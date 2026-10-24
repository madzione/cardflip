package stats

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// largeCollection returns a collection of 50,000 cards in ten decks with
// 500,000 review log entries over two years.
func largeCollection(tb testing.TB) *core.Collection {
	tb.Helper()
	const notes = 50_000
	clk := fixtureNow
	c, _, err := core.Open(storage.NewMemory(), core.Options{Now: func() time.Time { return clk }})
	if err != nil {
		tb.Fatal(err)
	}
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	rng := rand.New(rand.NewPCG(1, 2))
	b := &core.ImportBatch{NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}}}
	reviewCards := 0
	for i := range notes {
		deck := fmt.Sprintf("Bench::Deck %d", i%10)
		ic := core.ImportCard{DeckName: deck}
		if i%10 != 0 {
			// 45,000 review cards: 11 entries each, 12 for the first
			// 5,000, so 500,000 in total.
			perCard := 11
			if reviewCards < 5000 {
				perCard = 12
			}
			reviewCards++
			last := fixtureNow.Add(-time.Duration(rng.IntN(90*24)) * time.Hour)
			ivl := 1 + rng.IntN(400)
			ic.State = scheduler.CardState{State: scheduler.Review, Due: last.AddDate(0, 0, ivl), ScheduledDays: ivl,
				Memory: &scheduler.Memory{Stability: float64(ivl) * (0.5 + rng.Float64()), Difficulty: 1 + 9*rng.Float64()},
				Reps:   perCard, LastReview: last}
			for j := range perCard {
				state := scheduler.Review
				if j < 2 {
					state = scheduler.Learning
				}
				ic.History = append(ic.History, core.ImportReview{
					Time:   fixtureNow.Add(-time.Duration(rng.IntN(730*24*60)) * time.Minute),
					Rating: scheduler.Rating(1 + rng.IntN(4)), StateBefore: state,
					IvlBefore: rng.IntN(100), IvlAfter: 1 + rng.IntN(100), DurationMs: int64(rng.IntN(20000)),
				})
			}
		}
		b.Notes = append(b.Notes, core.ImportNote{TypeKey: "basic", GUID: fmt.Sprint(i), DeckName: deck,
			Fields: []string{fmt.Sprint("q", i), "a"}, Cards: []core.ImportCard{ic}})
	}
	if _, err := c.Import(b, nil); err != nil {
		tb.Fatal(err)
	}
	return c
}

// allSections computes every section of scope, as the screen does.
func allSections(tb testing.TB, c *core.Collection, scope Scope, period string) {
	var l Loader
	d, err := l.Load(c, scope)
	if err != nil {
		tb.Fatal(err)
	}
	TodaySummary(d)
	FutureDue(d, period)
	CardCounts(d)
	ReviewCalendar(d)
	Reviews(d, period)
	ReviewTime(d, period)
	Hourly(d, period)
	AnswerButtons(d, period)
	Intervals(d)
	Stability(d)
	Difficulty(d)
	Retrievability(d)
	TrueRetention(d)
}

// BenchmarkAllSections loads the review log afresh and computes every
// section for 50,000 cards and 500,000 entries. It must stay under 2
// seconds per operation.
func BenchmarkAllSections(b *testing.B) {
	c := largeCollection(b)
	months, _ := c.RevlogMonths()
	n := 0
	for _, m := range months {
		e, _ := c.ReadRevlog(m)
		n += len(e)
	}
	if n != 500_000 || len(c.AllCards()) != 50_000 {
		b.Fatalf("%d entries, %d cards", n, len(c.AllCards()))
	}
	for b.Loop() {
		allSections(b, c, Scope{Kind: ScopeCollection}, PeriodAll)
	}
	if per := b.Elapsed() / time.Duration(b.N); per > 2*time.Second {
		b.Errorf("all sections took %v, want under 2s", per)
	}
}
