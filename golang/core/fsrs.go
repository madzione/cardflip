package core

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"cardflip/golang/scheduler"
)

// FSRS parameter options: review histories for the optimizer and
// rescheduling when the desired retention or the parameters change
// (collection spec, "Reschedule on change").

// historyEntry is the part of a stored ReviewEntry FSRS needs.
type historyEntry struct {
	ID          ID               `json:"id"`
	CardID      ID               `json:"card"`
	Time        time.Time        `json:"time"`
	Rating      scheduler.Rating `json:"rating"`
	StateBefore scheduler.State  `json:"stateBefore"`
	IvlBefore   int              `json:"ivlBefore"`
	IvlAfter    int              `json:"ivlAfter"`
	DurationMs  int64            `json:"ms"`
	Preview     bool             `json:"preview"`
}

// histories reads the stored review log of the cards in want, oldest first
// (by time, then ID), without preview answers. It does not use c.mu.
func (c *Collection) histories(want map[ID]bool) (map[ID][]scheduler.HistoryReview, error) {
	out := map[ID][]scheduler.HistoryReview{}
	if len(want) == 0 {
		return out, nil
	}
	months, err := c.store.RevlogMonths()
	if err != nil {
		return nil, err
	}
	ids := map[ID][]ID{}
	needle := []byte(`"card":`)
	for _, m := range months {
		c.yield()
		entries, err := c.store.ReadRevlog(m)
		if err != nil {
			return nil, err
		}
		for i, raw := range entries {
			// Decoding only the entries of wanted cards keeps this fast.
			if !wantedEntry(raw, needle, want) {
				continue
			}
			var e historyEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				return nil, fmt.Errorf("%w: review log %s entry %d: %v", ErrCorrupt, m, i, err)
			}
			if !want[e.CardID] || e.Preview {
				continue
			}
			out[e.CardID] = append(out[e.CardID], scheduler.HistoryReview{Time: e.Time, Rating: e.Rating,
				StateBefore: e.StateBefore, IvlBefore: e.IvlBefore, IvlAfter: e.IvlAfter,
				Duration: time.Duration(e.DurationMs) * time.Millisecond})
			ids[e.CardID] = append(ids[e.CardID], e.ID)
		}
	}
	for card, h := range out {
		idx := make([]int, len(h))
		for i := range idx {
			idx[i] = i
		}
		slices.SortStableFunc(idx, func(a, b int) int {
			return cmp.Or(h[a].Time.Compare(h[b].Time), cmp.Compare(ids[card][a], ids[card][b]))
		})
		sorted := make([]scheduler.HistoryReview, len(h))
		for i, j := range idx {
			sorted[i] = h[j]
		}
		out[card] = sorted
	}
	return out, nil
}

// wantedEntry reports whether a raw review log entry may belong to a card
// in want, from the number after its "card": key.
func wantedEntry(raw, needle []byte, want map[ID]bool) bool {
	i := bytes.Index(raw, needle)
	if i < 0 {
		return true // let the decoder decide
	}
	rest := raw[i+len(needle):]
	j := 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	id, err := strconv.ParseInt(string(rest[:j]), 10, 64)
	return err != nil || want[ID(id)]
}

// historyOptions are how a deck's options turn histories into sequences.
// withCutoff applies IgnoreBefore, which only optimizing uses.
func (c *Collection) historyOptions(opts DeckOptions, withCutoff bool) scheduler.HistoryOptions {
	h := scheduler.HistoryOptions{RolloverHour: c.settings.RolloverHour, HistoricalRetention: opts.HistoricalRetention}
	if withCutoff && opts.IgnoreBefore != "" {
		if d, err := time.ParseInLocation(time.DateOnly, opts.IgnoreBefore, time.Local); err == nil {
			h.IgnoreBefore = time.Date(d.Year(), d.Month(), d.Day(), c.settings.RolloverHour, 0, 0, 0, time.Local)
		}
	}
	return h
}

// needsReschedule reports whether a deck's review cards are rescheduled
// when its options change from before to after.
func needsReschedule(before, after DeckOptions) bool {
	return after.RescheduleOnChange && (before.DesiredRetention != after.DesiredRetention ||
		!slices.Equal(before.Weights, after.Weights))
}

// rescheduledCards returns the review cards whose home deck is in decks.
// Callers hold a lock.
func (c *Collection) rescheduledCards(decks map[ID]bool) []*Card {
	var out []*Card
	for _, card := range c.cards {
		if card.State == scheduler.Review && decks[card.HomeDeck()] {
			out = append(out, card)
		}
	}
	return out
}

// rescheduleInto adds to o the review cards of decks, which get options
// after: their memory states replayed from their review history with
// after's parameters, and due dates at after's desired retention from their
// last review. Callers hold the write lock.
func (c *Collection) rescheduleInto(o *op, decks map[ID]bool, after DeckOptions) error {
	cards := c.rescheduledCards(decks)
	if len(cards) == 0 {
		return nil
	}
	want := map[ID]bool{}
	for _, card := range cards {
		want[card.ID] = true
	}
	hist, err := c.histories(want)
	if err != nil {
		return err
	}
	p := scheduler.DefaultParams(after.DesiredRetention)
	p.MaximumInterval = after.MaxInterval
	p.MinimumRelearnInterval = after.MinRelearnInterval
	p.RolloverHour = c.settings.RolloverHour
	p.Weights = after.Weights
	ho := c.historyOptions(after, false)
	for _, card := range cards {
		next := *card
		next.CardState = scheduler.Reschedule(card.CardState, scheduler.NewSequence(hist[card.ID], ho), p, int64(card.ID))
		o.Cards = append(o.Cards, &next)
	}
	return nil
}

// RescheduleCount returns how many review cards saving the deck options
// screen (see SaveDeckOptions) would reschedule.
func (c *Collection) RescheduleCount(deck, preset ID, opts DeckOptions) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.decks[deck]; !ok {
		return 0, userErr("deckGone", "That deck no longer exists.")
	}
	return len(c.rescheduledCards(c.decksToReschedule(deck, preset, opts))), nil
}

// decksToReschedule returns the decks whose review cards saving deck with
// preset and opts reschedules: deck and the other decks using preset, when
// their options change as needsReschedule says. Callers hold a lock.
func (c *Collection) decksToReschedule(deck, preset ID, opts DeckOptions) map[ID]bool {
	out := map[ID]bool{}
	for _, d := range c.decks {
		if d.Filter == nil && (d.ID == deck || d.PresetID == preset) && needsReschedule(c.presetOf(d).Options, opts) {
			out[d.ID] = true
		}
	}
	return out
}

// OptimizeSearch returns the search whose cards optimizing deck with opts
// learns from: opts.OptimizeSearch, or by default the deck and its
// subdecks.
func (c *Collection) OptimizeSearch(deck ID, opts DeckOptions) (string, error) {
	if s := strings.TrimSpace(opts.OptimizeSearch); s != "" {
		return s, nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.decks[deck]
	if !ok {
		return "", userErr("deckGone", "That deck no longer exists.")
	}
	return "deck:" + quoteSearch(d.Name), nil
}

// sequences returns the review sequences of the cards matching query, as
// opts turns them into training data. It reports the "reading" stage.
func (c *Collection) sequences(query string, opts DeckOptions, progress Progress) ([]scheduler.Sequence, error) {
	progress("reading", 0, 1)
	c.mu.RLock()
	ids, err := c.match(query, c.now())
	c.mu.RUnlock()
	if err != nil {
		return nil, userErr("optimizeSearchInvalid", "The optimize search is invalid: %s", err.Error())
	}
	want := make(map[ID]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	hist, err := c.histories(want)
	if err != nil {
		return nil, err
	}
	c.mu.RLock()
	ho := c.historyOptions(opts, true)
	c.mu.RUnlock()
	seqs := make([]scheduler.Sequence, 0, len(hist))
	for _, id := range ids {
		if h := hist[id]; len(h) > 0 {
			seqs = append(seqs, scheduler.NewSequence(h, ho))
		}
	}
	progress("reading", 1, 1)
	return seqs, nil
}

// Optimize fits FSRS parameters to the review history of the cards that
// optimizing deck with opts learns from (see OptimizeSearch), with
// opts.Weights as the current parameters. It saves nothing. Fewer than
// scheduler.MinOptimizeReviews reviews give a
// *scheduler.NotEnoughReviewsError; ctx stops it.
func (c *Collection) Optimize(ctx context.Context, deck ID, opts DeckOptions, progress Progress) (scheduler.OptimizeResult, error) {
	query, err := c.OptimizeSearch(deck, opts)
	if err != nil {
		return scheduler.OptimizeResult{}, err
	}
	seqs, err := c.sequences(query, opts, progress)
	if err != nil {
		return scheduler.OptimizeResult{}, err
	}
	return scheduler.Optimize(ctx, seqs, scheduler.OptimizeOptions{Current: opts.Weights, HealthCheck: opts.HealthCheck}, progress)
}

// PresetOptimization is the result of optimizing one preset.
type PresetOptimization struct {
	Preset Preset
	Result scheduler.OptimizeResult
	// NotEnough is set when the preset has too few reviews; its parameters
	// are unchanged.
	NotEnough *scheduler.NotEnoughReviewsError
	// Applied reports whether the preset got the fitted parameters.
	Applied bool
}

// OptimizeAll optimizes every preset: with its own optimize search, or
// by default the cards of the decks using it. Presets whose fitted
// parameters are better (and pass the health check, when asked for) get
// them, all in one undo step, rescheduling their review cards when they
// reschedule on change. progress gets the stages of each preset in turn,
// with done and total counting presets for the "preset" stage. When ctx
// is done nothing changes.
func (c *Collection) OptimizeAll(ctx context.Context, progress Progress) ([]PresetOptimization, error) {
	presets := c.Presets()
	var out []PresetOptimization
	for i, p := range presets {
		progress("preset", i, len(presets))
		query := strings.TrimSpace(p.Options.OptimizeSearch)
		if query == "" {
			query = "preset:" + quoteSearch(p.Name)
		}
		po := PresetOptimization{Preset: p}
		seqs, err := c.sequences(query, p.Options, progress)
		if err == nil {
			po.Result, err = scheduler.Optimize(ctx, seqs,
				scheduler.OptimizeOptions{Current: p.Options.Weights, HealthCheck: p.Options.HealthCheck}, progress)
		}
		if !errors.As(err, &po.NotEnough) && err != nil {
			return nil, fmt.Errorf("preset %q: %w", p.Name, err)
		}
		out = append(out, po)
	}
	progress("preset", len(presets), len(presets))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, c.applyOptimized(out)
}

// applyOptimized gives the presets of results their fitted parameters
// where they are to be used, in one step, and marks them applied.
func (c *Collection) applyOptimized(results []PresetOptimization) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	o := &op{Kind: "optimize"}
	for i := range results {
		r := &results[i]
		p, ok := c.presets[r.Preset.ID]
		if !ok || r.NotEnough != nil || !r.Result.Use() {
			continue
		}
		np := *p
		np.Options.Weights = r.Result.Weights
		o.Presets = append(o.Presets, &np)
		reschedule := map[ID]bool{}
		for _, d := range c.decks {
			if d.Filter == nil && c.presetOf(d).ID == p.ID && needsReschedule(p.Options, np.Options) {
				reschedule[d.ID] = true
			}
		}
		if err := c.rescheduleInto(o, reschedule, np.Options); err != nil {
			return err
		}
		r.Applied = true
	}
	if len(o.Presets) == 0 {
		return nil
	}
	return c.commit(o)
}

// RescheduleCountAll returns the number of review cards in decks whose
// preset reschedules on change: those OptimizeAll may reschedule.
func (c *Collection) RescheduleCountAll() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	decks := map[ID]bool{}
	for _, d := range c.decks {
		if d.Filter == nil && c.presetOf(d).Options.RescheduleOnChange {
			decks[d.ID] = true
		}
	}
	return len(c.rescheduledCards(decks))
}
