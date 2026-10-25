package stats

import (
	"slices"
	"time"

	"cardflip/golang/core"
)

// StreakStats are the streaks of a set of study days.
type StreakStats struct {
	// Current is the number of consecutive study days ending today, or
	// ending yesterday when today is not a study day yet.
	Current int `json:"current"`
	// AtRisk reports a current streak that today has not extended yet:
	// Current > 0 and today is not a study day.
	AtRisk bool `json:"atRisk"`
	// TodayStudied reports whether today is a study day.
	TodayStudied bool `json:"todayStudied"`
	// Longest is the longest run of consecutive study days, from the study
	// day LongestFirst to LongestLast (both 0 without study days). Of runs
	// of equal length the latest is reported.
	Longest      int `json:"longest"`
	LongestFirst int `json:"longestFirst"`
	LongestLast  int `json:"longestLast"`
}

// Streaks returns the streaks of days, a sorted set of study day numbers
// (scheduler.DayNumber) without duplicates, on the study day today. Days
// after today are left out of the current streak.
func Streaks(days []int, today int) StreakStats {
	var out StreakStats
	run, first := 0, 0
	for i, day := range days {
		if i > 0 && day == days[i-1]+1 {
			run++
		} else {
			run, first = 1, day
		}
		if run >= out.Longest {
			out.Longest, out.LongestFirst, out.LongestLast = run, first, day
		}
		switch day {
		case today:
			out.TodayStudied = true
			out.Current = run
		case today - 1:
			out.Current = run
		}
	}
	out.AtRisk = out.Current > 0 && !out.TodayStudied
	return out
}

// Goal is a daily goal: Kind core.GoalCards or core.GoalMinutes with a
// number of cards or minutes, or Kind empty for none.
type Goal struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}

// Met reports whether a day with cards answers taking ms milliseconds met
// the goal; false without a goal.
func (g Goal) Met(cards int, ms int64) bool {
	switch g.Kind {
	case core.GoalCards:
		return cards >= g.Value
	case core.GoalMinutes:
		return ms >= int64(g.Value)*time.Minute.Milliseconds()
	}
	return false
}

// StreakSection is the statistics' Streaks section.
type StreakSection struct {
	Current      int  `json:"current"`
	AtRisk       bool `json:"atRisk"`
	TodayStudied bool `json:"todayStudied"`
	// Longest is the longest streak, from LongestFirst to LongestLast
	// ("2006-01-02"; empty without study days).
	Longest      int    `json:"longest"`
	LongestFirst string `json:"longestFirst,omitempty"`
	LongestLast  string `json:"longestLast,omitempty"`
	// StudyDays is the number of study days among the period's PeriodDays.
	StudyDays  int `json:"studyDays"`
	PeriodDays int `json:"periodDays"`
	// GoalDays is the number of days of the period that met the current
	// daily goal; nil without a goal.
	GoalDays *int `json:"goalDays,omitempty"`
}

// StudyDays returns the sorted study days of reviews.
func StudyDays(reviews []Review) []int {
	days := make([]int, 0, len(reviews))
	for i := range reviews {
		days = append(days, reviews[i].Day)
	}
	slices.Sort(days)
	return slices.Compact(days)
}

// StreakSummary returns the Streaks section of the scope for period, judging
// the period's days against the current goal. The streaks cover the whole
// history; the counts the period.
func StreakSummary(d *Data, goal Goal, period string) StreakSection {
	s := Streaks(StudyDays(d.Reviews), d.Today)
	first := d.firstDay(period)
	out := StreakSection{Current: s.Current, AtRisk: s.AtRisk, TodayStudied: s.TodayStudied, Longest: s.Longest,
		PeriodDays: d.Today - first + 1}
	if s.Longest > 0 {
		out.LongestFirst, out.LongestLast = Date(s.LongestFirst), Date(s.LongestLast)
	}
	type day struct {
		cards int
		ms    int64
	}
	days := map[int]*day{}
	for i := range d.Reviews {
		r := &d.Reviews[i]
		if r.Day < first || r.Day > d.Today {
			continue
		}
		dd := days[r.Day]
		if dd == nil {
			dd = &day{}
			days[r.Day] = dd
		}
		dd.cards++
		dd.ms += r.DurationMs
	}
	out.StudyDays = len(days)
	if goal.Kind != "" {
		met := 0
		for _, dd := range days {
			if goal.Met(dd.cards, dd.ms) {
				met++
			}
		}
		out.GoalDays = &met
	}
	return out
}
