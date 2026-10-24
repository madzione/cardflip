package core

import (
	"fmt"
	"math"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

func TestRowColumns(t *testing.T) {
	c, _, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	nid, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic").ID, DeckID: spanish,
		Fields: []string{"<b>comer</b>  &amp; beber", "to eat<br>and drink"}, Tags: []string{"verb", "food"}})
	if err != nil {
		t.Fatal(err)
	}
	card := c.CardsOfNote(nid)[0].ID
	added := clk.t
	cloze := c.CardsOfNote(addNote(t, c, "Cloze", "{{c1::uno}} {{c2::dos}}", ""))[1].ID

	// A New card: no interval, memory state or due time.
	row := c.Rows([]ID{card})[0]
	want := Row{CardID: card, NoteID: nid, State: scheduler.New, Question: "comer & beber", Answer: "to eat and drink",
		SortField: "comer & beber", CardName: "Card 1", NoteType: "Basic", Deck: "Spanish", Tags: "verb food",
		Position: c.mustCard(t, card).NewPos, Interval: math.NaN(), Difficulty: math.NaN(), Stability: math.NaN(),
		Retrievability: math.NaN(), Created: time.UnixMilli(int64(card)), CardModified: added, NoteModified: added}
	if !sameRow(row, want) {
		t.Fatalf("new card row:\n got %+v\nwant %+v", row, want)
	}
	if row := c.Rows([]ID{cloze}, ColCardType)[0]; row.CardName != "Cloze 2" {
		t.Fatalf("cloze card type = %q", row.CardName)
	}

	// A Review card, changed an hour later.
	clk.advance(time.Hour)
	due := clk.t.AddDate(0, 0, 3)
	setState(t, c, []ID{card}, scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: 7, Reps: 9,
		Lapses: 2, LastReview: clk.t.AddDate(0, 0, -4), Memory: &scheduler.Memory{Stability: 12.5, Difficulty: 5.5}})
	row = c.Rows([]ID{card})[0]
	r, _ := scheduler.Retrievability(c.mustCard(t, card).CardState, clk.t)
	want.State, want.Position, want.Due, want.Interval = scheduler.Review, 0, due, 7
	want.Difficulty, want.Stability, want.Retrievability = 0.5, 12.5, r
	want.Reviews, want.Lapses, want.CardModified = 9, 2, clk.t
	if !sameRow(row, want) {
		t.Fatalf("review card row:\n got %+v\nwant %+v", row, want)
	}
	if r <= 0 || r >= 1 {
		t.Fatalf("retrievability = %v", r)
	}

	// Only the requested columns are filled.
	got := c.Rows([]ID{card}, ColDeck, ColLapses)[0]
	if got.Deck != "Spanish" || got.Lapses != 2 || got.Question != "" || got.Answer != "" || got.Reviews != 0 ||
		!got.Created.IsZero() || got.CardID != card || got.State != scheduler.Review {
		t.Fatalf("deck and lapses row = %+v", got)
	}

	// Editing the note updates the cached text and the note's modified
	// time, not the card's.
	clk.advance(time.Hour)
	if err := c.UpdateNote(nid, []string{"vivir", "<i>to live</i>"}, []string{"verb"}); err != nil {
		t.Fatal(err)
	}
	row = c.Rows([]ID{card})[0]
	if row.Question != "vivir" || row.Answer != "to live" || row.SortField != "vivir" || row.Tags != "verb" ||
		!row.NoteModified.Equal(clk.t) || !row.CardModified.Equal(want.CardModified) {
		t.Fatalf("row after edit = %+v", row)
	}

	// Long text is cut to 200 characters.
	long := c.CardsOfNote(addNote(t, c, "Basic", fmt.Sprintf("%0300d", 7), "ä"))[0].ID
	if row := c.Rows([]ID{long}, ColQuestion, ColSortField)[0]; len(row.Question) != 200 || len(row.SortField) != 200 {
		t.Fatalf("long question %d, sort field %d characters", len(row.Question), len(row.SortField))
	}
}

// sameRow compares rows, with NaN equal to NaN.
func sameRow(a, b Row) bool {
	nan := func(f float64) float64 {
		if math.IsNaN(f) {
			return -1
		}
		return f
	}
	for _, r := range []*Row{&a, &b} {
		r.Interval, r.Difficulty, r.Stability, r.Retrievability = nan(r.Interval), nan(r.Difficulty), nan(r.Stability), nan(r.Retrievability)
	}
	return a.Due.Equal(b.Due) && a.Created.Equal(b.Created) && a.CardModified.Equal(b.CardModified) &&
		a.NoteModified.Equal(b.NoteModified) && fmt.Sprint(a) == fmt.Sprint(b)
}

func TestAnswerText(t *testing.T) {
	for _, tc := range []struct{ question, answer, want string }{
		{"Front", "Front\n\n<hr id=answer>\n\nBack", "Back"},
		{"Front", `Front <hr ID="answer" class=x> <b>Back</b> text`, "Back text"},
		// Without the line, the repeated question is left out.
		{"Front", "<div>Front</div><div>Back</div>", "Back"},
		{"Front", "Back only", "Back only"},
		{"", "Back", "Back"},
	} {
		if got := answerText(tc.question, tc.answer); got != tc.want {
			t.Errorf("answerText(%q, %q) = %q, want %q", tc.question, tc.answer, got, tc.want)
		}
	}
}

func TestCardModified(t *testing.T) {
	c, _, clk := testCollection(t)
	card := c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0].ID
	added := clk.t
	if m := c.mustCard(t, card).Modified; !m.Equal(added) {
		t.Fatalf("modified after adding = %v", m)
	}
	// Changing a card stamps it; undo brings back the earlier time.
	clk.advance(time.Minute)
	if err := c.SetFlag([]ID{card}, 2); err != nil {
		t.Fatal(err)
	}
	if m := c.mustCard(t, card).Modified; !m.Equal(clk.t) {
		t.Fatalf("modified after flagging = %v, want %v", m, clk.t)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if m := c.mustCard(t, card).Modified; !m.Equal(added) {
		t.Fatalf("modified after undo = %v, want %v", m, added)
	}

	// Cards stored before the time was recorded show their latest review,
	// or their creation.
	old := &Card{ID: card}
	if got := cardModified(old); !got.Equal(time.UnixMilli(int64(card))) {
		t.Fatalf("unrecorded = %v", got)
	}
	old.LastReview = added.Add(time.Hour)
	if got := cardModified(old); !got.Equal(old.LastReview) {
		t.Fatalf("unrecorded, reviewed = %v", got)
	}
}

func TestSortByColumn(t *testing.T) {
	c, _, clk := testCollection(t)
	basic := noteType(t, c, "Basic").ID
	add := func(front, back string, tags ...string) ID {
		id, err := c.AddNote(NoteInput{TypeID: basic, DeckID: defaultDeck(c), Fields: []string{front, back}, Tags: tags})
		if err != nil {
			t.Fatal(err)
		}
		return c.CardsOfNote(id)[0].ID
	}
	a := add("one", "Cherry", "zeta")
	b := add("two", "apple")
	d := add("three", "<b>banana</b>", "alpha", "beta")
	e := add("four", "apple", "Alpha")
	review := func(id ID, ivl, lapses int) {
		setState(t, c, []ID{id}, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, ivl),
			ScheduledDays: ivl, Lapses: lapses, Memory: &scheduler.Memory{Stability: float64(ivl), Difficulty: 5}})
	}
	review(a, 10, 1)
	review(b, 3, 4)
	review(d, 30, 0)

	check := func(col Column, desc bool, want ...ID) {
		t.Helper()
		ids, err := c.Search("", col, desc)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(ids) != fmt.Sprint(want) {
			t.Errorf("%s desc=%v: %v, want %v", col, desc, ids, want)
		}
	}
	// Sort by lapses (spec): the card with the most lapses first.
	check(ColLapses, true, b, a, e, d)
	check(ColLapses, false, d, e, a, b)
	// Cards without an interval (New) come last either way.
	check(ColInterval, false, b, a, d, e)
	check(ColInterval, true, d, a, b, e)
	check(ColStability, true, d, a, b, e)
	// Text ignores case and HTML; equal values stay in creation order.
	check(ColAnswer, false, b, e, d, a)
	check(ColAnswer, true, a, d, e, b)
	// Notes without tags come last.
	check(ColTags, false, e, d, a, b)
	check(ColTags, true, a, d, e, b)
	check(ColSortField, false, e, a, d, b)

	// Every column sorts, both ways, without losing cards.
	for _, col := range Columns {
		for _, desc := range []bool{false, true} {
			if ids, _ := c.Search("", col, desc); len(ids) != 4 {
				t.Errorf("%s desc=%v: %v", col, desc, ids)
			}
		}
	}
}

// Sorting 50,000 cards by answer stays under 500 ms, the first time too,
// when every answer is rendered.
func TestSortByAnswer50k(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	for _, run := range []string{"first", "cached"} {
		start := time.Now()
		ids, err := c.Search("", ColAnswer, false)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s: %d cards in %v", run, len(ids), elapsed)
		if elapsed > 500*time.Millisecond && !raceDetector {
			t.Fatalf("%s sort by answer took %v", run, elapsed)
		}
	}
}

func BenchmarkSortByAnswer50k(b *testing.B) {
	c := largeCollection(b, 50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Search("", ColAnswer, false)
	}
}

// Sorting a large collection by text computes the keys in parallel.
func TestSortByQuestionConcurrent(t *testing.T) {
	c := largeCollection(t, 5000)
	ids, err := c.Search("", ColQuestion, true)
	if err != nil || len(ids) != 5000 {
		t.Fatalf("%d cards, %v", len(ids), err)
	}
	if row := c.Rows(ids[:1], ColQuestion)[0]; row.Question != "Question number 999 about topic 29" {
		t.Fatalf("first = %q", row.Question)
	}
}
