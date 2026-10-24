//go:build !js

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
	"cardflip/golang/stats"
)

// Paging of the search tools.
const (
	mcpDefaultLimit = 50
	mcpMaxLimit     = 500
)

// deckID returns the ID of the deck with the full name name ("A::B"),
// ignoring case and spaces around "::".
func deckID(c *core.Collection, name string) (core.ID, error) {
	want := normalizeDeckName(name)
	for _, d := range c.Decks() {
		if strings.EqualFold(normalizeDeckName(d.Name), want) {
			return d.ID, nil
		}
	}
	return 0, fmt.Errorf("There is no deck named %q. list_decks lists the decks.", name)
}

func normalizeDeckName(name string) string {
	parts := strings.Split(name, core.DeckSeparator)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return strings.Join(parts, core.DeckSeparator)
}

// noteTypeNamed returns the note type called name, ignoring case.
func (t *mcpTools) noteTypeNamed(name string) (NoteType, error) {
	types, err := t.notes.Types()
	if err != nil {
		return NoteType{}, err
	}
	for _, nt := range types {
		if nt.Name == name {
			return nt, nil
		}
	}
	for _, nt := range types {
		if strings.EqualFold(nt.Name, name) {
			return nt, nil
		}
	}
	return NoteType{}, fmt.Errorf("There is no note type named %q. list_note_types lists the note types.", name)
}

// page returns the IDs of a page and the offset of the next one (nil after
// the last page).
func page(ids []int64, limit, offset int) ([]int64, *int, error) {
	if limit <= 0 {
		limit = mcpDefaultLimit
	}
	limit = min(limit, mcpMaxLimit)
	if offset < 0 {
		return nil, nil, fmt.Errorf("The offset must not be negative.")
	}
	if offset >= len(ids) {
		return nil, nil, nil
	}
	end := min(offset+limit, len(ids))
	if end < len(ids) {
		return ids[offset:end], &end, nil
	}
	return ids[offset:end], nil, nil
}

func checkIDCount(n int, what string) error {
	if n == 0 {
		return fmt.Errorf("Give at least one %s ID.", what)
	}
	if n > mcpMaxLimit {
		return fmt.Errorf("Give at most %d %s IDs per call.", mcpMaxLimit, what)
	}
	return nil
}

type mcpNoInput struct{}

type mcpDeck struct {
	ID int64 `json:"id"`
	// Name is the full name, subdecks after "::".
	Name     string `json:"name" jsonschema:"the deck's full name; subdecks follow their parent after ::"`
	New      int    `json:"new" jsonschema:"new cards to study today, subdecks included"`
	Learning int    `json:"learning" jsonschema:"learning cards due today, subdecks included"`
	Review   int    `json:"review" jsonschema:"review cards due today, subdecks included"`
	Cards    int    `json:"cards" jsonschema:"cards in the deck and its subdecks"`
	Filtered bool   `json:"filtered,omitempty" jsonschema:"a filtered deck, which cannot get new notes"`
}

type mcpDeckList struct {
	Decks []mcpDeck `json:"decks"`
}

type mcpDeckInput struct {
	Deck string `json:"deck" jsonschema:"the deck's full name, e.g. Spanish::Verbs"`
}

type mcpDeckOverview struct {
	Deck        string     `json:"deck"`
	New         int        `json:"new"`
	Learning    int        `json:"learning"`
	Review      int        `json:"review"`
	Buried      int        `json:"buried"`
	Description string     `json:"description,omitempty"`
	NextDue     *time.Time `json:"nextDue,omitempty" jsonschema:"when a card is next due, when none is due now"`
}

type mcpTemplate struct {
	Name  string `json:"name"`
	Front string `json:"front" jsonschema:"the front (question) template"`
	Back  string `json:"back" jsonschema:"the back (answer) template"`
}

type mcpNoteType struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Cloze     bool          `json:"cloze,omitempty" jsonschema:"a cloze note type: one card per cloze number {{c1::...}}"`
	Fields    []string      `json:"fields"`
	Templates []mcpTemplate `json:"templates" jsonschema:"the card types; each makes one card per note"`
	CSS       string        `json:"css"`
	Notes     int           `json:"notes"`
}

type mcpNoteTypeList struct {
	NoteTypes []mcpNoteType `json:"noteTypes"`
}

func mcpNoteTypeOf(nt NoteType) mcpNoteType {
	out := mcpNoteType{ID: nt.ID, Name: nt.Name, Cloze: nt.Cloze, Fields: nt.Fields, CSS: nt.CSS, Notes: nt.NoteCount,
		Templates: []mcpTemplate{}}
	for _, tm := range nt.Templates {
		out.Templates = append(out.Templates, mcpTemplate{Name: tm.Name, Front: tm.QFmt, Back: tm.AFmt})
	}
	return out
}

type mcpTagList struct {
	Tags []string `json:"tags"`
}

type mcpSearchInput struct {
	Query  string `json:"query" jsonschema:"a search in the app's syntax (as Anki's), e.g. deck:Spanish is:due tag:verb; empty finds everything"`
	Limit  int    `json:"limit,omitempty" jsonschema:"results per call, 50 when not given, at most 500"`
	Offset int    `json:"offset,omitempty" jsonschema:"results to skip, from nextOffset of the previous call"`
}

type mcpCardRow struct {
	CardID   int64      `json:"cardId"`
	NoteID   int64      `json:"noteId"`
	Deck     string     `json:"deck"`
	NoteType string     `json:"noteType"`
	CardType string     `json:"cardType"`
	Question string     `json:"question" jsonschema:"the question as plain text"`
	Answer   string     `json:"answer" jsonschema:"the answer as plain text, without the question"`
	State    string     `json:"state" jsonschema:"new, learning, review or relearning"`
	Due      *time.Time `json:"due,omitempty"`
	Position *int64     `json:"position,omitempty" jsonschema:"a new card's position in the new card queue"`
	Suspend  bool       `json:"suspended,omitempty"`
	Flag     int        `json:"flag,omitempty" jsonschema:"0 for none, 1 red, 2 orange, 3 green, 4 blue, 5 pink, 6 turquoise, 7 purple"`
	Tags     []string   `json:"tags"`
}

type mcpCardPage struct {
	Total      int          `json:"total" jsonschema:"all matching cards"`
	Cards      []mcpCardRow `json:"cards"`
	NextOffset *int         `json:"nextOffset,omitempty" jsonschema:"the offset of the next page; missing on the last page"`
}

type mcpNote struct {
	NoteID   int64             `json:"noteId"`
	NoteType string            `json:"noteType"`
	Fields   map[string]string `json:"fields" jsonschema:"the fields by name, as HTML"`
	Tags     []string          `json:"tags"`
	Cards    []int64           `json:"cards" jsonschema:"the note's card IDs, by card type"`
}

type mcpNotePage struct {
	Total      int       `json:"total" jsonschema:"all matching notes"`
	Notes      []mcpNote `json:"notes"`
	NextOffset *int      `json:"nextOffset,omitempty" jsonschema:"the offset of the next page; missing on the last page"`
}

type mcpNoteIDsInput struct {
	NoteIDs []int64 `json:"noteIds" jsonschema:"note IDs, at most 500"`
}

type mcpNoteList struct {
	Notes []mcpNote `json:"notes"`
}

type mcpCardIDsInput struct {
	CardIDs []int64 `json:"cardIds" jsonschema:"card IDs, at most 500"`
}

type mcpCard struct {
	CardID   int64             `json:"cardId"`
	NoteID   int64             `json:"noteId"`
	NoteType string            `json:"noteType"`
	Deck     string            `json:"deck"`
	CardType string            `json:"cardType"`
	Fields   map[string]string `json:"fields" jsonschema:"the note's fields by name, as HTML"`
	Tags     []string          `json:"tags"`
	Question string            `json:"question" jsonschema:"the question as plain text"`
	Answer   string            `json:"answer" jsonschema:"the answer as plain text, without the question"`
	State    string            `json:"state" jsonschema:"new, learning, review or relearning"`
	Due      *time.Time        `json:"due,omitempty"`
	Position *int64            `json:"position,omitempty" jsonschema:"a new card's position in the new card queue"`
	Interval int               `json:"intervalDays" jsonschema:"days until the next review after the last one"`
	Reviews  int               `json:"reviews"`
	Lapses   int               `json:"lapses"`
	Suspend  bool              `json:"suspended"`
	Flag     int               `json:"flag" jsonschema:"0 for none, 1 red, 2 orange, 3 green, 4 blue, 5 pink, 6 turquoise, 7 purple"`
}

type mcpCardList struct {
	Cards []mcpCard `json:"cards"`
}

type mcpStatsInput struct {
	Deck string `json:"deck,omitempty" jsonschema:"a deck's full name for that deck and its subdecks; the whole collection when not given"`
}

type mcpStats struct {
	Scope        string `json:"scope" jsonschema:"the deck, or collection"`
	StudiedToday int    `json:"studiedToday" jsonschema:"cards answered today"`
	SecondsToday int    `json:"secondsToday" jsonschema:"time spent answering today, in seconds"`
	New          int    `json:"new"`
	Learning     int    `json:"learning" jsonschema:"learning and relearning cards"`
	Review       int    `json:"review"`
	Suspended    int    `json:"suspended"`
	Buried       int    `json:"buried"`
	Total        int    `json:"total"`
	// DueNext30Days[i] is the number of cards due i days from today.
	DueNext30Days []int `json:"dueNext30Days" jsonschema:"cards due on each of the next 30 days; index 0 is today"`
	Overdue       int   `json:"overdue"`
}

func addReadTools(server *mcp.Server, t *mcpTools) {
	readTool(server, t, "list_decks", "List all decks with today's due counts and card totals (each including subdecks).",
		func(mcpNoInput) (mcpDeckList, error) {
			tree, err := t.decks.Tree()
			if err != nil {
				return mcpDeckList{}, err
			}
			out := mcpDeckList{Decks: []mcpDeck{}}
			for _, d := range tree {
				out.Decks = append(out.Decks, mcpDeck{ID: d.ID, Name: d.FullName, New: d.Counts.New, Learning: d.Counts.Learning,
					Review: d.Counts.Review, Cards: d.Cards, Filtered: d.Filtered})
			}
			return out, nil
		})
	readTool(server, t, "get_deck_overview", "Show a deck's overview: today's due cards, buried cards, its description, and when the next card is due.",
		func(in mcpDeckInput) (mcpDeckOverview, error) {
			c, err := t.col()
			if err != nil {
				return mcpDeckOverview{}, err
			}
			id, err := deckID(c, in.Deck)
			if err != nil {
				return mcpDeckOverview{}, err
			}
			o, err := t.decks.Overview(int64(id))
			if err != nil {
				return mcpDeckOverview{}, err
			}
			d, _ := c.Deck(id)
			return mcpDeckOverview{Deck: d.Name, New: o.Counts.New, Learning: o.Counts.Learning, Review: o.Counts.Review,
				Buried: o.Buried, Description: o.Description, NextDue: o.NextDue}, nil
		})
	readTool(server, t, "list_note_types", "List the note types with their fields, card templates and CSS.",
		func(mcpNoInput) (mcpNoteTypeList, error) {
			types, err := t.notes.Types()
			if err != nil {
				return mcpNoteTypeList{}, err
			}
			out := mcpNoteTypeList{NoteTypes: []mcpNoteType{}}
			for _, nt := range types {
				out.NoteTypes = append(out.NoteTypes, mcpNoteTypeOf(nt))
			}
			return out, nil
		})
	readTool(server, t, "list_tags", "List the tags in use.",
		func(mcpNoInput) (mcpTagList, error) {
			tags, err := t.notes.Tags()
			if tags == nil {
				tags = []string{}
			}
			return mcpTagList{Tags: tags}, err
		})
	readTool(server, t, "search_cards", "Search cards with the app's search syntax, oldest first, a page at a time. Returns each card's question and answer as text.",
		func(in mcpSearchInput) (mcpCardPage, error) {
			all, err := t.browse.Search(in.Query, string(core.ColCreated), false)
			if err != nil {
				return mcpCardPage{}, err
			}
			ids, next, err := page(all, in.Limit, in.Offset)
			if err != nil {
				return mcpCardPage{}, err
			}
			out := mcpCardPage{Total: len(all), Cards: []mcpCardRow{}, NextOffset: next}
			rows, err := t.browse.Rows(ids, []string{"question", "answer", "cardType", "noteType", "deck", "due", "tags"})
			if err != nil {
				return mcpCardPage{}, err
			}
			for _, r := range rows {
				out.Cards = append(out.Cards, mcpCardRow{CardID: r.CardID, NoteID: r.NoteID, Deck: r.Deck, NoteType: r.NoteType,
					CardType: r.CardType, Question: r.Question, Answer: r.Answer, State: r.State, Due: r.Due, Position: r.Position,
					Suspend: r.Suspended, Flag: r.Flag, Tags: splitTags(r.Tags)})
			}
			return out, nil
		})
	readTool(server, t, "search_notes", "Search notes (those with a card matching the search) with the app's search syntax, oldest first, a page at a time. Returns each note's fields by name.",
		func(in mcpSearchInput) (mcpNotePage, error) {
			all, err := t.browse.SearchNotes(in.Query, string(core.ColCreated), false)
			if err != nil {
				return mcpNotePage{}, err
			}
			ids, next, err := page(all, in.Limit, in.Offset)
			if err != nil {
				return mcpNotePage{}, err
			}
			notes, err := t.getNotes(ids)
			return mcpNotePage{Total: len(all), Notes: notes, NextOffset: next}, err
		})
	readTool(server, t, "get_notes", "Get notes by ID: note type, fields by name, tags and card IDs.",
		func(in mcpNoteIDsInput) (mcpNoteList, error) {
			if err := checkIDCount(len(in.NoteIDs), "note"); err != nil {
				return mcpNoteList{}, err
			}
			notes, err := t.getNotes(in.NoteIDs)
			return mcpNoteList{Notes: notes}, err
		})
	readTool(server, t, "get_cards", "Get cards by ID: the note's fields and tags, deck, card type, scheduling, and the question and answer as plain text.",
		func(in mcpCardIDsInput) (mcpCardList, error) {
			if err := checkIDCount(len(in.CardIDs), "card"); err != nil {
				return mcpCardList{}, err
			}
			c, err := t.col()
			if err != nil {
				return mcpCardList{}, err
			}
			out := mcpCardList{Cards: []mcpCard{}}
			for _, id := range in.CardIDs {
				d, err := t.browse.Card(id)
				if err != nil {
					return mcpCardList{}, fmt.Errorf("Card %d: %w", id, err)
				}
				q, a, err := t.cardText(c, core.ID(id))
				if err != nil {
					return mcpCardList{}, fmt.Errorf("Card %d: %w", id, err)
				}
				out.Cards = append(out.Cards, mcpCard{CardID: id, NoteID: d.Note.ID, NoteType: d.NoteType.Name, Deck: d.Deck,
					CardType: d.CardName, Fields: namedFields(d.NoteType.Fields, d.Note.Fields), Tags: nonNil(d.Note.Tags),
					Question: q, Answer: a, State: d.State, Due: d.Due, Position: d.Position, Interval: d.Interval,
					Reviews: d.Reviews, Lapses: d.Lapses, Suspend: d.Suspended, Flag: d.Flag})
			}
			return out, nil
		})
	readTool(server, t, "get_stats", "Collection or deck statistics: today's studied cards and time, cards by state, and the number of cards due on each of the next 30 days.",
		func(in mcpStatsInput) (mcpStats, error) {
			scope := stats.Scope{Kind: stats.ScopeCollection}
			out := mcpStats{Scope: "collection"}
			if in.Deck != "" {
				c, err := t.col()
				if err != nil {
					return mcpStats{}, err
				}
				id, err := deckID(c, in.Deck)
				if err != nil {
					return mcpStats{}, err
				}
				d, _ := c.Deck(id)
				scope, out.Scope = stats.Scope{Kind: stats.ScopeDeck, DeckID: int64(id)}, d.Name
			}
			today, err := t.stats.Today(scope, stats.PeriodMonth)
			if err != nil {
				return mcpStats{}, err
			}
			counts, err := t.stats.Counts(scope, stats.PeriodMonth)
			if err != nil {
				return mcpStats{}, err
			}
			due, err := t.stats.FutureDue(scope, stats.PeriodMonth)
			if err != nil {
				return mcpStats{}, err
			}
			out.StudiedToday, out.SecondsToday = today.Cards, int(today.Time.Seconds())
			out.New, out.Learning, out.Review = counts.New, counts.Learning+counts.Relearning, counts.Young+counts.Mature
			out.Suspended, out.Buried, out.Total = counts.Suspended, counts.Buried, counts.Total
			out.DueNext30Days = make([]int, 30)
			copy(out.DueNext30Days, due.Days)
			out.Overdue = due.Overdue
			return out, nil
		})
}

// getNotes returns notes by ID, failing on the first that does not exist.
func (t *mcpTools) getNotes(ids []int64) ([]mcpNote, error) {
	types, err := t.notes.Types()
	if err != nil {
		return nil, err
	}
	byID := map[int64]NoteType{}
	for _, nt := range types {
		byID[nt.ID] = nt
	}
	out := []mcpNote{}
	for _, id := range ids {
		n, err := t.notes.Get(id)
		if err != nil {
			return nil, fmt.Errorf("Note %d: %w", id, err)
		}
		nt := byID[n.TypeID]
		out = append(out, mcpNote{NoteID: n.ID, NoteType: nt.Name, Fields: namedFields(nt.Fields, n.Fields),
			Tags: nonNil(n.Tags), Cards: nonNil(n.CardIDs)})
	}
	return out, nil
}

func splitTags(tags string) []string { return nonNil(strings.Fields(tags)) }

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
