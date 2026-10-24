package scheduler

import (
	"context"
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"
)

// simDeck is a synthetic deck: studied cards with stabilities from 1 to 100
// days, due over the next weeks, and new cards.
func simDeck(studied, newCards int) SimDeck {
	rng := rand.New(rand.NewPCG(1, 2))
	d := SimDeck{NewCards: newCards}
	for range studied {
		s := 1 + rng.Float64()*99
		last := -rng.IntN(30)
		d.Cards = append(d.Cards, SimCard{Stability: s, Difficulty: 1 + rng.Float64()*9, LastDay: last,
			DueDay: last + int(s)})
	}
	return d
}

func simConfig(retention float64) SimConfig {
	return SimConfig{Days: 365, NewPerDay: 20, ReviewsPerDay: 9999, DesiredRetention: retention,
		MaximumInterval: 36500, MinimumRelearnInterval: 1, LearningSteps: 2, RelearningSteps: 1,
		Behaviour: DefaultBehaviour(), Seed: 42}
}

func sum[T int | float64](xs []T) T {
	var t T
	for _, x := range xs {
		t += x
	}
	return t
}

// Compare retentions (spec): a higher desired retention means more reviews.
func TestSimulateHigherRetentionMoreReviews(t *testing.T) {
	deck := simDeck(1000, 2000)
	var prev int
	for i, r := range []float64{0.80, 0.90, 0.95} {
		res, err := Simulate(context.Background(), deck, simConfig(r))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Reviews) != 365 || len(res.Memorized) != 365 {
			t.Fatalf("days = %d", len(res.Reviews))
		}
		total := sum(res.Reviews)
		if i > 0 && total <= prev {
			t.Errorf("retention %.2f: %d reviews, not more than %d", r, total, prev)
		}
		prev = total
	}
}

func TestSimulateIsReproducible(t *testing.T) {
	deck := simDeck(300, 300)
	a, _ := Simulate(context.Background(), deck, simConfig(0.9))
	b, _ := Simulate(context.Background(), deck, simConfig(0.9))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed, different results")
	}
	cfg := simConfig(0.9)
	cfg.Seed = 7
	if c, _ := Simulate(context.Background(), deck, cfg); reflect.DeepEqual(a, c) {
		t.Fatal("another seed gave the same result")
	}
	if len(deck.Cards) != 300 || deck.Cards[0].LastDay > 0 {
		t.Fatal("simulating changed the deck")
	}
}

func TestSimulateWithoutNewCards(t *testing.T) {
	deck := simDeck(500, 1000)
	cfg := simConfig(0.9)
	cfg.NewPerDay = 0
	res, err := Simulate(context.Background(), deck, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if n := sum(res.New); n != 0 {
		t.Fatalf("%d new cards learned", n)
	}
	for d, m := range res.Memorized {
		if m <= 0 || m > 500 {
			t.Fatalf("day %d: %.1f memorised of 500 cards", d, m)
		}
	}
	// Learning new cards adds to what is memorised.
	cfg.NewPerDay = 10
	more, _ := Simulate(context.Background(), deck, cfg)
	if sum(more.New) != 1000 || more.Memorized[364] <= res.Memorized[364] {
		t.Fatalf("new = %d, memorised %.0f vs %.0f", sum(more.New), more.Memorized[364], res.Memorized[364])
	}
}

func TestSimulateLimits(t *testing.T) {
	deck := simDeck(2000, 1000)
	cfg := simConfig(0.9)
	cfg.ReviewsPerDay = 50
	res, _ := Simulate(context.Background(), deck, cfg)
	for d := range res.Reviews {
		if res.Reviews[d] > 50 || res.Reviews[d]+res.New[d] > 50 {
			t.Fatalf("day %d: %d reviews and %d new over the limit", d, res.Reviews[d], res.New[d])
		}
	}
	// Day 0 has a large backlog: the limit is used up by reviews.
	if res.Reviews[0] != 50 || res.New[0] != 0 {
		t.Fatalf("day 0: %d reviews, %d new", res.Reviews[0], res.New[0])
	}
	cfg.NewIgnoresReviewLimit = true
	res, _ = Simulate(context.Background(), deck, cfg)
	if res.Reviews[0] != 50 || res.New[0] != 20 {
		t.Fatalf("ignoring the review limit, day 0: %d reviews, %d new", res.Reviews[0], res.New[0])
	}
	// The maximum interval keeps every card coming back.
	cfg = simConfig(0.9)
	cfg.MaximumInterval = 7
	res, _ = Simulate(context.Background(), SimDeck{Cards: []SimCard{{Stability: 100, Difficulty: 5, DueDay: 0}}}, cfg)
	if n := sum(res.Reviews); n < 365/7 {
		t.Fatalf("%d reviews with a 7-day maximum", n)
	}
}

func TestSimulateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Simulate(ctx, simDeck(10, 10), simConfig(0.9)); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

// Help me decide (spec): on a deck where lapses are expensive and reviews
// cheap, the cost per memorised card falls and then rises again, so the
// suggestion lies inside the range.
func TestSuggestRetentionInterior(t *testing.T) {
	deck := simDeck(500, 3000)
	cfg := simConfig(0)
	cfg.Behaviour.ReviewSeconds = [4]float64{120, 6, 5, 4}
	var calls []int
	s, err := SuggestRetention(context.Background(), deck, cfg, func(done, total int) {
		if total != 30 {
			t.Fatalf("total = %d", total)
		}
		calls = append(calls, done)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 31 || calls[30] != 30 {
		t.Fatalf("progress = %v", calls)
	}
	if len(s.Points) != 30 || s.Points[0].Retention != 0.70 || s.Points[29].Retention != 0.99 {
		t.Fatalf("points = %+v", s.Points)
	}
	if s.Retention <= 0.70 || s.Retention >= 0.99 {
		t.Fatalf("suggested %.2f, want inside the range; points %+v", s.Retention, s.Points)
	}
	var best RetentionPoint
	for _, p := range s.Points {
		if p.Retention == s.Retention {
			best = p
		}
	}
	if best.SecondsPerCard >= s.Points[0].SecondsPerCard || best.SecondsPerCard >= s.Points[29].SecondsPerCard {
		t.Fatalf("suggestion %+v not cheaper than the ends %+v %+v", best, s.Points[0], s.Points[29])
	}
	again, _ := SuggestRetention(context.Background(), deck, cfg, func(int, int) {})
	if !reflect.DeepEqual(s, again) {
		t.Fatal("suggestion not reproducible")
	}
}

func TestSuggestRetentionEmptyDeck(t *testing.T) {
	_, err := SuggestRetention(context.Background(), SimDeck{}, simConfig(0), func(int, int) {})
	if !errors.As(err, &NothingToSimulateError{}) {
		t.Fatalf("err = %v", err)
	}
}

func TestBehaviourOf(t *testing.T) {
	// A short history keeps the defaults.
	short := [][]HistoryReview{{{Time: day(1, 10), Rating: Good, StateBefore: New, Duration: time.Second}}}
	if b := BehaviourOf(short, 4); b != DefaultBehaviour() {
		t.Fatalf("short history: %+v", b)
	}
	// 200 cards: half start with Again (relearned after a minute, 10 s
	// each answer), half with Good; their reviews are recalled Good in 4 s,
	// or forgotten in 8 s with one 6 s relearning answer.
	var hist [][]HistoryReview
	for i := range 200 {
		var h []HistoryReview
		if i%2 == 0 {
			h = append(h, HistoryReview{Time: day(1, 10), Rating: Again, StateBefore: New, Duration: 10 * time.Second},
				HistoryReview{Time: day(1, 11), Rating: Good, StateBefore: Learning, Duration: 10 * time.Second})
		} else {
			h = append(h, HistoryReview{Time: day(1, 10), Rating: Good, StateBefore: New, Duration: 5 * time.Second})
		}
		h = append(h, HistoryReview{Time: day(3, 10), Rating: Good, StateBefore: Review, Duration: 4 * time.Second})
		if i%4 == 0 {
			h = append(h, HistoryReview{Time: day(9, 10), Rating: Again, StateBefore: Review, Duration: 8 * time.Second},
				HistoryReview{Time: day(9, 10).Add(10 * time.Minute), Rating: Good, StateBefore: Relearning, Duration: 6 * time.Second})
		}
		// A preview-like manual entry is ignored.
		h = append(h, HistoryReview{Time: day(10, 10), StateBefore: Review, IvlAfter: 5})
		hist = append(hist, h)
	}
	b := BehaviourOf(hist, 4)
	if b.FirstRating != [4]float64{0.5, 0, 0.5, 0} {
		t.Errorf("first ratings = %v", b.FirstRating)
	}
	if b.LearnSeconds[0] != 20 || b.LearnSeconds[2] != 5 || b.LearnSeconds[1] != DefaultBehaviour().LearnSeconds[1] {
		t.Errorf("learn seconds = %v", b.LearnSeconds)
	}
	if b.RecallRating != [3]float64{0, 1, 0} {
		t.Errorf("recall ratings = %v", b.RecallRating)
	}
	if b.ReviewSeconds[0] != 14 || b.ReviewSeconds[2] != 4 {
		t.Errorf("review seconds = %v", b.ReviewSeconds)
	}
}

func TestSimCardOf(t *testing.T) {
	now := day(20, 12)
	if _, ok := SimCardOf(CardState{}, now, 4, nil); ok {
		t.Fatal("a New card is not studied")
	}
	if _, ok := SimCardOf(CardState{State: Learning, Due: now}, now, 4, nil); ok {
		t.Fatal("a learning card without memory counts as new")
	}
	rev := CardState{State: Review, Memory: &Memory{Stability: 12, Difficulty: 4}, Due: day(25, 4),
		LastReview: day(13, 9), ScheduledDays: 12}
	c, ok := SimCardOf(rev, now, 4, nil)
	if !ok || c != (SimCard{Stability: 12, Difficulty: 4, LastDay: -7, DueDay: 5}) {
		t.Fatalf("review card = %+v", c)
	}
	learning := CardState{State: Learning, Memory: &Memory{Stability: 0.5, Difficulty: 6}, Due: now.Add(time.Hour),
		LastReview: now}
	if c, ok := SimCardOf(learning, now, 4, nil); !ok || c.DueDay != 0 || c.LastDay != 0 {
		t.Fatalf("learning card = %+v", c)
	}
	// Without a memory state, it is estimated from the interval.
	rev.Memory, rev.LastReview = nil, time.Time{}
	if c, ok := SimCardOf(rev, now, 4, nil); !ok || c.Stability <= 0 || c.LastDay != -7 {
		t.Fatalf("review card without memory = %+v", c)
	}
}
