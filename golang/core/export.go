package core

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
)

// ExportScope selects what an export contains.
type ExportScope struct {
	// DeckID exports the cards of the deck and its subdecks, and their
	// notes.
	DeckID ID
	// NoteIDs exports these notes with all their cards.
	NoteIDs []ID
	// CardIDs exports these cards and their notes.
	CardIDs []ID
	// All exports the whole collection.
	All bool
}

// ExportSet is a consistent copy of the part of the collection an export
// writes, everything sorted by ID.
type ExportSet struct {
	Settings Settings
	// Presets are the presets of the exported decks.
	Presets []Preset
	// Decks are the exported decks: the deck tree of a deck export, the
	// decks of the cards otherwise (every deck for the whole collection).
	// Filtered decks are left out; their cards are in their home decks.
	Decks     []Deck
	NoteTypes []NoteType
	Notes     []Note
	Cards     []Card
	// Media are the stored media files the notes and note types reference.
	Media []string
}

// Export returns the part of the collection scope selects.
func (c *Collection) Export(scope ExportScope) (*ExportSet, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	set := &ExportSet{Settings: c.settings}
	decks := map[ID]bool{}
	notes := map[ID]bool{}
	var cards []*Card
	switch {
	case scope.All:
		for id := range c.decks {
			decks[id] = true
		}
		for id := range c.notes {
			notes[id] = true
		}
		for _, card := range c.cards {
			cards = append(cards, card)
		}
	case scope.DeckID != 0:
		root, ok := c.decks[scope.DeckID]
		if !ok {
			return nil, userErr("deckGone", "That deck no longer exists.")
		}
		for _, d := range c.subtree(root) {
			decks[d.ID] = true
		}
		for _, card := range c.cards {
			if decks[card.DeckID] || decks[card.HomeDeck()] {
				cards = append(cards, card)
				notes[card.NoteID] = true
			}
		}
	case len(scope.CardIDs) > 0:
		seen := map[ID]bool{}
		for _, id := range scope.CardIDs {
			card, ok := c.cards[id]
			if !ok || seen[id] {
				continue
			}
			seen[id] = true
			cards = append(cards, card)
			notes[card.NoteID] = true
			decks[card.DeckID] = true
		}
		if len(cards) == 0 {
			return nil, userErr("noCardsSelected", "No cards are selected.")
		}
	default:
		for _, id := range scope.NoteIDs {
			if _, ok := c.notes[id]; !ok {
				continue
			}
			notes[id] = true
			for _, cid := range c.cardsByNote[id] {
				if card, ok := c.cards[cid]; ok {
					cards = append(cards, card)
					decks[card.DeckID] = true
				}
			}
		}
		if len(notes) == 0 {
			return nil, userErr("noNotesSelected", "No notes are selected.")
		}
	}

	// Filtered decks are not exported: their cards go to their home decks.
	for _, card := range cards {
		decks[card.HomeDeck()] = true
	}
	maps.DeleteFunc(decks, func(id ID, _ bool) bool { return c.decks[id].Filter != nil })
	presets := map[ID]bool{}
	for id := range decks {
		d := c.decks[id]
		cp := *d
		cp.Limits = DeckLimits{New: copyLimit(d.Limits.New), Review: copyLimit(d.Limits.Review)}
		cp.Today = TodayLimits{}
		set.Decks = append(set.Decks, cp)
		presets[c.presetOf(d).ID] = true
	}
	for id := range presets {
		p := *c.preset(id)
		p.Options.LearningSteps = slices.Clone(p.Options.LearningSteps)
		p.Options.RelearningSteps = slices.Clone(p.Options.RelearningSteps)
		set.Presets = append(set.Presets, p)
	}
	types := map[ID]bool{}
	var htmls []string
	for id := range notes {
		n := c.notes[id]
		set.Notes = append(set.Notes, cloneNote(n))
		types[n.TypeID] = true
		htmls = append(htmls, n.Fields...)
	}
	for id := range types {
		nt := c.noteTypes[id]
		set.NoteTypes = append(set.NoteTypes, *cloneNoteType(nt))
		for _, t := range nt.Templates {
			htmls = append(htmls, t.QFmt, t.AFmt)
		}
	}
	for _, card := range cards {
		cp := *card
		cp.returnHome()
		if card.Memory != nil {
			m := *card.Memory
			cp.Memory = &m
		}
		set.Cards = append(set.Cards, cp)
	}
	seen := map[string]bool{}
	for _, name := range MediaNames(htmls...) {
		if !seen[name] && c.store.HasMedia(name) {
			seen[name] = true
			set.Media = append(set.Media, name)
		}
	}
	sort.Strings(set.Media)
	sort.Slice(set.Presets, func(i, j int) bool { return set.Presets[i].ID < set.Presets[j].ID })
	sort.Slice(set.Decks, func(i, j int) bool { return set.Decks[i].ID < set.Decks[j].ID })
	sort.Slice(set.NoteTypes, func(i, j int) bool { return set.NoteTypes[i].ID < set.NoteTypes[j].ID })
	sort.Slice(set.Notes, func(i, j int) bool { return set.Notes[i].ID < set.Notes[j].ID })
	sort.Slice(set.Cards, func(i, j int) bool { return set.Cards[i].ID < set.Cards[j].ID })
	return set, nil
}

// ExportReviews returns the whole stored review log of the cards, in ID
// (time) order.
func (c *Collection) ExportReviews(cards []Card) ([]ReviewEntry, error) {
	want := make(map[ID]bool, len(cards))
	for _, card := range cards {
		want[card.ID] = true
	}
	months, err := c.store.RevlogMonths()
	if err != nil {
		return nil, err
	}
	var out []ReviewEntry
	for _, m := range months {
		c.yield()
		entries, err := c.store.ReadRevlog(m)
		if err != nil {
			return nil, err
		}
		for i, raw := range entries {
			var e ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return nil, fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, i, err)
			}
			if want[e.CardID] {
				out = append(out, e)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
