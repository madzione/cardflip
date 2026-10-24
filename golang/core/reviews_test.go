package core

import (
	"reflect"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

func TestReviewLogEntry(t *testing.T) {
	c, store, clk := testCollection(t)
	card := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0]
	// Log entry written (spec): rating Hard after 7 seconds.
	entry, err := c.Answer(card.ID, scheduler.Hard, 7*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if entry.CardID != card.ID || entry.Rating != scheduler.Hard || entry.DurationMs != 7000 ||
		entry.StateBefore != scheduler.New || !entry.Time.Equal(clk.t) {
		t.Fatalf("entry = %+v", entry)
	}
	got := c.ReviewsOfCard(card.ID)
	if len(got) != 1 || got[0].ID != entry.ID {
		t.Fatalf("reviews = %+v", got)
	}
	answered, _ := c.Card(card.ID)
	if answered.State != scheduler.Learning || answered.Memory == nil || answered.Reps != 1 {
		t.Fatalf("card after answer = %+v", answered)
	}
	c2 := reopen(t, c, store, clk)
	if got := c2.ReviewsOfCard(card.ID); len(got) != 1 || got[0].DurationMs != 7000 || got[0].Rating != scheduler.Hard {
		t.Fatalf("reviews after reopen = %+v", got)
	}
	_, err = c.Answer(card.ID, 9, time.Second)
	mustUserErr(t, err, "Invalid rating")
}

func TestAnswerTimeCapped(t *testing.T) {
	c, _, _ := testCollection(t)
	ids := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))
	// Long pause (spec): the deck's maximum is 60 seconds and the card is
	// answered after 5 minutes.
	entry, err := c.Answer(ids[0].ID, scheduler.Good, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if entry.DurationMs != 60000 {
		t.Fatalf("duration = %d ms", entry.DurationMs)
	}
	if got := c.ReviewsOfCard(ids[0].ID); got[0].DurationMs != 60000 {
		t.Fatalf("logged duration = %d ms", got[0].DurationMs)
	}
	// The cap is the card's deck option.
	opts := DefaultDeckOptions()
	opts.MaxAnswerSeconds = 10
	if err := c.SetDeckOptions(ids[0].DeckID, opts); err != nil {
		t.Fatal(err)
	}
	for in, want := range map[time.Duration]int64{9 * time.Second: 9000, 11 * time.Second: 10000, -time.Second: 0} {
		if entry, err := c.Answer(ids[0].ID, scheduler.Good, in); err != nil || entry.DurationMs != want {
			t.Errorf("answer after %v: %d ms, %v", in, entry.DurationMs, err)
		}
	}
}

func TestUndoRestoresExactState(t *testing.T) {
	c, store, clk := testCollection(t)
	card := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0]
	c.Answer(card.ID, scheduler.Good, time.Second)
	clk.advance(10 * time.Minute)
	c.Answer(card.ID, scheduler.Good, time.Second)
	clk.advance(3 * 24 * time.Hour)
	before, _ := c.Card(card.ID)
	if before.State != scheduler.Review {
		t.Fatalf("setup: %+v", before)
	}

	// Undo last answer (spec): rated Again by mistake, then undone.
	entry, err := c.Answer(card.ID, scheduler.Again, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if lapsed, _ := c.Card(card.ID); lapsed.State != scheduler.Relearning || lapsed.Lapses != 1 {
		t.Fatalf("after Again: %+v", lapsed)
	}
	// Only the most recent answer of a card can be undone.
	older := c.ReviewsOfCard(card.ID)[0].ID
	_, err = c.UndoReview(older)
	mustUserErr(t, err, "most recent")

	cardID, err := c.UndoReview(entry.ID)
	if err != nil || cardID != card.ID {
		t.Fatalf("undo: %v %v", cardID, err)
	}
	after, _ := c.Card(card.ID)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("not restored exactly:\n got %+v\nwant %+v", after, before)
	}
	if n := len(c.ReviewsOfCard(card.ID)); n != 2 {
		t.Fatalf("reviews after undo = %d", n)
	}
	_, err = c.UndoReview(entry.ID)
	mustUserErr(t, err, "nothing to undo")

	c2 := reopen(t, c, store, clk)
	if n := len(c2.ReviewsOfCard(card.ID)); n != 2 {
		t.Fatalf("reviews after reopen = %d", n)
	}
	if got, _ := c2.Card(card.ID); got.State != scheduler.Review || got.Lapses != 0 {
		t.Fatalf("card after reopen = %+v", got)
	}
}

// Late-night review counts for the previous day (spec): with rollover 04:00
// a review at 02:30 belongs to the day that started at 04:00 the day before.
func TestReviewAt0230CountsForPreviousDay(t *testing.T) {
	c, _, clk := testCollection(t)
	a := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0]
	b := c.CardsOfNote(addNote(t, c, "Basic", "b", "2"))[0]
	clk.t = time.Date(2026, 10, 5, 22, 0, 0, 0, time.Local)
	c.Answer(a.ID, scheduler.Good, time.Second)
	clk.t = time.Date(2026, 10, 6, 2, 30, 0, 0, time.Local)
	c.Answer(b.ID, scheduler.Good, time.Second)

	c.mu.RLock()
	n := len(c.todayEntries(clk.t))
	c.mu.RUnlock()
	if n != 2 {
		t.Fatalf("entries in the study day at 02:30 = %d, want 2", n)
	}
	clk.t = time.Date(2026, 10, 6, 4, 0, 0, 0, time.Local)
	c.mu.RLock()
	n = len(c.todayEntries(clk.t))
	c.mu.RUnlock()
	if n != 0 {
		t.Fatalf("entries after rollover = %d", n)
	}
}

func TestPreviewsMatchAnswer(t *testing.T) {
	c, _, _ := testCollection(t)
	card := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0]
	p, err := c.Previews(card.ID)
	if err != nil {
		t.Fatal(err)
	}
	// The default steps 1m and 10m, with step fuzz.
	if p[0].Interval < time.Minute || p[0].Interval > 75*time.Second ||
		p[2].Interval < 10*time.Minute || p[2].Interval > 12*time.Minute+30*time.Second ||
		p[2].Label != FormatInterval(p[2].Interval) {
		t.Fatalf("new card previews = %+v", p)
	}
	before, _ := c.Card(card.ID)
	if before.Reps != 0 {
		t.Fatal("preview changed the card")
	}
	c.Answer(card.ID, scheduler.Good, 0)
	after, _ := c.Card(card.ID)
	if got := after.Due.Sub(c.Now()); got != p[2].Interval {
		t.Fatalf("Good due in %v, preview %v", got, p[2].Interval)
	}
}

func TestFormatInterval(t *testing.T) {
	cases := map[time.Duration]string{
		30 * time.Second:     "<1m",
		time.Minute:          "<1m",
		10 * time.Minute:     "<10m",
		90 * time.Minute:     "1.5h",
		3 * 24 * time.Hour:   "3d",
		36 * 24 * time.Hour:  "1.2mo",
		60 * 24 * time.Hour:  "2mo",
		400 * 24 * time.Hour: "1.1y",
		730 * 24 * time.Hour: "2y",
	}
	for d, want := range cases {
		if got := FormatInterval(d); got != want {
			t.Errorf("FormatInterval(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestIsLeech(t *testing.T) {
	for _, tc := range []struct {
		threshold int
		leech     []int
	}{
		{8, []int{8, 12, 16, 20}},
		{5, []int{5, 8, 11}},
		{1, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
	} {
		want := map[int]bool{}
		for _, n := range tc.leech {
			want[n] = true
		}
		for lapses := 0; lapses <= tc.leech[len(tc.leech)-1]; lapses++ {
			if got := isLeech(lapses, tc.threshold); got != want[lapses] {
				t.Errorf("isLeech(%d, %d) = %v", lapses, tc.threshold, got)
			}
		}
	}
}

// leechSetup returns a Review card with this many lapses, in the default
// deck with the given leech action and threshold 8.
func leechSetup(t *testing.T, c *Collection, clk *clock, action string, lapses int) (Card, Note) {
	t.Helper()
	opts := DefaultDeckOptions()
	opts.LeechAction = action
	if err := c.SetDeckOptions(defaultDeck(c), opts); err != nil {
		t.Fatal(err)
	}
	noteID := addNote(t, c, "Basic", "q", "a")
	card := c.CardsOfNote(noteID)[0]
	c.Answer(card.ID, scheduler.Easy, time.Second)
	clk.advance(30 * 24 * time.Hour)
	card, _ = c.Card(card.ID)
	if card.State != scheduler.Review {
		t.Fatalf("setup: %+v", card)
	}
	card.Lapses = lapses
	c.mu.Lock()
	err := c.commit(&op{Kind: "test", Cards: []*Card{&card}})
	c.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	note, _ := c.Note(noteID)
	return card, note
}

func TestLeechThresholdReached(t *testing.T) {
	c, store, clk := testCollection(t)
	// Threshold reached (spec): threshold 8, suspend, 7 lapses, Again.
	card, _ := leechSetup(t, c, clk, LeechSuspend, 7)
	entry, err := c.Answer(card.ID, scheduler.Again, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !entry.Leech || !entry.LeechTagged || !entry.LeechSuspended {
		t.Fatalf("entry = %+v", entry)
	}
	check := func(c *Collection) {
		t.Helper()
		got, _ := c.Card(card.ID)
		note, _ := c.Note(card.NoteID)
		if got.Lapses != 8 || !got.Suspended || got.State != scheduler.Relearning {
			t.Fatalf("card = %+v", got)
		}
		if !reflect.DeepEqual(note.Tags, []string{"leech"}) {
			t.Fatalf("tags = %v", note.Tags)
		}
	}
	check(c)
	c2 := reopen(t, c, store, clk)
	check(c2)
	if got := c2.ReviewsOfCard(card.ID); !got[len(got)-1].LeechSuspended {
		t.Fatalf("entry after reopen = %+v", got[len(got)-1])
	}
}

func TestLeechRepeatWarning(t *testing.T) {
	c, _, clk := testCollection(t)
	// Not a leech between the threshold and threshold + half.
	card, _ := leechSetup(t, c, clk, LeechTag, 9)
	if e, _ := c.Answer(card.ID, scheduler.Again, time.Second); e.Leech {
		t.Fatalf("10 lapses: %+v", e)
	}
	// Repeat warning (spec): tag only, reaching 12 lapses. The note already
	// has the tag, so the answer does not change it.
	card, _ = leechSetup(t, c, clk, LeechTag, 11)
	if err := c.AddTags([]ID{card.NoteID}, []string{"Leech", "hard"}); err != nil {
		t.Fatal(err)
	}
	e, err := c.Answer(card.ID, scheduler.Again, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Leech || e.LeechTagged || e.LeechSuspended {
		t.Fatalf("entry = %+v", e)
	}
	got, _ := c.Card(card.ID)
	note, _ := c.Note(card.NoteID)
	if got.Lapses != 12 || got.Suspended || !reflect.DeepEqual(note.Tags, []string{"Leech", "hard"}) {
		t.Fatalf("card = %+v, tags = %v", got, note.Tags)
	}
	// Again on a card that is not in Review does not lapse it, so it is
	// not checked, whatever its lapse count.
	newCard := c.CardsOfNote(addNote(t, c, "Basic", "n", "a"))[0]
	newCard.Lapses = 8
	c.mu.Lock()
	c.commit(&op{Kind: "test", Cards: []*Card{&newCard}})
	c.mu.Unlock()
	if e, _ := c.Answer(newCard.ID, scheduler.Again, time.Second); e.Leech {
		t.Fatalf("new card: %+v", e)
	}
}

func TestUndoLeech(t *testing.T) {
	for _, viaHistory := range []bool{true, false} {
		c, store, clk := testCollection(t)
		card, _ := leechSetup(t, c, clk, LeechSuspend, 7)
		other := addNote(t, c, "Basic", "x", "y")
		e, err := c.Answer(card.ID, scheduler.Again, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if !viaHistory {
			// Another change after the answer: undoing the answer is a new
			// step instead of the top of the history.
			if err := c.AddTags([]ID{other}, []string{"later"}); err != nil {
				t.Fatal(err)
			}
		}
		// Undo leech (spec): unsuspended, no leech tag, shown again.
		if id, err := c.UndoReview(e.ID); err != nil || id != card.ID {
			t.Fatalf("undo (history %v): %v %v", viaHistory, id, err)
		}
		check := func(c *Collection) {
			t.Helper()
			got, _ := c.Card(card.ID)
			n, _ := c.Note(card.NoteID)
			if got.Suspended || got.Lapses != 7 || got.State != scheduler.Review || len(n.Tags) != 0 {
				t.Fatalf("after undo (history %v): card %+v, tags %v", viaHistory, got, n.Tags)
			}
			if !reflect.DeepEqual(got.CardState, card.CardState) {
				t.Fatalf("state not restored (history %v)", viaHistory)
			}
		}
		check(c)
		check(reopen(t, c, store, clk))
		if n, _ := c.Note(other); !viaHistory && !reflect.DeepEqual(n.Tags, []string{"later"}) {
			t.Fatalf("other note's tags = %v", n.Tags)
		}
	}

	// Undo when the note already had the tag: the tag stays.
	c, _, clk := testCollection(t)
	card, _ := leechSetup(t, c, clk, LeechSuspend, 7)
	c.AddTags([]ID{card.NoteID}, []string{"leech"})
	e, _ := c.Answer(card.ID, scheduler.Again, time.Second)
	c.AddTags([]ID{card.NoteID}, []string{"later"})
	if _, err := c.UndoReview(e.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(card.ID)
	n, _ := c.Note(card.NoteID)
	if got.Suspended || !reflect.DeepEqual(n.Tags, []string{"leech", "later"}) {
		t.Fatalf("card %+v, tags %v", got, n.Tags)
	}
}

// Position kept (collection spec): the new-card position at the first
// answer is remembered, once, across reviews, moves and backups.
func TestOriginalPosition(t *testing.T) {
	c, store, clk := testCollection(t)
	for range 2 {
		addNote(t, c, "Basic", "earlier", "x")
	}
	card := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0]
	if _, err := c.Reposition([]ID{card.ID}, 42, 1, false, false); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Card(card.ID); got.OriginalPos != nil {
		t.Fatalf("original position before the first answer: %d", *got.OriginalPos)
	}
	for i := range 6 {
		rating := scheduler.Good
		if i == 3 {
			rating = scheduler.Again
		}
		if _, err := c.Answer(card.ID, rating, time.Second); err != nil {
			t.Fatal(err)
		}
		clk.advance(30 * 24 * time.Hour)
	}
	other, _ := c.CreateDeck("Other")
	if err := c.MoveCards([]ID{card.ID}, other); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Card(card.ID)
	if got.OriginalPos == nil || *got.OriginalPos != 42 || got.Reps != 6 {
		t.Fatalf("card after reviews = %+v", got)
	}
	c2 := reopen(t, c, store, clk)
	if got, _ := c2.Card(card.ID); got.OriginalPos == nil || *got.OriginalPos != 42 {
		t.Fatalf("card after reopening = %+v", got)
	}
	snap, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	c.ResetCards([]ID{card.ID}, ResetOptions{})
	if err := c.Restore(snap, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Card(card.ID); got.OriginalPos == nil || *got.OriginalPos != 42 || got.State == scheduler.New {
		t.Fatalf("card after restoring the backup = %+v", got)
	}
}

// A reset's manual entry is no answer: it cannot be undone as one, and it
// does not use up today's limits or bury siblings.
func TestResetEntryIsNoAnswer(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	if err := c.SetDeckOptions(deck, deckOpts(20, 1, 0.9)); err != nil {
		t.Fatal(err)
	}
	var reviews []Card
	for _, f := range []string{"r1", "r2"} {
		card := c.CardsOfNote(addNote(t, c, "Basic", f, "x"))[0]
		card.CardState = scheduler.CardState{State: scheduler.Review, Due: clk.t, ScheduledDays: 3, Reps: 2,
			Memory: &scheduler.Memory{Stability: 3, Difficulty: 5}, LastReview: clk.t.AddDate(0, 0, -3)}
		c.mu.Lock()
		c.commit(&op{Kind: "test", Cards: []*Card{&card}})
		c.mu.Unlock()
		reviews = append(reviews, card)
	}
	if err := c.ResetCards([]ID{reviews[0].ID}, ResetOptions{RestorePosition: true}); err != nil {
		t.Fatal(err)
	}
	if got := c.DeckCounts()[deck]; got.New != 1 || got.Review != 1 {
		t.Fatalf("counts after reset = %+v", got)
	}
	entry := c.ReviewsOfCard(reviews[0].ID)[0]
	_, err := c.UndoReview(entry.ID)
	mustUserErr(t, err, "Only answers given in Cardflip")
}
