package core

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// Card returns a card.
func (c *Collection) Card(id ID) (Card, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[id]
	if !ok {
		return Card{}, userErr("cardGone", "That card no longer exists.")
	}
	return *card, nil
}

// RenderCard renders a card's question and answer for display in env.
func (c *Collection) RenderCard(id ID, env render.Env) (render.Card, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[id]
	if !ok {
		return render.Card{}, userErr("cardGone", "That card no longer exists.")
	}
	return c.renderCard(card, env), nil
}

// RenderCardTyped is RenderCard with the answer comparing typed, what the
// user typed in, with the card's {{type:…}} field.
func (c *Collection) RenderCardTyped(id ID, typed string, env render.Env) (render.Card, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[id]
	if !ok {
		return render.Card{}, userErr("cardGone", "That card no longer exists.")
	}
	tmpl, css, rn, clozeOrd := c.renderInput(card)
	return render.RenderCardTyped(tmpl, css, rn, clozeOrd, typed, env), nil
}

func (c *Collection) renderCard(card *Card, env render.Env) render.Card {
	tmpl, css, rn, clozeOrd := c.renderInput(card)
	return render.RenderCard(tmpl, css, rn, clozeOrd, env)
}

// renderInput returns what a card is rendered from: its template, the note
// type's CSS, the note and the cloze number.
func (c *Collection) renderInput(card *Card) (render.Template, string, render.Note, int) {
	n := c.notes[card.NoteID]
	nt := c.noteTypes[n.TypeID]
	tmpl, clozeOrd, name := c.template(nt, card)
	rn := render.Note{Fields: fieldMap(nt, n.Fields), Tags: n.Tags, NoteType: nt.Name, CardName: name, CardOrd: card.Ord + 1, Flag: card.Flag,
		LatexSVG: nt.LatexSVG, HasMedia: c.store.HasMedia}
	// {{Deck}} is the home deck, as in Anki.
	if d, ok := c.decks[card.HomeDeck()]; ok {
		rn.Deck = d.Name
	}
	return tmpl, nt.CSS, rn, clozeOrd
}

// template returns the template, cloze number and card name of a card.
func (c *Collection) template(nt *NoteType, card *Card) (render.Template, int, string) {
	if nt.Kind == Cloze {
		t := nt.Templates[0]
		return render.Template{QFmt: t.QFmt, AFmt: t.AFmt}, card.Ord + 1, fmt.Sprintf("Cloze %d", card.Ord+1)
	}
	if card.Ord < len(nt.Templates) {
		t := nt.Templates[card.Ord]
		return render.Template{QFmt: t.QFmt, AFmt: t.AFmt}, 0, t.Name
	}
	return render.Template{}, 0, ""
}

// browserTemplate returns tmpl, the template of card, with its card type's
// browser appearance templates in place of the normal ones where they are
// set.
func (c *Collection) browserTemplate(card *Card, tmpl render.Template) render.Template {
	if t := c.noteTypes[c.notes[card.NoteID].TypeID].cardType(card.Ord); t != nil {
		if strings.TrimSpace(t.BrowserQFmt) != "" {
			tmpl.QFmt = t.BrowserQFmt
		}
		if strings.TrimSpace(t.BrowserAFmt) != "" {
			tmpl.AFmt = t.BrowserAFmt
		}
	}
	return tmpl
}

// QuestionText is a card's question as plain text, for lists.
func (c *Collection) QuestionText(id ID) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[id]
	if !ok {
		return ""
	}
	return questionText(c.renderCard(card, render.Env{}).Question)
}

func questionText(html string) string {
	return collapseSpace(render.StripHTML(html))
}

func (c *Collection) cardCopies(ids []ID) []*Card {
	var out []*Card
	seen := map[ID]bool{}
	for _, id := range ids {
		if card, ok := c.cards[id]; ok && !seen[id] {
			seen[id] = true
			cp := *card
			out = append(out, &cp)
		}
	}
	return out
}

// MoveCards moves cards to another deck, which becomes their home deck:
// cards in filtered decks leave them. Their scheduling is unchanged.
func (c *Collection) MoveCards(ids []ID, deckID ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if d, ok := c.decks[deckID]; !ok {
		return userErr("deckGone", "That deck no longer exists.")
	} else if d.Filter != nil {
		return userErr("moveIntoFiltered", "Cards cannot be moved into a filtered deck; it gathers its cards with its search.")
	}
	o := &op{Kind: "moveCards"}
	for _, card := range c.cardCopies(ids) {
		if card.DeckID != deckID {
			card.returnHome()
			card.DeckID = deckID
			o.Cards = append(o.Cards, card)
		}
	}
	return c.commit(o)
}

// SetSuspended suspends or unsuspends cards. Suspended cards are never
// studied or counted as due.
func (c *Collection) SetSuspended(ids []ID, suspended bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	o := &op{Kind: "suspend"}
	if !suspended {
		o.Kind = "unsuspend"
	}
	for _, card := range c.cardCopies(ids) {
		if card.Suspended != suspended {
			card.Suspended = suspended
			o.Cards = append(o.Cards, card)
		}
	}
	return c.commit(o)
}

// MaxFlag is the highest flag number; flags are 1 to MaxFlag, 0 is none.
const MaxFlag = 7

// SetFlag sets flag on cards, or clears it when every one of them already
// has it (design decision 2). Flag 0 clears the cards' flags. Scheduling is
// unchanged.
func (c *Collection) SetFlag(ids []ID, flag int) error {
	if flag < 0 || flag > MaxFlag {
		return userErr("unknownFlag", "There is no flag %d.", flag)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	cards := c.cardCopies(ids)
	if flag != 0 {
		all := len(cards) > 0
		for _, card := range cards {
			all = all && card.Flag == flag
		}
		if all {
			flag = 0
		}
	}
	o := &op{Kind: "flag"}
	if flag == 0 {
		o.Kind = "unflag"
	}
	for _, card := range cards {
		if card.Flag != flag {
			card.Flag = flag
			o.Cards = append(o.Cards, card)
		}
	}
	return c.commit(o)
}

// BuryCards hides cards from study until the next day rollover.
func (c *Collection) BuryCards(ids []ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	until := scheduler.NextDayStart(c.now(), c.settings.RolloverHour)
	o := &op{Kind: "bury"}
	for _, card := range c.cardCopies(ids) {
		card.BuriedUntil = until
		o.Cards = append(o.Cards, card)
	}
	return c.commit(o)
}

// Unbury makes the buried cards of a deck and its subdecks available
// again: it clears manual buries and stops today's answers so far from
// burying siblings in these decks (design decision 2).
func (c *Collection) Unbury(deckID ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deckID]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	now := c.now()
	o := &op{Kind: "unbury"}
	in := map[ID]bool{}
	for _, t := range c.subtree(d) {
		in[t.ID] = true
		nd := *t
		nd.UnburiedAt = now
		o.Decks = append(o.Decks, &nd)
	}
	var ids []ID
	for _, card := range c.cards {
		if in[card.DeckID] && isBuried(card, now) {
			ids = append(ids, card.ID)
		}
	}
	sortIDs(ids)
	for _, card := range c.cardCopies(ids) {
		card.BuriedUntil = time.Time{}
		o.Cards = append(o.Cards, card)
	}
	return c.commit(o)
}

// isBuried reports whether a card is buried at time now.
func isBuried(card *Card, now time.Time) bool {
	return !card.BuriedUntil.IsZero() && now.Before(card.BuriedUntil)
}

// ResetOptions are the options of ResetCards.
type ResetOptions struct {
	// RestorePosition puts each card back at its original position among
	// new cards (Card.OriginalPos) where it is known; other cards go after
	// all new cards.
	RestorePosition bool
	// ResetCounts sets the cards' review and lapse counts to zero.
	ResetCounts bool
}

// ResetCards makes cards New again: at their original position or at the
// end of the new card queue (see ResetOptions). Their memory state is
// cleared; their review history is kept, and each card gets a manual
// review log entry.
func (c *Collection) ResetCards(ids []ID, opts ResetOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	o := &op{Kind: "resetCards"}
	for _, card := range c.cardCopies(ids) {
		before := card.CardState
		card.CardState = scheduler.CardState{State: scheduler.New}
		if !opts.ResetCounts {
			card.Reps, card.Lapses = before.Reps, before.Lapses
		}
		if opts.RestorePosition && card.OriginalPos != nil {
			card.NewPos = *card.OriginalPos
		} else {
			card.NewPos = c.newPos()
		}
		card.BuriedUntil = time.Time{}
		o.Cards = append(o.Cards, card)
		o.AddReviews = append(o.AddReviews, c.manualEntry(card.ID, before, card.CardState, now))
	}
	if len(o.Cards) == 0 {
		return nil
	}
	rev, err := reviewWrites(o.AddReviews, nil)
	if err != nil {
		return err
	}
	return c.commit(o, rev...)
}

// CardsPerDeck counts the cards directly in each deck.
func (c *Collection) CardsPerDeck() map[ID]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := map[ID]int{}
	for _, card := range c.cards {
		out[card.DeckID]++
	}
	return out
}

// CardCount returns the number of cards in the collection.
func (c *Collection) CardCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cards)
}

// AllCards returns every card, in no particular order.
func (c *Collection) AllCards() []Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Card, 0, len(c.cards))
	for _, card := range c.cards {
		out = append(out, *card)
	}
	return out
}

// EmptyCard is a card whose question is empty: its template no longer
// renders anything for the note, or its cloze number is gone from the note.
type EmptyCard struct {
	CardID ID
	// Name is the card type name, or "Cloze N".
	Name string
}

// EmptyNote is a note with empty cards.
type EmptyNote struct {
	NoteID ID
	// Text is the note's first field as plain text.
	Text  string
	Cards []EmptyCard
	// AllEmpty is set when every card of the note is empty; deleting empty
	// cards keeps one of them, so the note is not deleted.
	AllEmpty bool
}

// EmptyCardList is the result of EmptyCards.
type EmptyCardList struct {
	Notes []EmptyNote
	// Total counts the empty cards; Deletable those DeleteEmptyCards would
	// delete (Total minus one kept card per AllEmpty note).
	Total, Deletable int
}

// EmptyCards lists the empty cards, grouped by note in the order notes were
// added, each note's cards by ord.
func (c *Collection) EmptyCards() EmptyCardList {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ids := make([]ID, 0, len(c.notes))
	for id := range c.notes {
		ids = append(ids, id)
	}
	sortIDs(ids)
	var out EmptyCardList
	for _, nid := range ids {
		cards := c.emptyCardsOf(nid)
		if len(cards) == 0 {
			continue
		}
		n := c.notes[nid]
		nt := c.noteTypes[n.TypeID]
		en := EmptyNote{NoteID: nid, AllEmpty: len(cards) == len(c.cardsByNote[nid])}
		if len(n.Fields) > 0 {
			en.Text = questionText(n.Fields[0])
		}
		for _, card := range cards {
			_, _, name := c.template(nt, card)
			if name == "" {
				name = fmt.Sprintf("Card %d", card.Ord+1)
			}
			en.Cards = append(en.Cards, EmptyCard{CardID: card.ID, Name: name})
		}
		out.Notes = append(out.Notes, en)
		out.Total += len(cards)
		out.Deletable += len(cards)
		if en.AllEmpty {
			out.Deletable--
		}
	}
	return out
}

// emptyCardsOf returns a note's empty cards ordered by ord.
func (c *Collection) emptyCardsOf(nid ID) []*Card {
	n := c.notes[nid]
	want := map[int]bool{}
	for _, ord := range cardOrds(c.noteTypes[n.TypeID], n.Fields) {
		want[ord] = true
	}
	var out []*Card
	for _, cid := range c.cardsByNote[nid] {
		if card := c.cards[cid]; !want[card.Ord] {
			out = append(out, card)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ord < out[j].Ord })
	return out
}

// IsEmptyCard reports whether a card's question is empty.
func (c *Collection) IsEmptyCard(id ID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[id]
	if !ok {
		return false
	}
	for _, e := range c.emptyCardsOf(card.NoteID) {
		if e.ID == id {
			return true
		}
	}
	return false
}

// DeleteEmptyCards deletes those of cards that are empty, in one change.
// Cards that are no longer empty are skipped. A note keeps its first card
// when all its cards would be deleted, so no note is deleted.
func (c *Collection) DeleteEmptyCards(ids []ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	chosen := map[ID]bool{}
	var notes []ID
	for _, id := range ids {
		if card, ok := c.cards[id]; ok && !chosen[id] {
			chosen[id] = true
			notes = append(notes, card.NoteID)
		}
	}
	o := &op{Kind: "deleteEmptyCards"}
	seen := map[ID]bool{}
	for _, nid := range notes {
		if seen[nid] {
			continue
		}
		seen[nid] = true
		var del []ID
		for _, card := range c.emptyCardsOf(nid) {
			if chosen[card.ID] {
				del = append(del, card.ID)
			}
		}
		if len(del) == len(c.cardsByNote[nid]) {
			del = del[1:]
		}
		o.DelCards = append(o.DelCards, del...)
	}
	sortIDs(o.DelCards)
	return c.commit(o)
}
