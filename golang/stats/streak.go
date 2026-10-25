package stats

import (
	"slices"
	"time"

	"cardflip/golang/core"
)

// StreakStats are the streaks of a set of study days.
type StreakStats struct {
	// Current is the number of study days of the streak ending today, or
	// ending yesterday (or on the last study day before it, over rest
	// days) when today is not a study day yet.
	Current int `json:"current"`
	// AtRisk reports a current streak that today has not extended yet:
	// Current > 0 and today is not a study day.
	AtRisk bool `json:"atRisk"`
	// TodayStudied reports whether today is a study day.
	TodayStudied bool `json:"todayStudied"`
	// Longest is the longest streak, from the study day LongestFirst to
	// LongestLast (both 0 without study days). Of streaks of equal length
	// the latest is reported.
	Longest      int `json:"longest"`
	LongestFirst int `json:"longestFirst"`
	LongestLast  int `json:"longestLast"`
}

// streak follows a streak through the study days in order. With rest
// days, a streak goes on over a day without study as long as no 7 days in
// a row of the streak have more than rest such days; one more breaks it,
// and the next study day starts a new one. Rest days do not add to the
// length.
type streak struct {
	rest  int
	run   int // study days
	first int // the first study day
	// skipped are the days without study inside the streak, of the last
	// 7 days at most.
	skipped []int
}

// skip reports whether the streak survives day without study, and records
// it if so.
func (s *streak) skip(day int) bool {
	s.skipped = slices.DeleteFunc(s.skipped, func(d int) bool { return d <= day-7 })
	if len(s.skipped) >= s.rest {
		return false
	}
	s.skipped = append(s.skipped, day)
	return true
}

// study adds the study day day, after the previous one (if any) prev.
func (s *streak) study(prev, day int) {
	alive := s.run > 0
	for d := prev + 1; alive && d < day; d++ {
		alive = s.skip(d)
	}
	if !alive {
		s.run, s.first, s.skipped = 0, day, s.skipped[:0]
	}
	s.run++
}

// alive reports whether the streak, last studied on last, survives the
// days without study up to today, where today itself does not count yet.
// It leaves s as it was.
func (s streak) alive(last, today int) bool {
	s.skipped = slices.Clone(s.skipped)
	for d := last + 1; d < today; d++ {
		if !s.skip(d) {
			return false
		}
	}
	return s.run > 0
}

// Streaks returns the streaks of days, a sorted set of study day numbers
// (scheduler.DayNumber) without duplicates, on the study day today, with
// rest days per week (0 to core.MaxRestDaysPerWeek) that a streak may skip. Days after
// today are left out of the current streak.
func Streaks(days []int, today, rest int) StreakStats {
	var out StreakStats
	s := streak{rest: rest}
	for i, day := range days {
		prev := 0
		if i > 0 {
			prev = days[i-1]
		}
		s.study(prev, day)
		if s.run >= out.Longest {
			out.Longest, out.LongestFirst, out.LongestLast = s.run, s.first, day
		}
		if day == today {
			out.TodayStudied = true
			out.Current = s.run
		} else if day < today && (i+1 == len(days) || days[i+1] > today) && s.alive(day, today) {
			out.Current = s.run
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
// the period's days against the current goal and allowing rest days per
// week. The streaks cover the whole history; the counts the period.
func StreakSummary(d *Data, goal Goal, rest int, period string) StreakSection {
	s := Streaks(StudyDays(d.Reviews), d.Today, rest)
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
