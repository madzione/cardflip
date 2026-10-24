package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cardflip/golang/internal/usermsg"
	"cardflip/golang/storage"
)

// ErrCorrupt wraps errors caused by unreadable stored data.
var ErrCorrupt = storage.ErrCorrupt

// Error is a user-facing validation error: the operation was rejected and
// nothing changed. Its message is meant to be shown as is: it is tagged with
// a code and arguments (see package usermsg), so the app can show it
// translated.
type Error struct{ Msg string }

func (e *Error) Error() string { return e.Msg }

// Code returns the code of the error's message, such as "deckNameTaken".
func (e *Error) Code() string { return usermsg.Code(e.Msg) }

// userErr returns an *Error with the message format makes of args, tagged
// with code (see usermsg.Text).
func userErr(code, format string, args ...any) error {
	return &Error{Msg: usermsg.Text(code, format, args...)}
}

// NewError is userErr for the app's services and importers.
func NewError(code, format string, args ...any) error { return userErr(code, format, args...) }

// CheckUIScale fails unless percent is a UI size: MinUIScale to MaxUIScale
// in steps of UIScaleStep.
func CheckUIScale(percent int) error {
	if percent < MinUIScale || percent > MaxUIScale || percent%UIScaleStep != 0 {
		return userErr("uiScaleRange", "The UI size must be between %d and %d %% in steps of %d %%.", MinUIScale, MaxUIScale, UIScaleStep)
	}
	return nil
}

// Collection is an opened collection. All methods are safe for concurrent
// use: mutations are serialised, reads share a read lock.
type Collection struct {
	mu    sync.RWMutex
	store *storage.Store
	ids   *idGen
	now   func() time.Time

	settings  Settings
	presets   map[ID]*Preset
	decks     map[ID]*Deck
	noteTypes map[ID]*NoteType
	notes     map[ID]*Note
	cards     map[ID]*Card
	// nextPos is the position the next added New card gets: above every
	// card's NewPos seen since the collection was loaded.
	nextPos int64

	cardsByNote map[ID][]ID
	// noteFields is each note's fields as search matches them (see
	// searchFields; design D10), kept up to date as notes change.
	noteFields map[ID][]string
	// firstFields lists the notes by note type and normalised first field,
	// for the duplicate warning (see Duplicates).
	firstFields map[dupKey][]ID

	// texts and sortFields cache the browser's card text (see Rows).
	texts      textCache[cardText]
	sortFields textCache[string]

	// revlog holds the review log entries of the loaded months, which cover
	// the current study day, and of the months after them (those began while
	// the app was running, so all their entries were made in this session).
	revlog       []*ReviewEntry
	revlogMonths map[string]bool
	// reviews indexes the whole review log by card; nil until a search
	// first needs it (see loadReviewIndex).
	reviews reviewIndex
	// load counts the review cards due on each study day (see dueLoad).
	load dueLoad

	// hist is the undo history of this app session.
	hist history

	events *broadcaster

	// mediaIndex maps the SHA-1 of stored media files to their names; nil
	// until AddMedia first needs it, and after imports and restores.
	mediaIndex map[string]string
	// added holds the media files AddMedia stored or reused in this
	// session; compaction keeps them (see AddMedia).
	added map[string]bool

	// revision counts the changes since Open; see Revision.
	revision atomic.Uint64
	// yield is called periodically in long loops (web responsiveness).
	yield func()
	// builtin names what built-in things are created with.
	builtin BuiltinNames

	// Sync metadata (sync design decision 2, see sync_meta.go): the stamps
	// of the registers and the tombstones, the local change counter, the
	// collection's lineage, the device slot table, and the highest sequence
	// number of the review log entries of each month.
	clock  *hlc
	meta   map[metaKey]objMeta
	tombs  map[metaKey]objMeta
	seq    int64
	epoch  string
	slots  map[string]int
	revSeq map[string]int64
	// device is this device's ID (see SetDevice); empty until sync is set
	// up.
	device string
	// media is the media manifest (see media_sync.go), nil until it is
	// built; mediaBuild collects the files that change while it is being
	// built.
	media      map[string]mediaEntry
	mediaBuild map[string]bool
}

// Options configures Open.
type Options struct {
	// Now is the clock; nil means time.Now.
	Now func() time.Time
	// Yield is called periodically during long work; nil means no-op.
	Yield func()
	// Names are the names of built-in things in the interface language
	// (see SetBuiltinNames); the zero value means English.
	Names BuiltinNames
}

// Open loads the collection from b, creating a new one (with the Default
// deck and the built-in note types) when b is empty. fresh reports whether
// it was created. An error wrapping ErrCorrupt means the stored data could
// not be read; it was left untouched.
func Open(b storage.Backend, opts Options) (c *Collection, fresh bool, err error) {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	yield := opts.Yield
	if yield == nil {
		yield = func() {}
	}
	c = &Collection{now: now, yield: yield, ids: &idGen{now: now}, events: newBroadcaster(), added: map[string]bool{},
		builtin: opts.Names, clock: &hlc{wall: now}}
	c.reset()

	st, loaded, err := storage.Open(b)
	if err != nil {
		return nil, false, err
	}
	st.SetYield(yield)
	c.store = st
	if loaded.Fresh {
		c.initFresh()
		c.ensureMeta()
		snap, err := c.snapshot()
		if err != nil {
			return nil, false, err
		}
		if err := st.Init(snap); err != nil {
			return nil, false, err
		}
		if err := c.loadRevlog(); err != nil {
			return nil, false, err
		}
		return c, true, nil
	}

	normalized, version, err := c.loadSnapshot(loaded.Snapshot)
	if err != nil {
		return nil, false, err
	}
	for i, raw := range loaded.Ops {
		var o op
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, false, fmt.Errorf("%w: journal op %d: %v", ErrCorrupt, i, err)
		}
		c.apply(&o)
		if i%500 == 0 {
			c.yield()
		}
	}
	var renamed []string
	if !normalized {
		if renamed, err = c.migrateNFC(); err != nil {
			return nil, false, err
		}
	}
	if err := c.loadRevlog(); err != nil {
		return nil, false, err
	}
	presets := c.migrateDecks()
	if version < 3 {
		// Compacted below whether Default changes or not, so that a
		// later change to it is not undone on the next open.
		if p := unburyDefault(c.presets[DefaultPresetID]); p != nil {
			c.apply(&op{Presets: []*Preset{p}})
		}
	}
	migrated := c.migratePositions()
	c.ensureMeta()
	if len(loaded.Ops) > 0 || migrated || presets || !normalized || version < snapshotVersion {
		if err := c.compact(); err != nil {
			return nil, false, err
		}
	}
	// The media files are stored under their new names and the saved
	// collection refers to those, so the old names can go. Compaction
	// trashed most of them already, as nothing refers to them; a failure
	// leaves an unused file behind.
	_ = c.store.TrashMedia(renamed, c.now())
	return c, false, nil
}

func (c *Collection) reset() {
	c.texts = textCache[cardText]{}
	c.sortFields = textCache[string]{}
	c.settings = DefaultSettings()
	c.presets = map[ID]*Preset{}
	c.decks = map[ID]*Deck{}
	c.noteTypes = map[ID]*NoteType{}
	c.notes = map[ID]*Note{}
	c.cards = map[ID]*Card{}
	c.nextPos = 1
	c.cardsByNote = map[ID][]ID{}
	c.noteFields = map[ID][]string{}
	c.firstFields = map[dupKey][]ID{}
	c.revlog = nil
	c.revlogMonths = map[string]bool{}
	c.reviews = nil
	c.resetDueLoad()
	c.meta = map[metaKey]objMeta{}
	c.tombs = map[metaKey]objMeta{}
	c.seq = 0
	c.epoch = ""
	c.slots = nil
	c.revSeq = map[string]int64{}
	c.media = nil
}

func (c *Collection) initFresh() {
	p := defaultPreset()
	p.Name = c.builtin.presetName()
	c.presets[DefaultPresetID] = p
	d := &Deck{ID: c.ids.next(), Name: c.builtin.deckName(), PresetID: DefaultPresetID}
	c.decks[d.ID] = d
	for _, nt := range builtinNoteTypes(c.builtin) {
		nt.ID = c.ids.next()
		c.noteTypes[nt.ID] = nt
	}
	c.settings.LastDeckID = d.ID
}

// loadSnapshot loads a snapshot and reports whether it is normalised (see
// state.Normalized) and its version.
func (c *Collection) loadSnapshot(raw []byte) (normalized bool, version int, err error) {
	var s state
	if err := json.Unmarshal(raw, &s); err != nil {
		return false, 0, fmt.Errorf("%w: snapshot: %v", ErrCorrupt, err)
	}
	if s.Version > snapshotVersion {
		return false, 0, fmt.Errorf("%w: collection was saved by a newer Cardflip (format %d); update Cardflip", ErrCorrupt, s.Version)
	}
	if len(s.Decks) == 0 {
		return false, 0, fmt.Errorf("%w: snapshot has no decks", ErrCorrupt)
	}
	c.applyState(&s)
	return s.Normalized, s.Version, nil
}

// applyState replaces the in-memory collection with s.
func (c *Collection) applyState(s *state) {
	c.reset()
	c.settings = s.Settings
	if c.settings.Theme == "" {
		c.settings = DefaultSettings()
	}
	c.ids.observe(s.NextID)
	c.apply(&op{Presets: s.Presets, Decks: s.Decks, NoteTypes: s.NoteTypes, Notes: s.Notes, Cards: s.Cards})
	c.loadMeta(s)
}

// snapshot serialises the collection. Callers hold at least a read lock.
func (c *Collection) snapshot() ([]byte, error) {
	s := c.state()
	return json.Marshal(s)
}

func (c *Collection) state() *state {
	s := &state{Version: snapshotVersion, NextID: c.ids.high(), Settings: c.settings, Normalized: true}
	c.saveMeta(s)
	for _, p := range c.presets {
		s.Presets = append(s.Presets, p)
	}
	for _, d := range c.decks {
		s.Decks = append(s.Decks, d)
	}
	for _, nt := range c.noteTypes {
		s.NoteTypes = append(s.NoteTypes, nt)
	}
	for _, n := range c.notes {
		s.Notes = append(s.Notes, n)
	}
	for _, card := range c.cards {
		s.Cards = append(s.Cards, card)
	}
	sort.Slice(s.Presets, func(i, j int) bool { return s.Presets[i].ID < s.Presets[j].ID })
	sort.Slice(s.Decks, func(i, j int) bool { return s.Decks[i].ID < s.Decks[j].ID })
	sort.Slice(s.NoteTypes, func(i, j int) bool { return s.NoteTypes[i].ID < s.NoteTypes[j].ID })
	sort.Slice(s.Notes, func(i, j int) bool { return s.Notes[i].ID < s.Notes[j].ID })
	sort.Slice(s.Cards, func(i, j int) bool { return s.Cards[i].ID < s.Cards[j].ID })
	return s
}

// compact replaces the journal with a fresh snapshot and moves media that
// nothing references, not even the undo history, to the media trash.
// Callers hold the write lock (or are in Open).
func (c *Collection) compact() error {
	snap, err := c.snapshot()
	if err != nil {
		return err
	}
	return c.store.Compact(snap, c.referencedMedia(), staticMedia, c.now())
}

// mediaRef matches the files elements show or play, [sound:] tags and CSS
// url() values.
var mediaRef = regexp.MustCompile(`(?i)<(?:img|audio|video|source)\b[^>]*?\ssrc\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))` +
	`|\[sound:([^\]]*)\]` +
	`|\burl\(\s*(?:"([^"]*)"|'([^']*)'|([^)\s'"]*))\s*\)`)

// mediaRefs returns the raw media file names in fields, templates or CSS,
// without remote and data URLs.
func mediaRefs(htmls ...string) []string {
	var out []string
	for _, h := range htmls {
		for _, m := range mediaRef.FindAllStringSubmatch(h, -1) {
			name := strings.TrimSpace(strings.Join(m[1:], ""))
			lower := strings.ToLower(name)
			if name == "" || strings.Contains(name, "://") || strings.HasPrefix(name, "//") || strings.HasPrefix(lower, "data:") {
				continue
			}
			out = append(out, name)
		}
	}
	return out
}

// MediaNames returns the media file names referenced by fields, templates
// or CSS. A name with HTML entities or URL escapes is listed both as
// written and decoded, since any of them may be the stored name.
func MediaNames(htmls ...string) []string {
	var out []string
	for _, raw := range mediaRefs(htmls...) {
		out = append(out, mediaNameForms(raw)...)
	}
	return out
}

// mediaNameForms returns a referenced name as written, with HTML entities
// decoded, and then URL-decoded (as the app looks files up), without
// repeats. The second form is the name the reference means.
func mediaNameForms(raw string) []string {
	out := []string{raw}
	name := htmlUnescape(raw)
	if name != raw {
		out = append(out, name)
	}
	if u, err := url.PathUnescape(name); err == nil && u != name && u != raw {
		out = append(out, u)
	}
	return out
}

// staticMedia reports whether a media file is a static file of templates
// and styling, by the convention that their names start with "_". Such
// files are always in use: templates can refer to them in ways a search
// cannot see, such as from scripts.
func staticMedia(name string) bool { return strings.HasPrefix(name, "_") }

// referencedMedia lists the media that notes and note types use, including
// those the undo history can bring back and those added in this session.
func (c *Collection) referencedMedia() []string {
	var htmls []string
	notes, types := c.hist.notes()
	for _, n := range c.notes {
		notes = append(notes, n)
	}
	for _, nt := range c.noteTypes {
		types = append(types, nt)
	}
	for _, n := range notes {
		htmls = append(htmls, n.Fields...)
	}
	for _, nt := range types {
		htmls = append(htmls, noteTypeHTML(nt)...)
	}
	names := MediaNames(htmls...)
	for name := range c.added {
		names = append(names, name)
	}
	return names
}

// noteTypeHTML returns a note type's templates and CSS.
func noteTypeHTML(nt *NoteType) []string {
	out := []string{nt.CSS}
	for _, t := range nt.Templates {
		out = append(out, t.QFmt, t.AFmt)
	}
	return out
}

// op is one journaled change: the new versions of changed objects and the
// IDs of deleted ones. Applying it is deterministic, so replaying the
// journal rebuilds exactly the state the user saw.
type op struct {
	Kind         string      `json:"k"`
	NextID       int64       `json:"n,omitempty"`
	Settings     *Settings   `json:"s,omitempty"`
	Presets      []*Preset   `json:"p,omitempty"`
	Decks        []*Deck     `json:"d,omitempty"`
	NoteTypes    []*NoteType `json:"t,omitempty"`
	Notes        []*Note     `json:"no,omitempty"`
	Cards        []*Card     `json:"c,omitempty"`
	DelPresets   []ID        `json:"dp,omitempty"`
	DelDecks     []ID        `json:"dd,omitempty"`
	DelNoteTypes []ID        `json:"dt,omitempty"`
	DelNotes     []ID        `json:"dn,omitempty"`
	DelCards     []ID        `json:"dc,omitempty"`
	// Review is the entry of an answer, ReviewUndone the entry an undo
	// removes. Entries are stored in the review log segments (passed to
	// Commit as revlog writes), not in the journaled op; replay reloads them
	// from there.
	Review       *ReviewEntry `json:"-"`
	ReviewUndone ID           `json:"ru,omitempty"`
	// AddReviews and DelReviews are review log entries an import adds, and
	// those it replaces (and the same for its undo and redo). They are
	// stored in the review log segments like Review.
	AddReviews []*ReviewEntry `json:"-"`
	DelReviews []*ReviewEntry `json:"-"`
	// Trimmed are notes, as they were, whose extra fields with content
	// Check database removes. They go to the deletion log, not the journal.
	Trimmed []*Note `json:"-"`

	// Sync metadata (see stampSync): the stamps and sequence numbers of the
	// registers and tombstones the op changes, the highest sequence number
	// of the review log entries it adds in each month, a new collection
	// lineage and a new slot table.
	Meta   []syncEntry      `json:"m,omitempty"`
	Tombs  []syncEntry      `json:"x,omitempty"`
	RevSeq map[string]int64 `json:"rs,omitempty"`
	Epoch  string           `json:"e,omitempty"`
	Slots  map[string]int   `json:"sl,omitempty"`
	// Media are changes of the media manifest; MediaInit starts the
	// manifest (see BuildMediaManifest).
	Media     []mediaChange `json:"md,omitempty"`
	MediaInit bool          `json:"mi,omitempty"`
	// Redirects are the decks and note types the op deletes because they
	// were combined with others (see resolveClashes); stampSync records
	// them in the tombstones.
	Redirects []redirect `json:"-"`
	// Pruned are tombstones no link needs any more (PruneTombstones).
	Pruned []syncEntry `json:"pr,omitempty"`
}

// opSync is the kind of the op that applies changes received by device
// sync (see ApplyBatch).
const opSync = "sync"

// opSlots is the kind of the op that saves a new slot table (SetSlots).
const opSlots = "slots"

func (o *op) empty() bool {
	return o.Settings == nil && len(o.Presets)+len(o.Decks)+len(o.NoteTypes)+len(o.Notes)+len(o.Cards)+
		len(o.DelPresets)+len(o.DelDecks)+len(o.DelNoteTypes)+len(o.DelNotes)+len(o.DelCards) == 0 &&
		o.Review == nil && o.ReviewUndone == 0 && len(o.AddReviews)+len(o.DelReviews) == 0 &&
		len(o.Meta)+len(o.Tombs)+len(o.Media)+len(o.Pruned) == 0 && o.Epoch == "" && o.Slots == nil && !o.MediaInit
}

// commit saves o durably, then applies it, records it as an undo step and
// announces the change. If the save fails nothing changes. Callers hold the
// write lock.
func (c *Collection) commit(o *op, rev ...storage.RevlogWrite) error {
	return c.commitAs(o, rev, modeNew)
}

func (c *Collection) commitAs(o *op, rev []storage.RevlogWrite, mode commitMode) error {
	if o.empty() {
		if mode != modeNew {
			c.record(mode, o, &step{revert: &op{}})
		}
		return nil
	}
	if mode == modeNew && o.Kind != opSync {
		nfcOp(o)
		c.placeAdded(o)
		c.stampModified(o)
	}
	var inv *step
	if mode != modeNew || !untracked[o.Kind] {
		var err error
		if inv, err = c.inverse(o); err != nil {
			return err
		}
	}
	var dels []storage.Deletion
	if o.Kind != opRenumber {
		var err error
		if dels, err = c.deletions(o); err != nil {
			return err
		}
	}
	o.NextID = c.ids.high()
	rev = c.stampSync(o, rev)
	raw, err := json.Marshal(o)
	if err != nil {
		return err
	}
	if err := c.store.Commit(raw, rev, dels); err != nil {
		return fmt.Errorf("could not save the change: %w", err)
	}
	c.apply(o)
	c.record(mode, o, inv)
	c.revision.Add(1)
	c.events.publish(changeOf(o))
	if c.store.NeedsCompaction() {
		// The change is already safe in the journal; a failed compaction is
		// retried after the next change.
		_ = c.compact()
	}
	return nil
}

// stampModified sets Modified on the op's cards that are added or changed,
// unless the op set it itself: imported cards keep their source's time.
func (c *Collection) stampModified(o *op) {
	now := c.now()
	for _, card := range o.Cards {
		if old, ok := c.cards[card.ID]; ok {
			if !card.Modified.Equal(old.Modified) || reflect.DeepEqual(*old, *card) {
				continue
			}
		} else if !card.Modified.IsZero() {
			continue
		}
		card.Modified = now
	}
}

// apply changes the in-memory collection. It must not fail: everything was
// validated before the op was built.
func (c *Collection) apply(o *op) {
	c.ids.observe(o.NextID)
	c.applyMeta(o)
	c.applyMedia(o)
	c.texts.invalidate(o, c.cardsByNote)
	c.sortFields.invalidate(o, c.cardsByNote)
	if o.Settings != nil {
		s := *o.Settings
		c.settings = s
	}
	for _, p := range o.Presets {
		c.ids.observe(int64(p.ID))
		c.presets[p.ID] = p
	}
	for _, d := range o.Decks {
		c.ids.observe(int64(d.ID))
		c.decks[d.ID] = d
	}
	for _, nt := range o.NoteTypes {
		c.ids.observe(int64(nt.ID))
		c.noteTypes[nt.ID] = nt
	}
	for _, n := range o.Notes {
		c.ids.observe(int64(n.ID))
		c.indexFirstField(c.notes[n.ID], n)
		c.notes[n.ID] = n
		c.noteFields[n.ID] = searchFields(n.Fields)
	}
	for _, card := range o.Cards {
		c.ids.observe(int64(card.ID))
		c.nextPos = max(c.nextPos, card.NewPos+1)
		if old, ok := c.cards[card.ID]; ok && old.NoteID != card.NoteID {
			c.unlinkCard(old)
		}
		if old, ok := c.cards[card.ID]; ok {
			c.load.add(old, -1)
		} else {
			c.cardsByNote[card.NoteID] = append(c.cardsByNote[card.NoteID], card.ID)
		}
		c.cards[card.ID] = card
		c.load.add(card, 1)
	}
	for _, id := range o.DelCards {
		if card, ok := c.cards[id]; ok {
			c.unlinkCard(card)
			c.load.add(card, -1)
			delete(c.cards, id)
		}
	}
	if c.load.rollover != c.settings.RolloverHour {
		c.rebuildDueLoad()
	}
	for _, id := range o.DelNotes {
		if n, ok := c.notes[id]; ok {
			c.indexFirstField(n, nil)
		}
		delete(c.notes, id)
		delete(c.noteFields, id)
		delete(c.cardsByNote, id)
	}
	for _, id := range o.DelNoteTypes {
		delete(c.noteTypes, id)
	}
	for _, id := range o.DelDecks {
		delete(c.decks, id)
	}
	for _, id := range o.DelPresets {
		delete(c.presets, id)
	}
	c.applyReviews(o)
	if o.Review != nil {
		// Keep the log in ID order when redo brings an entry back.
		i := len(c.revlog)
		for i > 0 && c.revlog[i-1].ID > o.Review.ID {
			i--
		}
		c.revlog = append(c.revlog[:i:i], append([]*ReviewEntry{o.Review}, c.revlog[i:]...)...)
	}
	if o.ReviewUndone != 0 {
		for i, e := range c.revlog {
			if e.ID == o.ReviewUndone {
				c.revlog = append(c.revlog[:i:i], c.revlog[i+1:]...)
				break
			}
		}
	}
	if len(o.DelReviews) > 0 {
		del := make(map[ID]bool, len(o.DelReviews))
		for _, e := range o.DelReviews {
			del[e.ID] = true
		}
		kept := c.revlog[:0:0]
		for _, e := range c.revlog {
			if !del[e.ID] {
				kept = append(kept, e)
			}
		}
		c.revlog = kept
	}
	added := false
	for _, e := range o.AddReviews {
		if c.revlogInMemory(monthKey(e.Time)) {
			c.revlog = append(c.revlog, e)
			added = true
		}
	}
	if added {
		sort.SliceStable(c.revlog, func(i, j int) bool { return c.revlog[i].ID < c.revlog[j].ID })
	}
}

func (c *Collection) unlinkCard(card *Card) {
	ids := c.cardsByNote[card.NoteID]
	for i, id := range ids {
		if id == card.ID {
			ids = append(ids[:i:i], ids[i+1:]...)
			break
		}
	}
	if len(ids) == 0 {
		delete(c.cardsByNote, card.NoteID)
	} else {
		c.cardsByNote[card.NoteID] = ids
	}
}

// Close ends all change subscriptions. The storage backend belongs to the
// caller, which closes it.
func (c *Collection) Close() {
	c.events.close()
}

// Settings returns the app settings.
func (c *Collection) Settings() Settings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := c.settings
	s.LastDeckByNoteType = maps.Clone(s.LastDeckByNoteType)
	return s
}

// languageCode matches a language code such as "de" or "pt-BR".
var languageCode = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// UpdateSettings validates and saves new settings.
func (c *Collection) UpdateSettings(s Settings) error {
	switch s.Theme {
	case ThemeSystem, ThemeLight, ThemeDark:
	default:
		return userErr("unknownTheme", "Unknown theme %q.", s.Theme)
	}
	if s.Language != "" && !languageCode.MatchString(s.Language) {
		return userErr("unknownLanguage", "Unknown language %q.", s.Language)
	}
	if s.RolloverHour < 0 || s.RolloverHour > 23 {
		return userErr("rolloverHourRange", "The day rollover hour must be between 0 and 23.")
	}
	if s.TimeboxMinutes < 0 || s.TimeboxMinutes > MaxTimeboxMinutes {
		return userErr("timeboxRange", "The timebox time limit must be between 0 and %d minutes.", MaxTimeboxMinutes)
	}
	if f := s.QuestionFadeMs; f != nil && (*f < 0 || *f > MaxQuestionFadeMs) {
		return userErr("questionFadeRange", "The question fade must be between 0 and %d milliseconds.", MaxQuestionFadeMs)
	}
	if s.UIScale != 0 {
		if err := CheckUIScale(s.UIScale); err != nil {
			return err
		}
	}
	switch s.ReduceMotion {
	case "", ReduceMotionOn, ReduceMotionOff:
	default:
		return userErr("unknownReduceMotion", "Unknown reduce motion choice %q.", s.ReduceMotion)
	}
	if b := s.Backup; b.IntervalMinutes < 0 || b.IntervalMinutes > MaxBackupInterval {
		return userErr("backupIntervalRange", "The backup interval must be between 0 and %d minutes.", MaxBackupInterval)
	} else if min(b.Daily, b.Weekly, b.Monthly) < 0 || max(b.Daily, b.Weekly, b.Monthly) > MaxBackupsKept {
		return userErr("backupsKeptRange", "The number of backups to keep must be between 0 and %d.", MaxBackupsKept)
	}
	switch s.BrowserMode {
	case "", BrowserCards, BrowserNotes:
	default:
		return userErr("unknownBrowserMode", "Unknown browser mode %q.", s.BrowserMode)
	}
	if s.BrowserSplit < 0 || s.BrowserSplit >= 1 {
		return userErr("browserSplitRange", "The browser split must be between 0 and 1.")
	}
	for i, col := range s.BrowserColumns {
		if !col.Valid() {
			return userErr("unknownBrowserColumn", "Unknown browser column %q.", col)
		}
		if slices.Contains(s.BrowserColumns[:i], col) {
			return userErr("browserColumnTwice", "The browser column %q is shown twice.", col)
		}
	}
	for i, saved := range s.SavedSearches {
		if strings.TrimSpace(saved.Name) == "" {
			return userErr("savedSearchNameEmpty", "A saved search needs a name.")
		}
		if slices.ContainsFunc(s.SavedSearches[:i], func(o SavedSearch) bool { return strings.EqualFold(o.Name, saved.Name) }) {
			return userErr("savedSearchNameTaken", "A saved search named %q already exists.", saved.Name)
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.equal(c.settings) {
		return nil
	}
	return c.commit(&op{Kind: "settings", Settings: &s})
}

// SetCurrentDeck records the deck most recently studied (see
// Settings.CurrentDeckID and Settings.RecentDecks).
func (c *Collection) SetCurrentDeck(id ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.decks[id]; !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	recent := c.recentDecks()
	if c.settings.CurrentDeckID == id && len(recent) > 0 && recent[0] == id {
		return nil
	}
	recent = slices.DeleteFunc(recent, func(d ID) bool { return d == id })
	recent = append([]ID{id}, recent[:min(len(recent), MaxRecentDecks-1)]...)
	s := c.settings
	s.CurrentDeckID = id
	s.RecentDecks = recent
	return c.commit(&op{Kind: "settings", Settings: &s})
}

// RecentDecks returns the decks most recently studied, latest first,
// leaving out deleted ones (see Settings.RecentDecks).
func (c *Collection) RecentDecks() []ID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.recentDecks()
}

// recentDecks returns a new slice of the recent decks that still exist.
func (c *Collection) recentDecks() []ID {
	out := make([]ID, 0, len(c.settings.RecentDecks))
	for _, id := range c.settings.RecentDecks {
		if _, ok := c.decks[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// Revision increases with every saved change, restores included. Equal
// revisions of the same Collection mean nothing changed in between.
func (c *Collection) Revision() uint64 { return c.revision.Load() }

// Now returns the collection's clock reading.
func (c *Collection) Now() time.Time { return c.now() }

// IsUserError reports whether err is a validation error meant for the user.
func IsUserError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}
