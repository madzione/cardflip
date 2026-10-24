package core

import (
	"cardflip/golang/internal/usermsg"
	"slices"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// setState overwrites cards' scheduling state directly (test helper).
func setState(t *testing.T, c *Collection, ids []ID, cs scheduler.CardState) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	o := &op{Kind: "test"}
	for _, card := range c.cardCopies(ids) {
		card.CardState = cs
		o.Cards = append(o.Cards, card)
	}
	if err := c.commit(o); err != nil {
		t.Fatal(err)
	}
}

func addCards(t *testing.T, c *Collection, deck ID, n int) []ID {
	t.Helper()
	var ids []ID
	for i := 0; i < n; i++ {
		note := addNoteIn(t, c, deck, "Basic", "q", "a")
		ids = append(ids, c.CardsOfNote(note)[0].ID)
	}
	return ids
}

func reviewDue(clk *clock, daysAgo int) scheduler.CardState {
	return scheduler.CardState{State: scheduler.Review, Due: clk.t.Add(-time.Hour), ScheduledDays: daysAgo,
		Memory: &scheduler.Memory{Stability: float64(daysAgo), Difficulty: 5}, Reps: 3, LastReview: clk.t.AddDate(0, 0, -daysAgo)}
}

func counts(t *testing.T, c *Collection, deck ID) Counts {
	t.Helper()
	q, err := c.Queue(deck)
	if err != nil {
		t.Fatal(err)
	}
	if dc := c.DeckCounts()[deck]; dc != q.Counts() {
		t.Fatalf("DeckCounts %+v != Queue counts %+v", dc, q.Counts())
	}
	return q.Counts()
}

func TestNewLimitReached(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 50)
	if got := counts(t, c, deck); got.New != 20 {
		t.Fatalf("new = %d, want 20", got.New)
	}
	// New limit reached (spec): 20 introduced today → 0 new until rollover.
	for _, id := range ids[:20] {
		if _, err := c.Answer(id, scheduler.Easy, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	if got := counts(t, c, deck); got.New != 0 {
		t.Fatalf("new after 20 introduced = %d", got.New)
	}
	clk.t = scheduler.NextDayStart(clk.t, 4)
	if got := counts(t, c, deck); got.New != 20 {
		t.Fatalf("new after rollover = %d", got.New)
	}
	// Undoing an answer gives the slot back.
	clk.t = clk.t.Add(time.Hour)
	e, _ := c.Answer(ids[20], scheduler.Good, time.Second)
	if got := counts(t, c, deck); got.New != 19 {
		t.Fatalf("new = %d", got.New)
	}
	c.UndoReview(e.ID)
	if got := counts(t, c, deck); got.New != 20 {
		t.Fatalf("new after undo = %d", got.New)
	}
}

func TestLearningIgnoresLimits(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	c.SetDeckOptions(deck, deckOpts(0, 3, 0.9))
	reviews := addCards(t, c, deck, 5)
	setState(t, c, reviews, reviewDue(clk, 10))
	if got := counts(t, c, deck); got.Review != 3 {
		t.Fatalf("reviews = %d, want 3", got.Review)
	}
	for _, id := range reviews[:3] {
		c.Answer(id, scheduler.Good, time.Second)
	}
	if got := counts(t, c, deck); got.Review != 0 {
		t.Fatalf("reviews after limit used = %d", got.Review)
	}
	// Learning ignores limits (spec): a relearning card that becomes due is
	// still shown.
	relearn := addCards(t, c, deck, 1)
	setState(t, c, relearn, scheduler.CardState{State: scheduler.Relearning, Due: clk.t.Add(-time.Minute),
		Memory: &scheduler.Memory{Stability: 1, Difficulty: 6}, Reps: 4, Lapses: 1, LastReview: clk.t.Add(-11 * time.Minute)})
	q, _ := c.Queue(deck)
	if len(q.Learning) != 1 || q.Learning[0] != relearn[0] {
		t.Fatalf("learning queue = %v", q.Learning)
	}
	if id, ok := PickNext(q, 0, 3); !ok || id != relearn[0] {
		t.Fatalf("next = %v %v", id, ok)
	}
}

// burySiblings turns new and review sibling burying on in deck's preset;
// they are off by default.
func burySiblings(t *testing.T, c *Collection, deck ID) {
	t.Helper()
	o := deckOptions(c, deck)
	o.BuryNew, o.BuryReview = true, true
	if err := c.SetDeckOptions(deck, o); err != nil {
		t.Fatal(err)
	}
}

// Siblings by default (spec): with the default options, a note's other
// cards are shown on the day one of them is studied.
func TestSiblingsByDefault(t *testing.T) {
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	cards := c.CardsOfNote(addNoteIn(t, c, deck, "Cloze", "{{c1::uno}} {{c2::dos}} {{c3::tres}}", ""))
	if len(cards) != 3 {
		t.Fatalf("cards = %+v", cards)
	}
	if _, err := c.Answer(cards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	q, _ := c.Queue(deck)
	if !slices.Equal(q.New, []ID{cards[1].ID, cards[2].ID}) {
		t.Fatalf("new = %v", q.New)
	}
	if o := overview(t, c, deck); o.Buried != 0 {
		t.Fatalf("overview = %+v", o)
	}
}

func TestSiblingBuried(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	burySiblings(t, c, deck)
	note := addNoteIn(t, c, deck, "Basic (and reversed card)", "hola", "hello")
	cards := c.CardsOfNote(note)
	q, _ := c.Queue(deck)
	if len(q.New) != 2 {
		t.Fatalf("new = %v", q.New)
	}
	// Sibling buried (spec): once the forward card is studied today the
	// reverse card waits for the next study day.
	if _, err := c.Answer(cards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	q, _ = c.Queue(deck)
	if len(q.New) != 0 {
		t.Fatalf("reverse card still queued: %v", q.New)
	}
	clk.t = scheduler.NextDayStart(clk.t, 4).Add(time.Minute)
	q, _ = c.Queue(deck)
	if len(q.New) != 1 || q.New[0] != cards[1].ID {
		t.Fatalf("next day new = %v", q.New)
	}
}

func TestBurySiblingOptions(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	newCards := c.CardsOfNote(addNoteIn(t, c, deck, "Basic (and reversed card)", "hola", "hello"))
	reviews := c.CardsOfNote(addNoteIn(t, c, deck, "Basic (and reversed card)", "uno", "one"))
	setState(t, c, []ID{reviews[0].ID, reviews[1].ID}, reviewDue(clk, 3))
	for _, id := range []ID{newCards[0].ID, reviews[0].ID} {
		if _, err := c.Answer(id, scheduler.Good, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	// Defaults: neither sibling is buried.
	if got := counts(t, c, deck); got.New != 1 || got.Review != 1 {
		t.Fatalf("default counts = %+v", got)
	}
	burySiblings(t, c, deck)
	if got := counts(t, c, deck); got.New != 0 || got.Review != 0 {
		t.Fatalf("counts with burying = %+v", got)
	}
	// Review siblings allowed (spec): both cards are shown today.
	o := deckOptions(c, deck)
	o.BuryReview = false
	c.SetDeckOptions(deck, o)
	if got := counts(t, c, deck); got.New != 0 || got.Review != 1 {
		t.Fatalf("counts without review burying = %+v", got)
	}
	o.BuryNew = false
	c.SetDeckOptions(deck, o)
	if got := counts(t, c, deck); got.New != 1 || got.Review != 1 {
		t.Fatalf("counts without burying = %+v", got)
	}
	// The options of the sibling's own deck decide.
	other, _ := c.CreateDeck("Other")
	ownPreset(t, c, other)
	burySiblings(t, c, other)
	if err := c.MoveCards([]ID{newCards[1].ID}, other); err != nil {
		t.Fatal(err)
	}
	if got := counts(t, c, other); got.New != 0 {
		t.Fatalf("sibling in burying deck = %+v", got)
	}
}

func TestParentAndChildLimits(t *testing.T) {
	c, _, _ := testCollection(t)
	parent, _ := c.CreateDeck("P")
	a, _ := c.CreateDeck("P::A")
	b, _ := c.CreateDeck("P::B")
	ownPreset(t, c, parent)
	ownPreset(t, c, a)
	c.SetDeckOptions(parent, deckOpts(12, 200, 0.9))
	c.SetDeckOptions(a, deckOpts(5, 200, 0.9))
	addCards(t, c, parent, 2)
	addCards(t, c, a, 10)
	addCards(t, c, b, 10)
	// Child limits apply on their own...
	if got := counts(t, c, a); got.New != 5 {
		t.Fatalf("A new = %d", got.New)
	}
	if got := counts(t, c, b); got.New != 10 {
		t.Fatalf("B new = %d", got.New)
	}
	// ...and the parent includes its children, capped by its own limit.
	if got := counts(t, c, parent); got.New != 12 {
		t.Fatalf("P new = %d", got.New)
	}
	q, _ := c.Queue(parent)
	inA := 0
	for _, id := range q.New {
		if card, _ := c.Card(id); card.DeckID == a {
			inA++
		}
	}
	if inA > 5 {
		t.Fatalf("parent queue takes %d cards from A, over A's limit", inA)
	}
}

func TestSuspendedAndBuriedNotDue(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 3)
	setState(t, c, ids, reviewDue(clk, 5))
	if got := counts(t, c, deck); got.Review != 3 {
		t.Fatalf("reviews = %d", got.Review)
	}
	// Suspended card skipped (spec).
	c.SetSuspended(ids[:1], true)
	c.BuryCards(ids[1:2])
	q, _ := c.Queue(deck)
	if len(q.Review) != 1 || q.Review[0] != ids[2] {
		t.Fatalf("queue = %+v", q)
	}
	c.SetSuspended(ids[:1], false)
	if got := counts(t, c, deck); got.Review != 2 {
		t.Fatalf("reviews after unsuspend = %d", got.Review)
	}
}

func TestCountsDecreaseAfterAnswer(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 5)
	setState(t, c, ids, reviewDue(clk, 4))
	// Counts decrease (study-session spec): 5 reviews, answer Good → 4.
	if got := counts(t, c, deck); got.Review != 5 {
		t.Fatalf("reviews = %d", got.Review)
	}
	c.Answer(ids[0], scheduler.Good, time.Second)
	if got := counts(t, c, deck); got.Review != 4 {
		t.Fatalf("reviews after Good = %d", got.Review)
	}
}

func TestReviewsDueByEndOfDay(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 3)
	end := scheduler.NextDayStart(clk.t, 4)
	setState(t, c, ids[:1], scheduler.CardState{State: scheduler.Review, Due: end.Add(-time.Minute), ScheduledDays: 3, Memory: &scheduler.Memory{Stability: 3, Difficulty: 5}})
	setState(t, c, ids[1:2], scheduler.CardState{State: scheduler.Review, Due: end.Add(time.Minute), ScheduledDays: 3, Memory: &scheduler.Memory{Stability: 3, Difficulty: 5}})
	setState(t, c, ids[2:], scheduler.CardState{State: scheduler.Learning, Due: clk.t.Add(5 * time.Minute), Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}})
	q, _ := c.Queue(deck)
	if len(q.Review) != 1 || q.Review[0] != ids[0] || len(q.Learning) != 0 {
		t.Fatalf("queue = %+v", q)
	}
	// Within the default learn ahead limit of 20 minutes.
	if !slices.Equal(q.LearningAhead, ids[2:]) || !q.LearningLater.IsZero() {
		t.Fatalf("learning ahead = %v, later = %v", q.LearningAhead, q.LearningLater)
	}
}

func setLearnAhead(t *testing.T, c *Collection, minutes int) {
	t.Helper()
	st := c.Settings()
	st.LearnAheadMinutes = &minutes
	if err := c.UpdateSettings(st); err != nil {
		t.Fatal(err)
	}
}

func TestLearnAhead(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 4)
	learning := func(in time.Duration) scheduler.CardState {
		return scheduler.CardState{State: scheduler.Learning, Due: clk.t.Add(in), Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}}
	}
	setState(t, c, ids[:1], reviewDue(clk, 1))
	setState(t, c, ids[1:2], learning(15*time.Minute))
	setState(t, c, ids[2:3], learning(5*time.Minute))
	setState(t, c, ids[3:], learning(time.Hour))

	q, _ := c.Queue(deck)
	if !slices.Equal(q.LearningAhead, []ID{ids[2], ids[1]}) || len(q.Learning) != 0 {
		t.Fatalf("learning = %v, ahead = %v", q.Learning, q.LearningAhead)
	}
	// The next learning card beyond the limit.
	if !q.LearningLater.Equal(clk.t.Add(time.Hour)) {
		t.Fatalf("learning later = %v", q.LearningLater)
	}
	// Cards ahead are not counted, and wait for the other queues.
	if got := q.Counts(); got != (Counts{Review: 1}) {
		t.Fatalf("counts = %+v", got)
	}
	if id, ok := PickNext(q, 0, 0); !ok || id != ids[0] {
		t.Fatalf("picked %d", id)
	}

	// Finish without waiting (study-session spec): the last review card is
	// answered and a learning card is due in 5 minutes.
	if _, err := c.Answer(ids[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	q, _ = c.Queue(deck)
	if id, ok := PickNext(q, 0, 1); !ok || id != ids[2] {
		t.Fatalf("picked %d (%v)", id, ok)
	}

	// Limit off (study-session spec): the card is only shown once due.
	setLearnAhead(t, c, 0)
	q, _ = c.Queue(deck)
	if id, ok := PickNext(q, 0, 1); ok || len(q.LearningAhead) != 0 {
		t.Fatalf("picked %d, ahead = %v", id, q.LearningAhead)
	}
	if !q.LearningLater.Equal(clk.t.Add(5 * time.Minute)) {
		t.Fatalf("learning later = %v", q.LearningLater)
	}
	clk.t = clk.t.Add(5 * time.Minute)
	q, _ = c.Queue(deck)
	if id, ok := PickNext(q, 0, 1); !ok || id != ids[2] {
		t.Fatalf("picked %d once due", id)
	}
}

func TestNextDue(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	if !c.NextDue(deck).IsZero() {
		t.Fatal("empty deck has a next due time")
	}
	ids := addCards(t, c, deck, 1)
	// Nothing due (study-session spec): the next card is due tomorrow.
	setState(t, c, ids, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 1), ScheduledDays: 1, Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}})
	if got := counts(t, c, deck); got.Total() != 0 {
		t.Fatalf("counts = %+v", got)
	}
	tomorrow := scheduler.NextDayStart(clk.t, 4)
	if got := c.NextDue(deck); !got.Equal(tomorrow) {
		t.Fatalf("next due = %v, want %v", got, tomorrow)
	}
	setState(t, c, ids, scheduler.CardState{State: scheduler.Learning, Due: clk.t.Add(7 * time.Minute), Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}})
	if got := c.NextDue(deck); !got.Equal(clk.t.Add(7 * time.Minute)) {
		t.Fatalf("next due = %v", got)
	}
}

func TestPickNextInterleaves(t *testing.T) {
	q := Queue{}
	for i := 0; i < 4; i++ {
		q.New = append(q.New, ID(100+i))
	}
	for i := 0; i < 8; i++ {
		q.Review = append(q.Review, ID(200+i))
	}
	var order []string
	newShown, revShown := 0, 0
	for len(q.New)+len(q.Review) > 0 {
		id, _ := PickNext(q, newShown, revShown)
		if id < 200 {
			order = append(order, "n")
			q.New = q.New[1:]
			newShown++
		} else {
			order = append(order, "r")
			q.Review = q.Review[1:]
			revShown++
		}
	}
	got := ""
	for _, s := range order {
		got += s
	}
	if got != "nrrnrrnrrnrr" {
		t.Fatalf("order = %s", got)
	}
}

func TestDeckCountsLargeCollection(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	basic := noteType(t, c, "Basic").ID
	c.mu.Lock()
	o := &op{Kind: "bulk"}
	nt := c.noteTypes[basic]
	for i := 0; i < 50000; i++ {
		n := &Note{ID: c.ids.next(), TypeID: nt.ID, Fields: []string{"q", "a"}}
		o.Notes = append(o.Notes, n)
		card := &Card{ID: c.ids.next(), NoteID: n.ID, DeckID: deck, NewPos: int64(i)}
		if i%2 == 0 {
			card.CardState = reviewDue(clk, 3)
			card.Due = clk.t.AddDate(0, 0, i%30-10)
		}
		o.Cards = append(o.Cards, card)
	}
	c.commit(o)
	c.mu.Unlock()
	for _, orders := range [][2]string{{GatherAscending, ReviewDueRandom}, {GatherRandomCards, ReviewRetrievabilityAscending}} {
		setOrders(t, c, deck, orders[0], SortRandom, NewMix, orders[1])
		start := time.Now()
		got := c.DeckCounts()[deck]
		t.Logf("deck counts over 50,000 cards with %v in %v: %+v", orders, time.Since(start), got)
		if time.Since(start) > time.Second {
			t.Fatal("too slow")
		}
	}
}

func TestNextCountChange(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	rollover := scheduler.NextDayStart(clk.t, c.Settings().RolloverHour)
	// No learning cards: the next day start.
	addCards(t, c, deck, 1)
	if got := c.NextCountChange(); !got.Equal(rollover) {
		t.Fatalf("next = %v, want rollover %v", got, rollover)
	}
	learning := func(due time.Time) scheduler.CardState {
		return scheduler.CardState{State: scheduler.Learning, Due: due, Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}}
	}
	// A learning card due in 1 minute wins over the rollover; one already
	// due does not count.
	ids := addCards(t, c, deck, 4)
	setState(t, c, ids[:1], learning(clk.t.Add(time.Minute)))
	setState(t, c, ids[1:2], learning(clk.t.Add(-time.Minute)))
	if got := c.NextCountChange(); !got.Equal(clk.t.Add(time.Minute)) {
		t.Fatalf("next = %v, want in 1 minute", got)
	}
	// Suspended and buried learning cards are ignored.
	setState(t, c, ids[2:3], learning(clk.t.Add(30*time.Second)))
	setState(t, c, ids[3:4], learning(clk.t.Add(20*time.Second)))
	if err := c.SetSuspended(ids[2:3], true); err != nil {
		t.Fatal(err)
	}
	if err := c.BuryCards(ids[3:4]); err != nil {
		t.Fatal(err)
	}
	if got := c.NextCountChange(); !got.Equal(clk.t.Add(time.Minute)) {
		t.Fatalf("next = %v, want in 1 minute", got)
	}
	// Once it is due, the next change is the rollover again.
	clk.advance(2 * time.Minute)
	if got := c.NextCountChange(); !got.Equal(rollover) {
		t.Fatalf("next = %v, want rollover %v", got, rollover)
	}
}

// interdayDue is an interday learning card due from the start of the study
// day containing due.
func interdayDue(due time.Time) scheduler.CardState {
	return scheduler.CardState{State: scheduler.Learning, Interday: true, Step: 1, ScheduledDays: 1,
		Due: scheduler.DayStart(due, 4), Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}, Reps: 2,
		LastReview: due.AddDate(0, 0, -1)}
}

func TestInterdayLearningDueByDay(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	card := addCards(t, c, deck, 1)
	tomorrow := scheduler.NextDayStart(clk.t, 4)
	setState(t, c, card, interdayDue(tomorrow.Add(10*time.Hour)))
	q, _ := c.Queue(deck)
	if q.Counts().Total() != 0 || !q.LearningLater.IsZero() || len(q.LearningAhead) != 0 {
		t.Fatalf("before its day: %+v", q)
	}
	if next := c.NextDue(deck); !next.Equal(tomorrow) {
		t.Fatalf("next due = %v, want %v", next, tomorrow)
	}
	// Shown from the start of its day, counted as learning (spec).
	clk.t = tomorrow
	q, _ = c.Queue(deck)
	if len(q.Learning) != 0 || len(q.Review) != 1 || q.Review[0] != card[0] {
		t.Fatalf("on its day: %+v", q)
	}
	if got := counts(t, c, deck); got != (Counts{Learning: 1}) {
		t.Fatalf("counts = %+v", got)
	}
	// Still due later that day.
	clk.t = tomorrow.Add(20 * time.Hour)
	if got := counts(t, c, deck); got != (Counts{Learning: 1}) {
		t.Fatalf("counts late in the day = %+v", got)
	}
}

func TestInterdayLearningReviewLimit(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	c.SetDeckOptions(deck, deckOpts(0, 3, 0.9))
	interday := addCards(t, c, deck, 2)
	setState(t, c, interday, interdayDue(clk.t))
	setState(t, c, addCards(t, c, deck, 5), reviewDue(clk, 10))
	// Interday learning cards count against the review limit, first.
	if got := counts(t, c, deck); got != (Counts{Learning: 2, Review: 1}) {
		t.Fatalf("counts = %+v", got)
	}
	// Answering one uses up a review slot.
	if _, err := c.Answer(interday[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	if got := counts(t, c, deck); got != (Counts{Learning: 1, Review: 1}) {
		t.Fatalf("counts after answer = %+v", got)
	}
}

func TestInterdayLearningOrder(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	interday := addCards(t, c, deck, 2)
	setState(t, c, interday, interdayDue(clk.t))
	setState(t, c, addCards(t, c, deck, 10), reviewDue(clk, 10))
	positions := func(order string) []int {
		t.Helper()
		o := DefaultDeckOptions()
		o.InterdayLearningOrder = order
		if err := c.SetDeckOptions(deck, o); err != nil {
			t.Fatal(err)
		}
		q, _ := c.Queue(deck)
		if len(q.Review) != 12 || q.InterdayLearning != 2 {
			t.Fatalf("%s: %+v", order, q)
		}
		var pos []int
		for i, id := range q.Review {
			if id == interday[0] || id == interday[1] {
				pos = append(pos, i)
			}
		}
		return pos
	}
	// Before reviews (spec): both interday learning cards come first.
	if p := positions(InterdayBefore); p[0] != 0 || p[1] != 1 {
		t.Fatalf("before: %v", p)
	}
	if p := positions(InterdayAfter); p[0] != 10 || p[1] != 11 {
		t.Fatalf("after: %v", p)
	}
	if p := positions(InterdayMix); p[0] != 0 || p[1] < 4 || p[1] > 8 {
		t.Fatalf("mix: %v", p)
	}
	// Intraday learning cards still come first.
	learning := addCards(t, c, deck, 1)
	setState(t, c, learning, scheduler.CardState{State: scheduler.Learning, Due: clk.t.Add(-time.Minute),
		Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}, Reps: 1, LastReview: clk.t.Add(-11 * time.Minute)})
	positions(InterdayBefore)
	q, _ := c.Queue(deck)
	if id, _ := PickNext(q, 0, 0); id != learning[0] {
		t.Fatalf("picked %v, want the intraday learning card", id)
	}
	if got := q.Counts(); got != (Counts{Learning: 3, Review: 10}) {
		t.Fatalf("counts = %+v", got)
	}
}

func TestBuryInterdayLearningSiblings(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	cards := c.CardsOfNote(addNoteIn(t, c, deck, "Basic (and reversed card)", "hola", "hello"))
	setState(t, c, []ID{cards[1].ID}, interdayDue(clk.t))
	if _, err := c.Answer(cards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	// Off by default: the interday learning sibling stays due.
	if got := counts(t, c, deck); got.Learning != 1 {
		t.Fatalf("counts = %+v", got)
	}
	o := DefaultDeckOptions()
	o.BuryInterdayLearning = true
	c.SetDeckOptions(deck, o)
	if got := counts(t, c, deck); got.Learning != 0 {
		t.Fatalf("counts with burying = %+v", got)
	}
}

func TestTodayOnlyIncrease(t *testing.T) {
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	ids := addCards(t, c, deck, 50)
	for _, id := range ids[:20] {
		if _, err := c.Answer(id, scheduler.Easy, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	if x, _ := c.ExtraCards(deck); x.New != 30 || x.Review != 0 {
		t.Fatalf("extra = %+v", x)
	}
	// Today-only increase (spec): 20 of 20 introduced, +10 → 10 new due.
	if err := c.IncreaseTodayLimits(deck, 10, 0); err != nil {
		t.Fatal(err)
	}
	if got := counts(t, c, deck); got.New != 10 {
		t.Fatalf("new = %d, want 10", got.New)
	}
	if x, _ := c.ExtraCards(deck); x.New != 20 {
		t.Fatalf("extra = %+v", x)
	}
	if o := deckOptions(c, deck); o.NewPerDay != 20 {
		t.Fatalf("options changed: %+v", o)
	}
}

func TestTodayReviewIncrease(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	c.SetDeckOptions(deck, deckOpts(0, 3, 0.9))
	setState(t, c, addCards(t, c, deck, 10), reviewDue(clk, 5))
	if x, _ := c.ExtraCards(deck); x.Review != 7 || x.New != 0 {
		t.Fatalf("extra = %+v", x)
	}
	c.IncreaseTodayLimits(deck, 0, 4)
	if got := counts(t, c, deck); got.Review != 7 {
		t.Fatalf("reviews = %d, want 7", got.Review)
	}
}

func TestReviewLimitCapsNewCards(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	opts := deckOpts(20, 100, 0.9)
	opts.NewIgnoresReviewLimit = false
	c.SetDeckOptions(deck, opts)
	setState(t, c, addCards(t, c, deck, 95), reviewDue(clk, 5))
	news := addCards(t, c, deck, 30)
	// Review limit caps new cards (spec): 95 of 100 reviews due → 5 new.
	if got := counts(t, c, deck); got.Review != 95 || got.New != 5 {
		t.Fatalf("counts = %+v, want 95 review and 5 new", got)
	}
	// New cards introduced today use up the room too.
	if _, err := c.Answer(news[0], scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	if got := counts(t, c, deck); got.New != 4 {
		t.Fatalf("new after one introduced = %d, want 4", got.New)
	}
	// With the option on (the default), new cards ignore the review limit.
	opts.NewIgnoresReviewLimit = true
	c.SetDeckOptions(deck, opts)
	if got := counts(t, c, deck); got.New != 19 {
		t.Fatalf("new with option on = %d, want 19", got.New)
	}
}

func TestLimitsStartFromTop(t *testing.T) {
	c, _, _ := testCollection(t)
	languages, _ := c.CreateDeck("Languages")
	spanish, _ := c.CreateDeck("Languages::Spanish")
	ownPreset(t, c, languages)
	c.SetDeckOptions(languages, deckOpts(10, 200, 0.9))
	opts := deckOpts(20, 200, 0.9)
	c.SetDeckOptions(spanish, opts)
	addCards(t, c, spanish, 30)
	if got := counts(t, c, spanish); got.New != 20 {
		t.Fatalf("new without the option = %d, want 20", got.New)
	}
	// Limits from the top (spec): Languages' limit of 10 applies too.
	opts.LimitsStartFromTop = true
	c.SetDeckOptions(spanish, opts)
	if got := counts(t, c, spanish); got.New != 10 {
		t.Fatalf("new = %d, want 10", got.New)
	}
	// Raising Spanish's limit cannot pass its parent's.
	if x, _ := c.ExtraCards(spanish); x.New != 0 {
		t.Fatalf("extra = %+v", x)
	}
	c.IncreaseTodayLimits(languages, 5, 0)
	if got := counts(t, c, spanish); got.New != 15 {
		t.Fatalf("new after parent increase = %d, want 15", got.New)
	}
}

// setOrders sets a deck's display order options, keeping its limits.
func setOrders(t *testing.T, c *Collection, deck ID, gather, sortOrder, newReview, review string) {
	t.Helper()
	o := deckOptions(c, deck)
	o.NewGatherOrder, o.NewSortOrder, o.NewReviewOrder, o.ReviewSortOrder = gather, sortOrder, newReview, review
	if err := c.SetDeckOptions(deck, o); err != nil {
		t.Fatal(err)
	}
}

func newQueue(t *testing.T, c *Collection, deck ID) []ID {
	t.Helper()
	q, err := c.Queue(deck)
	if err != nil {
		t.Fatal(err)
	}
	return q.New
}

func cardsOf(t *testing.T, c *Collection, ids []ID) []Card {
	t.Helper()
	out := make([]Card, len(ids))
	for i, id := range ids {
		card, err := c.Card(id)
		if err != nil {
			t.Fatal(err)
		}
		out[i] = card
	}
	return out
}

func TestNewGatherOrders(t *testing.T) {
	c, _, _ := testCollection(t)
	parent, _ := c.CreateDeck("P")
	b, _ := c.CreateDeck("P::B")
	a, _ := c.CreateDeck("P::A")
	// Added B first, then A, then P, so positions run B, A, P.
	inB := addCards(t, c, b, 3)
	inA := addCards(t, c, a, 3)
	inP := addCards(t, c, parent, 3)
	byPosition := slices.Concat(inB, inA, inP)
	check := func(gather string, want []ID) {
		t.Helper()
		setOrders(t, c, parent, gather, SortGathered, NewMix, ReviewDueRandom)
		if got := newQueue(t, c, parent); !slices.Equal(got, want) {
			t.Errorf("%s: %v, want %v", gather, got, want)
		}
	}
	// The default order is today's: the order cards were added.
	check(GatherAscending, byPosition)
	descending := slices.Clone(byPosition)
	slices.Reverse(descending)
	check(GatherDescending, descending)
	// Deck: the studied deck's own cards, then its subdecks by name.
	check(GatherDeck, slices.Concat(inP, inA, inB))

	// Deck then random notes: grouped by deck in the same order.
	setOrders(t, c, parent, GatherDeckRandomNotes, SortGathered, NewMix, ReviewDueRandom)
	var decks []ID
	for _, card := range cardsOf(t, c, newQueue(t, c, parent)) {
		if len(decks) == 0 || decks[len(decks)-1] != card.DeckID {
			decks = append(decks, card.DeckID)
		}
	}
	if !slices.Equal(decks, []ID{parent, a, b}) {
		t.Errorf("deck then random notes: decks in order %v", decks)
	}
	// Child limits still apply with random orders.
	setOrders(t, c, parent, GatherRandomCards, SortGathered, NewMix, ReviewDueRandom)
	o := DefaultDeckOptions()
	o.NewPerDay = 1
	ownPreset(t, c, a)
	c.SetDeckOptions(a, o)
	fromA := 0
	for _, card := range cardsOf(t, c, newQueue(t, c, parent)) {
		if card.DeckID == a {
			fromA++
		}
	}
	if fromA != 1 {
		t.Errorf("random cards: %d cards from A, over its limit of 1", fromA)
	}
}

func TestRandomNotesStableWithinDay(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	o := deckOpts(10, 200, 0.9)
	o.NewGatherOrder = GatherRandomNotes
	c.SetDeckOptions(deck, o)
	all := addCards(t, c, deck, 500)
	// Random notes (spec): today's 10 new cards come from randomly chosen
	// notes, and the selection does not change after answering one.
	today := newQueue(t, c, deck)
	if len(today) != 10 || slices.Equal(today, all[:10]) {
		t.Fatalf("random notes picked %v", today)
	}
	if _, err := c.Answer(today[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	if got := newQueue(t, c, deck); !slices.Equal(got, today[1:]) {
		t.Fatalf("after answering: %v, want %v", got, today[1:])
	}
	clk.t = scheduler.NextDayStart(clk.t, 4).Add(time.Minute)
	if got := newQueue(t, c, deck); slices.Equal(got[:9], today[1:]) {
		t.Fatal("the next day has the same selection")
	}
}

func TestRandomNotesKeepSiblingsTogether(t *testing.T) {
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	burySiblings(t, c, deck)
	var notes []ID
	for i := 0; i < 20; i++ {
		notes = append(notes, addNoteIn(t, c, deck, "Basic (and reversed card)", "q", "a"))
	}
	setOrders(t, c, deck, GatherRandomNotes, SortGathered, NewMix, ReviewDueRandom)
	q := cardsOf(t, c, newQueue(t, c, deck))
	if len(q) != 20 {
		t.Fatalf("queue has %d cards", len(q))
	}
	for i := 0; i < len(q); i += 2 {
		if q[i].NoteID != q[i+1].NoteID || q[i].Ord != 0 || q[i+1].Ord != 1 {
			t.Fatalf("cards %d and %d are not one note's cards in order: %+v %+v", i, i+1, q[i], q[i+1])
		}
	}
	if q[0].NoteID == notes[0] && q[2].NoteID == notes[1] && q[4].NoteID == notes[2] {
		t.Fatal("notes are in the order added")
	}
	// Sibling burying still applies.
	if _, err := c.Answer(q[0].ID, scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	if got := newQueue(t, c, deck); slices.Contains(got, q[1].ID) {
		t.Fatal("sibling of a studied card still queued")
	}
}

func TestNewSortOrders(t *testing.T) {
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	var forward, reverse []ID
	for i := 0; i < 10; i++ {
		cards := c.CardsOfNote(addNoteIn(t, c, deck, "Basic (and reversed card)", "q", "a"))
		forward, reverse = append(forward, cards[0].ID), append(reverse, cards[1].ID)
	}
	order := func(sortOrder string) []Card {
		t.Helper()
		setOrders(t, c, deck, GatherAscending, sortOrder, NewMix, ReviewDueRandom)
		q := newQueue(t, c, deck)
		if len(q) != 20 {
			t.Fatalf("%s: %d cards", sortOrder, len(q))
		}
		if again := newQueue(t, c, deck); !slices.Equal(q, again) {
			t.Fatalf("%s: order changed between calls", sortOrder)
		}
		return cardsOf(t, c, q)
	}
	ids := func(cards []Card) []ID {
		out := make([]ID, len(cards))
		for i, card := range cards {
			out[i] = card.ID
		}
		return out
	}
	// Order gathered (default): the order added, a note's cards together.
	gathered := ids(order(SortGathered))
	for i := range forward {
		if gathered[2*i] != forward[i] || gathered[2*i+1] != reverse[i] {
			t.Fatalf("gathered: %v", gathered)
		}
	}
	if got := ids(order(SortTypeThenGathered)); !slices.Equal(got, slices.Concat(forward, reverse)) {
		t.Errorf("card type then gathered: %v", got)
	}
	typeRandom := ids(order(SortTypeThenRandom))
	if !sameSet(typeRandom[:10], forward) || !sameSet(typeRandom[10:], reverse) ||
		slices.Equal(typeRandom, slices.Concat(forward, reverse)) {
		t.Errorf("card type then random: %v", typeRandom)
	}
	noteType := order(SortRandomNoteThenType)
	for i := 0; i < 20; i += 2 {
		if noteType[i].NoteID != noteType[i+1].NoteID || noteType[i].Ord != 0 {
			t.Fatalf("random note then card type: %+v", noteType)
		}
	}
	if slices.Equal(ids(noteType), gathered) {
		t.Error("random note then card type: notes in the order added")
	}
	random := ids(order(SortRandom))
	if !sameSet(random, gathered) || slices.Equal(random, gathered) {
		t.Errorf("random: %v", random)
	}
}

func sameSet(a, b []ID) bool {
	return slices.Equal(slices.Sorted(slices.Values(a)), slices.Sorted(slices.Values(b)))
}

func TestReviewSortOrders(t *testing.T) {
	c, _, clk := testCollection(t)
	parent, _ := c.CreateDeck("P")
	child, _ := c.CreateDeck("P::B")
	review := func(deck ID, dueDaysAgo, ivl, lastDaysAgo int, mem *scheduler.Memory) ID {
		id := addCards(t, c, deck, 1)
		cs := scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, -dueDaysAgo).Add(-time.Hour),
			ScheduledDays: ivl, Memory: mem, Reps: 3}
		if lastDaysAgo > 0 {
			cs.LastReview = clk.t.AddDate(0, 0, -lastDaysAgo)
		}
		setState(t, c, id, cs)
		return id[0]
	}
	// Retrievability: r1 0.75, r2 0.86, r4 0.92, r3 0.99, r5 none.
	// Relative overdueness: r3 2.0, r2 1.5, r1 1.3, r5 1.01, r4 1.0.
	r1 := review(child, 3, 10, 13, &scheduler.Memory{Stability: 4, Difficulty: 8})
	r2 := review(parent, 1, 2, 3, &scheduler.Memory{Stability: 2, Difficulty: 3})
	r3 := review(child, 1, 1, 2, &scheduler.Memory{Stability: 20, Difficulty: 5})
	r4 := review(parent, 0, 30, 30, &scheduler.Memory{Stability: 40, Difficulty: 6})
	r5 := review(parent, 0, 5, 0, nil)
	order := func(o string) []ID {
		t.Helper()
		setOrders(t, c, parent, GatherAscending, SortGathered, NewMix, o)
		q, _ := c.Queue(parent)
		if again, _ := c.Queue(parent); !slices.Equal(q.Review, again.Review) {
			t.Fatalf("%s: order changed between calls", o)
		}
		return q.Review
	}
	// in checks a queue against groups of cards, each group in any order.
	in := func(o string, groups ...[]ID) {
		t.Helper()
		got := order(o)
		for _, g := range groups {
			if len(got) < len(g) || !sameSet(got[:len(g)], g) {
				t.Errorf("%s: %v, want groups %v", o, order(o), groups)
				return
			}
			got = got[len(g):]
		}
	}
	in(ReviewDueRandom, []ID{r1}, []ID{r2, r3}, []ID{r4, r5})
	in(ReviewDueDeck, []ID{r1}, []ID{r2}, []ID{r3}, []ID{r4, r5})
	in(ReviewDeckDue, []ID{r2}, []ID{r4, r5}, []ID{r1}, []ID{r3})
	for _, tc := range []struct {
		order string
		want  []ID
	}{
		{ReviewIntervalsAscending, []ID{r3, r2, r5, r1, r4}},
		{ReviewIntervalsDescending, []ID{r4, r1, r5, r2, r3}},
		{ReviewDifficultyAscending, []ID{r2, r3, r4, r1, r5}},
		{ReviewDifficultyDescending, []ID{r1, r4, r3, r2, r5}},
		// Most at risk first (spec): least likely to be recalled first.
		{ReviewRetrievabilityAscending, []ID{r1, r2, r4, r3, r5}},
		{ReviewRelativeOverdueness, []ID{r3, r2, r1, r5, r4}},
	} {
		if got := order(tc.order); !slices.Equal(got, tc.want) {
			t.Errorf("%s: %v, want %v", tc.order, got, tc.want)
		}
	}
	// The review sort order picks which reviews fill the limit; interday
	// learning cards still take it first.
	interday := addCards(t, c, parent, 1)
	setState(t, c, interday, interdayDue(clk.t))
	o := DefaultDeckOptions()
	o.ReviewsPerDay, o.ReviewSortOrder = 3, ReviewRetrievabilityAscending
	c.SetDeckOptions(parent, o)
	if q, _ := c.Queue(parent); !sameSet(q.Review, []ID{interday[0], r1, r2}) {
		t.Errorf("limited to 3: %v", q.Review)
	}
}

func TestNewReviewOrder(t *testing.T) {
	// study answers the deck's cards as a session does and returns the
	// kinds shown, n for new and r for review.
	study := func(newReview string) string {
		t.Helper()
		c, _, clk := testCollection(t)
		deck := defaultDeck(c)
		setState(t, c, addCards(t, c, deck, 10), reviewDue(clk, 10))
		addCards(t, c, deck, 5)
		setOrders(t, c, deck, GatherAscending, SortGathered, newReview, ReviewDueRandom)
		shown, newShown, reviewShown := "", 0, 0
		for {
			q, _ := c.Queue(deck)
			id, ok := PickNext(q, newShown, reviewShown)
			if !ok {
				return shown
			}
			if slices.Contains(q.New, id) {
				shown, newShown = shown+"n", newShown+1
			} else {
				shown, reviewShown = shown+"r", reviewShown+1
			}
			if _, err := c.Answer(id, scheduler.Easy, time.Second); err != nil {
				t.Fatal(err)
			}
		}
	}
	// New cards after reviews (spec): all 10 review cards come first.
	if got := study(NewAfterReviews); got != "rrrrrrrrrrnnnnn" {
		t.Errorf("after: %s", got)
	}
	if got := study(NewBeforeReviews); got != "nnnnnrrrrrrrrrr" {
		t.Errorf("before: %s", got)
	}
	if got := study(NewMix); got != "nrrnrrnrrnrrnrr" {
		t.Errorf("mix: %s", got)
	}
}

func TestDeckLimitOverridesPreset(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	french, _ := c.CreateDeck("French")
	preset, _ := c.AddPreset(spanish, "Languages")
	c.SetDeckPreset(french, preset)
	addCards(t, c, spanish, 30)
	addCards(t, c, french, 30)
	// One deck lower (spec): French shows at most 5 new cards, Spanish up
	// to its preset's 20.
	five := 5
	if err := c.SetDeckLimits(french, DeckLimits{New: &five}); err != nil {
		t.Fatal(err)
	}
	if got := counts(t, c, french); got.New != 5 {
		t.Fatalf("French new = %d", got.New)
	}
	if got := counts(t, c, spanish); got.New != 20 {
		t.Fatalf("Spanish new = %d", got.New)
	}
	// Removing the deck limit uses the preset's again.
	if err := c.SetDeckLimits(french, DeckLimits{}); err != nil {
		t.Fatal(err)
	}
	if got := counts(t, c, french); got.New != 20 {
		t.Fatalf("French new without its limit = %d", got.New)
	}
}

func overview(t *testing.T, c *Collection, deck ID) Overview {
	t.Helper()
	o, err := c.Overview(deck)
	if err != nil {
		t.Fatal(err)
	}
	if o.Counts != counts(t, c, deck) {
		t.Fatalf("overview counts %+v != queue counts", o.Counts)
	}
	return o
}

// The buried count of a deck tree includes manually buried cards and cards
// held back by sibling burying that would otherwise be due, but not
// suspended cards or siblings that are not due.
func TestBuriedCount(t *testing.T) {
	c, _, clk := testCollection(t)
	parent, _ := c.CreateDeck("Spanish")
	child, _ := c.CreateDeck("Spanish::Verbs")
	burySiblings(t, c, child)
	reviews := addCards(t, c, parent, 3)
	setState(t, c, reviews, reviewDue(clk, 5))
	if err := c.BuryCards(reviews[:2]); err != nil {
		t.Fatal(err)
	}
	// A suspended buried card is not counted.
	if err := c.SetSuspended(reviews[1:2], true); err != nil {
		t.Fatal(err)
	}
	note := addNoteIn(t, c, child, "Basic (and reversed card)", "hablar", "to speak")
	cards := c.CardsOfNote(note)
	if _, err := c.Answer(cards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	// A review sibling due next week is not held back today.
	later := addNoteIn(t, c, child, "Basic (and reversed card)", "comer", "to eat")
	laterCards := c.CardsOfNote(later)
	if _, err := c.Answer(laterCards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	st := reviewDue(clk, 5)
	st.Due = clk.t.AddDate(0, 0, 7)
	setState(t, c, []ID{laterCards[1].ID}, st)

	if o := overview(t, c, parent); o.Buried != 2 || o.Counts.Review != 1 || o.Counts.New != 0 {
		t.Fatalf("parent overview = %+v", o)
	}
	if o := overview(t, c, child); o.Buried != 1 {
		t.Fatalf("child overview = %+v", o)
	}
	// Both kinds of burying end with the study day.
	clk.t = scheduler.NextDayStart(clk.t, 4).Add(time.Minute)
	if o := overview(t, c, parent); o.Buried != 0 {
		t.Fatalf("overview next day = %+v", o)
	}
	if _, err := c.Overview(12345); err == nil {
		t.Fatal("overview of a missing deck")
	}
}

// Unbury manually buried cards (collection spec): the cards are due again
// and the buried count is 0, in the deck and its subdecks only.
func TestUnburyManual(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	verbs, _ := c.CreateDeck("Spanish::Verbs")
	other, _ := c.CreateDeck("French")
	ids := addCards(t, c, spanish, 2)
	ids = append(ids, addCards(t, c, verbs, 1)...)
	french := addCards(t, c, other, 1)
	setState(t, c, append(ids, french...), reviewDue(clk, 5))
	if err := c.BuryCards(append(ids, french...)); err != nil {
		t.Fatal(err)
	}
	if o := overview(t, c, spanish); o.Buried != 3 || o.Counts.Review != 0 {
		t.Fatalf("overview = %+v", o)
	}
	clk.advance(time.Minute)
	if err := c.Unbury(spanish); err != nil {
		t.Fatal(err)
	}
	if o := overview(t, c, spanish); o.Buried != 0 || o.Counts.Review != 3 {
		t.Fatalf("overview after unbury = %+v", o)
	}
	if o := overview(t, c, other); o.Buried != 1 {
		t.Fatalf("other deck = %+v", o)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Unbury" {
		t.Fatalf("undo = %q", undo)
	}
	c = reopen(t, c, store, clk)
	if err := c.Unbury(12345); err == nil {
		t.Fatal("unbury of a missing deck")
	}
}

// Unbury siblings (collection spec): the reverse card of a note studied
// today is available again; answers after the unbury bury siblings again,
// and the marker has no effect the next day.
func TestUnburySiblings(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	burySiblings(t, c, deck)
	note := addNoteIn(t, c, deck, "Basic (and reversed card)", "hola", "hello")
	cards := c.CardsOfNote(note)
	if _, err := c.Answer(cards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	if o := overview(t, c, deck); o.Buried != 1 || o.Counts.New != 0 {
		t.Fatalf("overview = %+v", o)
	}
	clk.advance(time.Minute)
	if err := c.Unbury(deck); err != nil {
		t.Fatal(err)
	}
	q, _ := c.Queue(deck)
	if len(q.New) != 1 || q.New[0] != cards[1].ID {
		t.Fatalf("new after unbury = %v", q.New)
	}
	if o := overview(t, c, deck); o.Buried != 0 {
		t.Fatalf("overview after unbury = %+v", o)
	}
	// Undo buries the sibling again.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if o := overview(t, c, deck); o.Buried != 1 {
		t.Fatalf("overview after undo = %+v", o)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}

	// A note studied after the unbury buries its siblings as usual.
	other := addNoteIn(t, c, deck, "Basic (and reversed card)", "adiós", "bye")
	otherCards := c.CardsOfNote(other)
	clk.advance(time.Minute)
	if _, err := c.Answer(otherCards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	q, _ = c.Queue(deck)
	if slices.Contains(q.New, otherCards[1].ID) || !slices.Contains(q.New, cards[1].ID) {
		t.Fatalf("new = %v", q.New)
	}

	// The marker expires by itself: the next day, an answer buries its
	// siblings again.
	clk.t = scheduler.NextDayStart(clk.t, 4).Add(time.Minute)
	third := addNoteIn(t, c, deck, "Basic (and reversed card)", "gracias", "thanks")
	thirdCards := c.CardsOfNote(third)
	if _, err := c.Answer(thirdCards[0].ID, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	q, _ = c.Queue(deck)
	if slices.Contains(q.New, thirdCards[1].ID) {
		t.Fatalf("sibling not buried the next day: %v", q.New)
	}
	if o := overview(t, c, deck); o.Buried != 1 {
		t.Fatalf("overview next day = %+v", o)
	}
}
