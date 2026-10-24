// Package scheduler decides when cards are due: FSRS memory states (via
// go-fsrs), learning and relearning steps, interval fuzz, previews, the
// study day boundary and conversion of Anki scheduling data.
package scheduler

import "time"

// State is a card's scheduling state.
type State int

const (
	New State = iota
	Learning
	Review
	Relearning
)

func (s State) String() string {
	switch s {
	case New:
		return "new"
	case Learning:
		return "learning"
	case Review:
		return "review"
	case Relearning:
		return "relearning"
	}
	return "unknown"
}

// Memory is the FSRS memory state of a card.
type Memory struct {
	Stability  float64 `json:"s"`
	Difficulty float64 `json:"d"`
}

// CardState is everything the scheduler knows about a card.
type CardState struct {
	State State `json:"state"`
	// Due is when the card is next shown; zero for New cards.
	Due time.Time `json:"due,omitzero"`
	// Memory is absent while the card is New.
	Memory *Memory `json:"mem,omitempty"`
	// Step is the index of the current learning or relearning step.
	Step int `json:"step,omitempty"`
	// ScheduledDays is the interval of the last scheduling, in days (0 for
	// intraday steps).
	ScheduledDays int `json:"ivl,omitempty"`
	// Interday marks a (re)learning card whose step ends on a later study
	// day: it is due from the start of the day Due falls in, like a review.
	Interday   bool      `json:"interday,omitempty"`
	Reps       int       `json:"reps,omitempty"`
	Lapses     int       `json:"lapses,omitempty"`
	LastReview time.Time `json:"last,omitzero"`
}

// Rating is the user's answer to a card.
type Rating int

const (
	Again Rating = 1
	Hard  Rating = 2
	Good  Rating = 3
	Easy  Rating = 4
)

// Valid reports whether r is one of the four ratings.
func (r Rating) Valid() bool { return r >= Again && r <= Easy }

func (r Rating) String() string {
	switch r {
	case Again:
		return "again"
	case Hard:
		return "hard"
	case Good:
		return "good"
	case Easy:
		return "easy"
	}
	return "invalid"
}
