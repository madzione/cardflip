package core

import (
	"strings"
	"testing"
)

// textBatch is a text import of Basic rows (GUID, deck, front, back), as
// the CSV importer builds it.
func textBatch(c *Collection, dup DupMode, scope DupScope, rows ...[4]string) *ImportBatch {
	var basic ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == StockBasic {
			basic = nt.ID
		}
	}
	b := &ImportBatch{NoteTypes: []ImportNoteType{{Key: "csv", ExistingID: basic}}, Duplicates: dup, DupScope: scope}
	for _, r := range rows {
		b.Notes = append(b.Notes, ImportNote{TypeKey: "csv", GUID: r[0], DeckName: r[1], Fields: []string{r[2], r[3]}})
	}
	return b
}

func importTextBatch(t *testing.T, c *Collection, b *ImportBatch) ImportSummary {
	t.Helper()
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sum
}

func TestTextImportSkipDuplicates(t *testing.T) {
	// Skip duplicates (data-import spec): no new note, counted as skipped.
	c, _, _ := testCollection(t)
	importTextBatch(t, c, textBatch(c, DupSkip, "", [4]string{"", "Default", "hola", "hello"}))
	sum := importTextBatch(t, c, textBatch(c, DupSkip, "", [4]string{"", "Default", "<b>hola</b>", "hi"}))
	if sum.NotesAdded != 0 || sum.NotesSkipped != 1 {
		t.Fatalf("summary = %+v", sum)
	}
	// A row with a GUID no note has is a duplicate by its first field too.
	sum = importTextBatch(t, c, textBatch(c, DupSkip, "", [4]string{"new-guid", "Default", "hola", "hi"}))
	if sum.NotesAdded != 0 || sum.NotesSkipped != 1 {
		t.Fatalf("unknown GUID summary = %+v", sum)
	}
	// Rows the importer left out count as skipped.
	b := textBatch(c, DupSkip, "")
	b.Skipped = 3
	if sum := importTextBatch(t, c, b); sum.NotesSkipped != 3 {
		t.Fatalf("pre-skipped summary = %+v", sum)
	}
}

func TestTextImportGUIDUpdate(t *testing.T) {
	// GUID update (data-import spec): the existing note's Back is updated,
	// whatever the duplicate choice, also though the text has no
	// modification time.
	c, _, _ := testCollection(t)
	importTextBatch(t, c, textBatch(c, DupSkip, "", [4]string{"g1", "Default", "hola", "hello"}))
	for i, dup := range []DupMode{DupSkip, DupUpdate, DupDuplicate} {
		back := []string{"hi", "hey", "howdy"}[i]
		sum := importTextBatch(t, c, textBatch(c, dup, "", [4]string{"g1", "Default", "hola", back}))
		if sum.NotesUpdated != 1 || sum.NotesAdded != 0 {
			t.Fatalf("%s: summary = %+v", dup, sum)
		}
		if n := noteByGUID(t, c, "g1"); n.Fields[1] != back {
			t.Fatalf("%s: fields = %q", dup, n.Fields)
		}
	}
	// A GUID can change the first field: it is the same note.
	importTextBatch(t, c, textBatch(c, DupSkip, "", [4]string{"g1", "Default", "¡hola!", "hi"}))
	if n := noteByGUID(t, c, "g1"); n.Fields[0] != "¡hola!" {
		t.Fatalf("fields = %q", n.Fields)
	}

	// A note of another note type with the GUID is not updated.
	b := textBatch(c, DupSkip, "", [4]string{"g1", "Default", "x", "y"})
	for _, nt := range c.NoteTypes() {
		if nt.Name == StockReversed {
			b.NoteTypes[0].ExistingID = nt.ID
		}
	}
	sum := importTextBatch(t, c, b)
	if sum.NotesSkipped != 1 || sum.NotesUpdated != 0 || len(sum.Warnings) != 1 || !strings.Contains(sum.Warnings[0], "different note type") {
		t.Fatalf("other note type summary = %+v", sum)
	}
}

func TestTextImportDeckScope(t *testing.T) {
	// Deck scope (data-import spec): a row whose Front matches a Basic note
	// in another deck is added.
	c, _, _ := testCollection(t)
	importTextBatch(t, c, textBatch(c, DupSkip, "", [4]string{"", "Spanish", "hola", "hello"}))
	sum := importTextBatch(t, c, textBatch(c, DupSkip, DupScopeDeck,
		[4]string{"", "Travel", "hola", "hi"},
		[4]string{"", "Spanish", "hola", "hey"},
		// Rows of the file count too: this one is in Travel already.
		[4]string{"", "Travel", "hola", "howdy"}))
	if sum.NotesAdded != 1 || sum.NotesSkipped != 2 {
		t.Fatalf("deck scope summary = %+v", sum)
	}
	if ids, _ := c.Search("deck:Travel", ColCreated, false); len(ids) != 1 {
		t.Fatalf("%d cards in Travel", len(ids))
	}
	// Updating within the deck scope updates the note in that deck.
	sum = importTextBatch(t, c, textBatch(c, DupUpdate, DupScopeDeck, [4]string{"", "Travel", "hola", "hiya"}))
	if sum.NotesUpdated != 1 {
		t.Fatalf("deck scope update = %+v", sum)
	}
	ids, _ := c.Search("deck:Travel", ColCreated, false)
	if n, _ := c.Note(c.Rows(ids)[0].NoteID); n.Fields[1] != "hiya" {
		t.Fatalf("Travel note = %q", n.Fields)
	}
	// The note type scope (the default) finds the note in any deck.
	sum = importTextBatch(t, c, textBatch(c, DupSkip, DupScopeNoteType, [4]string{"", "Elsewhere", "hola", "hi"}))
	if sum.NotesAdded != 0 || sum.NotesSkipped != 1 {
		t.Fatalf("note type scope summary = %+v", sum)
	}
	if _, err := c.Import(textBatch(c, DupSkip, "everywhere"), nil); err == nil {
		t.Fatal("unknown scope accepted")
	}
}
