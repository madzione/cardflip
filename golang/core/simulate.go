package core

import (
	"context"

	"cardflip/golang/scheduler"
)

// The workload simulator and retention suggestion (collection spec,
// "Workload simulation" and "Retention suggestion"). Both only read the
// collection.

// Simulation are the simulator's inputs besides the deck options.
type Simulation struct {
	Days          int
	NewPerDay     int
	ReviewsPerDay int
}

// MaxSimulationDays is the longest simulation.
const MaxSimulationDays = 3650

// validate checks s, with the retentions to simulate.
func (s Simulation) validate(retentions ...float64) error {
	if s.Days < 1 || s.Days > MaxSimulationDays {
		return userErr("simulationDaysRange", "Days to simulate must be between 1 and %d.", MaxSimulationDays)
	}
	if s.NewPerDay < 0 || s.NewPerDay > 9999 {
		return userErr("newPerDayRange", "New cards per day must be between 0 and 9999.")
	}
	if s.ReviewsPerDay < 0 || s.ReviewsPerDay > 99999 {
		return userErr("reviewsPerDayRange", "Maximum reviews per day must be between 0 and 99999.")
	}
	for _, r := range retentions {
		if r < MinRetention || r > MaxRetention {
			return userErr("desiredRetentionRange", "Desired retention must be between %.2f and %.2f.", MinRetention, MaxRetention)
		}
	}
	return nil
}

// simulationInput returns the cards of deck and its subdecks (by home
// deck, without suspended cards) as the simulation starts from them, and
// the user's behaviour from their review history. It reports the "reading"
// stage.
func (c *Collection) simulationInput(deck ID, opts DeckOptions, progress Progress) (scheduler.SimDeck, scheduler.SimBehaviour, error) {
	progress("reading", 0, 1)
	c.mu.RLock()
	root, ok := c.decks[deck]
	if !ok {
		c.mu.RUnlock()
		return scheduler.SimDeck{}, scheduler.SimBehaviour{}, userErr("deckGone", "That deck no longer exists.")
	}
	inTree := map[ID]bool{}
	for _, d := range c.subtree(root) {
		inTree[d.ID] = true
	}
	now, rollover := c.now(), c.settings.RolloverHour
	// In ID order, so the same collection gives the same simulation.
	var ids []ID
	for _, card := range c.cards {
		if inTree[card.HomeDeck()] {
			ids = append(ids, card.ID)
		}
	}
	sortIDs(ids)
	var sd scheduler.SimDeck
	want := make(map[ID]bool, len(ids))
	for _, id := range ids {
		card := c.cards[id]
		want[id] = true
		if card.Suspended {
			continue
		}
		if sc, ok := scheduler.SimCardOf(card.CardState, now, rollover, opts.Weights); ok {
			sd.Cards = append(sd.Cards, sc)
		} else {
			sd.NewCards++
		}
	}
	c.mu.RUnlock()
	hist, err := c.histories(want)
	if err != nil {
		return scheduler.SimDeck{}, scheduler.SimBehaviour{}, err
	}
	all := make([][]scheduler.HistoryReview, 0, len(hist))
	for _, id := range ids {
		if h := hist[id]; len(h) > 0 {
			all = append(all, h)
		}
	}
	progress("reading", 1, 1)
	return sd, scheduler.BehaviourOf(all, rollover), nil
}

// simConfig is the simulation of deck with opts, s, the user's behaviour b
// and the desired retention.
func simConfig(deck ID, opts DeckOptions, s Simulation, b scheduler.SimBehaviour, retention float64) scheduler.SimConfig {
	return scheduler.SimConfig{
		Days: s.Days, NewPerDay: s.NewPerDay, ReviewsPerDay: s.ReviewsPerDay,
		NewIgnoresReviewLimit: opts.NewIgnoresReviewLimit, DesiredRetention: retention,
		MaximumInterval: opts.MaxInterval, MinimumRelearnInterval: opts.MinRelearnInterval,
		LearningSteps: len(opts.LearningSteps), RelearningSteps: len(opts.RelearningSteps),
		Weights: opts.Weights, Behaviour: b, Seed: uint64(deck),
	}
}

// Simulate simulates deck and its subdecks with the options shown on the
// deck options screen (which may be unsaved) and s, once for each of the
// retentions. The same deck, options and inputs
// give the same results. progress gets the "reading" stage, then
// "simulating" with done and total counting simulations.
func (c *Collection) Simulate(ctx context.Context, deck ID, opts DeckOptions, s Simulation, retentions []float64, progress Progress) ([]scheduler.SimResult, error) {
	if err := s.validate(retentions...); err != nil {
		return nil, err
	}
	sd, b, err := c.simulationInput(deck, opts, progress)
	if err != nil {
		return nil, err
	}
	out := make([]scheduler.SimResult, 0, len(retentions))
	for i, r := range retentions {
		progress("simulating", i, len(retentions))
		res, err := scheduler.Simulate(ctx, sd, simConfig(deck, opts, s, b, r))
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	progress("simulating", len(retentions), len(retentions))
	return out, nil
}

// SuggestRetention suggests the desired retention for deck with opts and
// s (see scheduler.SuggestRetention). progress
// gets the "reading" stage, then "simulating" with done and total counting
// simulations.
func (c *Collection) SuggestRetention(ctx context.Context, deck ID, opts DeckOptions, s Simulation, progress Progress) (scheduler.RetentionSuggestion, error) {
	if err := s.validate(); err != nil {
		return scheduler.RetentionSuggestion{}, err
	}
	sd, b, err := c.simulationInput(deck, opts, progress)
	if err != nil {
		return scheduler.RetentionSuggestion{}, err
	}
	return scheduler.SuggestRetention(ctx, sd, simConfig(deck, opts, s, b, 0), func(done, total int) { progress("simulating", done, total) })
}
