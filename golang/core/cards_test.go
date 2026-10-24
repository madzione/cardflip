package core

import (
	"cardflip/golang/internal/usermsg"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

func addNoteIn(t *testing.T, c *Collection, deck ID, typeName string, fields ...string) ID {
	t.Helper()
	id, err := c.AddNote(NoteInput{TypeID: noteType(t, c, typeName).ID, DeckID: deck, Fields: fields})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestDeleteDeckWithCards(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	verbs, _ := c.CreateDeck("Spanish::Verbs")
	keep := defaultDeck(c)
	for i := 0; i < 4; i++ {
		addNoteIn(t, c, spanish, "Basic (and reversed card)", "a", "b") // 2 cards each
	}
	for i := 0; i < 4; i++ {
		addNoteIn(t, c, verbs, "Basic", "v", "w") // 1 card each
	}
	// A note with one card inside the deleted tree and one outside.
	split := addNoteIn(t, c, keep, "Basic (and reversed card)", "x", "y")
	cards := c.CardsOfNote(split)
	if err := c.MoveCards([]ID{cards[0].ID}, verbs); err != nil {
		t.Fatal(err)
	}
	survivor := addNoteIn(t, c, keep, "Basic", "stays", "here")

	// Delete deck with cards (spec): the confirmation mentions 12 cards.
	info, err := c.DeckDeletionInfo(spanish)
	if err != nil {
		t.Fatal(err)
	}
	if info.Cards != 13 || info.Decks != 2 {
		// 8 + 4 + the moved card of the split note.
		t.Fatalf("info = %+v", info)
	}
	if err := c.DeleteDeck(spanish); err != nil {
		t.Fatal(err)
	}
	if got := deckNames(c); got != "Default" {
		t.Fatalf("decks = %q", got)
	}
	// Orphaned notes are gone; the split note keeps its other card.
	if len(c.notes) != 2 {
		t.Fatalf("%d notes left", len(c.notes))
	}
	if got := c.CardsOfNote(split); len(got) != 1 || got[0].ID != cards[1].ID {
		t.Fatalf("split note cards = %+v", got)
	}
	if _, err := c.Note(survivor); err != nil {
		t.Fatal(err)
	}
	reopen(t, c, store, clk)
}

func TestDeckDeletionInfoCountsTwelve(t *testing.T) {
	c, _, _ := testCollection(t)
	parent, _ := c.CreateDeck("P")
	child, _ := c.CreateDeck("P::C")
	for i := 0; i < 5; i++ {
		addNoteIn(t, c, parent, "Basic", "q", "a")
	}
	for i := 0; i < 7; i++ {
		addNoteIn(t, c, child, "Basic", "q", "a")
	}
	if info, _ := c.DeckDeletionInfo(parent); info.Cards != 12 {
		t.Fatalf("cards = %d, want 12", info.Cards)
	}
}

func TestMovePreservesSchedule(t *testing.T) {
	c, store, clk := testCollection(t)
	target, _ := c.CreateDeck("Target")
	id := addNote(t, c, "Basic", "q", "a")
	card := c.CardsOfNote(id)[0]
	due := clk.t.AddDate(0, 0, 10)
	card.CardState = scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: 10,
		Memory: &scheduler.Memory{Stability: 9, Difficulty: 4}, Reps: 2, LastReview: clk.t}
	c.mu.Lock()
	c.commit(&op{Kind: "test", Cards: []*Card{&card}})
	c.mu.Unlock()

	// Move preserves schedule (spec).
	if err := c.MoveCards([]ID{card.ID}, target); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(card.ID)
	if got.DeckID != target || !got.Due.Equal(due) || got.State != scheduler.Review || got.Memory.Stability != 9 {
		t.Fatalf("moved card = %+v", got)
	}
	mustUserErr(t, c.MoveCards([]ID{card.ID}, 12345), "no longer exists")
	reopen(t, c, store, clk)
}

func TestSuspendAndBury(t *testing.T) {
	c, store, clk := testCollection(t)
	id := addNote(t, c, "Basic", "q", "a")
	card := c.CardsOfNote(id)[0]
	if err := c.SetSuspended([]ID{card.ID}, true); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Card(card.ID); !got.Suspended {
		t.Fatal("not suspended")
	}
	c.SetSuspended([]ID{card.ID}, false)
	if got, _ := c.Card(card.ID); got.Suspended {
		t.Fatal("still suspended")
	}

	// Buried card returns (spec): excluded for the rest of the day, eligible
	// again after the next rollover (04:00).
	clk.t = time.Date(2026, 10, 5, 23, 0, 0, 0, time.Local)
	if err := c.BuryCards([]ID{card.ID}); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(card.ID)
	for _, at := range []time.Time{
		time.Date(2026, 10, 5, 23, 30, 0, 0, time.Local),
		time.Date(2026, 10, 6, 3, 59, 0, 0, time.Local),
	} {
		if !isBuried(&got, at) {
			t.Fatalf("not buried at %v", at)
		}
	}
	if isBuried(&got, time.Date(2026, 10, 6, 4, 0, 0, 0, time.Local)) {
		t.Fatal("still buried after rollover")
	}
	reopen(t, c, store, clk)
}

func TestResetCards(t *testing.T) {
	c, store, clk := testCollection(t)
	a := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0]
	b := c.CardsOfNote(addNote(t, c, "Basic", "b", "2"))[0]
	a.CardState = scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 3),
		Memory: &scheduler.Memory{Stability: 3, Difficulty: 5}, Reps: 4, Lapses: 1}
	c.mu.Lock()
	c.commit(&op{Kind: "test", Cards: []*Card{&a}})
	c.mu.Unlock()
	// Reset to new (card-browser spec): New, memory cleared, counts kept,
	// and without a known original position at the end.
	if err := c.ResetCards([]ID{a.ID}, ResetOptions{RestorePosition: true}); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(a.ID)
	if got.State != scheduler.New || got.Memory != nil || !got.Due.IsZero() || got.Reps != 4 || got.Lapses != 1 ||
		got.NewPos <= b.NewPos {
		t.Fatalf("reset card = %+v (b pos %d)", got, b.NewPos)
	}
	// Every reset adds a manual entry to the card's history.
	h := c.ReviewsOfCard(a.ID)
	if len(h) != 1 || h[0].Rating != 0 || h[0].StateBefore != scheduler.Review || h[0].IvlAfter != 0 {
		t.Fatalf("history = %+v", h)
	}
	reopen(t, c, store, clk)
}

// answerTimes answers a card Good n times, a day apart.
func answerTimes(t *testing.T, c *Collection, clk *clock, card ID, n int) {
	t.Helper()
	for range n {
		if _, err := c.Answer(card, scheduler.Good, time.Second); err != nil {
			t.Fatal(err)
		}
		clk.advance(24 * time.Hour)
	}
}

// Restore position (collection spec).
func TestResetRestoresPosition(t *testing.T) {
	c, store, clk := testCollection(t)
	var cards []Card
	for _, f := range []string{"a", "b", "c"} {
		cards = append(cards, c.CardsOfNote(addNote(t, c, "Basic", f, "x"))[0])
	}
	third := cards[2]
	answerTimes(t, c, clk, third.ID, 3)
	later := c.CardsOfNote(addNote(t, c, "Basic", "d", "x"))[0]

	if err := c.ResetCards([]ID{third.ID}, ResetOptions{RestorePosition: true}); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(third.ID)
	if got.State != scheduler.New || got.NewPos != third.NewPos || got.NewPos >= later.NewPos {
		t.Fatalf("restored card pos %d (was %d, later card %d)", got.NewPos, third.NewPos, later.NewPos)
	}
	q, err := c.Queue(defaultDeck(c))
	if err != nil {
		t.Fatal(err)
	}
	if i, j := slices.Index(q.New, third.ID), slices.Index(q.New, later.ID); i < 0 || j < 0 || i > j {
		t.Fatalf("new queue %v: reset card %d not before %d", q.New, third.ID, later.ID)
	}

	// Without restoring, the card goes after all new cards; its original
	// position is still known for a later reset.
	answerTimes(t, c, clk, third.ID, 1)
	if err := c.ResetCards([]ID{third.ID}, ResetOptions{}); err != nil {
		t.Fatal(err)
	}
	got, _ = c.Card(third.ID)
	if got.NewPos <= later.NewPos || got.OriginalPos == nil || *got.OriginalPos != third.NewPos {
		t.Fatalf("reset without restoring: %+v", got)
	}
	answerTimes(t, c, clk, third.ID, 1)
	c.ResetCards([]ID{third.ID}, ResetOptions{RestorePosition: true})
	if got, _ = c.Card(third.ID); got.NewPos != third.NewPos {
		t.Fatalf("second restore: pos %d, want %d", got.NewPos, third.NewPos)
	}
	reopen(t, c, store, clk)
}

// Reset counts (collection spec).
func TestResetCounts(t *testing.T) {
	c, store, clk := testCollection(t)
	card := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0]
	answerTimes(t, c, clk, card.ID, 2)
	cur, _ := c.Card(card.ID)
	cur.Reps, cur.Lapses = 12, 4
	c.mu.Lock()
	c.commit(&op{Kind: "test", Cards: []*Card{&cur}})
	c.mu.Unlock()

	if err := c.ResetCards([]ID{card.ID}, ResetOptions{ResetCounts: true}); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(card.ID)
	if got.State != scheduler.New || got.Reps != 0 || got.Lapses != 0 {
		t.Fatalf("reset card = %+v", got)
	}
	h, err := c.CardHistory(card.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 3 || h[0].Rating != scheduler.Good || h[1].Rating != scheduler.Good || h[2].Rating != 0 {
		t.Fatalf("history = %+v", h)
	}
	// Undo brings back the card and removes the manual entry.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Card(card.ID); got.Reps != 12 || got.Lapses != 4 || got.State == scheduler.New {
		t.Fatalf("after undo = %+v", got)
	}
	if h := c.ReviewsOfCard(card.ID); len(h) != 2 {
		t.Fatalf("history after undo = %+v", h)
	}
	reopen(t, c, store, clk)
}

// Every op announces a change to subscribers.
func TestChangeEventsForEveryOp(t *testing.T) {
	c, _, _ := testCollection(t)
	events, cancel := c.Subscribe()
	defer cancel()
	expect := func(name string, check func(Change) bool) {
		t.Helper()
		select {
		case ch := <-events:
			if !check(ch) {
				t.Fatalf("%s: change = %+v", name, ch)
			}
		default:
			t.Fatalf("%s: no change event", name)
		}
	}
	basic := noteType(t, c, "Basic")

	deck, _ := c.CreateDeck("D")
	expect("create deck", func(ch Change) bool { return ch.Decks })
	c.RenameDeck(deck, "E")
	expect("rename deck", func(ch Change) bool { return ch.Decks })
	c.SetDeckOptions(deck, deckOpts(1, 1, 0.8))
	expect("deck options", func(ch Change) bool { return ch.Decks })
	c.SetDeckCollapsed(deck, true)
	expect("collapse", func(ch Change) bool { return ch.Decks })
	id := addNoteIn(t, c, deck, "Basic", "q", "a")
	expect("add note", func(ch Change) bool { return ch.Notes && ch.Cards })
	c.UpdateNote(id, []string{"q2", "a"}, nil)
	expect("update note", func(ch Change) bool { return ch.Notes })
	card := c.CardsOfNote(id)[0].ID
	c.MoveCards([]ID{card}, defaultDeck(c))
	expect("move", func(ch Change) bool { return ch.Cards })
	c.SetSuspended([]ID{card}, true)
	expect("suspend", func(ch Change) bool { return ch.Cards })
	c.BuryCards([]ID{card})
	expect("bury", func(ch Change) bool { return ch.Cards })
	c.ResetCards([]ID{card}, ResetOptions{})
	expect("reset", func(ch Change) bool { return ch.Cards })
	c.AddTags([]ID{id}, []string{"t"})
	expect("add tags", func(ch Change) bool { return ch.Notes })
	c.RemoveTags([]ID{id}, []string{"t"})
	expect("remove tags", func(ch Change) bool { return ch.Notes })
	c.AddField(basic.ID, "Extra")
	expect("add field", func(ch Change) bool { return ch.NoteTypes })
	c.RenameField(basic.ID, 2, "More")
	expect("rename field", func(ch Change) bool { return ch.NoteTypes })
	c.RemoveField(basic.ID, 2)
	expect("remove field", func(ch Change) bool { return ch.NoteTypes && ch.Notes })
	c.AddTemplate(basic.ID, CardTemplate{Name: "R", QFmt: "{{Back}}", AFmt: "{{Front}}"})
	expect("add template", func(ch Change) bool { return ch.NoteTypes && ch.Cards })
	c.UpdateTemplate(basic.ID, 1, CardTemplate{Name: "R", QFmt: "{{Back}}!", AFmt: "{{Front}}"})
	expect("update template", func(ch Change) bool { return ch.NoteTypes })
	c.RemoveTemplate(basic.ID, 1)
	expect("remove template", func(ch Change) bool { return ch.NoteTypes && ch.Cards })
	c.SetNoteTypeCSS(basic.ID, ".card{}")
	expect("css", func(ch Change) bool { return ch.NoteTypes })
	c.RenameNoteType(basic.ID, "B")
	expect("rename note type", func(ch Change) bool { return ch.NoteTypes })
	s := c.Settings()
	s.Theme = ThemeDark
	c.UpdateSettings(s)
	expect("settings", func(ch Change) bool { return ch.Settings })
	c.DeleteNotes([]ID{id})
	expect("delete notes", func(ch Change) bool { return ch.Notes && ch.Cards })
	c.DeleteDeck(deck)
	expect("delete deck", func(ch Change) bool { return ch.Decks })

	// Rejected or no-op changes announce nothing.
	c.SetDeckOptions(defaultDeck(c), deckOpts(0, 0, 0.1))
	c.UpdateSettings(c.Settings())
	select {
	case ch := <-events:
		t.Fatalf("unexpected change %+v", ch)
	default:
	}
}

// A subscriber that does not keep up gets the pending changes merged.
func TestSlowSubscriberChangesMerge(t *testing.T) {
	c, _, _ := testCollection(t)
	events, cancel := c.Subscribe()
	c.CreateDeck("A")
	addNote(t, c, "Basic", "q", "a")
	s := c.Settings()
	s.Theme = ThemeLight
	c.UpdateSettings(s)
	ch := <-events
	if !ch.Decks || !ch.Notes || !ch.Settings {
		t.Fatalf("merged change = %+v", ch)
	}
	cancel()
	if _, ok := <-events; ok {
		t.Fatal("channel not closed after cancel")
	}
	cancel() // idempotent
	c.Close()
}

func TestRemoveEmptyCards(t *testing.T) {
	c, store, clk := testCollection(t)
	var notes []ID
	for i := 0; i < 3; i++ {
		notes = append(notes, addNote(t, c, "Basic (and reversed card)", fmt.Sprint("front ", i), "back"))
	}
	keep := addNote(t, c, "Basic (and reversed card)", "untouched", "back")
	var reverse []ID
	for _, id := range notes {
		if err := c.UpdateNote(id, []string{"front", ""}, nil); err != nil {
			t.Fatal(err)
		}
		reverse = append(reverse, c.CardsOfNote(id)[1].ID)
	}
	empty := c.EmptyCards()
	if empty.Total != 3 || empty.Deletable != 3 || len(empty.Notes) != 3 {
		t.Fatalf("empty cards = %+v", empty)
	}
	for i, en := range empty.Notes {
		if en.NoteID != notes[i] || en.AllEmpty || len(en.Cards) != 1 || en.Cards[0].CardID != reverse[i] || en.Cards[0].Name != "Card 2" {
			t.Fatalf("note %d = %+v", i, en)
		}
	}
	// Remove empty cards (spec): the 3 cards are deleted, every note keeps
	// its other cards, in one change.
	var ids []ID
	for _, en := range empty.Notes {
		ids = append(ids, en.Cards[0].CardID)
	}
	sub, stop := c.Subscribe()
	defer stop()
	if err := c.DeleteEmptyCards(ids); err != nil {
		t.Fatal(err)
	}
	if ch := <-sub; !ch.Cards {
		t.Fatalf("change = %+v", ch)
	}
	for _, id := range notes {
		if cards := c.CardsOfNote(id); len(cards) != 1 || cards[0].Ord != 0 {
			t.Fatalf("cards of note = %+v", cards)
		}
	}
	if n := len(c.CardsOfNote(keep)); n != 2 {
		t.Fatalf("untouched note has %d cards", n)
	}
	if empty := c.EmptyCards(); empty.Total != 0 {
		t.Fatalf("empty cards after delete = %+v", empty)
	}
	c = reopen(t, c, store, clk)
	if n := len(c.cards); n != 5 {
		t.Fatalf("cards after reopen = %d", n)
	}
}

func TestAllCardsEmpty(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	id := addNote(t, c, "Basic", "front", "")
	other := addNote(t, c, "Basic (and reversed card)", "a", "b")
	if err := c.UpdateTemplate(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Back}}", AFmt: "{{Back}}"}); err != nil {
		t.Fatal(err)
	}
	// All cards empty (spec): the list marks the note, deleting keeps the card.
	empty := c.EmptyCards()
	if empty.Total != 1 || empty.Deletable != 0 || !empty.Notes[0].AllEmpty || empty.Notes[0].Text != "front" {
		t.Fatalf("empty cards = %+v", empty)
	}
	card := empty.Notes[0].Cards[0].CardID
	if !c.IsEmptyCard(card) || c.IsEmptyCard(c.CardsOfNote(other)[0].ID) {
		t.Fatal("IsEmptyCard")
	}
	// A card that is not empty is never deleted.
	if err := c.DeleteEmptyCards([]ID{card, c.CardsOfNote(other)[0].ID}); err != nil {
		t.Fatal(err)
	}
	if len(c.CardsOfNote(id)) != 1 || len(c.CardsOfNote(other)) != 2 {
		t.Fatal("a card was deleted")
	}
	if _, err := c.Note(id); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveTemplateLeavingNoCardsRejected(t *testing.T) {
	c, _, _ := testCollection(t)
	rev := noteType(t, c, "Basic (and reversed card)")
	addNote(t, c, "Basic (and reversed card)", "front", "")
	mustUserErr(t, c.RemoveTemplate(rev.ID, 0), "without cards")
	if nt, _ := c.NoteType(rev.ID); len(nt.Templates) != 2 {
		t.Fatal("rejected change applied")
	}
}

func TestSetFlag(t *testing.T) {
	c, store, clk := testCollection(t)
	id := addNote(t, c, "Basic (and reversed card)", "q", "a")
	cards := c.CardsOfNote(id)
	fwd, rev := cards[0], cards[1]
	due := clk.t.AddDate(0, 0, 5)
	fwd.CardState = scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: 5,
		Memory: &scheduler.Memory{Stability: 5, Difficulty: 3}, Reps: 1, LastReview: clk.t}
	c.mu.Lock()
	c.commit(&op{Kind: "test", Cards: []*Card{&fwd}})
	c.mu.Unlock()
	flags := func() (int, int) {
		a, _ := c.Card(fwd.ID)
		b, _ := c.Card(rev.ID)
		return a.Flag, b.Flag
	}

	// Flag one card of a note (spec): the reverse card stays unflagged and
	// the scheduling is unchanged.
	if err := c.SetFlag([]ID{fwd.ID}, 1); err != nil {
		t.Fatal(err)
	}
	if a, b := flags(); a != 1 || b != 0 {
		t.Fatalf("flags = %d, %d", a, b)
	}
	got, _ := c.Card(fwd.ID)
	if !reflect.DeepEqual(got.CardState, fwd.CardState) || got.Suspended || !got.BuriedUntil.IsZero() {
		t.Fatalf("scheduling changed: %+v", got)
	}

	// Change flag (spec), then toggle it off by choosing it again.
	c.SetFlag([]ID{fwd.ID}, 3)
	if a, _ := flags(); a != 3 {
		t.Fatalf("flag = %d, want 3", a)
	}
	c.SetFlag([]ID{fwd.ID}, 3)
	if a, _ := flags(); a != 0 {
		t.Fatalf("flag = %d, want toggled off", a)
	}

	// A mixed selection gets the flag on every card.
	c.SetFlag([]ID{fwd.ID}, 4)
	c.SetFlag([]ID{fwd.ID, rev.ID}, 4)
	if a, b := flags(); a != 4 || b != 4 {
		t.Fatalf("flags = %d, %d", a, b)
	}
	// Flag 0 always clears.
	c.SetFlag([]ID{fwd.ID, rev.ID}, 0)
	if a, b := flags(); a != 0 || b != 0 {
		t.Fatalf("flags = %d, %d", a, b)
	}

	mustUserErr(t, c.SetFlag([]ID{fwd.ID}, 8), "no flag 8")
	mustUserErr(t, c.SetFlag([]ID{fwd.ID}, -1), "no flag -1")

	c.SetFlag([]ID{rev.ID}, 7)
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Flag 1 card" {
		t.Fatalf("undo = %q", undo)
	}
	c2 := reopen(t, c, store, clk)
	if b, _ := c2.Card(rev.ID); b.Flag != 7 {
		t.Fatalf("reopened flag = %d", b.Flag)
	}
}

// positionsOf returns the positions of cards.
func positionsOf(t *testing.T, c *Collection, ids []ID) []int64 {
	t.Helper()
	out := make([]int64, len(ids))
	for i, card := range cardsOf(t, c, ids) {
		out[i] = card.NewPos
	}
	return out
}

// byPosition returns ids ordered as new cards are introduced.
func byPosition(t *testing.T, c *Collection, ids []ID) []ID {
	t.Helper()
	cards := cardsOf(t, c, ids)
	slices.SortFunc(cards, func(a, b Card) int { return positionOrder(&a, &b) })
	out := make([]ID, len(cards))
	for i, card := range cards {
		out[i] = card.ID
	}
	return out
}

func TestRepositionChapterToFront(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 30)
	chapter, rest := ids[10:], ids[:10]
	// Move a chapter to the front (collection spec): start 0, step 1,
	// shifting on.
	skipped, err := c.Reposition(chapter, 0, 1, false, true)
	if err != nil || skipped != 0 {
		t.Fatalf("skipped %d, err %v", skipped, err)
	}
	for i, p := range positionsOf(t, c, chapter) {
		if p != int64(i) {
			t.Fatalf("chapter positions = %v", positionsOf(t, c, chapter))
		}
	}
	if got := byPosition(t, c, ids); !slices.Equal(got, append(slices.Clone(chapter), rest...)) {
		t.Fatalf("order = %v", got)
	}
	// The queue follows: the 20 new cards of the day are the chapter.
	if got := newQueue(t, c, deck); !slices.Equal(got, chapter) {
		t.Fatalf("queue = %v, want %v", got, chapter)
	}
	// Shifting made room: no position is shared.
	seen := map[int64]bool{}
	for _, p := range positionsOf(t, c, ids) {
		if seen[p] {
			t.Fatalf("position %d shared: %v", p, positionsOf(t, c, ids))
		}
		seen[p] = true
	}
	// Cards added later go after all of them.
	later := addCards(t, c, deck, 1)[0]
	if got := byPosition(t, c, append(slices.Clone(ids), later)); got[len(got)-1] != later {
		t.Fatalf("added card not last: %v", got)
	}
	reopen(t, c, store, clk)
}

func TestRepositionMixedSelection(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 8)
	reviews := ids[5:]
	setState(t, c, reviews, reviewDue(clk, 1))
	before := cardsOf(t, c, reviews)
	// Mixed selection (collection spec): only the 5 new cards move, and
	// the 3 others are reported.
	skipped, err := c.Reposition(ids, 100, 5, false, false)
	if err != nil || skipped != 3 {
		t.Fatalf("skipped %d, err %v", skipped, err)
	}
	if got := positionsOf(t, c, ids[:5]); !slices.Equal(got, []int64{100, 105, 110, 115, 120}) {
		t.Fatalf("positions = %v", got)
	}
	if got := cardsOf(t, c, reviews); !reflect.DeepEqual(got, before) {
		t.Fatalf("review cards changed: %+v", got)
	}
	// Without shifting, the other new cards keep their positions.
	other := addCards(t, c, deck, 1)
	pos := positionsOf(t, c, other)
	if _, err := c.Reposition(ids[:1], pos[0], 1, false, false); err != nil {
		t.Fatal(err)
	}
	if got := positionsOf(t, c, other); !slices.Equal(got, pos) {
		t.Fatalf("other card moved: %v -> %v", pos, got)
	}
}

func TestRepositionRandomAndInvalid(t *testing.T) {
	c, _, _ := testCollection(t)
	ids := addCards(t, c, defaultDeck(c), 50)
	if _, err := c.Reposition(ids, 1, 1, true, false); err != nil {
		t.Fatal(err)
	}
	got := positionsOf(t, c, ids)
	if slices.IsSorted(got) {
		t.Fatalf("randomize kept the order: %v", got)
	}
	slices.Sort(got)
	for i, p := range got {
		if p != int64(i+1) {
			t.Fatalf("positions = %v", got)
		}
	}
	for _, args := range [][2]int64{{-1, 1}, {0, 0}, {MaxPosition, 1}} {
		if _, err := c.Reposition(ids, args[0], args[1], false, false); !IsUserError(err) {
			t.Errorf("Reposition(start %d, step %d) = %v, want a user error", args[0], args[1], err)
		}
	}
}

// Positions saved before Reposition existed were card IDs; opening the
// collection renumbers them in order.
func TestLegacyPositionsMigrated(t *testing.T) {
	c, store, clk := testCollection(t)
	ids := addCards(t, c, defaultDeck(c), 5)
	cards := cardsOf(t, c, ids)
	c.mu.Lock()
	o := &op{Kind: "test"}
	for i := range cards {
		// Legacy: the position was the card ID, here in reverse order.
		cards[i].NewPos = int64(cards[len(cards)-1-i].ID)
		o.Cards = append(o.Cards, &cards[i])
	}
	if err := c.commit(o); err != nil {
		t.Fatal(err)
	}
	c.mu.Unlock()
	c2, _, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	if got := positionsOf(t, c2, ids); !slices.Equal(got, []int64{5, 4, 3, 2, 1}) {
		t.Fatalf("positions = %v", got)
	}
	added := addCards(t, c2, defaultDeck(c2), 1)
	if got := positionsOf(t, c2, added); got[0] != 6 {
		t.Fatalf("next position = %d, want 6", got[0])
	}
	reopen(t, c2, store, clk)
}

func setInsertionOrder(t *testing.T, c *Collection, deck ID, order string) {
	t.Helper()
	o := deckOptions(c, deck)
	o.InsertionOrder = order
	if err := c.SetDeckOptions(deck, o); err != nil {
		t.Fatal(err)
	}
}

func TestRandomInsertionOrder(t *testing.T) {
	c, store, clk := testCollection(t)
	deck, _ := c.CreateDeck("Words")
	setInsertionOrder(t, c, deck, InsertRandom)
	// Random insertion (collection spec): 100 imported cards are not
	// introduced in import order.
	b := &ImportBatch{NoteTypes: []ImportNoteType{{Key: "b", ExistingID: noteType(t, c, "Basic").ID}}}
	for i := range 100 {
		b.Notes = append(b.Notes, ImportNote{TypeKey: "b", DeckName: "Words", Fields: []string{fmt.Sprint("word ", i), "x"}})
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	var ids []ID
	for _, card := range c.cards {
		if card.DeckID == deck {
			ids = append(ids, card.ID)
		}
	}
	sortIDs(ids) // import order
	if len(ids) != 100 {
		t.Fatalf("%d cards imported", len(ids))
	}
	if slices.Equal(byPosition(t, c, ids), ids) {
		t.Fatal("random insertion kept the import order")
	}
	// A card added to the deck lands among its new cards.
	lo, hi := slices.Min(positionsOf(t, c, ids)), slices.Max(positionsOf(t, c, ids))
	added := addCards(t, c, deck, 1)
	if p := positionsOf(t, c, added)[0]; p < lo || p > hi+1 {
		t.Fatalf("added at %d, deck spans %d-%d", p, lo, hi)
	}
	ids = append(ids, added...)
	// Back to sequential (collection spec): the order added.
	setInsertionOrder(t, c, deck, InsertSequential)
	if got := byPosition(t, c, ids); !slices.Equal(got, ids) {
		t.Fatalf("order = %v, want %v", got, ids)
	}
	reopen(t, c, store, clk)
}

func TestSwitchToRandomShufflesDeck(t *testing.T) {
	c, _, _ := testCollection(t)
	words, _ := c.CreateDeck("Words")
	ownPreset(t, c, words)
	other := addCards(t, c, defaultDeck(c), 5)
	ids := addCards(t, c, words, 50)
	before := positionsOf(t, c, ids)
	otherBefore := positionsOf(t, c, other)
	setInsertionOrder(t, c, words, InsertRandom)
	if slices.Equal(byPosition(t, c, ids), ids) {
		t.Fatal("switching to random kept the order")
	}
	// The deck keeps its positions; other decks are unchanged.
	got := positionsOf(t, c, ids)
	slices.Sort(got)
	if !slices.Equal(got, before) {
		t.Fatalf("positions = %v, want %v", got, before)
	}
	if got := positionsOf(t, c, other); !slices.Equal(got, otherBefore) {
		t.Fatalf("other deck = %v, want %v", got, otherBefore)
	}
}

func TestParseDueDate(t *testing.T) {
	for in, want := range map[string]DueDateSpec{
		"0": {0, 0, false}, " 5 ": {5, 5, false}, "60-90": {60, 90, false}, "60 - 90": {60, 90, false},
		"30!": {30, 30, true}, "1-3!": {1, 3, true}, "4-4": {4, 4, false},
	} {
		if got, err := ParseDueDate(in); err != nil || got != want {
			t.Errorf("ParseDueDate(%q) = %+v, %v; want %+v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "!", "abc", "-3", "5-", "-", "1.5", "+2", "1-2-3", "3!!", "10-5", "99999", "99999999999999999999"} {
		if _, err := ParseDueDate(in); err == nil || !IsUserError(err) {
			t.Errorf("ParseDueDate(%q) err = %v, want a user error", in, err)
		}
	}
}

// Postpone (scheduling spec): a review card keeps its interval.
func TestSetDueDatePostpone(t *testing.T) {
	c, store, clk := testCollection(t)
	id := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	mem := &scheduler.Memory{Stability: 12, Difficulty: 5}
	setState(t, c, []ID{id}, scheduler.CardState{State: scheduler.Review, Due: clk.t, ScheduledDays: 10,
		Memory: mem, Reps: 5, LastReview: clk.t.AddDate(0, 0, -10)})
	if err := c.SetDueDate([]ID{id}, "7"); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(id)
	if got.State != scheduler.Review || !got.Due.Equal(clk.t.AddDate(0, 0, 7)) || got.ScheduledDays != 10 ||
		*got.Memory != *mem || got.Reps != 5 {
		t.Fatalf("card = %+v", got)
	}
	if s := c.Settings(); s.LastDueDate != "7" {
		t.Fatalf("last due date = %q", s.LastDueDate)
	}
	reopen(t, c, store, clk)
}

// Interval too (scheduling spec).
func TestSetDueDateInterval(t *testing.T) {
	c, _, clk := testCollection(t)
	id := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	setState(t, c, []ID{id}, scheduler.CardState{State: scheduler.Review, Due: clk.t, ScheduledDays: 10,
		Memory: &scheduler.Memory{Stability: 12, Difficulty: 5}, LastReview: clk.t.AddDate(0, 0, -10)})
	if err := c.SetDueDate([]ID{id}, "30!"); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Card(id); !got.Due.Equal(clk.t.AddDate(0, 0, 30)) || got.ScheduledDays != 30 {
		t.Fatalf("card = %+v", got)
	}
	// 0! makes it due today with the shortest interval, a day.
	if err := c.SetDueDate([]ID{id}, "0!"); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Card(id); !got.Due.Equal(clk.t) || got.ScheduledDays != 1 {
		t.Fatalf("card = %+v", got)
	}
}

// Range (scheduling spec): 100 new cards spread over the range.
func TestSetDueDateRange(t *testing.T) {
	c, store, clk := testCollection(t)
	var ids []ID
	for i := 0; i < 100; i++ {
		ids = append(ids, c.CardsOfNote(addNote(t, c, "Basic", fmt.Sprint("q", i), "a"))[0].ID)
	}
	if err := c.SetDueDate(ids, "1-10"); err != nil {
		t.Fatal(err)
	}
	perDay := map[int]int{}
	var first []time.Time
	for _, id := range ids {
		card, _ := c.Card(id)
		days := int(card.Due.Sub(clk.t).Hours() / 24)
		if card.State != scheduler.Review || days < 1 || days > 10 || !card.Due.Equal(clk.t.AddDate(0, 0, days)) ||
			card.ScheduledDays != days || card.Memory == nil {
			t.Fatalf("card = %+v", card)
		}
		perDay[days]++
		first = append(first, card.Due)
	}
	if len(perDay) != 10 {
		t.Fatalf("cards per day = %v, want every day of the range", perDay)
	}
	// The same day gives the same random days.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if err := c.SetDueDate(ids, "1-10"); err != nil {
		t.Fatal(err)
	}
	for i, id := range ids {
		if card, _ := c.Card(id); !card.Due.Equal(first[i]) {
			t.Fatalf("card %d due %v, then %v", i, first[i], card.Due)
		}
	}
	reopen(t, c, store, clk)
}

// Invalid input (scheduling spec): nothing changes.
func TestSetDueDateInvalid(t *testing.T) {
	c, _, _ := testCollection(t)
	id := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	before, _ := c.Card(id)
	if err := c.SetDueDate([]ID{id}, "10-5"); err == nil || !IsUserError(err) {
		t.Fatalf("err = %v", err)
	}
	if got, _ := c.Card(id); !reflect.DeepEqual(got, before) {
		t.Fatalf("card changed: %+v", got)
	}
	if h, _ := c.CardHistory(id); len(h) != 0 || c.Settings().LastDueDate != "" {
		t.Fatalf("history = %+v, last = %q", h, c.Settings().LastDueDate)
	}
}

// New card scheduled (scheduling spec), and undo.
func TestSetDueDateNewCard(t *testing.T) {
	c, store, clk := testCollection(t)
	id := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	before, _ := c.Card(id)
	if err := c.SetDueDate([]ID{id}, "2"); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(id)
	want := scheduler.FirstGoodMemory(clk.t, nil)
	if got.State != scheduler.Review || !got.Due.Equal(clk.t.AddDate(0, 0, 2)) || got.ScheduledDays != 2 ||
		got.Memory == nil || *got.Memory != want || got.Memory.Stability <= 0 || got.Memory.Difficulty <= 0 {
		t.Fatalf("card = %+v", got)
	}
	h, err := c.CardHistory(id)
	if err != nil || len(h) != 1 || h[0].Rating != 0 || h[0].StateBefore != scheduler.New || h[0].IvlAfter != 2 {
		t.Fatalf("history = %+v, %v", h, err)
	}
	// A copy, as c goes on with the store.
	reopen(t, c, store.Clone(), clk)
	if name, err := c.Undo(); err != nil || usermsg.Plain(name) != "Set due date of 1 card" {
		t.Fatalf("undo = %q, %v", name, err)
	}
	if got, _ := c.Card(id); got.State != scheduler.New || got.Memory != nil || got.NewPos != before.NewPos {
		t.Fatalf("undone card = %+v", got)
	}
	if h, _ := c.CardHistory(id); len(h) != 0 {
		t.Fatalf("history after undo = %+v", h)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if h, _ := c.CardHistory(id); len(h) != 1 {
		t.Fatalf("history after redo = %+v", h)
	}
	reopen(t, c, store, clk)
}

// Learning cards become Review cards, keeping their memory state; one
// without a memory state (e.g. imported) gets a first Good's.
func TestSetDueDateLearningCards(t *testing.T) {
	c, _, clk := testCollection(t)
	a := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	b := c.CardsOfNote(addNote(t, c, "Basic", "b", "2"))[0].ID
	mem := &scheduler.Memory{Stability: 0.5, Difficulty: 6}
	setState(t, c, []ID{a}, scheduler.CardState{State: scheduler.Learning, Due: clk.t.Add(10 * time.Minute), Step: 1,
		Memory: mem, Reps: 1, LastReview: clk.t.Add(-time.Minute)})
	setState(t, c, []ID{b}, scheduler.CardState{State: scheduler.Relearning, Due: clk.t.AddDate(0, 0, 1), Interday: true,
		Reps: 9, Lapses: 2})
	if err := c.SetDueDate([]ID{a, b}, "0"); err != nil {
		t.Fatal(err)
	}
	ca, _ := c.Card(a)
	if ca.State != scheduler.Review || ca.Step != 0 || !ca.Due.Equal(clk.t) || ca.ScheduledDays != 1 ||
		*ca.Memory != *mem || !ca.LastReview.Equal(clk.t.Add(-time.Minute)) {
		t.Fatalf("learning card = %+v", ca)
	}
	cb, _ := c.Card(b)
	if cb.State != scheduler.Review || cb.Interday || cb.Memory == nil || cb.Lapses != 2 || cb.ScheduledDays != 1 {
		t.Fatalf("relearning card = %+v", cb)
	}
	for _, id := range []ID{a, b} {
		if h, _ := c.CardHistory(id); len(h) != 1 || h[0].Rating != 0 {
			t.Fatalf("history of %d = %+v", id, h)
		}
	}
}
