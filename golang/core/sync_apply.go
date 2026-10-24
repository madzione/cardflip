package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"cardflip/golang/storage"
)

// Applying a delta (sync design decision 1): every register keeps the
// version with the greater stamp; equal stamps (possible only between
// collections without device slots) are decided by the greater JSON, so
// both sides agree. A tombstone deletes an object only if its stamp is
// greater than the object's, so an object changed after its deletion
// survives. Review log entries are combined. Deleting a note deletes its
// cards with it: card tombstones are only for cards deleted on their own,
// so that a note brought back by a later edit comes back with its cards.

// BatchSource describes where a batch comes from. Received are the local
// sequence numbers of the changes received earlier through the link, and
// LastSync the local sequence number when the link last finished a sync:
// when a media file has to make way for the peer's file of the same name,
// the notes changed since then that did not come through the link refer to
// the local file and are rewritten (decision 13).
type BatchSource struct {
	Link     string
	Received SeqRanges
	LastSync int64
}

// ApplyResult tells what a batch changed. First and Last are the local
// sequence numbers that the received changes got, 0 if nothing changed;
// they are not to be sent back to the peer they came from.
type ApplyResult struct {
	First, Last int64
	Changed     int
}

// opSyncFix is the kind of the local op that a batch can lead to: renamed
// media files and the notes that refer to them, and the clean-up after a
// merge (decision 5). Its changes are stamped as local changes, but it is
// not an undo step either.
const opSyncFix = "syncFix"

// merger builds the sync op of one batch. Callers hold the write lock.
type merger struct {
	c   *Collection
	src BatchSource
	o   *op
	seq int64
	// Objects the op puts or deletes, by ID.
	presets   map[ID]*Preset
	decks     map[ID]*Deck
	noteTypes map[ID]*NoteType
	notes     map[ID]*Note
	cards     map[ID]*Card
	deleted   map[metaKey]bool
	// reviews holds the review log entries the op adds, per month.
	reviews map[string]map[ID]bool
	// trash and restore are media files to move to and from the trash, and
	// renames the local files that make way for the peer's (old to new).
	trash, restore []string
	renames        map[string]string
	// resend gives local registers new sequence numbers, so that they are
	// sent to the peer again; it runs after the received changes got
	// theirs, which the peer must not get back.
	resend  []func()
	changed int
}

func (m *merger) next() int64 {
	m.seq++
	return m.seq
}

func (m *merger) meta(kind syncKind, id ID, st Stamp) {
	m.o.Meta = append(m.o.Meta, syncEntry{Kind: kind, ID: id, Stamp: st, Seq: m.next()})
}

func (m *merger) tomb(kind syncKind, id ID, st Stamp, aux int64) {
	m.o.Tombs = append(m.o.Tombs, syncEntry{Kind: kind, ID: id, Stamp: st, Seq: m.next(), Aux: aux})
	m.deleted[metaKey{kind, id}] = true
}

// wins reports whether an incoming register version beats the local one.
func wins(in, local Stamp, inData, localData func() []byte) bool {
	if in != local {
		return in > local
	}
	return bytes.Compare(inData(), localData()) > 0
}

func jsonOf(v any) func() []byte {
	return func() []byte {
		b, _ := json.Marshal(v)
		return b
	}
}

// ApplyBatch applies a page of a peer's delta as one op, which is not an
// undo step. The undo history is cleared only if the batch changed or
// deleted something a step in it changed (decision 14).
func (c *Collection) ApplyBatch(items []DeltaItem, src BatchSource) (ApplyResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := &merger{c: c, src: src, o: &op{Kind: opSync}, seq: c.seq,
		presets: map[ID]*Preset{}, decks: map[ID]*Deck{}, noteTypes: map[ID]*NoteType{}, notes: map[ID]*Note{},
		cards: map[ID]*Card{}, deleted: map[metaKey]bool{}, reviews: map[string]map[ID]bool{}, renames: map[string]string{}}
	var rev []storage.RevlogWrite
	for i := range items {
		it := &items[i]
		var err error
		switch it.Kind {
		case DeltaPresets:
			err = mergeObject(m, kindPreset, it, c.presets, m.presets, &m.o.Presets)
		case DeltaDecks:
			err = mergeObject(m, kindDeck, it, c.decks, m.decks, &m.o.Decks)
		case DeltaNoteTypes:
			err = mergeObject(m, kindNoteType, it, c.noteTypes, m.noteTypes, &m.o.NoteTypes)
		case DeltaNotes:
			err = mergeObject(m, kindNote, it, c.notes, m.notes, &m.o.Notes)
		case DeltaCards:
			err = m.card(it)
		case DeltaReviews:
			var w *storage.RevlogWrite
			if w, err = m.review(it); w != nil {
				rev = append(rev, *w)
			}
		case DeltaSettings:
			err = m.settings(it)
		case DeltaTombstones:
			var w *storage.RevlogWrite
			if w, err = m.tombstone(it); w != nil {
				rev = append(rev, *w)
			}
		case DeltaMedia:
			err = m.media(it)
		default:
			err = fmt.Errorf("unknown delta item kind %d", it.Kind)
		}
		if err != nil {
			return ApplyResult{}, fmt.Errorf("sync item %d (kind %d, ID %d): %w", i, it.Kind, it.ID, err)
		}
	}
	before, received := c.seq, m.seq
	for _, f := range m.resend {
		f()
	}
	if err := m.files(); err != nil {
		return ApplyResult{}, err
	}
	if err := c.commit(m.o, rev...); err != nil {
		return ApplyResult{}, err
	}
	res := ApplyResult{Changed: m.changed}
	if received > before {
		res.First, res.Last = before+1, received
	}
	keys := opKeys(m.o)
	fixes, err := c.afterMerge(m)
	for _, fix := range fixes {
		maps.Copy(keys, opKeys(fix))
	}
	if err != nil {
		return res, err
	}
	if c.hist.touches(keys) {
		c.hist.clear()
	}
	return res, nil
}

// mergeObject merges an incoming preset, deck, note type or note.
func mergeObject[T any](m *merger, kind syncKind, it *DeltaItem, cur, put map[ID]*T, out *[]*T) error {
	k := metaKey{kind, it.ID}
	local, exists := put[it.ID]
	if !exists {
		local, exists = cur[it.ID]
		exists = exists && !m.deleted[k]
	}
	if exists {
		lm := m.c.meta[k]
		if !wins(it.Stamp, lm.stamp, func() []byte { return it.Data }, jsonOf(local)) {
			return nil
		}
	} else if t, ok := m.c.tombs[k]; ok && t.stamp > it.Stamp {
		return nil
	}
	v := new(T)
	if err := json.Unmarshal(it.Data, v); err != nil {
		return err
	}
	put[it.ID] = v
	*out = append(*out, v)
	m.meta(kind, it.ID, it.Stamp)
	delete(m.deleted, k)
	m.changed++
	return nil
}

// card merges an incoming card's two registers separately.
func (m *merger) card(it *DeltaItem) error {
	in := new(Card)
	if err := json.Unmarshal(it.Data, in); err != nil {
		return err
	}
	local := m.cards[it.ID]
	if local == nil && !m.deleted[metaKey{kindCard, it.ID}] {
		local = m.c.cards[it.ID]
	}
	if local == nil {
		if !m.noteExists(in.NoteID) {
			// Its note was deleted, which deletes its cards.
			return nil
		}
		if t, ok := m.c.tombs[metaKey{kindCard, it.ID}]; ok && t.stamp > max(it.Stamp, it.Stamp2) {
			return nil
		}
		m.putCard(in)
		m.meta(kindCardSched, it.ID, it.Stamp)
		m.meta(kindCardPlace, it.ID, it.Stamp2)
		return nil
	}
	ms, mp := m.c.meta[metaKey{kindCardSched, it.ID}], m.c.meta[metaKey{kindCardPlace, it.ID}]
	takeS := wins(it.Stamp, ms.stamp, jsonOf(schedPart(in)), jsonOf(schedPart(local)))
	takeP := wins(it.Stamp2, mp.stamp, jsonOf(placePart(in)), jsonOf(placePart(local)))
	if !takeS && !takeP {
		return nil
	}
	merged := *local
	if takeS {
		merged.CardState, merged.NewPos, merged.OriginalPos = in.CardState, in.NewPos, in.OriginalPos
		merged.BuriedUntil, merged.PreviewDue = in.BuriedUntil, in.PreviewDue
		m.meta(kindCardSched, it.ID, it.Stamp)
	}
	if takeP {
		merged.NoteID, merged.Ord, merged.DeckID, merged.HomeDeckID = in.NoteID, in.Ord, in.DeckID, in.HomeDeckID
		merged.FilteredPos, merged.Flag, merged.Suspended = in.FilteredPos, in.Flag, in.Suspended
		m.meta(kindCardPlace, it.ID, it.Stamp2)
	}
	// The modification time goes with the newest register (undo can make
	// it go back).
	if max(it.Stamp, it.Stamp2) > max(ms.stamp, mp.stamp) {
		merged.Modified = in.Modified
	}
	m.putCard(&merged)
	return nil
}

func (m *merger) putCard(card *Card) {
	if _, ok := m.cards[card.ID]; !ok {
		m.o.Cards = append(m.o.Cards, card)
	} else {
		for i, c := range m.o.Cards {
			if c.ID == card.ID {
				m.o.Cards[i] = card
			}
		}
	}
	m.cards[card.ID] = card
	m.changed++
}

// noteExists reports whether a note exists after what the op does so far.
func (m *merger) noteExists(id ID) bool {
	if m.deleted[metaKey{kindNote, id}] {
		return false
	}
	_, a := m.notes[id]
	_, b := m.c.notes[id]
	return a || b
}

func schedPart(c *Card) any {
	return []any{c.CardState, c.NewPos, c.OriginalPos, c.BuriedUntil, c.PreviewDue}
}

func placePart(c *Card) any {
	return []any{c.NoteID, c.Ord, c.DeckID, c.HomeDeckID, c.FilteredPos, c.Flag, c.Suspended}
}

// review adds an incoming review log entry unless it is stored, or was
// deleted later than it was added.
func (m *merger) review(it *DeltaItem) (*storage.RevlogWrite, error) {
	k := metaKey{kindReview, it.ID}
	if t, ok := m.c.tombs[k]; ok && t.stamp > it.Stamp {
		return nil, nil
	}
	var e ReviewEntry
	if err := json.Unmarshal(it.Data, &e); err != nil {
		return nil, err
	}
	if e.ID != it.ID || e.Time.IsZero() {
		return nil, errors.New("invalid review log entry")
	}
	month := monthKey(e.Time)
	if m.reviews[month][e.ID] {
		return nil, nil
	}
	has, err := m.c.store.RevlogHas(month, []int64{int64(e.ID)})
	if err != nil {
		return nil, err
	}
	if has[int64(e.ID)] && !m.deleted[k] {
		if lm := m.c.meta[k]; it.Stamp > lm.stamp {
			m.meta(kindReview, it.ID, it.Stamp)
		}
		return nil, nil
	}
	if m.reviews[month] == nil {
		m.reviews[month] = map[ID]bool{}
	}
	m.reviews[month][e.ID] = true
	m.o.AddReviews = append(m.o.AddReviews, &e)
	if it.Stamp != 0 {
		m.meta(kindReview, it.ID, it.Stamp)
	}
	delete(m.deleted, k)
	seq := m.next()
	if m.o.RevSeq == nil {
		m.o.RevSeq = map[string]int64{}
	}
	m.o.RevSeq[month] = seq
	m.changed++
	return &storage.RevlogWrite{Month: month, ID: int64(e.ID), Entry: it.Data, Seq: seq}, nil
}

// settings merges the synced settings; the device's own stay.
func (m *merger) settings(it *DeltaItem) error {
	var in Settings
	if err := json.Unmarshal(it.Data, &in); err != nil {
		return err
	}
	cur := m.c.settings
	if m.o.Settings != nil {
		cur = *m.o.Settings
	}
	lm := m.c.meta[metaKey{kindSettings, 0}]
	if !wins(it.Stamp, lm.stamp, jsonOf(withSyncedSettings(Settings{}, in)), jsonOf(withSyncedSettings(Settings{}, cur))) {
		return nil
	}
	s := withSyncedSettings(cur, in)
	m.o.Settings = &s
	m.meta(kindSettings, 0, it.Stamp)
	m.changed++
	return nil
}

// tombstone applies an incoming tombstone. It returns the review log write
// that deletes an entry.
func (m *merger) tombstone(it *DeltaItem) (*storage.RevlogWrite, error) {
	kind := syncKind(it.Of)
	k := metaKey{kind, it.ID}
	if t, ok := m.c.tombs[k]; ok && t.stamp >= it.Stamp && !m.deleted[k] {
		return nil, nil
	}
	switch kind {
	case kindPreset:
		deleteObject(m, k, it, m.presets, m.c.presets, &m.o.DelPresets)
	case kindDeck:
		deleteObject(m, k, it, m.decks, m.c.decks, &m.o.DelDecks)
	case kindNoteType:
		deleteObject(m, k, it, m.noteTypes, m.c.noteTypes, &m.o.DelNoteTypes)
	case kindNote:
		if deleteObject(m, k, it, m.notes, m.c.notes, &m.o.DelNotes) {
			m.deleteCardsOf(it.ID)
		}
	case kindCard:
		card := m.cards[it.ID]
		if card == nil {
			card = m.c.cards[it.ID]
		}
		if card != nil && !m.deleted[k] {
			s := max(m.c.meta[metaKey{kindCardSched, it.ID}].stamp, m.c.meta[metaKey{kindCardPlace, it.ID}].stamp)
			if _, put := m.cards[it.ID]; put {
				s = 0
				for _, e := range m.o.Meta {
					if (e.Kind == kindCardSched || e.Kind == kindCardPlace) && e.ID == it.ID {
						s = max(s, e.Stamp)
					}
				}
			}
			if it.Stamp <= s {
				return nil, nil
			}
			m.o.DelCards = append(m.o.DelCards, it.ID)
			m.changed++
		}
		m.tomb(kind, it.ID, it.Stamp, it.Aux)
	case kindReview:
		return m.deleteReview(it)
	case kindSettings, kindCardSched, kindCardPlace:
		return nil, fmt.Errorf("invalid tombstone of kind %d", kind)
	default:
		return nil, fmt.Errorf("unknown tombstone kind %d", kind)
	}
	return nil, nil
}

// deleteObject applies the tombstone of a preset, deck, note type or note:
// it deletes the object unless it changed later, and keeps the tombstone
// unless the object survives. It reports whether it deleted the object.
func deleteObject[T any](m *merger, k metaKey, it *DeltaItem, put, cur map[ID]*T, del *[]ID) bool {
	_, inOp := put[k.id]
	_, exists := cur[k.id]
	if inOp || exists && !m.deleted[k] {
		st := m.c.meta[k].stamp
		if inOp {
			for _, e := range m.o.Meta {
				if e.key() == k {
					st = e.Stamp
				}
			}
		}
		if it.Stamp <= st {
			return false
		}
		*del = append(*del, k.id)
		m.tomb(k.kind, k.id, it.Stamp, it.Aux)
		m.changed++
		return true
	}
	m.tomb(k.kind, k.id, it.Stamp, it.Aux)
	return false
}

// deleteCardsOf deletes the cards of a deleted note, without tombstones.
func (m *merger) deleteCardsOf(note ID) {
	ids := slices.Clone(m.c.cardsByNote[note])
	for id, card := range m.cards {
		if card.NoteID == note {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	for _, id := range slices.Compact(ids) {
		if !slices.Contains(m.o.DelCards, id) {
			m.o.DelCards = append(m.o.DelCards, id)
		}
	}
}

// deleteReview applies a review log entry's tombstone.
func (m *merger) deleteReview(it *DeltaItem) (*storage.RevlogWrite, error) {
	k := metaKey{kindReview, it.ID}
	if lm, ok := m.c.meta[k]; ok && lm.stamp >= it.Stamp {
		return nil, nil
	}
	m.tomb(kindReview, it.ID, it.Stamp, it.Aux)
	if it.Aux == 0 {
		return nil, nil
	}
	month := monthName(it.Aux)
	if m.reviews[month][it.ID] {
		// Added by this batch: drop it again.
		delete(m.reviews[month], it.ID)
		m.o.AddReviews = slices.DeleteFunc(m.o.AddReviews, func(e *ReviewEntry) bool { return e.ID == it.ID })
		return &storage.RevlogWrite{Month: month, ID: int64(it.ID)}, nil
	}
	recs, err := m.c.store.ReadRevlogRecords(month)
	if err != nil {
		return nil, err
	}
	for _, r := range recs {
		if r.ID != int64(it.ID) {
			continue
		}
		var e ReviewEntry
		if err := json.Unmarshal(r.Entry, &e); err != nil {
			return nil, fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, month, r.ID, err)
		}
		m.o.DelReviews = append(m.o.DelReviews, &e)
		m.changed++
		return &storage.RevlogWrite{Month: month, ID: r.ID}, nil
	}
	return nil, nil
}

// media merges a media manifest entry (decision 13). Different content
// under one name is kept by both: the entry with the smaller stamp keeps
// the name, and the device whose file it is not renames its own.
func (m *merger) media(it *DeltaItem) error {
	c := m.c
	if c.media == nil {
		// The manifest is built before a link first syncs.
		return nil
	}
	name := cleanMediaName(it.Name)
	if name != it.Name || name == "" || len(it.SHA1) != 40 {
		return fmt.Errorf("invalid media name %q", it.Name)
	}
	in := mediaChange{Name: name, SHA1: it.SHA1, Trashed: it.Trashed, Stamp: it.Stamp}
	local, ok := c.media[name]
	switch {
	case !ok:
		in.Missing = !in.Trashed
	case local.SHA1 == in.SHA1:
		if !wins(in.Stamp, local.Stamp, jsonOf(in.Trashed), jsonOf(local.Trashed)) {
			return nil
		}
		in.Missing = local.Missing
		if in.Trashed && !local.Trashed {
			m.trash = append(m.trash, name)
		} else if !in.Trashed && local.Trashed {
			m.restore = append(m.restore, name)
			in.Missing = true
		}
	case in.Stamp < local.Stamp || in.Stamp == local.Stamp && in.SHA1 < local.SHA1:
		// The peer's file keeps the name; the local one makes way.
		stem, ext := splitMediaName(name)
		m.renames[name] = stem + "-" + local.SHA1[:8] + ext
		in.Missing = !in.Trashed
	default:
		// The local file keeps the name: send its entry again, so the peer
		// renames its own.
		lc := mediaChange{Name: name, SHA1: local.SHA1, Trashed: local.Trashed, Stamp: local.Stamp, Missing: local.Missing}
		m.resend = append(m.resend, func() {
			lc.Seq = m.next()
			m.o.Media = append(m.o.Media, lc)
		})
		return nil
	}
	in.Seq = m.next()
	m.o.Media = append(m.o.Media, in)
	m.changed++
	return nil
}

// files moves the media files that make way for the peer's to their new
// names, before the op that changes the manifest is saved.
func (m *merger) files() error {
	for _, from := range slices.Sorted(maps.Keys(m.renames)) {
		data, err := m.c.mediaBlob(from)
		if errors.Is(err, storage.ErrNotFound) {
			delete(m.renames, from)
			continue
		}
		if err != nil {
			return err
		}
		if err := m.c.store.PutMedia(m.renames[from], data); err != nil {
			return fmt.Errorf("could not save media file %s: %w", m.renames[from], err)
		}
	}
	return nil
}

// afterMerge does what a committed batch leads to: it moves media files to
// and from the trash, renames the local files that made way (with the notes
// that refer to them) and cleans up after the merge (decision 5). It
// returns the local ops it committed.
func (c *Collection) afterMerge(m *merger) ([]*op, error) {
	if len(m.trash) > 0 {
		if err := c.store.TrashMedia(m.trash, c.now()); err != nil {
			return nil, err
		}
		c.mediaIndex = nil
	}
	if len(m.restore) > 0 {
		have := &op{Kind: opSync}
		for _, name := range m.restore {
			data, err := c.store.TrashedMediaData(name)
			if err != nil {
				continue // fetched from the peer instead
			}
			e := c.media[name]
			if sha1Hex(data) != e.SHA1 {
				continue
			}
			if err := c.store.PutMedia(name, data); err != nil {
				return nil, err
			}
			have.Media = append(have.Media, mediaChange{Name: name, SHA1: e.SHA1, Stamp: e.Stamp, Seq: e.Seq})
		}
		if err := c.commit(have); err != nil {
			return nil, err
		}
	}
	fix := &op{Kind: opSyncFix}
	if len(m.renames) > 0 {
		for _, from := range slices.Sorted(maps.Keys(m.renames)) {
			to := m.renames[from]
			data, err := c.store.GetMedia(to)
			if err != nil {
				return nil, err
			}
			c.mediaChanged(fix, to, sha1Hex(data), false)
		}
		for _, n := range c.notesToRename(m) {
			nn := cloneNote(n)
			for i, f := range nn.Fields {
				nn.Fields[i] = rewriteMedia(f, m.renames)
			}
			nn.Modified = c.now()
			fix.Notes = append(fix.Notes, &nn)
		}
	}
	var done []*op
	if !fix.empty() {
		if err := c.commit(fix); err != nil {
			return done, err
		}
		done = append(done, fix)
	}
	cleanup, err := c.syncCleanup(m.o)
	return append(done, cleanup...), err
}

// notesToRename returns the notes that refer to a local media file that
// made way for the peer's: those changed since the link's last sync that
// did not come through it. Callers hold the write lock.
func (c *Collection) notesToRename(m *merger) []*Note {
	var out []*Note
	for _, id := range sortedIDs(c.notes) {
		mt := c.meta[metaKey{kindNote, id}]
		if mt.seq <= m.src.LastSync || m.src.Received.Contains(mt.seq) || received(m.o, kindNote, id) {
			continue
		}
		n := c.notes[id]
		for _, ref := range MediaNames(n.Fields...) {
			if _, ok := m.renames[ref]; ok {
				out = append(out, n)
				break
			}
		}
	}
	return out
}

// received reports whether op o set the register (kind, id).
func received(o *op, kind syncKind, id ID) bool {
	for _, e := range o.Meta {
		if e.Kind == kind && e.ID == id {
			return true
		}
	}
	return false
}

// mediaBlob returns a media file's content from the media or, failing
// that, the media trash.
func (c *Collection) mediaBlob(name string) ([]byte, error) {
	data, err := c.store.GetMedia(name)
	if errors.Is(err, storage.ErrNotFound) {
		return c.store.TrashedMediaData(name)
	}
	return data, err
}

// MediaBlob returns a media file a peer asks for, from the media or the
// media trash, if the manifest lists it with the content sha (SHA-1, hex).
func (c *Collection) MediaBlob(name, sha string) ([]byte, error) {
	c.mu.RLock()
	e, ok := c.media[name]
	c.mu.RUnlock()
	if !ok || e.Missing || e.SHA1 != sha {
		return nil, storage.ErrNotFound
	}
	data, err := c.mediaBlob(name)
	if err != nil {
		return nil, err
	}
	if sha1Hex(data) != e.SHA1 {
		return nil, storage.ErrNotFound
	}
	return data, nil
}

// MediaFileRef names a media file and its content's SHA-1 (hex).
type MediaFileRef struct {
	Name string `json:"name"`
	SHA1 string `json:"sha1"`
}

// MissingMedia lists the media files the manifest has that are not stored
// yet: received entries whose files are still to be fetched.
func (c *Collection) MissingMedia() []MediaFileRef {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []MediaFileRef
	for name, e := range c.media {
		if e.Missing && !e.Trashed {
			out = append(out, MediaFileRef{Name: name, SHA1: e.SHA1})
		}
	}
	slices.SortFunc(out, func(a, b MediaFileRef) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// PutSyncedMedia stores a media file fetched from a peer, if it is the
// content the manifest lists for the name; otherwise (the manifest changed
// meanwhile) it does nothing.
func (c *Collection) PutSyncedMedia(name string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.media[name]
	if !ok || !e.Missing || sha1Hex(data) != e.SHA1 {
		return nil
	}
	if err := c.store.PutMedia(name, data); err != nil {
		return err
	}
	c.mediaIndex = nil
	return c.commit(&op{Kind: opSync, Media: []mediaChange{{Name: name, SHA1: e.SHA1, Trashed: e.Trashed,
		Stamp: e.Stamp, Seq: e.Seq}}})
}

// opKeys returns the registers an op changes or deletes, and the review
// log entries it adds or removes.
func opKeys(o *op) map[metaKey]bool {
	keys := map[metaKey]bool{}
	if o == nil {
		return keys
	}
	add := func(kind syncKind, ids ...ID) {
		for _, id := range ids {
			keys[metaKey{kind, id}] = true
		}
	}
	for _, p := range o.Presets {
		add(kindPreset, p.ID)
	}
	for _, d := range o.Decks {
		add(kindDeck, d.ID)
	}
	for _, nt := range o.NoteTypes {
		add(kindNoteType, nt.ID)
	}
	for _, n := range o.Notes {
		add(kindNote, n.ID)
	}
	for _, card := range o.Cards {
		add(kindCard, card.ID)
	}
	add(kindPreset, o.DelPresets...)
	add(kindDeck, o.DelDecks...)
	add(kindNoteType, o.DelNoteTypes...)
	add(kindNote, o.DelNotes...)
	add(kindCard, o.DelCards...)
	if o.Review != nil {
		add(kindReview, o.Review.ID)
	}
	if o.ReviewUndone != 0 {
		add(kindReview, o.ReviewUndone)
	}
	for _, e := range append(slices.Clone(o.AddReviews), o.DelReviews...) {
		add(kindReview, e.ID)
	}
	return keys
}

// touches reports whether a step of the history changed one of keys.
func (h *history) touches(keys map[metaKey]bool) bool {
	if len(keys) == 0 {
		return false
	}
	for _, stack := range [][]*step{h.undo, h.redo} {
		for _, s := range stack {
			for k := range opKeys(s.revert) {
				if keys[k] {
					return true
				}
			}
			if s.review != 0 && keys[metaKey{kindReview, s.review}] {
				return true
			}
		}
	}
	return false
}
