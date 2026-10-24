package core

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"fmt"
	"reflect"

	"cardflip/golang/storage"
)

// Undo history bounds (design decision 2): the last MaxUndoSteps steps, or
// fewer when their before-images pass maxUndoBytes. The most recent step is
// always kept, however large.
const MaxUndoSteps = 30

var maxUndoBytes = 50 << 20 // a variable for tests

// untracked op kinds are not undo steps: they change view state (a deck
// collapsed in the list), app settings or sync state, which undo leaves
// alone. Undo reverts only the settings fields a step itself changed (the
// remembered note type and deck).
var untracked = map[string]bool{"settings": true, "collapseDeck": true, opSync: true, opSyncFix: true, opSlots: true,
	opMedia: true, opRenumber: true, opPrune: true}

// step is one entry of the undo or redo stack: committing revert (with its
// review log writes) reverts the step. revert holds the before-images of
// everything the step changed; they are the objects the collection held,
// which are never modified in place, so keeping them costs no copies.
type step struct {
	name   string
	revert *op
	rev    []storage.RevlogWrite
	// settings are the settings after the step, when it changed them;
	// revert.Settings holds them before.
	settings *Settings
	// review is the entry of an answer step, for the study session's view.
	review ID
	size   int
}

// history is the undo and redo stacks of this app session. It is not
// persisted.
type history struct {
	undo, redo []*step
	bytes      int // before-image bytes of the undo stack
	pushed     uint64
}

func (h *history) clear() { *h = history{} }

// push adds a step to the undo stack and drops the oldest steps beyond the
// bounds.
func (h *history) push(s *step) {
	h.pushed++
	h.undo = append(h.undo, s)
	h.bytes += s.size
	for len(h.undo) > 1 && (len(h.undo) > MaxUndoSteps || h.bytes > maxUndoBytes) {
		h.bytes -= h.undo[0].size
		h.undo[0] = nil
		h.undo = h.undo[1:]
	}
}

func (h *history) popUndo() {
	h.bytes -= h.undo[len(h.undo)-1].size
	h.undo = h.undo[:len(h.undo)-1]
}

// notes returns the notes and note types the history can bring back, whose
// media compaction must keep (design decision 3).
func (h *history) notes() ([]*Note, []*NoteType) {
	var notes []*Note
	var types []*NoteType
	for _, stack := range [][]*step{h.undo, h.redo} {
		for _, s := range stack {
			notes = append(notes, s.revert.Notes...)
			types = append(types, s.revert.NoteTypes...)
		}
	}
	return notes, types
}

// commit modes: a new change, or the revert of an undo or redo step.
type commitMode int

const (
	modeNew commitMode = iota
	modeUndo
	modeRedo
)

// before returns the current versions of the objects an op puts or deletes
// (restore), and the IDs of those it creates (remove).
func before[T any](cur map[ID]*T, put []*T, del []ID, idOf func(*T) ID) (restore []*T, remove []ID) {
	seen := map[ID]bool{}
	for _, v := range put {
		id := idOf(v)
		if seen[id] {
			continue
		}
		seen[id] = true
		if old, ok := cur[id]; ok {
			restore = append(restore, old)
		} else {
			remove = append(remove, id)
		}
	}
	for _, id := range del {
		if seen[id] {
			continue
		}
		seen[id] = true
		if old, ok := cur[id]; ok {
			restore = append(restore, old)
		}
	}
	return restore, remove
}

// inverse builds, before o is applied, the step that reverts it: the
// before-images of everything o changes (design decision 1). Settings are
// not reverted. Callers hold the write lock.
func (c *Collection) inverse(o *op) (*step, error) {
	inv := &op{}
	inv.Presets, inv.DelPresets = before(c.presets, o.Presets, o.DelPresets, func(p *Preset) ID { return p.ID })
	inv.Decks, inv.DelDecks = before(c.decks, o.Decks, o.DelDecks, func(d *Deck) ID { return d.ID })
	inv.NoteTypes, inv.DelNoteTypes = before(c.noteTypes, o.NoteTypes, o.DelNoteTypes, func(nt *NoteType) ID { return nt.ID })
	inv.Notes, inv.DelNotes = before(c.notes, o.Notes, o.DelNotes, func(n *Note) ID { return n.ID })
	inv.Cards, inv.DelCards = before(c.cards, o.Cards, o.DelCards, func(card *Card) ID { return card.ID })
	s := &step{name: stepName(o), revert: inv}
	if o.Settings != nil {
		old, now := c.settings, *o.Settings
		inv.Settings, s.settings = &old, &now
	}
	if o.Review != nil {
		s.review = o.Review.ID
		inv.ReviewUndone = o.Review.ID
		s.rev = append(s.rev, storage.RevlogWrite{Month: monthKey(o.Review.Time), ID: int64(o.Review.ID)})
	}
	if len(o.AddReviews)+len(o.DelReviews) > 0 {
		inv.AddReviews, inv.DelReviews = o.DelReviews, o.AddReviews
		rev, err := reviewWrites(inv.AddReviews, inv.DelReviews)
		if err != nil {
			return nil, err
		}
		s.rev = append(s.rev, rev...)
	}
	if o.ReviewUndone != 0 {
		if e := c.reviewEntry(o.ReviewUndone); e != nil {
			raw, err := json.Marshal(e)
			if err != nil {
				return nil, err
			}
			inv.Review = e
			s.rev = append(s.rev, storage.RevlogWrite{Month: monthKey(e.Time), ID: int64(e.ID), Entry: raw})
		}
	}
	s.size = inv.size()
	return s, nil
}

// size estimates the memory an op's objects hold, for the history bound.
func (o *op) size() int {
	const id, obj = 8, 64
	n := id * (len(o.DelPresets) + len(o.DelDecks) + len(o.DelNoteTypes) + len(o.DelNotes) + len(o.DelCards))
	for _, p := range o.Presets {
		n += obj + len(p.Name) + 256
	}
	for _, d := range o.Decks {
		n += obj + len(d.Name)
	}
	for _, nt := range o.NoteTypes {
		n += obj + len(nt.Name) + len(nt.CSS)
		for _, f := range nt.Fields {
			n += len(f)
		}
		for _, t := range nt.Templates {
			n += len(t.Name) + len(t.QFmt) + len(t.AFmt)
		}
	}
	for _, note := range o.Notes {
		n += obj + len(note.GUID)
		for _, f := range note.Fields {
			n += len(f) + 16
		}
		for _, t := range note.Tags {
			n += len(t) + 16
		}
	}
	n += len(o.Cards) * (obj + 96)
	if o.Review != nil {
		n += obj + 96
	}
	for _, e := range append(o.AddReviews[:len(o.AddReviews):len(o.AddReviews)], o.DelReviews...) {
		n += obj + 96 + len(e.Source)
	}
	return n
}

func (c *Collection) reviewEntry(id ID) *ReviewEntry {
	for _, e := range c.revlog {
		if e.ID == id {
			return e
		}
	}
	return nil
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// stepName names the step an op makes, for "Undo <name>". The name is
// tagged (see package usermsg), so the app shows it translated.
func stepName(o *op) string {
	switch o.Kind {
	case "addNote":
		return usermsg.Text("stepAddNote", "Add note")
	case "addNotes":
		return usermsg.Tag("stepAddNotes", usermsg.Args(len(o.Notes)), "Add "+plural(len(o.Notes), "note"))
	case "updateNote":
		return usermsg.Text("stepUpdateNote", "Edit note")
	case "deleteNotes":
		return usermsg.Tag("stepDeleteNotes", usermsg.Args(len(o.DelNotes)), "Delete "+plural(len(o.DelNotes), "note"))
	case "addTags":
		return usermsg.Text("stepAddTags", "Add tags")
	case "removeTags":
		return usermsg.Text("stepRemoveTags", "Remove tags")
	case "mark":
		return usermsg.Tag("stepMark", usermsg.Args(len(o.Notes)), "Mark "+plural(len(o.Notes), "note"))
	case "unmark":
		return usermsg.Tag("stepUnmark", usermsg.Args(len(o.Notes)), "Unmark "+plural(len(o.Notes), "note"))
	case "renameTag":
		return usermsg.Text("stepRenameTag", "Rename tag")
	case "moveTags":
		return usermsg.Text("stepMoveTags", "Move tags")
	case "deleteTags":
		return usermsg.Text("stepDeleteTags", "Delete tags")
	case "findReplace":
		return usermsg.Text("stepFindReplace", "Find and replace")
	case "changeNoteType":
		return usermsg.Text("stepChangeNoteType", "Change note type")
	case "moveCards":
		return usermsg.Tag("stepMoveCards", usermsg.Args(len(o.Cards)), "Move "+plural(len(o.Cards), "card"))
	case "suspend":
		return usermsg.Tag("stepSuspend", usermsg.Args(len(o.Cards)), "Suspend "+plural(len(o.Cards), "card"))
	case "unsuspend":
		return usermsg.Tag("stepUnsuspend", usermsg.Args(len(o.Cards)), "Unsuspend "+plural(len(o.Cards), "card"))
	case "bury":
		return usermsg.Tag("stepBury", usermsg.Args(len(o.Cards)), "Bury "+plural(len(o.Cards), "card"))
	case "unbury":
		return usermsg.Text("stepUnbury", "Unbury")
	case "flag":
		return usermsg.Tag("stepFlag", usermsg.Args(len(o.Cards)), "Flag "+plural(len(o.Cards), "card"))
	case "unflag":
		return usermsg.Tag("stepUnflag", usermsg.Args(len(o.Cards)), "Remove flag from "+plural(len(o.Cards), "card"))
	case "resetCards":
		return usermsg.Tag("stepResetCards", usermsg.Args(len(o.Cards)), "Reset "+plural(len(o.Cards), "card"))
	case "setDueDate":
		return usermsg.Tag("stepSetDueDate", usermsg.Args(len(o.Cards)), "Set due date of "+plural(len(o.Cards), "card"))
	case "reposition":
		return usermsg.Text("stepReposition", "Reposition")
	case "deleteEmptyCards":
		return usermsg.Tag("stepDeleteEmptyCards", usermsg.Args(len(o.DelCards)), "Delete "+plural(len(o.DelCards), "empty card"))
	case "createDeck":
		return usermsg.Text("stepCreateDeck", "Create deck")
	case "renameDeck":
		return usermsg.Text("stepRenameDeck", "Rename deck")
	case "deckOptions":
		return usermsg.Text("stepDeckOptions", "Change deck options")
	case "optimize":
		return usermsg.Text("stepOptimize", "Optimize all presets")
	case "addPreset":
		return usermsg.Text("stepAddPreset", "Add preset")
	case "clonePreset":
		return usermsg.Text("stepClonePreset", "Clone preset")
	case "renamePreset":
		return usermsg.Text("stepRenamePreset", "Rename preset")
	case "deletePreset":
		return usermsg.Text("stepDeletePreset", "Delete preset")
	case "deckPreset":
		return usermsg.Text("stepDeckPreset", "Change preset")
	case "presetToSubdecks":
		return usermsg.Text("stepPresetToSubdecks", "Save to all subdecks")
	case "deckLimits":
		return usermsg.Text("stepDeckLimits", "Change deck limits")
	case "todayLimits":
		return usermsg.Text("stepTodayLimits", "Custom study")
	case "deckDescription":
		return usermsg.Text("stepDeckDescription", "Edit deck description")
	case "deleteDeck":
		return usermsg.Text("stepDeleteDeck", "Delete deck")
	case "buildFiltered":
		return usermsg.Text("stepBuildFiltered", "Build filtered deck")
	case "emptyFiltered":
		return usermsg.Text("stepEmptyFiltered", "Empty filtered deck")
	case "customStudy":
		return usermsg.Text("stepCustomStudy", "Custom study")
	case "addNoteType":
		return usermsg.Text("stepAddNoteType", "Add note type")
	case "cloneNoteType":
		return usermsg.Text("stepCloneNoteType", "Clone note type")
	case "deleteNoteType":
		return usermsg.Text("stepDeleteNoteType", "Delete note type")
	case "renameNoteType":
		return usermsg.Text("stepRenameNoteType", "Rename note type")
	case "renameField":
		return usermsg.Text("stepRenameField", "Rename field")
	case "addField":
		return usermsg.Text("stepAddField", "Add field")
	case "removeField":
		return usermsg.Text("stepRemoveField", "Remove field")
	case "addTemplate":
		return usermsg.Text("stepAddTemplate", "Add card type")
	case "updateTemplate":
		return usermsg.Text("stepUpdateTemplate", "Edit card type")
	case "removeTemplate":
		return usermsg.Text("stepRemoveTemplate", "Remove card type")
	case "moveTemplate":
		return usermsg.Text("stepMoveTemplate", "Move card type")
	case "noteTypeCSS":
		return usermsg.Text("stepNoteTypeCSS", "Edit styling")
	case "fieldOptions":
		return usermsg.Text("stepFieldOptions", "Change field options")
	case "moveField":
		return usermsg.Text("stepMoveField", "Move field")
	case "sortField":
		return usermsg.Text("stepSortField", "Change sort field")
	case "import":
		return usermsg.Text("stepImport", "Import")
	case "answer":
		return usermsg.Text("stepAnswer", "Answer card")
	case "undoAnswer":
		return usermsg.Text("stepUndoAnswer", "Undo answer")
	case "checkDatabase":
		return usermsg.Text("stepCheckDatabase", "Check database")
	}
	return usermsg.Text("stepChange", "Change")
}

// UndoStatus names the steps Undo and Redo would act on; empty means there
// is none.
func (c *Collection) UndoStatus() (undo, redo string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n := len(c.hist.undo); n > 0 {
		undo = c.hist.undo[n-1].name
	}
	if n := len(c.hist.redo); n > 0 {
		redo = c.hist.redo[n-1].name
	}
	return undo, redo
}

// LastStep returns how many steps were pushed onto the undo stack in this
// app session, and the name of the step on top. A caller compares the count
// before and after a change to learn whether the change made a step.
func (c *Collection) LastStep() (uint64, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	name := ""
	if n := len(c.hist.undo); n > 0 {
		name = c.hist.undo[n-1].name
	}
	return c.hist.pushed, name
}

// Undo reverts the most recent step and returns its name. The revert is
// saved like any change, so it is durable; Redo re-applies the step.
func (c *Collection) Undo() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.undoLocked()
}

func (c *Collection) undoLocked() (string, error) {
	n := len(c.hist.undo)
	if n == 0 {
		return "", userErr("nothingToUndo", "There is nothing to undo.")
	}
	s := c.hist.undo[n-1]
	if err := c.commitAs(c.revertOp(s, "undo"), s.rev, modeUndo); err != nil {
		return "", err
	}
	return s.name, nil
}

// Redo re-applies the most recently undone step and returns its name.
func (c *Collection) Redo() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := len(c.hist.redo)
	if n == 0 {
		return "", userErr("nothingToRedo", "There is nothing to redo.")
	}
	s := c.hist.redo[n-1]
	if err := c.commitAs(c.revertOp(s, "redo"), s.rev, modeRedo); err != nil {
		return "", err
	}
	return s.name, nil
}

// revertOp returns the op that reverts s. Only the settings fields s
// changed are set back, and decks keep their current collapsed state, which
// is not part of the history.
func (c *Collection) revertOp(s *step, kind string) *op {
	o := *s.revert
	o.Kind = kind
	if s.revert.Settings != nil {
		st := revertFields(c.settings, *s.revert.Settings, *s.settings)
		o.Settings = &st
		if st.equal(c.settings) {
			o.Settings = nil
		}
	}
	o.Decks = make([]*Deck, len(s.revert.Decks))
	for i, d := range s.revert.Decks {
		if cur, ok := c.decks[d.ID]; ok && cur.Collapsed != d.Collapsed {
			nd := *d
			nd.Collapsed = cur.Collapsed
			d = &nd
		}
		o.Decks[i] = d
	}
	return &o
}

// revertFields returns cur with the fields that differ between before and
// after set back to before.
func revertFields(cur, before, after Settings) Settings {
	v, b, a := reflect.ValueOf(&cur).Elem(), reflect.ValueOf(before), reflect.ValueOf(after)
	for i := range v.NumField() {
		bf, af := b.Field(i), a.Field(i)
		if bf.Kind() == reflect.Map && bf.Len() == 0 && af.Len() == 0 {
			continue
		}
		if !reflect.DeepEqual(bf.Interface(), af.Interface()) {
			v.Field(i).Set(b.Field(i))
		}
	}
	return cur
}

// record updates the history after o, a change of mode, was saved; s
// reverts o. Callers hold the write lock.
func (c *Collection) record(mode commitMode, o *op, s *step) {
	switch mode {
	case modeNew:
		if untracked[o.Kind] {
			return
		}
		c.hist.redo = nil
		c.hist.push(s)
	case modeUndo:
		top := c.hist.undo[len(c.hist.undo)-1]
		c.hist.popUndo()
		s.name, s.review = top.name, 0
		c.hist.redo = append(c.hist.redo, s)
	case modeRedo:
		top := c.hist.redo[len(c.hist.redo)-1]
		c.hist.redo = c.hist.redo[:len(c.hist.redo)-1]
		s.name = top.name
		c.hist.push(s)
	}
}
