package core

import (
	"fmt"
	"maps"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// checkLoad checks that the kept per-day counts equal a fresh count of the
// review cards' due days, and returns them.
func checkLoad(t *testing.T, c *Collection, what string) map[int]int {
	t.Helper()
	c.mu.RLock()
	defer c.mu.RUnlock()
	want := map[int]int{}
	for _, card := range c.cards {
		if !card.Suspended && (card.State == scheduler.Review || card.Interday) {
			want[scheduler.DayNumber(card.Due.In(c.now().Location()), c.settings.RolloverHour)]++
		}
	}
	if !maps.Equal(c.load.days, want) {
		t.Fatalf("%s: load %v, want %v", what, c.load.days, want)
	}
	return maps.Clone(want)
}

func total(load map[int]int) int {
	n := 0
	for _, v := range load {
		n += v
	}
	return n
}

// The per-day counts follow answers, undo, redo, suspending, due date
// changes, deletions, imports, rollover changes and a reopen.
func TestDueLoadKeptUpToDate(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := c.Settings().LastDeckID
	var cards []ID
	for range 3 {
		cards = append(cards, reviewCard(t, c, clk, deck))
	}
	if load := checkLoad(t, c, "after answers"); total(load) != 3 {
		t.Fatalf("load %v, want 3 cards", load)
	}

	before := checkLoad(t, c, "before")
	got, _ := c.Card(cards[0])
	clk.t = got.Due.Add(time.Hour)
	if _, err := c.Answer(cards[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	after := checkLoad(t, c, "after answer")
	if maps.Equal(after, before) {
		t.Fatal("answer did not move the card's due day")
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := checkLoad(t, c, "after undo"); !maps.Equal(got, before) {
		t.Fatalf("after undo %v, want %v", got, before)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if got := checkLoad(t, c, "after redo"); !maps.Equal(got, after) {
		t.Fatalf("after redo %v, want %v", got, after)
	}

	if err := c.SetSuspended(cards[1:2], true); err != nil {
		t.Fatal(err)
	}
	if load := checkLoad(t, c, "after suspend"); total(load) != 2 {
		t.Fatalf("suspended card still counted: %v", load)
	}
	if err := c.SetDueDate(cards[2:], "5"); err != nil {
		t.Fatal(err)
	}
	checkLoad(t, c, "after set due date")
	note := c.mustCard(t, cards[2]).NoteID
	if err := c.DeleteNotes([]ID{note}); err != nil {
		t.Fatal(err)
	}
	if load := checkLoad(t, c, "after delete"); total(load) != 1 {
		t.Fatalf("deleted card still counted: %v", load)
	}

	if _, err := c.Import(historyBatch(c, "imported", 5), nil); err != nil {
		t.Fatal(err)
	}
	if load := checkLoad(t, c, "after import"); total(load) != 2 {
		t.Fatalf("imported card not counted: %v", load)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if load := checkLoad(t, c, "after undoing the import"); total(load) != 1 {
		t.Fatalf("undone import still counted: %v", load)
	}

	s := c.Settings()
	s.RolloverHour = 23
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	checkLoad(t, c, "after rollover change")
	checkLoad(t, reopen(t, c, store, clk), "after reopen")
}

// The deck's Easy Days levels and the counts reach the scheduler.
func TestParamsCarryEasyDays(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := c.Settings().LastDeckID
	card := reviewCard(t, c, clk, deck)
	opts, _ := c.DeckOptions(deck)
	opts.EasyDays = [7]int{5: EasyReduced, 6: EasyMinimum}
	saveOptions(t, c, deck, opts)
	got := c.mustCard(t, card)
	p := c.params(&got)
	if p.EasyDays != opts.EasyDays {
		t.Fatalf("easy days = %v", p.EasyDays)
	}
	due := scheduler.DayNumber(got.Due.In(c.now().Location()), p.RolloverHour)
	if p.DueLoad == nil || p.DueLoad(due) != 1 || p.DueLoad(due+1) != 0 {
		t.Fatal("due load does not count the card")
	}
	if o, _ := c.DeckOptions(deck); o.EasyDays != opts.EasyDays {
		t.Fatalf("stored easy days = %v", o.EasyDays)
	}
	opts.EasyDays[0] = 3
	d, _ := c.Deck(deck)
	if err := c.SaveDeckOptions(deck, d.PresetID, opts, DeckLimits{}); err == nil {
		t.Fatal("an unknown level was accepted")
	}
}

// simulateMonth imports 300 review cards due over the next ten days, then
// for 30 days answers every due card Good at noon. It returns how many
// answers were scheduled on each weekday.
func simulateMonth(t *testing.T, easyDays [7]int) map[time.Weekday]int {
	t.Helper()
	c, _, clk := testCollection(t)
	now := c.Now()
	b := &ImportBatch{NoteTypes: []ImportNoteType{{Key: "b", ExistingID: noteType(t, c, "Basic").ID}}, Duplicates: DupDuplicate}
	for i := range 300 {
		ivl := 5 + i%20
		due := now.AddDate(0, 0, i%10)
		state := scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: ivl, Reps: 4,
			Memory: &scheduler.Memory{Stability: float64(ivl), Difficulty: 5}, LastReview: due.AddDate(0, 0, -ivl)}
		b.Notes = append(b.Notes, ImportNote{TypeKey: "b", DeckName: "Spanish", Fields: []string{fmt.Sprintf("q%d", i), "a"},
			Cards: []ImportCard{{DeckName: "Spanish", State: state}}})
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	var deck ID
	for _, d := range c.Decks() {
		if d.Name == "Spanish" {
			deck = d.ID
		}
	}
	opts, _ := c.DeckOptions(deck)
	opts.EasyDays = easyDays
	saveOptions(t, c, deck, opts)

	scheduled := map[time.Weekday]int{}
	for range 30 {
		dayEnd := scheduler.NextDayStart(clk.t, c.Settings().RolloverHour)
		for _, card := range cardsIn(c, deck) {
			if card.State != scheduler.Review || !card.Due.Before(dayEnd) {
				continue
			}
			if _, err := c.Answer(card.ID, scheduler.Good, time.Second); err != nil {
				t.Fatal(err)
			}
			due := c.mustCard(t, card.ID).Due
			scheduled[scheduler.DayStart(due, c.Settings().RolloverHour).Weekday()]++
		}
		checkLoad(t, c, "during the simulation")
		clk.advance(24 * time.Hour)
	}
	return scheduled
}

// 30 days of reviews with Sunday as Minimum leave Sunday without reviews
// from answers (every interval here is long enough to move), where every
// day Normal gives Sunday its share.
func TestEasyDaysMonthSimulation(t *testing.T) {
	normal := simulateMonth(t, [7]int{})
	easy := simulateMonth(t, [7]int{6: EasyMinimum})
	total := 0
	for _, n := range normal {
		total += n
	}
	t.Logf("all Normal: %v", normal)
	t.Logf("Sunday Minimum: %v", easy)
	if normal[time.Sunday] < total/14 {
		t.Fatalf("all Normal gave Sunday only %d of %d", normal[time.Sunday], total)
	}
	if easy[time.Sunday] != 0 {
		t.Fatalf("Sunday Minimum still got %d reviews", easy[time.Sunday])
	}
}
