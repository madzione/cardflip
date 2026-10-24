package core

import (
	"fmt"
	"sort"
	"strings"

	"cardflip/golang/scheduler"
)

func defaultPreset() *Preset {
	return &Preset{ID: DefaultPresetID, Name: DefaultPresetName, Options: DefaultDeckOptions()}
}

// presetOf returns the preset deck d uses. Callers hold a lock.
func (c *Collection) presetOf(d *Deck) *Preset { return c.preset(d.PresetID) }

// preset returns preset id, or Default when it does not exist. Callers hold
// a lock.
func (c *Collection) preset(id ID) *Preset {
	if p, ok := c.presets[id]; ok {
		return p
	}
	if p, ok := c.presets[DefaultPresetID]; ok {
		return p
	}
	return defaultPreset()
}

// optionsOf returns the options deck id is studied with: its preset's, with
// the deck's own limits applied. Callers hold a lock.
func (c *Collection) optionsOf(id ID) DeckOptions {
	d, ok := c.decks[id]
	if !ok {
		return DefaultDeckOptions()
	}
	opts := c.presetOf(d).Options
	if d.Limits.New != nil {
		opts.NewPerDay = *d.Limits.New
	}
	if d.Limits.Review != nil {
		opts.ReviewsPerDay = *d.Limits.Review
	}
	return opts
}

// newLimit is deck d's new card limit on study day day: today's override
// when set for that day, otherwise its daily limit. Callers hold a lock.
func (c *Collection) newLimit(d *Deck, day int) int {
	if d.Today.Day == day && d.Today.New != nil {
		return *d.Today.New
	}
	return c.optionsOf(d.ID).NewPerDay
}

// reviewLimit is deck d's review limit on study day day, like newLimit.
func (c *Collection) reviewLimit(d *Deck, day int) int {
	if d.Today.Day == day && d.Today.Review != nil {
		return *d.Today.Review
	}
	return c.optionsOf(d.ID).ReviewsPerDay
}

// EffectiveLimits returns deck id's new and review limits on the current
// study day: its own limits or its preset's, today's increases included.
func (c *Collection) EffectiveLimits(id ID) (newLimit, reviewLimit int, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.decks[id]
	if !ok {
		return 0, 0, userErr("deckGone", "That deck no longer exists.")
	}
	day := scheduler.DayNumber(c.now(), c.settings.RolloverHour)
	return c.newLimit(d, day), c.reviewLimit(d, day), nil
}

// DeckOptions returns the options of deck id's preset, without the deck's
// own limits.
func (c *Collection) DeckOptions(id ID) (DeckOptions, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.decks[id]
	if !ok {
		return DeckOptions{}, userErr("deckGone", "That deck no longer exists.")
	}
	return c.presetOf(d).Options, nil
}

// setPresetOptions changes preset p's options. When its insertion order
// changes, the new cards of the decks using it are repositioned; with
// RescheduleOnChange, a change of desired retention or FSRS parameters
// reschedules their review cards. Callers hold the write lock.
func (c *Collection) setPresetOptions(p *Preset, opts DeckOptions) error {
	if p.Options.Equal(opts) {
		return nil
	}
	np := *p
	np.Options = opts
	o := &op{Kind: "deckOptions", Presets: []*Preset{&np}}
	reschedule := map[ID]bool{}
	for _, d := range c.decks {
		if c.presetOf(d).ID != p.ID {
			continue
		}
		if opts.InsertionOrder != p.Options.InsertionOrder {
			o.Cards = append(o.Cards, c.resortDeck(d.ID, opts.InsertionOrder == InsertRandom)...)
		}
		if d.Filter == nil && needsReschedule(p.Options, opts) {
			reschedule[d.ID] = true
		}
	}
	if err := c.rescheduleInto(o, reschedule, opts); err != nil {
		return err
	}
	return c.commit(o)
}

// Presets returns all presets sorted by name, ignoring case.
func (c *Collection) Presets() []Preset {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Preset, 0, len(c.presets))
	for _, p := range c.presets {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i].Name), strings.ToLower(out[j].Name)
		return a < b || a == b && out[i].ID < out[j].ID
	})
	return out
}

// PresetUsage returns how many decks use each preset; unused presets are
// missing.
func (c *Collection) PresetUsage() map[ID]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := map[ID]int{}
	for _, d := range c.decks {
		out[c.presetOf(d).ID]++
	}
	return out
}

// normalizePresetName trims a preset name and checks that no other preset
// than except has it, ignoring case.
func (c *Collection) normalizePresetName(name string, except ID) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "", userErr("presetNameEmpty", "Please enter a preset name.")
	}
	for _, p := range c.presets {
		if p.ID != except && strings.EqualFold(p.Name, name) {
			return "", userErr("presetNameTaken", "A preset named %q already exists.", p.Name)
		}
	}
	return name, nil
}

// AddPreset creates a preset with the default options and makes deck use
// it. It returns the new preset's ID.
func (c *Collection) AddPreset(deck ID, name string) (ID, error) {
	return c.newPreset(deck, name, "addPreset", func(*Preset) DeckOptions { return DefaultDeckOptions() })
}

// ClonePreset creates a preset with the options of deck's preset and makes
// deck use it. It returns the new preset's ID.
func (c *Collection) ClonePreset(deck ID, name string) (ID, error) {
	return c.newPreset(deck, name, "clonePreset", func(p *Preset) DeckOptions { return p.Options })
}

func (c *Collection) newPreset(deck ID, name, kind string, options func(*Preset) DeckOptions) (ID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deck]
	if !ok {
		return 0, userErr("deckGone", "That deck no longer exists.")
	}
	name, err := c.normalizePresetName(name, 0)
	if err != nil {
		return 0, err
	}
	p := &Preset{ID: c.ids.next(), Name: name, Options: options(c.presetOf(d))}
	o := &op{Kind: kind, Presets: []*Preset{p}}
	c.assignInto(o, []*Deck{d}, p)
	return p.ID, c.commit(o)
}

// RenamePreset renames a preset. Names are unique, ignoring case.
func (c *Collection) RenamePreset(id ID, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	p, ok := c.presets[id]
	if !ok {
		return userErr("presetGone", "That preset no longer exists.")
	}
	name, err := c.normalizePresetName(name, id)
	if err != nil {
		return err
	}
	if name == p.Name {
		return nil
	}
	np := *p
	np.Name = name
	return c.commit(&op{Kind: "renamePreset", Presets: []*Preset{&np}})
}

// DeletePreset deletes a preset; the decks using it use Default instead.
// Default cannot be deleted.
func (c *Collection) DeletePreset(id ID) error {
	if id == DefaultPresetID {
		return userErr("defaultPresetUndeletable", "The Default preset cannot be deleted.")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.presets[id]; !ok {
		return userErr("presetGone", "That preset no longer exists.")
	}
	o := &op{Kind: "deletePreset", DelPresets: []ID{id}}
	var users []*Deck
	for _, d := range c.decks {
		if d.PresetID == id {
			users = append(users, d)
		}
	}
	c.assignInto(o, users, c.preset(DefaultPresetID))
	return c.commit(o)
}

// SetDeckPreset makes deck use preset.
func (c *Collection) SetDeckPreset(deck, preset ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deck]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	p, ok := c.presets[preset]
	if !ok {
		return userErr("presetGone", "That preset no longer exists.")
	}
	o := &op{Kind: "deckPreset"}
	c.assignInto(o, []*Deck{d}, p)
	return c.commit(o)
}

// SavePresetToSubdecks makes every descendant of deck use deck's preset.
func (c *Collection) SavePresetToSubdecks(deck ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deck]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	var sub []*Deck
	for _, t := range c.subtree(d) {
		if t.ID != d.ID {
			sub = append(sub, t)
		}
	}
	o := &op{Kind: "presetToSubdecks"}
	c.assignInto(o, sub, c.presetOf(d))
	return c.commit(o)
}

// SetPresetOptions changes a preset's options, and so those of every deck
// using it. Invalid values are rejected and the previous options kept.
func (c *Collection) SetPresetOptions(id ID, opts DeckOptions) error {
	if err := c.validateOptions(opts); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	p, ok := c.presets[id]
	if !ok {
		return userErr("presetGone", "That preset no longer exists.")
	}
	return c.setPresetOptions(p, opts)
}

// assignInto adds making decks use preset p to o. Decks whose insertion
// order changes get their new cards repositioned. Callers hold the write
// lock; p may be in o.Presets rather than in the collection.
func (c *Collection) assignInto(o *op, decks []*Deck, p *Preset) {
	for _, d := range decks {
		if d.PresetID == p.ID {
			continue
		}
		old := c.presetOf(d)
		nd := *d
		nd.PresetID = p.ID
		o.Decks = append(o.Decks, &nd)
		if old.Options.InsertionOrder != p.Options.InsertionOrder {
			o.Cards = append(o.Cards, c.resortDeck(d.ID, p.Options.InsertionOrder == InsertRandom)...)
		}
	}
}

// SetDeckLimits sets deck's own new and review limits, which override its
// preset's for that deck only; nil uses the preset's.
func (c *Collection) SetDeckLimits(deck ID, limits DeckLimits) error {
	if err := validateLimits(limits); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deck]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	if equalLimit(d.Limits.New, limits.New) && equalLimit(d.Limits.Review, limits.Review) {
		return nil
	}
	nd := *d
	nd.Limits = DeckLimits{New: copyLimit(limits.New), Review: copyLimit(limits.Review)}
	return c.commit(&op{Kind: "deckLimits", Decks: []*Deck{&nd}})
}

func equalLimit(a, b *int) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }

func copyLimit(n *int) *int {
	if n == nil {
		return nil
	}
	v := *n
	return &v
}

// migratePresets turns the options of decks stored before presets existed
// (snapshot version 1) into presets (design decision 2): decks with the
// default options of that time (legacyDeckOptions) use Default, and each
// other distinct set of options becomes a preset named after the first
// deck, in deck order, that has it. It returns the presets to add (Default,
// with the default options of that time, too when hasDefault is false) and
// the migrated decks; both are empty when no deck needs migrating. decks
// are not changed.
func migratePresets(decks []*Deck, hasDefault bool, next func() ID) ([]*Preset, []*Deck) {
	var legacy []Deck
	for _, d := range decks {
		if d.legacy != nil {
			legacy = append(legacy, *d)
		}
	}
	if len(legacy) == 0 {
		return nil, nil
	}
	sortDecks(legacy)
	def := legacyDeckOptions()
	var presets []*Preset
	if !hasDefault {
		presets = append(presets, &Preset{ID: DefaultPresetID, Name: DefaultPresetName, Options: def})
	}
	taken := map[string]bool{strings.ToLower(DefaultPresetName): true}
	var out []*Deck
	for _, d := range legacy {
		opts := *d.legacy
		d.legacy = nil
		d.PresetID = DefaultPresetID
		if !opts.Equal(def) {
			var match *Preset
			for _, p := range presets {
				if p.ID != DefaultPresetID && p.Options.Equal(opts) {
					match = p
					break
				}
			}
			if match == nil {
				name := d.Name
				for i := 2; taken[strings.ToLower(name)]; i++ {
					name = fmt.Sprintf("%s (%d)", d.Name, i)
				}
				taken[strings.ToLower(name)] = true
				match = &Preset{ID: next(), Name: name, Options: opts}
				presets = append(presets, match)
			}
			d.PresetID = match.ID
		}
		out = append(out, &d)
	}
	return presets, out
}

// unburyDefault returns Default with sibling burying off, or nil when it
// buries nothing: snapshots before version 3 had it on by default, and the
// Default preset Cardflip ships buries nothing since. Other presets keep
// their options.
func unburyDefault(p *Preset) *Preset {
	if !p.Options.BuryNew && !p.Options.BuryReview {
		return nil
	}
	np := *p
	np.Options.BuryNew, np.Options.BuryReview = false, false
	return &np
}

// migrateState migrates a restored backup's decks to presets (see
// migratePresets), and turns sibling burying off in the Default preset of
// a backup before version 3 (see unburyDefault).
func migrateState(s *state, next func() ID) {
	hasDefault := false
	for _, p := range s.Presets {
		hasDefault = hasDefault || p.ID == DefaultPresetID
	}
	presets, migrated := migratePresets(s.Decks, hasDefault, next)
	if !hasDefault && len(presets) == 0 {
		presets = []*Preset{defaultPreset()}
	}
	s.Presets = append(s.Presets, presets...)
	byID := map[ID]*Deck{}
	for _, d := range migrated {
		byID[d.ID] = d
	}
	for i, d := range s.Decks {
		if m, ok := byID[d.ID]; ok {
			s.Decks[i] = m
		}
	}
	if s.Version < 3 {
		for i, p := range s.Presets {
			if p.ID != DefaultPresetID {
				continue
			}
			if np := unburyDefault(p); np != nil {
				s.Presets[i] = np
			}
		}
	}
	s.Version = snapshotVersion
}

// migrateDecks migrates the loaded collection's decks to presets (see
// migratePresets) and reports whether there were any to migrate. Callers
// hold the write lock or are in Open.
func (c *Collection) migrateDecks() bool {
	decks := make([]*Deck, 0, len(c.decks))
	for _, d := range c.decks {
		decks = append(decks, d)
	}
	_, hasDefault := c.presets[DefaultPresetID]
	presets, migrated := migratePresets(decks, hasDefault, c.ids.next)
	if len(migrated) == 0 && hasDefault {
		return false
	}
	if !hasDefault && len(presets) == 0 {
		presets = []*Preset{defaultPreset()}
	}
	c.apply(&op{Presets: presets, Decks: migrated})
	return true
}

// SaveDeckOptions saves the deck options screen in one step: deck uses
// preset, which gets opts, and deck gets its own limits. Decks whose
// insertion order changes get their new cards repositioned, and those that
// reschedule on change (see RescheduleCount) their review cards. Invalid
// values are rejected and nothing changes.
func (c *Collection) SaveDeckOptions(deck, preset ID, opts DeckOptions, limits DeckLimits) error {
	if err := c.validateOptions(opts); err != nil {
		return err
	}
	if err := validateLimits(limits); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[deck]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	p, ok := c.presets[preset]
	if !ok {
		return userErr("presetGone", "That preset no longer exists.")
	}
	o := &op{Kind: "deckOptions"}
	reschedule := c.decksToReschedule(deck, preset, opts)
	if !p.Options.Equal(opts) {
		np := *p
		np.Options = opts
		o.Presets = append(o.Presets, &np)
	}
	if d.PresetID != preset || !equalLimit(d.Limits.New, limits.New) || !equalLimit(d.Limits.Review, limits.Review) {
		nd := *d
		nd.PresetID = preset
		nd.Limits = DeckLimits{New: copyLimit(limits.New), Review: copyLimit(limits.Review)}
		o.Decks = append(o.Decks, &nd)
	}
	for _, x := range c.decks {
		before := c.presetOf(x).Options.InsertionOrder
		after := before
		if x.ID == deck || x.PresetID == preset {
			after = opts.InsertionOrder
		}
		if after != before {
			o.Cards = append(o.Cards, c.resortDeck(x.ID, after == InsertRandom)...)
		}
	}
	if err := c.rescheduleInto(o, reschedule, opts); err != nil {
		return err
	}
	return c.commit(o)
}

func validateLimits(limits DeckLimits) error {
	if n := limits.New; n != nil && (*n < 0 || *n > 9999) {
		return userErr("newPerDayRange", "New cards per day must be between 0 and 9999.")
	}
	if n := limits.Review; n != nil && (*n < 0 || *n > 99999) {
		return userErr("reviewsPerDayRange", "Maximum reviews per day must be between 0 and 99999.")
	}
	return nil
}
