package scheduler

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"time"
)

// The FSRS optimizer: fits a deck's parameters to its review history, as
// Anki does with fsrs-rs (scheduling spec, "Optimize parameters").

// MinOptimizeReviews is the fewest reviews optimizing works with.
const MinOptimizeReviews = 400

// DefaultHistoricalRetention is the recall assumed for the missing part of
// a card's history.
const DefaultHistoricalRetention = 0.90

// HistoryReview is one review log entry of a card, for optimizing.
type HistoryReview struct {
	Time time.Time
	// Rating is 0 for a manual entry: a reset to New when IvlAfter is 0,
	// otherwise a set due date, which FSRS ignores.
	Rating      Rating
	StateBefore State
	// IvlBefore and IvlAfter are the intervals before and after, in days.
	IvlBefore, IvlAfter int
	// Duration is the answer time; 0 when unknown.
	Duration time.Duration
}

// HistoryOptions say how review histories become sequences.
type HistoryOptions struct {
	RolloverHour int
	// IgnoreBefore drops the entries before it, when set.
	IgnoreBefore time.Time
	// HistoricalRetention is the recall assumed for the reviews before a
	// card's known history (see Sequence.Incomplete).
	HistoricalRetention float64
}

// SeqReview is one review of a sequence.
type SeqReview struct {
	Rating Rating
	// DeltaT is the number of study days since the previous review: 0 for
	// a card's first review and for further reviews on the same day.
	DeltaT int
	at     int64
}

// Sequence is a card's reviews as FSRS sees them.
type Sequence struct {
	Reviews []SeqReview
	// Incomplete marks a history that does not start with the card's first
	// review as a New card (it is older than the review log, or before
	// IgnoreBefore). Its memory state before Reviews is estimated from
	// StartInterval, the interval in days before the first one, at the
	// historical retention, and it is not used for training.
	Incomplete    bool
	StartInterval float64
	retention     float64
	// LastReview is the time of the last review.
	LastReview time.Time
}

// NewSequence turns a card's review log, oldest first and without preview
// answers, into its review sequence: from its last first review as a New
// card (a reset starts over), or, when there is none, from its first entry
// on or after o.IgnoreBefore. Manual entries other than resets are left
// out, and reviews are counted in study days.
func NewSequence(history []HistoryReview, o HistoryOptions) Sequence {
	var seg []HistoryReview
	var before *HistoryReview // the last answer before seg, when known
	for i := range history {
		h := &history[i]
		switch {
		case h.Rating == 0 && h.IvlAfter == 0:
			seg, before = nil, nil
			continue
		case !h.Rating.Valid():
			continue
		case !o.IgnoreBefore.IsZero() && h.Time.Before(o.IgnoreBefore):
			before = h
			continue
		}
		seg = append(seg, *h)
	}
	seq := Sequence{retention: o.HistoricalRetention}
	if seq.retention <= 0 || seq.retention >= 1 {
		seq.retention = DefaultHistoricalRetention
	}
	start := -1
	for i, h := range seg {
		if h.StateBefore == New {
			start = i
		}
	}
	if start >= 0 {
		seg = seg[start:]
	} else if len(seg) > 0 {
		seq.Incomplete = true
		seq.StartInterval = float64(seg[0].IvlBefore)
	}
	day := func(t time.Time) int { return DayNumber(t, o.RolloverHour) }
	for i, h := range seg {
		dt := 0
		switch {
		case i > 0:
			dt = max(day(h.Time)-day(seg[i-1].Time), 0)
		case seq.Incomplete && before != nil:
			dt = max(day(h.Time)-day(before.Time), 0)
		case seq.Incomplete:
			dt = h.IvlBefore
		}
		seq.Reviews = append(seq.Reviews, SeqReview{Rating: h.Rating, DeltaT: dt, at: h.Time.UnixMilli()})
		seq.LastReview = h.Time
	}
	return seq
}

// Memory returns the memory state after the sequence's reviews with
// parameters w (the defaults when w is not a full set). ok is false when it
// has no reviews.
func (s Sequence) Memory(w []float64) (m Memory, ok bool) {
	if len(s.Reviews) == 0 {
		return Memory{}, false
	}
	ws := weightsOf(w)
	var start *memState
	if s.Incomplete {
		st := ws.startState(s.StartInterval, s.retention)
		start = &st
	}
	st := ws.replay(s.Reviews, start, false)
	return Memory{Stability: st.s, Difficulty: st.d}, true
}

// Reschedule returns review card cs with its memory state replayed from
// seq with p.Weights (kept when seq has no reviews) and due at the interval
// for p.DesiredRetention from its last review, with the fuzz and limits of
// an answer. seed identifies the card, as for Answer.
func Reschedule(cs CardState, seq Sequence, p Params, seed int64) CardState {
	if m, ok := seq.Memory(p.Weights); ok {
		cs.Memory = &m
		if cs.LastReview.IsZero() {
			cs.LastReview = seq.LastReview
		}
	}
	if cs.Memory == nil || cs.LastReview.IsZero() {
		return cs
	}
	days := fuzz(intervalDays(cs.Memory.Stability, p), 1, cs, seed)
	cs.ScheduledDays = int(min(days, float64(max(p.MaximumInterval, 1))))
	cs.Due = cs.LastReview.AddDate(0, 0, cs.ScheduledDays)
	return cs
}

// trainingItems returns one item per review that can be predicted: each
// review of a complete sequence that is the first on its study day, except
// the card's first review. They are ordered by review time.
func trainingItems(seqs []Sequence) []item {
	var items []item
	for _, s := range seqs {
		if s.Incomplete {
			continue
		}
		for i := 1; i < len(s.Reviews); i++ {
			if s.Reviews[i].DeltaT > 0 {
				items = append(items, item{reviews: s.Reviews[: i+1 : i+1], weight: 1, time: s.Reviews[i].at})
			}
		}
	}
	slices.SortStableFunc(items, func(a, b item) int { return cmp.Compare(a.time, b.time) })
	return items
}

// TrainingReviews is the number of reviews of seqs that optimizing can
// learn from.
func TrainingReviews(seqs []Sequence) int { return len(trainingItems(seqs)) }

// NotEnoughReviewsError is returned when there are fewer than
// MinOptimizeReviews reviews to learn from.
type NotEnoughReviewsError struct{ Reviews int }

func (e *NotEnoughReviewsError) Error() string {
	return fmt.Sprintf("Optimizing needs at least %d reviews, but only %d can be used.", MinOptimizeReviews, e.Reviews)
}

// OptimizeOptions are the optimizer's settings.
type OptimizeOptions struct {
	// Current are the parameters in use; empty means the defaults.
	Current []float64
	// HealthCheck also checks that the fitted parameters predict later
	// reviews from earlier ones (see HealthCheck).
	HealthCheck bool
}

// OptimizeResult is what optimizing found.
type OptimizeResult struct {
	// Weights are the fitted parameters, rounded to four decimals.
	Weights []float64
	// Reviews is the number of reviews learned from.
	Reviews int
	// LogLoss and RMSE measure how well Weights predict the reviews
	// (lower is better); CurrentLogLoss and CurrentRMSE the same for the
	// current parameters.
	LogLoss, RMSE               float64
	CurrentLogLoss, CurrentRMSE float64
	// Better reports whether Weights predict the reviews better than the
	// current parameters.
	Better bool
	// Health is the health check, when asked for.
	Health *HealthCheck
}

// HealthCheck tells whether fitted parameters are reliable. The history is
// split by time, and each later part is predicted with parameters fitted to
// the parts before it. Those predictions must be at least as good as the
// default parameters', and close to the predictions of the parameters
// fitted to everything: parameters that keep changing as reviews are added
// come from too few or inconsistent reviews.
type HealthCheck struct {
	Passed bool
	// LogLoss and DefaultLogLoss are the losses of the predictions and of
	// the default parameters on the later parts.
	LogLoss, DefaultLogLoss float64
	// Drift is the mean difference between the predicted recall and that
	// of the parameters fitted to everything (0.03 is three percentage
	// points).
	Drift float64
}

// maxHealthDrift is the largest Drift a reliable fit has.
const maxHealthDrift = 0.03

// Use reports whether the fitted parameters should replace the current
// ones: they are better and pass the health check, when there was one.
func (r OptimizeResult) Use() bool { return r.Better && (r.Health == nil || r.Health.Passed) }

// Optimize fits FSRS parameters to the reviews of seqs. It reports
// progress per stage ("training", then "checking" for the health check),
// stops with ctx's error when ctx is done, and returns a
// *NotEnoughReviewsError when there are too few reviews.
func Optimize(ctx context.Context, seqs []Sequence, o OptimizeOptions, progress func(stage string, done, total int)) (OptimizeResult, error) {
	if progress == nil {
		progress = func(string, int, int) {}
	}
	items := trainingItems(seqs)
	if len(items) < MinOptimizeReviews {
		return OptimizeResult{}, &NotEnoughReviewsError{Reviews: len(items)}
	}
	w, err := fit(ctx, items, func(done, total int) { progress("training", done, total) })
	if err == errNoData {
		return OptimizeResult{}, &NotEnoughReviewsError{Reviews: len(items)}
	}
	if err != nil {
		return OptimizeResult{}, err
	}
	// Four decimals, as Anki stores them.
	for i := range w {
		w[i] = math.Round(w[i]*1e4) / 1e4
	}
	w.clip()
	cur := weightsOf(o.Current)
	weighted := slices.Clone(items)
	recencyWeight(weighted)
	fitted, current := evaluate(w, weighted), evaluate(cur, weighted)
	res := OptimizeResult{
		Weights: w[:], Reviews: len(items),
		LogLoss: fitted.LogLoss, RMSE: fitted.RMSE, CurrentLogLoss: current.LogLoss, CurrentRMSE: current.RMSE,
		Better: fitted.LogLoss < current.LogLoss,
	}
	if o.HealthCheck {
		h, err := healthCheck(ctx, items, w, func(done, total int) { progress("checking", done, total) })
		if err != nil {
			return OptimizeResult{}, err
		}
		res.Health = &h
	}
	return res, nil
}

// healthSplits is the number of later parts the health check predicts.
const healthSplits = 5

// healthCheck splits items, ordered by time, into healthSplits+1 parts and
// predicts each part but the first with parameters fitted to the parts
// before it, comparing with the defaults and with final, the parameters
// fitted to all items.
func healthCheck(ctx context.Context, items []item, final weights, progress progressFunc) (HealthCheck, error) {
	size := len(items) / (healthSplits + 1)
	if size == 0 {
		return HealthCheck{}, nil
	}
	const per = 1000 // progress steps per split
	var fitted, defaults []prediction
	drift := 0.0
	def := defaultWeights()
	for i := range healthSplits {
		start, end := (i+1)*size, (i+2)*size
		if i == healthSplits-1 {
			end = len(items)
		}
		w, err := fit(ctx, items[:start], func(done, total int) { progress(i*per+done*per/total, healthSplits*per) })
		if err == errNoData {
			w, err = def, nil
		}
		if err != nil {
			return HealthCheck{}, err
		}
		for _, it := range items[start:end] {
			_, p := w.itemLoss(&it, false, nil)
			_, pd := def.itemLoss(&it, false, nil)
			_, pf := final.itemLoss(&it, false, nil)
			recalled := it.last().Rating > 1
			fitted = append(fitted, prediction{p, recalled})
			defaults = append(defaults, prediction{pd, recalled})
			drift += math.Abs(p - pf)
		}
		progress((i+1)*per, healthSplits*per)
	}
	h := HealthCheck{LogLoss: logLoss(fitted), DefaultLogLoss: logLoss(defaults), Drift: drift / float64(len(fitted))}
	h.Passed = h.LogLoss <= h.DefaultLogLoss && h.Drift <= maxHealthDrift
	return h, nil
}

// prediction is a predicted recall probability and whether the card was
// recalled.
type prediction struct {
	p        float64
	recalled bool
}

func logLoss(preds []prediction) float64 {
	if len(preds) == 0 {
		return 0
	}
	l := 0.0
	for _, pr := range preds {
		p := min(max(pr.p, 1e-7), 1-1e-7)
		if pr.recalled {
			l -= math.Log(p)
		} else {
			l -= math.Log(1 - p)
		}
	}
	return l / float64(len(preds))
}
