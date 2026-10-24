package core

import (
	"bytes"
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// DupMode says what a CSV import does with a row whose first field matches
// an existing note of the same note type.
type DupMode string

const (
	DupSkip      DupMode = "skip"
	DupUpdate    DupMode = "update"
	DupDuplicate DupMode = "duplicate"
)

// DupScope says which notes a CSV row's first field is compared with: the
// notes of its note type (the default, also when empty), or only those of
// them with a card in the row's deck.
type DupScope string

const (
	DupScopeNoteType DupScope = "noteType"
	DupScopeDeck     DupScope = "deck"
)

// UpdatePolicy says when an import updates a note or note type that
// already exists.
type UpdatePolicy string

const (
	// UpdateIfNewer updates when the imported version was modified after the
	// collection's. It is the default.
	UpdateIfNewer UpdatePolicy = "ifNewer"
	UpdateAlways  UpdatePolicy = "always"
	UpdateNever   UpdatePolicy = "never"
)

// allows reports whether the policy updates a version modified at existing
// with one modified at imported. A zero time is older than any other.
func (p UpdatePolicy) allows(imported, existing time.Time) bool {
	switch p {
	case UpdateAlways:
		return true
	case UpdateNever:
		return false
	}
	return imported.After(existing)
}

func (p UpdatePolicy) valid() bool {
	switch p {
	case "", UpdateIfNewer, UpdateAlways, UpdateNever:
		return true
	}
	return false
}

// ImportBatch is everything one import adds; it is committed all at once.
type ImportBatch struct {
	NoteTypes []ImportNoteType
	Notes     []ImportNote
	// Media holds image, audio and video files by name; audio and video in
	// formats the app cannot play are filtered out by the importer (with a
	// warning).
	Media map[string][]byte
	// Duplicates is set for a text import, empty for a package. A row whose
	// GUID matches an existing note of the same note type updates it;
	// otherwise Duplicates says what happens when its first field matches
	// an existing note's within DupScope.
	Duplicates DupMode
	DupScope   DupScope
	// Skipped counts the rows the importer left out of the batch; they are
	// counted as skipped notes.
	Skipped int
	// UpdateNotes decides whether notes of a package that exist (matched by
	// GUID) are updated, UpdateNoteTypes whether existing note types get the
	// batch's templates and styling; empty means UpdateIfNewer. MergeNoteTypes
	// merges a note type whose fields or card types differ from the
	// existing one into it, instead of adding it as a separate note type.
	UpdateNotes     UpdatePolicy
	UpdateNoteTypes UpdatePolicy
	MergeNoteTypes  bool
	Warnings        []string
	// Presets are deck options presets for the decks the import creates.
	Presets []ImportPreset
	// DeckPresets maps deck names (ignoring case) to the Key of their
	// preset. Decks the import creates use it, or Default without one;
	// existing decks keep their preset.
	DeckPresets map[string]string
	// Decks are created even when no card goes into them (the empty decks
	// of a collection package).
	Decks []string
	// DeckDescriptions maps deck names (ignoring case) to their
	// description. Decks the import creates get it; existing decks keep
	// theirs.
	DeckDescriptions map[string]string
}

// ImportPreset is a deck options preset of the batch. It is added when a
// deck the import creates uses it, unless a preset with the same name (or
// that name numbered, as an earlier import named it) and options exists.
type ImportPreset struct {
	Key     string
	Name    string
	Options DeckOptions
}

// ImportNoteType is a note type used by the batch: an existing one
// (ExistingID), or one from the package, which is matched against the
// existing note types of the same origin (see Collection.Import) and
// created when none fits. Type.Modified is when it last changed there.
type ImportNoteType struct {
	Key        string
	ExistingID ID
	Type       NoteType
	// TemplateDecks names the deck each card type of Type overrides to,
	// parallel to Type.Templates ("" or missing for none). An override is
	// kept when the deck exists after the import and Type is added or
	// updated.
	TemplateDecks []string
}

// ImportNote is a note to add, or to update when a note with the same GUID
// exists.
type ImportNote struct {
	TypeKey string
	GUID    string
	// DeckName is the deck of generated cards (created if missing); DeckID
	// is used when DeckName is empty.
	DeckName string
	DeckID   ID
	Fields   []string
	Tags     []string
	// Modified is when the note last changed where it comes from; zero when
	// unknown, which counts as older than any time.
	Modified time.Time
	// Cards, when set, are the note's cards with their scheduling (Anki
	// import). When nil, cards are generated New.
	Cards []ImportCard
}

// ImportCard is an imported card.
type ImportCard struct {
	Ord       int
	DeckName  string
	State     scheduler.CardState
	Suspended bool
	// Flag is 0 for no flag, or 1 to 7; other values are ignored.
	Flag int
	// Order sorts imported New cards among themselves.
	Order int64
	// OriginalOrder is, for a card that is no longer New, where it was
	// among the new cards where it comes from, on the scale of Order; nil
	// when unknown. The card's OriginalPos is set from it.
	OriginalOrder *int64
	// Modified is when the card last changed where it comes from; zero
	// when unknown.
	Modified time.Time
	// History is the card's review history, oldest first. It is added with
	// the card, so cards of notes that already exist get none.
	History []ImportReview
}

// ImportReview is a review log entry of an imported card.
type ImportReview struct {
	Time time.Time
	// Rating is 0 for a manual entry (a reschedule or a reset).
	Rating      scheduler.Rating
	StateBefore scheduler.State
	// IvlBefore and IvlAfter are the intervals in days.
	IvlBefore, IvlAfter int
	DurationMs          int64
	// Source identifies the entry where it came from, e.g. "anki:<revlog
	// id>"; an entry whose source the collection already has is skipped.
	Source string
}

// ImportSummary reports what an import did.
type ImportSummary struct {
	NotesAdded   int `json:"notesAdded"`
	NotesUpdated int `json:"notesUpdated"`
	NotesSkipped int `json:"notesSkipped"`
	// NotesSkippedOlder counts the notes not updated because the batch's
	// version was not modified after the collection's ("if newer"); they
	// are not in NotesSkipped.
	NotesSkippedOlder int `json:"notesSkippedOlder"`
	CardsAdded        int `json:"cardsAdded"`
	// NoteTypesAdded, NoteTypesUpdated and NoteTypesMerged count the note
	// types created, existing ones given new templates or styling, and
	// existing ones that gained fields or card types.
	NoteTypesAdded   int `json:"noteTypesAdded"`
	NoteTypesUpdated int `json:"noteTypesUpdated"`
	NoteTypesMerged  int `json:"noteTypesMerged"`
	// Reviews counts the review history entries added.
	Reviews int `json:"reviews"`
	Media   int `json:"media"`
	// CommentLines counts the comment lines skipped in a text file; the
	// text importer sets it.
	CommentLines int      `json:"commentLines"`
	Warnings     []string `json:"warnings"`
}

// Progress reports import progress: done of total steps in a stage.
type Progress func(stage string, done, total int)

// sameOrigin reports whether the existing note type a is where the
// imported note type b came from: the same source or, without one, the same
// name. A package exported from this collection names its own note types by
// their ID ("anki:<ID>").
func sameOrigin(a, b *NoteType) bool {
	if b.Source == "" {
		return strings.EqualFold(a.Name, b.Name)
	}
	return a.Source == b.Source || b.Source == fmt.Sprintf("anki:%d", a.ID)
}

// sameStructure reports whether notes and cards of b map one to one onto a:
// the same kind, fields and card types, by name and in order. A cloze note
// type's only card type matches whatever its name.
func sameStructure(a, b *NoteType) bool {
	if a.Kind != b.Kind || len(a.Fields) != len(b.Fields) || len(a.Templates) != len(b.Templates) {
		return false
	}
	for i := range a.Fields {
		if a.Fields[i] != b.Fields[i] {
			return false
		}
	}
	for i := range a.Templates {
		if a.Kind != Cloze && !strings.EqualFold(a.Templates[i].Name, b.Templates[i].Name) {
			return false
		}
	}
	return true
}

// sameLook reports whether two note types of the same structure have the
// same templates and styling, LaTeX image format included.
func sameLook(a, b *NoteType) bool {
	if a.CSS != b.CSS || a.LatexSVG != b.LatexSVG {
		return false
	}
	for i := range a.Templates {
		if !sameFormats(a.Templates[i], b.Templates[i]) {
			return false
		}
	}
	return true
}

// sameFormats reports whether two card types have the same templates,
// browser appearance included.
func sameFormats(a, b CardTemplate) bool {
	return a.QFmt == b.QFmt && a.AFmt == b.AFmt && a.BrowserQFmt == b.BrowserQFmt && a.BrowserAFmt == b.BrowserAFmt
}

// takeFormats gives t the templates of from, browser appearance included.
func (t *CardTemplate) takeFormats(from CardTemplate) {
	t.QFmt, t.AFmt, t.BrowserQFmt, t.BrowserAFmt = from.QFmt, from.AFmt, from.BrowserQFmt, from.BrowserAFmt
}

// importType is the note type the notes of a batch note type go to, with
// the index there of each of the batch type's fields and card types (nil
// when they are the same). takeDecks is set when the import adds or updates
// that note type, so its card types take the batch's deck overrides.
type importType struct {
	nt        *NoteType
	fields    []int
	ords      []int
	takeDecks bool
}

// field returns the index of the batch type's field i, or -1.
func (t *importType) field(i int) int {
	if t.fields == nil {
		return i
	}
	if i < len(t.fields) {
		return t.fields[i]
	}
	return -1
}

// ord returns the card ord for the batch type's card ord i, or -1.
func (t *importType) ord(i int) int {
	if t.ords == nil || t.nt.Kind == Cloze {
		return i
	}
	if i >= 0 && i < len(t.ords) {
		return t.ords[i]
	}
	return -1
}

// importedTime is the modification time of something taken from a batch:
// the batch's time, or now when it has none.
func (c *Collection) importedTime(t time.Time) time.Time {
	if t.IsZero() {
		return c.now()
	}
	return t
}

// importNoteType finds or creates the note type for the batch note type in,
// among the existing note types of the same origin, oldest first: an
// identical one; else one of the same structure, which gets in's templates
// and styling when b.UpdateNoteTypes allows; else, with b.MergeNoteTypes,
// one of the same kind, merged with in (see mergeNoteType). Without such a
// note type, in is added as a new one. Note types it changes go into o and
// changed. Callers hold the write lock.
func (c *Collection) importNoteType(in *NoteType, b *ImportBatch, o *op, changed map[ID]*NoteType, sum *ImportSummary) *importType {
	var cands []*NoteType
	for _, nt := range c.noteTypes {
		if sameOrigin(nt, in) {
			if p := changed[nt.ID]; p != nil {
				nt = p
			}
			cands = append(cands, nt)
		}
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].ID < cands[j].ID })
	for _, nt := range cands {
		if sameStructure(nt, in) && sameLook(nt, in) {
			return &importType{nt: nt}
		}
	}
	update := func(nt *NoteType) bool { return b.UpdateNoteTypes.allows(in.Modified, nt.Modified) }
	for _, nt := range cands {
		if !sameStructure(nt, in) {
			continue
		}
		if !update(nt) {
			return &importType{nt: nt}
		}
		nn := cloneNoteType(nt)
		nn.CSS, nn.LatexSVG = in.CSS, in.LatexSVG
		for i := range nn.Templates {
			nn.Templates[i].takeFormats(in.Templates[i])
		}
		nn.Modified = c.importedTime(in.Modified)
		changeImportedType(nn, o, changed)
		sum.NoteTypesUpdated++
		return &importType{nt: nn, takeDecks: true}
	}
	if b.MergeNoteTypes {
		for _, nt := range cands {
			if nt.Kind == in.Kind {
				return c.mergeNoteType(nt, in, update(nt), o, changed, sum)
			}
		}
	}
	nt := cloneNoteType(in)
	nt.ID = c.ids.next()
	nt.Name = c.uniqueNoteTypeName(nt.Name, o)
	nt.Modified = c.importedTime(in.Modified)
	o.NoteTypes = append(o.NoteTypes, nt)
	sum.NoteTypesAdded++
	return &importType{nt: nt, takeDecks: true}
}

// mergeNoteType merges the batch note type in into the existing note type
// nt: fields and card types are matched by name, ignoring case (a cloze
// note type's only card type matches whatever its name), the missing ones
// are added at the end, and none is removed. With update, matched card
// types get in's templates and nt gets in's styling. Templates taken from in
// name the fields as nt does.
func (c *Collection) mergeNoteType(nt, in *NoteType, update bool, o *op, changed map[ID]*NoteType, sum *ImportSummary) *importType {
	nn := cloneNoteType(nt)
	t := &importType{nt: nn, fields: make([]int, len(in.Fields)), ords: make([]int, len(in.Templates))}
	added, restyled := false, false
	var renames [][2]string
	for i, f := range in.Fields {
		j := slices.IndexFunc(nn.Fields, func(g string) bool { return strings.EqualFold(f, g) })
		if j < 0 {
			nn.Fields = append(nn.Fields, f)
			j, added = len(nn.Fields)-1, true
		} else if nn.Fields[j] != f {
			renames = append(renames, [2]string{f, nn.Fields[j]})
		}
		t.fields[i] = j
	}
	for i, tm := range in.Templates {
		for _, r := range renames {
			for _, f := range []*string{&tm.QFmt, &tm.AFmt, &tm.BrowserQFmt, &tm.BrowserAFmt} {
				*f = render.RenameField(*f, r[0], r[1])
			}
		}
		j := 0
		if nn.Kind != Cloze {
			j = slices.IndexFunc(nn.Templates, func(o CardTemplate) bool { return strings.EqualFold(o.Name, tm.Name) })
		}
		switch {
		case j < 0:
			nn.Templates = append(nn.Templates, tm)
			j, added = len(nn.Templates)-1, true
		case update && !sameFormats(nn.Templates[j], tm):
			nn.Templates[j].takeFormats(tm)
			restyled = true
		}
		t.ords[i] = j
	}
	if update && (nn.CSS != in.CSS || nn.LatexSVG != in.LatexSVG) {
		nn.CSS, nn.LatexSVG = in.CSS, in.LatexSVG
		restyled = true
	}
	switch {
	case restyled:
		nn.Modified = c.importedTime(in.Modified)
	case added:
		nn.Modified = c.now()
	default:
		t.nt = nt
		return t
	}
	if added {
		sum.NoteTypesMerged++
	} else {
		sum.NoteTypesUpdated++
	}
	changeImportedType(nn, o, changed)
	t.takeDecks = true
	return t
}

// importDeckOverrides gives the card types of the note types the import
// adds or updates the deck overrides of the batch, by deck name, when that
// deck exists now: from before or created by the import, and not filtered.
// It runs once all decks of the import exist.
func (c *Collection) importDeckOverrides(b *ImportBatch, types map[string]*importType, changed map[ID]*NoteType, created map[string]*Deck) {
	for _, bt := range b.NoteTypes {
		it := types[bt.Key]
		if it == nil || !it.takeDecks {
			continue
		}
		nt := it.nt
		if p := changed[nt.ID]; p != nil {
			nt = p
		}
		for i, name := range bt.TemplateDecks {
			j := it.ord(i)
			if name == "" || j < 0 || j >= len(nt.Templates) {
				continue
			}
			name, err := normalizeDeckName(name)
			if err != nil {
				continue
			}
			d := c.deckByName(name)
			if d == nil {
				d = created[strings.ToLower(name)]
			}
			if d != nil && d.Filter == nil {
				nt.Templates[j].DeckOverride = d.ID
			}
		}
	}
}

// changeImportedType puts the changed version of an existing note type
// into o and changed, replacing an earlier change of the same import.
func changeImportedType(nt *NoteType, o *op, changed map[ID]*NoteType) {
	if changed[nt.ID] != nil {
		for i, p := range o.NoteTypes {
			if p.ID == nt.ID {
				o.NoteTypes[i] = nt
			}
		}
	} else {
		o.NoteTypes = append(o.NoteTypes, nt)
	}
	changed[nt.ID] = nt
}

// firstFieldKey is what duplicate detection compares: the first field,
// HTML-stripped and trimmed.
func firstFieldKey(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	return collapseSpace(stripHTML(fields[0]))
}

// Import adds a batch to the collection in one durable step: either all of
// it is saved, or nothing changes.
func (c *Collection) Import(b *ImportBatch, progress Progress) (ImportSummary, error) {
	if progress == nil {
		progress = func(string, int, int) {}
	}
	b = nfcBatch(b)
	c.mu.Lock()
	defer c.mu.Unlock()
	if !b.UpdateNotes.valid() || !b.UpdateNoteTypes.valid() {
		return ImportSummary{}, fmt.Errorf("import: unknown update policy %q/%q", b.UpdateNotes, b.UpdateNoteTypes)
	}
	switch b.DupScope {
	case "", DupScopeNoteType, DupScopeDeck:
	default:
		return ImportSummary{}, fmt.Errorf("import: unknown duplicate scope %q", b.DupScope)
	}
	// text is set for a text import (see ImportBatch.Duplicates).
	text := b.Duplicates != ""
	sum := ImportSummary{NotesSkipped: b.Skipped, Warnings: append([]string(nil), b.Warnings...)}
	o := &op{Kind: "import"}
	createdDecks := map[string]*Deck{}

	// Note types. changed holds the new versions of existing note types the
	// import updates or merges.
	types := map[string]*importType{}
	changed := map[ID]*NoteType{}
	for _, it := range b.NoteTypes {
		if it.ExistingID != 0 {
			nt, ok := c.noteTypes[it.ExistingID]
			if !ok {
				return ImportSummary{}, userErr("noteTypeGone", "That note type no longer exists.")
			}
			types[it.Key] = &importType{nt: nt}
			continue
		}
		types[it.Key] = c.importNoteType(&it.Type, b, o, changed, &sum)
	}
	typeOf := func(id ID) *NoteType {
		if nt := changed[id]; nt != nil {
			return nt
		}
		return c.noteTypes[id]
	}

	// Media: names that clash with different content get a new name, and
	// the batch's references are rewritten.
	renames := map[string]string{}
	media := map[string][]byte{}
	names := make([]string, 0, len(b.Media))
	for name := range b.Media {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data := b.Media[name]
		if existing, err := c.store.GetMedia(name); err == nil {
			if bytes.Equal(existing, data) {
				continue
			}
			renamed := hashedMediaName(name, data)
			renames[name] = renamed
			name = renamed
		}
		media[name] = data
	}

	// Existing notes by GUID, and by first field per note type for CSV
	// duplicate detection, oldest first.
	byGUID := map[string]*Note{}
	byFirst := map[ID]map[string][]*Note{}
	for _, n := range c.notes {
		if n.GUID != "" {
			byGUID[n.GUID] = n
		}
	}
	firstIndex := func(typeID ID) map[string][]*Note {
		if m, ok := byFirst[typeID]; ok {
			return m
		}
		m := map[string][]*Note{}
		for _, n := range c.notes {
			if n.TypeID == typeID {
				k := firstFieldKey(n.Fields)
				m[k] = append(m[k], n)
			}
		}
		for _, ns := range m {
			sort.Slice(ns, func(i, j int) bool { return ns[i].ID < ns[j].ID })
		}
		byFirst[typeID] = m
		return m
	}
	updated := map[ID]bool{}
	// pendingNotes are the notes o adds or changes, addedNotes the ones it
	// adds, with the deck of their cards.
	pendingNotes, addedNotes := map[ID]bool{}, map[ID]ID{}
	// existingDeck is the deck an import note goes to when the deck exists
	// already, or 0.
	existingDeck := func(in *ImportNote) ID {
		if in.DeckName == "" {
			return in.DeckID
		}
		name, err := normalizeDeckName(in.DeckName)
		if err != nil {
			return 0
		}
		if d := c.deckByName(name); d != nil {
			return d.ID
		}
		if d, ok := createdDecks[strings.ToLower(name)]; ok {
			return d.ID
		}
		return 0
	}
	// inDeck reports whether note n has a card in deck.
	inDeck := func(n *Note, deck ID) bool {
		if d, ok := addedNotes[n.ID]; ok {
			return d == deck
		}
		for _, cid := range c.cardsByNote[n.ID] {
			if c.cards[cid].HomeDeck() == deck {
				return true
			}
		}
		return false
	}
	// duplicate finds the note whose first field a text row's matches.
	duplicate := func(in *ImportNote, typeID ID, fields []string) *Note {
		cands := firstIndex(typeID)[firstFieldKey(fields)]
		if b.DupScope != DupScopeDeck {
			if len(cands) == 0 {
				return nil
			}
			return cands[0]
		}
		deck := existingDeck(in)
		if deck == 0 {
			return nil
		}
		for _, n := range cands {
			if inDeck(n, deck) {
				return n
			}
		}
		return nil
	}

	// pending is a card that gets the next position: a New card, or with
	// original set, a card that is not New and gets its OriginalPos.
	type pending struct {
		card     *Card
		order    int64
		original bool
	}
	var newCards []pending
	var hist []*ReviewEntry

	total := len(b.Notes)
	for i, in := range b.Notes {
		if i%200 == 0 {
			progress("notes", i, total)
			c.yield()
		}
		it, ok := types[in.TypeKey]
		if !ok {
			return ImportSummary{}, fmt.Errorf("import: unknown note type %q", in.TypeKey)
		}
		nt := it.nt
		fields := make([]string, len(nt.Fields))
		for j, f := range in.Fields {
			if k := it.field(j); k >= 0 && k < len(fields) {
				fields[k] = rewriteMedia(f, renames)
			}
		}
		tags := NormalizeTags(in.Tags)

		var existing *Note
		if in.GUID != "" {
			existing = byGUID[in.GUID]
			// A package's note may change note type when the fields still
			// fit; a text row only updates a note of its note type.
			if existing != nil && existing.TypeID != nt.ID && (text || len(typeOf(existing.TypeID).Fields) != len(fields)) {
				sum.NotesSkipped++
				sum.Warnings = append(sum.Warnings, usermsg.Text("importOtherNoteType", "A note already exists with a different note type and was not updated: %s", firstFieldKey(fields)))
				continue
			}
		}
		if existing == nil && text && b.Duplicates != DupDuplicate {
			existing = duplicate(&in, nt.ID, fields)
			if existing != nil && b.Duplicates != DupUpdate {
				sum.NotesSkipped++
				continue
			}
		}

		if existing != nil {
			if updated[existing.ID] {
				sum.NotesSkipped++
				continue
			}
			updated[existing.ID] = true
			etype := typeOf(existing.TypeID)
			if slices.Equal(padFields(existing.Fields, len(etype.Fields)), fields) && slices.Equal(existing.Tags, tags) {
				sum.NotesSkipped++
				continue
			}
			if !text && !b.UpdateNotes.allows(in.Modified, existing.Modified) {
				if b.UpdateNotes == UpdateNever {
					sum.NotesSkipped++
				} else {
					sum.NotesSkippedOlder++
				}
				continue
			}
			ords := cardOrds(etype, fields)
			if len(ords) == 0 {
				sum.NotesSkipped++
				sum.Warnings = append(sum.Warnings, usermsg.Text("importNotUpdatedNoCards", "Not updated, it would produce no cards: %s", firstFieldKey(existing.Fields)))
				continue
			}
			nn := cloneNote(existing)
			nn.Fields = fields
			nn.Tags = tags
			nn.Modified = c.importedTime(in.Modified)
			o.Notes = append(o.Notes, &nn)
			pendingNotes[nn.ID] = true
			before := len(o.Cards)
			c.syncCards(o, etype, &nn, ords, c.deckOfNote(nn.ID), nil)
			sum.CardsAdded += len(o.Cards) - before
			sum.NotesUpdated++
			continue
		}

		deckID := in.DeckID
		if in.DeckName != "" {
			id, err := c.deckForImport(in.DeckName, o, createdDecks)
			if err != nil && text {
				// A row's deck column: skip the row, not the file.
				sum.NotesSkipped++
				sum.Warnings = append(sum.Warnings, usermsg.Text("importSkipped", "Skipped %q: %v", truncate(strings.Join(in.Fields, " | "), 60), err))
				continue
			}
			if err != nil {
				return ImportSummary{}, err
			}
			deckID = id
		}
		if d, ok := c.decks[deckID]; !ok && !deckCreated(createdDecks, deckID) {
			return ImportSummary{}, userErr("deckGone", "That deck no longer exists.")
		} else if ok && d.Filter != nil {
			return ImportSummary{}, errFilteredTarget(d)
		}
		n := &Note{ID: c.ids.next(), TypeID: nt.ID, GUID: in.GUID, Fields: fields, Tags: tags, Modified: c.importedTime(in.Modified)}
		if n.GUID == "" {
			n.GUID = newGUID()
		}
		made := 0
		if in.Cards != nil {
			seen := map[int]bool{}
			for _, ic := range in.Cards {
				ord := it.ord(ic.Ord)
				if seen[ord] || ord < 0 || nt.Kind == Standard && ord >= len(nt.Templates) {
					continue
				}
				seen[ord] = true
				did := deckID
				if ic.DeckName != "" {
					id, err := c.deckForImport(ic.DeckName, o, createdDecks)
					if err != nil {
						return ImportSummary{}, err
					}
					did = id
				}
				card := &Card{ID: c.ids.next(), NoteID: n.ID, DeckID: did, Ord: ord, Suspended: ic.Suspended, CardState: ic.State,
					Modified: c.importedTime(ic.Modified)}
				if ic.Flag > 0 && ic.Flag <= MaxFlag {
					card.Flag = ic.Flag
				}
				o.Cards = append(o.Cards, card)
				made++
				for _, h := range ic.History {
					if h.Rating < 0 || h.Rating > scheduler.Easy {
						continue
					}
					hist = append(hist, &ReviewEntry{ID: c.ids.next(), CardID: card.ID, Time: h.Time, Rating: h.Rating,
						StateBefore: h.StateBefore, IvlBefore: h.IvlBefore, IvlAfter: h.IvlAfter,
						DurationMs: max(h.DurationMs, 0), Source: h.Source})
				}
				if card.State == scheduler.New {
					newCards = append(newCards, pending{card, ic.Order, false})
				} else if ic.OriginalOrder != nil {
					newCards = append(newCards, pending{card, *ic.OriginalOrder, true})
				}
			}
		}
		if made == 0 {
			ords := cardOrds(nt, fields)
			if len(ords) == 0 {
				sum.NotesSkipped++
				sum.Warnings = append(sum.Warnings, usermsg.Text("importSkippedNoCards", "Skipped, it would produce no cards: %q", truncate(strings.Join(in.Fields, " | "), 60)))
				continue
			}
			before := len(o.Cards)
			c.syncCards(o, nt, n, ords, deckID, nil)
			for _, card := range o.Cards[before:] {
				newCards = append(newCards, pending{card, int64(i), false})
			}
		}
		o.Notes = append(o.Notes, n)
		pendingNotes[n.ID] = true
		addedNotes[n.ID] = deckID
		if in.GUID != "" {
			byGUID[in.GUID] = n
		}
		if text {
			k := firstFieldKey(fields)
			firstIndex(nt.ID)[k] = append(firstIndex(nt.ID)[k], n)
		}
		sum.NotesAdded++
	}
	progress("notes", total, total)

	// The other notes of updated or merged note types get the added fields,
	// empty, and their cards are regenerated, as when the note type is
	// edited. Their modification time stays: their content did not change.
	changedIDs := slices.Sorted(maps.Keys(changed))
	for _, id := range changedIDs {
		nt := changed[id]
		var noteIDs []ID
		for nid, n := range c.notes {
			if n.TypeID == id && !pendingNotes[nid] {
				noteIDs = append(noteIDs, nid)
			}
		}
		sortIDs(noteIDs)
		for j, nid := range noteIDs {
			if j%500 == 0 {
				c.yield()
			}
			note := c.notes[nid]
			if len(note.Fields) < len(nt.Fields) {
				nn := cloneNote(note)
				nn.Fields = padFields(nn.Fields, len(nt.Fields))
				o.Notes = append(o.Notes, &nn)
				note = &nn
			}
			before := len(o.Cards)
			c.syncCards(o, nt, note, cardOrds(nt, note.Fields), c.deckOfNote(nid), nil)
			sum.CardsAdded += len(o.Cards) - before
		}
	}
	for _, name := range b.Decks {
		if _, err := c.deckForImport(name, o, createdDecks); err != nil {
			return ImportSummary{}, err
		}
	}
	c.importDeckOverrides(b, types, changed, createdDecks)
	c.importPresets(b, o, createdDecks, &sum)
	importDescriptions(b, createdDecks)

	// Count cards of added notes, and order the new ones.
	for _, card := range o.Cards {
		if _, ok := addedNotes[card.NoteID]; ok {
			sum.CardsAdded++
		}
	}
	// New cards go after all existing ones, in import order. Cards that
	// are no longer New but know their place among them keep it as their
	// original position.
	sort.SliceStable(newCards, func(i, j int) bool { return newCards[i].order < newCards[j].order })
	newPos := c.nextPos - 1
	for _, p := range newCards {
		newPos++
		if p.original {
			pos := newPos
			p.card.OriginalPos = &pos
		} else {
			p.card.NewPos = newPos
		}
	}
	c.nextPos = newPos + 1

	// Review history of the added cards. An entry whose source is stored
	// already is skipped, unless its card was deleted (a deleted note, or
	// an undone import): then the new card takes it over.
	if len(hist) > 0 {
		stored, err := c.sourcedReviews(hist, progress)
		if err != nil {
			return ImportSummary{}, err
		}
		seen := map[string]bool{}
		for _, e := range hist {
			if e.Source != "" {
				if seen[e.Source] {
					continue
				}
				seen[e.Source] = true
				if old := stored[e.Source]; old != nil {
					if _, ok := c.cards[old.CardID]; ok {
						continue
					}
					o.DelReviews = append(o.DelReviews, old)
				}
			}
			o.AddReviews = append(o.AddReviews, e)
		}
		sum.Reviews = len(o.AddReviews)
	}

	// Media first: if the commit fails, unreferenced files are removed at
	// the next compaction.
	c.mediaIndex = nil
	mnames := make([]string, 0, len(media))
	for name := range media {
		mnames = append(mnames, name)
	}
	sort.Strings(mnames)
	for i, name := range mnames {
		progress("media", i, len(mnames))
		if err := c.store.PutMedia(name, media[name]); err != nil {
			return ImportSummary{}, fmt.Errorf("could not save media file %s: %w", name, err)
		}
		c.mediaChanged(o, name, sha1Hex(media[name]), false)
		c.yield()
	}
	sum.Media = len(mnames)
	progress("saving", 0, 1)
	rev, err := reviewWrites(o.AddReviews, o.DelReviews)
	if err != nil {
		return ImportSummary{}, err
	}
	if err := c.commit(o, rev...); err != nil {
		return ImportSummary{}, err
	}
	progress("saving", 1, 1)
	return sum, nil
}

// sourcedReviews returns the stored review log entries that have a source,
// by source, from the months the entries of hist fall in.
func (c *Collection) sourcedReviews(hist []*ReviewEntry, progress Progress) (map[string]*ReviewEntry, error) {
	want := map[string]bool{}
	for _, e := range hist {
		if e.Source == "" {
			continue
		}
		// Months are local: after a time zone change, an entry stored by an
		// earlier import can be in the neighbouring month.
		for _, d := range []int{-2, 0, 2} {
			want[monthKey(e.Time.AddDate(0, 0, d))] = true
		}
	}
	all, err := c.store.RevlogMonths()
	if err != nil {
		return nil, err
	}
	var months []string
	for _, m := range all {
		if want[m] {
			months = append(months, m)
		}
	}
	out := map[string]*ReviewEntry{}
	for i, m := range months {
		progress("history", i, len(months))
		c.yield()
		entries, err := c.store.ReadRevlog(m)
		if err != nil {
			return nil, err
		}
		for j, raw := range entries {
			if !bytes.Contains(raw, []byte(`"src":`)) {
				continue
			}
			var e ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return nil, fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, j, err)
			}
			if e.Source != "" {
				out[e.Source] = &e
			}
		}
	}
	progress("history", len(months), len(months))
	return out, nil
}

// importPresets gives the decks the import creates the presets the batch
// assigns them. Callers hold the write lock.
func (c *Collection) importPresets(b *ImportBatch, o *op, created map[string]*Deck, sum *ImportSummary) {
	if len(b.DeckPresets) == 0 {
		return
	}
	byKey := map[string]*ImportPreset{}
	for i := range b.Presets {
		byKey[b.Presets[i].Key] = &b.Presets[i]
	}
	deckPreset := map[string]*ImportPreset{}
	for name, key := range b.DeckPresets {
		if n, err := normalizeDeckName(name); err == nil && byKey[key] != nil {
			deckPreset[strings.ToLower(n)] = byKey[key]
		}
	}
	names := make([]string, 0, len(created))
	for name := range created {
		names = append(names, name)
	}
	sort.Strings(names)
	ids := map[*ImportPreset]ID{}
	for _, name := range names {
		ip := deckPreset[name]
		if ip == nil {
			continue
		}
		if _, ok := ids[ip]; !ok {
			ids[ip] = c.importPreset(ip, o, sum)
		}
		created[name].PresetID = ids[ip]
	}
}

// importPreset returns the preset for ip: an existing one with its name, or
// that name numbered, and its options, or a new one under the first free
// name of these.
func (c *Collection) importPreset(ip *ImportPreset, o *op, sum *ImportSummary) ID {
	name := strings.Join(strings.Fields(ip.Name), " ")
	if name == "" {
		name = "Imported"
	}
	opts := ip.Options
	if err := ValidateDeckOptions(opts); err != nil {
		sum.Warnings = append(sum.Warnings, usermsg.Text("importPresetOptions", "The options of preset %q could not be imported (%v); it has the default options.", name, err))
		opts = DefaultDeckOptions()
	}
	named := func(n string) *Preset {
		for _, p := range c.presets {
			if strings.EqualFold(p.Name, n) {
				return p
			}
		}
		for _, p := range o.Presets {
			if strings.EqualFold(p.Name, n) {
				return p
			}
		}
		return nil
	}
	for i := 1; ; i++ {
		n := name
		if i > 1 {
			n = fmt.Sprintf("%s (%d)", name, i)
		}
		p := named(n)
		if p == nil {
			p = &Preset{ID: c.ids.next(), Name: n, Options: opts}
			o.Presets = append(o.Presets, p)
			return p.ID
		}
		if p.Options.Equal(opts) {
			return p.ID
		}
	}
}

// importDescriptions gives the decks the import creates the descriptions
// the batch has for them.
func importDescriptions(b *ImportBatch, created map[string]*Deck) {
	for name, desc := range b.DeckDescriptions {
		n, err := normalizeDeckName(name)
		if err != nil {
			continue
		}
		if d := created[strings.ToLower(n)]; d != nil {
			d.Description = strings.TrimSpace(desc)
		}
	}
}

func deckCreated(created map[string]*Deck, id ID) bool {
	for _, d := range created {
		if d.ID == id {
			return true
		}
	}
	return false
}

func (c *Collection) uniqueNoteTypeName(name string, o *op) string {
	taken := func(n string) bool {
		for _, nt := range c.noteTypes {
			if strings.EqualFold(nt.Name, n) {
				return true
			}
		}
		for _, nt := range o.NoteTypes {
			if strings.EqualFold(nt.Name, n) {
				return true
			}
		}
		return false
	}
	if !taken(name) {
		return name
	}
	for i := 2; ; i++ {
		if n := fmt.Sprintf("%s (%d)", name, i); !taken(n) {
			return n
		}
	}
}

// padFields returns fields with empty values appended up to n.
func padFields(fields []string, n int) []string {
	if len(fields) >= n {
		return fields
	}
	return append(slices.Clone(fields), make([]string, n-len(fields))...)
}

func rewriteMedia(field string, renames map[string]string) string {
	for from, to := range renames {
		field = strings.ReplaceAll(field, `"`+from+`"`, `"`+to+`"`)
		field = strings.ReplaceAll(field, `'`+from+`'`, `'`+to+`'`)
	}
	return field
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

// HasMedia reports whether a media file with this name is stored.
func (c *Collection) HasMedia(name string) bool { return c.store.HasMedia(name) }

// Media returns a stored media file.
func (c *Collection) Media(name string) ([]byte, error) {
	data, err := c.store.GetMedia(name)
	if err != nil {
		return nil, userErr("mediaMissing", "Media file %q is missing.", name)
	}
	return data, nil
}
