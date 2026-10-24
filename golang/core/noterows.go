package core

import (
	"strings"
	"time"

	"cardflip/golang/scheduler"
)

// NoteColumns are the columns the browser's Notes mode can sort by.
var NoteColumns = []Column{ColCreated, ColSortField, ColDue}

// SearchNotes returns the IDs of the notes with a card matching query,
// sorted by a column of NoteColumns: created, sort field (see
// NoteType.SortFieldIndex) or the earliest due of the note's cards. An
// invalid query returns a *search.Error naming the bad term.
func (c *Collection) SearchNotes(query string, order Column, descending bool) ([]ID, error) {
	switch order {
	case ColCreated, ColSortField, ColDue:
	default:
		return nil, userErr("unknownNoteSort", "Notes can't be sorted by %q.", order)
	}
	now := c.now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	cards, err := c.match(query, now)
	if err != nil {
		return nil, err
	}
	seen := make(map[ID]bool, len(cards))
	var ids []ID
	for _, id := range cards {
		if n := c.cards[id].NoteID; !seen[n] {
			seen[n] = true
			ids = append(ids, n)
		}
	}
	c.sortIDs(ids, func(id ID) sortKey { return c.noteSortKey(id, order, now) }, descending)
	return ids, nil
}

// noteSortKey is a note's key in column col. Callers hold a lock.
func (c *Collection) noteSortKey(id ID, col Column, now time.Time) sortKey {
	switch col {
	case ColSortField:
		if first := c.firstCard(id); first != nil {
			return c.sortKey(first, ColSortField, now)
		}
	case ColDue:
		if card := c.earliestDue(id, now); card != nil {
			return c.sortKey(card, ColDue, now)
		}
	}
	return sortKey{} // ColCreated: by ID alone
}

// firstCard is the card of a note with the lowest ord, which represents
// the note in Notes mode; nil when it has none. Callers hold a lock.
func (c *Collection) firstCard(note ID) *Card {
	var first *Card
	for _, id := range c.cardsByNote[note] {
		if card := c.cards[id]; first == nil || card.Ord < first.Ord {
			first = card
		}
	}
	return first
}

// earliestDue is the card of a note that comes first in the Due column:
// the scheduled card due first, or else the New card first in the queue.
// Callers hold a lock.
func (c *Collection) earliestDue(note ID, now time.Time) *Card {
	var best *Card
	var bestKey sortKey
	for _, id := range c.cardsByNote[note] {
		card := c.cards[id]
		key := c.sortKey(card, ColDue, now)
		if r, _ := key.compare(bestKey); best == nil || r < 0 || r == 0 && card.ID < best.ID {
			best, bestKey = card, key
		}
	}
	return best
}

// NoteRow is one note in the browser's Notes mode.
type NoteRow struct {
	NoteID ID
	// CardID is the note's first card, whose details the browser shows.
	CardID ID
	// SortField is the sort field as plain text, at most 200 characters.
	SortField string
	NoteType  string
	Cards     int
	// Deck is the deck of the note's cards; empty when they are in
	// several decks (MultipleDecks).
	Deck          string
	MultipleDecks bool
	// Tags are the note's tags, separated by spaces.
	Tags string
	// DueState is the state of the card due first (see earliestDue): Due
	// is its due time, or Position its place in the new card queue when it
	// is New.
	DueState scheduler.State
	Due      time.Time
	Position int64
	// Suspended is set when all of the note's cards are suspended.
	Suspended bool
	// Marked is set when the note has the tag MarkedTag.
	Marked  bool
	Created time.Time
}

// NoteRows returns Notes mode rows for notes, in the order of ids; unknown
// IDs and notes without cards are skipped.
func (c *Collection) NoteRows(ids []ID) []NoteRow {
	now := c.now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]NoteRow, 0, len(ids))
	for _, id := range ids {
		n, ok := c.notes[id]
		first := c.firstCard(id)
		if !ok || first == nil {
			continue
		}
		row := NoteRow{NoteID: id, CardID: first.ID, SortField: clip(c.sortFieldText(first)),
			NoteType: c.noteTypes[n.TypeID].Name, Tags: clip(strings.Join(n.Tags, " ")),
			Suspended: true, Marked: isMarked(n.Tags), Created: time.UnixMilli(int64(id))}
		for _, cid := range c.cardsByNote[id] {
			card := c.cards[cid]
			row.Cards++
			row.Suspended = row.Suspended && card.Suspended
			if deck := c.deckLabel(card); row.Deck == "" && !row.MultipleDecks {
				row.Deck = deck
			} else if deck != row.Deck {
				row.Deck, row.MultipleDecks = "", true
			}
		}
		due := c.earliestDue(id, now)
		row.DueState = due.State
		if due.State == scheduler.New {
			row.Position = due.NewPos
		} else {
			row.Due = due.Due
		}
		out = append(out, row)
	}
	return out
}

// CardsOfNotes returns the IDs of all cards of the notes, note by note;
// unknown notes are skipped.
func (c *Collection) CardsOfNotes(ids []ID) []ID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []ID
	for _, id := range ids {
		out = append(out, c.cardsByNote[id]...)
	}
	return out
}
