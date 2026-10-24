package stats

import (
	"slices"
	"testing"

	"cardflip/golang/core"
)

func cardIDs(d *Data) []core.ID {
	var out []core.ID
	for _, c := range d.Cards {
		out = append(out, c.ID)
	}
	return out
}

func reviewCards(d *Data) map[core.ID]int {
	out := map[core.ID]int{}
	for _, r := range d.Reviews {
		out[r.CardID]++
	}
	return out
}

func TestDeckScopeIncludesSubdecks(t *testing.T) {
	f := newFixture(t)
	// Deck scope (spec): statistics from the menu of Spanish cover Spanish
	// and its subdecks only.
	d := f.load(t, Scope{Kind: ScopeDeck, DeckID: int64(f.spanish)})
	if got := cardIDs(d); !slices.Equal(got, []core.ID{f.s1, f.v1}) {
		t.Fatalf("cards = %v, want S1 and V1", got)
	}
	if got := reviewCards(d); len(got) != 2 || got[f.s1] != 3 || got[f.v1] != 2 {
		t.Fatalf("reviews per card = %v (manual entry must be skipped)", got)
	}
	d = f.load(t, Scope{Kind: ScopeDeck, DeckID: int64(f.verbs)})
	if got := cardIDs(d); !slices.Equal(got, []core.ID{f.v1}) {
		t.Fatalf("subdeck cards = %v", got)
	}
}

func TestCollectionScopeIncludesDeletedCards(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeCollection})
	if len(d.Cards) != 4 {
		t.Fatalf("cards = %d, want 4 (D1 deleted)", len(d.Cards))
	}
	if len(d.Reviews) != 8 {
		t.Fatalf("reviews = %d, want 8 including the deleted card's", len(d.Reviews))
	}
	got := reviewCards(d)
	delete(got, f.s1)
	delete(got, f.v1)
	delete(got, f.f1)
	if len(got) != 1 {
		t.Fatalf("answers of deleted cards = %v, want one card", got)
	}
}

func TestSearchScope(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeSearch, Query: "deck:French"})
	if got := cardIDs(d); !slices.Equal(got, []core.ID{f.f1, f.n1}) {
		t.Fatalf("cards = %v", got)
	}
	if got := reviewCards(d); len(got) != 1 || got[f.f1] != 2 {
		t.Fatalf("reviews = %v; the deleted French card must not count", got)
	}
	var l Loader
	if _, err := l.Load(f.c, Scope{Kind: ScopeSearch, Query: "is:bogus"}); err == nil {
		t.Fatal("invalid search accepted")
	}
	if _, err := l.Load(f.c, Scope{Kind: ScopeDeck, DeckID: 999}); err == nil {
		t.Fatal("unknown deck accepted")
	}
}

func TestStudyDayBucketingAndKinds(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeDeck, DeckID: int64(f.spanish)})
	if d.Today != day(fixtureNow) {
		t.Fatalf("today = %d", d.Today)
	}
	var s1 []Review
	for _, r := range d.Reviews {
		if r.CardID == f.s1 {
			s1 = append(s1, r)
		}
	}
	slices.SortFunc(s1, func(a, b Review) int { return a.Day - b.Day })
	// The 02:30 answer belongs to the study day before.
	if s1[2].Day != d.Today-1 || s1[2].Hour != 2 || s1[2].Kind != Young {
		t.Fatalf("02:30 answer = %+v, want yesterday, hour 2, young", s1[2])
	}
	if s1[0].Kind != Learning || s1[1].Kind != Learning || s1[0].Day != day(at(9, 13, 10, 0)) {
		t.Fatalf("learning answers = %+v", s1[:2])
	}
	kinds := map[core.ID][]Kind{}
	for _, r := range f.load(t, Scope{Kind: ScopeCollection}).Reviews {
		kinds[r.CardID] = append(kinds[r.CardID], r.Kind)
	}
	if got := kinds[f.v1]; !slices.Equal(got, []Kind{Mature, Relearning}) {
		t.Fatalf("V1 kinds = %v", got)
	}
	if got := kinds[f.f1]; !slices.Equal(got, []Kind{Learning, Filtered}) {
		t.Fatalf("F1 kinds = %v", got)
	}
}

func TestPeriods(t *testing.T) {
	f := newFixture(t)
	d := f.load(t, Scope{Kind: ScopeCollection})
	for p, want := range map[string]int{PeriodMonth: 30, Period3Months: 90, PeriodYear: 365} {
		if got := d.Today - d.firstDay(p) + 1; got != want {
			t.Errorf("%s covers %d days, want %d", p, got, want)
		}
	}
	if got := d.firstDay(PeriodAll); got != day(at(9, 13, 10, 0)) {
		t.Errorf("all history starts on day %d, want the first answer's", got)
	}
	if !ValidPeriod(PeriodAll) || ValidPeriod("week") {
		t.Error("ValidPeriod")
	}
	if got := Date(d.Today); got != "2026-10-13" {
		t.Errorf("Date(today) = %s", got)
	}
}

func TestDesiredRetention(t *testing.T) {
	f := newFixture(t)
	opts, _ := f.c.DeckOptions(f.spanish)
	if d := f.load(t, Scope{Kind: ScopeDeck, DeckID: int64(f.spanish)}); d.Retention != opts.DesiredRetention {
		t.Fatalf("deck retention = %v", d.Retention)
	}
	if d := f.load(t, Scope{Kind: ScopeCollection}); d.Retention != core.DefaultRetention {
		t.Fatalf("collection retention = %v", d.Retention)
	}
}

func TestLoaderReusesReviewLogUntilChange(t *testing.T) {
	f := newFixture(t)
	var l Loader
	scope := Scope{Kind: ScopeCollection}
	a, err := l.Load(f.c, scope)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := l.Load(f.c, Scope{Kind: ScopeDeck, DeckID: int64(f.fr)})
	if &a.Reviews[0] != &l.all[0] || len(b.Reviews) != 2 {
		t.Fatal("review log not shared")
	}
	if _, err := f.c.Answer(f.n1, 3, 0); err != nil {
		t.Fatal(err)
	}
	c, _ := l.Load(f.c, scope)
	if len(c.Reviews) != len(a.Reviews)+1 {
		t.Fatalf("reviews after an answer = %d, want %d", len(c.Reviews), len(a.Reviews)+1)
	}
}
