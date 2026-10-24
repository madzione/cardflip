package core

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// historyBatch imports one Basic note whose card is a review card with n
// history entries, three days apart and ending today, the last one at
// 09:00. Every fourth entry is manual (rating 0).
func historyBatch(c *Collection, guid string, n int) *ImportBatch {
	now := c.Now()
	last := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	var h []ImportReview
	for i := range n {
		r := ImportReview{
			Time: last.AddDate(0, 0, -3*(n-1-i)), Rating: scheduler.Good, StateBefore: scheduler.Review,
			IvlBefore: 3, IvlAfter: 3, DurationMs: 4000, Source: fmt.Sprintf("test:%s:%d", guid, i),
		}
		if i == 0 {
			r.StateBefore, r.IvlBefore = scheduler.New, 0
		}
		if i%4 == 3 {
			r.Rating, r.DurationMs = 0, 0
		}
		h = append(h, r)
	}
	var basic ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	state := scheduler.CardState{State: scheduler.Review, Due: now.AddDate(0, 0, 3), ScheduledDays: 3,
		Memory: &scheduler.Memory{Stability: 3, Difficulty: 5}, Reps: n}
	return &ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "b", ExistingID: basic}},
		Notes: []ImportNote{{TypeKey: "b", GUID: guid, DeckName: "Imported", Fields: []string{guid, "back"},
			Cards: []ImportCard{{Ord: 0, DeckName: "Imported", State: state, History: h}}}},
	}
}

// storedHistory returns every stored review log entry of a card, from all
// months, oldest first.
func storedHistory(t *testing.T, c *Collection, card ID) []ReviewEntry {
	t.Helper()
	months, err := c.RevlogMonths()
	if err != nil {
		t.Fatal(err)
	}
	var out []ReviewEntry
	for _, m := range months {
		entries, err := c.ReadRevlog(m)
		if err != nil {
			t.Fatal(err)
		}
		for _, raw := range entries {
			var e ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				t.Fatal(err)
			}
			if card == 0 || e.CardID == card {
				out = append(out, e)
			}
		}
	}
	return out
}

func importedCard(t *testing.T, c *Collection, guid string) ID {
	t.Helper()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, n := range c.notes {
		if n.GUID == guid {
			return c.cardsByNote[n.ID][0]
		}
	}
	t.Fatalf("no note %s", guid)
	return 0
}

// History kept (data-import spec): a card reviewed 12 times keeps the 12
// entries with their dates and ratings, in memory for the loaded months and
// stored for all.
func TestImportHistoryKept(t *testing.T) {
	c, store, clk := testCollection(t)
	b := historyBatch(c, "h1", 12)
	sum, err := c.Import(b, nil)
	if err != nil || sum.Reviews != 12 || sum.CardsAdded != 1 {
		t.Fatalf("import: %+v, %v", sum, err)
	}
	card := importedCard(t, c, "h1")
	want := b.Notes[0].Cards[0].History
	got := storedHistory(t, c, card)
	if len(got) != 12 {
		t.Fatalf("stored history = %d entries", len(got))
	}
	for i, e := range got {
		w := want[i]
		if !e.Time.Equal(w.Time) || e.Rating != w.Rating || e.StateBefore != w.StateBefore || e.IvlBefore != w.IvlBefore ||
			e.IvlAfter != w.IvlAfter || e.DurationMs != w.DurationMs || e.Source != w.Source {
			t.Fatalf("entry %d = %+v, want %+v", i, e, w)
		}
	}
	// In memory: the entries of this month and the previous one.
	inMonths := 0
	for _, w := range want {
		if m := monthKey(w.Time); m == monthKey(clk.now()) || m == monthKey(clk.now().AddDate(0, 0, -2)) {
			inMonths++
		}
	}
	if n := len(c.ReviewsOfCard(card)); n != inMonths || n == 0 || n == 12 {
		t.Fatalf("in memory: %d entries, want %d", n, inMonths)
	}
	c2 := reopen(t, c, store, clk)
	if n := len(c2.ReviewsOfCard(card)); n != inMonths {
		t.Fatalf("after reopen: %d entries in memory, want %d", n, inMonths)
	}
	if n := len(storedHistory(t, c2, card)); n != 12 {
		t.Fatalf("after reopen: %d entries stored", n)
	}
}

// Import as new (data-import spec): cards without history add none.
func TestImportWithoutHistory(t *testing.T) {
	c, _, _ := testCollection(t)
	b := historyBatch(c, "h1", 12)
	b.Notes[0].Cards[0].History = nil
	b.Notes[0].Cards[0].State = scheduler.CardState{State: scheduler.New}
	sum, err := c.Import(b, nil)
	if err != nil || sum.Reviews != 0 || sum.CardsAdded != 1 {
		t.Fatalf("import: %+v, %v", sum, err)
	}
	if n := len(storedHistory(t, c, 0)); n != 0 {
		t.Fatalf("stored history = %d entries", n)
	}
}

// Re-import (data-import spec): importing the same package again adds no
// entries; neither does a batch whose entries are already stored for a card
// that exists.
func TestImportHistoryTwice(t *testing.T) {
	c, _, _ := testCollection(t)
	if _, err := c.Import(historyBatch(c, "h1", 12), nil); err != nil {
		t.Fatal(err)
	}
	card := importedCard(t, c, "h1")
	sum, err := c.Import(historyBatch(c, "h1", 12), nil)
	if err != nil || sum.Reviews != 0 || sum.NotesSkipped != 1 {
		t.Fatalf("second import: %+v, %v", sum, err)
	}
	// The same sources on another note: the stored entries' card exists.
	b := historyBatch(c, "h2", 12)
	for i := range b.Notes[0].Cards[0].History {
		b.Notes[0].Cards[0].History[i].Source = fmt.Sprintf("test:h1:%d", i)
	}
	if sum, err = c.Import(b, nil); err != nil || sum.Reviews != 0 || sum.NotesAdded != 1 {
		t.Fatalf("third import: %+v, %v", sum, err)
	}
	if n := len(storedHistory(t, c, card)); n != 12 {
		t.Fatalf("history = %d entries", n)
	}
	if n := len(storedHistory(t, c, 0)); n != 12 {
		t.Fatalf("all history = %d entries", n)
	}
}

// When the note was deleted, a re-import gives the new card the history:
// the old entries are replaced, not duplicated. Undoing that import brings
// the old entries back, so undoing the deletion restores the card with its
// history.
func TestImportHistoryAfterDelete(t *testing.T) {
	c, _, _ := testCollection(t)
	if _, err := c.Import(historyBatch(c, "h1", 12), nil); err != nil {
		t.Fatal(err)
	}
	old := importedCard(t, c, "h1")
	var note ID
	if card, err := c.Card(old); err == nil {
		note = card.NoteID
	}
	if err := c.DeleteNotes([]ID{note}); err != nil {
		t.Fatal(err)
	}
	sum, err := c.Import(historyBatch(c, "h1", 12), nil)
	if err != nil || sum.Reviews != 12 {
		t.Fatalf("re-import: %+v, %v", sum, err)
	}
	card := importedCard(t, c, "h1")
	if n := len(storedHistory(t, c, card)); n != 12 {
		t.Fatalf("new card history = %d entries", n)
	}
	if n := len(storedHistory(t, c, 0)); n != 12 {
		t.Fatalf("all history = %d entries", n)
	}
	if len(c.ReviewsOfCard(old)) != 0 || len(c.ReviewsOfCard(card)) == 0 {
		t.Fatal("history in memory not moved to the new card")
	}
	for _, step := range []string{"Import", "Delete 1 note"} {
		if name, err := c.Undo(); err != nil || usermsg.Plain(name) != step {
			t.Fatalf("undo = %q, %v", name, err)
		}
	}
	if n := len(storedHistory(t, c, old)); n != 12 || len(storedHistory(t, c, 0)) != 12 {
		t.Fatalf("after undo: old card history = %d entries", n)
	}
	if len(c.ReviewsOfCard(old)) == 0 {
		t.Fatal("after undo: history not in memory")
	}
}

// Imported entries are not undone as answers, and manual entries are not
// counted as today's answers.
func TestImportedHistoryNotAnswers(t *testing.T) {
	c, _, _ := testCollection(t)
	deck, err := c.CreateDeck("Imported")
	if err != nil {
		t.Fatal(err)
	}
	opts := DefaultDeckOptions()
	opts.NewPerDay = 1
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	addNoteIn(t, c, deck, "Basic", "new card", "")
	// Today's entry is manual (12 entries: the last one, index 11).
	if _, err := c.Import(historyBatch(c, "h1", 12), nil); err != nil {
		t.Fatal(err)
	}
	card := importedCard(t, c, "h1")
	reviews := c.ReviewsOfCard(card)
	latest := reviews[len(reviews)-1]
	if latest.Rating != 0 {
		t.Fatalf("latest entry = %+v", latest)
	}
	if got := c.DeckCounts()[deck].New; got != 1 {
		t.Fatalf("new count = %d, want 1: a manual entry is no answer", got)
	}
	if _, err := c.UndoReview(latest.ID); !IsUserError(err) {
		t.Fatalf("undo imported entry: %v", err)
	}
	if _, err := c.UndoReview(reviews[len(reviews)-2].ID); !IsUserError(err) {
		t.Fatalf("undo imported answer: %v", err)
	}
	// An imported answer today to a New card counts, as in Anki.
	b := historyBatch(c, "h2", 1)
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	if got := c.DeckCounts()[deck].New; got != 0 {
		t.Fatalf("new count = %d, want 0 after an imported answer to a new card today", got)
	}
}
