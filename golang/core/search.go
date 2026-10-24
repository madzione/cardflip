package core

import (
	"math"
	"strings"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/search"
)

// Search returns the IDs of the cards matching query, sorted by a column.
// An invalid query returns a *search.Error naming the bad term.
func (c *Collection) Search(query string, order Column, descending bool) ([]ID, error) {
	now := c.now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids, err := c.match(query, now)
	if err != nil {
		return nil, err
	}
	c.sortCards(ids, order, descending, now)
	return ids, nil
}

// CheckSearch returns the *search.Error searching for query would fail
// with, or nil when query is valid.
func (c *Collection) CheckSearch(query string) error {
	now := c.now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, err := c.parseQuery(query, now)
	return err
}

// match returns the IDs of the cards matching query, in no particular
// order. Callers hold the read lock, which match may release and take
// again to build the review index.
func (c *Collection) match(query string, now time.Time) ([]ID, error) {
	q, err := c.parseQuery(query, now)
	for err == nil && q.UsesReviews() && c.reviews == nil {
		// Building the index needs the write lock; a restore in between can
		// drop it again.
		c.mu.RUnlock()
		err = c.loadReviewIndex()
		c.mu.RLock()
		if err != nil {
			return nil, err
		}
		q, err = c.parseQuery(query, now)
	}
	if err != nil {
		return nil, err
	}
	return c.matchQuery(q, now), nil
}

// matchLocked is match for callers holding the write lock.
func (c *Collection) matchLocked(query string, now time.Time) ([]ID, error) {
	q, err := c.parseQuery(query, now)
	if err != nil {
		return nil, err
	}
	if q.UsesReviews() {
		if err := c.loadReviewIndexLocked(); err != nil {
			return nil, err
		}
	}
	return c.matchQuery(q, now), nil
}

// matchQuery returns the IDs of the cards matching q, in no particular
// order. When q uses reviews, the review index is loaded. Callers hold a
// lock.
func (c *Collection) matchQuery(q search.Query, now time.Time) []ID {
	reviews := q.UsesReviews()
	dayEnd := scheduler.NextDayStart(now, c.settings.RolloverHour)
	today := scheduler.DayNumber(now, c.settings.RolloverHour)
	var studied map[ID]map[ID]time.Time
	if q.UsesSiblingBurial() {
		studied = c.studied(c.todayEntries(now))
	}
	fieldNames := make(map[ID][]string, len(c.noteTypes))
	excluded := map[ID][]bool{}
	for id, nt := range c.noteTypes {
		names := make([]string, len(nt.Fields))
		for i, f := range nt.Fields {
			names[i] = strings.ToLower(f)
			if nt.FieldOption(i).ExcludeFromSearch {
				if excluded[id] == nil {
					excluded[id] = make([]bool, len(nt.Fields))
				}
				excluded[id][i] = true
			}
		}
		fieldNames[id] = names
	}
	var sc search.Card
	var ids []ID
	for _, card := range c.cards {
		n := c.notes[card.NoteID]
		sc.Fields = c.noteFields[n.ID]
		sc.Raw = n.Fields
		sc.FieldNames = fieldNames[n.TypeID]
		sc.Excluded = excluded[n.TypeID]
		d := c.decks[card.DeckID]
		sc.Deck = d.Name
		home := c.decks[card.HomeDeck()]
		sc.HomeDeck, sc.Filtered = home.Name, card.HomeDeckID != 0
		sc.Preset = c.presetOf(home).Name
		sc.Tags = n.Tags
		nt := c.noteTypes[n.TypeID]
		sc.NoteType = nt.Name
		sc.State = search.State(card.State)
		sc.Due = card.State != scheduler.New && card.Due.Before(dayEnd)
		sc.Suspended = card.Suspended
		sc.Flag = card.Flag
		sc.BuriedManually = isBuried(card, now)
		// Sibling burying holds back the cards today's queue would have
		// (see view).
		sc.BuriedSibling = studied != nil && !card.Suspended && !sc.BuriedManually &&
			(card.State == scheduler.New || (card.State == scheduler.Review || card.Interday) && card.Due.Before(dayEnd)) &&
			c.siblingBuried(card, studied)
		fillProperties(&sc, card, now, today, c.settings.RolloverHour)
		sc.CardOrd = card.Ord + 1
		// Cloze cards share the note type's one template.
		tmpl := card.Ord
		if nt.Kind == Cloze {
			tmpl = 0
		}
		sc.CardName = ""
		if tmpl < len(nt.Templates) {
			sc.CardName = nt.Templates[tmpl].Name
		}
		sc.NoteID, sc.CardID = int64(n.ID), int64(card.ID)
		// IDs are creation times (see ID).
		sc.Added = time.UnixMilli(int64(card.ID))
		sc.Edited = n.Modified
		if reviews {
			sc.Reviews = c.reviews.of(card.ID)
		}
		if q.Match(&sc) {
			ids = append(ids, card.ID)
		}
	}
	return ids
}

// fillProperties sets the numeric properties prop: compares, as the
// statistics count them: the interval of Review cards, the memory state of
// cards other than New. today is the current study day's DayNumber.
func fillProperties(sc *search.Card, card *Card, now time.Time, today, rolloverHour int) {
	sc.Reps, sc.Lapses = card.Reps, card.Lapses
	sc.Interval, sc.Position = reviewInterval(card), math.NaN()
	sc.Stability, sc.Difficulty, sc.Retrievability = stability(card), difficulty(card), retrievability(card, now)
	sc.DueDay = math.NaN()
	if card.State == scheduler.New {
		sc.Position = float64(card.NewPos)
		return
	}
	sc.DueDay = float64(scheduler.DayNumber(card.Due.In(now.Location()), rolloverHour) - today)
}

// reviewInterval is a Review card's interval in days; NaN for other cards.
func reviewInterval(card *Card) float64 {
	if card.State != scheduler.Review {
		return math.NaN()
	}
	return float64(max(card.ScheduledDays, 1))
}

// stability is a card's stability in days; NaN without a memory state.
func stability(card *Card) float64 {
	if card.State == scheduler.New || card.Memory == nil {
		return math.NaN()
	}
	return card.Memory.Stability
}

// difficulty is a card's difficulty from 0 to 1; NaN without a memory
// state.
func difficulty(card *Card) float64 {
	if card.State == scheduler.New || card.Memory == nil {
		return math.NaN()
	}
	return min(max((card.Memory.Difficulty-1)/9, 0), 1)
}

// retrievability is the probability of recalling a card now; NaN when it
// is unknown.
func retrievability(card *Card, now time.Time) float64 {
	if card.State == scheduler.New {
		return math.NaN()
	}
	if r, ok := scheduler.Retrievability(card.CardState, now); ok {
		return r
	}
	return math.NaN()
}

// parseQuery parses a browser query with the collection's settings.
// Callers hold a lock.
func (c *Collection) parseQuery(query string, now time.Time) (search.Query, error) {
	o := search.Options{IgnoreAccents: c.settings.IgnoreAccents, Now: now, RolloverHour: c.settings.RolloverHour}
	if d, ok := c.decks[c.settings.CurrentDeckID]; ok {
		o.CurrentDeck = d.Name
	}
	return search.ParseWith(query, o)
}
