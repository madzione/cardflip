package stats

import (
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
)

// testData returns empty Data at the fixture's clock.
func testData() *Data {
	return &Data{Now: fixtureNow, Today: day(fixtureNow), Rollover: core.DefaultRolloverHour, Retention: 0.9}
}

// answer adds n answers given daysAgo study days ago.
func (d *Data) answer(n, daysAgo int, kind Kind, rating scheduler.Rating, ivl int, ms int64) {
	for range n {
		d.Reviews = append(d.Reviews, Review{CardID: core.ID(len(d.Reviews) + 1), Day: d.Today - daysAgo, Hour: 10,
			Rating: rating, Kind: kind, IvlBefore: int32(ivl), DurationMs: ms})
	}
}

// review adds n review cards with interval ivl due inDays study days from
// today.
func (d *Data) review(n, inDays, ivl int) {
	for range n {
		d.Cards = append(d.Cards, core.Card{ID: core.ID(len(d.Cards) + 1), CardState: scheduler.CardState{
			State: scheduler.Review, Due: fixtureNow.AddDate(0, 0, inDays), ScheduledDays: ivl}})
	}
}

func TestTodayAfterASession(t *testing.T) {
	// After a session (spec): 40 cards today, 4 of them Again, in 10
	// minutes.
	d := testData()
	d.answer(30, 0, Young, scheduler.Good, 5, 15000)
	d.answer(4, 0, Mature, scheduler.Again, 40, 15000)
	d.answer(6, 0, Learning, scheduler.Good, 0, 15000)
	d.answer(7, 1, Young, scheduler.Again, 5, 15000) // yesterday
	got := TodaySummary(d)
	want := TodayStats{Cards: 40, Time: 10 * time.Minute, Again: 4, Correct: 0.9, Learning: 6, Review: 34}
	if got != want {
		t.Fatalf("today = %+v\nwant    %+v", got, want)
	}
	if empty := TodaySummary(testData()); empty != (TodayStats{}) {
		t.Fatalf("empty = %+v", empty)
	}
}

func TestTodayFixture(t *testing.T) {
	f := newFixture(t)
	got := TodaySummary(f.load(t, Scope{Kind: ScopeCollection}))
	// V1's Again and relearning Good, F1's cram review; S1's 02:30 answer
	// was yesterday's.
	want := TodayStats{Cards: 3, Time: 12 * time.Second, Again: 1, Correct: 2.0 / 3, Review: 1, Relearning: 1, Filtered: 1}
	if got != want {
		t.Fatalf("today = %+v\nwant    %+v", got, want)
	}
}

func TestFutureDueForecast(t *testing.T) {
	// Forecast (spec): 30 review cards due tomorrow, 12 the day after.
	d := testData()
	d.review(30, 1, 10)
	d.review(12, 2, 10)
	d.review(5, -3, 10)
	d.review(1, 45, 10)                           // beyond a month
	d.Cards = append(d.Cards, core.Card{ID: 999}) // new: not forecast
	got := FutureDue(d, PeriodMonth)
	if len(got.Days) != 30 || got.Days[1] != 30 || got.Days[2] != 12 || got.Days[0] != 0 {
		t.Fatalf("days = %v", got.Days)
	}
	if got.Overdue != 5 || got.Total != 42 || got.DueTomorrow != 30 || got.Cumulative[2] != 42 || got.Cumulative[29] != 42 {
		t.Fatalf("forecast = %+v", got)
	}
	if got.Average != 42.0/30 {
		t.Fatalf("average = %v", got.Average)
	}
	all := FutureDue(d, PeriodAll)
	if len(all.Days) != 46 || all.Days[45] != 1 || all.Total != 43 {
		t.Fatalf("all history: %d days, total %d", len(all.Days), all.Total)
	}
}

func TestFutureDueFixture(t *testing.T) {
	f := newFixture(t)
	got := FutureDue(f.load(t, Scope{Kind: ScopeCollection}), PeriodMonth)
	// S1 is due at 02:30 tomorrow, which is still today's study day; V1
	// later today; F1 is suspended.
	if got.Days[0] != 2 || got.Total != 2 || got.Overdue != 0 {
		t.Fatalf("forecast = %+v", got)
	}
}

func TestCardCounts(t *testing.T) {
	f := newFixture(t)
	got := CardCounts(f.load(t, Scope{Kind: ScopeCollection}))
	want := StateCounts{New: 1, Relearning: 1, Mature: 1, Suspended: 1, Total: 4}
	if got != want {
		t.Fatalf("counts = %+v\nwant     %+v", got, want)
	}
	d := testData()
	d.review(3, 1, 20)
	d.review(2, 1, 21)
	d.Cards = append(d.Cards,
		core.Card{ID: 10, BuriedUntil: fixtureNow.Add(time.Hour)},
		core.Card{ID: 11, BuriedUntil: fixtureNow.Add(-time.Hour)},
		core.Card{ID: 12, CardState: scheduler.CardState{State: scheduler.Learning}})
	want = StateCounts{New: 1, Learning: 1, Young: 3, Mature: 2, Buried: 1, Total: 8}
	if got := CardCounts(d); got != want {
		t.Fatalf("counts = %+v\nwant     %+v", got, want)
	}
}
