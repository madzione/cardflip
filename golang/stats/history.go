package stats

import (
	"time"
)

// Calendar is the number of answers per study day over the last year.
type Calendar struct {
	// Today is the current study day's date, "2006-01-02".
	Today string `json:"today"`
	// Days are the days with answers, oldest first.
	Days []CalendarDay `json:"days"`
	// Max is the most answers on one day.
	Max int `json:"max"`
}

// CalendarDay is one study day with answers.
type CalendarDay struct {
	// Day is relative to today: 0 is today, -1 yesterday.
	Day   int `json:"day"`
	Count int `json:"count"`
}

// calendarDays is the number of days the calendar covers, today included.
const calendarDays = 365

// ReviewCalendar returns the answers per day of the last year.
func ReviewCalendar(d *Data) Calendar {
	counts := make([]int, calendarDays)
	for i := range d.Reviews {
		if back := d.Today - d.Reviews[i].Day; back >= 0 && back < calendarDays {
			counts[calendarDays-1-back]++
		}
	}
	out := Calendar{Today: Date(d.Today), Days: []CalendarDay{}}
	for i, n := range counts {
		if n > 0 {
			out.Days = append(out.Days, CalendarDay{Day: i - calendarDays + 1, Count: n})
			out.Max = max(out.Max, n)
		}
	}
	return out
}

// History units.
const (
	UnitDay   = "day"
	UnitWeek  = "week"
	UnitMonth = "month"
)

// ReviewHistory is the answers of a period grouped by day, week or month, split
// by what the cards were.
type ReviewHistory struct {
	// Unit is UnitDay for a month or three, UnitWeek for a year and
	// UnitMonth for all history.
	Unit  string `json:"unit"`
	Today string `json:"today"`
	// Buckets are oldest first; the last one contains today.
	Buckets []ReviewBucket `json:"buckets"`
	// Total is the sum of all buckets.
	Total int64 `json:"total"`
	// DaysStudied is the number of days with answers among the period's
	// PeriodDays.
	DaysStudied int `json:"daysStudied"`
	PeriodDays  int `json:"periodDays"`
}

// ReviewBucket is one day, week or month: a number of answers, or for
// review time the milliseconds spent on them.
type ReviewBucket struct {
	// Start is the bucket's first study day, relative to today.
	Start      int   `json:"start"`
	Learning   int64 `json:"learning"`
	Young      int64 `json:"young"`
	Mature     int64 `json:"mature"`
	Relearning int64 `json:"relearning"`
	Filtered   int64 `json:"filtered"`
}

func (b *ReviewBucket) add(k Kind, v int64) {
	switch k {
	case Learning:
		b.Learning += v
	case Young:
		b.Young += v
	case Mature:
		b.Mature += v
	case Relearning:
		b.Relearning += v
	case Filtered:
		b.Filtered += v
	}
}

// Reviews returns the number of answers of period.
func Reviews(d *Data, period string) ReviewHistory {
	return history(d, period, func(*Review) int64 { return 1 })
}

// ReviewTime returns the time spent answering in period, in milliseconds.
func ReviewTime(d *Data, period string) ReviewHistory {
	return history(d, period, func(r *Review) int64 { return r.DurationMs })
}

func history(d *Data, period string, value func(*Review) int64) ReviewHistory {
	first := d.firstDay(period)
	out := ReviewHistory{Today: Date(d.Today), PeriodDays: d.Today - first + 1}
	// starts are the buckets' first days (absolute), oldest first; index
	// maps a day to its bucket.
	var starts []int
	var index func(day int) int
	switch {
	case period == PeriodAll:
		out.Unit = UnitMonth
		for m := monthStart(first); m <= d.Today; m = nextMonth(m) {
			starts = append(starts, m)
		}
		index = func(day int) int {
			lo, hi := 0, len(starts)-1
			for lo < hi {
				mid := (lo + hi + 1) / 2
				if starts[mid] <= day {
					lo = mid
				} else {
					hi = mid - 1
				}
			}
			return lo
		}
	case period == PeriodYear:
		out.Unit = UnitWeek
		weeks := (out.PeriodDays + 6) / 7
		for w := weeks - 1; w >= 0; w-- {
			starts = append(starts, max(d.Today-7*w-6, first))
		}
		index = func(day int) int { return weeks - 1 - (d.Today-day)/7 }
	default:
		out.Unit = UnitDay
		for day := first; day <= d.Today; day++ {
			starts = append(starts, day)
		}
		index = func(day int) int { return day - first }
	}
	out.Buckets = make([]ReviewBucket, len(starts))
	for i, s := range starts {
		out.Buckets[i].Start = s - d.Today
	}
	studied := map[int]bool{}
	for i := range d.Reviews {
		r := &d.Reviews[i]
		if r.Day < first || r.Day > d.Today {
			continue
		}
		v := value(r)
		out.Buckets[index(r.Day)].add(r.Kind, v)
		out.Total += v
		studied[r.Day] = true
	}
	out.DaysStudied = len(studied)
	return out
}

// monthStart returns the study day number of the first of day's month.
func monthStart(day int) int {
	t := time.Unix(int64(day)*86400, 0).UTC()
	return int(time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Unix() / 86400)
}

func nextMonth(start int) int {
	t := time.Unix(int64(start)*86400, 0).UTC()
	return int(t.AddDate(0, 1, 0).Unix() / 86400)
}

// HourBucket is the answers given in one hour of the day.
type HourBucket struct {
	Reviews int `json:"reviews"`
	// Correct is the share of them that were not Again.
	Correct float64 `json:"correct"`
}

// Hourly returns the answers of period by local hour of the day, 0 to 23.
func Hourly(d *Data, period string) []HourBucket {
	first := d.firstDay(period)
	var total, passed [24]int
	for i := range d.Reviews {
		r := &d.Reviews[i]
		if r.Day < first || r.Day > d.Today || r.Hour < 0 || r.Hour > 23 {
			continue
		}
		total[r.Hour]++
		if r.Passed() {
			passed[r.Hour]++
		}
	}
	out := make([]HourBucket, 24)
	for h := range out {
		out[h] = HourBucket{Reviews: total[h], Correct: ratio(passed[h], total[h])}
	}
	return out
}

// ButtonStats are the answer button counts of period for learning (and
// relearning), young and mature cards.
type ButtonStats struct {
	Learning ButtonGroup `json:"learning"`
	Young    ButtonGroup `json:"young"`
	Mature   ButtonGroup `json:"mature"`
}

// ButtonGroup counts the answers per button.
type ButtonGroup struct {
	// Counts are Again, Hard, Good and Easy.
	Counts []int `json:"counts"`
	Total  int   `json:"total"`
	// Correct is the share of answers that were not Again.
	Correct float64 `json:"correct"`
}

// AnswerButtons returns the answer button counts of period. Filtered
// reviews count as young or mature by their interval.
func AnswerButtons(d *Data, period string) ButtonStats {
	first := d.firstDay(period)
	var counts [3][4]int
	for i := range d.Reviews {
		r := &d.Reviews[i]
		if r.Day < first || r.Day > d.Today {
			continue
		}
		g := 0
		switch {
		case r.Kind == Learning || r.Kind == Relearning:
		case mature(int(r.IvlBefore)):
			g = 2
		default:
			g = 1
		}
		counts[g][r.Rating-1]++
	}
	group := func(c [4]int) ButtonGroup {
		total := c[0] + c[1] + c[2] + c[3]
		return ButtonGroup{Counts: c[:], Total: total, Correct: ratio(total-c[0], total)}
	}
	return ButtonStats{Learning: group(counts[0]), Young: group(counts[1]), Mature: group(counts[2])}
}
