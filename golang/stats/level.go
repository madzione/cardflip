package stats

import "math"

// XP earned from the review history (see the progress spec).
const (
	XPPerAnswer   = 1
	XPPerStudyDay = 10
	XPPerGoalDay  = 20
	XPPerLongTerm = 5
)

// XPPerTier is the XP an achievement tier earns by tier name, except Level
// tiers, which earn none.
var XPPerTier = map[string]int{Bronze: 25, Silver: 50, Gold: 100, Platinum: 200}

// LevelXP is the XP needed in total to reach level, 50 per level below
// it: 25·L·(L−1).
func LevelXP(level int) int { return 25 * level * (level - 1) }

// LevelOf returns the level of xp XP, the XP at which that level started
// and the XP at which the next one starts. Levels start at 1 and have no
// top.
func LevelOf(xp int) (level, start, next int) {
	level = max(int((1+math.Sqrt(1+4*float64(xp)/25))/2), 1)
	// Correct float rounding at exact thresholds.
	for level > 1 && LevelXP(level) > xp {
		level--
	}
	for LevelXP(level+1) <= xp {
		level++
	}
	return level, LevelXP(level), LevelXP(level + 1)
}

// LevelStatus is the level, its XP and where they came from.
type LevelStatus struct {
	Level int `json:"level"`
	XP    int `json:"xp"`
	// LevelStart is the XP at which Level started, NextLevelXP those at
	// which the next one starts.
	LevelStart  int `json:"levelStart"`
	NextLevelXP int `json:"nextLevelXp"`
	// The XP from answers, study days, days that met the daily goal,
	// cards reaching long-term memory and achievement tiers; they add up
	// to XP.
	FromAnswers      int `json:"fromAnswers"`
	FromStudyDays    int `json:"fromStudyDays"`
	FromGoalDays     int `json:"fromGoalDays"`
	FromLongTerm     int `json:"fromLongTerm"`
	FromAchievements int `json:"fromAchievements"`
}
