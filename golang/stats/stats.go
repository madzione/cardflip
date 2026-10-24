// Package stats computes the statistics screen's sections from a
// collection's cards and review log: for a deck with its subdecks, the
// whole collection or the cards matching a search, over a chosen period.
// Days follow the study day boundary (the rollover hour).
//
// A Loader resolves a scope to a Data, the input every section function
// takes. Decoding the whole review log is the expensive part, so the Loader
// keeps it until the collection changes and the sections, requested one by
// one, share it.
package stats

import (
	"cmp"
	"encoding/json"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/internal/yield"
	"cardflip/golang/scheduler"
)

// Scope kinds.
const (
	ScopeDeck       = "deck"
	ScopeCollection = "collection"
	ScopeSearch     = "search"
)

// Scope is what the statistics cover.
type Scope struct {
	// Kind is ScopeDeck (the deck DeckID with its subdecks),
	// ScopeCollection or ScopeSearch (the cards matching Query).
	Kind   string `json:"kind"`
	DeckID int64  `json:"deckId,omitempty"`
	Query  string `json:"query,omitempty"`
}

// Periods of the time-based sections.
const (
	PeriodMonth   = "month"
	Period3Months = "3months"
	PeriodYear    = "year"
	PeriodAll     = "all"
)

// periodDays is the number of study days of a period other than PeriodAll.
var periodDays = map[string]int{PeriodMonth: 30, Period3Months: 90, PeriodYear: 365}

// ValidPeriod reports whether p is a period.
func ValidPeriod(p string) bool { _, ok := periodDays[p]; return ok || p == PeriodAll }

// youngLimit is the interval, in days, from which a card is mature.
const youngLimit = 21

// Kind is what a review was, for the graphs that split reviews.
type Kind uint8

const (
	// Learning is an answer to a New or Learning card.
	Learning Kind = iota
	// Young and Mature are answers to Review cards with an interval under
	// 21 days and of 21 days or more.
	Young
	Mature
	Relearning
	// Filtered is a review in a filtered deck that did not reschedule the
	// card (an Anki cram review: a passed review card that kept no
	// interval).
	Filtered
)

// Review is a review log answer, reduced to what the statistics use.
// Manual entries (rating 0) are left out when loading.
type Review struct {
	CardID core.ID
	// Day is the study day (scheduler.DayNumber) and Hour the local hour
	// of the answer.
	Day        int
	Hour       int8
	Rating     scheduler.Rating
	Kind       Kind
	IvlBefore  int32
	DurationMs int64
}

// Passed reports whether the answer was not Again.
func (r *Review) Passed() bool { return r.Rating > scheduler.Again }

// mature reports whether a review card with interval ivl is mature.
func mature(ivl int) bool { return ivl >= youngLimit }

// reviewKind classifies a review log entry.
func reviewKind(e *entry) Kind {
	switch e.StateBefore {
	case scheduler.New, scheduler.Learning:
		return Learning
	case scheduler.Relearning:
		return Relearning
	}
	if e.Rating > scheduler.Again && e.IvlAfter == 0 {
		return Filtered
	}
	if mature(e.IvlBefore) {
		return Mature
	}
	return Young
}

// Data is the input of every section: the cards and answers in a scope.
type Data struct {
	Now time.Time
	// Today is the current study day (scheduler.DayNumber).
	Today    int
	Rollover int
	Cards    []core.Card
	// Reviews are the answers to the scope's cards; in the collection scope
	// they include answers to deleted cards.
	Reviews []Review
	// Retention is the desired retention to compare true retention with:
	// the deck's for a deck, otherwise the preset all scoped cards share,
	// or the default preset's.
	Retention float64
}

// DayOf returns the study day of t.
func (d *Data) DayOf(t time.Time) int {
	return scheduler.DayNumber(t.In(d.Now.Location()), d.Rollover)
}

// Date returns the calendar date of study day day, as "2006-01-02".
func Date(day int) string {
	return time.Unix(int64(day)*86400, 0).UTC().Format(time.DateOnly)
}

// firstDay returns the first study day of period, counting back from today;
// for PeriodAll the day of the oldest answer (today without answers).
func (d *Data) firstDay(period string) int {
	if n, ok := periodDays[period]; ok {
		return d.Today - n + 1
	}
	first := d.Today
	for i := range d.Reviews {
		first = min(first, d.Reviews[i].Day)
	}
	return first
}

// Loader resolves scopes to Data. It keeps the decoded review log of the
// collection it last read until the collection changes. It is safe for
// concurrent use; concurrent loads share one decode.
type Loader struct {
	mu       sync.Mutex
	col      *core.Collection
	rev      uint64
	rollover int
	loc      *time.Location
	all      []Review

	// last is the Data of the most recent load and lastScope its scope,
	// reused while the review log is.
	last      *Data
	lastScope Scope
}

// Load returns the Data of scope in c.
func (l *Loader) Load(c *core.Collection, scope Scope) (*Data, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := c.Now()
	rollover := c.Settings().RolloverHour
	rev := c.Revision()
	fresh := l.col == c && l.rev == rev && l.rollover == rollover && l.loc == now.Location()
	if fresh && l.last != nil && l.lastScope == scope && l.last.Today == scheduler.DayNumber(now, rollover) {
		d := *l.last
		d.Now = now
		return &d, nil
	}
	if !fresh {
		all, err := readReviews(c, now.Location(), rollover)
		if err != nil {
			return nil, err
		}
		l.col, l.rev, l.rollover, l.loc, l.all = c, rev, rollover, now.Location(), all
	}
	d, err := resolve(c, scope, l.all, now, rollover)
	if err != nil {
		return nil, err
	}
	l.last, l.lastScope = d, scope
	return d, nil
}

// resolve builds the Data of scope from the whole review log all.
func resolve(c *core.Collection, scope Scope, all []Review, now time.Time, rollover int) (*Data, error) {
	d := &Data{Now: now, Today: scheduler.DayNumber(now, rollover), Rollover: rollover}
	cards := c.AllCards()
	switch scope.Kind {
	case ScopeCollection:
		d.Cards, d.Reviews = cards, all
	case ScopeDeck:
		deck, err := c.Deck(core.ID(scope.DeckID))
		if err != nil {
			return nil, err
		}
		decks := map[core.ID]bool{}
		for _, sub := range c.Decks() {
			if sub.Name == deck.Name || strings.HasPrefix(sub.Name, deck.Name+core.DeckSeparator) {
				decks[sub.ID] = true
			}
		}
		// A deck's cards include those away in filtered decks, and a
		// filtered deck's the cards in it.
		d.Cards = slices.DeleteFunc(cards, func(card core.Card) bool { return !decks[card.DeckID] && !decks[card.HomeDeck()] })
	case ScopeSearch:
		ids, err := c.Search(scope.Query, core.ColCreated, false)
		if err != nil {
			return nil, err
		}
		match := make(map[core.ID]bool, len(ids))
		for _, id := range ids {
			match[id] = true
		}
		d.Cards = slices.DeleteFunc(cards, func(card core.Card) bool { return !match[card.ID] })
	default:
		return nil, fmt.Errorf("stats: unknown scope %q", scope.Kind)
	}
	if scope.Kind != ScopeCollection {
		in := make(map[core.ID]bool, len(d.Cards))
		for _, card := range d.Cards {
			in[card.ID] = true
		}
		for _, r := range all {
			if in[r.CardID] {
				d.Reviews = append(d.Reviews, r)
			}
		}
	}
	slices.SortFunc(d.Cards, func(a, b core.Card) int { return cmp.Compare(a.ID, b.ID) })
	retention, err := desiredRetention(c, scope, d.Cards)
	if err != nil {
		return nil, err
	}
	d.Retention = retention
	return d, nil
}

// desiredRetention is Data.Retention.
func desiredRetention(c *core.Collection, scope Scope, cards []core.Card) (float64, error) {
	if scope.Kind == ScopeDeck {
		opts, err := c.DeckOptions(core.ID(scope.DeckID))
		return opts.DesiredRetention, err
	}
	presetOf := map[core.ID]core.ID{}
	for _, d := range c.Decks() {
		presetOf[d.ID] = d.PresetID
	}
	shared := core.DefaultPresetID
	for i, card := range cards {
		p := presetOf[card.HomeDeck()]
		if i > 0 && p != shared {
			shared = core.DefaultPresetID
			break
		}
		shared = p
	}
	for _, p := range c.Presets() {
		if p.ID == shared {
			return p.Options.DesiredRetention, nil
		}
	}
	return core.DefaultRetention, nil
}

// entry is the part of a core.ReviewEntry the statistics read.
type entry struct {
	CardID      core.ID          `json:"card"`
	Time        time.Time        `json:"time"`
	Rating      scheduler.Rating `json:"rating"`
	StateBefore scheduler.State  `json:"stateBefore"`
	IvlBefore   int              `json:"ivlBefore"`
	IvlAfter    int              `json:"ivlAfter"`
	DurationMs  int64            `json:"ms"`
	// Preview answers changed no scheduling and are not counted.
	Preview bool `json:"preview"`
}

// readReviews decodes every answer in c's review log, months in parallel
// where Go has threads.
func readReviews(c *core.Collection, loc *time.Location, rollover int) ([]Review, error) {
	months, err := c.RevlogMonths()
	if err != nil {
		return nil, err
	}
	parts := make([][]Review, len(months))
	errs := make([]error, len(months))
	sem := make(chan struct{}, max(runtime.GOMAXPROCS(0), 1))
	var wg sync.WaitGroup
	for i, m := range months {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			parts[i], errs[i] = readMonth(c, m, loc, rollover)
		}()
	}
	wg.Wait()
	n := 0
	for i, err := range errs {
		if err != nil {
			return nil, err
		}
		n += len(parts[i])
	}
	out := make([]Review, 0, n)
	for _, p := range parts {
		out = append(out, p...)
	}
	return out, nil
}

func readMonth(c *core.Collection, month string, loc *time.Location, rollover int) ([]Review, error) {
	raws, err := c.ReadRevlog(month)
	if err != nil {
		return nil, err
	}
	pause := yield.Every(500)
	out := make([]Review, 0, len(raws))
	var e entry
	for i, raw := range raws {
		pause()
		e = entry{}
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("%w: review log %s entry %d: %v", core.ErrCorrupt, month, i, err)
		}
		if !e.Rating.Valid() || e.Preview {
			continue
		}
		t := e.Time.In(loc)
		out = append(out, Review{CardID: e.CardID, Day: scheduler.DayNumber(t, rollover), Hour: int8(t.Hour()),
			Rating: e.Rating, Kind: reviewKind(&e), IvlBefore: int32(e.IvlBefore), DurationMs: max(e.DurationMs, 0)})
	}
	return out, nil
}
