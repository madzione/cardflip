package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cardflip/golang/storage"
)

// deletionLog returns the log as "noteType:field|field:tags" strings.
func deletionLog(t *testing.T, c *Collection) []string {
	t.Helper()
	log, err := c.DeletionLog()
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range log {
		out = append(out, e.NoteType+":"+strings.Join(e.Fields, "|")+":"+strings.Join(e.Tags, " "))
	}
	return out
}

func TestDeletionLogDeleteNotes(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic", "hola", "hello")
	b := addNote(t, c, "Basic (and reversed card)", "gato", "cat")
	addNote(t, c, "Basic", "kept", "here")
	if err := c.AddTags([]ID{a}, []string{"spanish", "food"}); err != nil {
		t.Fatal(err)
	}
	clk.advance(time.Hour)
	if err := c.DeleteNotes([]ID{a, b}); err != nil {
		t.Fatal(err)
	}
	want := "[Basic:hola|hello:spanish food Basic (and reversed card):gato|cat:]"
	if got := fmt.Sprint(deletionLog(t, c)); got != want {
		t.Fatalf("log = %s, want %s", got, want)
	}
	log, _ := c.DeletionLog()
	if log[0].NoteID != a || log[1].NoteID != b || !log[0].Time.Equal(clk.now()) || log[0].ID == log[1].ID {
		t.Fatalf("entries = %+v", log)
	}
	// The log is stored with the collection.
	c2 := reopen(t, c, store, clk)
	if got := fmt.Sprint(deletionLog(t, c2)); got != want {
		t.Fatalf("log after reopen = %s", got)
	}
}

// Deleted with a deck (spec): the notes left without cards are logged; a
// note that keeps a card in another deck is not deleted and not logged.
func TestDeletionLogDeleteDeck(t *testing.T) {
	c, _, _ := testCollection(t)
	deck, err := c.CreateDeck("Doomed")
	if err != nil {
		t.Fatal(err)
	}
	basic := noteType(t, c, "Basic").ID
	for i := range 5 {
		if _, err := c.AddNote(NoteInput{TypeID: basic, DeckID: deck, Fields: []string{fmt.Sprint("q", i), "a"}, Tags: []string{"t"}}); err != nil {
			t.Fatal(err)
		}
	}
	split, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic (and reversed card)").ID, DeckID: deck, Fields: []string{"split", "x"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.MoveCards([]ID{c.CardsOfNote(split)[0].ID}, defaultDeck(c)); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteDeck(deck); err != nil {
		t.Fatal(err)
	}
	got := deletionLog(t, c)
	if len(got) != 5 {
		t.Fatalf("log = %v, want 5 entries", got)
	}
	for i, e := range got {
		if e != fmt.Sprintf("Basic:q%d|a:t", i) {
			t.Fatalf("entry %d = %s", i, e)
		}
	}
}

// Undo and redo delete notes too: undoing an added note logs it, redoing a
// deletion logs it again, and undoing a deletion keeps the entry.
func TestDeletionLogUndoRedo(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic", "typed", "carefully")
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(deletionLog(t, c)); got != "[Basic:typed|carefully:]" {
		t.Fatalf("after undoing the add: %s", got)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteNotes([]ID{id}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if len(deletionLog(t, c)) != 2 {
		t.Fatalf("undoing a deletion changed the log: %v", deletionLog(t, c))
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if len(deletionLog(t, c)) != 3 {
		t.Fatalf("redoing a deletion did not log it: %v", deletionLog(t, c))
	}
	// Ops that delete no notes log nothing.
	other := addNote(t, c, "Basic", "x", "y")
	if err := c.AddTags([]ID{other}, []string{"z"}); err != nil {
		t.Fatal(err)
	}
	if len(deletionLog(t, c)) != 3 {
		t.Fatalf("log = %v", deletionLog(t, c))
	}
}

// failingLog fails writes to the deletion log while fail is set, like a
// crash right after the journal record was saved.
type failingLog struct {
	storage.Backend
	fail bool
}

func (f *failingLog) Append(k string, r []byte) error {
	if f.fail && strings.HasPrefix(k, "dl") {
		return errors.New("simulated crash")
	}
	return f.Backend.Append(k, r)
}

// A deletion whose log write did not happen is logged when the journal is
// replayed at the next open, exactly once.
func TestDeletionLogCrashReplay(t *testing.T) {
	store := storage.NewMemoryStore()
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	b := &failingLog{Backend: store.Open()}
	c, _, err := Open(b, Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	id := addNote(t, c, "Basic", "lost", "found")
	b.fail = true
	if err := c.DeleteNotes([]ID{id}); err != nil {
		t.Fatalf("the deletion is durable in the journal: %v", err)
	}
	if got := deletionLog(t, c); len(got) != 0 {
		t.Fatalf("log written despite the failure: %v", got)
	}
	for range 2 {
		c2 := reopen(t, c, store, clk)
		if got := fmt.Sprint(deletionLog(t, c2)); got != "[Basic:lost|found:]" {
			t.Fatalf("log after replay = %s", got)
		}
	}
}

func TestClearDeletionLog(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic", "one", "1")
	b := addNote(t, c, "Basic", "two", "2")
	if err := c.DeleteNotes([]ID{a}); err != nil {
		t.Fatal(err)
	}
	r := c.Revision()
	if err := c.ClearDeletionLog(); err != nil {
		t.Fatal(err)
	}
	if got := deletionLog(t, c); len(got) != 0 || c.Revision() == r {
		t.Fatalf("log after clearing = %v", got)
	}
	// Undo still works across the clear, and logging continues.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteNotes([]ID{b}); err != nil {
		t.Fatal(err)
	}
	c2 := reopen(t, c, store, clk)
	if got := fmt.Sprint(deletionLog(t, c2)); got != "[Basic:two|2:]" {
		t.Fatalf("log after reopen = %s", got)
	}
}
