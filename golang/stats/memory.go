package stats

import (
	"math"
	"slices"

	"cardflip/golang/scheduler"
)

// Histogram is a distribution of card values.
type Histogram struct {
	// Buckets hold the values from Lo up to, not including, Hi; the last
	// one also holds Hi and, for day values, every larger value: its Hi is
	// then the largest value.
	Buckets []HistogramBucket `json:"buckets"`
	// Count is the number of values, Average their mean and Sum their sum.
	Count   int     `json:"count"`
	Average float64 `json:"average"`
	Sum     float64 `json:"sum"`
}

// HistogramBucket is one bar of a histogram.
type HistogramBucket struct {
	Lo    float64 `json:"lo"`
	Hi    float64 `json:"hi"`
	Count int     `json:"count"`
}

// maxDayBuckets is the most buckets a histogram of days has.
const maxDayBuckets = 30

// dayWidths are the bucket widths of day histograms, smallest first.
var dayWidths = []float64{1, 2, 3, 5, 7, 10, 14, 15, 30, 60, 90, 180, 365}

// dayHistogram groups day values starting at start into buckets of a
// width that fits the values up to the 99th percentile into at most
// maxDayBuckets; larger values go into the last bucket.
func dayHistogram(vals []float64, start float64) Histogram {
	out := summarise(vals)
	if len(vals) == 0 {
		return out
	}
	slices.Sort(vals)
	top := vals[((len(vals)-1)*99+99)/100]
	span := top - start + 1
	w := dayWidths[len(dayWidths)-1]
	for _, dw := range dayWidths {
		if math.Ceil(span/dw) <= maxDayBuckets {
			w = dw
			break
		}
	}
	if math.Ceil(span/w) > maxDayBuckets {
		w = 365 * math.Ceil(span/365/maxDayBuckets)
	}
	n := max(int(math.Ceil(span/w)), 1)
	out.Buckets = make([]HistogramBucket, n)
	for i := range out.Buckets {
		out.Buckets[i] = HistogramBucket{Lo: start + float64(i)*w, Hi: start + float64(i+1)*w}
	}
	last := &out.Buckets[n-1]
	for _, v := range vals {
		i := min(max(int((v-start)/w), 0), n-1)
		out.Buckets[i].Count++
		last.Hi = math.Max(last.Hi, v)
	}
	return out
}

// fractionBuckets is the number of buckets of a histogram of values from 0
// to 1.
const fractionBuckets = 20

// fractionHistogram groups values from 0 to 1 into 5 % buckets.
func fractionHistogram(vals []float64) Histogram {
	out := summarise(vals)
	out.Buckets = make([]HistogramBucket, fractionBuckets)
	for i := range out.Buckets {
		out.Buckets[i] = HistogramBucket{Lo: float64(i) / fractionBuckets, Hi: float64(i+1) / fractionBuckets}
	}
	for _, v := range vals {
		out.Buckets[min(max(int(v*fractionBuckets), 0), fractionBuckets-1)].Count++
	}
	return out
}

func summarise(vals []float64) Histogram {
	out := Histogram{Buckets: []HistogramBucket{}, Count: len(vals)}
	for _, v := range vals {
		out.Sum += v
	}
	if len(vals) > 0 {
		out.Average = out.Sum / float64(len(vals))
	}
	return out
}

// Intervals returns the distribution of review cards' intervals, in days.
func Intervals(d *Data) Histogram {
	var vals []float64
	for i := range d.Cards {
		if c := &d.Cards[i]; c.State == scheduler.Review {
			vals = append(vals, float64(max(c.ScheduledDays, 1)))
		}
	}
	return dayHistogram(vals, 1)
}

// Stability returns the distribution of the FSRS stability, in days, of
// cards with a memory state.
func Stability(d *Data) Histogram {
	var vals []float64
	for i := range d.Cards {
		if c := &d.Cards[i]; c.State != scheduler.New && c.Memory != nil {
			vals = append(vals, c.Memory.Stability)
		}
	}
	return dayHistogram(vals, 0)
}

// Difficulty returns the distribution of the FSRS difficulty of cards
// with a memory state, scaled from 1–10 to 0–1.
func Difficulty(d *Data) Histogram {
	var vals []float64
	for i := range d.Cards {
		if c := &d.Cards[i]; c.State != scheduler.New && c.Memory != nil {
			vals = append(vals, min(max((c.Memory.Difficulty-1)/9, 0), 1))
		}
	}
	return fractionHistogram(vals)
}

// Retrievability returns the distribution of cards' current probability
// of recall. Its Average is the average retrievability and its Sum the
// estimated knowledge: the number of cards one would expect to recall now.
func Retrievability(d *Data) Histogram {
	var vals []float64
	for i := range d.Cards {
		if c := &d.Cards[i]; c.State != scheduler.New {
			if r, ok := scheduler.Retrievability(c.CardState, d.Now); ok {
				vals = append(vals, r)
			}
		}
	}
	return fractionHistogram(vals)
}

// RetentionTable is the true retention table.
type RetentionTable struct {
	// Desired is the desired retention to compare with.
	Desired float64        `json:"desired"`
	Rows    []RetentionRow `json:"rows"`
}

// Retention periods.
const (
	RetentionToday     = "today"
	RetentionYesterday = "yesterday"
	RetentionWeek      = "week"
	RetentionMonth     = "month"
	RetentionYear      = "year"
	RetentionAll       = "all"
)

// RetentionRow is one period of the true retention table.
type RetentionRow struct {
	Period string         `json:"period"`
	Young  RetentionCount `json:"young"`
	Mature RetentionCount `json:"mature"`
	Total  RetentionCount `json:"total"`
}

// RetentionCount counts passed and failed reviews.
type RetentionCount struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	// Retention is the share passed; nil without reviews.
	Retention *float64 `json:"retention"`
}

func (c *RetentionCount) add(passed bool) {
	if passed {
		c.Passed++
	} else {
		c.Failed++
	}
}

func (c *RetentionCount) finish() {
	if n := c.Passed + c.Failed; n > 0 {
		r := float64(c.Passed) / float64(n)
		c.Retention = &r
	}
}

// TrueRetention returns the passed and failed reviews of young and mature
// cards (not learning, relearning or filtered answers) for today,
// yesterday, the last 7, 30 and 365 days and all time.
func TrueRetention(d *Data) RetentionTable {
	periods := []struct {
		name     string
		from, to int // days ago
	}{
		{RetentionToday, 0, 0}, {RetentionYesterday, 1, 1}, {RetentionWeek, 6, 0},
		{RetentionMonth, 29, 0}, {RetentionYear, 364, 0}, {RetentionAll, math.MaxInt32, 0},
	}
	rows := make([]RetentionRow, len(periods))
	for i, p := range periods {
		rows[i].Period = p.name
	}
	for i := range d.Reviews {
		r := &d.Reviews[i]
		ago := d.Today - r.Day
		for j, p := range periods {
			if ago >= p.to && ago <= p.from {
				rows[j].add(r)
			}
		}
	}
	for i := range rows {
		rows[i].finish()
	}
	return RetentionTable{Desired: d.Retention, Rows: rows}
}

// RetentionOfDays returns the true retention, counted as in TrueRetention,
// of the study days first to last (scheduler.DayNumber), both included.
// The row has no period name.
func RetentionOfDays(d *Data, first, last int) RetentionRow {
	var row RetentionRow
	for i := range d.Reviews {
		if r := &d.Reviews[i]; r.Day >= first && r.Day <= last {
			row.add(r)
		}
	}
	row.finish()
	return row
}

// add counts r if it is an answer to a young or mature card.
func (row *RetentionRow) add(r *Review) {
	switch r.Kind {
	case Young:
		row.Young.add(r.Passed())
	case Mature:
		row.Mature.add(r.Passed())
	default:
		return
	}
	row.Total.add(r.Passed())
}

func (row *RetentionRow) finish() {
	row.Young.finish()
	row.Mature.finish()
	row.Total.finish()
}
