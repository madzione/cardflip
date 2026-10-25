package stats

import "fmt"

// History is what the achievements are computed from: the whole review
// history by study day.
type History struct {
	// Days are the sorted study days (scheduler.DayNumber); Answers[i] is
	// the number of answers on Days[i] and Ms[i] the milliseconds spent on
	// them.
	Days    []int
	Answers []int
	Ms      []int64
	// Mature are the sorted study days on which a card first got an
	// interval of core.MatureInterval days or more, one per card.
	Mature []int
	// Goal is the current daily goal, Rest the rest days per week.
	Goal Goal
	Rest int
}

// Achievement is a milestone of the review history with tiers.
type Achievement struct {
	ID string
	// Tiers are the thresholds, ascending.
	Tiers []int
	// Eval returns the progress toward tiers (the achievement's) and, for
	// each tier reached, the study day on which it was.
	Eval func(h History, tiers []int) (progress int, earned []int)
}

// ComebackGap is the number of days without study after which studying
// again earns the Comeback achievement.
const ComebackGap = 14

// Achievements are all achievements, in the order they are shown. Adding
// one needs its strings and icon in the app.
var Achievements = []Achievement{
	{ID: "answers", Tiers: []int{100, 1_000, 10_000, 100_000}, Eval: func(h History, tiers []int) (int, []int) {
		return cumulative(h.Days, func(i int) int { return h.Answers[i] }, tiers)
	}},
	{ID: "streak", Tiers: []int{7, 30, 100, 365}, Eval: func(h History, tiers []int) (int, []int) {
		var earned []int
		s := streak{rest: h.Rest}
		longest := 0
		for i, day := range h.Days {
			prev := 0
			if i > 0 {
				prev = h.Days[i-1]
			}
			s.study(prev, day)
			longest = max(longest, s.run)
			for len(earned) < len(tiers) && longest >= tiers[len(earned)] {
				earned = append(earned, day)
			}
		}
		return longest, earned
	}},
	{ID: "study-days", Tiers: []int{30, 100, 365, 1_000}, Eval: func(h History, tiers []int) (int, []int) {
		return cumulative(h.Days, func(int) int { return 1 }, tiers)
	}},
	{ID: "daily-goal", Tiers: []int{7, 30, 100}, Eval: func(h History, tiers []int) (int, []int) {
		if h.Goal.Kind == "" {
			return 0, nil
		}
		return cumulative(h.Days, func(i int) int {
			if h.Goal.Met(h.Answers[i], h.Ms[i]) {
				return 1
			}
			return 0
		}, tiers)
	}},
	{ID: "long-term", Tiers: []int{100, 1_000, 5_000}, Eval: func(h History, tiers []int) (int, []int) {
		return cumulative(h.Mature, func(int) int { return 1 }, tiers)
	}},
	{ID: "comeback", Tiers: []int{1}, Eval: func(h History, _ []int) (int, []int) {
		for i := len(h.Days) - 1; i > 0; i-- {
			if h.Days[i]-h.Days[i-1]-1 >= ComebackGap {
				return 1, []int{h.Days[i]}
			}
		}
		return 0, nil
	}},
}

// cumulative adds up count(i) over days and returns the total and the day
// each of tiers was reached.
func cumulative(days []int, count func(i int) int, tiers []int) (int, []int) {
	var earned []int
	total := 0
	for i, day := range days {
		total += count(i)
		for len(earned) < len(tiers) && total >= tiers[len(earned)] {
			earned = append(earned, day)
		}
	}
	return total, earned
}

// Tier names, for the badge's rim.
const (
	Bronze   = "bronze"
	Silver   = "silver"
	Gold     = "gold"
	Platinum = "platinum"
)

// TierNames returns the names of an achievement's tiers: bronze, silver,
// gold and platinum in order, or gold for an achievement of one tier.
func (a Achievement) TierNames() []string {
	if len(a.Tiers) == 1 {
		return []string{Gold}
	}
	return []string{Bronze, Silver, Gold, Platinum}[:len(a.Tiers)]
}

// AchievementStatus is how far an achievement is.
type AchievementStatus struct {
	ID string `json:"id"`
	// Tiers are the thresholds and TierNames their names (see
	// Achievement.TierNames).
	Tiers     []int    `json:"tiers"`
	TierNames []string `json:"tierNames"`
	// Progress is the count toward the tiers: answers, days or cards.
	Progress int `json:"progress"`
	// Earned are the dates ("2006-01-02") of the study days the tiers were
	// reached on, for the tiers reached so far.
	Earned []string `json:"earned"`
	// NeedsGoal reports the Daily goal achievement while no daily goal is
	// set: it makes no progress.
	NeedsGoal bool `json:"needsGoal"`
}

// Key names tier i (from 0) of the achievement in the seen list.
func (a AchievementStatus) Key(i int) string { return fmt.Sprintf("%s:%d", a.ID, a.Tiers[i]) }

// EvalAchievements returns the status of every achievement.
func EvalAchievements(h History) []AchievementStatus {
	out := make([]AchievementStatus, 0, len(Achievements))
	for _, a := range Achievements {
		progress, earned := a.Eval(h, a.Tiers)
		st := AchievementStatus{ID: a.ID, Tiers: a.Tiers, TierNames: a.TierNames(), Progress: progress,
			Earned: make([]string, len(earned)), NeedsGoal: a.ID == "daily-goal" && h.Goal.Kind == ""}
		for i, day := range earned {
			st.Earned[i] = Date(day)
		}
		out = append(out, st)
	}
	return out
}
