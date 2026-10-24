package core

import (
	"cardflip/golang/internal/usermsg"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStockNoteTypes(t *testing.T) {
	c, _, _ := testCollection(t)
	want := "Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer), Cloze, Image Occlusion"
	if got := strings.Join(StockNoteTypes(), ", "); got != want {
		t.Fatalf("stock = %q", got)
	}
	// Fresh collections keep exactly the four built-ins, defined as their
	// stock types.
	types := c.NoteTypes()
	if len(types) != 4 {
		t.Fatalf("%d note types", len(types))
	}
	for _, nt := range types {
		stock := stockNoteType(nt.Name)
		stock.ID = nt.ID
		if !reflect.DeepEqual(*stock, nt) {
			t.Errorf("built-in %q = %+v, want %+v", nt.Name, nt, *stock)
		}
	}
	typeIn := stockNoteType(StockTypeIn)
	if typeIn.Templates[0].QFmt != "{{Front}}\n\n{{type:Back}}" {
		t.Fatalf("type-in front = %q", typeIn.Templates[0].QFmt)
	}
}

func TestAddNoteType(t *testing.T) {
	c, store, clk := testCollection(t)
	// Add from stock (spec).
	id, err := c.AddNoteType(StockBasic, " Vocabulary ")
	if err != nil {
		t.Fatal(err)
	}
	nt := noteType(t, c, "Vocabulary")
	if nt.ID != id || strings.Join(nt.Fields, ",") != "Front,Back" || len(nt.Templates) != 1 || nt.Kind != Standard {
		t.Fatalf("Vocabulary = %+v", nt)
	}
	if n := c.NoteCount(id); n != 0 {
		t.Fatalf("%d notes", n)
	}
	cloze, err := c.AddNoteType(StockCloze, "My cloze")
	if err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(cloze); nt.Kind != Cloze {
		t.Fatalf("cloze kind = %q", nt.Kind)
	}
	// Name taken (spec), ignoring case.
	_, err = c.AddNoteType(StockBasic, "basic")
	mustUserErr(t, err, `"Basic" already exists`)
	_, err = c.AddNoteType(StockBasic, "  ")
	mustUserErr(t, err, "Please enter a note type name")
	_, err = c.AddNoteType("Fancy", "Fancy")
	mustUserErr(t, err, "Unknown stock note type")
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Add note type" {
		t.Fatalf("undo = %q", undo)
	}
	reopen(t, c, store, clk)
}

func TestCloneNoteType(t *testing.T) {
	c, store, clk := testCollection(t)
	imported := NoteType{
		Name: "Japanese", Kind: Standard, Fields: []string{"Expression", "Reading", "Meaning", "Notes"},
		Templates: []CardTemplate{
			{Name: "Recognition", QFmt: "{{Expression}}", AFmt: "{{FrontSide}}<hr id=answer>{{Reading}} {{Meaning}}"},
			{Name: "Recall", QFmt: "{{Meaning}}", AFmt: "{{FrontSide}}<hr id=answer>{{Expression}}"},
		},
		CSS: ".card { color: teal; }", Source: "anki:123",
	}
	_, err := c.Import(&ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "j", Type: imported}},
		Notes: []ImportNote{{TypeKey: "j", GUID: "g1", DeckName: "Japanese",
			Fields: []string{"猫", "ねこ", "cat", ""}}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	orig := noteType(t, c, "Japanese")
	// Clone imported type (spec).
	id, err := c.CloneNoteType(orig.ID, "Experiment")
	if err != nil {
		t.Fatal(err)
	}
	clone := noteType(t, c, "Experiment")
	if clone.ID != id || clone.ID == orig.ID {
		t.Fatalf("clone ID %d, original %d", clone.ID, orig.ID)
	}
	if clone.Kind != orig.Kind || !reflect.DeepEqual(clone.Fields, orig.Fields) ||
		!reflect.DeepEqual(clone.Templates, orig.Templates) || clone.CSS != orig.CSS {
		t.Fatalf("clone = %+v, original %+v", clone, orig)
	}
	if clone.Source != "" {
		t.Fatalf("clone source = %q", clone.Source)
	}
	if n := c.NoteCount(id); n != 0 {
		t.Fatalf("clone has %d notes", n)
	}
	_, err = c.CloneNoteType(orig.ID, "JAPANESE")
	mustUserErr(t, err, "already exists")
	_, err = c.CloneNoteType(999999, "Other")
	mustUserErr(t, err, "no longer exists")
	reopen(t, c, store, clk)
}

func TestDeleteNoteType(t *testing.T) {
	c, store, clk := testCollection(t)
	reversed := noteType(t, c, "Basic (and reversed card)")
	for i := range 10 {
		addNote(t, c, reversed.Name, fmt.Sprint("front ", i), fmt.Sprint("back ", i))
	}
	kept := addNote(t, c, "Basic", "keep", "me")
	if s := c.Settings(); s.LastNoteTypeID != noteType(t, c, "Basic").ID {
		t.Fatalf("last note type = %d", s.LastNoteTypeID)
	}
	addNote(t, c, reversed.Name, "last", "used")
	info, err := c.NoteTypeDeletionInfo(reversed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if info != (NoteTypeDeletion{Notes: 11, Cards: 22}) {
		t.Fatalf("info = %+v", info)
	}
	// Delete with notes (spec).
	if err := c.DeleteNoteType(reversed.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := c.NoteType(reversed.ID); err == nil {
		t.Fatal("note type still exists")
	}
	if len(c.notes) != 1 || len(c.cards) != 1 {
		t.Fatalf("%d notes %d cards left", len(c.notes), len(c.cards))
	}
	if _, err := c.Note(kept); err != nil {
		t.Fatal(err)
	}
	// The editor's remembered note type pointed at the deleted one.
	if s := c.Settings(); s.LastNoteTypeID != 0 {
		t.Fatalf("last note type = %d", s.LastNoteTypeID)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Delete note type" {
		t.Fatalf("undo = %q", undo)
	}
	log, err := c.DeletionLog()
	if err != nil || len(log) != 11 || log[0].NoteType != reversed.Name {
		t.Fatalf("deletion log = %+v, %v", log, err)
	}

	// Undo brings the note type, its notes and cards back; redo deletes
	// them again.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if c.NoteCount(reversed.ID) != 11 || len(c.cards) != 23 {
		t.Fatalf("after undo: %d notes, %d cards", c.NoteCount(reversed.ID), len(c.cards))
	}
	if s := c.Settings(); s.LastNoteTypeID != reversed.ID {
		t.Fatalf("last note type after undo = %d", s.LastNoteTypeID)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	c = reopen(t, c, store, clk)
	if _, err := c.NoteType(reversed.ID); err == nil {
		t.Fatal("note type back after reopening")
	}
	if len(c.notes) != 1 || len(c.cards) != 1 {
		t.Fatalf("after reopening: %d notes %d cards", len(c.notes), len(c.cards))
	}
}

func TestDeleteLastNoteTypeRejected(t *testing.T) {
	c, store, clk := testCollection(t)
	_, err := c.NoteTypeDeletionInfo(999999)
	mustUserErr(t, err, "no longer exists")
	for _, nt := range c.NoteTypes()[1:] {
		if err := c.DeleteNoteType(nt.ID); err != nil {
			t.Fatal(err)
		}
	}
	// Last note type (spec).
	last := c.NoteTypes()[0]
	addNote(t, c, last.Name, "a", "b")
	mustUserErr(t, c.DeleteNoteType(last.ID), "last note type cannot be deleted")
	if len(c.NoteTypes()) != 1 || c.NoteCount(last.ID) != 1 {
		t.Fatalf("note types %v, %d notes", c.NoteTypes(), c.NoteCount(last.ID))
	}
	reopen(t, c, store, clk)
}

func TestOptionalReversedCard(t *testing.T) {
	c, store, clk := testCollection(t)
	if _, err := c.AddNoteType(StockOptionalReversed, "Optional"); err != nil {
		t.Fatal(err)
	}
	// Reverse on demand (spec).
	both := addNote(t, c, "Optional", "hola", "hello", "y")
	if got := questions(t, c, both); fmt.Sprint(got) != "[hola hello]" {
		t.Fatalf("questions = %v", got)
	}
	one := addNote(t, c, "Optional", "adiós", "goodbye", "")
	if got := questions(t, c, one); fmt.Sprint(got) != "[adiós]" {
		t.Fatalf("questions = %v", got)
	}
	// Filling Add Reverse later adds the reverse card.
	if err := c.UpdateNote(one, []string{"adiós", "goodbye", "y"}, nil); err != nil {
		t.Fatal(err)
	}
	if got := questions(t, c, one); fmt.Sprint(got) != "[adiós goodbye]" {
		t.Fatalf("questions after filling Add Reverse = %v", got)
	}
	reopen(t, c, store, clk)
}

// Every change to a note type, and creating one, records when it happened.
func TestNoteTypeModified(t *testing.T) {
	c, _, clk := testCollection(t)
	id, err := c.AddNoteType(StockBasic, "Vocabulary")
	if err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(id); !nt.Modified.Equal(clk.t) {
		t.Fatalf("added: modified %v", nt.Modified)
	}
	changes := []func() error{
		func() error { return c.RenameNoteType(id, "Words") },
		func() error { return c.AddField(id, "Extra") },
		func() error { return c.RenameField(id, 2, "Notes") },
		func() error { return c.SetFieldOptions(id, 0, FieldOptions{Sticky: true}) },
		func() error { return c.RemoveField(id, 2) },
		func() error {
			return c.AddTemplate(id, CardTemplate{Name: "Card 2", QFmt: "{{Back}}", AFmt: "{{Front}}"})
		},
		func() error {
			return c.UpdateTemplate(id, 1, CardTemplate{Name: "Reverse", QFmt: "{{Back}}?", AFmt: "{{Front}}"})
		},
		func() error { return c.RemoveTemplate(id, 1) },
		func() error { return c.SetNoteTypeCSS(id, ".card {}") },
	}
	for i, change := range changes {
		clk.t = clk.t.Add(time.Minute)
		if err := change(); err != nil {
			t.Fatalf("change %d: %v", i, err)
		}
		if nt, _ := c.NoteType(id); !nt.Modified.Equal(clk.t) {
			t.Fatalf("change %d: modified %v, want %v", i, nt.Modified, clk.t)
		}
	}
	clk.t = clk.t.Add(time.Minute)
	clone, err := c.CloneNoteType(id, "Copy")
	if err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(clone); !nt.Modified.Equal(clk.t) {
		t.Fatalf("clone: modified %v", nt.Modified)
	}
}
