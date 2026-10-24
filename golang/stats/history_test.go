package stats

import (
	"slices"
	"testing"

	"cardflip/golang/scheduler"
)

func TestCalendar(t *testing.T) {
	// Calendar (spec): reviews on 20 of the last 30 days, more on some.
	d := testData()
	for back := range 30 {
		if back%3 != 2 {
			d.answer(1+back%4, back, Young, scheduler.Good, 5, 1000)
		}
	}
	d.answer(3, calendarDays, Young, scheduler.Good, 5, 1000) // over a year ago
	got := ReviewCalendar(d)
	if len(got.Days) != 20 || got.Today != "2026-10-13" || got.Max != 4 {
		t.Fatalf("calendar: %d days, today %s, max %d", len(got.Days), got.Today, got.Max)
	}
	last := got.Days[len(got.Days)-1]
	if last != (CalendarDay{Day: 0, Count: 1}) || got.Days[len(got.Days)-2] != (CalendarDay{Day: -1, Count: 2}) {
		t.Fatalf("last days = %v", got.Days[len(got.Days)-2:])
	}
	if empty := ReviewCalendar(testData()); empty.Days == nil || len(empty.Days) != 0 {
		t.Fatalf("empty calendar = %+v", empty)
	}
}

func TestReviewsByDay(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeCollection})
	got := Reviews(d, PeriodMonth)
	if got.Unit != UnitDay || len(got.Buckets) != 30 || got.PeriodDays != 30 {
		t.Fatalf("unit %s, %d buckets, %d days", got.Unit, len(got.Buckets), got.PeriodDays)
	}
	// S1's September answers are 30 days ago, just outside the month.
	if got.Total != 6 || got.DaysStudied != 3 {
		t.Fatalf("total %d, days studied %d", got.Total, got.DaysStudied)
	}
	want := map[int]ReviewBucket{
		29: {Start: 0, Mature: 1, Relearning: 1, Filtered: 1},
		28: {Start: -1, Learning: 1, Young: 1},
		17: {Start: -12, Learning: 1},
	}
	for i, b := range got.Buckets {
		w, ok := want[i]
		if !ok {
			w = ReviewBucket{Start: i - 29}
		}
		if b != w {
			t.Errorf("bucket %d = %+v, want %+v", i, b, w)
		}
	}
	rt := ReviewTime(d, PeriodMonth)
	if rt.Total != 29000 || rt.Buckets[29].Mature != 8000 || rt.Buckets[29].Filtered != 1000 {
		t.Fatalf("review time: total %d, today %+v", rt.Total, rt.Buckets[29])
	}
}

func TestReviewsByWeekAndMonth(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeCollection})
	year := Reviews(d, PeriodYear)
	if year.Unit != UnitWeek || len(year.Buckets) != 53 || year.Total != 8 || year.DaysStudied != 4 || year.PeriodDays != 365 {
		t.Fatalf("year: %s, %d buckets, total %d, %d days studied", year.Unit, len(year.Buckets), year.Total, year.DaysStudied)
	}
	last := year.Buckets[52]
	if last != (ReviewBucket{Start: -6, Learning: 1, Young: 1, Mature: 1, Relearning: 1, Filtered: 1}) {
		t.Fatalf("this week = %+v", last)
	}
	if b := year.Buckets[51]; b != (ReviewBucket{Start: -13, Learning: 1}) {
		t.Fatalf("last week = %+v", b)
	}
	if b := year.Buckets[48]; b.Learning != 2 || b.Start != -34 {
		t.Fatalf("Sep 13's week = %+v", b)
	}
	if first := year.Buckets[0]; first.Start != -364 {
		t.Fatalf("first week starts %d days ago, want the period's start", -first.Start)
	}

	all := Reviews(d, PeriodAll)
	if all.Unit != UnitMonth || len(all.Buckets) != 2 || all.PeriodDays != 31 {
		t.Fatalf("all: %s, %d buckets, %d days", all.Unit, len(all.Buckets), all.PeriodDays)
	}
	if sep, oct := all.Buckets[0], all.Buckets[1]; sep != (ReviewBucket{Start: -42, Learning: 2}) || oct.Start != -12 ||
		oct.Learning != 2 || oct.Young != 1 || oct.Mature != 1 {
		t.Fatalf("months = %+v", all.Buckets)
	}
}

func TestHourly(t *testing.T) {
	f := newFixture(t)
	got := Hourly(f.load(t, Scope{Kind: ScopeCollection}), Period3Months)
	if len(got) != 24 {
		t.Fatalf("%d hours", len(got))
	}
	want := map[int]HourBucket{2: {1, 1}, 8: {1, 1}, 9: {2, 0.5}, 10: {3, 2.0 / 3}, 20: {1, 1}}
	for h, b := range got {
		if b != want[h] {
			t.Errorf("hour %d = %+v, want %+v", h, b, want[h])
		}
	}
	if month := Hourly(f.load(t, Scope{Kind: ScopeCollection}), PeriodMonth); month[10].Reviews != 1 {
		t.Fatalf("hour 10 in the last month = %+v", month[10])
	}
}

func TestAnswerButtons(t *testing.T) {
	f := newFixture(t)
	got := AnswerButtons(f.load(t, Scope{Kind: ScopeCollection}), Period3Months)
	check := func(name string, g ButtonGroup, counts []int, correct float64) {
		t.Helper()
		if !slices.Equal(g.Counts, counts) || g.Correct != correct || g.Total != counts[0]+counts[1]+counts[2]+counts[3] {
			t.Errorf("%s = %+v, want %v and %v correct", name, g, counts, correct)
		}
	}
	check("learning", got.Learning, []int{1, 0, 4, 0}, 0.8)
	check("young", got.Young, []int{0, 0, 2, 0}, 1)
	check("mature", got.Mature, []int{1, 0, 0, 0}, 0)
	empty := AnswerButtons(testData(), PeriodMonth)
	check("empty", empty.Mature, []int{0, 0, 0, 0}, 0)
}
