package core

import (
	"cmp"
	"crypto/rand"
	"encoding/base64"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"

	"cardflip/golang/internal/usermsg"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// NormalizeTags splits tag text on whitespace and drops duplicates (ignoring
// case, keeping the first spelling): "verb irregular" is two tags.
func NormalizeTags(input []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, s := range input {
		for _, tag := range strings.Fields(nfc(s)) {
			key := strings.ToLower(tag)
			if !seen[key] {
				seen[key] = true
				out = append(out, tag)
			}
		}
	}
	return out
}

func newGUID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func fieldMap(nt *NoteType, fields []string) map[string]string {
	m := make(map[string]string, len(nt.Fields))
	for i, name := range nt.Fields {
		if i < len(fields) {
			m[name] = fields[i]
		} else {
			m[name] = ""
		}
	}
	return m
}

// cardOrds returns the ords a note of type nt must have cards for: the
// templates whose question uses a non-empty field, or for a cloze note type
// one per cloze number (ord = number - 1).
func cardOrds(nt *NoteType, fields []string) []int {
	fm := fieldMap(nt, fields)
	var ords []int
	if nt.Kind == Cloze {
		if len(nt.Templates) == 0 {
			return nil
		}
		for _, n := range render.ClozeCards(nt.Templates[0].QFmt, fm) {
			ords = append(ords, n-1)
		}
		return ords
	}
	for i, t := range nt.Templates {
		if render.HasContent(t.QFmt, render.Note{Fields: fm}, 0) {
			ords = append(ords, i)
		}
	}
	return ords
}

// noCardsError explains why a note would produce no cards.
func noCardsError(nt *NoteType) error {
	if nt.Kind == Cloze {
		return userErr("noCardsCloze", "This would produce no cards. Add a cloze deletion such as {{c1::text}}.")
	}
	var used []string
	seen := map[string]bool{}
	for _, t := range nt.Templates {
		for _, f := range render.FieldsUsed(t.QFmt) {
			if !seen[f] && contains(nt.Fields, f) {
				seen[f] = true
				used = append(used, f)
			}
		}
	}
	if len(used) == 0 {
		return userErr("noCardsNoField", "This would produce no cards: no card's question uses a field.")
	}
	if len(used) == 1 {
		return userErr("noCardsFillInField", "This would produce no cards. Fill in the %s field.", used[0])
	}
	last := len(used) - 1
	return &Error{Msg: usermsg.Tag("noCardsFillInFields", []string{strings.Join(used[:last], ", "), used[last]},
		"This would produce no cards. Fill in "+joinOr(used)+".")}
}

func joinOr(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = "the " + n + " field"
	}
	return strings.Join(q[:len(q)-1], ", ") + " or " + q[len(q)-1]
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// syncCards adds to o the card changes that give note n (of type nt) a card
// for each of ords: existing cards keep their scheduling and missing ones are
// created New in deckID, or in their card type's deck override. Cards no longer generated are kept as empty cards,
// for Empty Cards to remove. remap, when set, maps current card ords to new
// ones (-1 deletes) for template removal. It returns how many cards the note
// has afterwards.
func (c *Collection) syncCards(o *op, nt *NoteType, n *Note, ords []int, deckID ID, remap func(int) int) int {
	ch := c.cardChanges(n.ID, ords, remap)
	o.DelCards = append(o.DelCards, ch.del...)
	o.Cards = append(o.Cards, ch.moved...)
	for _, ord := range ch.create {
		id := c.ids.next()
		o.Cards = append(o.Cards, &Card{
			ID: id, NoteID: n.ID, DeckID: c.newCardDeck(nt, ord, deckID), Ord: ord, NewPos: c.newPos(),
			CardState: scheduler.CardState{State: scheduler.New},
		})
	}
	return ch.count
}

// newCardDeck returns the deck a new card of card type ord goes to: the
// card type's deck override while that deck exists and is not filtered,
// else deckID.
func (c *Collection) newCardDeck(nt *NoteType, ord int, deckID ID) ID {
	if t := nt.cardType(ord); t != nil && t.DeckOverride != 0 {
		if d, ok := c.decks[t.DeckOverride]; ok && d.Filter == nil {
			return t.DeckOverride
		}
	}
	return deckID
}

// cardSync is what syncCards does to a note's cards.
type cardSync struct {
	del    []ID    // cards deleted
	moved  []*Card // changed copies of cards whose ord changes
	create []int   // ords that get a new card
	count  int     // cards the note has afterwards
}

// cardChanges works out syncCards without changing anything, so previews
// can count with a read lock.
func (c *Collection) cardChanges(nid ID, ords []int, remap func(int) int) cardSync {
	var ch cardSync
	have := map[int]bool{}
	for _, cid := range c.cardsByNote[nid] {
		card := c.cards[cid]
		ord := card.Ord
		if remap != nil {
			ord = remap(ord)
		}
		if ord < 0 || have[ord] {
			ch.del = append(ch.del, cid)
			continue
		}
		have[ord] = true
		if ord != card.Ord {
			cc := *card
			cc.Ord = ord
			ch.moved = append(ch.moved, &cc)
		}
	}
	for _, ord := range ords {
		if !have[ord] {
			have[ord] = true
			ch.create = append(ch.create, ord)
		}
	}
	ch.count = len(have)
	return ch
}

// deckOfNote returns the deck new cards of an existing note go to: the deck
// of its first card, else the last used deck, else any deck.
func (c *Collection) deckOfNote(nid ID) ID {
	ids := c.cardsByNote[nid]
	if len(ids) > 0 {
		first := c.cards[ids[0]]
		for _, id := range ids {
			if c.cards[id].Ord < first.Ord {
				first = c.cards[id]
			}
		}
		return first.HomeDeck()
	}
	return c.fallbackDeck()
}

func (c *Collection) fallbackDeck() ID {
	if d, ok := c.decks[c.settings.LastDeckID]; ok && d.Filter == nil {
		return c.settings.LastDeckID
	}
	var best ID
	for id, d := range c.decks {
		if d.Filter == nil && (best == 0 || id < best) {
			best = id
		}
	}
	return best
}

// NoteInput is a note to add.
type NoteInput struct {
	TypeID ID
	DeckID ID
	Fields []string
	Tags   []string
}

func (c *Collection) checkFields(nt *NoteType, fields []string) ([]string, error) {
	if len(fields) > len(nt.Fields) {
		return nil, userErr("noteFieldCount", "The note has %d fields, but note type %q has %d.", len(fields), nt.Name, len(nt.Fields))
	}
	out := make([]string, len(nt.Fields))
	copy(out, fields)
	for i, f := range out {
		out[i] = nfc(f)
	}
	return out, nil
}

// AddNote adds a note and its cards. The note type and deck are remembered
// for the next add, and the deck also for the note type.
func (c *Collection) AddNote(in NoteInput) (ID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p, err := c.prepareNote(in)
	if err != nil {
		return 0, err
	}
	o := &op{Kind: "addNote"}
	n := c.buildNote(o, p)
	c.rememberDefaults(o, p.nt.ID, in.DeckID)
	if err := c.commit(o); err != nil {
		return 0, err
	}
	return n.ID, nil
}

// preparedNote is a checked NoteInput, ready for buildNote.
type preparedNote struct {
	in     NoteInput
	nt     *NoteType
	fields []string
	ords   []int
}

// prepareNote checks a note to add without changing anything. Callers hold
// the write lock.
func (c *Collection) prepareNote(in NoteInput) (preparedNote, error) {
	nt, ok := c.noteTypes[in.TypeID]
	if !ok {
		return preparedNote{}, userErr("noteTypeGone", "That note type no longer exists.")
	}
	if d, ok := c.decks[in.DeckID]; !ok {
		return preparedNote{}, userErr("deckGone", "That deck no longer exists.")
	} else if d.Filter != nil {
		return preparedNote{}, errFilteredTarget(d)
	}
	fields, err := c.checkFields(nt, in.Fields)
	if err != nil {
		return preparedNote{}, err
	}
	ords := cardOrds(nt, fields)
	if len(ords) == 0 {
		return preparedNote{}, noCardsError(nt)
	}
	return preparedNote{in: in, nt: nt, fields: fields, ords: ords}, nil
}

// buildNote adds a prepared note and its new cards to o.
func (c *Collection) buildNote(o *op, p preparedNote) *Note {
	n := &Note{ID: c.ids.next(), TypeID: p.nt.ID, GUID: newGUID(), Fields: p.fields, Tags: NormalizeTags(p.in.Tags), Modified: c.now()}
	o.Notes = append(o.Notes, n)
	c.syncCards(o, p.nt, n, p.ords, p.in.DeckID, nil)
	return n
}

// rememberDefaults makes o remember note type typeID and deck deckID as the
// Add screen's defaults, unless they already are.
func (c *Collection) rememberDefaults(o *op, typeID, deckID ID) {
	if c.settings.LastNoteTypeID == typeID && c.settings.LastDeckID == deckID &&
		c.settings.LastDeckByNoteType[typeID] == deckID {
		return
	}
	s := c.settings
	s.LastNoteTypeID, s.LastDeckID = typeID, deckID
	s.LastDeckByNoteType = maps.Clone(s.LastDeckByNoteType)
	if s.LastDeckByNoteType == nil {
		s.LastDeckByNoteType = map[ID]ID{}
	}
	s.LastDeckByNoteType[typeID] = deckID
	o.Settings = &s
}

// AddNotesOptions are the options of AddNotes.
type AddNotesOptions struct {
	// KeepDefaults leaves the Add screen's remembered note type and deck
	// alone.
	KeepDefaults bool
	// SkipDuplicates skips notes whose first field matches a note of the
	// same note type, existing or earlier in the batch, as the duplicate
	// warning compares them.
	SkipDuplicates bool
}

// AddResult is what AddNotes did with one note: added with its cards, or
// skipped as a duplicate.
type AddResult struct {
	NoteID ID   // the added note, 0 when skipped
	Cards  []ID // its cards, by ord
	// DuplicateOf is the oldest note the skipped note duplicates (an earlier
	// note of the batch when no existing note matches).
	DuplicateOf ID
}

// AddNotesError is the error of AddNotes for a note that cannot be added.
type AddNotesError struct {
	Index int // in the input, from 0
	Err   error
}

func (e *AddNotesError) Error() string { return e.Err.Error() }
func (e *AddNotesError) Unwrap() error { return e.Err }

// AddNotes adds notes and their cards as one undo step, "Add N notes". Every
// note is checked first: when one cannot be added, nothing is and the error
// is an *AddNotesError. The results are in the input's order.
func (c *Collection) AddNotes(in []NoteInput, opts AddNotesOptions) ([]AddResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	prepared := make([]preparedNote, len(in))
	for i, n := range in {
		p, err := c.prepareNote(n)
		if err != nil {
			return nil, &AddNotesError{Index: i, Err: err}
		}
		prepared[i] = p
	}
	o := &op{Kind: "addNotes"}
	out := make([]AddResult, len(in))
	added := map[dupKey]ID{}
	var last *preparedNote
	for i, p := range prepared {
		key := dupKey{p.nt.ID, duplicateText(p.fields)}
		if opts.SkipDuplicates && key.first != "" {
			if ids := slices.Clone(c.firstFields[key]); len(ids) > 0 {
				sortIDs(ids)
				out[i].DuplicateOf = ids[0]
				continue
			}
			if id, ok := added[key]; ok {
				out[i].DuplicateOf = id
				continue
			}
		}
		from := len(o.Cards)
		n := c.buildNote(o, p)
		if _, ok := added[key]; !ok {
			added[key] = n.ID
		}
		out[i].NoteID = n.ID
		cards := o.Cards[from:]
		sorted := slices.Clone(cards)
		slices.SortFunc(sorted, func(a, b *Card) int { return cmp.Compare(a.Ord, b.Ord) })
		for _, card := range sorted {
			out[i].Cards = append(out[i].Cards, card.ID)
		}
		last = &prepared[i]
	}
	if last != nil && !opts.KeepDefaults {
		c.rememberDefaults(o, last.nt.ID, last.in.DeckID)
	}
	if err := c.commit(o); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateNote changes a note's fields and tags and regenerates its cards:
// existing cards keep their scheduling, new ones are added, and ones no
// longer generated are kept as empty cards.
func (c *Collection) UpdateNote(id ID, fields []string, tags []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.notes[id]
	if !ok {
		return userErr("noteGone", "That note no longer exists.")
	}
	nt := c.noteTypes[old.TypeID]
	fields, err := c.checkFields(nt, fields)
	if err != nil {
		return err
	}
	ords := cardOrds(nt, fields)
	if len(ords) == 0 {
		return noCardsError(nt)
	}
	n := *old
	n.Fields = fields
	n.Tags = NormalizeTags(tags)
	n.Modified = c.now()
	o := &op{Kind: "updateNote", Notes: []*Note{&n}}
	c.syncCards(o, nt, &n, ords, c.deckOfNote(id), nil)
	return c.commit(o)
}

// Note returns a note.
func (c *Collection) Note(id ID) (Note, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n, ok := c.notes[id]
	if !ok {
		return Note{}, userErr("noteGone", "That note no longer exists.")
	}
	return cloneNote(n), nil
}

func cloneNote(n *Note) Note {
	cp := *n
	cp.Fields = append([]string(nil), n.Fields...)
	cp.Tags = append([]string(nil), n.Tags...)
	return cp
}

// CardsOfNote returns a note's cards ordered by ord.
func (c *Collection) CardsOfNote(id ID) []Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []Card
	for _, cid := range c.cardsByNote[id] {
		out = append(out, *c.cards[cid])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ord < out[j].Ord })
	return out
}

// DeleteNotes deletes notes and all their cards.
func (c *Collection) DeleteNotes(ids []ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	o := &op{Kind: "deleteNotes"}
	for _, id := range ids {
		if _, ok := c.notes[id]; !ok {
			continue
		}
		o.DelNotes = append(o.DelNotes, id)
		o.DelCards = append(o.DelCards, c.cardsByNote[id]...)
	}
	return c.commit(o)
}

// AllTags lists the tags in use, sorted ignoring case. Tags that differ only
// in case are listed once.
func (c *Collection) AllTags() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids := make([]ID, 0, len(c.notes))
	for id := range c.notes {
		ids = append(ids, id)
	}
	sortIDs(ids)
	// The oldest note's spelling wins.
	seen := map[string]string{}
	for _, id := range ids {
		for _, t := range c.notes[id].Tags {
			if _, ok := seen[strings.ToLower(t)]; !ok {
				seen[strings.ToLower(t)] = t
			}
		}
	}
	out := make([]string, 0, len(seen))
	for _, t := range seen {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}

// AddTags adds tags to notes.
func (c *Collection) AddTags(noteIDs []ID, tags []string) error {
	add := NormalizeTags(tags)
	return c.editTags("addTags", noteIDs, func(have []string) []string {
		return NormalizeTags(append(append([]string(nil), have...), add...))
	})
}

// RemoveTags removes tags (ignoring case) from notes.
func (c *Collection) RemoveTags(noteIDs []ID, tags []string) error {
	drop := map[string]bool{}
	for _, t := range NormalizeTags(tags) {
		drop[strings.ToLower(t)] = true
	}
	return c.editTags("removeTags", noteIDs, func(have []string) []string {
		var out []string
		for _, t := range have {
			if !drop[strings.ToLower(t)] {
				out = append(out, t)
			}
		}
		return out
	})
}

// MarkedTag is the tag that marks a note, as in Anki. It is an ordinary tag.
const MarkedTag = "marked"

// isMarked reports whether tags contain MarkedTag (ignoring case).
func isMarked(tags []string) bool {
	return slices.ContainsFunc(tags, func(t string) bool { return strings.EqualFold(t, MarkedTag) })
}

// IsMarked reports whether the note has the tag MarkedTag.
func (c *Collection) IsMarked(noteID ID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n, ok := c.notes[noteID]
	return ok && isMarked(n.Tags)
}

// ToggleMark adds the tag MarkedTag to all of the notes when any of them
// lacks it, and otherwise removes it from all of them.
func (c *Collection) ToggleMark(noteIDs []ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	add := false
	for _, id := range noteIDs {
		if n, ok := c.notes[id]; ok && !isMarked(n.Tags) {
			add = true
			break
		}
	}
	kind := "unmark"
	if add {
		kind = "mark"
	}
	return c.editTagsLocked(kind, noteIDs, func(have []string) []string {
		if add {
			return NormalizeTags(append(append([]string(nil), have...), MarkedTag))
		}
		var out []string
		for _, t := range have {
			if !strings.EqualFold(t, MarkedTag) {
				out = append(out, t)
			}
		}
		return out
	})
}

func (c *Collection) editTags(kind string, noteIDs []ID, edit func([]string) []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.editTagsLocked(kind, noteIDs, edit)
}

// editTagsLocked is editTags for callers that hold c.mu.
func (c *Collection) editTagsLocked(kind string, noteIDs []ID, edit func([]string) []string) error {
	o := &op{Kind: kind}
	for _, id := range noteIDs {
		n, ok := c.notes[id]
		if !ok {
			continue
		}
		tags := edit(n.Tags)
		if strings.Join(tags, " ") == strings.Join(n.Tags, " ") {
			continue
		}
		nn := cloneNote(n)
		nn.Tags = tags
		nn.Modified = c.now()
		o.Notes = append(o.Notes, &nn)
	}
	return c.commit(o)
}

// FindReplace is a find and replace across notes.
type FindReplace struct {
	Find    string
	Replace string
	// Field is the name of the one field to change; "" changes every field.
	// Notes whose note type has no field of that name are left unchanged.
	Field string
	// Tags changes the tags instead of the fields.
	Tags bool
	// Regex reads Find as an RE2 regular expression, and Replace may then
	// refer to its groups as $1 or ${name}. Otherwise both are literal.
	Regex      bool
	IgnoreCase bool
}

func (fr FindReplace) compile() (*regexp.Regexp, error) {
	if fr.Find == "" {
		return nil, userErr("findTextEmpty", "Enter the text to find.")
	}
	expr := nfc(fr.Find)
	if !fr.Regex {
		expr = regexp.QuoteMeta(expr)
	}
	if fr.IgnoreCase {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, userErr("invalidRegex", "Invalid regular expression: %s", strings.TrimPrefix(err.Error(), "error parsing regexp: "))
	}
	return re, nil
}

// findReplaceEdits returns the notes among noteIDs that fr changes, edited
// copies in the order of noteIDs. Callers hold a lock.
func (c *Collection) findReplaceEdits(noteIDs []ID, fr FindReplace) ([]*Note, error) {
	re, err := fr.compile()
	if err != nil {
		return nil, err
	}
	with, field := nfc(fr.Replace), nfc(fr.Field)
	replace := func(s string) string {
		if fr.Regex {
			return nfc(re.ReplaceAllString(s, with))
		}
		return nfc(re.ReplaceAllLiteralString(s, with))
	}
	var out []*Note
	seen := map[ID]bool{}
	for i, id := range noteIDs {
		if i%200 == 199 {
			c.yield()
		}
		n, ok := c.notes[id]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		nn := cloneNote(n)
		changed := false
		if fr.Tags {
			var tags []string
			for _, t := range n.Tags {
				tags = append(tags, replace(t))
			}
			tags = NormalizeTags(tags)
			changed = strings.Join(tags, " ") != strings.Join(n.Tags, " ")
			nn.Tags = tags
		} else {
			nt := c.noteTypes[n.TypeID]
			for j := range nn.Fields {
				if field != "" && (j >= len(nt.Fields) || nt.Fields[j] != field) {
					continue
				}
				if v := replace(nn.Fields[j]); v != nn.Fields[j] {
					nn.Fields[j] = v
					changed = true
				}
			}
		}
		if changed {
			out = append(out, &nn)
		}
	}
	return out, nil
}

// FindReplacePreview returns how many of the notes FindReplace would change.
func (c *Collection) FindReplacePreview(noteIDs []ID, fr FindReplace) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	edits, err := c.findReplaceEdits(noteIDs, fr)
	return len(edits), err
}

// FindReplace replaces text in the fields or tags of the notes, in one
// operation, and returns how many notes changed. Cards are regenerated as
// for UpdateNote, except that a note whose fields no longer generate any
// card is changed anyway: its cards are kept as empty cards.
func (c *Collection) FindReplace(noteIDs []ID, fr FindReplace) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	edits, err := c.findReplaceEdits(noteIDs, fr)
	if err != nil {
		return 0, err
	}
	o := &op{Kind: "findReplace"}
	for _, n := range edits {
		n.Modified = c.now()
		o.Notes = append(o.Notes, n)
		if !fr.Tags {
			nt := c.noteTypes[n.TypeID]
			c.syncCards(o, nt, n, cardOrds(nt, n.Fields), c.deckOfNote(n.ID), nil)
		}
	}
	if err := c.commit(o); err != nil {
		return 0, err
	}
	return len(edits), nil
}

// NoteTypeChange is what changing notes to another note type does, for the
// confirmation.
type NoteTypeChange struct {
	Notes        int `json:"notes"`
	CardsDeleted int `json:"cardsDeleted"`
	CardsCreated int `json:"cardsCreated"`
	// LostFields are the names of the source fields that no target field
	// takes and that have content in at least one of the notes.
	LostFields []string `json:"lostFields"`
}

// noteTypePlan is one note's part of a note type change.
type noteTypePlan struct {
	note  *Note // changed copy
	ords  []int
	remap func(int) int
	deck  ID
}

// planNoteTypeChange works out the change of noteIDs to note type target.
// fieldMap has one entry per target field: the index of the source field it
// takes, or -1 to leave it empty. templateMap has one entry per target card
// type: the index of the source card type whose cards it takes, or -1; it
// is used only between standard note types. Otherwise a card keeps its
// scheduling when the changed note generates a card with the same ord.
// Callers hold a lock.
func (c *Collection) planNoteTypeChange(noteIDs []ID, target ID, fieldMap, templateMap []int) ([]noteTypePlan, NoteTypeChange, error) {
	var info NoteTypeChange
	dst, ok := c.noteTypes[target]
	if !ok {
		return nil, info, userErr("noteTypeGone", "That note type no longer exists.")
	}
	var src *NoteType
	var notes []*Note
	seen := map[ID]bool{}
	for _, id := range noteIDs {
		n, ok := c.notes[id]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		if src != nil && n.TypeID != src.ID {
			return nil, info, userErr("sameNoteType", "All selected cards must use the same note type.")
		}
		src = c.noteTypes[n.TypeID]
		notes = append(notes, n)
	}
	if len(notes) == 0 {
		return nil, info, userErr("selectCardsToChange", "Select the cards of the notes to change.")
	}
	if len(fieldMap) != len(dst.Fields) {
		return nil, info, userErr("chooseFieldFor", "Choose a field for each field of %q.", dst.Name)
	}
	mapped := make([]bool, len(src.Fields))
	for _, s := range fieldMap {
		if s < -1 || s >= len(src.Fields) {
			return nil, info, userErr("fieldGone", "That field no longer exists.")
		}
		if s >= 0 {
			mapped[s] = true
		}
	}
	var remap func(int) int
	if src.Kind == Standard && dst.Kind == Standard {
		if len(templateMap) != len(dst.Templates) {
			return nil, info, userErr("chooseCardTypeFor", "Choose a card type for each card type of %q.", dst.Name)
		}
		to := map[int]int{}
		for t, s := range templateMap {
			if s < -1 || s >= len(src.Templates) {
				return nil, info, userErr("cardTypeGone", "That card type no longer exists.")
			}
			if s < 0 {
				continue
			}
			if _, dup := to[s]; dup {
				return nil, info, userErr("cardTypeOneTarget", "Card type %q can only go to one card type.", src.Templates[s].Name)
			}
			to[s] = t
		}
		remap = func(ord int) int {
			if t, ok := to[ord]; ok {
				return t
			}
			return -1
		}
	}

	lost := make([]bool, len(src.Fields))
	plans := make([]noteTypePlan, 0, len(notes))
	empty := 0
	for i, n := range notes {
		if i%200 == 199 {
			c.yield()
		}
		nn := cloneNote(n)
		nn.TypeID = dst.ID
		nn.Fields = make([]string, len(dst.Fields))
		for t, s := range fieldMap {
			if s >= 0 && s < len(n.Fields) {
				nn.Fields[t] = n.Fields[s]
			}
		}
		for s, v := range n.Fields {
			if s < len(mapped) && !mapped[s] && v != "" {
				lost[s] = true
			}
		}
		p := noteTypePlan{note: &nn, ords: cardOrds(dst, nn.Fields), remap: remap, deck: c.deckOfNote(n.ID)}
		if p.remap == nil {
			gen := map[int]bool{}
			for _, ord := range p.ords {
				gen[ord] = true
			}
			p.remap = func(ord int) int {
				if gen[ord] {
					return ord
				}
				return -1
			}
		}
		ch := c.cardChanges(n.ID, p.ords, p.remap)
		if ch.count == 0 {
			empty++
		}
		info.CardsDeleted += len(ch.del)
		info.CardsCreated += len(ch.create)
		plans = append(plans, p)
	}
	if empty > 0 {
		return nil, info, &Error{Msg: usermsg.Tag("notesWouldHaveNoCards", usermsg.Args(empty),
			plural(empty, "note")+" would have no cards. Map the fields its card types use, or keep a card type.")}
	}
	info.Notes = len(plans)
	for s, l := range lost {
		if l {
			info.LostFields = append(info.LostFields, src.Fields[s])
		}
	}
	return plans, info, nil
}

// ChangeNoteTypePreview returns what ChangeNoteType would do, without
// changing anything.
func (c *Collection) ChangeNoteTypePreview(noteIDs []ID, target ID, fieldMap, templateMap []int) (NoteTypeChange, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, info, err := c.planNoteTypeChange(noteIDs, target, fieldMap, templateMap)
	return info, err
}

// ChangeNoteType changes notes that share one note type to note type
// target, in one operation (see planNoteTypeChange for the maps). Cards that
// are kept keep their ID, so their scheduling and review history stay; the
// other cards are deleted, and cards the changed notes generate but lack
// are created New in the deck of the note's first card. Content of source
// fields no target field takes is lost.
func (c *Collection) ChangeNoteType(noteIDs []ID, target ID, fieldMap, templateMap []int) (NoteTypeChange, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	plans, info, err := c.planNoteTypeChange(noteIDs, target, fieldMap, templateMap)
	if err != nil {
		return info, err
	}
	nt := c.noteTypes[target]
	o := &op{Kind: "changeNoteType"}
	for _, p := range plans {
		p.note.Modified = c.now()
		o.Notes = append(o.Notes, p.note)
		c.syncCards(o, nt, p.note, p.ords, p.deck, p.remap)
	}
	if err := c.commit(o); err != nil {
		return NoteTypeChange{}, err
	}
	return info, nil
}

// AllNoteIDs returns the IDs of every note, oldest first.
func (c *Collection) AllNoteIDs() []ID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids := make([]ID, 0, len(c.notes))
	for id := range c.notes {
		ids = append(ids, id)
	}
	sortIDs(ids)
	return ids
}

// tagInTree reports whether tag is root or one of its child tags (root::…),
// ignoring case.
func tagInTree(tag, root string) bool {
	t, r := strings.ToLower(tag), strings.ToLower(root)
	return t == r || strings.HasPrefix(t, r+"::")
}

// normalizeTagName trims a tag name and the "::" around it. Tags never
// contain whitespace.
func normalizeTagName(name string) (string, error) {
	name = strings.Trim(strings.TrimSpace(nfc(name)), ":")
	if name == "" {
		return "", userErr("tagNameEmpty", "Enter a tag name.")
	}
	if strings.ContainsAny(name, " \t\r\n") {
		return "", userErr("tagNameSpaces", "Tag names cannot contain spaces.")
	}
	return name, nil
}

// RenameTag renames a tag and its child tags (ignoring case) on all notes, in
// one operation, and returns how many notes changed. Renaming into a tag a
// note already has merges the two. Moving a tag under another is a rename
// to parent::tag.
func (c *Collection) RenameTag(old, newName string) (int, error) {
	old = strings.TrimSpace(nfc(old))
	newName, err := normalizeTagName(newName)
	if err != nil {
		return 0, err
	}
	if old == "" {
		return 0, userErr("tagNameEmpty", "Enter a tag name.")
	}
	return c.editAllTags("renameTag", func(have []string) []string {
		out := make([]string, 0, len(have))
		for _, t := range have {
			if tagInTree(t, old) {
				t = newName + t[len(old):]
			}
			out = append(out, t)
		}
		return NormalizeTags(out)
	})
}

// ReparentTags moves tags, with their child tags, under parent ("" moves
// them to the top level), in one operation, and returns how many notes
// changed. Each tag keeps its last name part: moving verb::irregular under
// grammar makes it grammar::irregular, spelled as given. A tag whose parent
// is also moved moves with it, and moving into an existing tag merges the
// two.
func (c *Collection) ReparentTags(tags []string, parent string) (int, error) {
	parent = strings.TrimSpace(nfc(parent))
	if parent != "" {
		var err error
		if parent, err = normalizeTagName(parent); err != nil {
			return 0, err
		}
	}
	roots := tagRoots(tags)
	for _, t := range roots {
		if parent != "" && tagInTree(parent, t) {
			return 0, userErr("tagMoveIntoItself", "A tag cannot be moved into itself or one of its child tags.")
		}
	}
	renames := map[string]string{}
	for _, t := range roots {
		name := t
		if i := strings.LastIndex(t, "::"); i >= 0 {
			name = t[i+2:]
		}
		if parent != "" {
			name = parent + "::" + name
		}
		renames[t] = name
	}
	return c.editAllTags("moveTags", func(have []string) []string {
		out := make([]string, 0, len(have))
		for _, t := range have {
			for _, root := range roots {
				if tagInTree(t, root) {
					t = renames[root] + t[len(root):]
					break
				}
			}
			out = append(out, t)
		}
		return NormalizeTags(out)
	})
}

// tagRoots returns tags without blanks, duplicates (ignoring case) and tags
// whose parent is also listed.
func tagRoots(tags []string) []string {
	tags = nfcAll(tags)
	var out []string
	for _, t := range tags {
		t = strings.Trim(strings.TrimSpace(t), ":")
		if t == "" {
			continue
		}
		covered := false
		for _, o := range tags {
			o = strings.Trim(strings.TrimSpace(o), ":")
			if o != "" && !strings.EqualFold(o, t) && tagInTree(t, o) {
				covered = true
				break
			}
		}
		if !covered && !slices.ContainsFunc(out, func(o string) bool { return strings.EqualFold(o, t) }) {
			out = append(out, t)
		}
	}
	return out
}

// DeleteTags removes tags and their child tags (ignoring case) from all
// notes, in one operation, and returns how many notes changed. Notes are
// kept.
func (c *Collection) DeleteTags(tags []string) (int, error) {
	roots := tagRoots(tags)
	if len(roots) == 0 {
		return 0, nil
	}
	return c.editAllTags("deleteTags", func(have []string) []string {
		var out []string
		for _, t := range have {
			if !slices.ContainsFunc(roots, func(r string) bool { return tagInTree(t, r) }) {
				out = append(out, t)
			}
		}
		return out
	})
}

// TagUsage counts the notes that have any of the tags or one of their child
// tags.
func (c *Collection) TagUsage(tags []string) int {
	roots := tagRoots(tags)
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, note := range c.notes {
		if slices.ContainsFunc(note.Tags, func(t string) bool {
			return slices.ContainsFunc(roots, func(r string) bool { return tagInTree(t, r) })
		}) {
			n++
		}
	}
	return n
}

// ClearUnusedTags removes tags no note uses from the tag list and returns how
// many it removed. The tag list is derived from the notes, so there are never
// any: it reports zero.
func (c *Collection) ClearUnusedTags() int { return 0 }

// editAllTags applies edit to the tags of every note, in one operation, and
// returns how many notes changed.
func (c *Collection) editAllTags(kind string, edit func([]string) []string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ids := make([]ID, 0, len(c.notes))
	for id := range c.notes {
		ids = append(ids, id)
	}
	sortIDs(ids)
	o := &op{Kind: kind}
	for i, id := range ids {
		if i%200 == 199 {
			c.yield()
		}
		n := c.notes[id]
		tags := edit(n.Tags)
		if strings.Join(tags, " ") == strings.Join(n.Tags, " ") {
			continue
		}
		nn := cloneNote(n)
		nn.Tags = tags
		nn.Modified = c.now()
		o.Notes = append(o.Notes, &nn)
	}
	if err := c.commit(o); err != nil {
		return 0, err
	}
	return len(o.Notes), nil
}
