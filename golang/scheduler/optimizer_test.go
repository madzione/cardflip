package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func day(d int, hour int) time.Time { return time.Date(2026, 3, d, hour, 0, 0, 0, time.UTC) }

func ratings(rs []SeqReview) (out [][2]int) {
	for _, r := range rs {
		out = append(out, [2]int{int(r.Rating), r.DeltaT})
	}
	return out
}

func TestNewSequenceCountsStudyDays(t *testing.T) {
	h := []HistoryReview{
		{Time: day(1, 10), Rating: Again, StateBefore: New},
		{Time: day(1, 10).Add(time.Minute), Rating: Good, StateBefore: Learning},
		{Time: day(1, 11), Rating: Good, StateBefore: Learning},
		// 02:00 on the 3rd still belongs to study day 2 (rollover 04:00).
		{Time: day(3, 2), Rating: Good, StateBefore: Review, IvlBefore: 1},
		{Time: day(6, 9), Rating: Hard, StateBefore: Review, IvlBefore: 3},
	}
	s := NewSequence(h, HistoryOptions{RolloverHour: 4})
	if s.Incomplete {
		t.Fatal("history starting as a New card is complete")
	}
	want := [][2]int{{1, 0}, {3, 0}, {3, 0}, {3, 1}, {2, 4}}
	if got := ratings(s.Reviews); !slices.Equal(got, want) {
		t.Fatalf("reviews = %v, want %v", got, want)
	}
	if !s.LastReview.Equal(day(6, 9)) {
		t.Errorf("last review = %v", s.LastReview)
	}
}

func TestNewSequenceManualEntries(t *testing.T) {
	h := []HistoryReview{
		{Time: day(1, 10), Rating: Good, StateBefore: New},
		{Time: day(2, 10), Rating: Good, StateBefore: Review, IvlBefore: 1},
		// A set due date is ignored, a reset starts over.
		{Time: day(3, 10), StateBefore: Review, IvlBefore: 3, IvlAfter: 10},
		{Time: day(4, 10), Rating: Again, StateBefore: Review, IvlBefore: 10},
		{Time: day(5, 10), StateBefore: Relearning},
		{Time: day(7, 10), Rating: Easy, StateBefore: New},
		{Time: day(9, 10), Rating: Good, StateBefore: Review, IvlBefore: 2},
	}
	s := NewSequence(h, HistoryOptions{})
	want := [][2]int{{4, 0}, {3, 2}}
	if got := ratings(s.Reviews); !slices.Equal(got, want) || s.Incomplete {
		t.Fatalf("reviews = %v (incomplete %v), want %v", got, s.Incomplete, want)
	}
}

func TestNewSequenceIncompleteHistory(t *testing.T) {
	h := []HistoryReview{
		{Time: day(1, 10), Rating: Good, StateBefore: New},
		{Time: day(5, 10), Rating: Good, StateBefore: Review, IvlBefore: 4},
		{Time: day(15, 10), Rating: Again, StateBefore: Review, IvlBefore: 10},
		{Time: day(15, 11), Rating: Good, StateBefore: Relearning},
		{Time: day(17, 10), Rating: Good, StateBefore: Review, IvlBefore: 2},
	}
	// The first review is before the cut-off: the history is incomplete,
	// and its first kept review is counted from the last ignored one.
	s := NewSequence(h, HistoryOptions{IgnoreBefore: day(3, 0), HistoricalRetention: 0.8})
	want := [][2]int{{3, 4}, {1, 10}, {3, 0}, {3, 2}}
	if got := ratings(s.Reviews); !slices.Equal(got, want) {
		t.Fatalf("reviews = %v, want %v", got, want)
	}
	if !s.Incomplete || s.StartInterval != 4 || s.retention != 0.8 {
		t.Fatalf("incomplete %v, start interval %v, retention %v", s.Incomplete, s.StartInterval, s.retention)
	}
	if n := TrainingReviews([]Sequence{s}); n != 0 {
		t.Errorf("incomplete history gives %d training reviews, want 0", n)
	}

	// Without any earlier entry the first review waited its interval.
	s = NewSequence(h[2:], HistoryOptions{})
	if got := ratings(s.Reviews); !slices.Equal(got, [][2]int{{1, 10}, {3, 0}, {3, 2}}) {
		t.Fatalf("reviews = %v", got)
	}

	// A lower historical retention means the card was forgotten sooner, so
	// its estimated stability is lower.
	m90, _ := NewSequence(h[2:], HistoryOptions{HistoricalRetention: 0.9}).Memory(nil)
	m70, _ := NewSequence(h[2:], HistoryOptions{HistoricalRetention: 0.7}).Memory(nil)
	if m70.Stability >= m90.Stability {
		t.Errorf("stability at 70%% = %v, at 90%% = %v", m70.Stability, m90.Stability)
	}
}

func TestTrainingItemsArePerStudyDay(t *testing.T) {
	h := []HistoryReview{
		{Time: day(1, 10), Rating: Again, StateBefore: New},
		{Time: day(1, 11), Rating: Good, StateBefore: Learning},
		{Time: day(2, 10), Rating: Good, StateBefore: Review},
		{Time: day(5, 10), Rating: Again, StateBefore: Review},
		{Time: day(5, 11), Rating: Good, StateBefore: Relearning},
		{Time: day(6, 10), Rating: Good, StateBefore: Review},
	}
	items := trainingItems([]Sequence{NewSequence(h, HistoryOptions{})})
	var lens []int
	for _, it := range items {
		lens = append(lens, len(it.reviews))
	}
	// Only the first review of each later day is predicted.
	if !slices.Equal(lens, []int{3, 4, 6}) {
		t.Fatalf("item lengths = %v, want [3 4 6]", lens)
	}
}

// fsrsRsBatch is the batch of fsrs-rs's test_loss_and_grad: four items
// (columns) with six history reviews each.
func fsrsRsBatch() []item {
	ts := [6][4]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 1}, {0, 1, 1, 3}, {1, 3, 3, 5}, {3, 6, 6, 12}}
	rs := [6][4]int{{1, 2, 3, 4}, {3, 4, 2, 4}, {1, 4, 4, 3}, {4, 3, 3, 3}, {3, 1, 3, 3}, {2, 3, 3, 4}}
	deltas := [4]int{4, 11, 12, 23}
	labels := [4]Rating{Good, Good, Good, Again}
	var items []item
	for j := range 4 {
		var revs []SeqReview
		for i := range 6 {
			revs = append(revs, SeqReview{Rating: Rating(rs[i][j]), DeltaT: ts[i][j]})
		}
		revs = append(revs, SeqReview{Rating: labels[j], DeltaT: deltas[j]})
		items = append(items, item{reviews: revs, weight: 1})
	}
	return items
}

func batchLoss(w *weights, items []item) (float64, grad) {
	var g grad
	loss := 0.0
	for i := range items {
		l, _ := w.itemLoss(&items[i], true, &g)
		loss += l
	}
	return loss, g
}

func approx(t *testing.T, what string, got, want []float64, tol float64) {
	t.Helper()
	for i := range want {
		if d := math.Abs(got[i] - want[i]); d > tol*math.Max(1, math.Abs(want[i])) {
			t.Errorf("%s[%d] = %v, want %v", what, i, got[i], want[i])
		}
	}
}

// TestLossAndGradientMatchFsrsRs replays fsrs-rs's test_loss_and_grad: the
// loss, its gradient, two Adam steps and the L2 penalty must give the
// reference's values.
func TestLossAndGradientMatchFsrsRs(t *testing.T) {
	items := fsrsRsBatch()
	w := defaultWeights()
	init := w

	loss, g := batchLoss(&w, items)
	approx(t, "loss", []float64{loss}, []float64{4.4467363}, 1e-5)
	approx(t, "grad", g[:], []float64{
		-0.05832, -0.00682, -0.00255, 0.010539, -0.05128, 1.364291, 0.083658, -0.95023,
		0.534472, -2.89288, 0.514163, -0.01306, 0.041905, -0.11830, -0.00092, -0.14452,
		0.202374, 0.214104, 0.032307,
	}, 1e-3)

	var opt adam
	opt.step(&w, &g, 0.04)
	w.clip()
	approx(t, "weights after step 1", w[:], []float64{
		0.44255, 1.22385, 3.2129998, 15.65105, 7.2349, 0.4945, 1.4204, 0.0446, 1.5057501,
		0.1592, 0.97925, 1.9794999, 0.07000001, 0.33605, 2.3097994, 0.2715, 2.9498,
		0.47655, 0.62210006,
	}, 1e-5)

	var pg grad
	p := w.penalty(&init, 512, 1000, 2, &pg)
	approx(t, "penalty", []float64{p}, []float64{0.64689976}, 1e-5)
	approx(t, "penalty grad", pg[:], []float64{
		0.0018749383, 0.00090389, 0.00026177685, -0.00010645759, 0.27080965, -1.0448978,
		-0.18249036, 5.688889, -0.5119995, 2.528395, -0.7086509, 1.1237301, -12.799997,
		4.179591, 0.25213587, 1.3107198, -0.07721739, -1.1237309, -0.5385926,
	}, 1e-4)

	loss, g = batchLoss(&w, items)
	approx(t, "loss 2", []float64{loss}, []float64{4.176347}, 1e-5)
	approx(t, "grad 2", g[:], []float64{
		-0.0401341, -0.0061790533, -0.00288913, 0.01216853, -0.05624995, 1.147413, 0.068084724,
		-0.6906936, 0.48760873, -2.5428302, 0.49044546, -0.011574259, 0.037729632, -0.09633919,
		-0.0009513022, -0.12789416, 0.19088513, 0.2574597, 0.049311582,
	}, 1e-3)
	opt.step(&w, &g, 0.04)
	w.clip()
	approx(t, "weights after step 2", w[:], []float64{
		0.48150504, 1.2636971, 3.2530522, 15.611003, 7.2749534, 0.45482785, 1.3808222,
		0.083782874, 1.4658877, 0.19898315, 0.9393105, 2.0193, 0.030164223, 0.37562984,
		2.3498251, 0.3112984, 2.909878, 0.43652722, 0.5825156,
	}, 1e-4)
}

// TestGradientMatchesFiniteDifferences checks the analytic derivatives of
// every branch of the model against central differences.
func TestGradientMatchesFiniteDifferences(t *testing.T) {
	w := weights{0.6, 1.3, 4.4, 9.9, 7.6, 0.78, 2.3, 0.05, 1.3, 0.13, 0.83, 1.9, 0.13, 0.27, 2.4, 0.1, 3.1, 0.21, 0.28}
	items := fsrsRsBatch()
	// Failures and successes after long intervals, which the batch lacks.
	items = append(items, item{weight: 1, reviews: []SeqReview{{3, 0, 0}, {1, 30, 0}, {3, 0, 0}, {2, 9, 0}, {4, 40, 0}, {1, 2, 0}}})
	_, g := batchLoss(&w, items)
	for k := range w {
		const h = 1e-6
		up, down := w, w
		up[k] += h
		down[k] -= h
		lu, _ := batchLoss(&up, items)
		ld, _ := batchLoss(&down, items)
		if fd := (lu - ld) / (2 * h); math.Abs(fd-g[k]) > 1e-5*math.Max(1, math.Abs(fd)) {
			t.Errorf("d loss / d w[%d] = %v, finite difference %v", k, g[k], fd)
		}
	}
}

func TestCosineScheduleMatchesFsrsRs(t *testing.T) {
	// fsrs-rs cosine_annealing.rs, lr_scheduler test.
	lr := newCosineLR(5, 4e-2)
	var got []float64
	for range 11 {
		got = append(got, lr.next())
	}
	approx(t, "lr", got, []float64{
		0.04, 0.03618033988749895, 0.026180339887498946, 0.013819660112501051,
		0.0038196601125010526, 0.0, 0.003819660112501051, 0.013819660112501048,
		0.026180339887498943, 0.03618033988749895, 0.039999999999999994,
	}, 1e-12)
}

// simulate generates review histories of cards from the model with
// parameters w: a first rating, sometimes a same-day step, then reviews at
// intervals close to 90 % retention, each recalled with the model's
// probability, and relearning on the day of a lapse.
func simulate(w weights, cards int, seed uint64) []Sequence {
	rng := rand.New(rand.NewPCG(seed, 7))
	pick := func(probs ...float64) Rating {
		x := rng.Float64()
		for i, p := range probs {
			if x < p {
				return Rating(i + 1)
			}
			x -= p
		}
		return Rating(len(probs))
	}
	var seqs []Sequence
	for c := range cards {
		d := rng.IntN(300)
		at := func() int64 { return int64(d)*86_400_000 + int64(c) }
		r := pick(0.25, 0.1, 0.55, 0.1)
		st := w.first(int(r), false)
		revs := []SeqReview{{Rating: r, at: at()}}
		if r < Good {
			st = w.step(st, 0, int(Good), false)
			revs = append(revs, SeqReview{Rating: Good, at: at() + 1})
		}
		for len(revs) < 16 && d < 700 {
			ivl := st.s / fsrsFactor * (math.Pow(0.9, 1/fsrsDecay) - 1) * (0.6 + 0.8*rng.Float64())
			dt := max(int(math.Round(ivl)), 1)
			d += dt
			r = Again
			if rng.Float64() < forgettingCurve(float64(dt), st.s) {
				r = pick(0, 0.15, 0.75, 0.1)
			}
			st = w.step(st, float64(dt), int(r), false)
			revs = append(revs, SeqReview{Rating: r, DeltaT: dt, at: at()})
			if r == Again {
				st = w.step(st, 0, int(Good), false)
				revs = append(revs, SeqReview{Rating: Good, at: at() + 1})
			}
		}
		seqs = append(seqs, Sequence{Reviews: revs})
	}
	return seqs
}

// trueWeights are the parameters fsrs-rs's tests fit to its sample
// collection: realistic, and different from the defaults.
var trueWeights = weights{
	0.6032805, 1.3376843, 4.4167747, 9.933699, 7.654044, 0.78219295, 2.336606, 0.001,
	1.3264198, 0.12967199, 0.82880765, 1.9360433, 0.13298263, 0.27427456, 2.4304862,
	0.10340813, 3.108867, 0.2114512, 0.2826002,
}

// TestOptimizeRecoversParameters fits histories generated from known
// parameters: the fit must predict them about as well as the true
// parameters, better than the defaults, and give similar stabilities.
func TestOptimizeRecoversParameters(t *testing.T) {
	seqs := simulate(trueWeights, 3000, 1)
	res, err := Optimize(context.Background(), seqs, OptimizeOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Better {
		t.Fatalf("fitted loss %v not below default loss %v", res.LogLoss, res.CurrentLogLoss)
	}
	items := trainingItems(seqs)
	recencyWeight(items)
	truth := evaluate(trueWeights, items)
	if res.LogLoss > truth.LogLoss+0.002 {
		t.Errorf("fitted loss %v, true parameters' loss %v", res.LogLoss, truth.LogLoss)
	}
	// Predicted recall of every item within a few percent of the truth.
	fitted := weights(res.Weights)
	diff := 0.0
	for i := range items {
		_, p := fitted.itemLoss(&items[i], false, nil)
		_, q := trueWeights.itemLoss(&items[i], false, nil)
		diff += math.Abs(p - q)
	}
	if mean := diff / float64(len(items)); mean > 0.015 {
		t.Errorf("mean difference of predicted recall = %v", mean)
	}
	// The initial stabilities of the common first ratings.
	for _, r := range []int{0, 2} {
		if rel := math.Abs(fitted[r]-trueWeights[r]) / trueWeights[r]; rel > 0.25 {
			t.Errorf("initial stability %d = %v, true %v", r+1, fitted[r], trueWeights[r])
		}
	}
}

// reference is a result of fsrs-rs 2.0.4's compute_parameters (FSRS-5,
// short-term on) on the fixture, made with the example in
// testdata/fsrs_reference.md.
type reference struct {
	Cards          [][][3]int64 `json:"cards"`
	Params         []float64    `json:"params"`
	DefaultLogLoss float64      `json:"defaultLogLoss"`
	DefaultRMSE    float64      `json:"defaultRmse"`
	FittedLogLoss  float64      `json:"fittedLogLoss"`
	FittedRMSE     float64      `json:"fittedRmse"`
}

func (r reference) sequences() []Sequence {
	var seqs []Sequence
	for _, c := range r.Cards {
		var s Sequence
		for _, rv := range c {
			s.Reviews = append(s.Reviews, SeqReview{Rating: Rating(rv[0]), DeltaT: int(rv[1]), at: rv[2]})
		}
		seqs = append(seqs, s)
	}
	return seqs
}

func loadReference(t *testing.T) reference {
	raw, err := os.ReadFile("testdata/fsrs_reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var ref reference
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatal(err)
	}
	return ref
}

// TestOptimizeMatchesFsrsRs compares with fsrs-rs on the same fixture: the
// evaluation must agree exactly, and the fitted parameters must fit as
// well as fsrs-rs's (the batch order differs, so they are not identical).
func TestOptimizeMatchesFsrsRs(t *testing.T) {
	ref := loadReference(t)
	seqs := ref.sequences()
	items := trainingItems(seqs)
	recencyWeight(items)
	def := evaluate(defaultWeights(), items)
	approx(t, "default evaluation", []float64{def.LogLoss, def.RMSE}, []float64{ref.DefaultLogLoss, ref.DefaultRMSE}, 1e-4)
	theirs := evaluate(weights(ref.Params), items)
	approx(t, "reference evaluation", []float64{theirs.LogLoss, theirs.RMSE}, []float64{ref.FittedLogLoss, ref.FittedRMSE}, 1e-4)

	res, err := Optimize(context.Background(), seqs, OptimizeOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("parameters %.4f\nfsrs-rs    %.4f", res.Weights, ref.Params)
	t.Logf("log loss %.5f (fsrs-rs %.5f), RMSE %.5f (fsrs-rs %.5f)", res.LogLoss, theirs.LogLoss, res.RMSE, theirs.RMSE)
	if math.Abs(res.LogLoss-theirs.LogLoss) > 0.001 {
		t.Errorf("log loss = %v, fsrs-rs %v", res.LogLoss, theirs.LogLoss)
	}
	if math.Abs(res.RMSE-theirs.RMSE) > 0.01 {
		t.Errorf("RMSE = %v, fsrs-rs %v", res.RMSE, theirs.RMSE)
	}
	approx(t, "parameters", res.Weights, ref.Params, 0.1)
}

// TestWriteReferenceFixture writes the fixture's items in the input format
// of the fsrs-rs example when FSRS_FIXTURE_OUT is set (see
// testdata/fsrs_reference.md).
func TestWriteReferenceFixture(t *testing.T) {
	out := os.Getenv("FSRS_FIXTURE_OUT")
	if out == "" {
		t.Skip("FSRS_FIXTURE_OUT not set")
	}
	seqs := simulate(trueWeights, 400, 42)
	var cards [][][3]int64
	for _, s := range seqs {
		var c [][3]int64
		for _, r := range s.Reviews {
			c = append(c, [3]int64{int64(r.Rating), int64(r.DeltaT), r.at})
		}
		cards = append(cards, c)
	}
	raw, _ := json.Marshal(map[string]any{"cards": cards})
	if err := os.WriteFile(out+".cards.json", raw, 0o644); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, it := range trainingItems(seqs) {
		for i, r := range it.reviews {
			if i > 0 {
				b.WriteByte(' ')
			}
			fmt.Fprintf(&b, "%d,%d", r.Rating, r.DeltaT)
		}
		b.WriteByte('\n')
	}
	if err := os.WriteFile(out+".items.txt", []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestOptimizeNeedsEnoughReviews(t *testing.T) {
	seqs := simulate(trueWeights, 20, 3)
	_, err := Optimize(context.Background(), seqs, OptimizeOptions{}, nil)
	var nr *NotEnoughReviewsError
	if !errors.As(err, &nr) || nr.Reviews != TrainingReviews(seqs) || nr.Reviews >= MinOptimizeReviews {
		t.Fatalf("err = %v", err)
	}
}

func TestOptimizeKeepsBetterCurrentParameters(t *testing.T) {
	seqs := simulate(trueWeights, 1500, 4)
	// The true parameters cannot be beaten by much; a fit that does not
	// beat them is not used.
	res, err := Optimize(context.Background(), seqs, OptimizeOptions{Current: trueWeights[:]}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Better != (res.LogLoss < res.CurrentLogLoss) || res.Use() != res.Better {
		t.Fatalf("better %v use %v: loss %v, current %v", res.Better, res.Use(), res.LogLoss, res.CurrentLogLoss)
	}
	res.LogLoss, res.Better = res.CurrentLogLoss+1, false
	if res.Use() {
		t.Error("worse parameters are used")
	}
}

func TestOptimizeCancel(t *testing.T) {
	seqs := simulate(trueWeights, 1500, 5)
	ctx, cancel := context.WithCancel(context.Background())
	steps := 0
	_, err := Optimize(ctx, seqs, OptimizeOptions{}, func(stage string, done, total int) {
		if steps++; steps == 3 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestHealthCheck(t *testing.T) {
	var stages []string
	progress := func(stage string, done, total int) {
		if len(stages) == 0 || stages[len(stages)-1] != stage {
			stages = append(stages, stage)
		}
		if done > total {
			t.Fatalf("%s progress %d of %d", stage, done, total)
		}
	}
	res, err := Optimize(context.Background(), simulate(trueWeights, 3000, 6), OptimizeOptions{HealthCheck: true}, progress)
	if err != nil {
		t.Fatal(err)
	}
	if res.Health == nil || !res.Health.Passed {
		t.Fatalf("consistent history fails the health check: %+v", res.Health)
	}
	if !slices.Equal(stages, []string{"training", "checking"}) {
		t.Errorf("stages = %v", stages)
	}

	// Reviews that contradict each other over time: the first half of the
	// history comes from one learner, the rest from a very different one.
	// Parameters fitted to the earlier parts keep changing.
	a := simulate(trueWeights, 1200, 7)
	other := trueWeights
	other[0], other[1], other[2], other[3] = 0.1, 0.2, 0.3, 0.5
	other[8], other[9], other[10] = 0.2, 0.7, 0.2
	b := simulate(other, 1200, 8)
	for i := range b {
		for j := range b[i].Reviews {
			b[i].Reviews[j].at += 2000 * 86_400_000
		}
	}
	res, err = Optimize(context.Background(), append(a, b...), OptimizeOptions{HealthCheck: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Health.Passed || res.Use() {
		t.Errorf("inconsistent history passes the health check: %+v", res.Health)
	}
}

func TestWeightsOfIgnoresWrongLength(t *testing.T) {
	if weightsOf([]float64{1, 2, 3}) != defaultWeights() || weightsOf(nil) != defaultWeights() {
		t.Error("parameters of the wrong length are not the defaults")
	}
	if weightsOf(trueWeights[:]) != trueWeights {
		t.Error("a full set of parameters is not used")
	}
}

// TestOptimizeTiming times optimizing 100,000 simulated reviews, with and
// without the health check, when FSRS_TIMING is set.
func TestOptimizeTiming(t *testing.T) {
	if os.Getenv("FSRS_TIMING") == "" {
		t.Skip("FSRS_TIMING not set")
	}
	seqs := simulate(trueWeights, 9000, 9)
	for TrainingReviews(seqs) < 100_000 {
		seqs = append(seqs, simulate(trueWeights, 500, uint64(len(seqs)))...)
	}
	for _, health := range []bool{false, true} {
		start := time.Now()
		res, err := Optimize(context.Background(), seqs, OptimizeOptions{HealthCheck: health}, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%d reviews, health check %v: %v", res.Reviews, health, time.Since(start).Round(time.Millisecond))
	}
}
