package core

import (
	"math"
	"slices"
	"time"

	"cardflip/golang/scheduler"
)

// noteReviewChange advances c.reviewGen for an op that changes the answers
// in the review log other than by a new answer: undo, redo, imports, sync
// and their reverts. Manual entries are no answers and are ignored. apply
// calls it for every op; reset advances the generation too.
func (c *Collection) noteReviewChange(o *op) {
	changed := o.ReviewUndone != 0 || o.Review != nil && o.Kind != "answer"
	for _, e := range append(o.AddReviews[:len(o.AddReviews):len(o.AddReviews)], o.DelReviews...) {
		changed = changed || e.Rating.Valid() && !e.Preview
	}
	if changed {
		c.reviewGen++
	}
}

// StudyDays are the study days of the whole review log.
type StudyDays struct {
	// Days are the sorted study days (scheduler.DayNumber) with at least
	// one answer. Previews and manual entries do not count.
	Days []int
	// Gen is the review generation they were read at (see RecentAnswers),
	// Last the highest entry ID among their answers.
	Gen  uint64
	Last ID
}

// StudyDays returns the study days of every answer in the review log, by
// the rollover hour in loc. It reads the whole log the first time.
func (c *Collection) StudyDays(rollover int, loc *time.Location) (StudyDays, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.loadReviewIndexLocked(); err != nil {
		return StudyDays{}, err
	}
	out := StudyDays{Gen: c.reviewGen}
	n, first, last := 0, int64(math.MaxInt64), int64(math.MinInt64)
	for _, cr := range c.reviews {
		for i, e := range cr.entries {
			if scheduler.Rating(e.Rating).Valid() {
				n++
				first, last = min(first, e.Time), max(last, e.Time)
				out.Last = max(out.Last, cr.ids[i])
			}
		}
	}
	if n == 0 {
		return out, nil
	}
	// starts are the beginnings of the study days from the first answer's
	// to the last's, and nums their day numbers. Each answer finds its day
	// from an estimate, which daylight saving changes move by at most one.
	var starts []int64
	var nums []int
	for t := scheduler.DayStart(time.UnixMilli(first).In(loc), rollover); t.UnixMilli() <= last; t = scheduler.NextDayStart(t, rollover) {
		if len(starts) == n { // a span of more days than answers: sort them
			return c.studyDaysSorted(out, rollover, loc), nil
		}
		starts = append(starts, t.UnixMilli())
		nums = append(nums, scheduler.DayNumber(t, rollover))
	}
	studied := make([]bool, len(starts))
	for _, cr := range c.reviews {
		for _, e := range cr.entries {
			if !scheduler.Rating(e.Rating).Valid() {
				continue
			}
			i := min(int((e.Time-starts[0])/msPerDay), len(starts)-1)
			for i+1 < len(starts) && starts[i+1] <= e.Time {
				i++
			}
			for starts[i] > e.Time {
				i--
			}
			studied[i] = true
		}
	}
	for i, ok := range studied {
		if ok {
			out.Days = append(out.Days, nums[i])
		}
	}
	return out, nil
}

const msPerDay = 24 * 60 * 60 * 1000

// studyDaysSorted fills in the Days of out by sorting the answer times, for
// histories spanning more days than they have answers. Callers hold a
// lock.
func (c *Collection) studyDaysSorted(out StudyDays, rollover int, loc *time.Location) StudyDays {
	var times []int64
	for _, cr := range c.reviews {
		for _, e := range cr.entries {
			if scheduler.Rating(e.Rating).Valid() {
				times = append(times, e.Time)
			}
		}
	}
	slices.Sort(times)
	// Only the first answer of each study day is mapped to its day.
	end := int64(math.MinInt64)
	for _, ms := range times {
		if ms < end {
			continue
		}
		t := time.UnixMilli(ms).In(loc)
		if day := scheduler.DayNumber(t, rollover); len(out.Days) == 0 || day != out.Days[len(out.Days)-1] {
			out.Days = append(out.Days, day)
		}
		end = scheduler.NextDayStart(t, rollover).UnixMilli()
	}
	return out
}

// RecentAnswer is an answer of the loaded review log months.
type RecentAnswer struct {
	ID         ID
	Time       time.Time
	DurationMs int64
}

// RecentAnswers returns the answers in the loaded review log months with
// an ID above after or a time from since on, and the review generation.
// The loaded months hold the current study day and every answer of this
// app session. Previews and manual entries are left out.
//
// The generation advances when answers are removed or added other than
// by answering (see noteReviewChange): while it stays, the study days of a
// StudyDays result change only by those of the answers above its Last.
func (c *Collection) RecentAnswers(after ID, since time.Time) ([]RecentAnswer, uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []RecentAnswer
	for _, e := range c.revlog {
		if !e.Rating.Valid() || e.Preview || e.ID <= after && e.Time.Before(since) {
			continue
		}
		out = append(out, RecentAnswer{ID: e.ID, Time: e.Time, DurationMs: max(e.DurationMs, 0)})
	}
	return out, c.reviewGen
}
