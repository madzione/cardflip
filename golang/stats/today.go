package stats

import (
	"time"

	"cardflip/golang/scheduler"
)

// TodayStats summarises the current study day.
type TodayStats struct {
	// Cards is the number of answers, Time the time spent on them.
	Cards int           `json:"cards"`
	Time  time.Duration `json:"time"`
	Again int           `json:"again"`
	// Correct is the share of answers that were not Again; 0 without
	// answers.
	Correct float64 `json:"correct"`
	// Learning, Review, Relearning and Filtered split Cards by what the
	// card was before the answer; Review counts young and mature cards.
	Learning   int `json:"learning"`
	Review     int `json:"review"`
	Relearning int `json:"relearning"`
	Filtered   int `json:"filtered"`
}

// TodaySummary returns the summary of the current study day.
func TodaySummary(d *Data) TodayStats {
	var out TodayStats
	var ms int64
	for i := range d.Reviews {
		r := &d.Reviews[i]
		if r.Day != d.Today {
			continue
		}
		out.Cards++
		ms += r.DurationMs
		if !r.Passed() {
			out.Again++
		}
		switch r.Kind {
		case Learning:
			out.Learning++
		case Young, Mature:
			out.Review++
		case Relearning:
			out.Relearning++
		case Filtered:
			out.Filtered++
		}
	}
	out.Time = time.Duration(ms) * time.Millisecond
	out.Correct = ratio(out.Cards-out.Again, out.Cards)
	return out
}

func ratio(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total)
}

// Forecast is the forecast of due cards.
type Forecast struct {
	// Days[i] is the number of cards due i study days from today; today
	// counts cards due later today, not overdue ones.
	Days []int `json:"days"`
	// Cumulative[i] is the sum of Days[0..i].
	Cumulative []int `json:"cumulative"`
	// Overdue is the number of cards due before today.
	Overdue int `json:"overdue"`
	// Total is the sum of Days, Average the cards due per day.
	Total       int     `json:"total"`
	Average     float64 `json:"average"`
	DueTomorrow int     `json:"dueTomorrow"`
}

// FutureDue returns the forecast of due cards for period, counting the
// scope's cards that are not new or suspended. For PeriodAll it reaches
// the last due card.
func FutureDue(d *Data, period string) Forecast {
	n, ok := periodDays[period]
	due := make([]int, 0, len(d.Cards))
	var out Forecast
	for i := range d.Cards {
		c := &d.Cards[i]
		if c.State == scheduler.New || c.Suspended {
			continue
		}
		day := d.DayOf(c.Due) - d.Today
		if day < 0 {
			out.Overdue++
			continue
		}
		due = append(due, day)
		if !ok {
			n = max(n, day+1)
		}
	}
	n = max(n, 2)
	out.Days = make([]int, n)
	for _, day := range due {
		if day < n {
			out.Days[day]++
		}
	}
	out.Cumulative = make([]int, n)
	for i, v := range out.Days {
		out.Total += v
		out.Cumulative[i] = out.Total
	}
	out.Average = float64(out.Total) / float64(n)
	out.DueTomorrow = out.Days[1]
	return out
}

// StateCounts are the scope's cards by state. Suspended and buried cards are
// counted there only, not by their scheduling state.
type StateCounts struct {
	New        int `json:"new"`
	Learning   int `json:"learning"`
	Relearning int `json:"relearning"`
	Young      int `json:"young"`
	Mature     int `json:"mature"`
	Suspended  int `json:"suspended"`
	Buried     int `json:"buried"`
	Total      int `json:"total"`
}

// CardCounts counts the scope's cards by state.
func CardCounts(d *Data) StateCounts {
	var out StateCounts
	for i := range d.Cards {
		c := &d.Cards[i]
		out.Total++
		switch {
		case c.Suspended:
			out.Suspended++
		case c.BuriedUntil.After(d.Now):
			out.Buried++
		case c.State == scheduler.New:
			out.New++
		case c.State == scheduler.Learning:
			out.Learning++
		case c.State == scheduler.Relearning:
			out.Relearning++
		case mature(c.ScheduledDays):
			out.Mature++
		default:
			out.Young++
		}
	}
	return out
}
