package core

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"cardflip/golang/scheduler"
)

// DueDateSpec is a parsed set-due-date input: a delay in days from today,
// or a random delay from From to To for each card.
type DueDateSpec struct {
	From, To int
	// SetInterval also sets the cards' intervals to the delay (a trailing
	// "!").
	SetInterval bool
}

const dueDateHelp = "Enter days from today: 0 for today, 1 for tomorrow, 60-90 for a random day in that range, with ! to set the interval too."

// ParseDueDate parses set-due-date input: "N" days from today (0 is
// today), "N-M" for a random day in that range per card, either followed by
// "!" to set the interval too. Spaces are ignored.
func ParseDueDate(input string) (DueDateSpec, error) {
	s := strings.Join(strings.Fields(input), "")
	var spec DueDateSpec
	if rest, ok := strings.CutSuffix(s, "!"); ok {
		s, spec.SetInterval = rest, true
	}
	if s == "" {
		return DueDateSpec{}, userErr("dueDateHelp", dueDateHelp)
	}
	from, to, isRange := strings.Cut(s, "-")
	var err error
	if spec.From, err = dueDays(from); err != nil {
		return DueDateSpec{}, err
	}
	spec.To = spec.From
	if isRange {
		if spec.To, err = dueDays(to); err != nil {
			return DueDateSpec{}, err
		}
		if spec.To < spec.From {
			return DueDateSpec{}, userErr("dueRangeReversed", "The range %d-%d ends before it starts; put the smaller number first.", spec.From, spec.To)
		}
	}
	return spec, nil
}

func dueDays(s string) (int, error) {
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, userErr("dueDateHelp", dueDateHelp)
	}
	// Only digits, so Atoi fails only when the number is too large.
	n, err := strconv.Atoi(s)
	if err != nil || n > MaxIntervalDays {
		return 0, userErr("dueDateTooFar", "A due date can be at most %d days away.", MaxIntervalDays)
	}
	return n, nil
}

// days is the delay for a card. A range picks a day per card, seeded by the
// card and the study day, so the result is reproducible.
func (s DueDateSpec) days(card ID, today int) int {
	if s.From == s.To {
		return s.From
	}
	r := rand.New(rand.NewPCG(uint64(card), uint64(today)))
	return s.From + r.IntN(s.To-s.From+1)
}

// SetDueDate makes cards Review cards due the number of days from today
// that input gives (see ParseDueDate). A review card keeps its interval
// unless input ends in "!"; other cards get the delay as interval. Memory
// states are kept, and a card without one gets the memory state of a first
// Good rating. Each card gets a manual review log entry. The input is
// remembered in Settings.LastDueDate. Invalid input changes nothing.
func (c *Collection) SetDueDate(ids []ID, input string) error {
	spec, err := ParseDueDate(input)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	today := scheduler.DayNumber(now, c.settings.RolloverHour)
	o := &op{Kind: "setDueDate"}
	for _, card := range c.cardCopies(ids) {
		days := spec.days(card.ID, today)
		before := card.CardState
		next := before
		if next.Memory == nil {
			m := scheduler.FirstGoodMemory(now, c.deckOptions(card).Weights)
			next.Memory = &m
			if next.LastReview.IsZero() {
				next.LastReview = now
			}
		}
		if before.State != scheduler.Review || spec.SetInterval {
			next.ScheduledDays = max(days, 1)
		}
		next.State, next.Step, next.Interday = scheduler.Review, 0, false
		next.Due = now.AddDate(0, 0, days)
		card.CardState = next
		o.Cards = append(o.Cards, card)
		o.AddReviews = append(o.AddReviews, c.manualEntry(card.ID, before, next, now))
	}
	if len(o.Cards) == 0 {
		return nil
	}
	if last := strings.TrimSpace(input); c.settings.LastDueDate != last {
		s := c.settings
		s.LastDueDate = last
		o.Settings = &s
	}
	rev, err := reviewWrites(o.AddReviews, nil)
	if err != nil {
		return err
	}
	return c.commit(o, rev...)
}

// manualEntry is the review log entry (rating 0) for a change of a card's
// scheduling by hand from before to after. Callers hold the write lock.
func (c *Collection) manualEntry(card ID, before, after scheduler.CardState, now time.Time) *ReviewEntry {
	return &ReviewEntry{ID: c.ids.next(), CardID: card, Time: now, StateBefore: before.State,
		IvlBefore: before.ScheduledDays, IvlAfter: after.ScheduledDays, Before: before}
}
