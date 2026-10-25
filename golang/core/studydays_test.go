package core

import (
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
	if sorted := c.studyDaysSorted(StudyDays{}, DefaultRolloverHour, berlin); !slices.Equal(sorted.Days, want) {
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
