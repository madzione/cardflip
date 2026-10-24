package core

import (
	"cmp"
	"math/rand/v2"
	"slices"

	"cardflip/golang/scheduler"
)

// MaxPosition is the highest position Reposition gives a card.
const MaxPosition = 1_000_000_000

// legacyPosFloor separates positions from the card IDs that Cardflip used
// as positions before Reposition existed: IDs are epoch milliseconds, far
// above any position.
const legacyPosFloor = 100_000_000_000

// newPos returns the position for a card added now: after every card.
// Callers hold the write lock.
func (c *Collection) newPos() int64 {
	p := c.nextPos
	c.nextPos++
	return p
}

// compactPositions renumbers the cards whose position is a legacy ID-based
// one, in their order, after the highest compact position. It reports
// whether it changed any card.
func compactPositions(cards []*Card) bool {
	var legacy []*Card
	var top int64
	for _, card := range cards {
		if card.NewPos >= legacyPosFloor {
			legacy = append(legacy, card)
		} else {
			top = max(top, card.NewPos)
		}
	}
	slices.SortFunc(legacy, positionOrder)
	for i, card := range legacy {
		card.NewPos = top + int64(i) + 1
	}
	return len(legacy) > 0
}

// migratePositions renumbers legacy positions after loading. The caller
// saves the result (compaction writes a snapshot).
func (c *Collection) migratePositions() bool {
	cards := make([]*Card, 0, len(c.cards))
	for _, card := range c.cards {
		cards = append(cards, card)
	}
	if !compactPositions(cards) {
		return false
	}
	c.nextPos = 1
	for _, card := range cards {
		c.nextPos = max(c.nextPos, card.NewPos+1)
	}
	return true
}

// placeAdded gives the New cards that o adds to decks with random insertion
// order a random position among the deck's new cards. Cards added to an
// empty deck are shuffled among themselves. Callers hold the write lock.
func (c *Collection) placeAdded(o *op) {
	added := map[ID][]*Card{}
	for _, card := range o.Cards {
		if _, ok := c.cards[card.ID]; ok || card.State != scheduler.New {
			continue
		}
		if c.optionsOf(card.DeckID).InsertionOrder == InsertRandom {
			added[card.DeckID] = append(added[card.DeckID], card)
		}
	}
	if len(added) == 0 {
		return
	}
	type span struct{ lo, hi int64 }
	existing := map[ID]span{}
	for _, card := range c.cards {
		if _, ok := added[card.DeckID]; !ok || card.State != scheduler.New {
			continue
		}
		s, ok := existing[card.DeckID]
		if !ok {
			s = span{card.NewPos, card.NewPos}
		}
		existing[card.DeckID] = span{min(s.lo, card.NewPos), max(s.hi, card.NewPos)}
	}
	for deck, cards := range added {
		s, ok := existing[deck]
		if ok {
			// Room for each added card at the end, too.
			s.hi += int64(len(cards))
		} else {
			s = span{cards[0].NewPos, cards[0].NewPos}
			for _, card := range cards {
				s = span{min(s.lo, card.NewPos), max(s.hi, card.NewPos)}
			}
		}
		for _, card := range cards {
			card.NewPos = s.lo + rand.Int64N(s.hi-s.lo+1)
		}
	}
}

// resortDeck returns copies of a deck's New cards with the deck's positions
// handed out again: in random order, or in the order the cards were added.
// Callers hold the write lock.
func (c *Collection) resortDeck(deck ID, random bool) []*Card {
	var cards []*Card
	var pos []int64
	for _, card := range c.cards {
		if card.HomeDeck() == deck && card.State == scheduler.New {
			cp := *card
			cards = append(cards, &cp)
			pos = append(pos, card.NewPos)
		}
	}
	slices.Sort(pos)
	slices.SortFunc(cards, func(a, b *Card) int { return cmp.Compare(a.ID, b.ID) })
	if random {
		rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	}
	var out []*Card
	for i, card := range cards {
		if card.NewPos != pos[i] {
			card.NewPos = pos[i]
			out = append(out, card)
		}
	}
	return out
}

// Reposition gives the New cards among ids the positions start, start+step,
// start+2*step, … in their current order, or in random order with
// randomize. With shift, the other New cards at or after start move back by
// step for each repositioned card, so no position is shared. Cards that are
// not New are left unchanged; skipped counts them.
func (c *Collection) Reposition(ids []ID, start, step int64, randomize, shift bool) (skipped int, err error) {
	if start < 0 || start > MaxPosition {
		return 0, userErr("startPositionRange", "The start position must be between 0 and %d.", MaxPosition)
	}
	if step < 1 || step > MaxPosition {
		return 0, userErr("positionStepRange", "The step must be between 1 and %d.", MaxPosition)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var cards []*Card
	for _, card := range c.cardCopies(ids) {
		if card.State == scheduler.New {
			cards = append(cards, card)
		} else {
			skipped++
		}
	}
	if len(cards) == 0 {
		return skipped, nil
	}
	if end := start + step*int64(len(cards)-1); end > MaxPosition {
		return 0, userErr("lastPositionTooHigh", "The last position would be %d; positions must be at most %d.", end, MaxPosition)
	}
	if randomize {
		rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	} else {
		slices.SortFunc(cards, positionOrder)
	}
	o := &op{Kind: "reposition"}
	chosen := map[ID]bool{}
	for i, card := range cards {
		chosen[card.ID] = true
		card.NewPos = start + step*int64(i)
		o.Cards = append(o.Cards, card)
	}
	if shift {
		by := step * int64(len(cards))
		for _, card := range c.cards {
			if !chosen[card.ID] && card.State == scheduler.New && card.NewPos >= start {
				cp := *card
				cp.NewPos += by
				o.Cards = append(o.Cards, &cp)
			}
		}
	}
	return skipped, c.commit(o)
}
