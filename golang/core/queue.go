package core

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"time"

	"cardflip/golang/scheduler"
)

// Counts are the due cards of a deck, including its subdecks.
type Counts struct {
	New      int `json:"new"`
	Learning int `json:"learning"`
	Review   int `json:"review"`
}

// Total is the number of due cards.
func (c Counts) Total() int { return c.New + c.Learning + c.Review }

// Queue is what can be studied in a deck right now.
type Queue struct {
	// Learning cards due by now, earliest first. Daily limits do not apply.
	Learning []ID
	// Review cards and interday learning cards due today, within the review
	// limits (interday learning cards take it first). Interday learning
	// cards are earliest first, review cards in the deck's review sort
	// order; they are placed per the deck's interday learning/review order.
	Review []ID
	// InterdayLearning is how many cards of Review are interday learning
	// cards; they count as learning.
	InterdayLearning int
	// New cards within the new card limits, chosen by the deck's new card
	// gather order and shown in its new card sort order.
	New []ID
	// NewReviewOrder places New relative to Review: NewMix,
	// NewBeforeReviews or NewAfterReviews.
	NewReviewOrder string
	// LearningAhead are the learning cards due after now but within the
	// learn ahead limit, earliest first. They are shown only when nothing
	// else is due, and not counted.
	LearningAhead []ID
	// LearningLater is when the next learning card beyond the learn ahead
	// limit becomes due today (zero if none).
	LearningLater time.Time
}

// Counts returns the queue's sizes.
func (q Queue) Counts() Counts {
	return Counts{New: len(q.New), Learning: len(q.Learning) + q.InterdayLearning, Review: len(q.Review) - q.InterdayLearning}
}

// dayView is everything about the current study day the queue needs.
type dayView struct {
	now      time.Time
	day      int // scheduler.DayNumber
	dayEnd   time.Time
	children map[ID][]ID
	parent   map[ID]ID
	// deckRank orders decks as the deck list does: parents before their
	// children, siblings by name.
	deckRank map[ID]int
	// reviewKeys holds the sort keys of the cards in ownReview.
	reviewKeys map[ID]reviewKey
	// own lists the eligible cards directly in each deck; ownReview
	// includes interday learning cards.
	ownNew, ownReview, ownLearning map[ID][]*Card
	// laterLearning lists the learning cards directly in each deck that
	// become due later today.
	laterLearning map[ID][]*Card
	// aheadUntil is the end of the learn ahead limit.
	aheadUntil time.Time
	// doneNew and doneReview count today's answers per deck (own cards);
	// doneReview includes interday learning cards.
	doneNew, doneReview map[ID]int
	// buried counts the cards directly in each deck that are buried:
	// manually, or held back today by sibling burying.
	buried map[ID]int
	// filtered lists the unsuspended, unburied cards of each filtered deck;
	// they are in no other list (see filteredQueue).
	filtered map[ID][]*Card
}

// view collects the study day's state. Callers hold a lock.
func (c *Collection) view(now time.Time) *dayView {
	v := &dayView{
		now:           now,
		day:           scheduler.DayNumber(now, c.settings.RolloverHour),
		dayEnd:        scheduler.NextDayStart(now, c.settings.RolloverHour),
		children:      map[ID][]ID{},
		parent:        map[ID]ID{},
		ownNew:        map[ID][]*Card{},
		ownReview:     map[ID][]*Card{},
		ownLearning:   map[ID][]*Card{},
		laterLearning: map[ID][]*Card{},
		aheadUntil:    now.Add(c.settings.LearnAhead()),
		doneNew:       map[ID]int{},
		doneReview:    map[ID]int{},
		buried:        map[ID]int{},
		deckRank:      make(map[ID]int, len(c.decks)),
		reviewKeys:    map[ID]reviewKey{},
		filtered:      map[ID][]*Card{},
	}
	byName := make(map[string]ID, len(c.decks))
	decks := make([]Deck, 0, len(c.decks))
	for _, d := range c.decks {
		byName[strings.ToLower(d.Name)] = d.ID
		decks = append(decks, Deck{ID: d.ID, Name: d.Name})
	}
	sortDecks(decks)
	for i, d := range decks {
		v.deckRank[d.ID] = i
	}
	for _, d := range c.decks {
		if p, ok := byName[strings.ToLower(parentName(d.Name))]; ok {
			v.children[p] = append(v.children[p], d.ID)
			v.parent[d.ID] = p
		}
	}
	// Notes studied today: their other new and review cards wait for the
	// next day when their deck buries them (sibling burying).
	entries := c.todayEntries(now)
	studied := c.studied(entries)
	for _, e := range entries {
		card, ok := c.cards[e.CardID]
		if !ok || e.Rating == 0 || e.Preview { // manual entries and previews are no answers
			continue
		}
		switch {
		case e.StateBefore == scheduler.New:
			v.doneNew[card.DeckID]++
		case e.StateBefore == scheduler.Review || e.Before.Interday:
			v.doneReview[card.DeckID]++
		}
	}
	// siblingBuried reports whether a card that would be due is held back
	// by sibling burying, and counts it as buried.
	siblingBuried := func(card *Card) bool {
		if c.siblingBuried(card, studied) {
			v.buried[card.DeckID]++
			return true
		}
		return false
	}
	for _, card := range c.cards {
		if card.Suspended {
			continue
		}
		if isBuried(card, now) {
			v.buried[card.DeckID]++
			continue
		}
		if card.HomeDeckID != 0 {
			v.filtered[card.DeckID] = append(v.filtered[card.DeckID], card)
			continue
		}
		switch card.State {
		case scheduler.Learning, scheduler.Relearning:
			if card.Interday {
				// Due by day, like reviews, and limited with them.
				if card.Due.Before(v.dayEnd) && !siblingBuried(card) {
					v.ownReview[card.DeckID] = append(v.ownReview[card.DeckID], card)
				}
			} else if !card.Due.After(now) {
				v.ownLearning[card.DeckID] = append(v.ownLearning[card.DeckID], card)
			} else if card.Due.Before(v.dayEnd) {
				v.laterLearning[card.DeckID] = append(v.laterLearning[card.DeckID], card)
			}
		case scheduler.Review:
			if card.Due.Before(v.dayEnd) && !siblingBuried(card) {
				v.ownReview[card.DeckID] = append(v.ownReview[card.DeckID], card)
				v.reviewKeys[card.ID] = v.reviewKeyOf(card, c.settings.RolloverHour)
			}
		case scheduler.New:
			if !siblingBuried(card) {
				v.ownNew[card.DeckID] = append(v.ownNew[card.DeckID], card)
			}
		}
	}
	return v
}

// studied maps each note with answers among entries to the cards
// answered, each with the time of its last answer. Callers hold a lock.
func (c *Collection) studied(entries []*ReviewEntry) map[ID]map[ID]time.Time {
	studied := map[ID]map[ID]time.Time{}
	for _, e := range entries {
		card, ok := c.cards[e.CardID]
		if !ok || e.Rating == 0 || e.Preview { // manual entries and previews are no answers
			continue
		}
		if studied[card.NoteID] == nil {
			studied[card.NoteID] = map[ID]time.Time{}
		}
		if e.Time.After(studied[card.NoteID][card.ID]) {
			studied[card.NoteID][card.ID] = e.Time
		}
	}
	return studied
}

// siblingBuried reports whether the card's deck buries cards of its kind
// and another card of its note was answered (see studied) after the deck
// was last unburied. Callers hold a lock.
func (c *Collection) siblingBuried(card *Card, studied map[ID]map[ID]time.Time) bool {
	o := c.deckOptions(card)
	var bury bool
	switch card.State {
	case scheduler.New:
		bury = o.BuryNew
	case scheduler.Review:
		bury = o.BuryReview
	default:
		bury = o.BuryInterdayLearning
	}
	if !bury {
		return false
	}
	unburied := c.decks[card.DeckID].UnburiedAt
	for cid, at := range studied[card.NoteID] {
		if cid != card.ID && at.After(unburied) {
			return true
		}
	}
	return false
}

// subtreeDone sums per-deck numbers (today's answers, buried cards) over a
// deck and its subdecks.
func (v *dayView) subtreeDone(id ID, done map[ID]int) int {
	n := done[id]
	for _, ch := range v.children[id] {
		n += v.subtreeDone(ch, done)
	}
	return n
}

// limits describes one kind of daily limit (new or review) of a queue.
type limits struct {
	own   map[ID][]*Card
	done  map[ID]int
	limit func(d *Deck) int
	// order is the order in which cards fill the limits.
	order func(a, b *Card) int
	// room optionally caps a deck's cards further; ok is false for no cap.
	room func(id ID) (n int, ok bool)
	// shown, when set, records how many cards each deck of the tree shows.
	shown map[ID]int
}

// limited returns the cards of a deck tree for one queue: the deck's own
// cards plus each child's (already limited) cards, in order, cut to the
// deck's remaining daily limit.
func (c *Collection) limited(v *dayView, id ID, l *limits) []*Card {
	cards := append([]*Card(nil), l.own[id]...)
	for _, ch := range v.children[id] {
		cards = append(cards, c.limited(v, ch, l)...)
	}
	slices.SortFunc(cards, l.order)
	remaining := l.limit(c.decks[id]) - v.subtreeDone(id, l.done)
	if l.room != nil {
		if n, ok := l.room(id); ok {
			remaining = min(remaining, n)
		}
	}
	remaining = max(remaining, 0)
	if len(cards) > remaining {
		cards = cards[:remaining]
	}
	if l.shown != nil {
		l.shown[id] = len(cards)
	}
	return cards
}

// Salts keep the random orders of gathering, sorting and reviews apart.
const (
	saltGather = iota + 1
	saltSort
	saltReview
)

// dayRandom is a pseudo-random number for id that stays the same within
// study day day, so random orders do not reshuffle after every answer.
func dayRandom(id ID, day int, salt uint64) uint64 {
	x := uint64(id) ^ uint64(day)*0x9e3779b97f4a7c15 ^ salt<<56
	// splitmix64 finalizer
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	return x ^ x>>31
}

// newGather is the order in which new cards fill the daily limits, for
// the gather order of the studied deck.
func (v *dayView) newGather(order string) func(a, b *Card) int {
	rnd := func(id ID) uint64 { return dayRandom(id, v.day, saltGather) }
	switch order {
	case GatherDeck:
		return func(a, b *Card) int {
			return cmp.Or(cmp.Compare(v.deckRank[a.DeckID], v.deckRank[b.DeckID]), positionOrder(a, b))
		}
	case GatherDeckRandomNotes:
		return func(a, b *Card) int {
			return cmp.Or(cmp.Compare(v.deckRank[a.DeckID], v.deckRank[b.DeckID]),
				cmp.Compare(rnd(a.NoteID), rnd(b.NoteID)), noteTypeOrder(a, b))
		}
	case GatherDescending:
		return func(a, b *Card) int { return positionOrder(b, a) }
	case GatherRandomNotes:
		return func(a, b *Card) int { return cmp.Or(cmp.Compare(rnd(a.NoteID), rnd(b.NoteID)), noteTypeOrder(a, b)) }
	case GatherRandomCards:
		return func(a, b *Card) int { return cmp.Or(cmp.Compare(rnd(a.ID), rnd(b.ID)), cmp.Compare(a.ID, b.ID)) }
	}
	return positionOrder
}

// sortNew puts gathered new cards in the order they are shown, for the
// sort order of the studied deck.
func (v *dayView) sortNew(cards []*Card, order string) {
	rnd := func(id ID) uint64 { return dayRandom(id, v.day, saltSort) }
	switch order {
	case SortTypeThenGathered:
		slices.SortStableFunc(cards, func(a, b *Card) int { return cmp.Compare(a.Ord, b.Ord) })
	case SortTypeThenRandom:
		slices.SortFunc(cards, func(a, b *Card) int {
			return cmp.Or(cmp.Compare(a.Ord, b.Ord), cmp.Compare(rnd(a.ID), rnd(b.ID)), cmp.Compare(a.ID, b.ID))
		})
	case SortRandomNoteThenType:
		slices.SortFunc(cards, func(a, b *Card) int { return cmp.Or(cmp.Compare(rnd(a.NoteID), rnd(b.NoteID)), noteTypeOrder(a, b)) })
	case SortRandom:
		slices.SortFunc(cards, func(a, b *Card) int { return cmp.Or(cmp.Compare(rnd(a.ID), rnd(b.ID)), cmp.Compare(a.ID, b.ID)) })
	}
}

// positionOrder is the order cards were added in.
func positionOrder(a, b *Card) int {
	return cmp.Or(cmp.Compare(a.NewPos, b.NewPos), cmp.Compare(a.ID, b.ID))
}

// noteTypeOrder keeps the cards of a note together, by card type.
func noteTypeOrder(a, b *Card) int {
	return cmp.Or(cmp.Compare(a.NoteID, b.NoteID), cmp.Compare(a.Ord, b.Ord), cmp.Compare(a.ID, b.ID))
}

// reviewKey holds what review sort orders compare, computed once per view.
type reviewKey struct {
	// dueDay is the study day the card is due on.
	dueDay int
	// difficulty and retrievability are NaN without a memory state.
	difficulty, retrievability float64
	// overdueness is the days since the last review divided by the
	// interval.
	overdueness float64
	random      uint64
}

func (v *dayView) reviewKeyOf(card *Card, rolloverHour int) reviewKey {
	k := reviewKey{
		dueDay: scheduler.DayNumber(card.Due, rolloverHour), difficulty: math.NaN(), retrievability: math.NaN(),
		random: dayRandom(card.ID, v.day, saltReview),
	}
	if card.Memory != nil {
		k.difficulty = card.Memory.Difficulty
	}
	if r, ok := scheduler.Retrievability(card.CardState, v.now); ok {
		k.retrievability = r
	}
	ivl := float64(max(card.ScheduledDays, 1))
	last := card.LastReview
	if last.IsZero() {
		last = card.Due.AddDate(0, 0, -card.ScheduledDays)
	}
	k.overdueness = v.now.Sub(last).Hours() / 24 / ivl
	return k
}

// reviewOrder is the order in which due cards fill the review limits:
// interday learning cards first, earliest first, then review cards in the
// review sort order of the studied deck.
func (v *dayView) reviewOrder(order string) func(a, b *Card) int {
	return func(a, b *Card) int {
		if a.Interday != b.Interday {
			if a.Interday {
				return -1
			}
			return 1
		}
		if a.Interday {
			return dueOrder(a, b)
		}
		ka, kb := v.reviewKeys[a.ID], v.reviewKeys[b.ID]
		due := cmp.Compare(ka.dueDay, kb.dueDay)
		var c int
		switch order {
		case ReviewDueDeck:
			c = cmp.Or(due, cmp.Compare(v.deckRank[a.DeckID], v.deckRank[b.DeckID]))
		case ReviewDeckDue:
			c = cmp.Or(cmp.Compare(v.deckRank[a.DeckID], v.deckRank[b.DeckID]), due)
		case ReviewIntervalsAscending:
			c = cmp.Compare(a.ScheduledDays, b.ScheduledDays)
		case ReviewIntervalsDescending:
			c = cmp.Compare(b.ScheduledDays, a.ScheduledDays)
		case ReviewDifficultyAscending:
			c = knownFirst(ka.difficulty, kb.difficulty, false)
		case ReviewDifficultyDescending:
			c = knownFirst(ka.difficulty, kb.difficulty, true)
		case ReviewRetrievabilityAscending:
			c = knownFirst(ka.retrievability, kb.retrievability, false)
		case ReviewRelativeOverdueness:
			c = cmp.Compare(kb.overdueness, ka.overdueness)
		default: // ReviewDueRandom
			c = due
		}
		return cmp.Or(c, cmp.Compare(ka.random, kb.random), cmp.Compare(a.ID, b.ID))
	}
}

// knownFirst compares x and y ascending or descending, with missing
// values (NaN) last either way.
func knownFirst(x, y float64, descending bool) int {
	switch xn, yn := math.IsNaN(x), math.IsNaN(y); {
	case xn && yn:
		return 0
	case xn:
		return 1
	case yn:
		return -1
	case descending:
		return cmp.Compare(y, x)
	}
	return cmp.Compare(x, y)
}

func dueOrder(a, b *Card) int {
	return cmp.Or(a.Due.Compare(b.Due), cmp.Compare(a.ID, b.ID))
}

func (c *Collection) queue(v *dayView, id ID) Queue { return c.buildQueue(v, id, false, false) }

// buildQueue builds the queue of deck id. liftNew and liftReview ignore
// that deck's own new or review limit, to find the cards a custom study
// increase could add.
func (c *Collection) buildQueue(v *dayView, id ID, liftNew, liftReview bool) Queue {
	if d := c.decks[id]; d.Filter != nil {
		return c.filteredQueue(v, d)
	}
	newLimit := func(d *Deck) int {
		if liftNew && d.ID == id {
			return math.MaxInt32
		}
		return c.newLimit(d, v.day)
	}
	reviewLimit := func(d *Deck) int {
		if liftReview && d.ID == id {
			return math.MaxInt32
		}
		return c.reviewLimit(d, v.day)
	}
	// With "limits start from top", the studied deck's cards are also cut
	// to what its parents' limits leave.
	var parents []*Deck
	opts := c.optionsOf(id)
	if opts.LimitsStartFromTop {
		for p := v.parent[id]; p != 0; p = v.parent[p] {
			parents = append(parents, c.decks[p])
		}
	}
	fromTop := func(limit func(*Deck) int, done map[ID]int) func(ID) (int, bool) {
		return func(d ID) (int, bool) {
			if d != id || len(parents) == 0 {
				return 0, false
			}
			n := math.MaxInt
			for _, p := range parents {
				n = min(n, limit(p)-v.subtreeDone(p.ID, done))
			}
			return n, true
		}
	}

	q := Queue{NewReviewOrder: opts.NewReviewOrder}
	var interday, reviews []ID
	shownReviews := map[ID]int{}
	for _, card := range c.limited(v, id, &limits{own: v.ownReview, done: v.doneReview, limit: reviewLimit,
		order: v.reviewOrder(opts.ReviewSortOrder), room: fromTop(reviewLimit, v.doneReview), shown: shownReviews}) {
		if card.Interday {
			interday = append(interday, card.ID)
		} else {
			reviews = append(reviews, card.ID)
		}
	}
	q.InterdayLearning = len(interday)
	switch opts.InterdayLearningOrder {
	case InterdayBefore:
		q.Review = append(interday, reviews...)
	case InterdayAfter:
		q.Review = append(reviews, interday...)
	default:
		q.Review = mix(interday, reviews)
	}
	top := fromTop(newLimit, v.doneNew)
	// Unless new cards ignore the review limit, they only fill the room
	// the review limit leaves after today's answers and the due reviews.
	room := func(d ID) (int, bool) {
		n, ok := top(d)
		if dk := c.decks[d]; !c.optionsOf(d).NewIgnoresReviewLimit {
			r := reviewLimit(dk) - v.subtreeDone(d, v.doneReview) - v.subtreeDone(d, v.doneNew) - shownReviews[d]
			if !ok || r < n {
				n, ok = r, true
			}
		}
		return n, ok
	}
	newCards := c.limited(v, id, &limits{own: v.ownNew, done: v.doneNew, limit: newLimit,
		order: v.newGather(opts.NewGatherOrder), room: room})
	v.sortNew(newCards, opts.NewSortOrder)
	for _, card := range newCards {
		q.New = append(q.New, card.ID)
	}
	var learning, ahead []*Card
	var walk func(ID)
	walk = func(d ID) {
		learning = append(learning, v.ownLearning[d]...)
		for _, card := range v.laterLearning[d] {
			if !card.Due.After(v.aheadUntil) {
				ahead = append(ahead, card)
			} else if q.LearningLater.IsZero() || card.Due.Before(q.LearningLater) {
				q.LearningLater = card.Due
			}
		}
		for _, ch := range v.children[d] {
			walk(ch)
		}
	}
	walk(id)
	slices.SortFunc(learning, dueOrder)
	for _, card := range learning {
		q.Learning = append(q.Learning, card.ID)
	}
	slices.SortFunc(ahead, dueOrder)
	for _, card := range ahead {
		q.LearningAhead = append(q.LearningAhead, card.ID)
	}
	return q
}

// filteredQueue builds the queue of filtered deck d: all its cards, without
// daily limits. New cards are new and the others reviews, each in the
// deck's order, also when they are not due yet (review ahead); learning
// cards due later today wait as in other decks. In preview mode a card
// answered Again or Hard waits for its preview delay like a learning card,
// and the others are reviews whatever their state.
func (c *Collection) filteredQueue(v *dayView, d *Deck) Queue {
	q := Queue{NewReviewOrder: NewMix}
	preview := !d.Filter.Reschedule
	var news, reviews []*Card
	type waiting struct {
		card *Card
		due  time.Time
	}
	var learning, ahead []waiting
	wait := func(card *Card, due time.Time) {
		switch {
		case !due.After(v.now):
			learning = append(learning, waiting{card, due})
		case !due.After(v.aheadUntil):
			ahead = append(ahead, waiting{card, due})
		case due.Before(v.dayEnd) && (q.LearningLater.IsZero() || due.Before(q.LearningLater)):
			q.LearningLater = due
		}
	}
	for _, card := range v.filtered[d.ID] {
		switch {
		case preview && !card.PreviewDue.IsZero():
			wait(card, card.PreviewDue)
		case card.State == scheduler.New:
			news = append(news, card)
		case !preview && (card.State == scheduler.Learning || card.State == scheduler.Relearning) && !card.Interday:
			wait(card, card.Due)
		default:
			reviews = append(reviews, card)
			if !preview && card.Interday {
				q.InterdayLearning++
			}
		}
	}
	byPos := func(a, b *Card) int {
		return cmp.Or(cmp.Compare(a.FilteredPos, b.FilteredPos), cmp.Compare(a.ID, b.ID))
	}
	byDue := func(a, b waiting) int { return cmp.Or(a.due.Compare(b.due), cmp.Compare(a.card.ID, b.card.ID)) }
	slices.SortFunc(news, byPos)
	slices.SortFunc(reviews, byPos)
	slices.SortFunc(learning, byDue)
	slices.SortFunc(ahead, byDue)
	for _, card := range news {
		q.New = append(q.New, card.ID)
	}
	for _, card := range reviews {
		q.Review = append(q.Review, card.ID)
	}
	for _, w := range learning {
		q.Learning = append(q.Learning, w.card.ID)
	}
	for _, w := range ahead {
		q.LearningAhead = append(q.LearningAhead, w.card.ID)
	}
	return q
}

// mix spreads the cards of a evenly among those of b, starting with a.
func mix(a, b []ID) []ID {
	out := make([]ID, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		// Take from a while it is not ahead of its share.
		if i < len(a) && (j == len(b) || i*len(b) <= j*len(a)) {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	return out
}

// Queue returns what can be studied in a deck and its subdecks right now.
func (c *Collection) Queue(deckID ID) (Queue, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.decks[deckID]; !ok {
		return Queue{}, userErr("deckGone", "That deck no longer exists.")
	}
	return c.queue(c.view(c.now()), deckID), nil
}

// Extra are the cards a custom study increase of today's limits could add
// to a deck's queue.
type Extra struct {
	New    int `json:"new"`
	Review int `json:"review"`
}

// ExtraCards counts the new and review cards beyond the deck's own limits
// today: those that raising its new or review limit would add. Its
// subdecks' limits and, when they start from the top, its parents' still
// apply.
func (c *Collection) ExtraCards(deckID ID) (Extra, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.decks[deckID]; !ok {
		return Extra{}, userErr("deckGone", "That deck no longer exists.")
	}
	v := c.view(c.now())
	q := c.queue(v, deckID)
	return Extra{
		New:    len(c.buildQueue(v, deckID, true, false).New) - len(q.New),
		Review: len(c.buildQueue(v, deckID, false, true).Review) - len(q.Review),
	}, nil
}

// Overview is what a deck's overview shows about today, for the deck and
// its subdecks.
type Overview struct {
	Counts Counts
	// Buried counts the unsuspended cards that are buried: manually, or
	// held back today by sibling burying while they would be due.
	Buried int
}

// Overview returns a deck's due counts and buried cards.
func (c *Collection) Overview(deckID ID) (Overview, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.decks[deckID]; !ok {
		return Overview{}, userErr("deckGone", "That deck no longer exists.")
	}
	v := c.view(c.now())
	return Overview{Counts: c.queue(v, deckID).Counts(), Buried: v.subtreeDone(deckID, v.buried)}, nil
}

// DeckCounts returns the due counts of every deck.
func (c *Collection) DeckCounts() map[ID]Counts {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v := c.view(c.now())
	out := make(map[ID]Counts, len(c.decks))
	for id := range c.decks {
		out[id] = c.queue(v, id).Counts()
	}
	return out
}

// NextCountChange returns when the due counts next change on their own,
// without a write: the earliest due time of an unsuspended, unburied
// learning card after now, or the start of the next study day.
func (c *Collection) NextCountChange() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := c.now()
	next := scheduler.NextDayStart(now, c.settings.RolloverHour)
	for _, card := range c.cards {
		if card.Suspended || isBuried(card, now) {
			continue
		}
		if (card.State == scheduler.Learning || card.State == scheduler.Relearning) &&
			card.Due.After(now) && card.Due.Before(next) {
			next = card.Due
		}
		if card.PreviewDue.After(now) && card.PreviewDue.Before(next) {
			next = card.PreviewDue
		}
	}
	return next
}

// NextDue returns when something in a deck tree can next be studied, after
// now: a learning card's due time, the start of the study day in which a
// review card is due, or the next day when new cards wait for tomorrow's
// limit. Zero means nothing will ever be due (no unsuspended cards).
func (c *Collection) NextDue(deckID ID) time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := c.now()
	d, ok := c.decks[deckID]
	if !ok {
		return time.Time{}
	}
	in := map[ID]bool{}
	for _, t := range c.subtree(d) {
		in[t.ID] = true
	}
	rollover := c.settings.RolloverHour
	tomorrow := scheduler.NextDayStart(now, rollover)
	var next time.Time
	consider := func(t time.Time) {
		if t.After(now) && (next.IsZero() || t.Before(next)) {
			next = t
		}
	}
	for _, card := range c.cards {
		if !in[card.DeckID] || card.Suspended {
			continue
		}
		if !card.PreviewDue.IsZero() {
			if card.PreviewDue.After(now) {
				consider(card.PreviewDue)
			} else {
				consider(tomorrow) // buried
			}
			continue
		}
		switch card.State {
		case scheduler.New:
			consider(tomorrow)
		case scheduler.Learning, scheduler.Relearning:
			if card.Due.After(now) {
				consider(card.Due)
			} else {
				consider(tomorrow) // buried or sibling-buried
			}
		case scheduler.Review:
			start := scheduler.DayStart(card.Due, rollover)
			if start.After(now) {
				consider(start)
			} else {
				consider(tomorrow)
			}
		}
	}
	return next
}

// PickNext chooses the next card of a queue for a session that has already
// shown newShown new and reviewShown review (or interday learning) cards:
// due learning cards first, otherwise new and review cards per the queue's
// new/review order, mixed evenly over the session by default, and when
// nothing else is due the learning card within the learn ahead limit that
// is due soonest.
func PickNext(q Queue, newShown, reviewShown int) (ID, bool) {
	if len(q.Learning) > 0 {
		return q.Learning[0], true
	}
	n, r := len(q.New), len(q.Review)
	switch {
	case n == 0 && r == 0:
		if len(q.LearningAhead) > 0 {
			return q.LearningAhead[0], true
		}
		return 0, false
	case n == 0:
		return q.Review[0], true
	case r == 0:
		return q.New[0], true
	case q.NewReviewOrder == NewBeforeReviews:
		return q.New[0], true
	case q.NewReviewOrder == NewAfterReviews:
		return q.Review[0], true
	}
	// Show a new card when new cards are behind their share of the session.
	totalNew, totalReview := newShown+n, reviewShown+r
	if newShown*totalReview <= reviewShown*totalNew {
		return q.New[0], true
	}
	return q.Review[0], true
}
