package core

import (
	"cmp"
	"math"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// Column is a browser column. The browser can sort by any column.
type Column string

const (
	ColQuestion       Column = "question"
	ColAnswer         Column = "answer"
	ColSortField      Column = "sortField"
	ColCardType       Column = "cardType"
	ColNoteType       Column = "noteType"
	ColDeck           Column = "deck"
	ColDue            Column = "due"
	ColInterval       Column = "interval"
	ColDifficulty     Column = "difficulty"
	ColStability      Column = "stability"
	ColRetrievability Column = "retrievability"
	ColReviews        Column = "reviews"
	ColLapses         Column = "lapses"
	ColTags           Column = "tags"
	ColCreated        Column = "created"
	ColCardModified   Column = "cardModified"
	ColNoteModified   Column = "noteModified"
)

// Columns lists every browser column.
var Columns = []Column{ColQuestion, ColAnswer, ColSortField, ColCardType, ColNoteType, ColDeck, ColDue,
	ColInterval, ColDifficulty, ColStability, ColRetrievability, ColReviews, ColLapses, ColTags,
	ColCreated, ColCardModified, ColNoteModified}

// DefaultColumns are the browser's columns until the user changes them.
var DefaultColumns = []Column{ColQuestion, ColDeck, ColDue, ColNoteType}

// Valid reports whether col is a browser column.
func (col Column) Valid() bool { return slices.Contains(Columns, col) }

// textCache keeps per-card text for the browser; entries are dropped when a
// card, its note or its note type changes.
type textCache[T any] struct {
	mu sync.Mutex
	m  map[ID]T
}

func (q *textCache[T]) get(id ID, compute func() T) T {
	q.mu.Lock()
	v, ok := q.m[id]
	q.mu.Unlock()
	if ok {
		return v
	}
	v = compute()
	q.mu.Lock()
	if q.m == nil {
		q.m = map[ID]T{}
	}
	q.m[id] = v
	q.mu.Unlock()
	return v
}

func (q *textCache[T]) invalidate(o *op, cardsByNote map[ID][]ID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(o.NoteTypes)+len(o.DelNoteTypes)+len(o.Decks)+len(o.DelDecks) > 0 {
		q.m = nil // templates or deck names changed
		return
	}
	for _, n := range o.Notes {
		for _, cid := range cardsByNote[n.ID] {
			delete(q.m, cid)
		}
	}
	for _, card := range o.Cards {
		delete(q.m, card.ID)
	}
	for _, id := range o.DelCards {
		delete(q.m, id)
	}
}

// cardText is a card's question and answer as plain text, rendered with
// its card type's browser appearance templates where set.
type cardText struct{ question, answer string }

func (c *Collection) cardText(card *Card) cardText {
	return c.texts.get(card.ID, func() cardText {
		tmpl, css, rn, clozeOrd := c.renderInput(card)
		r := render.RenderCard(c.browserTemplate(card, tmpl), css, rn, clozeOrd, render.Env{})
		q := questionText(r.Question)
		return cardText{q, answerText(q, r.Answer)}
	})
}

// answerText is a rendered answer as plain text without the question it
// repeats: what follows its <hr id=answer> line, or else what follows the
// question when the answer starts with it.
func answerText(question, answer string) string {
	if part := render.AnswerPart(answer); len(part) < len(answer) {
		return questionText(part)
	}
	a := questionText(answer)
	if rest, ok := strings.CutPrefix(a, question); ok && question != "" {
		return strings.TrimSpace(rest)
	}
	return a
}

// sortFieldText is the sort field of a card's note (see SortFieldIndex) as
// plain text.
func (c *Collection) sortFieldText(card *Card) string {
	return c.sortFields.get(card.ID, func() string {
		n := c.notes[card.NoteID]
		if i := c.noteTypes[n.TypeID].SortFieldIndex(); i < len(n.Fields) {
			return questionText(n.Fields[i])
		}
		return ""
	})
}

// cardModified is when a card last changed: its Modified time, or for a
// card stored before that was recorded, its latest review or creation.
func cardModified(card *Card) time.Time {
	if !card.Modified.IsZero() {
		return card.Modified
	}
	created := time.UnixMilli(int64(card.ID))
	if card.LastReview.After(created) {
		return card.LastReview
	}
	return created
}

// Row is one card in the browser list. Rows has it fill only the requested
// columns; the values of other columns are zero.
type Row struct {
	CardID    ID
	NoteID    ID
	State     scheduler.State
	Suspended bool
	Flag      int
	// Marked is set when the note has the tag MarkedTag.
	Marked bool

	// Question, Answer and SortField are plain text, at most 200
	// characters.
	Question, Answer, SortField string
	// CardName is the Card type column.
	CardName string
	NoteType string
	Deck     string
	// Tags are the note's tags, separated by spaces.
	Tags string
	// Due is a scheduled card's due time; Position a New card's position
	// in the new card queue.
	Due      time.Time
	Position int64
	// Interval is a Review card's interval in days; NaN for other cards.
	Interval float64
	// Difficulty (0 to 1), Stability (days) and Retrievability (0 to 1)
	// are NaN when the card has no memory state.
	Difficulty, Stability, Retrievability float64
	Reviews, Lapses                       int
	Created, CardModified, NoteModified   time.Time
}

// rowTextLimit is how many characters of text a row shows at most.
const rowTextLimit = 200

func clip(s string) string {
	if len(s) <= rowTextLimit {
		return s
	}
	if r := []rune(s); len(r) > rowTextLimit {
		return string(r[:rowTextLimit])
	}
	return s
}

// Rows returns browser rows for cards with the values of the given
// columns, or of every column when none is given, in the order of ids;
// unknown IDs are skipped.
func (c *Collection) Rows(ids []ID, columns ...Column) []Row {
	if len(columns) == 0 {
		columns = Columns
	}
	now := c.now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Row, 0, len(ids))
	for _, id := range ids {
		card, ok := c.cards[id]
		if !ok {
			continue
		}
		n := c.notes[card.NoteID]
		row := Row{CardID: id, NoteID: n.ID, State: card.State, Suspended: card.Suspended, Flag: card.Flag,
			Marked: isMarked(n.Tags)}
		for _, col := range columns {
			c.fillColumn(&row, col, card, n, now)
		}
		out = append(out, row)
	}
	return out
}

// fillColumn sets the value of column col in row. Callers hold a lock.
func (c *Collection) fillColumn(row *Row, col Column, card *Card, n *Note, now time.Time) {
	switch col {
	case ColQuestion:
		row.Question = clip(c.cardText(card).question)
	case ColAnswer:
		row.Answer = clip(c.cardText(card).answer)
	case ColSortField:
		row.SortField = clip(c.sortFieldText(card))
	case ColCardType:
		_, _, row.CardName = c.template(c.noteTypes[n.TypeID], card)
	case ColNoteType:
		row.NoteType = c.noteTypes[n.TypeID].Name
	case ColDeck:
		row.Deck = c.deckLabel(card)
	case ColDue:
		if card.State == scheduler.New {
			row.Position = card.NewPos
		} else {
			row.Due = card.Due
		}
	case ColInterval:
		row.Interval = reviewInterval(card)
	case ColDifficulty:
		row.Difficulty = difficulty(card)
	case ColStability:
		row.Stability = stability(card)
	case ColRetrievability:
		row.Retrievability = retrievability(card, now)
	case ColReviews:
		row.Reviews = card.Reps
	case ColLapses:
		row.Lapses = card.Lapses
	case ColTags:
		row.Tags = clip(strings.Join(n.Tags, " "))
	case ColCreated:
		row.Created = time.UnixMilli(int64(card.ID))
	case ColCardModified:
		row.CardModified = cardModified(card)
	case ColNoteModified:
		row.NoteModified = n.Modified
	}
}

// sortKey is what a card sorts by in a column: group, then num, then str.
// Cards without a value (missing) come last in both directions.
type sortKey struct {
	missing bool
	group   int
	num     float64
	str     string
}

func numKey(v float64) sortKey { return sortKey{missing: math.IsNaN(v), num: v} }

func timeKey(t time.Time) sortKey { return sortKey{num: float64(t.UnixMilli())} }

// sortKey returns card's key in column col. Callers hold a lock.
func (c *Collection) sortKey(card *Card, col Column, now time.Time) sortKey {
	n := c.notes[card.NoteID]
	switch col {
	case ColQuestion:
		return sortKey{str: strings.ToLower(c.cardText(card).question)}
	case ColAnswer:
		return sortKey{str: strings.ToLower(c.cardText(card).answer)}
	case ColSortField:
		return sortKey{str: strings.ToLower(c.sortFieldText(card))}
	case ColCardType:
		// By name, cloze cards and same-named card types by number.
		nt := c.noteTypes[n.TypeID]
		if nt.Kind == Cloze {
			return sortKey{str: "cloze", num: float64(card.Ord)}
		}
		_, _, name := c.template(nt, card)
		return sortKey{str: strings.ToLower(name), num: float64(card.Ord)}
	case ColNoteType:
		return sortKey{str: strings.ToLower(c.noteTypes[n.TypeID].Name)}
	case ColDeck:
		return sortKey{str: strings.ToLower(c.deckLabel(card))}
	case ColDue:
		// Scheduled cards by due time, then new cards in queue order.
		if card.State == scheduler.New {
			return sortKey{group: 1, num: float64(card.NewPos)}
		}
		return timeKey(card.Due)
	case ColInterval:
		return numKey(reviewInterval(card))
	case ColDifficulty:
		return numKey(difficulty(card))
	case ColStability:
		return numKey(stability(card))
	case ColRetrievability:
		return numKey(retrievability(card, now))
	case ColReviews:
		return sortKey{num: float64(card.Reps)}
	case ColLapses:
		return sortKey{num: float64(card.Lapses)}
	case ColTags:
		return sortKey{str: strings.ToLower(strings.Join(n.Tags, " ")), missing: len(n.Tags) == 0}
	case ColCardModified:
		return timeKey(cardModified(card))
	case ColNoteModified:
		return timeKey(n.Modified)
	}
	return sortKey{} // ColCreated: by ID alone
}

// sortCards sorts ids by column col; cards with equal values stay in
// creation order (reversed when descending). Callers hold a lock.
func (c *Collection) sortCards(ids []ID, col Column, descending bool, now time.Time) {
	c.sortIDs(ids, func(id ID) sortKey { return c.sortKey(c.cards[id], col, now) }, descending)
}

// compare orders two sort keys; a missing value comes last in both
// directions, so descending only applies to r when both have one.
func (a sortKey) compare(b sortKey) (r int, missing bool) {
	if a.missing != b.missing {
		if a.missing {
			return 1, true
		}
		return -1, true
	}
	return cmp.Or(cmp.Compare(a.group, b.group), cmp.Compare(a.num, b.num), strings.Compare(a.str, b.str)), false
}

// sortIDs sorts ids by their keys; equal keys keep ID order (reversed when
// descending). Callers hold a lock.
func (c *Collection) sortIDs(ids []ID, key func(ID) sortKey, descending bool) {
	type entry struct {
		id  ID
		key sortKey
	}
	es := make([]entry, len(ids))
	keys := func(from, to int, yield func()) {
		for i := from; i < to; i++ {
			es[i] = entry{ids[i], key(ids[i])}
			if i%2000 == 1999 {
				yield()
			}
		}
	}
	// Text keys render every card the first time: spread that over the
	// threads Go has, or let the web page respond meanwhile.
	if workers := runtime.GOMAXPROCS(0); workers > 1 && len(es) > 1000 {
		var wg sync.WaitGroup
		chunk := (len(es) + workers - 1) / workers
		for from := 0; from < len(es); from += chunk {
			wg.Go(func() { keys(from, min(from+chunk, len(es)), func() {}) })
		}
		wg.Wait()
	} else {
		keys(0, len(es), c.yield)
	}
	slices.SortFunc(es, func(a, b entry) int {
		r, missing := a.key.compare(b.key)
		if missing {
			return r
		}
		r = cmp.Or(r, cmp.Compare(a.id, b.id))
		if descending {
			return -r
		}
		return r
	})
	for i, e := range es {
		ids[i] = e.id
	}
}
