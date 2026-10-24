package scheduler

import (
	"cardflip/golang/internal/usermsg"
	"errors"
	"math"
)

// The FSRS-5 model as the optimizer trains it, ported from fsrs-rs
// (model.rs): the forgetting curve and the memory state after each review,
// with exact derivatives of the stability with respect to the parameters
// (forward-mode differentiation), so training needs no autodiff library and
// runs in WebAssembly. fsrs-rs is under the BSD-3-Clause license, see
// THIRD_PARTY_NOTICES.md.

// NumWeights is the number of FSRS parameters of the model go-fsrs
// implements (FSRS-5).
const NumWeights = 19

const (
	fsrsDecay  = -0.5
	fsrsFactor = 19.0 / 81.0
	sMin       = 0.01
	sMax       = 36500.0
	// initSMax is the largest initial stability.
	initSMax = 100.0
)

// weights are the model parameters, and grad derivatives with respect to
// them.
type (
	weights [NumWeights]float64
	grad    [NumWeights]float64
)

// DefaultWeights returns go-fsrs's default parameters.
func DefaultWeights() []float64 {
	w := lib.W
	return w[:]
}

func defaultWeights() weights { return weights(lib.W) }

// weightsOf returns w as model parameters, or the defaults when w does not
// have exactly NumWeights values (empty, or from another model version).
func weightsOf(w []float64) weights {
	if len(w) != NumWeights {
		return defaultWeights()
	}
	return weights(w)
}

// clampBounds are the allowed range of each parameter (fsrs-rs
// parameter_clipper.rs).
var clampBounds = [NumWeights][2]float64{
	{sMin, initSMax}, {sMin, initSMax}, {sMin, initSMax}, {sMin, initSMax},
	{1, 10}, {0.001, 4}, {0.001, 4}, {0.001, 0.75}, {0, 4.5}, {0, 0.8},
	{0.001, 3.5}, {0.001, 5}, {0.001, 0.25}, {0.001, 0.9}, {0, 4},
	{0, 1}, {1, 6}, {0, 2}, {0, 2},
}

// CheckWeights reports what is wrong with FSRS parameters: they must be
// empty (the defaults) or NumWeights numbers, each within its range.
func CheckWeights(w []float64) error {
	if len(w) == 0 {
		return nil
	}
	if len(w) != NumWeights {
		return errors.New(usermsg.Text("fsrsParamsCount", "FSRS parameters must be %d numbers, not %d.", NumWeights, len(w)))
	}
	for i, v := range w {
		if b := clampBounds[i]; math.IsNaN(v) || v < b[0] || v > b[1] {
			return errors.New(usermsg.Text("fsrsParamRange", "FSRS parameter %d must be between %g and %g.", i+1, b[0], b[1]))
		}
	}
	return nil
}

func (w *weights) clip() {
	for i, b := range clampBounds {
		w[i] = min(max(w[i], b[0]), b[1])
	}
}

// memState is a memory state and, while training, the derivatives of its
// stability and difficulty with respect to the parameters.
type memState struct {
	s, d   float64
	ds, dd grad
}

// clampGrad limits v to [lo, hi]; a limited value has no derivative.
func clampGrad(v float64, d *grad, lo, hi float64) float64 {
	if v < lo || v > hi {
		*d = grad{}
		return min(max(v, lo), hi)
	}
	return v
}

func forgettingCurve(t, s float64) float64 { return math.Pow(1+fsrsFactor*t/s, fsrsDecay) }

// curve returns the recall probability after t days at stability s and its
// derivative with respect to s.
func curve(t, s float64) (r, drds float64) {
	base := 1 + fsrsFactor*t/s
	r = math.Pow(base, fsrsDecay)
	drds = fsrsDecay * math.Pow(base, fsrsDecay-1) * (-fsrsFactor * t / (s * s))
	return r, drds
}

// first is the memory state after a New card's first rating r.
func (w *weights) first(r int, g bool) memState {
	var st memState
	st.s = w[r-1]
	e := math.Exp(w[5] * float64(r-1))
	st.d = w[4] - e + 1
	if g {
		st.ds[r-1] = 1
		st.dd[4] = 1
		st.dd[5] = -float64(r-1) * e
	}
	st.d = clampGrad(st.d, &st.dd, 1, 10)
	st.s = clampGrad(st.s, &st.ds, sMin, sMax)
	return st
}

// step is the memory state after rating r, t days after the review that
// left st. A review on the same day (t = 0) uses the short-term stability.
func (w *weights) step(st memState, t float64, r int, g bool) memState {
	rf := float64(r)
	var nx memState
	switch {
	case t == 0:
		e := math.Exp(w[17] * (rf - 3 + w[18]))
		nx.s = st.s * e
		if g {
			for k := range nx.ds {
				nx.ds[k] = st.ds[k] * e
			}
			nx.ds[17] += nx.s * (rf - 3 + w[18])
			nx.ds[18] += nx.s * w[17]
		}
	case r == 1:
		rr, drds := curve(t, st.s)
		e14 := math.Exp((1 - rr) * w[14])
		dp := math.Pow(st.d, -w[12])
		sp := math.Pow(st.s+1, w[13])
		f := w[11] * dp * (sp - 1) * e14
		em := math.Exp(w[17] * w[18])
		if smin := st.s / em; smin < f {
			nx.s = smin
			if g {
				for k := range nx.ds {
					nx.ds[k] = st.ds[k] / em
				}
				nx.ds[17] -= smin * w[18]
				nx.ds[18] -= smin * w[17]
			}
			break
		}
		nx.s = f
		if g {
			viaD := f * -w[12] / st.d
			viaS := w[11]*dp*w[13]*math.Pow(st.s+1, w[13]-1)*e14 + f*-w[14]*drds
			for k := range nx.ds {
				nx.ds[k] = viaD*st.dd[k] + viaS*st.ds[k]
			}
			nx.ds[11] += dp * (sp - 1) * e14
			nx.ds[12] += f * -math.Log(st.d)
			nx.ds[13] += w[11] * dp * sp * math.Log(st.s+1) * e14
			nx.ds[14] += f * (1 - rr)
		}
	default:
		rr, drds := curve(t, st.s)
		hp, eb := 1.0, 1.0
		if r == 2 {
			hp = w[15]
		}
		if r == 4 {
			eb = w[16]
		}
		ew8 := math.Exp(w[8])
		spw := math.Pow(st.s, -w[9])
		e10 := math.Exp((1 - rr) * w[10])
		base := ew8 * (11 - st.d) * spw * (e10 - 1)
		inner := base * hp * eb
		nx.s = st.s * (inner + 1)
		if g {
			viaD := -hp * eb * ew8 * spw * (e10 - 1)
			viaS := hp * eb * ew8 * (11 - st.d) * ((e10-1)*-w[9]*math.Pow(st.s, -w[9]-1) + spw*e10*-w[10]*drds)
			var di grad
			for k := range di {
				di[k] = viaD*st.dd[k] + viaS*st.ds[k]
			}
			di[8] += inner
			di[9] += inner * -math.Log(st.s)
			di[10] += hp * eb * ew8 * (11 - st.d) * spw * e10 * (1 - rr)
			if r == 2 {
				di[15] += base * eb
			}
			if r == 4 {
				di[16] += base * hp
			}
			for k := range nx.ds {
				nx.ds[k] = st.ds[k]*(inner+1) + st.s*di[k]
			}
		}
	}

	// Difficulty: linear damping, then mean reversion towards the initial
	// difficulty of Easy (not limited to 1–10, as in fsrs-rs).
	delta := -w[6] * (rf - 3)
	nd := st.d + (10-st.d)*delta/9
	e5 := math.Exp(3 * w[5])
	init4 := w[4] - e5 + 1
	nx.d = w[7]*(init4-nd) + nd
	if g {
		var dnd grad
		for k := range dnd {
			dnd[k] = st.dd[k] * (1 - delta/9)
		}
		dnd[6] += (10 - st.d) * -(rf - 3) / 9
		for k := range nx.dd {
			nx.dd[k] = (1 - w[7]) * dnd[k]
		}
		nx.dd[4] += w[7]
		nx.dd[5] += w[7] * -3 * e5
		nx.dd[7] += init4 - nd
	}
	nx.d = clampGrad(nx.d, &nx.dd, 1, 10)
	nx.s = clampGrad(nx.s, &nx.ds, sMin, sMax)
	return nx
}

// replay returns the memory state after the reviews of seq, starting from
// start when the history before them is missing. g asks for derivatives.
func (w *weights) replay(reviews []SeqReview, start *memState, g bool) memState {
	var st memState
	i := 0
	if start != nil {
		st = *start
	} else {
		st = w.first(int(reviews[0].Rating), g)
		i = 1
	}
	for _, rv := range reviews[i:] {
		st = w.step(st, float64(rv.DeltaT), int(rv.Rating), g)
	}
	return st
}

// startState is the memory state a card with missing history is assumed to
// have had before its first known review, Anki's estimate from SM-2: the
// stability at which recall falls to the historical retention after ivl
// days, and the difficulty that would have given an ease of 250 %.
func (w *weights) startState(ivl, retention float64) memState {
	const ease = 2.5
	s := math.Max(ivl, sMin) * fsrsFactor / (math.Pow(retention, 1/fsrsDecay) - 1)
	d := 11 - (ease-1)/(math.Exp(w[8])*math.Pow(s, -w[9])*math.Expm1((1-retention)*w[10]))
	if math.IsNaN(d) || math.IsInf(d, 0) {
		d = 5
	}
	return memState{s: min(max(s, sMin), sMax), d: min(max(d, 1), 10)}
}

// itemLoss is the weighted log loss of one item: its history predicts the
// recall of its last review. With g the derivatives are added to gr.
func (w *weights) itemLoss(it *item, g bool, gr *grad) (loss, pred float64) {
	hist := it.reviews[:len(it.reviews)-1]
	last := it.reviews[len(it.reviews)-1]
	st := w.replay(hist, nil, g)
	r, drds := curve(float64(last.DeltaT), st.s)
	y := 0.0
	if last.Rating > 1 {
		y = 1
	}
	loss = -it.weight * (y*math.Log(r) + (1-y)*math.Log(1-r))
	if g {
		dl := -it.weight * (y/r - (1-y)/(1-r)) * drds
		for k := range gr {
			gr[k] += dl * st.ds[k]
		}
	}
	return loss, r
}

// paramsStddev scales the penalty for moving away from the initial
// parameters (fsrs-rs training.rs).
var paramsStddev = [NumWeights]float64{
	6.61, 9.52, 17.69, 27.74, 0.55, 0.28, 0.67, 0.12, 0.4, 0.18, 0.34, 0.27, 0.08, 0.14, 0.57,
	0.25, 1.03, 0.27, 0.39,
}

// penalty is the L2 regularisation of a batch of batch items out of total,
// added to the gradient gr.
func (w *weights) penalty(init *weights, batch, total int, gamma float64, gr *grad) float64 {
	scale := gamma * float64(batch) / float64(total)
	p := 0.0
	for k := range w {
		diff := w[k] - init[k]
		v := paramsStddev[k] * paramsStddev[k]
		p += diff * diff / v
		if gr != nil {
			gr[k] += scale * 2 * diff / v
		}
	}
	return p * scale
}
