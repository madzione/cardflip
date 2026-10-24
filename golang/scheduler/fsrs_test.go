package scheduler

import (
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

func reviewCard(stability, difficulty float64, lastReviewDaysAgo int, lapses int) CardState {
	return CardState{
		State: Review, Memory: &Memory{Stability: stability, Difficulty: difficulty},
		ScheduledDays: lastReviewDaysAgo, Due: t0, Reps: 5, Lapses: lapses,
		LastReview: t0.AddDate(0, 0, -lastReviewDaysAgo),
	}
}

func TestHigherRetentionShorterInterval(t *testing.T) {
	// Higher retention means shorter intervals (spec).
	card := reviewCard(20, 5, 20, 0)
	hi := Answer(card, Good, t0, DefaultParams(0.95), 42)
	lo := Answer(card, Good, t0, DefaultParams(0.85), 42)
	if hi.Interval >= lo.Interval {
		t.Fatalf("0.95 → %v, 0.85 → %v", hi.Interval, lo.Interval)
	}
	t.Logf("Good at 0.95: %dd, at 0.85: %dd", hi.State.ScheduledDays, lo.State.ScheduledDays)
}

func TestRatingsAreOrdered(t *testing.T) {
	// Ratings are ordered (spec), for every kind of card.
	cards := map[string]CardState{
		"new":             {State: New},
		"learning step 0": {State: Learning, Step: 0, Memory: &Memory{Stability: 0.4, Difficulty: 7}, LastReview: t0.Add(-time.Minute), Reps: 1},
		"learning step 1": {State: Learning, Step: 1, Memory: &Memory{Stability: 3, Difficulty: 5}, LastReview: t0.Add(-10 * time.Minute), Reps: 1},
		"relearning":      {State: Relearning, Memory: &Memory{Stability: 2, Difficulty: 8}, LastReview: t0.Add(-10 * time.Minute), Reps: 6, Lapses: 1},
		"young review":    reviewCard(3, 5, 3, 0),
		"mature review":   reviewCard(120, 4, 110, 0),
		"overdue review":  reviewCard(10, 6, 60, 2),
		"hard card":       reviewCard(5, 9.5, 5, 4),
	}
	for name, card := range cards {
		for seed := int64(0); seed < 50; seed++ {
			for _, ret := range []float64{0.7, 0.8, 0.9, 0.99} {
				out := Preview(card, t0, DefaultParams(ret), seed)
				for i := 1; i < 4; i++ {
					if out[i].Interval < out[i-1].Interval {
						t.Fatalf("%s (seed %d, retention %v): %v < %v at rating %d", name, seed, ret, out[i].Interval, out[i-1].Interval, i+1)
					}
				}
				if out[3].Interval <= out[0].Interval {
					t.Fatalf("%s: Easy %v not above Again %v", name, out[3].Interval, out[0].Interval)
				}
			}
		}
	}
}

func TestLapse(t *testing.T) {
	// Forgotten review card (spec): 3 lapses, Relearning, due the same day.
	card := reviewCard(30, 5, 30, 2)
	out := Answer(card, Again, t0, DefaultParams(0.9), 1)
	if out.State.Lapses != 3 || out.State.State != Relearning {
		t.Fatalf("after Again: %+v", out.State)
	}
	if out.Interval > 12*time.Hour || out.State.Due.Day() != t0.Day() {
		t.Fatalf("due %v (in %v)", out.State.Due, out.Interval)
	}
	if out.State.Memory.Stability >= 30 {
		t.Fatalf("stability did not drop: %v", out.State.Memory.Stability)
	}
	// Good on the relearning step graduates again.
	back := Answer(out.State, Good, out.State.Due, DefaultParams(0.9), 1)
	if back.State.State != Review || back.State.ScheduledDays < 1 {
		t.Fatalf("after relearning Good: %+v", back.State)
	}
}

func TestNewCardGoodTwiceGraduates(t *testing.T) {
	p := DefaultParams(0.9)
	card := CardState{State: New}
	first := Answer(card, Good, t0, p, 7)
	if first.State.State != Learning || !fuzzedStep(first.Interval, 10*time.Minute) || first.State.Memory == nil {
		t.Fatalf("after first Good: %+v in %v", first.State, first.Interval)
	}
	// New card rated Good twice (spec): Review, due at least a day away.
	second := Answer(first.State, Good, first.State.Due, p, 7)
	if second.State.State != Review || second.Interval < 24*time.Hour || second.State.ScheduledDays < 1 {
		t.Fatalf("after second Good: %+v in %v", second.State, second.Interval)
	}
	if second.State.Reps != 2 {
		t.Fatalf("reps = %d", second.State.Reps)
	}
}

// fuzzedStep reports whether d is step plus its step fuzz: up to 25 %, at
// most 5 minutes.
func fuzzedStep(d, step time.Duration) bool {
	return d >= step && d <= step+min(step/4, 5*time.Minute)
}

func TestLearningSteps(t *testing.T) {
	p := DefaultParams(0.9)
	card := CardState{State: New}
	if o := Answer(card, Again, t0, p, 1); o.State.State != Learning || !fuzzedStep(o.Interval, time.Minute) {
		t.Fatalf("Again: %+v %v", o.State, o.Interval)
	}
	if o := Answer(card, Hard, t0, p, 1); o.State.State != Learning || !fuzzedStep(o.Interval, 5*time.Minute+30*time.Second) {
		t.Fatalf("Hard: %+v %v", o.State, o.Interval)
	}
	// Rating Easy graduates immediately (spec).
	if o := Answer(card, Easy, t0, p, 1); o.State.State != Review || o.State.ScheduledDays < 1 {
		t.Fatalf("Easy: %+v", o.State)
	}
	// Again in a later step goes back to the first step.
	step1 := Answer(card, Good, t0, p, 1).State
	if o := Answer(step1, Again, t0.Add(10*time.Minute), p, 1); o.State.Step != 0 || !fuzzedStep(o.Interval, time.Minute) {
		t.Fatalf("Again from step 1: %+v %v", o.State, o.Interval)
	}
	if o := Answer(step1, Hard, t0.Add(10*time.Minute), p, 1); o.State.Step != 1 || !fuzzedStep(o.Interval, 15*time.Minute) {
		t.Fatalf("Hard on last step: %+v %v", o.State, o.Interval)
	}
}

func TestCustomSteps(t *testing.T) {
	// Custom steps (spec): 1m 10m 1h, Good twice → about an hour, learning.
	p := DefaultParams(0.9)
	p.LearningSteps = []time.Duration{time.Minute, 10 * time.Minute, time.Hour}
	first := Answer(CardState{State: New}, Good, t0, p, 3)
	second := Answer(first.State, Good, first.State.Due, p, 3)
	if second.State.State != Learning || second.State.Step != 2 || !fuzzedStep(second.Interval, time.Hour) {
		t.Fatalf("after second Good: %+v in %v", second.State, second.Interval)
	}
}

func TestStepFuzz(t *testing.T) {
	// Step fuzz bounded (spec): Good onto a 10-minute step is due 10 to
	// 12.5 minutes later, and the preview shows that delay.
	p := DefaultParams(0.9)
	seen := map[time.Duration]bool{}
	for seed := int64(0); seed < 100; seed++ {
		preview := Preview(CardState{State: New}, t0, p, seed)
		out := Answer(CardState{State: New}, Good, t0, p, seed)
		if out.Interval < 10*time.Minute || out.Interval > 12*time.Minute+30*time.Second {
			t.Fatalf("seed %d: Good due in %v", seed, out.Interval)
		}
		if out.Interval != preview[Good-1].Interval || !out.State.Due.Equal(t0.Add(out.Interval)) {
			t.Fatalf("seed %d: result %v, preview %v", seed, out.Interval, preview[Good-1].Interval)
		}
		seen[out.Interval] = true
	}
	if len(seen) < 10 {
		t.Fatalf("step fuzz produced only %d delays", len(seen))
	}
	// Steps under a minute get no fuzz; long steps at most 5 minutes.
	p.LearningSteps = []time.Duration{30 * time.Second, 2 * time.Hour}
	for seed := int64(0); seed < 50; seed++ {
		if o := Answer(CardState{State: New}, Again, t0, p, seed); o.Interval != 30*time.Second {
			t.Fatalf("30s step fuzzed to %v", o.Interval)
		}
		if o := Answer(CardState{State: New}, Good, t0, p, seed); o.Interval < 2*time.Hour || o.Interval > 2*time.Hour+5*time.Minute {
			t.Fatalf("2h step fuzzed to %v", o.Interval)
		}
	}
}

func TestEmptyStepsUseFSRSShortTerm(t *testing.T) {
	p := DefaultParams(0.9)
	p.LearningSteps, p.RelearningSteps = nil, nil
	for seed := int64(0); seed < 20; seed++ {
		card := CardState{State: New}
		now := t0
		// Ordered delays: Again ≤ Hard ≤ Good ≤ Easy, Easy graduates.
		for range 3 {
			preview := Preview(card, now, p, seed)
			for i := 1; i < 4; i++ {
				if preview[i].Interval < preview[i-1].Interval {
					t.Fatalf("seed %d, %v: rating %d %v < %v", seed, card.State, i+1, preview[i].Interval, preview[i-1].Interval)
				}
			}
			if preview[Easy-1].State.State != Review {
				t.Fatalf("Easy did not graduate: %+v", preview[Easy-1].State)
			}
			if preview[Again-1].State.State != Learning || preview[Again-1].Interval >= 24*time.Hour {
				t.Fatalf("Again: %+v in %v", preview[Again-1].State, preview[Again-1].Interval)
			}
			// Preview equals answer.
			for r := Again; r <= Easy; r++ {
				if out := Answer(card, r, now, p, seed); out.Interval != preview[r-1].Interval || out.State.State != preview[r-1].State.State {
					t.Fatalf("rating %v: answer %v, preview %v", r, out.Interval, preview[r-1].Interval)
				}
			}
			out := Answer(card, Hard, now, p, seed)
			card, now = out.State, out.State.Due
		}
		// Eventual graduation: Good graduates with at least a day.
		out := Answer(card, Good, now, p, seed)
		if out.State.State != Review || out.State.ScheduledDays < 1 {
			t.Fatalf("Good after Hard: %+v", out.State)
		}
	}
	// Relearning without steps: Again on a review card relearns briefly.
	out := Answer(reviewCard(30, 5, 30, 2), Again, t0, p, 1)
	if out.State.State != Relearning || out.State.Lapses != 3 || out.Interval >= time.Hour {
		t.Fatalf("lapse without relearning steps: %+v in %v", out.State, out.Interval)
	}
	if back := Answer(out.State, Good, out.State.Due, p, 1); back.State.State != Review || back.State.ScheduledDays < 1 {
		t.Fatalf("relearning Good: %+v", back.State)
	}
}

func TestInterdaySteps(t *testing.T) {
	loc := time.FixedZone("X", 2*3600)
	at := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, loc) }
	p := DefaultParams(0.9)
	p.LearningSteps = []time.Duration{10 * time.Minute, 24 * time.Hour}
	// One-day step (spec): Good on the 10-minute step at 15:00 → due from
	// the start of the next study day.
	onStep0 := Answer(CardState{State: New}, Again, at(5, 14, 50), p, 1).State
	out := Answer(onStep0, Good, at(5, 15, 0), p, 1)
	if out.State.State != Learning || !out.State.Interday || out.State.Step != 1 {
		t.Fatalf("after Good: %+v", out.State)
	}
	if !out.State.Due.Equal(at(6, 4, 0)) || out.State.ScheduledDays != 1 || out.Interval != 24*time.Hour {
		t.Fatalf("due %v, %d days, interval %v", out.State.Due, out.State.ScheduledDays, out.Interval)
	}
	if prev := Preview(onStep0, at(5, 15, 0), p, 1)[Good-1]; prev.Interval != out.Interval || !prev.State.Due.Equal(out.State.Due) {
		t.Fatalf("preview %+v", prev)
	}
	// Answering it again leaves interday learning.
	again := Answer(out.State, Again, at(6, 9, 0), p, 1)
	if again.State.Interday || again.State.State != Learning || again.State.ScheduledDays != 0 {
		t.Fatalf("Again on interday step: %+v", again.State)
	}
	if grad := Answer(out.State, Good, at(6, 9, 0), p, 1); grad.State.Interday || grad.State.State != Review {
		t.Fatalf("Good on last step: %+v", grad.State)
	}
	// A 1h step at 23:30 with rollover 04:00 ends in the same study day.
	p.LearningSteps = []time.Duration{time.Hour}
	late := Answer(CardState{State: New}, Again, at(5, 23, 30), p, 1)
	if late.State.Interday || !fuzzedStep(late.Interval, time.Hour) || late.State.Due.Day() != 6 {
		t.Fatalf("1h at 23:30: %+v in %v", late.State, late.Interval)
	}
	// The same step at 03:30 ends after the rollover: due that study day.
	early := Answer(CardState{State: New}, Again, at(6, 3, 30), p, 1)
	if !early.State.Interday || !early.State.Due.Equal(at(6, 4, 0)) || early.State.ScheduledDays != 1 {
		t.Fatalf("1h at 03:30: %+v", early.State)
	}
}

func TestPreviewEqualsResult(t *testing.T) {
	// Preview is side-effect free (spec): the Good result equals its preview.
	card := reviewCard(15, 5, 14, 0)
	p := DefaultParams(0.9)
	before := card
	preview := Preview(card, t0, p, 99)
	if card.Memory != before.Memory || *card.Memory != *before.Memory || card.Reps != before.Reps {
		t.Fatal("preview changed the card")
	}
	result := Answer(card, Good, t0, p, 99)
	if result.Interval != preview[Good-1].Interval || result.State.ScheduledDays != preview[Good-1].State.ScheduledDays {
		t.Fatalf("result %v, preview %v", result.Interval, preview[Good-1].Interval)
	}
}

func TestFuzzIsDeterministicAndSpreads(t *testing.T) {
	card := reviewCard(40, 5, 40, 0)
	p := DefaultParams(0.9)
	seen := map[int]bool{}
	for seed := int64(0); seed < 100; seed++ {
		a := Answer(card, Good, t0, p, seed).State.ScheduledDays
		b := Answer(card, Good, t0, p, seed).State.ScheduledDays
		if a != b {
			t.Fatal("fuzz not deterministic")
		}
		seen[a] = true
	}
	if len(seen) < 3 {
		t.Fatalf("fuzz produced only %v", seen)
	}
}

func TestDayBoundary(t *testing.T) {
	loc := time.FixedZone("X", 2*3600)
	at := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, loc) }
	// Late-night review counts for the previous day (spec).
	if got := DayStart(at(6, 2, 30), 4); !got.Equal(at(5, 4, 0)) {
		t.Fatalf("day start of 02:30 = %v", got)
	}
	if DayNumber(at(6, 2, 30), 4) != DayNumber(at(5, 4, 0), 4) {
		t.Fatal("02:30 is not in the previous study day")
	}
	if DayNumber(at(6, 4, 0), 4) != DayNumber(at(5, 4, 0), 4)+1 {
		t.Fatal("04:00 does not start a new study day")
	}
	if got := NextDayStart(at(6, 2, 30), 4); !got.Equal(at(6, 4, 0)) {
		t.Fatalf("next day start = %v", got)
	}
	// With rollover 6, 04:00 still belongs to the previous day.
	if DayNumber(at(6, 4, 0), 6) != DayNumber(at(5, 12, 0), 6) {
		t.Fatal("rollover hour not applied")
	}
	if got := DayStart(at(6, 0, 0), 0); !got.Equal(at(6, 0, 0)) {
		t.Fatalf("midnight rollover = %v", got)
	}
}

func TestMaximumInterval(t *testing.T) {
	// Capped interval (spec): with a maximum of 180 days, Easy on a mature
	// card shows and gives 180 days, whatever the fuzz.
	card := reviewCard(400, 3, 380, 0)
	if days := Preview(card, t0, DefaultParams(0.9), 1)[3].State.ScheduledDays; days <= 180 {
		t.Fatalf("uncapped Easy is only %d days", days)
	}
	p := DefaultParams(0.9)
	p.MaximumInterval = 180
	for seed := int64(0); seed < 50; seed++ {
		out := Preview(card, t0, p, seed)
		if easy := Answer(card, Easy, t0, p, seed); easy.Interval != out[3].Interval || easy.State.ScheduledDays != 180 || easy.Interval != 180*24*time.Hour {
			t.Fatalf("seed %d: Easy %+v, preview %+v", seed, easy.State, out[3].State)
		}
		for i, o := range out[1:] {
			if o.State.ScheduledDays > 180 {
				t.Fatalf("seed %d, rating %d: %d days", seed, i+2, o.State.ScheduledDays)
			}
		}
	}
	// Below the fuzz range, Easy's "Good + 1" is capped too.
	p.MaximumInterval = 1
	for _, card := range []CardState{{State: New}, reviewCard(3, 5, 3, 0)} {
		for i, o := range Preview(card, t0, p, 1) {
			if o.State.ScheduledDays > 1 {
				t.Fatalf("%v, rating %d: %d days", card.State, i+1, o.State.ScheduledDays)
			}
		}
	}
}

func TestMinimumRelearnInterval(t *testing.T) {
	// Relearned card (spec): with a minimum of 3 days, a relearning card
	// whose computed interval is 1 day is due in 3 days.
	card := CardState{State: Relearning, Memory: &Memory{Stability: 0.5, Difficulty: 8}, LastReview: t0.Add(-10 * time.Minute), Reps: 6, Lapses: 1}
	if days := Answer(card, Good, t0, DefaultParams(0.9), 1).State.ScheduledDays; days != 1 {
		t.Fatalf("computed interval is %d days, want 1", days)
	}
	p := DefaultParams(0.9)
	p.MinimumRelearnInterval = 3
	noSteps := p
	noSteps.RelearningSteps = nil
	for name, p := range map[string]Params{"steps": p, "no steps": noSteps} {
		for seed := int64(0); seed < 50; seed++ {
			out := Preview(card, t0, p, seed)
			for _, r := range []Rating{Good, Easy} {
				o := Answer(card, r, t0, p, seed)
				if o.Interval != out[r-1].Interval || o.State.State != Review || o.State.ScheduledDays < 3 {
					t.Fatalf("%s, seed %d, rating %d: %+v", name, seed, r, o.State)
				}
			}
			if out[2].State.ScheduledDays > out[3].State.ScheduledDays {
				t.Fatalf("%s, seed %d: Good %d > Easy %d", name, seed, out[2].State.ScheduledDays, out[3].State.ScheduledDays)
			}
			if out[0].State.State != Relearning {
				t.Fatalf("%s, seed %d: Again graduated: %+v", name, seed, out[0].State)
			}
		}
	}
	// With fuzz: a longer interval stays fuzzed but never drops below the
	// minimum.
	p.MinimumRelearnInterval = 10
	seen := map[int]bool{}
	for seed := int64(0); seed < 100; seed++ {
		days := Answer(card, Good, t0, p, seed).State.ScheduledDays
		if days < 10 {
			t.Fatalf("seed %d: %d days", seed, days)
		}
		seen[days] = true
	}
	if len(seen) < 2 {
		t.Fatalf("no fuzz: %v", seen)
	}
	// Review cards are not affected.
	if days := Answer(reviewCard(0.5, 8, 1, 0), Hard, t0, p, 1).State.ScheduledDays; days >= 10 {
		t.Fatalf("review card got the relearning minimum: %d days", days)
	}
}

func TestRetrievability(t *testing.T) {
	// At the interval for 90% retention, recall probability is 90%.
	card := reviewCard(10, 5, 0, 0)
	p := DefaultParams(0.9)
	if r, ok := Retrievability(card, card.LastReview); !ok || r != 1 {
		t.Fatalf("R right after the review = %v %v", r, ok)
	}
	due := card.LastReview.AddDate(0, 0, int(intervalDays(10, p)))
	if r, ok := Retrievability(card, due); !ok || r < 0.89 || r > 0.91 {
		t.Fatalf("R at due = %v %v", r, ok)
	}
	if r, _ := Retrievability(card, due.AddDate(0, 0, 30)); r >= 0.89 {
		t.Fatalf("R after 30 more days = %v", r)
	}
	if _, ok := Retrievability(CardState{State: Review, Due: t0}, t0); ok {
		t.Fatal("card without memory has a retrievability")
	}
}

func sundayMinimum() Params {
	p := DefaultParams(0.9)
	p.EasyDays[6] = EasyMinimum
	return p
}

func TestEasyDaysAvoidMinimumDay(t *testing.T) {
	// Avoid a minimum day (spec): t0 is a Monday, so 4 to 7 days from it
	// span Friday to Monday.
	p := sundayMinimum()
	for i := range 100 {
		if got := p.easyDay(4, 7, t0, float64(i)/100); got == 6 {
			t.Fatalf("u=%v chose Sunday", float64(i)/100)
		}
	}
	// The same through answers: no Good interval lands on a Sunday.
	card := reviewCard(40, 5, 40, 0)
	for seed := int64(0); seed < 200; seed++ {
		o := Answer(card, Good, t0, p, seed)
		if o.State.Due.Weekday() == time.Sunday {
			t.Fatalf("seed %d: due on Sunday (%dd)", seed, o.State.ScheduledDays)
		}
	}
}

func TestEasyDaysOnlyMinimumDay(t *testing.T) {
	// Only a minimum day (spec): a range of only Sunday keeps Sunday.
	p := sundayMinimum()
	for _, u := range []float64{0, 0.5, 0.99} {
		if got := p.easyDay(6, 6, t0, u); got != 6 {
			t.Fatalf("u=%v: %d days", u, got)
		}
	}
}

func TestEasyDaysAllNormalUnchanged(t *testing.T) {
	// No Easy Days set (spec): due days are chosen as before, whatever the
	// load.
	p := DefaultParams(0.9)
	loaded := p
	loaded.DueLoad = func(day int) int { return day % 5 * 100 }
	card := reviewCard(40, 5, 40, 0)
	for seed := int64(0); seed < 100; seed++ {
		want := t0.AddDate(0, 0, int(fuzz(intervalDays(memoryAfter(card, Good, t0, nil).Stability, p), 1, card, seed)))
		if got := Answer(card, Good, t0, loaded, seed).State.Due; !got.Equal(want) {
			t.Fatalf("seed %d: due %v, want %v", seed, got, want)
		}
	}
}

func TestEasyDaysReducedAndLoad(t *testing.T) {
	// Reduced days get fewer reviews than Normal days.
	p := DefaultParams(0.9)
	p.EasyDays[5], p.EasyDays[6] = EasyReduced, EasyReduced
	counts := map[time.Weekday]int{}
	for i := range 1000 {
		counts[t0.AddDate(0, 0, p.easyDay(4, 7, t0, float64(i)/1000)).Weekday()]++
	}
	if counts[time.Saturday] >= counts[time.Friday] || counts[time.Sunday] >= counts[time.Monday] {
		t.Fatalf("picks per weekday: %v", counts)
	}
	// Days with more reviews due are chosen less often.
	friday := DayNumber(t0.AddDate(0, 0, 4), p.RolloverHour)
	p.DueLoad = func(day int) int {
		if day == friday {
			return 50
		}
		return 0
	}
	busy := 0
	for i := range 1000 {
		if p.easyDay(4, 7, t0, float64(i)/1000) == 4 {
			busy++
		}
	}
	if busy >= counts[time.Friday]/5 {
		t.Fatalf("busy Friday chosen %d times, idle %d", busy, counts[time.Friday])
	}
}

func TestEasyDaysPreviewEqualsResult(t *testing.T) {
	p := sundayMinimum()
	p.EasyDays[5] = EasyReduced
	p.DueLoad = func(day int) int { return day % 7 }
	for seed := int64(0); seed < 50; seed++ {
		card := reviewCard(30, 5, 30, 0)
		preview := Preview(card, t0, p, seed)
		for r := Hard; r <= Easy; r++ {
			if got := Answer(card, r, t0, p, seed); got.State.Due != preview[r-1].State.Due {
				t.Fatalf("seed %d %v: result %v, preview %v", seed, r, got.State.Due, preview[r-1].State.Due)
			}
		}
	}
}

func TestEasyDaysUseStudyDayWeekday(t *testing.T) {
	// A review at 02:00 on a Monday belongs to Sunday's study day, and so
	// does a due time 7 days later.
	p := sundayMinimum()
	now := time.Date(2026, 10, 5, 2, 0, 0, 0, time.UTC)
	if got := p.easyDay(6, 8, now, 0.99); got == 7 {
		t.Fatal("chose the Sunday study day")
	}
}
