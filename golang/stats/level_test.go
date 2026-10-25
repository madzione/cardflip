package stats

import (
	"slices"
	"testing"

	"cardflip/golang/core"
)

func TestLevelOf(t *testing.T) {
	for _, tc := range []struct{ xp, level, start, next int }{
		// Never studied (spec): level 1 with 0 / 50 XP.
		{0, 1, 0, 50},
		{49, 1, 0, 50},
		{50, 2, 50, 150},
		// Just below a level (spec): level 9 with 449 / 450 XP.
		{2_249, 9, 1_800, 2_250},
		// Exactly at a level (spec): level 10 with 0 / 500 XP.
		{2_250, 10, 2_250, 2_750},
		{3_760, 12, 3_300, 3_900},
		{61_250, 50, 61_250, 63_750},
		{LevelXP(100_000) - 1, 99_999, LevelXP(99_999), LevelXP(100_000)},
		{LevelXP(100_000), 100_000, LevelXP(100_000), LevelXP(100_001)},
	} {
		level, start, next := LevelOf(tc.xp)
		if level != tc.level || start != tc.start || next != tc.next {
			t.Errorf("LevelOf(%d) = %d, %d, %d; want %d, %d, %d", tc.xp, level, start, next, tc.level, tc.start, tc.next)
		}
	}
	// Every threshold up to level 2,000 exactly.
	for l := 1; l <= 2_000; l++ {
		if got, _, _ := LevelOf(LevelXP(l)); got != l {
			t.Errorf("LevelOf(%d) = %d, want %d", LevelXP(l), got, l)
		}
		if got, _, _ := LevelOf(LevelXP(l+1) - 1); got != l {
			t.Errorf("LevelOf(%d) = %d, want %d", LevelXP(l+1)-1, got, l)
		}
	}
}

func TestXP(t *testing.T) {
	const d = 20000
	cards20 := Goal{Kind: core.GoalCards, Value: 20}
	for _, tc := range []struct {
		name string
		h    History
		want LevelStatus // the breakdown and XP
	}{
		{"nothing studied", History{}, LevelStatus{}},
		// One day (spec): 30 answers, the study day, the goal and 2 cards
		// reaching long-term memory.
		{"one day", History{Days: []int{d}, Answers: []int{30}, Ms: []int64{0}, Mature: []int{d, d}, Goal: cards20},
			LevelStatus{XP: 70, FromAnswers: 30, FromStudyDays: 10, FromGoalDays: 20, FromLongTerm: 10}},
		{"goal not met", History{Days: []int{d}, Answers: []int{19}, Ms: []int64{0}, Goal: cards20},
			LevelStatus{XP: 29, FromAnswers: 19, FromStudyDays: 10}},
		// Achievement tier (spec): the 100th answer earns bronze Answers.
		{"a tier", daily(d, []int{60, 40}, nil),
			LevelStatus{XP: 100 + 20 + 25, FromAnswers: 100, FromStudyDays: 20, FromAchievements: 25}},
		// Comeback's single tier counts as gold.
		{"comeback", daily(d, append(append([]int{1}, repeat(0, 15)...), 1), nil),
			LevelStatus{XP: 2 + 20 + 100, FromAnswers: 2, FromStudyDays: 20, FromAchievements: 100}},
	} {
		_, got := EvalProgress(tc.h)
		got.Level, got.LevelStart, got.NextLevelXP = 0, 0, 0
		if got != tc.want {
			t.Errorf("%s: %+v\nwant %+v", tc.name, got, tc.want)
		}
	}
	// Each XP belongs to its study day: the tier on the second day.
	h := daily(d, []int{60, 40}, nil)
	earned := make([][]int, len(Achievements))
	earned[0] = []int{d + 1} // answers
	h.XP, _ = xpOf(h, earned)
	if h.XP[0] != 70 || h.XP[1] != 75 {
		t.Errorf("XP per day %v", h.XP)
	}
}

func TestLevelAchievement(t *testing.T) {
	const d = 20000
	// 200 answers a day for 10 days: 2,000 + 100 XP, plus bronze Answers
	// (25) on the first day, silver (50) on the fifth and bronze Streak
	// (25) on the seventh: 2,200, level 9.
	h := daily(d, repeat(200, 10), nil)
	_, level := EvalProgress(h)
	if level.XP != 2_200 || level.Level != 9 || level.FromAchievements != 100 {
		t.Fatalf("level %+v", level)
	}
	// Level 5 (500 XP) on the third day: 235, 445, 655.
	lvl := achievement(t, h, "level")
	if lvl.Progress != 9 || !slices.Equal(lvl.Earned, []string{Date(d + 2)}) {
		t.Errorf("level achievement %+v", lvl)
	}
	// Level tier (spec): earned on the day XP reach 2,250. 75 answers on
	// the 11th day make 2,200 + 85 = 2,285.
	h = daily(d, append(repeat(200, 10), 75), nil)
	_, level = EvalProgress(h)
	lvl = achievement(t, h, "level")
	if level.XP != 2_285 || level.Level != 10 || !slices.Equal(lvl.Earned, []string{Date(d + 2), Date(d + 10)}) {
		t.Errorf("level %+v, achievement %+v", level, lvl)
	}
	// Level tiers earn no XP: the two of them add nothing.
	if level.FromAchievements != 100 {
		t.Errorf("from achievements %d", level.FromAchievements)
	}
	// Raising the goal can lower the level and un-earn a tier: with a goal
	// of 200 cards every day earns 20 more and bronze Daily goal 25 (2,425,
	// level 10), with 201 none (level 9).
	h = daily(d, repeat(200, 10), nil)
	h.Goal = Goal{Kind: core.GoalCards, Value: 200}
	if _, level = EvalProgress(h); level.XP != 2_425 || level.Level != 10 || len(achievement(t, h, "level").Earned) != 2 {
		t.Errorf("goal 200: %+v", level)
	}
	h.Goal.Value = 201
	if _, level = EvalProgress(h); level.XP != 2_200 || level.Level != 9 || len(achievement(t, h, "level").Earned) != 1 {
		t.Errorf("goal 201: %+v", level)
	}
}
