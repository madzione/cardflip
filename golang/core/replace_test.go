package core

import (
	"strings"
	"testing"
)

func TestReplaceWithEmptyImport(t *testing.T) {
	c, store, clk := testCollection(t)
	st := c.Settings()
	st.Theme = "dark"
	noErr(t, c.UpdateSettings(st))
	addNote(t, c, "Basic", "old", "")
	// A batch without decks or note types still gives a usable collection:
	// the Default deck and the built-in note types.
	if _, err := c.ReplaceWithImport(&ImportBatch{}, nil); err != nil {
		t.Fatal(err)
	}
	if ds := c.Decks(); len(ds) != 1 || ds[0].Name != DefaultDeckName {
		t.Fatalf("decks = %+v", ds)
	}
	if len(c.NoteTypes()) != len(builtinNoteTypes(BuiltinNames{})) || len(c.notes) != 0 {
		t.Fatalf("%d note types, %d notes", len(c.NoteTypes()), len(c.notes))
	}
	if c.Settings().Theme != "dark" {
		t.Fatal("settings not kept")
	}
	reopen(t, c, store, clk)
}

// If the new collection cannot be saved, the current one is unchanged.
func TestReplaceWithImportAllOrNothing(t *testing.T) {
	c, store, clk := testCollection(t)
	addNote(t, c, "Basic", "kept", "")
	b := &ImportBatch{Decks: []string{"New"}, Media: map[string][]byte{"a.png": []byte("png")}}
	store.FailPut = func(key string) bool { return strings.HasPrefix(key, "m2/") }
	if _, err := c.ReplaceWithImport(b, nil); err == nil {
		t.Fatal("no error")
	}
	store.FailPut = nil
	// The stored collection is unchanged too. (Not compared with reopen:
	// the failed attempt has used up IDs, which is harmless.)
	c2, _, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	for _, col := range []*Collection{c, c2} {
		if len(col.notes) != 1 || len(col.Decks()) != 1 || col.HasMedia("a.png") {
			t.Fatalf("collection changed: %d notes, %d decks", len(col.notes), len(col.Decks()))
		}
	}
}
