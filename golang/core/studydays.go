package core

import (
	"cmp"
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

// MatureInterval is the interval, in days, from which a card is mature:
// in long-term memory.
const MatureInterval = 21

// StudyDays are the study days of the whole review log.
type StudyDays struct {
	// Days are the sorted study days (scheduler.DayNumber) with at least
	// one answer. Previews and manual entries do not count.
	Days []int
	// Answers[i] is the number of answers on Days[i], Ms[i] the
	// milliseconds spent on them.
	Answers []int
	Ms      []int64
	// Mature are the sorted study days on which a card first got an
	// interval of MatureInterval days or more by an answer, one per card.
	Mature []int
	// Gen is the review generation they were read at (see RecentAnswers),
	// Last the highest entry ID among their answers.
	Gen  uint64
	Last ID
}

// StudyDays returns the study days of every answer in the review log, by
// the rollover hour in loc, with their answers and the days cards became
// mature. It reads the whole log the first time.
func (c *Collection) StudyDays(rollover int, loc *time.Location) (StudyDays, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.loadReviewIndexLocked(); err != nil {
		return StudyDays{}, err
	}
	out := StudyDays{Gen: c.reviewGen}
	n, first, last := 0, int64(math.MaxInt64), int64(math.MinInt64)
	var mature []int64
	for _, cr := range c.reviews {
		for i, e := range cr.entries {
			if scheduler.Rating(e.Rating).Valid() {
				n++
				first, last = min(first, e.Time), max(last, e.Time)
				out.Last = max(out.Last, cr.ids[i])
			}
		}
		if i := cr.mature(); i >= 0 {
			mature = append(mature, cr.entries[i].Time)
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
			return c.studyDaysSorted(out, mature, rollover, loc), nil
		}
		starts = append(starts, t.UnixMilli())
		nums = append(nums, scheduler.DayNumber(t, rollover))
	}
	dayOf := func(ms int64) int {
		i := min(max(int((ms-starts[0])/msPerDay), 0), len(starts)-1)
		for i+1 < len(starts) && starts[i+1] <= ms {
			i++
		}
		for starts[i] > ms {
			i--
		}
		return i
	}
	answers := make([]int, len(starts))
	spent := make([]int64, len(starts))
	for _, cr := range c.reviews {
		for i, e := range cr.entries {
			if !scheduler.Rating(e.Rating).Valid() {
				continue
			}
			d := dayOf(e.Time)
			answers[d]++
			spent[d] += int64(cr.progress[i].ms)
		}
	}
	for i, k := range answers {
		if k > 0 {
			out.Days = append(out.Days, nums[i])
			out.Answers = append(out.Answers, k)
			out.Ms = append(out.Ms, spent[i])
		}
	}
	for _, ms := range mature {
		out.Mature = append(out.Mature, nums[dayOf(ms)])
	}
	slices.Sort(out.Mature)
	return out, nil
}

const msPerDay = 24 * 60 * 60 * 1000

// studyDaysSorted fills in the days of out by sorting the answer times, for
// histories spanning more days than they have answers; mature are the
// times cards became mature. Callers hold a lock.
func (c *Collection) studyDaysSorted(out StudyDays, mature []int64, rollover int, loc *time.Location) StudyDays {
	type answer struct{ time, ms int64 }
	var all []answer
	for _, cr := range c.reviews {
		for i, e := range cr.entries {
			if scheduler.Rating(e.Rating).Valid() {
				all = append(all, answer{e.Time, int64(cr.progress[i].ms)})
			}
		}
	}
	slices.SortFunc(all, func(a, b answer) int { return cmp.Compare(a.time, b.time) })
	// Only the first answer of each study day is mapped to its day.
	end := int64(math.MinInt64)
	for _, a := range all {
		if a.time >= end {
			t := time.UnixMilli(a.time).In(loc)
			if day := scheduler.DayNumber(t, rollover); len(out.Days) == 0 || day != out.Days[len(out.Days)-1] {
				out.Days = append(out.Days, day)
				out.Answers = append(out.Answers, 0)
				out.Ms = append(out.Ms, 0)
			}
			end = scheduler.NextDayStart(t, rollover).UnixMilli()
		}
		out.Answers[len(out.Answers)-1]++
		out.Ms[len(out.Ms)-1] += a.ms
	}
	for _, ms := range mature {
		out.Mature = append(out.Mature, scheduler.DayNumber(time.UnixMilli(ms).In(loc), rollover))
	}
	slices.Sort(out.Mature)
	return out
}

// RecentAnswer is an answer of the loaded review log months.
type RecentAnswer struct {
	ID         ID
	Time       time.Time
	DurationMs int64
	// FirstMature reports the card's first answer that gave it an interval
	// of MatureInterval days or more (see StudyDays.Mature).
	FirstMature bool
}

// RecentAnswers returns the answers in the loaded review log months with
// an ID above after or a time from since on, and the review generation.
// The loaded months hold the current study day and every answer of this
// app session. Previews and manual entries are left out.
//
// The generation advances when answers are removed or added other than
// by answering (see noteReviewChange): while it stays, the study days of a
// StudyDays result change only by those of the answers above its Last.
// FirstMature needs the whole log, which StudyDays reads.
func (c *Collection) RecentAnswers(after ID, since time.Time) ([]RecentAnswer, uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []RecentAnswer
	for _, e := range c.revlog {
		if !e.Rating.Valid() || e.Preview || e.ID <= after && e.Time.Before(since) {
			continue
		}
		a := RecentAnswer{ID: e.ID, Time: e.Time, DurationMs: max(e.DurationMs, 0)}
		if cr := c.reviews[e.CardID]; e.IvlAfter >= MatureInterval && cr != nil {
			i := cr.mature()
			a.FirstMature = i >= 0 && cr.ids[i] == e.ID
		}
		out = append(out, a)
	}
	return out, c.reviewGen
}
