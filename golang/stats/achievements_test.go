package stats

import (
	"slices"
	"testing"

	"cardflip/golang/core"
)

// daily returns a history with an answer count and milliseconds per day
// (one entry per day from first on, with days of 0 answers left out).
func daily(first int, answers []int, ms []int64) History {
	var h History
	for i, n := range answers {
		if n == 0 {
			continue
		}
		h.Days = append(h.Days, first+i)
		h.Answers = append(h.Answers, n)
		if ms != nil {
			h.Ms = append(h.Ms, ms[i])
		} else {
			h.Ms = append(h.Ms, 0)
		}
	}
	return h
}

// repeat returns n copies of v.
func repeat[T any](v T, n int) []T {
	out := make([]T, n)
	for i := range out {
		out[i] = v
	}
	return out
}

func achievement(t *testing.T, h History, id string) AchievementStatus {
	t.Helper()
	all := EvalAchievements(h)
	i := slices.IndexFunc(all, func(a AchievementStatus) bool { return a.ID == id })
	if i < 0 {
		t.Fatalf("no achievement %s", id)
	}
	return all[i]
}

const (
	minute = int64(60_000)
	hour   = 60 * minute
)

func TestAchievements(t *testing.T) {
	const d = 20000
	streakDays := daily(d, repeat(1, 8), nil)
	withRest := daily(d, []int{1, 1, 1, 0, 1, 1, 1, 1}, nil)
	withRest.Rest = 1
	cardsGoal := daily(d, []int{20, 5, 30, 19, 25, 20, 40, 21, 20}, nil)
	cardsGoal.Goal = Goal{Kind: core.GoalCards, Value: 20}
	minutesGoal := daily(d, repeat(1, 7), []int64{600_000, 599_999, 600_000, 600_000, 600_000, 600_000, 900_000})
	minutesGoal.Goal = Goal{Kind: core.GoalMinutes, Value: 10}
	var mature History
	for i := range 1_000 {
		mature.Mature = append(mature.Mature, d+i/10)
	}
	for _, tc := range []struct {
		name     string
		h        History
		id       string
		progress int
		earned   []int // days, from d
	}{
		{"nothing studied", History{}, "answers", 0, nil},
		// Answers tier (spec): 1,020 answers earn the first two tiers.
		{"answers", daily(d, []int{60, 40, 900, 20}, nil), "answers", 1_020, []int{1, 2}},
		// Undo (spec): one answer short of 100 locks the tier again.
		{"answers, undone below 100", daily(d, []int{60, 39}, nil), "answers", 99, nil},
		{"streak", streakDays, "streak", 8, []int{6}},
		{"streak with a rest day", withRest, "streak", 7, []int{7}},
		{"streak without the rest day", func() History { h := withRest; h.Rest = 0; return h }(), "streak", 4, nil},
		{"study days", daily(d, repeat(3, 100), nil), "study-days", 100, []int{29, 99}},
		{"study days with gaps", daily(d, append(repeat(0, 50), repeat(1, 30)...), nil), "study-days", 30, []int{79}},
		{"daily goal in cards", cardsGoal, "daily-goal", 7, []int{8}},
		{"daily goal in minutes", minutesGoal, "daily-goal", 6, nil},
		// No goal (spec): no progress.
		{"no goal", daily(d, repeat(50, 10), nil), "daily-goal", 0, nil},
		// Study time tier (spec): 9 hours 50 minutes, then 15 minutes.
		{"study time", daily(d, []int{1, 1}, []int64{590 * minute, 15 * minute}), "study-time", 10, []int{1}},
		{"study time in whole hours", daily(d, []int{1, 1}, []int64{590 * minute, 9 * minute}), "study-time", 9, nil},
		{"study time tiers", daily(d, repeat(1, 4), []int64{100 * hour, 400 * hour, 499 * hour, 2 * hour}),
			"study-time", 1_001, []int{0, 0, 1, 3}},
		// Undo below a threshold locks the tier again.
		{"study time, undone below 10 hours", daily(d, []int{1, 1}, []int64{590 * minute, 10*minute - 1}), "study-time", 9, nil},
		{"long-term memory", mature, "long-term", 1_000, []int{9, 99}},
		// Comeback (spec): studying today after 15 days without study.
		{"comeback after 15 days", daily(d, append(append([]int{1}, repeat(0, 15)...), 1), nil), "comeback", 1, []int{16}},
		{"comeback after exactly 14 days", daily(d, append(append([]int{1}, repeat(0, 14)...), 1), nil), "comeback", 1, []int{15}},
		{"no comeback after 13 days", daily(d, append(append([]int{1}, repeat(0, 13)...), 1), nil), "comeback", 0, nil},
		{"the latest comeback", daily(d, append(append(append(append([]int{1}, repeat(0, 20)...), 1, 1), repeat(0, 30)...), 1), nil),
			"comeback", 1, []int{53}},
		{"the first study day is no comeback", daily(d+100, []int{1}, nil), "comeback", 0, nil},
	} {
		got := achievement(t, tc.h, tc.id)
		var want []string
		for _, day := range tc.earned {
			want = append(want, Date(d+day))
		}
		if got.Progress != tc.progress || !slices.Equal(got.Earned, want) {
			t.Errorf("%s: progress %d, earned %v\nwant %d, %v", tc.name, got.Progress, got.Earned, tc.progress, want)
		}
	}
}

func TestAchievementCatalogue(t *testing.T) {
	// Achievement catalogue (spec).
	want := map[string][]int{
		"answers": {100, 1_000, 10_000, 100_000}, "streak": {7, 30, 100, 365}, "study-days": {30, 100, 365, 1_000},
		"study-time": {10, 100, 500, 1_000}, "daily-goal": {7, 30, 100}, "long-term": {100, 1_000, 5_000}, "level": {5, 10, 25, 50}, "comeback": {1},
	}
	all := EvalAchievements(History{})
	if len(all) != len(want) {
		t.Fatalf("%d achievements, want %d", len(all), len(want))
	}
	for _, a := range all {
		progress := 0
		if a.ID == "level" {
			progress = 1 // levels start at 1
		}
		if !slices.Equal(a.Tiers, want[a.ID]) || len(a.TierNames) != len(a.Tiers) || a.Progress != progress || len(a.Earned) != 0 {
			t.Errorf("%+v", a)
		}
	}
	names := func(id string) []string { return achievement(t, History{}, id).TierNames }
	if got := names("answers"); !slices.Equal(got, []string{Bronze, Silver, Gold, Platinum}) {
		t.Errorf("answers tiers %v", got)
	}
	if got := names("long-term"); !slices.Equal(got, []string{Bronze, Silver, Gold}) {
		t.Errorf("long-term tiers %v", got)
	}
	if got := names("comeback"); !slices.Equal(got, []string{Gold}) {
		t.Errorf("comeback tiers %v", got)
	}
	var order []string
	for _, a := range all {
		order = append(order, a.ID)
	}
	if want := []string{"answers", "streak", "study-days", "study-time", "daily-goal", "long-term", "level", "comeback"}; !slices.Equal(order, want) {
		t.Errorf("order %v", order)
	}
	if !achievement(t, History{}, "daily-goal").NeedsGoal || achievement(t, History{Goal: Goal{Kind: core.GoalCards, Value: 5}}, "daily-goal").NeedsGoal {
		t.Error("NeedsGoal")
	}
	if key := achievement(t, History{}, "answers").Key(1); key != "answers:1000" {
		t.Errorf("key %q", key)
	}
}
