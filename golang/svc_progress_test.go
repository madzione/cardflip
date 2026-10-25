package main

import (
	"fmt"
	"reflect"
	"slices"
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
	if p := must(s.progress.Today()); !reflect.DeepEqual(p, Progress{}) {
		t.Fatalf("no reviews: %+v", p)
	}
	// Goal in progress (spec, card goal): 2 of 3 cards.
	check(t, s.app.SetDailyGoal(core.GoalCards, 3))
	studyNew(t, s, 2, 1000)
	want := Progress{Current: 1, TodayStudied: true, Longest: 1, TodayCards: 2, GoalKind: core.GoalCards, Goal: 3}
	if p := must(s.progress.Today()); !reflect.DeepEqual(p, want) {
		t.Fatalf("goal in progress: %+v\nwant              %+v", p, want)
	}
	// Goal reached.
	studyNew(t, s, 1, 1000)
	want.TodayCards, want.GoalReached = 3, true
	if p := must(s.progress.Today()); !reflect.DeepEqual(p, want) {
		t.Fatalf("goal reached: %+v\nwant          %+v", p, want)
	}
	// At risk: the next day, before studying.
	s.advance(24 * time.Hour)
	want = Progress{Current: 1, AtRisk: true, Longest: 1, GoalKind: core.GoalCards, Goal: 3}
	if p := must(s.progress.Today()); !reflect.DeepEqual(p, want) {
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
	today := scheduler.DayNumber(*s.clock, core.DefaultRolloverHour)
	checkDays(t, s, "rollover 4", []int{today - 2}, []int{2}, []int64{2000})
	check(t, s.app.SetRolloverHour(0))
	if p := must(s.progress.Today()); p.Current != 2 || !p.AtRisk || p.Longest != 2 {
		t.Fatalf("rollover 0: %+v", p)
	}
	checkDays(t, s, "rollover 0", []int{today - 2, today - 1}, []int{1, 1}, []int64{1000, 1000})
	// The cached days follow new answers without rereading the log.
	gen := s.progress.gen
	studyNew(t, s, 2, 1500)
	if p := must(s.progress.Today()); p.Current != 3 || !p.TodayStudied || s.progress.gen != gen {
		t.Fatalf("after an answer: %+v, generation %d → %d", p, gen, s.progress.gen)
	}
	checkDays(t, s, "after answers", []int{today - 2, today - 1, today}, []int{1, 1, 2}, []int64{1000, 1000, 3000})
}

// checkDays fails unless the progress service holds these study days with
// these answers and milliseconds.
func checkDays(t *testing.T, s *services, what string, days, answers []int, ms []int64) {
	t.Helper()
	p := s.progress
	if !slices.Equal(p.days, days) || !slices.Equal(p.answers, answers) || !slices.Equal(p.ms, ms) {
		t.Fatalf("%s: days %v, answers %v, ms %v\nwant %v, %v, %v", what, p.days, p.answers, p.ms, days, answers, ms)
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

func TestProgressRestDays(t *testing.T) {
	s, _ := opened(t)
	// Studied on each of the last 11 days but 4 days ago.
	importHistory(t, s, "rest", 10, 1000, 10, 9, 8, 7, 6, 5, 3, 2, 1, 0)
	if p := must(s.progress.Today()); p.Current != 4 || p.Longest != 6 {
		t.Fatalf("no rest days: %+v", p)
	}
	// Applied to history (spec): with one rest day the whole run counts.
	check(t, s.app.SetRestDaysPerWeek(1))
	if p := must(s.progress.Today()); p.Current != 10 || p.Longest != 10 {
		t.Fatalf("one rest day: %+v", p)
	}
	sec := must(s.stats.Streaks(stats.Scope{Kind: stats.ScopeCollection}, stats.PeriodAll))
	if sec.Current != 10 || sec.Longest != 10 {
		t.Fatalf("statistics with one rest day: %+v", sec)
	}
}

// achievementOf returns the status of achievement id.
func achievementOf(t *testing.T, s *services, id string) stats.AchievementStatus {
	t.Helper()
	for _, a := range must(s.progress.Achievements()) {
		if a.ID == id {
			return a
		}
	}
	t.Fatalf("no achievement %s", id)
	return stats.AchievementStatus{}
}

// newKeys returns the keys of the new achievements.
func newKeys(p Progress) []string {
	var out []string
	for _, a := range p.NewAchievements {
		out = append(out, a.Key)
	}
	return out
}

func TestProgressAchievements(t *testing.T) {
	s, _ := opened(t)
	c := must(s.app.a.collection())
	// The first evaluation of a new collection has nothing to mark seen.
	if p := must(s.progress.Today()); len(p.NewAchievements) != 0 {
		t.Fatalf("new collection: %+v", p.NewAchievements)
	}
	if seen := c.Settings().SeenAchievements; seen == nil || len(seen) != 0 {
		t.Fatalf("seen after the first evaluation: %#v", seen)
	}
	// An import of 99 days of history up to today earns three tiers at
	// once, with their dates in the history.
	days := make([]int, 99)
	for i := range days {
		days[i] = 98 - i
	}
	importHistory(t, s, "history", 10, 1000, days...)
	want := []string{"streak:7", "streak:30", "study-days:30"}
	if got := newKeys(must(s.progress.Today())); !slices.Equal(got, want) {
		t.Fatalf("after the import: %v\nwant %v", got, want)
	}
	date := func(daysAgo int) string {
		return stats.Date(scheduler.DayNumber(s.clock.AddDate(0, 0, -daysAgo), core.DefaultRolloverHour))
	}
	if a := achievementOf(t, s, "study-days"); a.Progress != 99 || !slices.Equal(a.Earned, []string{date(69)}) {
		t.Fatalf("study days: %+v", a)
	}
	// Setting a goal earns the Daily goal tiers from the history: five new
	// tiers at once, more than three.
	if a := achievementOf(t, s, "daily-goal"); !a.NeedsGoal || a.Progress != 0 {
		t.Fatalf("daily goal without a goal: %+v", a)
	}
	check(t, s.app.SetDailyGoal(core.GoalCards, 1))
	want = append(want, "daily-goal:7", "daily-goal:30")
	if got := newKeys(must(s.progress.Today())); !slices.Equal(got, want) {
		t.Fatalf("with a goal: %v\nwant %v", got, want)
	}
	// Seen: marked tiers are not new again.
	check(t, s.progress.MarkSeen(want))
	if p := must(s.progress.Today()); len(p.NewAchievements) != 0 {
		t.Fatalf("after marking seen: %+v", p.NewAchievements)
	}
	// Today's 100th answer is new; undo (spec) locks the tier again with
	// 99 / 100.
	studyNew(t, s, 1, 1000)
	p := must(s.progress.Today())
	if got := newKeys(p); !slices.Equal(got, []string{"answers:100"}) {
		t.Fatalf("100th answer: %v", got)
	}
	if a := p.NewAchievements[0]; a.ID != "answers" || a.Threshold != 100 || a.TierName != stats.Bronze {
		t.Fatalf("100th answer: %+v", a)
	}
	must(s.study.Undo())
	if a := achievementOf(t, s, "answers"); a.Progress != 99 || len(a.Earned) != 0 {
		t.Fatalf("answers after undo: %+v", a)
	}
	if got := newKeys(must(s.progress.Today())); len(got) != 0 {
		t.Fatalf("new after undo: %v", got)
	}
}

func TestProgressAchievementsAfterUpgrade(t *testing.T) {
	s, _ := opened(t)
	c := must(s.app.a.collection())
	days := make([]int, 40)
	for i := range days {
		days[i] = i
	}
	importHistory(t, s, "old", 10, 1000, days...)
	// A collection from before achievements: nothing seen yet.
	st := c.Settings()
	st.SeenAchievements = nil
	check(t, c.UpdateSettings(st))
	s.progress = &ProgressService{a: s.app.a}
	if p := must(s.progress.Today()); len(p.NewAchievements) != 0 {
		t.Fatalf("first evaluation after an upgrade: %+v", p.NewAchievements)
	}
	if seen := c.Settings().SeenAchievements; !slices.Equal(seen, []string{"streak:7", "streak:30", "study-days:30"}) {
		t.Fatalf("seen: %v", seen)
	}
	// The achievements screen still shows them with their dates.
	if a := achievementOf(t, s, "streak"); a.Progress != 40 || len(a.Earned) != 2 {
		t.Fatalf("streak: %+v", a)
	}
	// The seen list stays on reload, also when empty.
	file := must(s.transfer.ExportBackup())
	other, _ := opened(t)
	check(t, other.transfer.Restore(must(other.transfer.StageBytes(file.Name, file.Data)).ID))
	if seen := must(other.app.a.collection()).Settings().SeenAchievements; len(seen) != 3 {
		t.Fatalf("seen after restore: %v", seen)
	}
}

func TestProgressMatureAfterAnswer(t *testing.T) {
	s, _ := opened(t)
	c := must(s.app.a.collection())
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	// A review card due today with an interval of 15 days.
	b := &core.ImportBatch{
		NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []core.ImportNote{{TypeKey: "basic", GUID: "due", DeckName: "Default", Fields: []string{"due", "a"},
			Cards: []core.ImportCard{{DeckName: "Default",
				State: scheduler.CardState{State: scheduler.Review, Due: s.clock.Add(-time.Hour), ScheduledDays: 15, Reps: 3,
					LastReview: s.clock.AddDate(0, 0, -15), Memory: &scheduler.Memory{Stability: 15, Difficulty: 3}},
				History: []core.ImportReview{{Time: s.clock.AddDate(0, 0, -15), Rating: scheduler.Good,
					StateBefore: scheduler.Review, IvlAfter: 15, DurationMs: 1000}}}}}},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	if a := achievementOf(t, s, "long-term"); a.Progress != 0 {
		t.Fatalf("before: %+v", a)
	}
	gen := s.progress.gen
	st := must(s.study.Start(must(s.notes.Defaults(0)).DeckID))
	must(s.study.Answer(st.Card.CardID, int(scheduler.Good), 1000))
	if a := achievementOf(t, s, "long-term"); a.Progress != 1 || s.progress.gen != gen {
		t.Fatalf("after the answer: %+v, generation %d → %d", a, gen, s.progress.gen)
	}
}

func TestProgressAchievementsImportedHistory(t *testing.T) {
	s, _ := opened(t)
	c := must(s.app.a.collection())
	must(s.progress.Today()) // the first evaluation, before the import
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	// Imported history (spec): an Anki collection whose 1,000th answer was
	// on 2024-03-12. 14 answers a day from January 1st to March 11th (994),
	// 6 on March 12th and 20 on March 13th, over 20 cards.
	cards := make([][]core.ImportReview, 20)
	for day := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local); day.Before(time.Date(2024, 3, 14, 0, 0, 0, 0, time.Local)); day = day.AddDate(0, 0, 1) {
		n := 14
		switch day.Format(time.DateOnly) {
		case "2024-03-12":
			n = 6
		case "2024-03-13":
			n = 20
		}
		for k := range n {
			cards[k] = append(cards[k], core.ImportReview{Time: day.Add(time.Duration(k) * time.Minute), Rating: scheduler.Good,
				StateBefore: scheduler.Review, IvlBefore: 1, IvlAfter: 1, DurationMs: 4000,
				Source: fmt.Sprintf("anki:%d-%d", day.Unix(), k)})
		}
	}
	b := &core.ImportBatch{NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}}}
	for k, history := range cards {
		b.Notes = append(b.Notes, core.ImportNote{TypeKey: "basic", GUID: fmt.Sprint("anki", k), DeckName: "Default",
			Fields: []string{fmt.Sprint("q", k), "a"}, Cards: []core.ImportCard{{DeckName: "Default", History: history,
				State: scheduler.CardState{State: scheduler.Review, Due: s.clock.AddDate(0, 0, 1), ScheduledDays: 1}}}})
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	if a := achievementOf(t, s, "answers"); a.Progress != 1_020 || !slices.Equal(a.Earned, []string{"2024-01-08", "2024-03-12"}) {
		t.Fatalf("answers: %+v", a)
	}
	if a := achievementOf(t, s, "study-days"); a.Progress != 73 || !slices.Equal(a.Earned, []string{"2024-01-30"}) {
		t.Fatalf("study days: %+v", a)
	}
	// Import earns many (spec): all tiers are new at once.
	want := []string{"answers:100", "answers:1000", "streak:7", "streak:30", "study-days:30"}
	if got := newKeys(must(s.progress.Today())); !slices.Equal(got, want) {
		t.Fatalf("new: %v\nwant %v", got, want)
	}
}
