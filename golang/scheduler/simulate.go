package scheduler

import (
	"context"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"cardflip/golang/internal/yield"
)

// The workload simulator: a day-by-day Monte Carlo simulation of a deck
// with the FSRS model, as Anki's simulator (fsrs-rs simulation.rs) does,
// and the search for the desired retention with the lowest review time per
// memorised card (collection spec, "Workload simulation" and "Retention
// suggestion"). fsrs-rs is under the BSD-3-Clause license, see
// THIRD_PARTY_NOTICES.md.

// SimCard is a studied card as the simulation starts from it.
type SimCard struct {
	Stability, Difficulty float64
	// LastDay and DueDay are the study days of the last review and of the
	// due date, relative to today (day 0): LastDay is 0 or less, DueDay is
	// below 0 for an overdue card.
	LastDay, DueDay int
}

// SimCardOf returns the card with state cs as the simulation starts from it
// on the study day of now, with FSRS parameters w (see Params.Weights). ok
// is false for a card the simulation treats as new: a New card, or a
// learning card without a memory state.
func SimCardOf(cs CardState, now time.Time, rolloverHour int, w []float64) (c SimCard, ok bool) {
	if cs.State == New || (cs.Memory == nil && cs.State != Review) {
		return SimCard{}, false
	}
	today := DayNumber(now, rolloverHour)
	if cs.State == Review || cs.Interday {
		c.DueDay = DayNumber(cs.Due, rolloverHour) - today
	}
	if cs.LastReview.IsZero() {
		c.LastDay = min(c.DueDay-cs.ScheduledDays, 0)
	} else {
		c.LastDay = min(DayNumber(cs.LastReview, rolloverHour)-today, 0)
	}
	if cs.Memory == nil {
		// A review card without a memory state (it has no review history):
		// estimate it from its interval, as optimizing does.
		ww := weightsOf(w)
		st := ww.startState(float64(max(cs.ScheduledDays, 1)), DefaultHistoricalRetention)
		c.Stability, c.Difficulty = st.s, st.d
		return c, true
	}
	c.Stability = min(max(cs.Memory.Stability, sMin), sMax)
	c.Difficulty = min(max(cs.Memory.Difficulty, 1), 10)
	return c, true
}

// SimDeck is what a simulation starts from.
type SimDeck struct {
	Cards []SimCard
	// NewCards is the number of new cards, learned in turn.
	NewCards int
}

// SimBehaviour is how the user answers: answer times and rating
// probabilities, from the deck's review history (see BehaviourOf).
type SimBehaviour struct {
	// LearnSeconds is the time of a new card's first learning session, by
	// its first rating (index 0 = Again … 3 = Easy).
	LearnSeconds [4]float64
	// ReviewSeconds is the time of a review by rating; Again includes the
	// relearning steps after it.
	ReviewSeconds [4]float64
	// FirstRating is the probability of each first rating of a new card.
	FirstRating [4]float64
	// RecallRating is the probability of Hard, Good and Easy when a review
	// card is recalled.
	RecallRating [3]float64
}

// DefaultBehaviour is the behaviour fsrs-rs assumes without a history.
func DefaultBehaviour() SimBehaviour {
	return SimBehaviour{
		LearnSeconds:  [4]float64{33.79, 24.3, 13.68, 6.5},
		ReviewSeconds: [4]float64{23.0, 11.68, 7.33, 5.6},
		FirstRating:   [4]float64{0.24, 0.094, 0.495, 0.171},
		RecallRating:  [3]float64{0.224, 0.631, 0.145},
	}
}

const (
	// minBehaviourSamples is the fewest first answers (or recalled reviews)
	// the rating probabilities and times are taken from; with fewer, the
	// defaults are used for them.
	minBehaviourSamples = 100
	// minTimeSamples is the fewest timed answers of one rating its average
	// time is taken from.
	minTimeSamples = 10
)

// BehaviourOf derives the behaviour from review histories (each card's
// oldest first, without preview answers, with answer times), falling back
// to DefaultBehaviour where the history is short. A first learning session
// is a New card's first answer and the learning answers after it on the
// same study day; the relearning after a lapse is the relearning answers
// that follow it.
func BehaviourOf(histories [][]HistoryReview, rolloverHour int) SimBehaviour {
	var (
		firstN, recallN       [4]int
		learnSum, reviewSum   [4]float64
		learnN, reviewN       [4]int
		relearnSum            float64
		relearnN              int
		firstTotal, recallTot int
	)
	secs := func(h HistoryReview) float64 { return h.Duration.Seconds() }
	for _, hist := range histories {
		for i := 0; i < len(hist); i++ {
			h := hist[i]
			if !h.Rating.Valid() {
				continue
			}
			switch h.StateBefore {
			case New:
				firstN[h.Rating-1]++
				firstTotal++
				day := DayNumber(h.Time, rolloverHour)
				sum, timed := secs(h), h.Duration > 0
				for j := i + 1; j < len(hist) && hist[j].StateBefore == Learning &&
					DayNumber(hist[j].Time, rolloverHour) == day; j++ {
					if hist[j].Rating.Valid() {
						sum += secs(hist[j])
						timed = timed && hist[j].Duration > 0
					}
				}
				if timed {
					learnSum[h.Rating-1] += sum
					learnN[h.Rating-1]++
				}
			case Review:
				if h.Duration > 0 {
					reviewSum[h.Rating-1] += secs(h)
					reviewN[h.Rating-1]++
				}
				if h.Rating != Again {
					recallN[h.Rating-1]++
					recallTot++
					continue
				}
				sum, timed := 0.0, true
				for j := i + 1; j < len(hist) && hist[j].StateBefore == Relearning; j++ {
					if hist[j].Rating.Valid() {
						sum += secs(hist[j])
						timed = timed && hist[j].Duration > 0
					}
				}
				if timed {
					relearnSum += sum
					relearnN++
				}
			}
		}
	}
	b := DefaultBehaviour()
	if firstTotal >= minBehaviourSamples {
		for r := range b.FirstRating {
			b.FirstRating[r] = float64(firstN[r]) / float64(firstTotal)
		}
		for r := range b.LearnSeconds {
			if learnN[r] >= minTimeSamples {
				b.LearnSeconds[r] = learnSum[r] / float64(learnN[r])
			}
		}
	}
	if recallTot >= minBehaviourSamples {
		for r := range b.RecallRating {
			b.RecallRating[r] = float64(recallN[r+1]) / float64(recallTot)
		}
		for r := range b.ReviewSeconds {
			if reviewN[r] >= minTimeSamples {
				b.ReviewSeconds[r] = reviewSum[r] / float64(reviewN[r])
			}
		}
		if reviewN[0] >= minTimeSamples && relearnN >= minTimeSamples {
			b.ReviewSeconds[0] += relearnSum / float64(relearnN)
		}
	}
	return b
}

// SimConfig are the settings a simulation runs with.
type SimConfig struct {
	// Days is the number of days simulated, from today.
	Days int
	// NewPerDay and ReviewsPerDay are the daily limits; unless
	// NewIgnoresReviewLimit, new cards also count against the review
	// limit.
	NewPerDay, ReviewsPerDay int
	NewIgnoresReviewLimit    bool
	DesiredRetention         float64
	// MaximumInterval and MinimumRelearnInterval are as in Params.
	MaximumInterval, MinimumRelearnInterval int
	// LearningSteps and RelearningSteps are the numbers of (re)learning
	// steps: after the first answer, each is assumed answered Good on the
	// same day.
	LearningSteps, RelearningSteps int
	// Weights are the FSRS parameters (see Params.Weights).
	Weights   []float64
	Behaviour SimBehaviour
	// Seed makes the simulation reproducible.
	Seed uint64
}

// SimResult is a simulation's outcome, per simulated day.
type SimResult struct {
	// Reviews is the number of reviews of studied cards, New the number of
	// new cards learned.
	Reviews []int
	New     []int
	// Seconds is the time spent learning and reviewing.
	Seconds []float64
	// Memorized is the number of cards memorised at the end of the day: the
	// sum of every studied card's recall probability.
	Memorized []float64
}

// TotalSeconds is the time spent over the whole simulation.
func (r SimResult) TotalSeconds() float64 {
	t := 0.0
	for _, s := range r.Seconds {
		t += s
	}
	return t
}

// Simulate simulates deck day by day with cfg. Due cards are reviewed
// oldest first up to the review limit, the others stay due; a review is
// recalled with the FSRS recall probability and then rated as the
// behaviour says. It does not change deck. ctx stops it.
func Simulate(ctx context.Context, deck SimDeck, cfg SimConfig) (SimResult, error) {
	res, _, err := runSimulation(ctx, deck, cfg, true)
	return res, err
}

// simCard is a studied card during a simulation.
type simCard struct {
	s, d      float64
	last, due int
}

// runSimulation runs the simulation. With daily it fills in every day of
// Memorized; otherwise only the last. It also returns the memorised cards
// at the end.
func runSimulation(ctx context.Context, deck SimDeck, cfg SimConfig, daily bool) (SimResult, float64, error) {
	days := max(cfg.Days, 0)
	res := SimResult{Reviews: make([]int, days), New: make([]int, days), Seconds: make([]float64, days),
		Memorized: make([]float64, days)}
	w := weightsOf(cfg.Weights)
	b := cfg.Behaviour
	rng := rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9e3779b97f4a7c15))
	// A tick is a review or one card's recall probability, tens of
	// nanoseconds in WebAssembly: yielding every 200,000 keeps the UI
	// smooth without browsers' timer clamping dominating the run time.
	tick := yield.Every(200_000)
	maxIvl := max(cfg.MaximumInterval, 1)
	interval := func(s float64) int {
		ivl := s / fsrsFactor * (math.Pow(cfg.DesiredRetention, 1/fsrsDecay) - 1)
		return min(max(int(math.Round(ivl)), 1), maxIvl)
	}
	pick := func(probs []float64) int {
		x := rng.Float64()
		for i, p := range probs {
			if x -= p; x < 0 {
				return i
			}
		}
		return len(probs) - 1
	}

	cards := make([]simCard, 0, len(deck.Cards)+max(deck.NewCards, 0))
	for _, c := range deck.Cards {
		cards = append(cards, simCard{s: c.Stability, d: c.Difficulty, last: c.LastDay, due: c.DueDay})
	}
	// buckets[d] are the cards falling due on day d; carry the due cards
	// left over by the review limit, oldest first.
	buckets := make([][]int32, days)
	var carry []int32
	order := make([]int32, len(cards))
	for i := range order {
		order[i] = int32(i)
	}
	slices.SortStableFunc(order, func(a, b int32) int { return cards[a].due - cards[b].due })
	for _, i := range order {
		switch due := cards[i].due; {
		case due <= 0:
			carry = append(carry, i)
		case due < days:
			buckets[due] = append(buckets[due], i)
		}
	}
	schedule := func(i int32) {
		if due := cards[i].due; due < days {
			buckets[due] = append(buckets[due], i)
		}
	}
	reviewLimit := max(cfg.ReviewsPerDay, 0)
	newLeft := max(deck.NewCards, 0)

	for day := 0; day < days; day++ {
		if err := ctx.Err(); err != nil {
			return SimResult{}, 0, err
		}
		queue := append(carry, buckets[day]...)
		buckets[day] = nil
		n := min(len(queue), reviewLimit)
		for _, i := range queue[:n] {
			c := &cards[i]
			t := float64(day - c.last)
			rating := Again
			if rng.Float64() < forgettingCurve(t, c.s) {
				rating = Rating(pick(b.RecallRating[:]) + 2)
			}
			st := w.step(memState{s: c.s, d: c.d}, t, int(rating), false)
			res.Seconds[day] += b.ReviewSeconds[rating-1]
			ivl := 0
			if rating == Again {
				for range cfg.RelearningSteps {
					st = w.step(st, 0, int(Good), false)
				}
				ivl = max(interval(st.s), min(max(cfg.MinimumRelearnInterval, 1), maxIvl))
			} else {
				ivl = interval(st.s)
			}
			c.s, c.d, c.last, c.due = st.s, st.d, day, day+ivl
			schedule(i)
			tick()
		}
		res.Reviews[day] = n
		carry = slices.Clone(queue[n:])

		newLimit := max(cfg.NewPerDay, 0)
		if !cfg.NewIgnoresReviewLimit {
			newLimit = min(newLimit, reviewLimit-n)
		}
		learn := min(newLimit, newLeft)
		for range learn {
			r := pick(b.FirstRating[:]) + 1
			st := w.first(r, false)
			extra := cfg.LearningSteps
			switch Rating(r) {
			case Good:
				extra = max(extra-1, 0)
			case Easy:
				extra = 0
			}
			for range extra {
				st = w.step(st, 0, int(Good), false)
			}
			res.Seconds[day] += b.LearnSeconds[r-1]
			cards = append(cards, simCard{s: st.s, d: st.d, last: day, due: day + interval(st.s)})
			schedule(int32(len(cards) - 1))
			tick()
		}
		res.New[day] = learn
		newLeft -= learn

		if daily || day == days-1 {
			m := 0.0
			for i := range cards {
				m += forgettingCurve(float64(day-cards[i].last), cards[i].s)
				tick()
			}
			res.Memorized[day] = m
		}
	}
	memorized := 0.0
	if days > 0 {
		memorized = res.Memorized[days-1]
	}
	return res, memorized, nil
}

// Retention search range and step of SuggestRetention.
const (
	MinSuggestedRetention  = 0.70
	MaxSuggestedRetention  = 0.99
	suggestedRetentionStep = 0.01
)

// RetentionPoint is the simulated cost at one desired retention.
type RetentionPoint struct {
	Retention float64
	// SecondsPerCard is the review time per memorised card: the total time
	// over the simulation divided by the cards memorised at its end.
	SecondsPerCard float64
}

// RetentionSuggestion is the result of SuggestRetention.
type RetentionSuggestion struct {
	// Retention is the retention with the lowest cost.
	Retention float64
	// Points is the cost at every retention tried, in increasing
	// retention; retentions that memorise nothing are left out.
	Points []RetentionPoint
}

// NothingToSimulateError is returned by SuggestRetention when no retention
// memorises any card: the deck has no cards, or nothing is simulated.
type NothingToSimulateError struct{}

func (NothingToSimulateError) Error() string { return "The deck has no cards to simulate." }

// SuggestRetention simulates deck at each desired retention from
// MinSuggestedRetention to MaxSuggestedRetention, all with the same seed,
// and suggests the one with the lowest review time per memorised card. The
// rest of cfg is used as it is. progress gets the number of simulations
// done. ctx stops it.
func SuggestRetention(ctx context.Context, deck SimDeck, cfg SimConfig, progress func(done, total int)) (RetentionSuggestion, error) {
	steps := int(math.Round((MaxSuggestedRetention-MinSuggestedRetention)/suggestedRetentionStep)) + 1
	out := RetentionSuggestion{Points: []RetentionPoint{}}
	best := math.Inf(1)
	for i := range steps {
		progress(i, steps)
		cfg.DesiredRetention = math.Round((MinSuggestedRetention+float64(i)*suggestedRetentionStep)*100) / 100
		res, memorized, err := runSimulation(ctx, deck, cfg, false)
		if err != nil {
			return RetentionSuggestion{}, err
		}
		if memorized <= 0 {
			continue
		}
		cost := res.TotalSeconds() / memorized
		out.Points = append(out.Points, RetentionPoint{Retention: cfg.DesiredRetention, SecondsPerCard: cost})
		if cost < best {
			best, out.Retention = cost, cfg.DesiredRetention
		}
	}
	progress(steps, steps)
	if len(out.Points) == 0 {
		return RetentionSuggestion{}, NothingToSimulateError{}
	}
	return out, nil
}
