package stats

import (
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
)

func TestStreakSummaryFixture(t *testing.T) {
	f := newFixture(t)
	// Collection: study days Sep 13, Oct 1 (the deleted card), Oct 12 (F1
	// and S1's 02:30 answer) and Oct 13. The manual entry of Sep 20 is no
	// study day.
	d := f.load(t, Scope{Kind: ScopeCollection})
	got := StreakSummary(d, Goal{}, PeriodMonth)
	want := StreakSection{Current: 2, TodayStudied: true, Longest: 2, LongestFirst: "2026-10-12", LongestLast: "2026-10-13",
		StudyDays: 3, PeriodDays: 30}
	if got != want {
		t.Fatalf("collection = %+v\nwant         %+v", got, want)
	}
	if all := StreakSummary(d, Goal{}, PeriodAll); all.StudyDays != 4 || all.PeriodDays != day(fixtureNow)-day(at(9, 13, 12, 0))+1 {
		t.Fatalf("all = %+v", all)
	}
	// Goal days: Oct 13 had 3 answers, Oct 12 two, Oct 1 one; none took a
	// minute.
	for _, tc := range []struct {
		goal Goal
		want int
	}{{Goal{core.GoalCards, 2}, 2}, {Goal{core.GoalCards, 3}, 1}, {Goal{core.GoalMinutes, 1}, 0}} {
		if got := StreakSummary(d, tc.goal, PeriodMonth).GoalDays; got == nil || *got != tc.want {
			t.Errorf("%+v: goal days %v, want %d", tc.goal, got, tc.want)
		}
	}
	// Deck scope: Spanish::Verbs was only studied today.
	verbs := StreakSummary(f.load(t, Scope{Kind: ScopeDeck, DeckID: int64(f.verbs)}), Goal{}, PeriodMonth)
	if verbs.Current != 1 || verbs.Longest != 1 || verbs.StudyDays != 1 || verbs.LongestFirst != "2026-10-13" {
		t.Fatalf("Spanish::Verbs = %+v", verbs)
	}
	// Previews do not count: N1 previewed today in a filtered deck that
	// does not reschedule.
	cfg := core.DefaultFilterConfig("deck:French is:new")
	cfg.Reschedule = false
	if _, _, err := f.c.CreateFilteredDeck("Preview", cfg); err != nil {
		t.Fatal(err)
	}
	if e, err := f.c.Answer(f.n1, scheduler.Good, time.Second); err != nil || !e.Preview {
		t.Fatalf("preview %+v, %v", e, err)
	}
	n1 := StreakSummary(f.load(t, Scope{Kind: ScopeSearch, Query: "n1"}), Goal{}, PeriodMonth)
	if n1 != (StreakSection{PeriodDays: 30}) {
		t.Fatalf("previewed card = %+v", n1)
	}
}

func TestStreakSummaryGoalDays(t *testing.T) {
	// Collection streaks (spec): 24 of the last 30 days, the last 9 in a
	// row; the longest run, 41 days, ended 60 days ago.
	d := testData()
	for back := range 9 {
		d.answer(30, back, Young, scheduler.Good, 5, 1000)
	}
	for back := 10; back < 25; back++ {
		d.answer(10, back, Young, scheduler.Good, 5, 1000)
	}
	for back := 60; back < 101; back++ {
		d.answer(1, back, Young, scheduler.Good, 5, 1000)
	}
	got := StreakSummary(d, Goal{core.GoalCards, 30}, PeriodMonth)
	if got.Current != 9 || got.Longest != 41 || got.LongestLast != Date(d.Today-60) || got.LongestFirst != Date(d.Today-100) ||
		got.StudyDays != 24 || got.PeriodDays != 30 {
		t.Fatalf("section = %+v", got)
	}
	// Goal days: the 9 days with 30 cards. Goal raised (spec): with 50
	// cards, none.
	if got.GoalDays == nil || *got.GoalDays != 9 {
		t.Fatalf("goal days %v", got.GoalDays)
	}
	if g := StreakSummary(d, Goal{core.GoalCards, 50}, PeriodMonth).GoalDays; *g != 0 {
		t.Fatalf("raised goal: %d days", *g)
	}
	// Minutes: 30 answers of a second each are half a minute.
	if g := StreakSummary(d, Goal{core.GoalMinutes, 1}, PeriodMonth).GoalDays; *g != 0 {
		t.Fatalf("minute goal: %d days", *g)
	}
	// No goal (spec): no goal days.
	if g := StreakSummary(d, Goal{}, PeriodMonth).GoalDays; g != nil {
		t.Fatalf("no goal: %d days", *g)
	}
}

func TestGoalMet(t *testing.T) {
	// Card goal and minute goal (spec).
	if (Goal{core.GoalCards, 30}).Met(18, 0) || !(Goal{core.GoalCards, 30}).Met(30, 0) {
		t.Fatal("card goal")
	}
	if !(Goal{core.GoalMinutes, 15}).Met(1, 16*60_000) || (Goal{core.GoalMinutes, 15}).Met(100, 14*60_000) {
		t.Fatal("minute goal")
	}
	if (Goal{}).Met(1000, 1e9) {
		t.Fatal("no goal met")
	}
}
