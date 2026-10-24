package core

import (
	"cardflip/golang/internal/usermsg"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// fittedWeights are FSRS parameters fitted to a real collection (fsrs-rs's
// sample), far from the defaults: an Easy first rating means a stability
// of 9.9 days instead of 15.7.
var fittedWeights = []float64{
	0.6032805, 1.3376843, 4.4167747, 9.933699, 7.654044, 0.78219295, 2.336606, 0.001,
	1.3264198, 0.12967199, 0.82880765, 1.9360433, 0.13298263, 0.27427456, 2.4304862,
	0.10340813, 3.108867, 0.2114512, 0.2826002,
}

// TestWeightsChangeIntervals: cards are scheduled with their deck's FSRS
// parameters (scheduling spec, "FSRS scheduling").
func TestWeightsChangeIntervals(t *testing.T) {
	c, _, _ := testCollection(t)
	plain := defaultDeck(c)
	fitted, err := c.CreateDeck("Fitted")
	if err != nil {
		t.Fatal(err)
	}
	ownPreset(t, c, fitted)
	opts := deckOptions(c, fitted)
	opts.Weights = fittedWeights
	if err := c.SetDeckOptions(fitted, opts); err != nil {
		t.Fatal(err)
	}
	easy := func(deck ID) time.Duration {
		t.Helper()
		card := c.CardsOfNote(addNoteIn(t, c, deck, "Basic", "q", "a"))[0]
		p, err := c.Previews(card.ID)
		if err != nil {
			t.Fatal(err)
		}
		return p[3].Interval
	}
	// About 16 and 10 days, each with interval fuzz.
	const day = 24 * time.Hour
	if got := easy(plain); got < 13*day || got > 19*day {
		t.Errorf("Easy with default parameters = %v, want 13 to 19 days", got)
	}
	if got := easy(fitted); got < 8*day || got > 12*day {
		t.Errorf("Easy with fitted parameters = %v, want 8 to 12 days", got)
	}
}

func TestFSRSOptionsValidation(t *testing.T) {
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	opts := DefaultDeckOptions()
	opts.Weights = fittedWeights
	opts.OptimizeSearch = "deck:French -is:suspended"
	opts.IgnoreBefore = "2026-01-31"
	opts.HistoricalRetention = 0.8
	opts.HealthCheck, opts.RescheduleOnChange = true, true
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	if o := deckOptions(c, deck); !o.Equal(opts) {
		t.Fatalf("options = %+v", o)
	}
	for _, tc := range []struct {
		change func(*DeckOptions)
		msg    string
	}{
		{func(o *DeckOptions) { o.Weights = fittedWeights[:17] }, "must be 19 numbers"},
		{func(o *DeckOptions) { o.Weights = append([]float64{-1}, fittedWeights[1:]...) }, "parameter 1 must be between"},
		{func(o *DeckOptions) { o.HistoricalRetention = 0.49 }, "between 0.50 and 0.99"},
		{func(o *DeckOptions) { o.HistoricalRetention = 1 }, "between 0.50 and 0.99"},
		{func(o *DeckOptions) { o.IgnoreBefore = "31.01.2026" }, "not a date"},
		{func(o *DeckOptions) { o.OptimizeSearch = "deck:(" }, "search"},
	} {
		bad := opts
		tc.change(&bad)
		mustUserErr(t, c.SetDeckOptions(deck, bad), tc.msg)
	}
	if o := deckOptions(c, deck); !o.Equal(opts) {
		t.Fatalf("rejected options changed them: %+v", o)
	}
}

// TestStoredWeightsOfOtherModelAreDropped: parameters of another model
// version (another length) are treated as absent (design decision 2).
func TestStoredWeightsOfOtherModelAreDropped(t *testing.T) {
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"weights":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17]}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.Weights != nil || o.HistoricalRetention != DefaultHistoricalRetention {
		t.Fatalf("options = %+v", o)
	}
	raw, _ := json.Marshal(DeckOptions{Weights: fittedWeights})
	if err := json.Unmarshal(raw, &o); err != nil || len(o.Weights) != scheduler.NumWeights {
		t.Fatalf("weights = %v, %v", o.Weights, err)
	}
}

// reviewCard adds a note to deck and studies its card into a review card
// with a few reviews of history.
func reviewCard(t *testing.T, c *Collection, clk *clock, deck ID) ID {
	t.Helper()
	card := c.CardsOfNote(addNoteIn(t, c, deck, "Basic", "q", "a"))[0].ID
	for range 2 {
		if _, err := c.Answer(card, scheduler.Good, time.Second); err != nil {
			t.Fatal(err)
		}
		clk.advance(15 * time.Minute)
	}
	for range 3 {
		got, _ := c.Card(card)
		clk.t = got.Due.Add(time.Hour)
		if _, err := c.Answer(card, scheduler.Good, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := c.Card(card); got.State != scheduler.Review {
		t.Fatalf("card is %v", got.State)
	}
	return card
}

func saveOptions(t *testing.T, c *Collection, deck ID, opts DeckOptions) {
	t.Helper()
	d, _ := c.Deck(deck)
	if err := c.SaveDeckOptions(deck, d.PresetID, opts, DeckLimits{}); err != nil {
		t.Fatal(err)
	}
}

// TestLowerRetentionReschedules: with "Reschedule cards on change", lowering
// desired retention gives the deck's review cards later due dates, after
// a count for the confirmation; undo restores them.
func TestLowerRetentionReschedules(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	opts := deckOptions(c, deck)
	opts.RescheduleOnChange = true
	saveOptions(t, c, deck, opts)
	card := reviewCard(t, c, clk, deck)
	newCard := c.CardsOfNote(addNoteIn(t, c, deck, "Basic", "new", "a"))[0].ID
	before, _ := c.Card(card)

	lower := opts
	lower.DesiredRetention = 0.85
	d, _ := c.Deck(deck)
	if n, err := c.RescheduleCount(deck, d.PresetID, lower); err != nil || n != 1 {
		t.Fatalf("count = %d, %v; want 1", n, err)
	}
	saveOptions(t, c, deck, lower)
	after, _ := c.Card(card)
	if !after.Due.After(before.Due) || after.ScheduledDays <= before.ScheduledDays {
		t.Fatalf("due %v (%d days) → %v (%d days)", before.Due, before.ScheduledDays, after.Due, after.ScheduledDays)
	}
	if !after.Due.Equal(after.LastReview.AddDate(0, 0, after.ScheduledDays)) {
		t.Errorf("due %v is not %d days after the last review %v", after.Due, after.ScheduledDays, after.LastReview)
	}
	if n, _ := c.Card(newCard); n.State != scheduler.New {
		t.Errorf("new card changed: %+v", n)
	}
	// Saving without a change reschedules nothing.
	if n, _ := c.RescheduleCount(deck, d.PresetID, lower); n != 0 {
		t.Errorf("count without change = %d", n)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if undone, _ := c.Card(card); !undone.Due.Equal(before.Due) || deckOptions(c, deck).DesiredRetention != 0.9 {
		t.Errorf("undo: due %v, want %v", undone.Due, before.Due)
	}
}

// TestRescheduleOptionOff: without the option, changing retention or
// parameters leaves due dates alone until cards are answered.
func TestRescheduleOptionOff(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	card := reviewCard(t, c, clk, deck)
	before, _ := c.Card(card)
	opts := deckOptions(c, deck)
	opts.DesiredRetention = 0.8
	opts.Weights = fittedWeights
	d, _ := c.Deck(deck)
	if n, _ := c.RescheduleCount(deck, d.PresetID, opts); n != 0 {
		t.Fatalf("count = %d, want 0", n)
	}
	saveOptions(t, c, deck, opts)
	if after, _ := c.Card(card); !after.Due.Equal(before.Due) || *after.Memory != *before.Memory {
		t.Fatalf("card changed: %+v → %+v", before.CardState, after.CardState)
	}
}

// TestNewWeightsReplayHistory: new parameters replay the review history
// into new memory states.
func TestNewWeightsReplayHistory(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	opts := deckOptions(c, deck)
	opts.RescheduleOnChange = true
	saveOptions(t, c, deck, opts)
	card := reviewCard(t, c, clk, deck)
	opts.Weights = fittedWeights
	saveOptions(t, c, deck, opts)
	got, _ := c.Card(card)
	want, ok := scheduler.NewSequence(cardHistory(t, c, card), scheduler.HistoryOptions{RolloverHour: 4}).Memory(fittedWeights)
	if !ok || *got.Memory != want {
		t.Fatalf("memory = %+v, want %+v", *got.Memory, want)
	}
}

func cardHistory(t *testing.T, c *Collection, card ID) []scheduler.HistoryReview {
	t.Helper()
	h, err := c.histories(map[ID]bool{card: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(h[card]) != 5 {
		t.Fatalf("history has %d entries", len(h[card]))
	}
	return h[card]
}

// studyCards adds n cards to deck and gives each a learning day and three
// reviews (three reviews to learn from), answering Again now and then.
func studyCards(t *testing.T, c *Collection, clk *clock, deck ID, n int) {
	t.Helper()
	start := clk.t
	var cards []ID
	for i := range n {
		cards = append(cards, c.CardsOfNote(addNoteIn(t, c, deck, "Basic", fmt.Sprintf("q%d", i), "a"))[0].ID)
	}
	for i, id := range cards {
		clk.t = start
		for range 2 {
			if _, err := c.Answer(id, scheduler.Good, time.Second); err != nil {
				t.Fatal(err)
			}
			clk.advance(15 * time.Minute)
		}
		for r := range 3 {
			got, _ := c.Card(id)
			clk.t = got.Due.Add(time.Hour)
			rating := scheduler.Good
			if (i+r)%7 == 0 {
				rating = scheduler.Again
			}
			if _, err := c.Answer(id, rating, time.Second); err != nil {
				t.Fatal(err)
			}
			if rating == scheduler.Again {
				clk.advance(15 * time.Minute)
				c.Answer(id, scheduler.Good, time.Second)
			}
		}
	}
}

func TestOptimizeDeck(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	studyCards(t, c, clk, deck, 150)
	opts := deckOptions(c, deck)
	var stages []string
	res, err := c.Optimize(context.Background(), deck, opts, func(stage string, done, total int) {
		if len(stages) == 0 || stages[len(stages)-1] != stage {
			stages = append(stages, stage)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reviews != 450 || len(res.Weights) != scheduler.NumWeights {
		t.Fatalf("result = %+v", res)
	}
	if !slices.Equal(stages, []string{"reading", "training"}) {
		t.Errorf("stages = %v", stages)
	}
	// Optimizing saves nothing.
	if o := deckOptions(c, deck); len(o.Weights) != 0 {
		t.Errorf("weights saved: %v", o.Weights)
	}

	// A search selecting fewer cards has too few reviews.
	opts.OptimizeSearch = "q1*"
	_, err = c.Optimize(context.Background(), deck, opts, func(string, int, int) {})
	var nr *scheduler.NotEnoughReviewsError
	if !errors.As(err, &nr) || nr.Reviews != 3*61 {
		t.Fatalf("err = %v", err)
	}
}

func TestOptimizeAll(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	studyCards(t, c, clk, deck, 150)
	other, _ := c.CreateDeck("Other")
	ownPreset(t, c, other)
	studyCards(t, c, clk, other, 10)
	results, err := c.OptimizeAll(context.Background(), func(string, int, int) {})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %+v", results)
	}
	for _, r := range results {
		switch r.Preset.Name {
		case DefaultPresetName:
			if r.NotEnough != nil || r.Applied != r.Result.Use() {
				t.Errorf("Default: %+v", r)
			}
			if got := deckOptions(c, deck).Weights; r.Applied && !slices.Equal(got, r.Result.Weights) {
				t.Errorf("Default weights = %v, want %v", got, r.Result.Weights)
			}
		default:
			if r.NotEnough == nil || r.Applied || len(deckOptions(c, other).Weights) != 0 {
				t.Errorf("%s: %+v", r.Preset.Name, r)
			}
		}
	}
	if !results[0].Applied && !results[1].Applied {
		t.Fatalf("nothing applied: %+v", results)
	}
	if label, err := c.Undo(); err != nil || usermsg.Plain(label) != "Optimize all presets" {
		t.Fatalf("undo = %q, %v", label, err)
	}
	if len(deckOptions(c, deck).Weights) != 0 {
		t.Error("undo kept the parameters")
	}
}
