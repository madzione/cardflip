package stats

import (
	"slices"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// dayRange returns the study days from first to last.
func dayRange(first, last int) []int {
	var out []int
	for d := first; d <= last; d++ {
		out = append(out, d)
	}
	return out
}

func TestStreaks(t *testing.T) {
	const today = 20000
	for _, tc := range []struct {
		name string
		days []int
		want StreakStats
	}{
		{"empty history", nil, StreakStats{}},
		{"only today", []int{today},
			StreakStats{Current: 1, TodayStudied: true, CurrentFirst: today, Longest: 1,
				LongestFirst: today, LongestLast: today}},
		// Today not studied yet (spec): the last 12 days, not today.
		{"today not studied yet", dayRange(today-12, today-1),
			StreakStats{Current: 12, AtRisk: true, CurrentFirst: today - 12, Longest: 12,
				LongestFirst: today - 12, LongestLast: today - 1}},
		// Gap of one day (spec): today and 5 days before yesterday.
		{"one-day gap", append(dayRange(today-6, today-2), today),
			StreakStats{Current: 1, TodayStudied: true, CurrentFirst: today, Longest: 5,
				LongestFirst: today - 6, LongestLast: today - 2}},
		{"single long run", dayRange(today-99, today),
			StreakStats{Current: 100, TodayStudied: true, CurrentFirst: today - 99, Longest: 100,
				LongestFirst: today - 99, LongestLast: today}},
		// Broken streak (spec): 40 days, then three days off, then today.
		{"longest run in the past", append(dayRange(today-43, today-4), today),
			StreakStats{Current: 1, TodayStudied: true, CurrentFirst: today, Longest: 40,
				LongestFirst: today - 43, LongestLast: today - 4}},
		{"broken, nothing since", dayRange(today-43, today-4),
			StreakStats{Longest: 40, LongestFirst: today - 43, LongestLast: today - 4}},
		{"equal runs report the latest", append(dayRange(today-10, today-8), dayRange(today-3, today-1)...),
			StreakStats{Current: 3, AtRisk: true, CurrentFirst: today - 3, Longest: 3,
				LongestFirst: today - 3, LongestLast: today - 1}},
	} {
		if got := Streaks(tc.days, today, 0); got != tc.want {
			t.Errorf("%s: %+v\nwant %+v", tc.name, got, tc.want)
		}
	}
}

// without returns days without the days skip.
func without(days []int, skip ...int) []int {
	return slices.DeleteFunc(slices.Clone(days), func(d int) bool { return slices.Contains(skip, d) })
}

func TestStreaksWithRestDays(t *testing.T) {
	const today = 20000
	for _, tc := range []struct {
		name string
		rest int
		days []int
		want StreakStats
	}{
		{"no rest days: a skip breaks", 0, without(dayRange(today-9, today), today-4),
			StreakStats{Current: 4, TodayStudied: true, CurrentFirst: today - 3, Longest: 5,
				LongestFirst: today - 9, LongestLast: today - 5}},
		// One rest day (spec): the last 10 days except 4 days ago.
		{"one rest day", 1, without(dayRange(today-9, today), today-4),
			StreakStats{Current: 9, TodayStudied: true, CurrentFirst: today - 9, Longest: 9,
				LongestFirst: today - 9, LongestLast: today}},
		// Too many rest days (spec): two skips within a week; the streak
		// restarts after the second.
		{"two skips in one window", 1, without(dayRange(today-8, today), today-5, today-2),
			StreakStats{Current: 2, TodayStudied: true, CurrentFirst: today - 1, Longest: 5,
				LongestFirst: today - 8, LongestLast: today - 3}},
		{"two rest days allow two skips", 2, without(dayRange(today-8, today), today-5, today-2),
			StreakStats{Current: 7, TodayStudied: true, CurrentFirst: today - 8, Longest: 7,
				LongestFirst: today - 8, LongestLast: today}},
		{"two rest days, three skips in one window", 2, without(dayRange(today-9, today), today-7, today-5, today-3),
			StreakStats{Current: 3, TodayStudied: true, CurrentFirst: today - 2, Longest: 4,
				LongestFirst: today - 9, LongestLast: today - 4}},
		{"skips a week apart", 1, without(dayRange(today-10, today), today-8, today-1),
			StreakStats{Current: 9, TodayStudied: true, CurrentFirst: today - 10, Longest: 9,
				LongestFirst: today - 10, LongestLast: today}},
		{"skips six days apart", 1, without(dayRange(today-10, today), today-7, today-1),
			StreakStats{Current: 1, TodayStudied: true, CurrentFirst: today, Longest: 8,
				LongestFirst: today - 10, LongestLast: today - 2}},
		{"two days off in a row with one rest day", 1, append(dayRange(today-6, today-3), today),
			StreakStats{Current: 1, TodayStudied: true, CurrentFirst: today, Longest: 4,
				LongestFirst: today - 6, LongestLast: today - 3}},
		// Skips at the streak's edges: after the first study day, and
		// yesterday with today not studied yet.
		{"skip after the first day", 1, append([]int{today - 5}, dayRange(today-3, today)...),
			StreakStats{Current: 5, TodayStudied: true, CurrentFirst: today - 5, Longest: 5,
				LongestFirst: today - 5, LongestLast: today}},
		{"yesterday skipped, at risk today", 1, dayRange(today-6, today-2),
			StreakStats{Current: 5, AtRisk: true, CurrentFirst: today - 6, Longest: 5,
				LongestFirst: today - 6, LongestLast: today - 2}},
		{"yesterday skipped, rest day already used", 1, without(dayRange(today-6, today-2), today-4),
			StreakStats{Longest: 4, LongestFirst: today - 6, LongestLast: today - 2}},
		{"two days skipped, at risk with two rest days", 2, dayRange(today-6, today-3),
			StreakStats{Current: 4, AtRisk: true, CurrentFirst: today - 6, Longest: 4,
				LongestFirst: today - 6, LongestLast: today - 3}},
		{"two days skipped, broken with one rest day", 1, dayRange(today-6, today-3),
			StreakStats{Longest: 4, LongestFirst: today - 6, LongestLast: today - 3}},
		// Applied to history (spec): the longest streak counts rest days.
		{"longest streak with a rest day", 1, append(without(dayRange(today-30, today-20), today-25), today),
			StreakStats{Current: 1, TodayStudied: true, CurrentFirst: today, Longest: 10,
				LongestFirst: today - 30, LongestLast: today - 20}},
	} {
		if got := Streaks(tc.days, today, tc.rest); got != tc.want {
			t.Errorf("%s: %+v\nwant %+v", tc.name, got, tc.want)
		}
	}
}

// studyDays maps answer times to their sorted, distinct study days, as the
// statistics and the progress service do.
func studyDays(times []time.Time, rollover int) []int {
	var out []int
	for _, t := range times {
		out = append(out, scheduler.DayNumber(t, rollover))
	}
	slices.Sort(out)
	return slices.Compact(out)
}

func TestStudyDaysAcrossDST(t *testing.T) {
	// Daylight saving change (spec): one answer every day across both of
	// Berlin's changes, at 02:30 (the previous study day with rollover 4,
	// and on the spring change an hour that does not exist) and at 04:30.
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skip("no time zone data:", err)
	}
	for _, change := range []time.Time{
		time.Date(2026, 3, 29, 12, 0, 0, 0, berlin),
		time.Date(2026, 10, 25, 12, 0, 0, 0, berlin),
	} {
		for _, hour := range []int{2, 4} {
			var times []time.Time
			for i := -5; i <= 5; i++ {
				y, m, d := change.AddDate(0, 0, i).Date()
				times = append(times, time.Date(y, m, d, hour, 30, 0, 0, berlin))
			}
			days := studyDays(times, 4)
			if len(days) != len(times) {
				t.Fatalf("%s at %02d:30: %d answers on %d study days: %v", change.Format(time.DateOnly), hour, len(times), len(days), days)
			}
			today := scheduler.DayNumber(times[len(times)-1], 4)
			if s := Streaks(days, today, 0); s.Current != len(times) || !s.TodayStudied {
				t.Fatalf("%s at %02d:30: streak %+v, want %d days", change.Format(time.DateOnly), hour, s, len(times))
			}
		}
	}
	// Rollover hour (spec): with rollover 4, 02:30 belongs to the previous
	// study day.
	night := time.Date(2026, 10, 13, 2, 30, 0, 0, berlin)
	evening := time.Date(2026, 10, 12, 20, 0, 0, 0, berlin)
	if days := studyDays([]time.Time{evening, night}, 4); len(days) != 1 {
		t.Fatalf("02:30 and the previous evening: %v", days)
	}
}
