package scheduler

import (
	"math"
	"testing"
	"time"
)

var ankiCrt = time.Date(2026, 1, 1, 4, 0, 0, 0, time.UTC)

func ankiDay(d int, minutes int) time.Time {
	return ankiCrt.Add(time.Duration(d)*24*time.Hour + 2*time.Hour + time.Duration(minutes)*time.Minute)
}

// smTwoHistory is a card reviewed 8 times under Anki's SM-2 scheduler:
// three learning steps, then five reviews.
func smTwoHistory() []AnkiReview {
	return []AnkiReview{
		{ankiDay(100, 0), 3, -60, 0, 0},
		{ankiDay(100, 1), 3, -600, 0, 0},
		{ankiDay(100, 11), 3, 1, 2500, 0},
		{ankiDay(101, 0), 3, 3, 2500, 1},
		{ankiDay(104, 0), 3, 8, 2500, 1},
		{ankiDay(112, 0), 2, 10, 2350, 1},
		{ankiDay(122, 0), 3, 30, 2350, 1},
		{ankiDay(210, 0), 2, 20, 2300, 1},
	}
}

func smTwoCard(history []AnkiReview) AnkiCard {
	return AnkiCard{Type: 2, Queue: 2, Due: 230, Ivl: 20, Factor: 2300, Reps: 8, History: history}
}

// replay rates a card through ratings in the given states, by hand.
func replay(steps []struct {
	state State
	r     Rating
	at    time.Time
}) Memory {
	var cs CardState
	for _, s := range steps {
		cs.State = s.state
		m := memoryAfter(cs, s.r, s.at, nil)
		cs.Memory, cs.LastReview = &m, s.at
	}
	return *cs.Memory
}

// Card scheduled with SM-2 (data-import spec): its stability and difficulty
// result from replaying its 8 reviews, and its due date is kept.
func TestMemoryFromSMTwoHistory(t *testing.T) {
	h := smTwoHistory()
	cs := FromAnki(smTwoCard(h), ankiCrt)
	type step = struct {
		state State
		r     Rating
		at    time.Time
	}
	want := replay([]step{
		{New, Good, h[0].Time}, {Learning, Good, h[1].Time}, {Learning, Good, h[2].Time},
		{Review, Good, h[3].Time}, {Review, Good, h[4].Time}, {Review, Hard, h[5].Time},
		{Review, Good, h[6].Time}, {Review, Hard, h[7].Time},
	})
	if cs.Memory == nil || *cs.Memory != want {
		t.Fatalf("memory = %+v, want %+v", cs.Memory, want)
	}
	if est := estimateMemory(smTwoCard(nil)); *est == want {
		t.Fatalf("memory equals the estimate %+v", est)
	}
	if !cs.Due.Equal(ankiCrt.AddDate(0, 0, 230)) || cs.State != Review || cs.ScheduledDays != 20 {
		t.Fatalf("scheduling changed: %+v", cs)
	}
}

// Manual and rescheduled entries and cram reviews do not count; a review in
// a filtered deck that reschedules does.
func TestHistorySkipsManualAndCram(t *testing.T) {
	base := *FromAnki(smTwoCard(smTwoHistory()), ankiCrt).Memory
	h := smTwoHistory()
	with := append(append(append([]AnkiReview{}, h[:5]...),
		AnkiReview{ankiDay(105, 0), 0, 40, 2500, 5},  // set due date
		AnkiReview{ankiDay(106, 0), 0, 20, 2500, 4},  // manual reschedule
		AnkiReview{ankiDay(107, 0), 3, 0, 0, 3},      // cram, no rescheduling
		AnkiReview{ankiDay(109, 0), 7, 10, 2500, 1}), // invalid rating
		h[5:]...)
	if got := *FromAnki(smTwoCard(with), ankiCrt).Memory; got != base {
		t.Fatalf("memory = %+v, want %+v", got, base)
	}
	filtered := append(append(append([]AnkiReview{}, h[:5]...), AnkiReview{ankiDay(107, 0), 3, 9, 2500, 3}), h[5:]...)
	if got := *FromAnki(smTwoCard(filtered), ankiCrt).Memory; got == base {
		t.Fatal("a rescheduling filtered review was skipped")
	}
}

// A reset to new starts the replay over.
func TestHistoryResetStartsOver(t *testing.T) {
	h := smTwoHistory()
	after := []AnkiReview{
		{ankiDay(150, 0), 1, -60, 0, 0},
		{ankiDay(150, 10), 3, 1, 2500, 0},
	}
	reset := append(append(append([]AnkiReview{}, h[:7]...), AnkiReview{ankiDay(140, 0), 0, 0, 0, 4}), after...)
	got := FromAnki(smTwoCard(reset), ankiCrt).Memory
	want := FromAnki(smTwoCard(after), ankiCrt).Memory
	if got == nil || *got != *want {
		t.Fatalf("memory = %+v, want %+v", got, want)
	}
}

// Anki's memory state wins over the history; without either, the memory
// state is estimated from interval and ease. A learning card gets a memory
// state from its history too.
func TestAnkiMemorySources(t *testing.T) {
	a := smTwoCard(smTwoHistory())
	a.Memory = &Memory{Stability: 14.25, Difficulty: 5.5}
	if got := FromAnki(a, ankiCrt).Memory; *got != *a.Memory {
		t.Fatalf("memory = %+v", got)
	}
	got := FromAnki(smTwoCard(nil), ankiCrt).Memory
	if got.Stability != 20 || math.Abs(got.Difficulty-(10-9*1000.0/2200)) > 1e-9 {
		t.Fatalf("estimate = %+v", got)
	}
	// Only manual entries: estimated.
	if got := FromAnki(smTwoCard([]AnkiReview{{ankiDay(5, 0), 0, 20, 2300, 5}}), ankiCrt).Memory; got.Stability != 20 {
		t.Fatalf("estimate with manual history = %+v", got)
	}
	learning := AnkiCard{Type: 1, Queue: 1, Due: 1790000000, Left: 1001, Reps: 1,
		History: []AnkiReview{{ankiDay(263, 0), 1, -60, 0, 0}}}
	cs := FromAnki(learning, ankiCrt)
	if cs.State != Learning || cs.Memory == nil || cs.Due.Unix() != 1790000000 || cs.Interday {
		t.Fatalf("learning card = %+v", cs)
	}
	// Day learning (queue 3) is due on a day: interday learning.
	if cs := FromAnki(AnkiCard{Type: 3, Queue: 3, Due: 264, Left: 1001, Ivl: 10}, ankiCrt); !cs.Interday ||
		cs.State != Relearning || !cs.Due.Equal(ankiCrt.AddDate(0, 0, 264)) {
		t.Fatalf("day learning card = %+v", cs)
	}
	if FromAnki(AnkiCard{Type: 1, Queue: 1, Due: 1790000000, Left: 1001}, ankiCrt).Memory != nil {
		t.Fatal("learning card without history got a memory state")
	}
}
