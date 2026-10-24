package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// DecksService manages decks.
type DecksService struct{ a *app }

// DeckOptions are a deck's study settings.
type DeckOptions struct {
	NewPerDay        int     `json:"newPerDay"`
	ReviewsPerDay    int     `json:"reviewsPerDay"`
	DesiredRetention float64 `json:"desiredRetention"`
	// DisableAutoplay: "Don't play audio automatically".
	DisableAutoplay bool `json:"disableAutoplay"`
	// SkipQuestionOnReplay: "Skip question when replaying answer".
	SkipQuestionOnReplay bool `json:"skipQuestionOnReplay"`
	// LearningSteps and RelearningSteps are step delays in seconds; empty
	// lets FSRS choose the short-term delays.
	LearningSteps   []int `json:"learningSteps"`
	RelearningSteps []int `json:"relearningSteps"`
	// InterdayLearningOrder is "mix", "before" or "after" (reviews).
	InterdayLearningOrder string `json:"interdayLearningOrder"`
	// BuryNew: "Bury new siblings".
	BuryNew bool `json:"buryNew"`
	// BuryReview: "Bury review siblings".
	BuryReview bool `json:"buryReview"`
	// BuryInterdayLearning: "Bury interday learning siblings".
	BuryInterdayLearning bool `json:"buryInterdayLearning"`
	// LeechThreshold is the lapse count that makes a card a leech (1–99).
	LeechThreshold int `json:"leechThreshold"`
	// LeechAction is "tag" (tag only) or "suspend" (suspend card and tag).
	LeechAction string `json:"leechAction"`
	// NewIgnoresReviewLimit: "New cards ignore review limit".
	NewIgnoresReviewLimit bool `json:"newIgnoresReviewLimit"`
	// LimitsStartFromTop: "Limits start from top".
	LimitsStartFromTop bool `json:"limitsStartFromTop"`
	// MaxInterval: "Maximum interval" in days (1–36500).
	MaxInterval int `json:"maxInterval"`
	// MinRelearnInterval: "Minimum interval" after relearning in days
	// (1–36500, not above MaxInterval).
	MinRelearnInterval int `json:"minRelearnInterval"`
	// NewGatherOrder is "deck", "deckRandomNotes", "ascending" (position),
	// "descending", "randomNotes" or "randomCards".
	NewGatherOrder string `json:"newGatherOrder"`
	// NewSortOrder is "gathered", "typeThenGathered", "typeThenRandom",
	// "randomNoteThenType" or "random".
	NewSortOrder string `json:"newSortOrder"`
	// NewReviewOrder is "mix", "before" or "after" (reviews).
	NewReviewOrder string `json:"newReviewOrder"`
	// ReviewSortOrder is "dueRandom", "dueDeck", "deckDue",
	// "intervalsAscending", "intervalsDescending", "difficultyAscending",
	// "difficultyDescending", "retrievabilityAscending" or
	// "relativeOverdueness".
	ReviewSortOrder string `json:"reviewSortOrder"`
	// InsertionOrder is "sequential" (order added) or "random".
	InsertionOrder string `json:"insertionOrder"`
	// MaxAnswerSeconds: "Maximum answer seconds" (1–9999), the cap on the
	// answer time recorded.
	MaxAnswerSeconds int `json:"maxAnswerSeconds"`
	// ShowTimer: "Show answer timer" on the study screen.
	ShowTimer bool `json:"showTimer"`
	// StopTimerOnAnswer: "Stop timer on answer".
	StopTimerOnAnswer bool `json:"stopTimerOnAnswer"`
	// SecondsToShowQuestion and SecondsToShowAnswer: "Seconds to show
	// question for" and "Seconds to show answer for" (0–9999) in auto
	// advance; 0 turns it off for that side.
	SecondsToShowQuestion int `json:"secondsToShowQuestion"`
	SecondsToShowAnswer   int `json:"secondsToShowAnswer"`
	// QuestionAction is "showAnswer" or "showReminder".
	QuestionAction string `json:"questionAction"`
	// AnswerAction is "buryCard", "answerAgain", "answerHard",
	// "answerGood" or "showReminder".
	AnswerAction string `json:"answerAction"`
	// Weights: "FSRS parameters", 19 numbers; empty uses the defaults.
	Weights []float64 `json:"weights"`
	// OptimizeSearch: the search selecting the reviews optimizing learns
	// from; empty means the deck and its subdecks.
	OptimizeSearch string `json:"optimizeSearch"`
	// IgnoreBefore: "Ignore cards reviewed before", a date like
	// "2026-01-31", or empty.
	IgnoreBefore string `json:"ignoreBefore"`
	// HistoricalRetention: "Historical retention" (0.50–0.99), the recall
	// assumed where a card's review history is missing.
	HistoricalRetention float64 `json:"historicalRetention"`
	// HealthCheck: "Health check" after optimizing.
	HealthCheck bool `json:"healthCheck"`
	// RescheduleOnChange: "Reschedule cards on change" of desired
	// retention or parameters.
	RescheduleOnChange bool `json:"rescheduleOnChange"`
	// EasyDays: "Easy Days", the level of each weekday, Monday first:
	// 0 normal, 1 reduced, 2 minimum.
	EasyDays [7]int `json:"easyDays"`
}

// Counts are due cards.
type Counts struct {
	New      int `json:"new"`
	Learning int `json:"learning"`
	Review   int `json:"review"`
}

func countsOf(c core.Counts) Counts {
	return Counts{New: c.New, Learning: c.Learning, Review: c.Review}
}

// DeckNode is one row of the deck tree.
type DeckNode struct {
	ID int64 `json:"id"`
	// Name is the deck's own name; FullName includes its parents.
	Name        string `json:"name"`
	FullName    string `json:"fullName"`
	Depth       int    `json:"depth"`
	ParentID    int64  `json:"parentId"`
	HasChildren bool   `json:"hasChildren"`
	Collapsed   bool   `json:"collapsed"`
	Counts      Counts `json:"counts"`
	// Cards is the number of cards in the deck and its subdecks.
	Cards   int         `json:"cards"`
	Options DeckOptions `json:"options"`
	// Filtered marks a filtered deck: it has no subdecks or options of its
	// own and cannot get new notes.
	Filtered bool `json:"filtered"`
}

// Tree returns all decks in display order (parents before children) with
// their due counts, which include subdecks.
func (s *DecksService) Tree() ([]DeckNode, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	decks := c.Decks()
	counts := c.DeckCounts()
	own := c.CardsPerDeck()
	byName := map[string]int64{}
	for _, d := range decks {
		byName[strings.ToLower(d.Name)] = int64(d.ID)
	}
	out := make([]DeckNode, 0, len(decks))
	index := map[int64]int{}
	for _, d := range decks {
		parts := strings.Split(d.Name, core.DeckSeparator)
		opts, err := c.DeckOptions(d.ID)
		if err != nil {
			return nil, err
		}
		n := DeckNode{
			ID: int64(d.ID), Name: parts[len(parts)-1], FullName: d.Name, Depth: len(parts) - 1,
			Collapsed: d.Collapsed, Counts: countsOf(counts[d.ID]),
			Options: DeckOptions(opts), Filtered: d.Filter != nil,
		}
		if len(parts) > 1 {
			n.ParentID = byName[strings.ToLower(strings.Join(parts[:len(parts)-1], core.DeckSeparator))]
			if i, ok := index[n.ParentID]; ok {
				out[i].HasChildren = true
			}
		}
		index[n.ID] = len(out)
		out = append(out, n)
	}
	// Card totals include subdecks: add each deck's own cards to it and all
	// its ancestors.
	for _, d := range decks {
		for id := int64(d.ID); id != 0; id = out[index[id]].ParentID {
			out[index[id]].Cards += own[d.ID]
		}
	}
	return out, nil
}

// NextChange returns when the due counts next change on their own: when a
// learning card becomes due, or when the next study day starts.
func (s *DecksService) NextChange() (time.Time, error) {
	c, err := s.a.collection()
	if err != nil {
		return time.Time{}, err
	}
	return c.NextCountChange(), nil
}

// Create creates a deck; "A::B" creates B inside A (and A when missing).
func (s *DecksService) Create(name string) (int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	id, err := c.CreateDeck(name)
	return int64(id), err
}

// Rename renames or moves a deck, with its subdecks.
func (s *DecksService) Rename(id int64, name string) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.RenameDeck(core.ID(id), name)
}

// DeletionInfo counts what deleting a deck removes, for the confirmation.
type DeletionInfo struct {
	Decks int `json:"decks"`
	// Cards are deleted; InFiltered of them are currently in filtered
	// decks elsewhere.
	Cards      int `json:"cards"`
	InFiltered int `json:"inFiltered"`
	// Returned are cards in filtered decks among the deleted decks, which
	// return to their home decks.
	Returned int `json:"returned"`
}

// DeletionInfo returns what Delete(id) would remove.
func (s *DecksService) DeletionInfo(id int64) (DeletionInfo, error) {
	c, err := s.a.collection()
	if err != nil {
		return DeletionInfo{}, err
	}
	d, err := c.DeckDeletionInfo(core.ID(id))
	return DeletionInfo(d), err
}

// Delete deletes a deck, its subdecks and the cards whose home deck they
// are; cards in a deleted filtered deck return to their home decks.
func (s *DecksService) Delete(id int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.DeleteDeck(core.ID(id))
}

// SetOptions changes the options of a deck's preset, and so of every deck
// using it; invalid values are rejected.
func (s *DecksService) SetOptions(id int64, options DeckOptions) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.SetDeckOptions(core.ID(id), core.DeckOptions(options))
}

// Preset is a named set of deck options shared by the decks using it.
type Preset struct {
	ID      int64       `json:"id"`
	Name    string      `json:"name"`
	Options DeckOptions `json:"options"`
	// Decks is the number of decks using the preset.
	Decks int `json:"decks"`
}

// DeckLimits are a deck's own daily limits, which override its preset's
// for that deck only; null uses the preset's.
type DeckLimits struct {
	NewPerDay     *int `json:"newPerDay"`
	ReviewsPerDay *int `json:"reviewsPerDay"`
}

// DeckSettings is what the deck options screen shows: the deck's preset,
// all presets, and the deck's own limits.
type DeckSettings struct {
	PresetID int64 `json:"presetId"`
	// Presets are sorted by name; Default cannot be deleted.
	Presets         []Preset   `json:"presets"`
	DefaultPresetID int64      `json:"defaultPresetId"`
	Limits          DeckLimits `json:"limits"`
}

// Settings returns the deck options screen's content for a deck.
func (s *DecksService) Settings(deckID int64) (DeckSettings, error) {
	c, err := s.a.collection()
	if err != nil {
		return DeckSettings{}, err
	}
	d, err := c.Deck(core.ID(deckID))
	if err != nil {
		return DeckSettings{}, err
	}
	usage := c.PresetUsage()
	out := DeckSettings{PresetID: int64(d.PresetID), DefaultPresetID: int64(core.DefaultPresetID),
		Limits: DeckLimits{NewPerDay: d.Limits.New, ReviewsPerDay: d.Limits.Review}}
	for _, p := range c.Presets() {
		out.Presets = append(out.Presets, Preset{ID: int64(p.ID), Name: p.Name, Options: DeckOptions(p.Options), Decks: usage[p.ID]})
	}
	return out, nil
}

// SaveSettings saves the deck options screen as one undo step: the deck
// uses the preset, which gets the options, and the deck gets its own
// limits. Invalid values are rejected and nothing changes.
func (s *DecksService) SaveSettings(deckID, presetID int64, options DeckOptions, limits DeckLimits) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.SaveDeckOptions(core.ID(deckID), core.ID(presetID), core.DeckOptions(options),
		core.DeckLimits{New: limits.NewPerDay, Review: limits.ReviewsPerDay})
}

// AddPreset creates a preset with the default options, used by the deck
// from now on. Names are unique, ignoring case.
func (s *DecksService) AddPreset(deckID int64, name string) (int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	id, err := c.AddPreset(core.ID(deckID), name)
	return int64(id), err
}

// ClonePreset creates a copy of the deck's preset, used by the deck from
// now on.
func (s *DecksService) ClonePreset(deckID int64, name string) (int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	id, err := c.ClonePreset(core.ID(deckID), name)
	return int64(id), err
}

// RenamePreset renames a preset.
func (s *DecksService) RenamePreset(id int64, name string) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.RenamePreset(core.ID(id), name)
}

// DeletePreset deletes a preset; the decks using it use Default instead.
func (s *DecksService) DeletePreset(id int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.DeletePreset(core.ID(id))
}

// SavePresetToSubdecks makes every subdeck of the deck use its preset.
func (s *DecksService) SavePresetToSubdecks(deckID int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.SavePresetToSubdecks(core.ID(deckID))
}

// SetCollapsed collapses or expands a parent deck in the list.
func (s *DecksService) SetCollapsed(id int64, collapsed bool) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.SetDeckCollapsed(core.ID(id), collapsed)
}

// CustomStudyInfo is what custom study can add to a deck today.
type CustomStudyInfo struct {
	// NewAvailable and ReviewAvailable are the cards beyond the deck's
	// current limits that increasing them would add.
	NewAvailable    int `json:"newAvailable"`
	ReviewAvailable int `json:"reviewAvailable"`
}

// CustomStudyInfo returns how many more new and review cards a deck has
// available today.
func (s *DecksService) CustomStudyInfo(deckID int64) (CustomStudyInfo, error) {
	c, err := s.a.collection()
	if err != nil {
		return CustomStudyInfo{}, err
	}
	x, err := c.ExtraCards(core.ID(deckID))
	return CustomStudyInfo{NewAvailable: x.New, ReviewAvailable: x.Review}, err
}

// IncreaseToday raises the deck's new and review limits for today only, by
// newBy and reviewBy cards; its options stay unchanged.
func (s *DecksService) IncreaseToday(deckID int64, newBy, reviewBy int) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.IncreaseTodayLimits(core.ID(deckID), newBy, reviewBy)
}

// DeckOverview is what a deck's overview shows before studying.
type DeckOverview struct {
	// Counts are today's due cards of the deck and its subdecks.
	Counts Counts `json:"counts"`
	// Buried is the number of buried cards (manually or by sibling
	// burying), which Unbury makes available again.
	Buried int `json:"buried"`
	// Description is the deck's description as entered.
	Description string `json:"description"`
	// DescriptionHTML is the description sanitized for display, without
	// scripts. Plain text keeps its line breaks.
	DescriptionHTML string `json:"descriptionHtml"`
	// NextDue is when something can next be studied, when nothing is due
	// now; nil when cards are due or the deck has nothing to study at all.
	NextDue *time.Time `json:"nextDue"`
}

// Overview returns a deck's overview: its due and buried cards, its
// description, and when nothing is due, when the next card is.
func (s *DecksService) Overview(deckID int64) (DeckOverview, error) {
	c, err := s.a.collection()
	if err != nil {
		return DeckOverview{}, err
	}
	d, err := c.Deck(core.ID(deckID))
	if err != nil {
		return DeckOverview{}, err
	}
	o, err := c.Overview(d.ID)
	if err != nil {
		return DeckOverview{}, err
	}
	out := DeckOverview{Counts: countsOf(o.Counts), Buried: o.Buried, Description: d.Description,
		DescriptionHTML: descriptionHTML(d.Description)}
	if o.Counts.Total() == 0 {
		if next := c.NextDue(d.ID); !next.IsZero() {
			out.NextDue = &next
		}
	}
	return out, nil
}

// descriptionHTML sanitizes a deck description for display. Text without
// tags is plain text: it is escaped and keeps its line breaks.
func descriptionHTML(desc string) string {
	if !strings.ContainsAny(desc, "<&") {
		desc = strings.ReplaceAll(html.EscapeString(desc), "\n", "<br>")
	}
	return render.SanitizeHTML(desc)
}

// SetDescription changes a deck's description (plain text or simple HTML).
func (s *DecksService) SetDescription(deckID int64, description string) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.SetDeckDescription(core.ID(deckID), description)
}

// Unbury makes every buried card of a deck and its subdecks available for
// study again, including cards held back today by sibling burying.
func (s *DecksService) Unbury(deckID int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.Unbury(core.ID(deckID))
}

// FilterTerm is one search of a filtered deck: up to Limit cards matching
// Search, picked and studied in Order: "oldestSeen", "random",
// "intervalsAscending", "intervalsDescending", "mostLapses", "added",
// "due", "latestAdded" or "retrievability".
type FilterTerm struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
	Order  string `json:"order"`
}

// FilteredDeckOptions are what the filtered deck options screen edits.
type FilteredDeckOptions struct {
	Name string `json:"name"`
	// Terms are the first filter and, when it is on, the second.
	Terms []FilterTerm `json:"terms"`
	// Reschedule: "Reschedule cards based on my answers in this deck".
	// Without it the deck previews its cards.
	Reschedule bool `json:"reschedule"`
	// PreviewDelays are the seconds before a card answered Again, Hard or
	// Good is shown again in preview mode; 0 returns it to its home deck.
	PreviewDelays []int `json:"previewDelays"`
	// SecondLimit and SecondOrder are what a second filter starts with.
	SecondLimit int    `json:"secondLimit"`
	SecondOrder string `json:"secondOrder"`
}

func filteredOptionsOf(name string, f core.FilterConfig) FilteredDeckOptions {
	out := FilteredDeckOptions{Name: name, Reschedule: f.Reschedule, PreviewDelays: f.PreviewDelays[:],
		SecondLimit: core.DefaultSecondFilterLimit, SecondOrder: core.FilterDue}
	for _, t := range f.Terms {
		out.Terms = append(out.Terms, FilterTerm(t))
	}
	return out
}

func (o FilteredDeckOptions) core() core.FilterConfig {
	cfg := core.FilterConfig{Reschedule: o.Reschedule}
	for _, t := range o.Terms {
		cfg.Terms = append(cfg.Terms, core.FilterTerm(t))
	}
	copy(cfg.PreviewDelays[:], o.PreviewDelays)
	return cfg
}

// NewFilteredOptions returns the options of a new filtered deck searching
// for search, named "Filtered Deck N" with the first free N.
func (s *DecksService) NewFilteredOptions(search string) (FilteredDeckOptions, error) {
	c, err := s.a.collection()
	if err != nil {
		return FilteredDeckOptions{}, err
	}
	taken := map[string]bool{}
	for _, d := range c.Decks() {
		taken[strings.ToLower(d.Name)] = true
	}
	name := ""
	for i := 1; name == "" || taken[strings.ToLower(name)]; i++ {
		name = fmt.Sprintf("Filtered Deck %d", i)
	}
	return filteredOptionsOf(name, core.DefaultFilterConfig(search)), nil
}

// FilteredOptions returns a filtered deck's options.
func (s *DecksService) FilteredOptions(deckID int64) (FilteredDeckOptions, error) {
	c, err := s.a.collection()
	if err != nil {
		return FilteredDeckOptions{}, err
	}
	d, err := c.Deck(core.ID(deckID))
	if err != nil {
		return FilteredDeckOptions{}, err
	}
	if d.Filter == nil {
		return FilteredDeckOptions{}, core.NewError("notFilteredDeck", "%q is not a filtered deck.", d.Name)
	}
	return filteredOptionsOf(d.Name, *d.Filter), nil
}

// FilteredBuild is the result of building a filtered deck.
type FilteredBuild struct {
	DeckID int64 `json:"deckId"`
	// Cards is the number of cards the deck holds.
	Cards int `json:"cards"`
}

// CreateFiltered creates a filtered deck and builds it, as one undo step.
// When no card matches, nothing is created and the error says so.
func (s *DecksService) CreateFiltered(options FilteredDeckOptions) (FilteredBuild, error) {
	c, err := s.a.collection()
	if err != nil {
		return FilteredBuild{}, err
	}
	id, n, err := c.CreateFilteredDeck(options.Name, options.core())
	return FilteredBuild{DeckID: int64(id), Cards: n}, err
}

// UpdateFiltered saves a filtered deck's options and rebuilds it, as one
// undo step. When no card matches, nothing changes.
func (s *DecksService) UpdateFiltered(deckID int64, options FilteredDeckOptions) (FilteredBuild, error) {
	c, err := s.a.collection()
	if err != nil {
		return FilteredBuild{}, err
	}
	n, err := c.UpdateFilteredDeck(core.ID(deckID), options.Name, options.core())
	return FilteredBuild{DeckID: deckID, Cards: n}, err
}

// Rebuild returns a filtered deck's cards home and gathers them again.
// When no card matches, nothing changes.
func (s *DecksService) Rebuild(deckID int64) (FilteredBuild, error) {
	c, err := s.a.collection()
	if err != nil {
		return FilteredBuild{}, err
	}
	n, err := c.RebuildFilteredDeck(core.ID(deckID))
	return FilteredBuild{DeckID: deckID, Cards: n}, err
}

// Empty returns a filtered deck's cards to their home decks.
func (s *DecksService) Empty(deckID int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.EmptyFilteredDeck(core.ID(deckID))
}

// CustomStudyChoice is a custom study session to create.
type CustomStudyChoice struct {
	// Kind is "forgotten" (rated Again in the last Days days), "ahead" (due
	// in the next Days days), "previewNew" (new cards added in the last
	// Days days) or "byState" (up to Cards cards of State, optionally with
	// one of Tags).
	Kind string `json:"kind"`
	Days int    `json:"days"`
	// State is "new", "due" or "all".
	State string   `json:"state"`
	Tags  []string `json:"tags"`
	Cards int      `json:"cards"`
}

// CustomStudy creates the filtered deck "Custom Study Session" for a
// choice on the deck and its subdecks, replacing the previous session.
// When no card matches, nothing changes and the error says so.
func (s *DecksService) CustomStudy(deckID int64, choice CustomStudyChoice) (FilteredBuild, error) {
	c, err := s.a.collection()
	if err != nil {
		return FilteredBuild{}, err
	}
	id, n, err := c.CustomStudy(core.ID(deckID), core.CustomStudy{Kind: choice.Kind, Days: choice.Days,
		State: choice.State, Tags: choice.Tags, Cards: choice.Cards})
	return FilteredBuild{DeckID: int64(id), Cards: n}, err
}

// OptimizeResult is what optimizing found.
type OptimizeResult struct {
	// Weights are the fitted FSRS parameters.
	Weights []float64 `json:"weights"`
	// Use reports whether they should replace the current parameters:
	// they predict the review history better and, when the health check
	// ran, pass it.
	Use bool `json:"use"`
	// Better reports whether they predict the history better than the
	// current parameters.
	Better bool `json:"better"`
	// Reviews is the number of reviews learned from.
	Reviews int `json:"reviews"`
	// LogLoss and RMSE measure how well Weights predict the reviews,
	// CurrentLogLoss and CurrentRMSE the current parameters (lower is
	// better).
	LogLoss        float64 `json:"logLoss"`
	RMSE           float64 `json:"rmse"`
	CurrentLogLoss float64 `json:"currentLogLoss"`
	CurrentRMSE    float64 `json:"currentRmse"`
	// HealthChecked reports whether the health check ran, HealthPassed
	// whether the parameters passed it.
	HealthChecked bool `json:"healthChecked"`
	HealthPassed  bool `json:"healthPassed"`
	// NotEnoughReviews is set when there were fewer than MinReviews
	// reviews; then only Reviews is set besides it.
	NotEnoughReviews bool `json:"notEnoughReviews"`
	MinReviews       int  `json:"minReviews"`
}

func optimizeResultOf(r scheduler.OptimizeResult, notEnough *scheduler.NotEnoughReviewsError) OptimizeResult {
	if notEnough != nil {
		return OptimizeResult{NotEnoughReviews: true, Reviews: notEnough.Reviews, MinReviews: scheduler.MinOptimizeReviews}
	}
	out := OptimizeResult{Weights: r.Weights, Use: r.Use(), Better: r.Better, Reviews: r.Reviews,
		LogLoss: r.LogLoss, RMSE: r.RMSE, CurrentLogLoss: r.CurrentLogLoss, CurrentRMSE: r.CurrentRMSE,
		MinReviews: scheduler.MinOptimizeReviews}
	if r.Health != nil {
		out.HealthChecked, out.HealthPassed = true, r.Health.Passed
	}
	return out
}

// PresetOptimization is the result of optimizing one preset.
type PresetOptimization struct {
	PresetID int64          `json:"presetId"`
	Name     string         `json:"name"`
	Result   OptimizeResult `json:"result"`
	// Applied reports whether the preset got the fitted parameters.
	Applied bool `json:"applied"`
}

// OptimizeEvent is progress, then the result or an error.
type OptimizeEvent struct {
	// Stage is "reading" (the review history), "training", "checking"
	// (the health check), "preset" (Optimize all: Done of Total presets
	// finished) or "done".
	Stage string `json:"stage"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
	// Result is the result of Optimize, Presets that of Optimize all.
	Result  *OptimizeResult      `json:"result"`
	Presets []PresetOptimization `json:"presets"`
	Error   string               `json:"error"`
}

// Optimize fits FSRS parameters to the review history selected by the
// options shown on the deck options screen (which may be unsaved): their
// optimize search, or the deck and its subdecks, their ignore date and
// historical retention, and their parameters as the current ones. It
// streams progress and ends with the result, which it does not save.
// Cancelling the stream stops it.
func (s *DecksService) Optimize(ctx context.Context, deckID int64, options DeckOptions) <-chan OptimizeEvent {
	return optimizeStream(ctx, func(progress core.Progress) (OptimizeEvent, error) {
		c, err := s.a.collection()
		if err != nil {
			return OptimizeEvent{}, err
		}
		r, err := c.Optimize(ctx, core.ID(deckID), core.DeckOptions(options), progress)
		var notEnough *scheduler.NotEnoughReviewsError
		if errors.As(err, &notEnough) {
			err = nil
		}
		if err != nil {
			return OptimizeEvent{}, err
		}
		res := optimizeResultOf(r, notEnough)
		return OptimizeEvent{Result: &res}, nil
	})
}

// OptimizeAll optimizes every preset with its saved options and gives the
// presets whose fitted parameters should be used (see OptimizeResult.Use)
// those parameters, as one undo step. Presets that reschedule on change
// reschedule their review cards (see RescheduleCountAll). It streams
// progress and ends with the results; cancelling the stream stops it and
// changes nothing.
func (s *DecksService) OptimizeAll(ctx context.Context) <-chan OptimizeEvent {
	return optimizeStream(ctx, func(progress core.Progress) (OptimizeEvent, error) {
		c, err := s.a.collection()
		if err != nil {
			return OptimizeEvent{}, err
		}
		results, err := c.OptimizeAll(ctx, progress)
		if err != nil {
			return OptimizeEvent{}, err
		}
		ev := OptimizeEvent{Presets: []PresetOptimization{}}
		for _, r := range results {
			ev.Presets = append(ev.Presets, PresetOptimization{PresetID: int64(r.Preset.ID), Name: r.Preset.Name,
				Result: optimizeResultOf(r.Result, r.NotEnough), Applied: r.Applied})
		}
		return ev, nil
	})
}

// optimizeStream runs an optimization in the background, streaming its
// progress and then its final event or the error.
func optimizeStream(ctx context.Context, run func(core.Progress) (OptimizeEvent, error)) <-chan OptimizeEvent {
	out := make(chan OptimizeEvent, 16)
	go func() {
		defer close(out)
		send := func(e OptimizeEvent) {
			select {
			case out <- e:
			case <-ctx.Done():
			}
		}
		ev, err := run(func(stage string, done, total int) {
			// Progress is best effort: never block training on a slow UI.
			select {
			case out <- OptimizeEvent{Stage: stage, Done: done, Total: total}:
			default:
			}
		})
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			send(OptimizeEvent{Error: err.Error()})
			return
		}
		ev.Stage, ev.Done, ev.Total = "done", 1, 1
		send(ev)
	}()
	return out
}

// RescheduleCount returns how many review cards saving the deck options
// screen with these values would reschedule (for the confirmation): 0
// unless they reschedule on change and change the desired retention or
// the FSRS parameters.
func (s *DecksService) RescheduleCount(deckID, presetID int64, options DeckOptions) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.RescheduleCount(core.ID(deckID), core.ID(presetID), core.DeckOptions(options))
}

// RescheduleCountAll returns how many review cards Optimize all may
// reschedule: those of decks whose preset reschedules on change.
func (s *DecksService) RescheduleCountAll() (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.RescheduleCountAll(), nil
}

// SimulationInput are the simulator's inputs besides the deck options.
type SimulationInput struct {
	// Days is the number of days to simulate (1–3650).
	Days          int `json:"days"`
	NewPerDay     int `json:"newPerDay"`
	ReviewsPerDay int `json:"reviewsPerDay"`
	// Retention is the desired retention to simulate; CompareRetention a
	// second one to compare with, or 0 for none.
	Retention        float64 `json:"retention"`
	CompareRetention float64 `json:"compareRetention"`
}

func (input SimulationInput) core() core.Simulation {
	return core.Simulation{Days: input.Days, NewPerDay: input.NewPerDay, ReviewsPerDay: input.ReviewsPerDay}
}

// SimulationSeries is the simulation at one desired retention, per day.
type SimulationSeries struct {
	Retention float64 `json:"retention"`
	// Reviews are the expected reviews of studied cards, New the new cards
	// learned, Seconds the time spent learning and reviewing, Memorized the
	// cards memorised at the end of the day.
	Reviews   []int     `json:"reviews"`
	New       []int     `json:"new"`
	Seconds   []float64 `json:"seconds"`
	Memorized []float64 `json:"memorized"`
	// The totals over the simulation.
	TotalReviews int     `json:"totalReviews"`
	TotalNew     int     `json:"totalNew"`
	TotalSeconds float64 `json:"totalSeconds"`
}

func seriesOf(retention float64, r scheduler.SimResult) SimulationSeries {
	out := SimulationSeries{Retention: retention, Reviews: r.Reviews, New: r.New, Seconds: r.Seconds,
		Memorized: r.Memorized, TotalSeconds: r.TotalSeconds()}
	for d := range r.Reviews {
		out.TotalReviews += r.Reviews[d]
		out.TotalNew += r.New[d]
	}
	return out
}

// RetentionPoint is the simulated cost at one desired retention: the
// review time per memorised card, in seconds.
type RetentionPoint struct {
	Retention      float64 `json:"retention"`
	SecondsPerCard float64 `json:"secondsPerCard"`
}

// RetentionSuggestion is the retention with the lowest cost, and the cost
// at every retention tried (in increasing retention).
type RetentionSuggestion struct {
	Retention float64          `json:"retention"`
	Points    []RetentionPoint `json:"points"`
}

// SimulationEvent is progress, then the result or an error.
type SimulationEvent struct {
	// Stage is "reading" (the deck and its review history), "simulating"
	// (Done of Total simulations finished) or "done".
	Stage string `json:"stage"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
	// Series is the result of Simulate: one per retention, the compared
	// one second. Suggestion is that of SuggestRetention.
	Series     []SimulationSeries   `json:"series"`
	Suggestion *RetentionSuggestion `json:"suggestion"`
	Error      string               `json:"error"`
}

// Simulate simulates the expected daily reviews and memorised cards of the
// deck and its subdecks, starting from their current cards, with the
// options shown on the deck options screen (which may be unsaved) and
// input: at input.Retention and, when set, input.CompareRetention. Answer times and
// rating probabilities come from the deck's review history. It changes
// nothing, and the same inputs and collection give the same result. It
// streams progress and ends with the result; cancelling the stream stops
// it.
func (s *DecksService) Simulate(ctx context.Context, deckID int64, options DeckOptions, input SimulationInput) <-chan SimulationEvent {
	return simulationStream(ctx, func(progress core.Progress) (SimulationEvent, error) {
		c, err := s.a.collection()
		if err != nil {
			return SimulationEvent{}, err
		}
		retentions := []float64{input.Retention}
		if input.CompareRetention != 0 {
			retentions = append(retentions, input.CompareRetention)
		}
		results, err := c.Simulate(ctx, core.ID(deckID), core.DeckOptions(options), input.core(), retentions, progress)
		if err != nil {
			return SimulationEvent{}, err
		}
		ev := SimulationEvent{}
		for i, r := range results {
			ev.Series = append(ev.Series, seriesOf(retentions[i], r))
		}
		return ev, nil
	})
}

// SuggestRetention simulates the deck as Simulate does at every desired
// retention from 0.70 to 0.99 and suggests the one with the lowest review
// time per memorised card (input.Retention and input.CompareRetention
// are ignored). It changes nothing: the screen asks before applying the
// suggestion. It streams progress and ends with the result; cancelling the
// stream stops it.
func (s *DecksService) SuggestRetention(ctx context.Context, deckID int64, options DeckOptions, input SimulationInput) <-chan SimulationEvent {
	return simulationStream(ctx, func(progress core.Progress) (SimulationEvent, error) {
		c, err := s.a.collection()
		if err != nil {
			return SimulationEvent{}, err
		}
		r, err := c.SuggestRetention(ctx, core.ID(deckID), core.DeckOptions(options), input.core(), progress)
		if err != nil {
			return SimulationEvent{}, err
		}
		sug := RetentionSuggestion{Retention: r.Retention, Points: []RetentionPoint{}}
		for _, p := range r.Points {
			sug.Points = append(sug.Points, RetentionPoint{Retention: p.Retention, SecondsPerCard: p.SecondsPerCard})
		}
		return SimulationEvent{Suggestion: &sug}, nil
	})
}

// simulationStream runs a simulation in the background, streaming its
// progress and then its final event or the error.
func simulationStream(ctx context.Context, run func(core.Progress) (SimulationEvent, error)) <-chan SimulationEvent {
	out := make(chan SimulationEvent, 16)
	go func() {
		defer close(out)
		ev, err := run(func(stage string, done, total int) {
			// Progress is best effort: never block the simulation on a slow UI.
			select {
			case out <- SimulationEvent{Stage: stage, Done: done, Total: total}:
			default:
			}
		})
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			ev = SimulationEvent{Error: err.Error()}
		} else {
			ev.Stage, ev.Done, ev.Total = "done", 1, 1
		}
		select {
		case out <- ev:
		case <-ctx.Done():
		}
	}()
	return out
}
