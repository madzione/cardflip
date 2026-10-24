package core

import (
	"cmp"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"cardflip/golang/scheduler"
)

// errFilteredTarget is the error for adding cards to filtered deck d.
func errFilteredTarget(d *Deck) error {
	return userErr("filteredNoNotes", "%q is a filtered deck, so notes cannot be added to it.", d.Name)
}

// checkParents rejects a deck name inside a filtered deck: filtered decks
// have no subdecks. Callers hold a lock.
func (c *Collection) checkParents(name string) error {
	for p := parentName(name); p != ""; p = parentName(p) {
		if d := c.deckByName(p); d != nil && d.Filter != nil {
			return userErr("filteredNoSubdecks", "%q is a filtered deck, which cannot have subdecks.", d.Name)
		}
	}
	return nil
}

// deckLabel names a card's deck for display: a card in a filtered deck
// shows its home deck in parentheses, as in Anki. Callers hold a lock.
func (c *Collection) deckLabel(card *Card) string {
	name := c.decks[card.DeckID].Name
	if h, ok := c.decks[card.HomeDeckID]; ok && card.HomeDeckID != 0 {
		return name + " (" + h.Name + ")"
	}
	return name
}

// DeckLabel names a card's deck for display (see deckLabel).
func (c *Collection) DeckLabel(id ID) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	card, ok := c.cards[id]
	if !ok {
		return ""
	}
	return c.deckLabel(card)
}

// ValidateFilterConfig checks a filtered deck's options against their
// allowed ranges. The searches are checked when the deck is built.
func ValidateFilterConfig(cfg FilterConfig) error {
	if len(cfg.Terms) < 1 || len(cfg.Terms) > 2 {
		return userErr("filteredFilterCount", "A filtered deck has one or two filters.")
	}
	for _, t := range cfg.Terms {
		if t.Limit < 1 || t.Limit > MaxFilterLimit {
			return userErr("cardCountRange", "The number of cards must be between 1 and %d.", MaxFilterLimit)
		}
		switch t.Order {
		case FilterOldestSeen, FilterRandom, FilterIntervalsAscending, FilterIntervalsDescending, FilterMostLapses,
			FilterAdded, FilterDue, FilterLatestAdded, FilterRetrievability:
		default:
			return userErr("unknownFilteredOrder", "Unknown filtered deck order %q.", t.Order)
		}
	}
	for _, d := range cfg.PreviewDelays {
		if d < 0 || d > MaxPreviewDelay {
			return userErr("previewDelayRange", "Preview delays must be between 0 and %d seconds.", MaxPreviewDelay)
		}
	}
	return nil
}

func cloneFilter(cfg FilterConfig) *FilterConfig {
	cfg.Terms = slices.Clone(cfg.Terms)
	for i := range cfg.Terms {
		cfg.Terms[i].Search = strings.TrimSpace(cfg.Terms[i].Search)
	}
	return &cfg
}

// CreateFilteredDeck creates a filtered deck and builds it, in one step. It
// returns the new deck's ID and the number of cards it holds. Nothing
// changes when no card matches.
func (c *Collection) CreateFilteredDeck(name string, cfg FilterConfig) (ID, int, error) {
	name, err := normalizeDeckName(name)
	if err != nil {
		return 0, 0, err
	}
	if err := ValidateFilterConfig(cfg); err != nil {
		return 0, 0, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if d := c.deckByName(name); d != nil {
		return 0, 0, userErr("deckNameTaken", "A deck named %q already exists.", d.Name)
	}
	if err := c.checkParents(name); err != nil {
		return 0, 0, err
	}
	resolved, parents := c.resolvePath(name, nil)
	d := &Deck{ID: c.ids.next(), Name: resolved, PresetID: DefaultPresetID, Filter: cloneFilter(cfg)}
	o := &op{Kind: "buildFiltered", Decks: append(parents, d)}
	n, err := c.buildInto(o, d)
	if err != nil {
		return 0, 0, err
	}
	return d.ID, n, c.commit(o)
}

// UpdateFilteredDeck renames a filtered deck, changes its options and
// rebuilds it, in one step. It returns the number of cards the deck holds.
// Nothing changes when no card matches.
func (c *Collection) UpdateFilteredDeck(id ID, name string, cfg FilterConfig) (int, error) {
	name, err := normalizeDeckName(name)
	if err != nil {
		return 0, err
	}
	if err := ValidateFilterConfig(cfg); err != nil {
		return 0, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.filteredDeck(id)
	if err != nil {
		return 0, err
	}
	nd := *d
	nd.Filter = cloneFilter(cfg)
	o := &op{Kind: "buildFiltered"}
	if name != d.Name {
		if other := c.deckByName(name); other != nil && other.ID != id {
			return 0, userErr("deckNameTaken", "A deck named %q already exists.", other.Name)
		}
		if err := c.checkParents(name); err != nil {
			return 0, err
		}
		resolved, parents := c.resolvePath(name, map[ID]bool{id: true})
		nd.Name = resolved
		o.Decks = parents
	}
	o.Decks = append(o.Decks, &nd)
	n, err := c.buildInto(o, &nd)
	if err != nil {
		return 0, err
	}
	return n, c.commit(o)
}

// RebuildFilteredDeck returns a filtered deck's cards home and builds it
// again from its searches, in one step. It returns the number of cards it
// holds. Nothing changes when no card matches.
func (c *Collection) RebuildFilteredDeck(id ID) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, err := c.filteredDeck(id)
	if err != nil {
		return 0, err
	}
	o := &op{Kind: "buildFiltered"}
	n, err := c.buildInto(o, d)
	if err != nil {
		return 0, err
	}
	return n, c.commit(o)
}

// EmptyFilteredDeck returns a filtered deck's cards to their home decks.
// Their scheduling stays as it is.
func (c *Collection) EmptyFilteredDeck(id ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.filteredDeck(id); err != nil {
		return err
	}
	o := &op{Kind: "emptyFiltered"}
	for _, card := range c.filteredCards(id) {
		cp := *card
		cp.returnHome()
		o.Cards = append(o.Cards, &cp)
	}
	return c.commit(o)
}

// filteredDeck returns filtered deck id. Callers hold a lock.
func (c *Collection) filteredDeck(id ID) (*Deck, error) {
	d, ok := c.decks[id]
	if !ok {
		return nil, userErr("deckGone", "That deck no longer exists.")
	}
	if d.Filter == nil {
		return nil, userErr("notFilteredDeck", "%q is not a filtered deck.", d.Name)
	}
	return d, nil
}

// filteredCards returns the cards in filtered deck id, sorted by ID.
// Callers hold a lock.
func (c *Collection) filteredCards(id ID) []*Card {
	var out []*Card
	for _, card := range c.cards {
		if card.DeckID == id {
			out = append(out, card)
		}
	}
	slices.SortFunc(out, func(a, b *Card) int { return cmp.Compare(a.ID, b.ID) })
	return out
}

// errNoMatch is the error for building a filtered deck no card matches.
func errNoMatch() error {
	return userErr("filteredNoMatch", "No cards matched the search. Suspended and buried cards and cards in other filtered decks are left out.")
}

// buildInto adds building filtered deck d (as it will be after o) to o:
// the cards now in it return home, then each search moves up to its limit
// of matching cards into it, in its order. Suspended and buried cards and
// cards in other filtered decks are left out. It returns the number of
// cards the deck then holds, or an error when that is none. Callers hold
// the write lock.
func (c *Collection) buildInto(o *op, d *Deck) (int, error) {
	now := c.now()
	pending := map[ID]*Card{}
	for _, card := range c.filteredCards(d.ID) {
		cp := *card
		cp.returnHome()
		pending[card.ID] = &cp
	}
	chosen := map[ID]bool{}
	pos := 0
	for _, t := range d.Filter.Terms {
		ids, err := c.matchLocked(t.Search, now)
		if err != nil {
			return 0, err
		}
		var cards []*Card
		for _, id := range ids {
			card := c.cards[id]
			if chosen[id] || card.Suspended || isBuried(card, now) || card.HomeDeckID != 0 && card.DeckID != d.ID {
				continue
			}
			cards = append(cards, card)
		}
		slices.SortFunc(cards, filterOrder(t.Order, now))
		for _, card := range cards[:min(len(cards), t.Limit)] {
			chosen[card.ID] = true
			cp, ok := pending[card.ID]
			if !ok {
				v := *card
				cp = &v
				pending[card.ID] = cp
			}
			pos++
			cp.HomeDeckID, cp.DeckID, cp.FilteredPos = cp.DeckID, d.ID, pos
		}
	}
	if pos == 0 {
		return 0, errNoMatch()
	}
	ids := make([]ID, 0, len(pending))
	for id := range pending {
		ids = append(ids, id)
	}
	sortIDs(ids)
	for _, id := range ids {
		if cur := c.cards[id]; *cur != *pending[id] {
			o.Cards = append(o.Cards, pending[id])
		}
	}
	return pos, nil
}

// filterOrder is the order in which a filtered deck's search picks its
// cards, and they are studied.
func filterOrder(order string, now time.Time) func(a, b *Card) int {
	byID := func(a, b *Card) int { return cmp.Compare(a.ID, b.ID) }
	switch order {
	case FilterOldestSeen:
		return func(a, b *Card) int { return cmp.Or(a.LastReview.Compare(b.LastReview), byID(a, b)) }
	case FilterIntervalsAscending:
		return func(a, b *Card) int { return cmp.Or(cmp.Compare(a.ScheduledDays, b.ScheduledDays), byID(a, b)) }
	case FilterIntervalsDescending:
		return func(a, b *Card) int { return cmp.Or(cmp.Compare(b.ScheduledDays, a.ScheduledDays), byID(a, b)) }
	case FilterMostLapses:
		return func(a, b *Card) int { return cmp.Or(cmp.Compare(b.Lapses, a.Lapses), byID(a, b)) }
	case FilterAdded:
		return func(a, b *Card) int {
			return cmp.Or(cmp.Compare(a.NoteID, b.NoteID), cmp.Compare(a.Ord, b.Ord), byID(a, b))
		}
	case FilterLatestAdded:
		return func(a, b *Card) int {
			return cmp.Or(cmp.Compare(b.NoteID, a.NoteID), cmp.Compare(a.Ord, b.Ord), byID(a, b))
		}
	case FilterDue:
		// As Anki: by card type, new cards by position, others by due time.
		return func(a, b *Card) int {
			if c := cmp.Compare(a.State, b.State); c != 0 {
				return c
			}
			if a.State == scheduler.New {
				return positionOrder(a, b)
			}
			return dueOrder(a, b)
		}
	case FilterRetrievability:
		r := func(card *Card) float64 {
			if v, ok := scheduler.Retrievability(card.CardState, now); ok && card.State != scheduler.New {
				return v
			}
			return math.NaN()
		}
		return func(a, b *Card) int { return cmp.Or(knownFirst(r(a), r(b), false), byID(a, b)) }
	}
	// FilterRandom: a new shuffle every build.
	keys := map[ID]uint64{}
	key := func(id ID) uint64 {
		k, ok := keys[id]
		if !ok {
			k = rand.Uint64()
			keys[id] = k
		}
		return k
	}
	return func(a, b *Card) int { return cmp.Or(cmp.Compare(key(a.ID), key(b.ID)), byID(a, b)) }
}

// CustomStudyDeckName is the filtered deck custom study creates, replacing
// the previous session.
const CustomStudyDeckName = "Custom Study Session"

// Custom study choices.
const (
	// CustomForgotten reviews the cards rated Again in the last Days days.
	CustomForgotten = "forgotten"
	// CustomAhead reviews the cards due in the next Days days.
	CustomAhead = "ahead"
	// CustomPreviewNew previews the new cards added in the last Days days.
	CustomPreviewNew = "previewNew"
	// CustomByState studies up to Cards cards of a State, optionally only
	// those with one of Tags.
	CustomByState = "byState"
)

// Card states of CustomByState.
const (
	CustomNew = "new"
	CustomDue = "due"
	CustomAll = "all"
)

// Custom study limits.
const (
	MaxCustomDays  = 365
	MaxCustomCards = 9999
)

// CustomStudy is a custom study choice.
type CustomStudy struct {
	Kind  string
	Days  int
	State string
	Tags  []string
	Cards int
}

// quoteSearch quotes text as one search term value.
func quoteSearch(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// customStudyFilter returns the filtered deck options of a custom study
// choice for the deck named deck (design decision 4).
func customStudyFilter(deck string, cs CustomStudy) (FilterConfig, error) {
	days := func() error {
		if cs.Days < 1 || cs.Days > MaxCustomDays {
			return userErr("dayCountRange", "The number of days must be between 1 and %d.", MaxCustomDays)
		}
		return nil
	}
	cfg := DefaultFilterConfig("")
	t := &cfg.Terms[0]
	t.Limit = MaxFilterLimit
	in := "deck:" + quoteSearch(deck)
	switch cs.Kind {
	case CustomForgotten:
		if err := days(); err != nil {
			return cfg, err
		}
		t.Search, t.Order = fmt.Sprintf("%s rated:%d:1", in, cs.Days), FilterRandom
	case CustomAhead:
		if err := days(); err != nil {
			return cfg, err
		}
		t.Search, t.Order = fmt.Sprintf("%s prop:due<=%d prop:due>0", in, cs.Days), FilterDue
	case CustomPreviewNew:
		if err := days(); err != nil {
			return cfg, err
		}
		t.Search, t.Order = fmt.Sprintf("%s is:new added:%d", in, cs.Days), FilterAdded
		cfg.Reschedule = false
	case CustomByState:
		if cs.Cards < 1 || cs.Cards > MaxCustomCards {
			return cfg, userErr("cardCountRange", "The number of cards must be between 1 and %d.", MaxCustomCards)
		}
		t.Limit = cs.Cards
		switch cs.State {
		case CustomNew:
			t.Search, t.Order = in+" is:new", FilterAdded
		case CustomDue:
			t.Search, t.Order = in+" is:due", FilterDue
		case CustomAll:
			// As in Anki, all cards are studied without rescheduling.
			t.Search, t.Order = in, FilterRandom
			cfg.Reschedule = false
		default:
			return cfg, userErr("unknownCardState", "Unknown card state %q.", cs.State)
		}
		var tags []string
		for _, tag := range NormalizeTags(cs.Tags) {
			tags = append(tags, "tag:"+quoteSearch(tag))
		}
		if len(tags) > 0 {
			t.Search += " (" + strings.Join(tags, " or ") + ")"
		}
	default:
		return cfg, userErr("unknownCustomStudy", "Unknown custom study choice %q.", cs.Kind)
	}
	return cfg, nil
}

// CustomStudy creates the filtered deck "Custom Study Session" for a
// custom study choice on deck and its subdecks, replacing the previous
// session: its cards return home and it is built again. It returns the
// session's deck and number of cards. Nothing changes when no card
// matches.
func (c *Collection) CustomStudy(deck ID, cs CustomStudy) (ID, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deck]
	if !ok {
		return 0, 0, userErr("deckGone", "That deck no longer exists.")
	}
	if d.Filter != nil {
		return 0, 0, userErr("customStudyNeedsRegular", "Custom study needs a regular deck; %q is a filtered deck.", d.Name)
	}
	cfg, err := customStudyFilter(d.Name, cs)
	if err != nil {
		return 0, 0, err
	}
	o := &op{Kind: "customStudy"}
	session := c.deckByName(CustomStudyDeckName)
	switch {
	case session == nil:
		session = &Deck{ID: c.ids.next(), Name: CustomStudyDeckName, PresetID: DefaultPresetID}
	case session.Filter == nil:
		return 0, 0, userErr("customStudyNameTaken", "A regular deck is named %q. Rename it to use custom study.", session.Name)
	default:
		cp := *session
		session = &cp
	}
	session.Filter = cloneFilter(cfg)
	o.Decks = append(o.Decks, session)
	n, err := c.buildInto(o, session)
	if err != nil {
		return 0, 0, err
	}
	return session.ID, n, c.commit(o)
}
