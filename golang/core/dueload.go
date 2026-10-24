package core

import (
	"time"

	"cardflip/golang/scheduler"
)

// dueLoad counts the cards due as reviews on each study day, for Easy Days
// (scheduler.Params.DueLoad). apply keeps it up to date with every change
// of a card, and rebuilds it when the rollover hour changes.
type dueLoad struct {
	rollover int
	loc      *time.Location
	days     map[int]int
}

// resetDueLoad empties the counts for the current rollover hour.
func (c *Collection) resetDueLoad() {
	c.load = dueLoad{rollover: c.settings.RolloverHour, loc: c.now().Location(), days: map[int]int{}}
}

// rebuildDueLoad counts every card again.
func (c *Collection) rebuildDueLoad() {
	c.resetDueLoad()
	for _, card := range c.cards {
		c.load.add(card, 1)
	}
}

// add changes the count of card's due day by n, if the card is due as a
// review: a review or interday learning card that is not suspended.
func (l *dueLoad) add(card *Card, n int) {
	if card.Suspended || card.Due.IsZero() || (card.State != scheduler.Review && !card.Interday) {
		return
	}
	day := scheduler.DayNumber(card.Due.In(l.loc), l.rollover)
	if l.days[day] += n; l.days[day] == 0 {
		delete(l.days, day)
	}
}

// count returns how many cards are due on a study day.
func (l *dueLoad) count(day int) int { return l.days[day] }
