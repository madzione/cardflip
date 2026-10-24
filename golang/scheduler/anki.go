package scheduler

import (
	"math"
	"time"
)

// AnkiCard is the scheduling data of a card in an Anki collection.
type AnkiCard struct {
	// Type: 0 new, 1 learning, 2 review, 3 relearning.
	Type int
	// Queue: -1 suspended, -2/-3 buried, 0 new, 1 learning (Due is epoch
	// seconds), 2 review and 3 day learning (Due is a day number), 4 preview.
	Queue                          int
	Due, Ivl, Factor, Reps, Lapses int64
	// Left encodes the remaining learning steps (Left % 1000).
	Left int64
	// Memory is the FSRS state from the card's data, when Anki had FSRS on.
	Memory *Memory
	// History is the card's review log, oldest first.
	History []AnkiReview
}

// AnkiReview is one entry of an Anki card's review log.
type AnkiReview struct {
	Time time.Time
	// Ease is the rating, 1 to 4, or 0 for a manual entry.
	Ease int
	// Ivl is the interval after the entry: days when positive, seconds when
	// negative.
	Ivl int64
	// Factor is the ease factor; 0 for a cram review in a filtered deck.
	Factor int64
	// Type: 0 learning, 1 review, 2 relearning, 3 filtered, 4 manual,
	// 5 rescheduled.
	Type int
}

// FromAnki converts Anki scheduling data. crt is the collection's creation
// time, from which review due days are counted. The memory state comes
// from Anki when present, is otherwise computed from the review history,
// and without history estimated: stability from the interval, difficulty
// mapped linearly from ease 130%–350% to 10–1.
func FromAnki(a AnkiCard, crt time.Time) CardState {
	if a.Type == 0 {
		return CardState{State: New}
	}
	cs := CardState{Reps: int(a.Reps), Lapses: int(a.Lapses), ScheduledDays: int(max(a.Ivl, 0))}
	dayDue := func() time.Time { return crt.Add(time.Duration(a.Due) * 24 * time.Hour) }
	mem := a.Memory
	if mem == nil {
		mem = memoryFromHistory(a.History)
	}
	switch a.Type {
	case 2:
		cs.State = Review
		cs.Due = dayDue()
		if cs.ScheduledDays < 1 {
			cs.ScheduledDays = 1
		}
		if mem == nil {
			mem = estimateMemory(a)
		}
	default: // 1 learning, 3 relearning
		cs.State = Learning
		steps := 2
		if a.Type == 3 {
			cs.State = Relearning
			steps = 1
			if mem == nil {
				mem = estimateMemory(a)
			}
		}
		if a.Queue == 1 || a.Due > 1_000_000_000 {
			cs.Due = time.Unix(a.Due, 0)
		} else {
			cs.Due = dayDue()
			cs.Interday = true
		}
		remaining := int(a.Left % 1000)
		cs.Step = min(max(steps-remaining, 0), steps-1)
		cs.ScheduledDays = 0
	}
	if mem != nil {
		m := *mem
		cs.Memory = &m
	}
	if cs.State == Review {
		cs.LastReview = cs.Due.AddDate(0, 0, -cs.ScheduledDays)
	}
	return cs
}

// memoryFromHistory replays FSRS over an Anki review history, as Anki does
// to compute memory states: manual and rescheduled entries and cram reviews
// (filtered, factor 0) are skipped, and a reset to new (a manual entry with
// interval 0) starts over. Each review is rated in the state its entry
// records. It returns nil when no review is left.
func memoryFromHistory(rs []AnkiReview) *Memory {
	var cs CardState
	for _, r := range rs {
		switch {
		case r.Ease == 0 || r.Type == 4 || r.Type == 5:
			if r.Ivl == 0 {
				cs = CardState{}
			}
			continue
		case r.Type == 3 && r.Factor == 0, r.Ease < 1 || r.Ease > 4:
			continue
		}
		switch r.Type {
		case 0:
			cs.State = Learning
		case 2:
			cs.State = Relearning
		default: // 1 review, 3 filtered
			cs.State = Review
		}
		if cs.Memory == nil {
			cs.State = New
		}
		mem := memoryAfter(cs, Rating(r.Ease), r.Time, nil)
		cs.Memory = &mem
		cs.LastReview = r.Time
	}
	return cs.Memory
}

// RecoverMemory returns a memory state for a card that lost its own: computed
// from its review history, oldest first, as for imported cards, or without
// a usable history estimated from its interval in days. fromHistory reports
// which.
func RecoverMemory(history []AnkiReview, ivlDays int) (m Memory, fromHistory bool) {
	if mem := memoryFromHistory(history); mem != nil {
		return *mem, true
	}
	return *estimateMemory(AnkiCard{Ivl: int64(ivlDays)}), false
}

func estimateMemory(a AnkiCard) *Memory {
	s := math.Max(float64(a.Ivl), 0.5)
	d := 5.0
	if a.Factor > 0 {
		d = 10 - 9*(float64(a.Factor)-1300)/2200
		d = math.Min(math.Max(d, 1), 10)
	}
	return &Memory{Stability: s, Difficulty: d}
}
