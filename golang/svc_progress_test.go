package main

import (
	"fmt"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/stats"
)

// studyNew adds n basic notes to the default deck and answers each new
// card Good, taking ms milliseconds.
func studyNew(t *testing.T, s *services, n int, ms int64) {
	t.Helper()
	for i := range n {
		basicNote(t, s, fmt.Sprintf("q%d-%d", s.clock.Unix(), i), "a")
	}
	st := must(s.study.Start(must(s.notes.Defaults(0)).DeckID))
	for range n {
		if st.Card == nil {
			t.Fatalf("no card to answer: %+v", st)
		}
		st = must(s.study.Answer(st.Card.CardID, int(scheduler.Good), ms))
	}
}

// importHistory imports one note whose card has an Anki answer at hour on
// each of the study days daysAgo (0 is today), taking ms milliseconds.
func importHistory(t *testing.T, s *services, guid string, hour int, ms int64, daysAgo ...int) {
	t.Helper()
	c := must(s.app.a.collection())
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	var history []core.ImportReview
	for _, back := range daysAgo {
		y, m, d := s.clock.AddDate(0, 0, -back).Date()
		history = append(history, core.ImportReview{Time: time.Date(y, m, d, hour, 0, 0, 0, time.Local),
			Rating: scheduler.Good, StateBefore: scheduler.Review, IvlBefore: 1, IvlAfter: 1, DurationMs: ms,
			Source: fmt.Sprintf("anki:%s-%d", guid, back)})
	}
	b := &core.ImportBatch{
		NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []core.ImportNote{{TypeKey: "basic", GUID: guid, DeckName: "Default", Fields: []string{guid, "a"},
			Cards: []core.ImportCard{{DeckName: "Default", History: history,
				State: scheduler.CardState{State: scheduler.Review, Due: s.clock.AddDate(0, 0, 1), ScheduledDays: 1}}}}},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
}

func (s *services) advance(d time.Duration) { *s.clock = s.clock.Add(d) }

func TestProgressToday(t *testing.T) {
	s, _ := opened(t)
	// No reviews: no streak and nothing at risk.
	if p := must(s.progress.Today()); p != (Progress{}) {
		t.Fatalf("no reviews: %+v", p)
	}
	// Goal in progress (spec, card goal): 2 of 3 cards.
	check(t, s.app.SetDailyGoal(core.GoalCards, 3))
	studyNew(t, s, 2, 1000)
	want := Progress{Current: 1, TodayStudied: true, Longest: 1, TodayCards: 2, GoalKind: core.GoalCards, Goal: 3}
	if p := must(s.progress.Today()); p != want {
		t.Fatalf("goal in progress: %+v\nwant              %+v", p, want)
	}
	// Goal reached.
	studyNew(t, s, 1, 1000)
	want.TodayCards, want.GoalReached = 3, true
	if p := must(s.progress.Today()); p != want {
		t.Fatalf("goal reached: %+v\nwant          %+v", p, want)
	}
	// At risk: the next day, before studying.
	s.advance(24 * time.Hour)
	want = Progress{Current: 1, AtRisk: true, Longest: 1, GoalKind: core.GoalCards, Goal: 3}
	if p := must(s.progress.Today()); p != want {
		t.Fatalf("at risk: %+v\nwant     %+v", p, want)
	}
	// Undo the only answer of the day (spec): the streak goes back to at
	// risk; redo brings it back.
	studyNew(t, s, 1, 1000)
	if p := must(s.progress.Today()); p.Current != 2 || !p.TodayStudied || p.TodayCards != 1 {
		t.Fatalf("second day: %+v", p)
	}
	must(s.study.Undo())
	if p := must(s.progress.Today()); p.Current != 1 || !p.AtRisk || p.TodayStudied || p.TodayCards != 0 || p.Longest != 1 {
		t.Fatalf("after undo: %+v", p)
	}
	must(s.app.Redo())
	if p := must(s.progress.Today()); p.Current != 2 || p.TodayCards != 1 || p.Longest != 2 {
		t.Fatalf("after redo: %+v", p)
	}
	// Minute goal (spec): 16 minutes of 15.
	check(t, s.app.SetDailyGoal(core.GoalMinutes, 15))
	importHistory(t, s, "long", 10, 16*60_000, 0)
	if p := must(s.progress.Today()); p.TodayMinutes != 16 || !p.GoalReached || p.GoalKind != core.GoalMinutes || p.Goal != 15 {
		t.Fatalf("minute goal: %+v", p)
	}
}

func TestProgressImportedHistory(t *testing.T) {
	s, _ := opened(t)
	// Imported Anki history (spec): reviews on each of the last 50 days.
	days := make([]int, 50)
	for i := range days {
		days[i] = i
	}
	must(s.progress.Today()) // reads the study days before the import
	importHistory(t, s, "anki", 10, 5000, days...)
	if p := must(s.progress.Today()); p.Current != 50 || !p.TodayStudied || p.Longest != 50 || p.TodayCards != 1 {
		t.Fatalf("imported history: %+v", p)
	}
}

func TestProgressRolloverChange(t *testing.T) {
	s, _ := opened(t)
	// Answers on Oct 11 at 12:00 and Oct 12 at 03:00. With the rollover at
	// 04:00 both are Oct 11, two days ago: no streak. With the rollover at
	// midnight they are Oct 11 and 12: a two-day streak at risk.
	importHistory(t, s, "noon", 12, 1000, 2)
	importHistory(t, s, "night", 3, 1000, 1)
	if p := must(s.progress.Today()); p.Current != 0 || p.AtRisk || p.Longest != 1 {
		t.Fatalf("rollover 4: %+v", p)
	}
	check(t, s.app.SetRolloverHour(0))
	if p := must(s.progress.Today()); p.Current != 2 || !p.AtRisk || p.Longest != 2 {
		t.Fatalf("rollover 0: %+v", p)
	}
	// The cached days follow new answers without rereading the log.
	gen := s.progress.gen
	studyNew(t, s, 1, 1000)
	if p := must(s.progress.Today()); p.Current != 3 || !p.TodayStudied || s.progress.gen != gen {
		t.Fatalf("after an answer: %+v, generation %d → %d", p, gen, s.progress.gen)
	}
}

func TestProgressAgreesWithStatistics(t *testing.T) {
	s, _ := opened(t)
	// The same history through both paths: a 6-day run long ago, a gap,
	// and 4 days up to yesterday; then today's answers.
	importHistory(t, s, "old", 10, 1000, 30, 29, 28, 27, 26, 25)
	importHistory(t, s, "recent", 9, 1000, 4, 3, 2, 1)
	studyNew(t, s, 2, 1000)
	for _, period := range []string{stats.PeriodMonth, stats.PeriodAll} {
		sec := must(s.stats.Streaks(stats.Scope{Kind: stats.ScopeCollection}, period))
		p := must(s.progress.Today())
		if sec.Current != p.Current || sec.Longest != p.Longest || sec.AtRisk != p.AtRisk || sec.TodayStudied != p.TodayStudied {
			t.Fatalf("%s: statistics %+v, progress %+v", period, sec, p)
		}
		if p.Current != 5 || p.Longest != 6 {
			t.Fatalf("progress %+v", p)
		}
	}
}

func TestProgressIgnoresManualEntries(t *testing.T) {
	s, _ := opened(t)
	// Set due date does not count (spec): the only activity today is
	// setting the due date of 10 cards, then resetting them.
	var ids []int64
	for i := range 10 {
		basicNote(t, s, fmt.Sprintf("q%d", i), "a")
	}
	for _, card := range must(s.app.a.collection()).AllCards() {
		ids = append(ids, int64(card.ID))
	}
	must(s.progress.Today())
	gen := s.progress.gen
	check(t, s.browse.SetDueDate(ids, "3"))
	check(t, s.browse.ResetToNew(ids, ResetOptions{}))
	if p := must(s.progress.Today()); p.Current != 0 || p.TodayStudied || p.TodayCards != 0 || p.Longest != 0 {
		t.Fatalf("after manual entries: %+v", p)
	}
	if s.progress.gen != gen {
		t.Fatalf("manual entries reread the log: generation %d → %d", gen, s.progress.gen)
	}
}
