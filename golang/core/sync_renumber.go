package core

import (
	"cmp"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"slices"
	"strconv"
	"time"
)

// Two collections created separately can use the same ID for different
// objects, as IDs are times. Before their first merge (sync design decision
// 4), one of them gives its colliding notes, cards and review log entries
// new IDs. An object counts as the same in both when its identity matches:
// a note's GUID, a card's note GUID and card type, a review log entry's
// card (by its identity), time and rating. Decks, presets and note types with the same ID are the
// same object, as in collections copied by backup.

// Identities are the identity hashes of a collection's notes, cards and
// review log entries, by ID.
type Identities struct {
	Notes   map[ID]uint64 `json:"notes"`
	Cards   map[ID]uint64 `json:"cards"`
	Reviews map[ID]uint64 `json:"reviews"`
}

func identity(parts ...string) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return h.Sum64()
}

func noteIdentity(n *Note) uint64 { return identity("note", n.GUID) }

// cardIdentity is a card's note's GUID and its card type; a card whose
// note is missing has its note ID instead.
func (c *Collection) cardIdentity(card *Card) uint64 {
	ord := strconv.Itoa(card.Ord)
	if n, ok := c.notes[card.NoteID]; ok {
		return identity("card", n.GUID, ord)
	}
	return identity("card#", strconv.FormatInt(int64(card.NoteID), 10), ord)
}

// reviewIdentity is a review log entry's card's identity (its ID if the
// card is gone), time and rating. Callers hold a lock.
func (c *Collection) reviewIdentity(e *ReviewEntry) uint64 {
	card := strconv.FormatInt(int64(e.CardID), 10)
	if cd, ok := c.cards[e.CardID]; ok {
		card = strconv.FormatUint(c.cardIdentity(cd), 10)
	}
	return identity("review", card, strconv.FormatInt(e.Time.UnixMilli(), 10), strconv.Itoa(int(e.Rating)))
}

// Identities returns the identities of the collection's objects, for the
// peer of a first merge.
func (c *Collection) Identities() (Identities, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids := Identities{Notes: make(map[ID]uint64, len(c.notes)), Cards: make(map[ID]uint64, len(c.cards)),
		Reviews: map[ID]uint64{}}
	for id, n := range c.notes {
		ids.Notes[id] = noteIdentity(n)
	}
	for id, card := range c.cards {
		ids.Cards[id] = c.cardIdentity(card)
	}
	err := c.eachReview(func(e *ReviewEntry, _ string) error {
		ids.Reviews[e.ID] = c.reviewIdentity(e)
		return nil
	})
	return ids, err
}

// eachReview calls fn for every stored review log entry. Callers hold a
// lock.
func (c *Collection) eachReview(fn func(e *ReviewEntry, month string) error) error {
	months, err := c.store.RevlogMonths()
	if err != nil {
		return err
	}
	for _, m := range months {
		recs, err := c.store.ReadRevlog(m)
		if err != nil {
			return err
		}
		for i, raw := range recs {
			var e ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, i, err)
			}
			if err := fn(&e, m); err != nil {
				return err
			}
		}
		c.yield()
	}
	return nil
}

// opRenumber is the kind of the op that gives colliding objects new IDs.
// It removes the objects under their old IDs without tombstones, as those
// IDs belong to the peer's objects, and without deletion log entries.
const opRenumber = "renumber"

// Renumber gives the local notes, cards and review log entries whose IDs
// the peer uses for different objects new IDs, with every reference to
// them, in one change that is not an undo step. It returns how many objects
// it renumbered. Call it before the first merge's exchange, once the
// collection has its device slot, so the new IDs cannot collide again.
func (c *Collection) Renumber(peer Identities) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	o := &op{Kind: opRenumber}
	notes := map[ID]ID{}
	for _, id := range sortedIDs(c.notes) {
		if h, ok := peer.Notes[id]; ok && h != noteIdentity(c.notes[id]) {
			notes[id] = c.ids.next()
		}
	}
	cards := map[ID]ID{}
	for _, id := range sortedIDs(c.cards) {
		if h, ok := peer.Cards[id]; ok && h != c.cardIdentity(c.cards[id]) {
			cards[id] = c.ids.next()
		}
	}
	n := len(notes) + len(cards)
	for _, id := range sortedIDs(c.notes) {
		to, ok := notes[id]
		if !ok {
			continue
		}
		nn := cloneNote(c.notes[id])
		nn.ID = to
		o.Notes = append(o.Notes, &nn)
		o.DelNotes = append(o.DelNotes, id)
	}
	for _, id := range sortedIDs(c.cards) {
		card := c.cards[id]
		newID, renamed := cards[id]
		newNote, moved := notes[card.NoteID]
		if !renamed && !moved {
			continue
		}
		cp := *card
		if renamed {
			cp.ID = newID
			o.DelCards = append(o.DelCards, id)
		}
		if moved {
			cp.NoteID = newNote
		}
		o.Cards = append(o.Cards, &cp)
	}
	// Review log entries: those whose ID collides get a new one, and those
	// of renumbered cards their card's new ID.
	err := c.eachReview(func(e *ReviewEntry, _ string) error {
		h, collides := peer.Reviews[e.ID]
		collides = collides && h != c.reviewIdentity(e)
		newCard, moved := cards[e.CardID]
		if !collides && !moved {
			return nil
		}
		ne := *e
		if moved {
			ne.CardID = newCard
		}
		if collides {
			ne.ID = c.ids.next()
			n++
		}
		o.DelReviews = append(o.DelReviews, e)
		o.AddReviews = append(o.AddReviews, &ne)
		return nil
	})
	if err != nil {
		return 0, err
	}
	if n == 0 && len(o.AddReviews) == 0 {
		return 0, nil
	}
	rev, err := reviewWrites(o.AddReviews, o.DelReviews)
	if err != nil {
		return 0, err
	}
	if err := c.commit(o, rev...); err != nil {
		return 0, err
	}
	// The history's before-images use the old IDs.
	c.hist.clear()
	return n, nil
}

// TombstoneRetention is how long tombstones are kept at most (decision
// 10): a link that has not synced for longer asks the first-sync question
// again.
const TombstoneRetention = 180 * 24 * time.Hour

// opPrune is the kind of the op that drops tombstones no link needs.
const opPrune = "prune"

// PruneTombstones drops the tombstones every link has received, those with
// a sequence number up to acked (-1 when there are no links), and those
// older than retention. It returns how many it dropped.
func (c *Collection) PruneTombstones(acked int64, retention time.Duration) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := c.now().Add(-retention).UnixMilli()
	o := &op{Kind: opPrune}
	for k, m := range c.tombs {
		if m.seq <= acked || m.stamp.Millis() < cutoff {
			o.Pruned = append(o.Pruned, entryOf(k, m))
		}
	}
	if len(o.Pruned) == 0 {
		return 0, nil
	}
	slices.SortFunc(o.Pruned, func(a, b syncEntry) int { return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.ID, b.ID)) })
	return len(o.Pruned), c.commit(o)
}
