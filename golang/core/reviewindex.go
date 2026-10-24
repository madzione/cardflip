package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/search"
)

// reviewIndex holds every card's review log entries, the whole stored log
// and not only the loaded months, for the searches that look at history
// (rated:, introduced:, resched:, prop:rated, prop:resched). Times are kept
// rather than study days, so a change of the rollover hour needs no
// rebuild.
type reviewIndex map[ID]*cardReviews

// cardReviews are one card's entries, oldest first (by time, then ID).
// ids[i] is the entry ID of entries[i].
type cardReviews struct {
	ids     []ID
	entries []search.LogEntry
}

// indexedEntry is the part of a stored ReviewEntry the index needs.
type indexedEntry struct {
	ID     ID               `json:"id"`
	CardID ID               `json:"card"`
	Time   time.Time        `json:"time"`
	Rating scheduler.Rating `json:"rating"`
	// Preview answers are left out: they changed no scheduling.
	Preview bool `json:"preview"`
}

// loadReviewIndex reads the whole stored review log into c.reviews, unless
// it is already there. It takes the write lock, so callers hold no lock.
func (c *Collection) loadReviewIndex() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loadReviewIndexLocked()
}

// loadReviewIndexLocked is loadReviewIndex for callers holding the write
// lock.
func (c *Collection) loadReviewIndexLocked() error {
	if c.reviews != nil {
		return nil
	}
	months, err := c.store.RevlogMonths()
	if err != nil {
		return err
	}
	idx := reviewIndex{}
	for _, m := range months {
		c.yield()
		entries, err := c.store.ReadRevlog(m)
		if err != nil {
			return err
		}
		for i, raw := range entries {
			var e indexedEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, i, err)
			}
			if !e.Preview {
				idx.add(e.ID, e.CardID, e.Time, e.Rating)
			}
		}
	}
	c.reviews = idx
	return nil
}

// add inserts an entry in its place in the card's list.
func (idx reviewIndex) add(id, cardID ID, t time.Time, rating scheduler.Rating) {
	cr := idx[cardID]
	if cr == nil {
		cr = &cardReviews{}
		idx[cardID] = cr
	}
	e := search.LogEntry{Time: t.UnixMilli(), Rating: int(rating)}
	// Entries mostly arrive in order, so the place is found from the end.
	i := len(cr.entries)
	for i > 0 && (cr.entries[i-1].Time > e.Time || cr.entries[i-1].Time == e.Time && cr.ids[i-1] > id) {
		i--
	}
	cr.ids = slices.Insert(cr.ids, i, id)
	cr.entries = slices.Insert(cr.entries, i, e)
}

// remove deletes entry id of a card; cardID 0 means the card is not known
// and every card is looked at.
func (idx reviewIndex) remove(id, cardID ID) {
	if cardID == 0 {
		for cid, cr := range idx {
			if slices.Contains(cr.ids, id) {
				cardID = cid
				break
			}
		}
	}
	cr := idx[cardID]
	if cr == nil {
		return
	}
	if i := slices.Index(cr.ids, id); i >= 0 {
		cr.ids = slices.Delete(cr.ids, i, i+1)
		cr.entries = slices.Delete(cr.entries, i, i+1)
	}
	if len(cr.ids) == 0 {
		delete(idx, cardID)
	}
}

// of returns a card's entries for matching, oldest first.
func (idx reviewIndex) of(cardID ID) []search.LogEntry {
	if cr := idx[cardID]; cr != nil {
		return cr.entries
	}
	return nil
}

// applyReviews keeps c.reviews up to date with the review log changes of o.
// apply calls it before it changes c.revlog, which still holds the entry an
// undo removes.
func (c *Collection) applyReviews(o *op) {
	if c.reviews == nil {
		return
	}
	if e := o.Review; e != nil && !e.Preview {
		c.reviews.add(e.ID, e.CardID, e.Time, e.Rating)
	}
	if o.ReviewUndone != 0 {
		var cardID ID
		if e := c.reviewEntry(o.ReviewUndone); e != nil {
			cardID = e.CardID
		}
		c.reviews.remove(o.ReviewUndone, cardID)
	}
	for _, e := range o.DelReviews {
		c.reviews.remove(e.ID, e.CardID)
	}
	for _, e := range o.AddReviews {
		if e.Preview {
			continue
		}
		c.reviews.add(e.ID, e.CardID, e.Time, e.Rating)
	}
}
