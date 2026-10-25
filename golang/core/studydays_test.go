package core

import (
	"maps"
	"slices"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// importAnswers imports a note whose card has an answer at each of times,
// and a manual entry an hour after the last.
func importAnswers(t *testing.T, c *Collection, guid string, times []time.Time) {
	t.Helper()
	basic, _ := c.BasicNoteType()
	var history []ImportReview
	for _, tm := range times {
		history = append(history, ImportReview{Time: tm, Rating: scheduler.Good, StateBefore: scheduler.Review})
	}
	history = append(history, ImportReview{Time: times[len(times)-1].Add(time.Hour), StateBefore: scheduler.Review})
	b := &ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []ImportNote{{TypeKey: "basic", GUID: guid, DeckName: "Default", Fields: []string{guid, "a"},
			Cards: []ImportCard{{DeckName: "Default", History: history}}}},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
}

func TestStudyDays(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skip("no time zone data:", err)
	}
	c, _, _ := testCollection(t)
	if sd, err := c.StudyDays(DefaultRolloverHour, berlin); err != nil || len(sd.Days) != 0 {
		t.Fatalf("empty: %+v, %v", sd, err)
	}
	// Two answers a day across the autumn change, at 02:30 (the previous
	// study day) and 22:00, plus a manual entry.
	var times []time.Time
	var want []int
	for d := 20; d <= 30; d++ {
		times = append(times, time.Date(2026, 10, d, 2, 30, 0, 0, berlin), time.Date(2026, 10, d, 22, 0, 0, 0, berlin))
		want = append(want, scheduler.DayNumber(time.Date(2026, 10, d, 12, 0, 0, 0, berlin), DefaultRolloverHour))
	}
	want = append([]int{want[0] - 1}, want...) // Oct 20 02:30 is Oct 19
	importAnswers(t, c, "dst", times)
	sd, err := c.StudyDays(DefaultRolloverHour, berlin)
	if err != nil || !slices.Equal(sd.Days, want) {
		t.Fatalf("days %v, %v\nwant %v", sd.Days, err, want)
	}
	if sorted := c.studyDaysSorted(StudyDays{}, nil, DefaultRolloverHour, berlin); !slices.Equal(sorted.Days, want) {
		t.Fatalf("sorted days %v\nwant        %v", sorted.Days, want)
	}
	// With the rollover at midnight the 02:30 answers are on their own day.
	if sd, _ := c.StudyDays(0, berlin); !slices.Equal(sd.Days, want[1:]) {
		t.Fatalf("rollover 0: %v\nwant       %v", sd.Days, want[1:])
	}
	// An answer decades earlier spans more days than there are answers.
	importAnswers(t, c, "old", []time.Time{time.Date(1990, 1, 1, 12, 0, 0, 0, berlin)})
	old := scheduler.DayNumber(time.Date(1990, 1, 1, 12, 0, 0, 0, berlin), DefaultRolloverHour)
	if sd, _ := c.StudyDays(DefaultRolloverHour, berlin); !slices.Equal(sd.Days, append([]int{old}, want...)) {
		t.Fatalf("with 1990: %v", sd.Days)
	}
}

func TestReviewGeneration(t *testing.T) {
	c, _, clk := testCollection(t)
	importAnswers(t, c, "a", []time.Time{clk.now().Add(-48 * time.Hour)})
	card := c.AllCards()[0]
	sd, _ := c.StudyDays(DefaultRolloverHour, time.Local)
	// A new answer keeps the generation and is above Last.
	e, err := c.Answer(card.ID, scheduler.Good, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	answers, gen := c.RecentAnswers(sd.Last, clk.now().Add(time.Hour))
	if gen != sd.Gen || len(answers) != 1 || answers[0].ID != e.ID || answers[0].DurationMs != 1000 {
		t.Fatalf("after an answer: %+v, generation %d → %d", answers, sd.Gen, gen)
	}
	// Manual entries keep it; undo, redo and imports advance it.
	if err := c.SetDueDate([]ID{card.ID}, "5"); err != nil {
		t.Fatal(err)
	}
	if _, g := c.RecentAnswers(0, clk.now()); g != gen {
		t.Fatalf("set due date advanced the generation")
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, g := c.RecentAnswers(0, clk.now()); g != gen {
		t.Fatalf("undoing set due date advanced the generation")
	}
	if _, err := c.UndoReview(e.ID); err != nil {
		t.Fatal(err)
	}
	_, g1 := c.RecentAnswers(0, clk.now())
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	_, g2 := c.RecentAnswers(0, clk.now())
	importAnswers(t, c, "b", []time.Time{clk.now().Add(-72 * time.Hour)})
	_, g3 := c.RecentAnswers(0, clk.now())
	if !(gen < g1 && g1 < g2 && g2 < g3) {
		t.Fatalf("generations %d, %d, %d, %d", gen, g1, g2, g3)
	}
}

// importHistory imports a note with one card in state st and the answers
// in history, and returns the card.
func importHistory(t *testing.T, c *Collection, guid string, st scheduler.CardState, history ...ImportReview) ID {
	t.Helper()
	basic, _ := c.BasicNoteType()
	b := &ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []ImportNote{{TypeKey: "basic", GUID: guid, DeckName: "Default", Fields: []string{guid, "a"},
			Cards: []ImportCard{{DeckName: "Default", State: st, History: history}}}},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	for _, card := range c.AllCards() {
		if n, err := c.Note(card.NoteID); err == nil && n.Fields[0] == guid {
			return card.ID
		}
	}
	t.Fatalf("no card %s", guid)
	return 0
}

// checkHistory fails unless the study days of c have these answers and
// milliseconds per day, counted from day, and these mature days.
func checkHistory(t *testing.T, c *Collection, what string, day int, answers map[int]int, ms map[int]int64, mature []int) {
	t.Helper()
	sd, err := c.StudyDays(DefaultRolloverHour, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	gotAnswers, gotMs := map[int]int{}, map[int]int64{}
	for i, d := range sd.Days {
		gotAnswers[d-day], gotMs[d-day] = sd.Answers[i], sd.Ms[i]
	}
	var gotMature []int
	for _, d := range sd.Mature {
		gotMature = append(gotMature, d-day)
	}
	if !maps.Equal(gotAnswers, answers) || !maps.Equal(gotMs, ms) || !slices.Equal(gotMature, mature) {
		t.Fatalf("%s: answers %v, ms %v, mature %v\nwant %v, %v, %v", what, gotAnswers, gotMs, gotMature, answers, ms, mature)
	}
}

func TestProgressHistory(t *testing.T) {
	c, _, clk := testCollection(t)
	today := scheduler.DayNumber(clk.now(), DefaultRolloverHour)
	at := func(daysAgo int) time.Time { return clk.now().AddDate(0, 0, -daysAgo) }
	if _, err := c.StudyDays(DefaultRolloverHour, time.Local); err != nil { // loads the index before the changes
		t.Fatal(err)
	}
	// Import: a card mature 8 days ago (25 days), lapsed, and mature again
	// (counted once); a manual entry of 40 days on another card counts
	// neither as an answer nor as mature.
	importHistory(t, c, "lapsed", scheduler.CardState{State: scheduler.Review, Due: at(-10), ScheduledDays: 30},
		ImportReview{Time: at(10), Rating: scheduler.Good, StateBefore: scheduler.Review, IvlAfter: 10, DurationMs: 2000},
		ImportReview{Time: at(8), Rating: scheduler.Good, StateBefore: scheduler.Review, IvlBefore: 10, IvlAfter: 25, DurationMs: 3000},
		ImportReview{Time: at(6), Rating: scheduler.Again, StateBefore: scheduler.Review, IvlBefore: 25, IvlAfter: 1, DurationMs: 4000},
		ImportReview{Time: at(5), Rating: scheduler.Good, StateBefore: scheduler.Relearning, IvlBefore: 1, IvlAfter: 30, DurationMs: 1000},
		ImportReview{Time: at(5).Add(time.Hour), Rating: scheduler.Good, StateBefore: scheduler.Review, IvlBefore: 30, IvlAfter: 31, DurationMs: 500})
	due := importHistory(t, c, "due", scheduler.CardState{State: scheduler.Review, Due: at(0), ScheduledDays: 15, Reps: 3, LastReview: at(15), Memory: &scheduler.Memory{Stability: 15, Difficulty: 3}},
		ImportReview{Time: at(15), Rating: scheduler.Good, StateBefore: scheduler.Review, IvlAfter: 15, DurationMs: 6000},
		ImportReview{Time: at(3), StateBefore: scheduler.Review, IvlBefore: 15, IvlAfter: 40})
	answers := map[int]int{-15: 1, -10: 1, -8: 1, -6: 1, -5: 2}
	ms := map[int]int64{-15: 6000, -10: 2000, -8: 3000, -6: 4000, -5: 1500}
	checkHistory(t, c, "import", today, answers, ms, []int{-8})

	// Answer: the due card becomes mature today.
	e, err := c.Answer(due, scheduler.Good, 7*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if e.IvlAfter < MatureInterval {
		t.Fatalf("answer gave an interval of %d days", e.IvlAfter)
	}
	answers[0], ms[0] = 1, 7000
	checkHistory(t, c, "answer", today, answers, ms, []int{-8, 0})
	if recent, _ := c.RecentAnswers(0, at(0).Add(-time.Hour)); len(recent) != 1 || !recent[0].FirstMature || recent[0].DurationMs != 7000 {
		t.Fatalf("recent answers: %+v", recent)
	}
	// Undo takes it out again.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	delete(answers, 0)
	delete(ms, 0)
	checkHistory(t, c, "undo", today, answers, ms, []int{-8})
}

func TestProgressHistorySynced(t *testing.T) {
	var card ID
	a, b := linkedPair(t, func(c *Collection) {
		card = c.CardsOfNote(addNote(t, c, "Basic", "q", "a"))[0].ID
	})
	today := scheduler.DayNumber(b.clk.now(), DefaultRolloverHour)
	checkHistory(t, b.c, "before", today, map[int]int{}, map[int]int64{}, nil) // loads b's index
	advance(time.Minute, a, b)
	if _, err := a.c.Answer(card, scheduler.Good, 3*time.Second); err != nil {
		t.Fatal(err)
	}
	syncPair(t, a, b)
	checkHistory(t, b.c, "synced answer", today, map[int]int{0: 1}, map[int]int64{0: 3000}, nil)
	// Undone on a and synced again: the entry's deletion reaches b.
	if _, err := a.c.Undo(); err != nil {
		t.Fatal(err)
	}
	syncPair(t, a, b)
	checkHistory(t, b.c, "synced undo", today, map[int]int{}, map[int]int64{}, nil)
}
