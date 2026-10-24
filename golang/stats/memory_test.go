package stats

import (
	"math"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
)

// remembered adds n review cards with stability s reviewed elapsed days
// ago; after s days the retrievability is 0.9.
func (d *Data) remembered(n int, s, difficulty, elapsed float64) {
	for range n {
		d.Cards = append(d.Cards, core.Card{ID: core.ID(len(d.Cards) + 1), CardState: scheduler.CardState{
			State: scheduler.Review, ScheduledDays: int(s), Memory: &scheduler.Memory{Stability: s, Difficulty: difficulty},
			LastReview: d.Now.Add(-time.Duration(elapsed * 24 * float64(time.Hour)))}})
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestRetrievabilityAverageAndKnowledge(t *testing.T) {
	// Retrievability (spec): 100 review cards averaging 0.92.
	d := testData()
	d.remembered(20, 10, 5, 0)  // just reviewed: 1.0
	d.remembered(80, 10, 5, 10) // stability elapsed: 0.9
	d.Cards = append(d.Cards, core.Card{ID: 999})
	got := Retrievability(d)
	if got.Count != 100 || !near(got.Average, 0.92) || !near(got.Sum, 92) {
		t.Fatalf("count %d, average %v, knowledge %v", got.Count, got.Average, got.Sum)
	}
	if len(got.Buckets) != 20 || got.Buckets[19].Count != 20 || got.Buckets[18].Count != 80 || got.Buckets[18].Lo != 0.9 {
		t.Fatalf("buckets = %+v", got.Buckets[17:])
	}
	if empty := Retrievability(testData()); empty.Count != 0 || empty.Average != 0 || len(empty.Buckets) != 20 {
		t.Fatalf("empty = %+v", empty)
	}
}

func TestDifficultyAndStability(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeCollection})
	// S1 (s 30, d 5), V1 (s 4, d 7), F1 (s 5, d 3); N1 is new.
	diff := Difficulty(d)
	if diff.Count != 3 || !near(diff.Average, 4.0/9) {
		t.Fatalf("difficulty: %d cards, average %v", diff.Count, diff.Average)
	}
	for i, want := range map[int]int{4: 1, 8: 1, 13: 1} { // 2/9, 4/9, 6/9
		if diff.Buckets[i].Count != want {
			t.Errorf("difficulty bucket %d = %+v", i, diff.Buckets[i])
		}
	}
	st := Stability(d)
	if st.Count != 3 || !near(st.Average, 13) || len(st.Buckets) != 16 || st.Buckets[0].Hi != 2 {
		t.Fatalf("stability = %+v", st)
	}
	if st.Buckets[2].Count != 2 || st.Buckets[15].Count != 1 {
		t.Fatalf("stability buckets = %+v", st.Buckets)
	}
}

func TestIntervals(t *testing.T) {
	d := testData()
	d.review(3, 1, 1)
	d.review(2, 1, 7)
	d.review(1, 1, 21)
	got := Intervals(d)
	if got.Count != 6 || len(got.Buckets) != 21 || got.Buckets[0] != (HistogramBucket{1, 2, 3}) || got.Buckets[6].Count != 2 || got.Buckets[20] != (HistogramBucket{21, 22, 1}) {
		t.Fatalf("intervals = %+v", got)
	}
	// One outlier above the 99th percentile goes into the last bucket
	// instead of stretching the buckets.
	d = testData()
	d.review(150, 1, 10)
	d.review(1, 1, 3650)
	got = Intervals(d)
	last := got.Buckets[len(got.Buckets)-1]
	if len(got.Buckets) != 10 || last.Count != 151 || last.Hi != 3650 {
		t.Fatalf("buckets = %+v", got.Buckets)
	}
	d = testData()
	d.review(100, 1, 1)
	d.review(100, 1, 2000)
	if got := Intervals(d); len(got.Buckets) != 23 || got.Buckets[0].Count != 100 || got.Buckets[0].Hi != 91 || got.Buckets[22].Count != 100 {
		t.Fatalf("long intervals: %+v", got.Buckets)
	}
	if empty := Intervals(testData()); empty.Buckets == nil || len(empty.Buckets) != 0 {
		t.Fatal("empty histogram must have an empty bucket list")
	}
}

func TestTrueRetention(t *testing.T) {
	// Compare with target (spec): last month's mature reviews 180 passed,
	// 20 failed, desired retention 0.90.
	d := testData()
	d.answer(150, 10, Mature, scheduler.Good, 30, 0)
	d.answer(30, 0, Mature, scheduler.Easy, 30, 0)
	d.answer(20, 1, Mature, scheduler.Again, 30, 0)
	d.answer(4, 3, Young, scheduler.Hard, 5, 0)
	d.answer(1, 3, Young, scheduler.Again, 5, 0)
	d.answer(9, 40, Mature, scheduler.Again, 30, 0)
	d.answer(5, 0, Learning, scheduler.Again, 0, 0)
	d.answer(5, 0, Relearning, scheduler.Again, 0, 0)
	d.answer(5, 0, Filtered, scheduler.Good, 30, 0)
	got := TrueRetention(d)
	if got.Desired != 0.9 || len(got.Rows) != 6 {
		t.Fatalf("desired %v, %d rows", got.Desired, len(got.Rows))
	}
	month := got.Rows[3]
	if month.Period != RetentionMonth || month.Mature.Passed != 180 || month.Mature.Failed != 20 || *month.Mature.Retention != 0.9 {
		t.Fatalf("month = %+v", month)
	}
	if y := month.Young; y.Passed != 4 || y.Failed != 1 || *y.Retention != 0.8 {
		t.Fatalf("month young = %+v", y)
	}
	if tot := month.Total; tot.Passed != 184 || tot.Failed != 21 {
		t.Fatalf("month total = %+v", tot)
	}
	rows := map[string]RetentionRow{}
	for _, r := range got.Rows {
		rows[r.Period] = r
	}
	if r := rows[RetentionToday]; r.Mature.Passed != 30 || r.Mature.Failed != 0 || r.Young.Retention != nil {
		t.Fatalf("today = %+v", r)
	}
	if r := rows[RetentionYesterday]; r.Mature.Failed != 20 || r.Total.Passed != 0 {
		t.Fatalf("yesterday = %+v", r)
	}
	if r := rows[RetentionWeek]; r.Total.Passed != 34 || r.Total.Failed != 21 {
		t.Fatalf("week = %+v", r)
	}
	if r := rows[RetentionAll]; r.Mature.Failed != 29 || rows[RetentionYear].Mature.Failed != 29 {
		t.Fatalf("all = %+v", r)
	}
}
