package core

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Versions of a shared deck, as an Anki package's batch would hold them:
// version 1 has a typo in vocab-hola, version 2 (a month later) fixes it and
// restyles the note type, version 3 (a month after that) adds the field
// Audio, shown on the back, and the note vocab-perro.
var (
	vocabT1 = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	vocabT2 = vocabT1.AddDate(0, 1, 0)
	vocabT3 = vocabT2.AddDate(0, 1, 0)
)

func vocabBatch(version int) *ImportBatch {
	nt := NoteType{Name: "Vocab", Kind: Standard, Fields: []string{"Front", "Back"}, CSS: ".card {}", Source: "anki:42",
		Templates: []CardTemplate{{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}}"}},
		Modified:  vocabT1}
	note := func(guid string, mod time.Time, fields ...string) ImportNote {
		return ImportNote{TypeKey: "v", GUID: guid, DeckName: "Vocab", Fields: fields, Modified: mod}
	}
	notes := []ImportNote{note("vocab-hola", vocabT1, "hola", "helo"), note("vocab-gato", vocabT1, "gato", "cat")}
	if version >= 2 {
		nt.CSS, nt.Modified = ".card { color: red; }", vocabT2
		notes[0] = note("vocab-hola", vocabT2, "hola", "hello")
	}
	if version >= 3 {
		nt.Fields = append(nt.Fields, "Audio")
		nt.Templates[0].AFmt += "{{Audio}}"
		nt.Modified = vocabT3
		notes = []ImportNote{note("vocab-hola", vocabT3, "hola", "hello", "hola.mp3"),
			note("vocab-gato", vocabT3, "gato", "cat", ""), note("vocab-perro", vocabT3, "perro", "dog", "perro.mp3")}
	}
	return &ImportBatch{NoteTypes: []ImportNoteType{{Key: "v", Type: nt}}, Notes: notes}
}

func importVocab(t *testing.T, c *Collection, version int, edit func(b *ImportBatch)) ImportSummary {
	t.Helper()
	b := vocabBatch(version)
	if edit != nil {
		edit(b)
	}
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatalf("import version %d: %v", version, err)
	}
	return sum
}

func noteByGUID(t *testing.T, c *Collection, guid string) Note {
	t.Helper()
	c.mu.RLock()
	defer c.mu.RUnlock()
	var found []Note
	for _, n := range c.notes {
		if n.GUID == guid {
			found = append(found, cloneNote(n))
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d notes %s", len(found), guid)
	}
	return found[0]
}

func noteTypesNamed(c *Collection, prefix string) []NoteType {
	var out []NoteType
	for _, nt := range c.NoteTypes() {
		if strings.HasPrefix(nt.Name, prefix) {
			out = append(out, nt)
		}
	}
	return out
}

// Re-importing a package (data-import spec): import twice.
func TestReimportTwice(t *testing.T) {
	c, _, _ := testCollection(t)
	if sum := importVocab(t, c, 1, nil); sum.NotesAdded != 2 || sum.NoteTypesAdded != 1 {
		t.Fatalf("first import = %+v", sum)
	}
	nt := noteTypesNamed(c, "Vocab")[0]
	if !nt.Modified.Equal(vocabT1) {
		t.Fatalf("note type modified %v, want the package's %v", nt.Modified, vocabT1)
	}
	if n := noteByGUID(t, c, "vocab-hola"); !n.Modified.Equal(vocabT1) {
		t.Fatalf("note modified %v, want the package's %v", n.Modified, vocabT1)
	}
	for _, policy := range []UpdatePolicy{"", UpdateIfNewer, UpdateAlways, UpdateNever} {
		sum := importVocab(t, c, 1, func(b *ImportBatch) { b.UpdateNotes, b.UpdateNoteTypes = policy, policy })
		want := ImportSummary{NotesSkipped: 2, Warnings: []string{}}
		sum.Warnings = []string{}
		if !reflect.DeepEqual(sum, want) {
			t.Fatalf("%q: second import = %+v", policy, sum)
		}
	}
	noteByGUID(t, c, "vocab-hola")
	if n := len(noteTypesNamed(c, "Vocab")); n != 1 {
		t.Fatalf("%d Vocab note types", n)
	}
}

// Re-importing a package (data-import spec): a local edit is kept with "if
// newer", for the same package and for a newer one modified before the edit.
func TestReimportKeepsLocalEdit(t *testing.T) {
	c, _, clk := testCollection(t)
	importVocab(t, c, 1, nil)
	hola := noteByGUID(t, c, "vocab-hola")
	cards := c.CardsOfNote(hola.ID)
	if err := c.UpdateNote(hola.ID, []string{"hola", "hello!"}, nil); err != nil {
		t.Fatal(err)
	}
	for _, version := range []int{1, 2} {
		sum := importVocab(t, c, version, nil)
		if sum.NotesUpdated != 0 || sum.NotesSkippedOlder != 1 || sum.NotesSkipped != 1 {
			t.Fatalf("version %d: summary %+v", version, sum)
		}
		if n := noteByGUID(t, c, "vocab-hola"); n.Fields[1] != "hello!" || !n.Modified.Equal(clk.t) {
			t.Fatalf("version %d: local edit lost: %q %v", version, n.Fields, n.Modified)
		}
	}
	// "Always" overwrites it; scheduling stays.
	sum := importVocab(t, c, 2, func(b *ImportBatch) { b.UpdateNotes = UpdateAlways })
	if sum.NotesUpdated != 1 || sum.NotesSkippedOlder != 0 {
		t.Fatalf("always: summary %+v", sum)
	}
	n := noteByGUID(t, c, "vocab-hola")
	if n.Fields[1] != "hello" || !n.Modified.Equal(vocabT2) {
		t.Fatalf("always: note %q %v", n.Fields, n.Modified)
	}
	if got := c.CardsOfNote(n.ID); !reflect.DeepEqual(got, cards) {
		t.Fatalf("cards changed:\n%+v\n%+v", cards, got)
	}
}

// Re-importing a package (data-import spec): the updated deck's fix arrives
// with "if newer" and "always", not with "never"; notes the author did not
// touch stay as they are.
func TestReimportUpdatedDeck(t *testing.T) {
	for _, policy := range []UpdatePolicy{"", UpdateIfNewer, UpdateAlways, UpdateNever} {
		c, _, _ := testCollection(t)
		importVocab(t, c, 1, nil)
		hola := noteByGUID(t, c, "vocab-hola")
		cards := c.CardsOfNote(hola.ID)
		sum := importVocab(t, c, 2, func(b *ImportBatch) { b.UpdateNotes = policy })
		n := noteByGUID(t, c, "vocab-hola")
		if policy == UpdateNever {
			if sum.NotesUpdated != 0 || sum.NotesSkipped != 2 || n.Fields[1] != "helo" {
				t.Fatalf("never: summary %+v, note %q", sum, n.Fields)
			}
			continue
		}
		if sum.NotesUpdated != 1 || sum.NotesSkipped != 1 || sum.NotesAdded != 0 || n.Fields[1] != "hello" {
			t.Fatalf("%q: summary %+v, note %q", policy, sum, n.Fields)
		}
		if !n.Modified.Equal(vocabT2) {
			t.Fatalf("%q: modified %v", policy, n.Modified)
		}
		if got := c.CardsOfNote(n.ID); !reflect.DeepEqual(got, cards) {
			t.Fatalf("%q: cards changed", policy)
		}
	}
}

func TestImportUnknownUpdatePolicy(t *testing.T) {
	c, _, _ := testCollection(t)
	b := vocabBatch(1)
	b.UpdateNotes = "sometimes"
	if _, err := c.Import(b, nil); err == nil {
		t.Fatal("unknown policy accepted")
	}
}

func TestUpdatePolicyAllows(t *testing.T) {
	var zero time.Time
	for _, tc := range []struct {
		p                  UpdatePolicy
		imported, existing time.Time
		want               bool
	}{
		{UpdateIfNewer, vocabT2, vocabT1, true},
		{UpdateIfNewer, vocabT1, vocabT1, false},
		{UpdateIfNewer, vocabT1, vocabT2, false},
		{"", vocabT2, vocabT1, true},
		// No time is older than any.
		{UpdateIfNewer, vocabT1, zero, true},
		{UpdateIfNewer, zero, vocabT1, false},
		{UpdateIfNewer, zero, zero, false},
		{UpdateAlways, zero, vocabT2, true},
		{UpdateNever, vocabT2, zero, false},
	} {
		if got := tc.p.allows(tc.imported, tc.existing); got != tc.want {
			t.Errorf("%q.allows(%v, %v) = %v", tc.p, tc.imported, tc.existing, got)
		}
	}
}

// Note type updates on import (data-import spec): merge adds a field.
func TestImportMergeAddsField(t *testing.T) {
	c, store, clk := testCollection(t)
	importVocab(t, c, 1, nil)
	vocab := noteTypesNamed(c, "Vocab")[0]
	gato := noteByGUID(t, c, "vocab-gato")
	sum := importVocab(t, c, 3, func(b *ImportBatch) { b.MergeNoteTypes = true })
	if sum.NoteTypesMerged != 1 || sum.NoteTypesAdded != 0 || sum.NotesUpdated != 1 || sum.NotesAdded != 1 || sum.NotesSkipped != 1 {
		t.Fatalf("merge import = %+v", sum)
	}
	types := noteTypesNamed(c, "Vocab")
	if len(types) != 1 || types[0].ID != vocab.ID {
		t.Fatalf("note types = %+v", types)
	}
	nt := types[0]
	if strings.Join(nt.Fields, ",") != "Front,Back,Audio" || !strings.Contains(nt.Templates[0].AFmt, "{{Audio}}") ||
		nt.CSS != ".card { color: red; }" || !nt.Modified.Equal(vocabT3) {
		t.Fatalf("merged note type = %+v", nt)
	}
	if n := noteByGUID(t, c, "vocab-hola"); strings.Join(n.Fields, "|") != "hola|hello|hola.mp3" || !n.Modified.Equal(vocabT3) {
		t.Fatalf("updated note = %q %v", n.Fields, n.Modified)
	}
	if n := noteByGUID(t, c, "vocab-perro"); n.TypeID != vocab.ID || strings.Join(n.Fields, "|") != "perro|dog|perro.mp3" {
		t.Fatalf("added note = %+v", n)
	}
	// A note the package left as it was gets the field, empty, and keeps
	// its modification time.
	if n := noteByGUID(t, c, "vocab-gato"); strings.Join(n.Fields, "|") != "gato|cat|" || !n.Modified.Equal(gato.Modified) {
		t.Fatalf("untouched note = %q %v", n.Fields, n.Modified)
	}
	reopen(t, c, store, clk)
	// Importing version 3 again changes nothing, merged or not.
	for _, merge := range []bool{true, false} {
		sum = importVocab(t, c, 3, func(b *ImportBatch) { b.MergeNoteTypes = merge })
		if sum.NotesSkipped != 3 || sum.NoteTypesMerged+sum.NoteTypesAdded+sum.NoteTypesUpdated != 0 {
			t.Fatalf("merge %v: version 3 again = %+v", merge, sum)
		}
	}
	// Undo restores the note type and notes from before the merge.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(vocab.ID); !reflect.DeepEqual(nt, vocab) {
		t.Fatalf("after undo: %+v, want %+v", nt, vocab)
	}
	if n := noteByGUID(t, c, "vocab-gato"); !reflect.DeepEqual(n, gato) {
		t.Fatalf("after undo: %+v", n)
	}
}

// Merging matches fields and card types by name, whatever their order, keeps
// the collection's own, and adds the missing card types with their cards.
func TestImportMergeByName(t *testing.T) {
	c, _, clk := testCollection(t)
	importVocab(t, c, 1, nil)
	vocab := noteTypesNamed(c, "Vocab")[0]
	hola := noteByGUID(t, c, "vocab-hola")
	// The user adds a field of their own, after the package's version.
	if err := c.AddField(vocab.ID, "Mine"); err != nil {
		t.Fatal(err)
	}
	clk.t = clk.t.Add(time.Hour)
	sum := importVocab(t, c, 1, func(b *ImportBatch) {
		b.MergeNoteTypes = true
		b.UpdateNotes = UpdateAlways
		nt := &b.NoteTypes[0].Type
		// Fields reversed and renamed in case, a second card type first.
		nt.Fields = []string{"back", "Front"}
		nt.Templates = []CardTemplate{{Name: "Reverse", QFmt: "{{back}}", AFmt: "{{Front}}"}, nt.Templates[0]}
		nt.Modified = vocabT2
		for i := range b.Notes {
			n := &b.Notes[i]
			n.Fields = []string{n.Fields[1] + "!", n.Fields[0]}
			n.Modified = clk.t
			n.Cards = []ImportCard{{Ord: 0, DeckName: "Vocab"}, {Ord: 1, DeckName: "Vocab"}}
		}
		b.Notes = append(b.Notes, ImportNote{TypeKey: "v", GUID: "vocab-new", DeckName: "Vocab", Fields: []string{"house", "casa"},
			Cards: []ImportCard{{Ord: 0, DeckName: "Vocab"}, {Ord: 1, DeckName: "Vocab"}}})
	})
	if sum.NoteTypesMerged != 1 || sum.NotesUpdated != 2 || sum.NotesAdded != 1 {
		t.Fatalf("merge = %+v", sum)
	}
	nt, _ := c.NoteType(vocab.ID)
	if strings.Join(nt.Fields, ",") != "Front,Back,Mine" || len(nt.Templates) != 2 || nt.Templates[0].Name != "Card 1" || nt.Templates[1].Name != "Reverse" {
		t.Fatalf("merged = %+v", nt)
	}
	if n := noteByGUID(t, c, "vocab-hola"); strings.Join(n.Fields, "|") != "hola|helo!|" {
		t.Fatalf("hola = %q", n.Fields)
	}
	// Existing notes keep their card and get one for the new card type.
	cards := c.CardsOfNote(hola.ID)
	if len(cards) != 2 || cards[0].Ord != 0 || cards[1].Ord != 1 {
		t.Fatalf("hola cards = %+v", cards)
	}
	if sum.CardsAdded != 4 {
		t.Fatalf("cards added = %d, want 2 for the existing notes and 2 for the new one", sum.CardsAdded)
	}
	n := noteByGUID(t, c, "vocab-new")
	if strings.Join(n.Fields, "|") != "casa|house|" {
		t.Fatalf("new note = %q", n.Fields)
	}
	var ords []int
	for _, card := range c.CardsOfNote(n.ID) {
		ords = append(ords, card.Ord)
	}
	if !reflect.DeepEqual(ords, []int{0, 1}) {
		t.Fatalf("new note ords = %v", ords)
	}
}

// Note type updates on import (data-import spec): no merge.
func TestImportWithoutMerge(t *testing.T) {
	c, _, _ := testCollection(t)
	importVocab(t, c, 1, nil)
	vocab := noteTypesNamed(c, "Vocab")[0]
	sum := importVocab(t, c, 3, nil)
	if sum.NoteTypesAdded != 1 || sum.NoteTypesMerged+sum.NoteTypesUpdated != 0 || sum.NotesAdded != 1 || sum.NotesUpdated != 0 {
		t.Fatalf("import = %+v", sum)
	}
	if nt, _ := c.NoteType(vocab.ID); !reflect.DeepEqual(nt, vocab) {
		t.Fatalf("existing note type changed: %+v", nt)
	}
	types := noteTypesNamed(c, "Vocab")
	if len(types) != 2 || types[1].Name != "Vocab (2)" || strings.Join(types[1].Fields, ",") != "Front,Back,Audio" {
		t.Fatalf("note types = %+v", types)
	}
	if n := noteByGUID(t, c, "vocab-perro"); n.TypeID != types[1].ID {
		t.Fatalf("package note uses %d, want the new note type %d", n.TypeID, types[1].ID)
	}
	// The existing notes keep their note type, so they are not updated.
	if n := noteByGUID(t, c, "vocab-hola"); n.TypeID != vocab.ID || n.Fields[1] != "helo" {
		t.Fatalf("existing note = %+v", n)
	}
	if len(sum.Warnings) != 2 {
		t.Fatalf("warnings = %q", sum.Warnings)
	}
	// The next import uses the new note type instead of adding another.
	sum = importVocab(t, c, 3, nil)
	if sum.NoteTypesAdded != 0 || sum.NotesAdded != 0 || len(noteTypesNamed(c, "Vocab")) != 2 {
		t.Fatalf("again = %+v", sum)
	}
	// A different kind is never merged.
	sum = importVocab(t, c, 1, func(b *ImportBatch) {
		b.MergeNoteTypes = true
		b.NoteTypes[0].Type.Kind = Cloze
		b.NoteTypes[0].Type.Templates[0].QFmt = "{{cloze:Front}}"
		b.Notes = []ImportNote{{TypeKey: "v", GUID: "vocab-cloze", DeckName: "Vocab", Fields: []string{"{{c1::x}}", ""}}}
	})
	if sum.NoteTypesAdded != 1 || sum.NoteTypesMerged != 0 {
		t.Fatalf("cloze = %+v", sum)
	}
}
