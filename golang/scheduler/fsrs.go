package scheduler

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// Params are the per-deck scheduling settings.
type Params struct {
	// DesiredRetention is the target recall probability at the due time.
	DesiredRetention float64
	// LearningSteps and RelearningSteps are the steps before a card
	// (re)graduates to Review. An empty list leaves the delays to FSRS's
	// short-term scheduling.
	LearningSteps   []time.Duration
	RelearningSteps []time.Duration
	// MaximumInterval caps review intervals, in days.
	MaximumInterval int
	// MinimumRelearnInterval is the shortest interval, in days, a card
	// graduating from Relearning gets.
	MinimumRelearnInterval int
	// RolloverHour is the hour the study day starts. A step ending on a
	// later study day makes the card due on that day (interday learning).
	RolloverHour int
	// Weights are the FSRS parameters; anything but NumWeights values
	// (empty, or from another model version) means the defaults.
	Weights []float64
	// EasyDays is the Easy Days level of each weekday, Monday first:
	// EasyNormal, EasyReduced or EasyMinimum. With every day Normal, due
	// days are fuzzed uniformly.
	EasyDays [7]int
	// DueLoad returns how many reviews are due on a study day (see
	// DayNumber). Easy Days prefers days with fewer; nil counts none.
	DueLoad func(day int) int
}

// Easy Days levels of a weekday (Params.EasyDays).
const (
	EasyNormal = iota
	EasyReduced
	EasyMinimum
)

// DefaultParams returns the defaults: 1m/10m learning, 10m relearning,
// intervals from 1 day after relearning up to 36,500 days, rollover at 04:00.
func DefaultParams(retention float64) Params {
	return Params{
		DesiredRetention:       retention,
		LearningSteps:          []time.Duration{time.Minute, 10 * time.Minute},
		RelearningSteps:        []time.Duration{10 * time.Minute},
		MaximumInterval:        36500,
		MinimumRelearnInterval: 1,
		RolloverHour:           4,
	}
}

// Outcome is the result of one rating.
type Outcome struct {
	State CardState
	// Interval is the time until the card is due again.
	Interval time.Duration
}

// Preview returns the outcome of each rating (index 0 = Again … 3 = Easy)
// without changing anything. seed identifies the card; together with its
// review count it seeds the interval fuzz, so Answer gives exactly the
// previewed outcome.
func Preview(cs CardState, now time.Time, p Params, seed int64) [4]Outcome {
	var out [4]Outcome
	for r := Again; r <= Easy; r++ {
		out[r-1] = answer(cs, r, now, p, seed)
	}
	// Fuzz must not reorder ratings that both graduate: Hard ≤ Good ≤ Easy.
	for i := 1; i < 4; i++ {
		prev := out[i-1].State
		if out[i].State.State == Review && prev.State == Review && out[i].State.ScheduledDays < prev.ScheduledDays {
			out[i] = withDays(out[i], prev.ScheduledDays, now)
		}
	}
	return out
}

// Answer returns the new state after rating r.
func Answer(cs CardState, r Rating, now time.Time, p Params, seed int64) Outcome {
	return Preview(cs, now, p, seed)[r-1]
}

func withDays(o Outcome, days int, now time.Time) Outcome {
	o.State.ScheduledDays = days
	o.State.Due = now.AddDate(0, 0, days)
	o.Interval = o.State.Due.Sub(now)
	return o
}

var lib = fsrs.DefaultParam()

// memoryAfter asks go-fsrs for the memory state after rating r, with FSRS
// parameters w (see Params.Weights).
func memoryAfter(cs CardState, r Rating, now time.Time, w []float64) Memory {
	next := fsrsAfter(cs, r, now, w)
	return Memory{Stability: next.Stability, Difficulty: next.Difficulty}
}

// FirstGoodMemory is the memory state a New card gets from a first Good
// rating at now, with FSRS parameters w (see Params.Weights).
func FirstGoodMemory(now time.Time, w []float64) Memory {
	return memoryAfter(CardState{}, Good, now, w)
}

// fsrsAfter asks go-fsrs for the card after rating r. It uses the library's
// short-term mode, so same-day reviews during (re)learning use the
// short-term stability formula and later reviews the long-term one, and
// (re)learning cards get minute-level delays until FSRS graduates them.
func fsrsAfter(cs CardState, r Rating, now time.Time, w []float64) fsrs.Card {
	card := fsrs.NewCard()
	card.State = fsrs.State(cs.State)
	card.Reps = uint64(cs.Reps)
	card.Lapses = uint64(cs.Lapses)
	card.LastReview = cs.LastReview
	if cs.Memory != nil {
		card.Stability = cs.Memory.Stability
		card.Difficulty = cs.Memory.Difficulty
	}
	if cs.State != New && cs.Memory == nil {
		// A card in learning without a memory state (e.g. imported): start
		// from the state a first Good would give.
		card.State = fsrs.New
	}
	params := lib
	params.W = fsrs.Weights(weightsOf(w))
	params.EnableShortTerm = true
	params.EnableFuzz = false
	return params.NewBasicScheduler(card, now).Review(fsrs.Rating(r)).Card
}

// Retrievability is the probability of recalling a card at now, from the
// FSRS forgetting curve of its memory state and last review. ok is false
// for cards without a memory state or review.
func Retrievability(cs CardState, now time.Time) (r float64, ok bool) {
	if cs.Memory == nil || cs.Memory.Stability <= 0 || cs.LastReview.IsZero() {
		return 0, false
	}
	elapsed := math.Max(now.Sub(cs.LastReview).Hours()/24, 0)
	return math.Pow(1+lib.Factor*elapsed/cs.Memory.Stability, lib.Decay), true
}

// intervalDays is the FSRS interval at which recall probability falls to the
// desired retention, for stability s. It is not capped yet: graduate caps
// it after fuzz, so a capped interval is exactly the maximum.
func intervalDays(s float64, p Params) float64 {
	ivl := s / lib.Factor * (math.Pow(p.DesiredRetention, 1/lib.Decay) - 1)
	return math.Max(math.Round(ivl), 1)
}

func answer(cs CardState, r Rating, now time.Time, p Params, seed int64) Outcome {
	short := fsrsAfter(cs, r, now, p.Weights)
	mem := Memory{Stability: short.Stability, Difficulty: short.Difficulty}
	next := cs
	next.Memory = &mem
	next.Reps = cs.Reps + 1
	next.LastReview = now

	// step schedules a (re)learning step of delay d. A step ending on a
	// later study day is due from the start of that day instead.
	step := func(state State, i int, d time.Duration) Outcome {
		next.State = state
		next.Step = i
		next.ScheduledDays = 0
		next.Interday = false
		next.Due = now.Add(stepFuzz(d, cs, seed))
		if next.Due.Before(NextDayStart(now, p.RolloverHour)) {
			return Outcome{State: next, Interval: next.Due.Sub(now)}
		}
		next.Interday = true
		next.Due = DayStart(next.Due, p.RolloverHour)
		next.ScheduledDays = DayNumber(next.Due, p.RolloverHour) - DayNumber(now, p.RolloverHour)
		return Outcome{State: next, Interval: time.Duration(next.ScheduledDays) * 24 * time.Hour}
	}
	// graduate schedules a review interval: at least the minimum after
	// relearning, fuzzed, and at most the maximum interval.
	graduate := func(days float64) Outcome {
		floor := 1.0
		if cs.State == Relearning {
			floor = float64(max(p.MinimumRelearnInterval, 1))
			days = math.Max(days, floor)
		}
		days = p.fuzzDays(days, floor, cs, now, seed)
		next.State = Review
		next.Step = 0
		next.Interday = false
		next.ScheduledDays = int(days)
		next.Due = now.AddDate(0, 0, int(days))
		return Outcome{State: next, Interval: next.Due.Sub(now)}
	}

	switch cs.State {
	case New, Learning, Relearning:
		steps := p.LearningSteps
		state := Learning
		if cs.State == Relearning {
			steps = p.RelearningSteps
			state = Relearning
		}
		i := cs.Step
		if cs.State == New {
			i = 0
		}
		if i >= len(steps) {
			i = len(steps) - 1
		}
		goodDays := intervalDays(memoryAfter(cs, Good, now, p.Weights).Stability, p)
		switch {
		case r == Easy:
			return graduate(math.Max(intervalDays(mem.Stability, p), goodDays+1))
		case len(steps) == 0:
			// No steps: FSRS's short-term delays until it graduates the card.
			if short.State == fsrs.Review {
				return graduate(intervalDays(mem.Stability, p))
			}
			return step(state, 0, short.Due.Sub(now))
		case r == Again:
			return step(state, 0, steps[0])
		case r == Hard:
			return step(state, i, hardDelay(steps, i))
		}
		// Good: on to the next step. A New card's first Good completes the
		// first step, so it moves to step 1.
		nextStep := i + 1
		if nextStep < len(steps) {
			return step(state, nextStep, steps[nextStep])
		}
		return graduate(intervalDays(mem.Stability, p))

	default: // Review
		if r == Again {
			next.Lapses = cs.Lapses + 1
			if len(p.RelearningSteps) == 0 {
				return step(Relearning, 0, short.Due.Sub(now))
			}
			return step(Relearning, 0, p.RelearningSteps[0])
		}
		days := intervalDays(mem.Stability, p)
		hard := intervalDays(memoryAfter(cs, Hard, now, p.Weights).Stability, p)
		good := math.Max(intervalDays(memoryAfter(cs, Good, now, p.Weights).Stability, p), hard+1)
		switch r {
		case Hard:
			days = math.Min(hard, good)
		case Good:
			days = good
		case Easy:
			days = math.Max(days, good+1)
		}
		return graduate(days)
	}
}

// hardDelay repeats the current step: on the first step with the average of
// the first two delays, on a last step with 1.5 times its delay (at most one
// day longer), as Anki does.
func hardDelay(steps []time.Duration, i int) time.Duration {
	if i == 0 && len(steps) > 1 {
		return (steps[0] + steps[1]) / 2
	}
	if i == len(steps)-1 {
		d := steps[i] * 3 / 2
		if d > steps[i]+24*time.Hour {
			d = steps[i] + 24*time.Hour
		}
		return d
	}
	return steps[i]
}

// fuzz spreads review intervals of 3 days or more over a small range, so
// cards added together do not stay together. The result stays at or above
// floor. It is deterministic: seeded by the card and its review count, so a
// preview equals the result.
func fuzz(days, floor float64, cs CardState, seed int64) float64 {
	lo, hi, ok := fuzzRange(days, floor)
	if !ok {
		return days
	}
	return math.Floor(fuzzFactor(cs, seed)*(hi-lo+1)) + lo
}

// fuzzRange is the range of intervals fuzz chooses from; ok is false for
// intervals too short to fuzz.
func fuzzRange(days, floor float64) (lo, hi float64, ok bool) {
	if days < 2.5 {
		return days, days, false
	}
	delta := 1.0
	for _, r := range []struct{ start, end, factor float64 }{
		{2.5, 7, 0.15}, {7, 20, 0.1}, {20, math.Inf(1), 0.05},
	} {
		delta += r.factor * math.Max(math.Min(days, r.end)-r.start, 0)
	}
	lo = math.Max(math.Max(2, floor), math.Round(days-delta))
	hi = math.Max(math.Round(days+delta), lo)
	return lo, hi, true
}

// fuzzDays is the fuzzed review interval for an answer at now, at most the
// maximum interval. With Easy Days set, the day is chosen within the fuzz
// range by easyDay instead of uniformly.
func (p Params) fuzzDays(days, floor float64, cs CardState, now time.Time, seed int64) float64 {
	maxDays := float64(max(p.MaximumInterval, 1))
	lo, hi, ok := fuzzRange(days, floor)
	if !ok || p.EasyDays == [7]int{} {
		return math.Min(fuzz(days, floor, cs, seed), maxDays)
	}
	return float64(p.easyDay(int(math.Min(lo, maxDays)), int(math.Min(hi, maxDays)), now, fuzzFactor(cs, seed)))
}

// reducedWeight is how much less likely a Reduced day is chosen than a
// Normal day with as many reviews due, so that Reduced days settle at about
// half the reviews of Normal days.
const reducedWeight = 0.5

// easyDay chooses an interval in [lo, hi] days from now for Easy Days.
// Minimum days are left out unless the range has no other day; the others
// are weighted by their level and inversely by the reviews already due on
// them, and u in [0,1) picks one by those weights.
func (p Params) easyDay(lo, hi int, now time.Time, u float64) int {
	type candidate struct {
		days   int
		weight float64
	}
	var cands []candidate
	onlyMinimum := true
	for d := lo; d <= hi; d++ {
		if p.easyLevel(now.AddDate(0, 0, d)) != EasyMinimum {
			onlyMinimum = false
		}
	}
	total := 0.0
	for d := lo; d <= hi; d++ {
		due := now.AddDate(0, 0, d)
		w := 1.0
		switch p.easyLevel(due) {
		case EasyReduced:
			w = reducedWeight
		case EasyMinimum:
			if !onlyMinimum {
				continue
			}
		}
		if p.DueLoad != nil {
			w /= float64(max(p.DueLoad(DayNumber(due, p.RolloverHour)), 0) + 1)
		}
		cands = append(cands, candidate{d, w})
		total += w
	}
	x := u * total
	for _, c := range cands {
		if x < c.weight {
			return c.days
		}
		x -= c.weight
	}
	return cands[len(cands)-1].days
}

// easyLevel is the Easy Days level of the study day due falls in.
func (p Params) easyLevel(due time.Time) int {
	wd := DayStart(due, p.RolloverHour).Weekday()
	return p.EasyDays[(int(wd)+6)%7]
}

// stepFuzz adds a random extra delay to a learning step of a minute or
// more: up to 25 % of the step, at most 5 minutes, so cards learned together
// do not keep returning in the same order. Seeded like fuzz.
func stepFuzz(d time.Duration, cs CardState, seed int64) time.Duration {
	if d < time.Minute {
		return d
	}
	extra := min(d/4, 5*time.Minute)
	return d + time.Duration(fuzzFactor(cs, seed)*float64(extra)).Truncate(time.Second)
}

// fuzzFactor is a value in [0,1) determined by the card and its review
// count.
func fuzzFactor(cs CardState, seed int64) float64 {
	h := fnv.New64a()
	var buf [16]byte
	binary.LittleEndian.PutUint64(buf[:8], uint64(seed))
	binary.LittleEndian.PutUint64(buf[8:], uint64(cs.Reps))
	h.Write(buf[:])
	return float64(h.Sum64()>>11) / float64(1<<53)
}
