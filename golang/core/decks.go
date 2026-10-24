package core

import (
	"maps"
	"sort"
	"strings"
	"time"

	"cardflip/golang/scheduler"
)

// normalizeDeckName trims each "::" component and rejects empty ones.
func normalizeDeckName(name string) (string, error) {
	name = nfc(name)
	parts := strings.Split(name, DeckSeparator)
	for i, p := range parts {
		p = strings.Join(strings.Fields(p), " ")
		if p == "" {
			if strings.TrimSpace(name) == "" {
				return "", userErr("deckNameEmpty", "Please enter a deck name.")
			}
			return "", userErr("deckNameEmptyPart", "Deck name %q has an empty part between '::'.", name)
		}
		parts[i] = p
	}
	return strings.Join(parts, DeckSeparator), nil
}

func parentName(name string) string {
	if i := strings.LastIndex(name, DeckSeparator); i >= 0 {
		return name[:i]
	}
	return ""
}

func baseName(name string) string {
	if i := strings.LastIndex(name, DeckSeparator); i >= 0 {
		return name[i+len(DeckSeparator):]
	}
	return name
}

// deckByName finds a deck by full name, ignoring case.
func (c *Collection) deckByName(name string) *Deck {
	for _, d := range c.decks {
		if strings.EqualFold(d.Name, name) {
			return d
		}
	}
	return nil
}

// isInTree reports whether deck name is root or one of its descendants.
func isInTree(name, root string) bool {
	return strings.EqualFold(name, root) || len(name) > len(root)+2 && strings.EqualFold(name[:len(root)+2], root+DeckSeparator)
}

// subtree returns root and all its descendant decks.
func (c *Collection) subtree(root *Deck) []*Deck {
	var out []*Deck
	for _, d := range c.decks {
		if isInTree(d.Name, root.Name) {
			out = append(out, d)
		}
	}
	return out
}

// resolvePath maps a normalized name onto existing decks: every component
// whose path exists (ignoring case) takes the existing spelling. It returns
// the resolved name and the parent decks that must be created.
func (c *Collection) resolvePath(name string, exclude map[ID]bool) (string, []*Deck) {
	parts := strings.Split(name, DeckSeparator)
	resolved := ""
	var create []*Deck
	for i, p := range parts {
		path := p
		if resolved != "" {
			path = resolved + DeckSeparator + p
		}
		if d := c.deckByName(path); d != nil && !exclude[d.ID] {
			resolved = d.Name
			continue
		}
		resolved = path
		if i < len(parts)-1 {
			create = append(create, &Deck{ID: c.ids.next(), Name: path, PresetID: DefaultPresetID})
		}
	}
	return resolved, create
}

// Decks returns all decks sorted by name, parents before their children.
func (c *Collection) Decks() []Deck {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Deck, 0, len(c.decks))
	for _, d := range c.decks {
		out = append(out, *d)
	}
	sortDecks(out)
	return out
}

func sortDecks(decks []Deck) {
	sort.Slice(decks, func(i, j int) bool {
		a := strings.Split(strings.ToLower(decks[i].Name), DeckSeparator)
		b := strings.Split(strings.ToLower(decks[j].Name), DeckSeparator)
		for k := 0; k < len(a) && k < len(b); k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return len(a) < len(b)
	})
}

// Deck returns one deck.
func (c *Collection) Deck(id ID) (Deck, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.decks[id]
	if !ok {
		return Deck{}, userErr("deckGone", "That deck no longer exists.")
	}
	return *d, nil
}

// CreateDeck creates a deck, and any missing parents of a nested name such
// as "Languages::Spanish". It returns the new deck's ID.
func (c *Collection) CreateDeck(name string) (ID, error) {
	name, err := normalizeDeckName(name)
	if err != nil {
		return 0, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.createDeckLocked(name)
}

func (c *Collection) createDeckLocked(name string) (ID, error) {
	if d := c.deckByName(name); d != nil {
		return 0, userErr("deckNameTaken", "A deck named %q already exists.", d.Name)
	}
	if err := c.checkParents(name); err != nil {
		return 0, err
	}
	resolved, parents := c.resolvePath(name, nil)
	d := &Deck{ID: c.ids.next(), Name: resolved, PresetID: DefaultPresetID}
	if err := c.commit(&op{Kind: "createDeck", Decks: append(parents, d)}); err != nil {
		return 0, err
	}
	return d.ID, nil
}

// deckForImport returns the deck with this name, creating it (and its
// parents) inside the op being built when it does not exist.
func (c *Collection) deckForImport(name string, o *op, created map[string]*Deck) (ID, error) {
	name, err := normalizeDeckName(name)
	if err != nil {
		return 0, err
	}
	if d := c.deckByName(name); d != nil {
		if d.Filter != nil {
			return 0, errFilteredTarget(d)
		}
		return d.ID, nil
	}
	if err := c.checkParents(name); err != nil {
		return 0, err
	}
	if d, ok := created[strings.ToLower(name)]; ok {
		return d.ID, nil
	}
	parts := strings.Split(name, DeckSeparator)
	path := ""
	var id ID
	for _, p := range parts {
		if path != "" {
			path += DeckSeparator
		}
		path += p
		if d := c.deckByName(path); d != nil {
			path, id = d.Name, d.ID
			continue
		}
		if d, ok := created[strings.ToLower(path)]; ok {
			path, id = d.Name, d.ID
			continue
		}
		d := &Deck{ID: c.ids.next(), Name: path, PresetID: DefaultPresetID}
		created[strings.ToLower(path)] = d
		o.Decks = append(o.Decks, d)
		id = d.ID
	}
	return id, nil
}

// RenameDeck renames or moves a deck. Its children move with it. The new
// name must not be taken by a sibling (ignoring case), and a deck cannot
// move into its own subtree.
func (c *Collection) RenameDeck(id ID, newName string) error {
	newName, err := normalizeDeckName(newName)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[id]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	if newName == d.Name {
		return nil
	}
	if isInTree(newName, d.Name) && !strings.EqualFold(newName, d.Name) {
		return userErr("deckMoveIntoItself", "A deck cannot be moved into itself or one of its subdecks.")
	}
	tree := c.subtree(d)
	exclude := map[ID]bool{}
	for _, t := range tree {
		exclude[t.ID] = true
	}
	if other := c.deckByName(newName); other != nil && other.ID != id {
		return userErr("deckNameTakenByOther", "The name %q is already taken by another deck.", baseName(other.Name))
	}
	if err := c.checkParents(newName); err != nil {
		return err
	}
	resolved, parents := c.resolvePath(newName, exclude)
	o := &op{Kind: "renameDeck", Decks: parents}
	for _, t := range tree {
		nd := *t
		nd.Name = resolved + t.Name[len(d.Name):]
		o.Decks = append(o.Decks, &nd)
	}
	return c.commit(o)
}

// validateOptions is ValidateDeckOptions plus a check of the optimize
// search. It takes the read lock, so callers hold no lock.
func (c *Collection) validateOptions(opts DeckOptions) error {
	if err := ValidateDeckOptions(opts); err != nil {
		return err
	}
	if strings.TrimSpace(opts.OptimizeSearch) != "" {
		if err := c.CheckSearch(opts.OptimizeSearch); err != nil {
			return userErr("optimizeSearchInvalid", "The optimize search is invalid: %s", err.Error())
		}
	}
	return nil
}

// ValidateDeckOptions checks deck options against their allowed ranges.
func ValidateDeckOptions(opts DeckOptions) error {
	if opts.DesiredRetention < MinRetention || opts.DesiredRetention > MaxRetention {
		return userErr("desiredRetentionRange", "Desired retention must be between %.2f and %.2f.", MinRetention, MaxRetention)
	}
	if opts.NewPerDay < 0 || opts.NewPerDay > 9999 {
		return userErr("newPerDayRange", "New cards per day must be between 0 and 9999.")
	}
	if opts.ReviewsPerDay < 0 || opts.ReviewsPerDay > 99999 {
		return userErr("reviewsPerDayRange", "Maximum reviews per day must be between 0 and 99999.")
	}
	for _, steps := range []struct {
		name, code string
		steps      []int
	}{{"Learning steps", "learningSteps", opts.LearningSteps}, {"Relearning steps", "relearningSteps", opts.RelearningSteps}} {
		if len(steps.steps) > MaxSteps {
			return userErr(steps.code+"TooMany", steps.name+" can have at most %d steps.", MaxSteps)
		}
		for _, s := range steps.steps {
			if s <= 0 || s > MaxStepSeconds {
				return userErr(steps.code+"Range", steps.name+" must each be longer than 0 and at most 365 days.")
			}
		}
	}
	switch opts.InterdayLearningOrder {
	case InterdayMix, InterdayBefore, InterdayAfter:
	default:
		return userErr("unknownInterdayOrder", "Unknown interday learning/review order %q.", opts.InterdayLearningOrder)
	}
	if opts.LeechThreshold < 1 || opts.LeechThreshold > MaxLeechThreshold {
		return userErr("leechThresholdRange", "Leech threshold must be between 1 and %d lapses.", MaxLeechThreshold)
	}
	switch opts.LeechAction {
	case LeechTag, LeechSuspend:
	default:
		return userErr("unknownLeechAction", "Unknown leech action %q.", opts.LeechAction)
	}
	if opts.MaxInterval < 1 || opts.MaxInterval > MaxIntervalDays {
		return userErr("maxIntervalRange", "Maximum interval must be between 1 and %d days.", MaxIntervalDays)
	}
	if opts.MinRelearnInterval < 1 || opts.MinRelearnInterval > MaxIntervalDays {
		return userErr("minIntervalRange", "Minimum interval must be between 1 and %d days.", MaxIntervalDays)
	}
	if opts.MinRelearnInterval > opts.MaxInterval {
		return userErr("minIntervalAboveMax", "Minimum interval (%d days) must not be above the maximum interval (%d days).", opts.MinRelearnInterval, opts.MaxInterval)
	}
	switch opts.NewGatherOrder {
	case GatherDeck, GatherDeckRandomNotes, GatherAscending, GatherDescending, GatherRandomNotes, GatherRandomCards:
	default:
		return userErr("unknownGatherOrder", "Unknown new card gather order %q.", opts.NewGatherOrder)
	}
	switch opts.NewSortOrder {
	case SortGathered, SortTypeThenGathered, SortTypeThenRandom, SortRandomNoteThenType, SortRandom:
	default:
		return userErr("unknownNewSortOrder", "Unknown new card sort order %q.", opts.NewSortOrder)
	}
	switch opts.NewReviewOrder {
	case NewMix, NewBeforeReviews, NewAfterReviews:
	default:
		return userErr("unknownNewReviewOrder", "Unknown new/review order %q.", opts.NewReviewOrder)
	}
	switch opts.ReviewSortOrder {
	case ReviewDueRandom, ReviewDueDeck, ReviewDeckDue, ReviewIntervalsAscending, ReviewIntervalsDescending,
		ReviewDifficultyAscending, ReviewDifficultyDescending, ReviewRetrievabilityAscending, ReviewRelativeOverdueness:
	default:
		return userErr("unknownReviewSortOrder", "Unknown review sort order %q.", opts.ReviewSortOrder)
	}
	switch opts.InsertionOrder {
	case InsertSequential, InsertRandom:
	default:
		return userErr("unknownInsertionOrder", "Unknown insertion order %q.", opts.InsertionOrder)
	}
	if opts.MaxAnswerSeconds < 1 || opts.MaxAnswerSeconds > MaxMaxAnswerSeconds {
		return userErr("maxAnswerSecondsRange", "Maximum answer seconds must be between 1 and %d.", MaxMaxAnswerSeconds)
	}
	for _, side := range []struct {
		name, code string
		seconds    int
	}{{"Seconds to show question for", "secondsQuestionRange", opts.SecondsToShowQuestion},
		{"Seconds to show answer for", "secondsAnswerRange", opts.SecondsToShowAnswer}} {
		if side.seconds < 0 || side.seconds > MaxAutoAdvanceSeconds {
			return userErr(side.code, side.name+" must be between 0 and %d.", MaxAutoAdvanceSeconds)
		}
	}
	switch opts.QuestionAction {
	case QuestionShowAnswer, QuestionShowReminder:
	default:
		return userErr("unknownQuestionAction", "Unknown question action %q.", opts.QuestionAction)
	}
	switch opts.AnswerAction {
	case AnswerBuryCard, AnswerAgain, AnswerHard, AnswerGood, AnswerShowReminder:
	default:
		return userErr("unknownAnswerAction", "Unknown answer action %q.", opts.AnswerAction)
	}
	if err := scheduler.CheckWeights(opts.Weights); err != nil {
		return &Error{Msg: err.Error()} // scheduler tags its messages
	}
	if opts.HistoricalRetention < MinHistoricalRetention || opts.HistoricalRetention > MaxHistoricalRetention {
		return userErr("historicalRetentionRange", "Historical retention must be between %.2f and %.2f.", MinHistoricalRetention, MaxHistoricalRetention)
	}
	if opts.IgnoreBefore != "" {
		if _, err := time.Parse(time.DateOnly, opts.IgnoreBefore); err != nil {
			return userErr("invalidDate", "%q is not a date like 2026-01-31.", opts.IgnoreBefore)
		}
	}
	for _, level := range opts.EasyDays {
		if level < EasyNormal || level > EasyMinimum {
			return userErr("unknownEasyDaysLevel", "Unknown Easy Days level %d.", level)
		}
	}
	return nil
}

// SetDeckOptions changes the options of deck id's preset, and so of every
// deck using it. Invalid values are rejected and the previous options kept.
func (c *Collection) SetDeckOptions(id ID, opts DeckOptions) error {
	if err := c.validateOptions(opts); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[id]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	return c.setPresetOptions(c.presetOf(d), opts)
}

// SetDeckCollapsed records whether a parent deck is collapsed in the list.
func (c *Collection) SetDeckCollapsed(id ID, collapsed bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[id]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	if d.Collapsed == collapsed {
		return nil
	}
	nd := *d
	nd.Collapsed = collapsed
	return c.commit(&op{Kind: "collapseDeck", Decks: []*Deck{&nd}})
}

// SetDeckDescription changes a deck's description; surrounding whitespace
// is dropped.
func (c *Collection) SetDeckDescription(id ID, description string) error {
	description = strings.TrimSpace(nfc(description))
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[id]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	if d.Description == description {
		return nil
	}
	nd := *d
	nd.Description = description
	return c.commit(&op{Kind: "deckDescription", Decks: []*Deck{&nd}})
}

// MaxTodayIncrease is the most a custom study increase may add to a limit.
const MaxTodayIncrease = 9999

// IncreaseTodayLimits raises a deck's new and review limits for the current
// study day only, by newBy and reviewBy over their effective limits; zero
// leaves a limit as it is. The deck's options do not change.
func (c *Collection) IncreaseTodayLimits(id ID, newBy, reviewBy int) error {
	for _, by := range []int{newBy, reviewBy} {
		if by < 0 || by > MaxTodayIncrease {
			return userErr("todayIncreaseRange", "The increase must be between 0 and %d cards.", MaxTodayIncrease)
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[id]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	if newBy == 0 && reviewBy == 0 {
		return nil
	}
	day := scheduler.DayNumber(c.now(), c.settings.RolloverHour)
	nd := *d
	if nd.Today.Day != day {
		nd.Today = TodayLimits{Day: day}
	}
	if newBy > 0 {
		n := c.newLimit(d, day) + newBy
		nd.Today.New = &n
	}
	if reviewBy > 0 {
		r := c.reviewLimit(d, day) + reviewBy
		nd.Today.Review = &r
	}
	return c.commit(&op{Kind: "todayLimits", Decks: []*Deck{&nd}})
}

// DeckDeletion describes what deleting a deck removes, for the confirmation.
type DeckDeletion struct {
	Decks int `json:"decks"`
	// Cards are the cards deleted: those whose home is one of the decks,
	// InFiltered of them currently in filtered decks elsewhere.
	Cards      int `json:"cards"`
	InFiltered int `json:"inFiltered"`
	// Returned are the cards in filtered decks among the decks that return
	// to their home decks.
	Returned int `json:"returned"`
}

// DeckDeletionInfo counts the decks and cards that DeleteDeck(id) removes.
func (c *Collection) DeckDeletionInfo(id ID) (DeckDeletion, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.decks[id]
	if !ok {
		return DeckDeletion{}, userErr("deckGone", "That deck no longer exists.")
	}
	tree := c.subtree(d)
	in := map[ID]bool{}
	for _, t := range tree {
		in[t.ID] = true
	}
	out := DeckDeletion{Decks: len(tree)}
	for _, card := range c.cards {
		switch {
		case in[card.HomeDeck()]:
			out.Cards++
			if !in[card.DeckID] {
				out.InFiltered++
			}
		case in[card.DeckID]:
			out.Returned++
		}
	}
	return out, nil
}

// DeleteDeck deletes a deck and its subdecks with every card whose home
// deck is one of them, wherever it is now; cards in filtered decks among
// them return to their home decks. Notes left without cards are deleted
// too. The last remaining regular deck cannot be deleted.
func (c *Collection) DeleteDeck(id ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.decks[id]
	if !ok {
		return userErr("deckGone", "That deck no longer exists.")
	}
	tree := c.subtree(d)
	o := &op{Kind: "deleteDeck"}
	in := map[ID]bool{}
	for _, t := range tree {
		in[t.ID] = true
		o.DelDecks = append(o.DelDecks, t.ID)
	}
	regular := false
	for _, other := range c.decks {
		regular = regular || !in[other.ID] && other.Filter == nil
	}
	if !regular {
		return userErr("lastDeck", "The last deck cannot be deleted. A collection always has at least one deck.")
	}
	var cardIDs, returned []ID
	for _, card := range c.cards {
		switch {
		case in[card.HomeDeck()]:
			cardIDs = append(cardIDs, card.ID)
		case in[card.DeckID]:
			returned = append(returned, card.ID)
		}
	}
	c.deleteCardsInto(o, cardIDs)
	sortIDs(returned)
	for _, card := range c.cardCopies(returned) {
		card.returnHome()
		o.Cards = append(o.Cards, card)
	}
	lastDecks := maps.Clone(c.settings.LastDeckByNoteType)
	maps.DeleteFunc(lastDecks, func(_, deck ID) bool { return in[deck] })
	if in[c.settings.LastDeckID] || len(lastDecks) != len(c.settings.LastDeckByNoteType) {
		s := c.settings
		if in[s.LastDeckID] {
			s.LastDeckID = 0
		}
		s.LastDeckByNoteType = lastDecks
		o.Settings = &s
	}
	return c.commit(o)
}

// deleteCardsInto adds the deletion of cards, and of notes that would be
// left without cards, to o.
func (c *Collection) deleteCardsInto(o *op, cardIDs []ID) {
	gone := map[ID]bool{}
	notes := map[ID]bool{}
	for _, id := range cardIDs {
		if card, ok := c.cards[id]; ok && !gone[id] {
			gone[id] = true
			o.DelCards = append(o.DelCards, id)
			notes[card.NoteID] = true
		}
	}
	for nid := range notes {
		left := false
		for _, cid := range c.cardsByNote[nid] {
			if !gone[cid] {
				left = true
				break
			}
		}
		if !left {
			o.DelNotes = append(o.DelNotes, nid)
		}
	}
	sortIDs(o.DelCards)
	sortIDs(o.DelNotes)
}

func sortIDs(ids []ID) { sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] }) }
