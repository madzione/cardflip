package scheduler

import (
	"cmp"
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"slices"

	"cardflip/golang/internal/yield"
)

// Training as fsrs-rs does it (training.rs, pre_training.rs, dataset.rs):
// pre-training fits the initial stabilities to the first long-term reviews,
// then Adam minimises the log loss of every item with a cosine-annealed
// learning rate over a few epochs, keeping the parameters of the epoch with
// the lowest loss. fsrs-rs is under the BSD-3-Clause license, see
// THIRD_PARTY_NOTICES.md.

const (
	trainEpochs    = 5
	trainBatchSize = 512
	trainRate      = 4e-2
	trainMaxSeqLen = 64
	trainGamma     = 1.0
	trainSeed      = 2023
)

// errNoData is returned when the items do not allow fitting initial
// stabilities.
var errNoData = errors.New("not enough data")

// item is one training example: a card's reviews up to the one whose
// recall is predicted. time orders items by that review.
type item struct {
	reviews []SeqReview
	weight  float64
	time    int64
}

func (it *item) last() SeqReview { return it.reviews[len(it.reviews)-1] }

func (it *item) longTermCount() int {
	n := 0
	for _, r := range it.reviews {
		if r.DeltaT > 0 {
			n++
		}
	}
	return n
}

func (it *item) firstLongTerm() SeqReview {
	for _, r := range it.reviews {
		if r.DeltaT > 0 {
			return r
		}
	}
	return SeqReview{}
}

// progressFunc reports done out of total steps; it is never nil inside the
// optimizer.
type progressFunc func(done, total int)

// fit returns the parameters fitted to items, which are ordered by the time
// of their last review.
func fit(ctx context.Context, items []item, progress progressFunc) (weights, error) {
	avg := averageRecall(items)
	pre, train := prepareTraining(items)
	if len(train) < 8 {
		return defaultWeights(), nil
	}
	init4, counts, err := pretrain(pre, avg)
	if err != nil {
		return weights{}, err
	}
	w := defaultWeights()
	copy(w[:4], init4[:])
	if len(train) == len(pre) || len(train) < 64 {
		return w, nil
	}
	recencyWeight(train)
	train = slices.DeleteFunc(train, func(it item) bool { return len(it.reviews) > trainMaxSeqLen })
	w, err = adamTrain(ctx, train, w, progress)
	if err != nil {
		return weights{}, err
	}
	stab := map[int]float64{1: w[0], 2: w[1], 3: w[2], 4: w[3]}
	s0, err := smoothAndFill(stab, counts)
	if err != nil {
		return weights{}, err
	}
	copy(w[:4], s0[:])
	return w, nil
}

func averageRecall(items []item) float64 {
	if len(items) == 0 {
		return 0
	}
	n := 0
	for i := range items {
		if items[i].last().Rating > 1 {
			n++
		}
	}
	return float64(n) / float64(len(items))
}

// prepareTraining splits off the items whose last review is the card's
// first long-term one (the pre-training set), removes outliers, and returns
// the pre-training set and the whole training set, which starts with it.
func prepareTraining(items []item) (pre, train []item) {
	var rest []item
	for _, it := range items {
		if it.longTermCount() == 1 {
			pre = append(pre, it)
		} else {
			rest = append(rest, it)
		}
	}
	pre, rest = filterOutliers(pre, rest)
	train = append(slices.Clone(pre), rest...)
	return pre, train
}

// filterOutliers drops, for each first rating, the rarest delta_t groups of
// first long-term reviews (5 %, at least 20 items), groups of fewer than six
// items and long delta_ts, and the training items whose first long-term
// review falls in a dropped group.
func filterOutliers(pre, rest []item) ([]item, []item) {
	groups := map[int]map[int][]item{}
	for _, it := range pre {
		r := int(it.reviews[0].Rating)
		if groups[r] == nil {
			groups[r] = map[int][]item{}
		}
		dt := it.last().DeltaT
		groups[r][dt] = append(groups[r][dt], it)
	}
	var kept []item
	removed := map[[2]int]bool{}
	for _, rating := range sortedKeys(groups) {
		type sub struct {
			dt    int
			items []item
		}
		var subs []sub
		total := 0
		for dt, its := range groups[rating] {
			subs = append(subs, sub{dt, its})
			total += len(its)
		}
		// Smallest groups first, and among equal sizes shorter delta_ts.
		slices.SortFunc(subs, func(a, b sub) int {
			return cmp.Or(cmp.Compare(len(a.items), len(b.items)), cmp.Compare(a.dt, b.dt))
		})
		limit := max(20, total/20)
		dropped := 0
		maxDT := 100
		if rating == 4 {
			maxDT = 365
		}
		for _, s := range subs {
			if dropped+len(s.items) >= limit {
				if len(s.items) >= 6 && s.dt <= maxDT {
					kept = append(kept, s.items...)
				} else {
					removed[[2]int{rating, s.dt}] = true
				}
			} else {
				dropped += len(s.items)
				removed[[2]int{rating, s.dt}] = true
			}
		}
	}
	rest = slices.DeleteFunc(rest, func(it item) bool {
		return removed[[2]int{int(it.reviews[0].Rating), it.firstLongTerm().DeltaT}]
	})
	return kept, rest
}

func sortedKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// recencyWeight weights items from 0.25 (first) to 1 (last), so later
// reviews count more.
func recencyWeight(items []item) {
	n := float64(len(items) - 1)
	if n <= 0 {
		for i := range items {
			items[i].weight = 1
		}
		return
	}
	for i := range items {
		items[i].weight = 0.25 + 0.75*math.Pow(float64(i)/n, 3)
	}
}

// pretrain fits the initial stability of each first rating to the recall
// of the first long-term reviews, and returns them with the number of
// items per first rating.
func pretrain(pre []item, avgRecall float64) ([4]float64, map[int]int, error) {
	type point struct{ dt, recall, count float64 }
	groups := map[int]map[int][]bool{}
	for _, it := range pre {
		r := int(it.reviews[0].Rating)
		lt := it.firstLongTerm()
		if groups[r] == nil {
			groups[r] = map[int][]bool{}
		}
		groups[r][lt.DeltaT] = append(groups[r][lt.DeltaT], lt.Rating > 1)
	}
	counts := map[int]int{}
	stab := map[int]float64{}
	def := defaultWeights()
	for r, byDT := range groups {
		var pts []point
		for _, dt := range sortedKeys(byDT) {
			passed := 0
			for _, ok := range byDT[dt] {
				if ok {
					passed++
				}
			}
			n := float64(len(byDT[dt]))
			// Laplace smoothing towards the average recall.
			pts = append(pts, point{float64(dt), (float64(passed) + avgRecall) / (n + 1), n})
			counts[r] += len(byDT[dt])
		}
		s0 := def[r-1]
		loss := func(s float64) float64 {
			l := 0.0
			for _, p := range pts {
				y := forgettingCurve(p.dt, s)
				l -= (p.recall*math.Log(y) + (1-p.recall)*math.Log(1-y)) * p.count
			}
			return l + math.Abs(s-s0)/16
		}
		lo, hi := sMin, initSMax
		best := s0
		for i := 0; hi-lo > 2.220446049250313e-16 && i < 1000; i++ {
			m1 := lo + (hi-lo)/3
			m2 := hi - (hi-lo)/3
			if loss(m1) < loss(m2) {
				hi = m2
			} else {
				lo = m1
			}
			best = (hi + lo) / 2
		}
		stab[r] = best
	}
	s0, err := smoothAndFill(stab, counts)
	return s0, counts, err
}

// smoothAndFill makes the initial stabilities grow with the rating (the
// rating with more items wins) and fills in ratings without data from the
// others (fsrs-rs smooth_and_fill).
func smoothAndFill(stab map[int]float64, counts map[int]int) ([4]float64, error) {
	for r := range stab {
		if _, ok := counts[r]; !ok {
			delete(stab, r)
		}
	}
	for _, p := range [][2]int{{1, 2}, {2, 3}, {3, 4}, {1, 3}, {2, 4}, {1, 4}} {
		small, ok1 := stab[p[0]]
		big, ok2 := stab[p[1]]
		if ok1 && ok2 && small > big {
			if counts[p[0]] > counts[p[1]] {
				stab[p[1]] = small
			} else {
				stab[p[0]] = big
			}
		}
	}
	const w1, w2 = 0.41, 0.54
	pow := math.Pow
	def := defaultWeights()
	var a [5]float64 // a[r] is rating r's stability, 0 when unknown
	for r := 1; r <= 4; r++ {
		a[r] = stab[r]
	}
	has := func(rs ...int) bool {
		for r := 1; r <= 4; r++ {
			if (a[r] != 0) != slices.Contains(rs, r) {
				return false
			}
		}
		return true
	}
	switch len(stab) {
	case 0:
		return [4]float64{}, errNoData
	case 1:
		for r, v := range stab {
			f := v / def[r-1]
			for i := 1; i <= 4; i++ {
				a[i] = def[i-1] * f
			}
		}
	case 2:
		d := w1 + w2 - w1*w2
		switch {
		case has(3, 4):
			a[2] = pow(a[3], 1/(1-w2)) * pow(a[4], 1-1/(1-w2))
			a[1] = pow(a[2], 1/w1) * pow(a[3], 1-1/w1)
		case has(2, 4):
			a[3] = pow(a[2], 1-w2) * pow(a[4], w2)
			a[1] = pow(a[2], 1/w1) * pow(a[3], 1-1/w1)
		case has(2, 3):
			a[4] = pow(a[2], 1-1/w2) * pow(a[3], 1/w2)
			a[1] = pow(a[2], 1/w1) * pow(a[3], 1-1/w1)
		case has(1, 4):
			a[2] = pow(a[1], w1/d) * pow(a[4], 1-w1/d)
			a[3] = pow(a[1], 1-w2/d) * pow(a[4], w2/d)
		case has(1, 3):
			a[2] = pow(a[1], w1) * pow(a[3], 1-w1)
			a[4] = pow(a[2], 1-1/w2) * pow(a[3], 1/w2)
		case has(1, 2):
			a[3] = pow(a[1], 1-1/(1-w1)) * pow(a[2], 1/(1-w1))
			a[4] = pow(a[2], 1-1/w2) * pow(a[3], 1/w2)
		}
	case 3:
		switch {
		case a[1] == 0:
			a[1] = pow(a[2], 1/w1) * pow(a[3], 1-1/w1)
		case a[2] == 0:
			a[2] = pow(a[1], w1) * pow(a[3], 1-w1)
		case a[3] == 0:
			a[3] = pow(a[2], 1-w2) * pow(a[4], w2)
		case a[4] == 0:
			a[4] = pow(a[2], 1-1/w2) * pow(a[3], 1/w2)
		}
	}
	var out [4]float64
	for i := range out {
		out[i] = min(max(a[i+1], sMin), initSMax)
	}
	return out, nil
}

// cosineLR is the cosine annealing learning rate schedule of fsrs-rs, from
// the initial rate to 0 over tMax steps.
type cosineLR struct {
	tMax, init, lr, step float64
}

func newCosineLR(tMax, init float64) *cosineLR {
	return &cosineLR{tMax: tMax, init: init, lr: init, step: -1}
}

func (c *cosineLR) next() float64 {
	c.step++
	switch {
	case c.step == 0:
		c.lr = c.init
	case math.Mod(c.step-1-c.tMax, 2*c.tMax) == 0:
		c.lr = c.init * (1 - math.Cos(math.Pi/c.tMax)) / 2
	default:
		c.lr = (1 + math.Cos(math.Pi*c.step/c.tMax)) / (1 + math.Cos(math.Pi*(c.step-1)/c.tMax)) * c.lr
	}
	return c.lr
}

// adam is the Adam optimiser (β1 0.9, β2 0.999, ε 1e-8).
type adam struct {
	m, v grad
	t    int
}

func (a *adam) step(w *weights, g *grad, lr float64) {
	const b1, b2, eps = 0.9, 0.999, 1e-8
	a.t++
	c1 := 1 - math.Pow(b1, float64(a.t))
	c2 := 1 - math.Pow(b2, float64(a.t))
	for k := range w {
		a.m[k] = b1*a.m[k] + (1-b1)*g[k]
		a.v[k] = b2*a.v[k] + (1-b2)*g[k]*g[k]
		w[k] -= lr * (a.m[k] / c1) / (math.Sqrt(a.v[k]/c2) + eps)
	}
}

// adamTrain minimises the weighted log loss of items, starting from init.
// Items are batched by length and the batches shuffled each epoch.
func adamTrain(ctx context.Context, items []item, init weights, progress progressFunc) (weights, error) {
	sorted := slices.Clone(items)
	slices.SortStableFunc(sorted, func(a, b item) int { return cmp.Compare(len(a.reviews), len(b.reviews)) })
	var batches [][]item
	for i := 0; i < len(sorted); i += trainBatchSize {
		batches = append(batches, sorted[i:min(i+trainBatchSize, len(sorted))])
	}
	total := len(sorted)
	lr := newCosineLR(float64((total/trainBatchSize+1)*trainEpochs), trainRate)
	rng := rand.New(rand.NewPCG(trainSeed, trainSeed))
	w := init
	var opt adam
	best, bestLoss := w, math.Inf(1)
	order := make([]int, len(batches))
	steps := len(batches) * trainEpochs
	for epoch := 0; epoch < trainEpochs; epoch++ {
		for i := range order {
			order[i] = i
		}
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for n, bi := range order {
			if err := ctx.Err(); err != nil {
				return weights{}, err
			}
			b := batches[bi]
			var g grad
			for i := range b {
				w.itemLoss(&b[i], true, &g)
			}
			w.penalty(&init, len(b), total, trainGamma, &g)
			opt.step(&w, &g, lr.next())
			w.clip()
			progress(epoch*len(batches)+n+1, steps)
			yield.Yield()
		}
		loss := 0.0
		for _, b := range batches {
			for i := range b {
				l, _ := w.itemLoss(&b[i], false, nil)
				loss += l
			}
			loss += w.penalty(&init, len(b), total, trainGamma, nil)
		}
		if loss /= float64(total); loss < bestLoss {
			best, bestLoss = w, loss
		}
	}
	for _, v := range best {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return weights{}, errors.New("optimizing gave invalid parameters")
		}
	}
	return best, nil
}

// evaluation is how well parameters predict items: the weighted log loss
// and the RMSE of the predicted against the actual recall over bins of
// similar items (fsrs-rs evaluate).
type evaluation struct {
	LogLoss, RMSE float64
}

func evaluate(w weights, items []item) evaluation {
	type bin struct{ pred, actual, count, weight float64 }
	bins := map[[3]int]*bin{}
	loss, wsum := 0.0, 0.0
	for i := range items {
		it := &items[i]
		l, p := w.itemLoss(it, false, nil)
		loss += l
		wsum += it.weight
		k := rMatrixIndex(it)
		b := bins[k]
		if b == nil {
			b = &bin{}
			bins[k] = b
		}
		b.pred += p
		if it.last().Rating > 1 {
			b.actual++
		}
		b.count++
		b.weight += it.weight
	}
	if wsum == 0 {
		return evaluation{}
	}
	se, bw := 0.0, 0.0
	for _, b := range bins {
		d := b.pred/b.count - b.actual/b.count
		se += d * d * b.weight
		bw += b.weight
	}
	return evaluation{LogLoss: loss / wsum, RMSE: math.Sqrt(se / bw)}
}

// rMatrixIndex groups an item by its last delta_t, its number of long-term
// reviews and its lapses, on logarithmic scales.
func rMatrixIndex(it *item) [3]int {
	logBin := func(v, base, scale float64) int {
		return int(math.Round(scale * math.Pow(base, math.Floor(math.Log(v)/math.Log(base)))))
	}
	dt := float64(it.last().DeltaT)
	dtBin := int(math.Round(2.48 * math.Pow(3.62, math.Floor(math.Log(dt)/math.Log(3.62))) * 100))
	length := float64(it.longTermCount() + 1)
	lapses := 0
	for _, r := range it.reviews[:len(it.reviews)-1] {
		if r.Rating == Again && r.DeltaT > 0 {
			lapses++
		}
	}
	lapseBin := 0
	if lapses > 0 {
		lapseBin = logBin(float64(lapses), 1.73, 1.65)
	}
	return [3]int{dtBin, logBin(length, 1.89, 1.99), lapseBin}
}
