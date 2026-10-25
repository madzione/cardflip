// Package core is the flash card collection: decks, note types, notes,
// cards, settings and the review log, held in memory and persisted through
// storage as a snapshot plus a journal of ops. Every exported mutation
// validates first, then commits one op durably, then applies it in memory
// and announces the change, so a failed save changes nothing.
package core

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"time"

	"cardflip/golang/scheduler"
)

// Deck options defaults and limits (collection spec, "Deck options").
const (
	DefaultNewPerDay     = 20
	DefaultReviewsPerDay = 200
	DefaultRetention     = 0.90
	MinRetention         = 0.70
	MaxRetention         = 0.99
	// Historical retention (collection spec, "FSRS parameter options").
	DefaultHistoricalRetention = scheduler.DefaultHistoricalRetention
	MinHistoricalRetention     = 0.50
	MaxHistoricalRetention     = 0.99
	DefaultRolloverHour        = 4
	DefaultDeckName            = "Default"
	DeckSeparator              = "::"
)

// Deck is a deck. Nesting is expressed by the full name: "Languages::Spanish"
// is the child Spanish of the deck Languages.
type Deck struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
	// PresetID is the preset whose options the deck uses.
	PresetID ID `json:"preset"`
	// Limits override the preset's daily limits for this deck only.
	Limits    DeckLimits `json:"limits,omitzero"`
	Collapsed bool       `json:"collapsed,omitempty"`
	// Today overrides the daily limits for one study day.
	Today TodayLimits `json:"today,omitzero"`
	// Description is shown on the deck's overview: plain text or simple
	// HTML, sanitized for display.
	Description string `json:"desc,omitempty"`
	// UnburiedAt is when the deck was last unburied: answers before it no
	// longer bury their siblings in the deck. Only today's answers bury
	// siblings, so it needs no reset (design decision 2).
	UnburiedAt time.Time `json:"unburied,omitzero"`
	// Filter makes the deck a filtered deck, which holds cards from other
	// decks (their home decks) gathered by searches; nil for a regular
	// deck. A filtered deck has no subdecks and no options of its own: its
	// cards are studied with their home deck's options.
	Filter *FilterConfig `json:"filter,omitempty"`
	// legacy holds the options of a deck stored before presets existed
	// (snapshot version 1), until migratePresets turns them into presets.
	legacy *DeckOptions
}

// UnmarshalJSON reads a deck, keeping the options of decks stored before
// presets existed for migratePresets.
func (d *Deck) UnmarshalJSON(b []byte) error {
	type plain Deck
	var p struct {
		plain
		Options *DeckOptions `json:"options"`
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	*d = Deck(p.plain)
	if d.PresetID == 0 {
		d.legacy = p.Options
		if d.legacy == nil {
			o := legacyDeckOptions()
			d.legacy = &o
		}
	}
	return nil
}

// FilterConfig is how a filtered deck gathers its cards and how they are
// studied.
type FilterConfig struct {
	// Terms are the deck's searches: one, or two when the second filter is
	// on. The second term's cards come after the first's.
	Terms []FilterTerm `json:"terms"`
	// Reschedule makes answers schedule cards as normal reviews; when it is
	// off the deck previews its cards (see PreviewDelays).
	Reschedule bool `json:"resched"`
	// PreviewDelays are the seconds after which a card answered Again, Hard
	// or Good in preview mode is shown again; 0 returns it to its home deck.
	PreviewDelays [3]int `json:"previewDelays"`
}

// FilterTerm is one search of a filtered deck: up to Limit matching cards,
// chosen in Order.
type FilterTerm struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
	Order  string `json:"order"`
}

// Filtered deck orders: which matching cards a search picks first, and the
// order they are studied in.
const (
	FilterOldestSeen          = "oldestSeen"
	FilterRandom              = "random"
	FilterIntervalsAscending  = "intervalsAscending"
	FilterIntervalsDescending = "intervalsDescending"
	FilterMostLapses          = "mostLapses"
	FilterAdded               = "added"
	FilterDue                 = "due"
	FilterLatestAdded         = "latestAdded"
	FilterRetrievability      = "retrievability"
)

// Filtered deck defaults and limits (filtered-decks spec).
const (
	DefaultFilterLimit       = 100
	DefaultSecondFilterLimit = 20
	MaxFilterLimit           = 99999
	MaxPreviewDelay          = 24 * 3600
)

// DefaultFilterConfig returns the options of a new filtered deck searching
// for search.
func DefaultFilterConfig(search string) FilterConfig {
	return FilterConfig{
		Terms:         []FilterTerm{{Search: search, Limit: DefaultFilterLimit, Order: FilterRandom}},
		Reschedule:    true,
		PreviewDelays: [3]int{60, 600, 0},
	}
}

// DeckLimits are a deck's own new and review limits, which override its
// preset's for that deck only. Nil uses the preset's limit.
type DeckLimits struct {
	New    *int `json:"new,omitempty"`
	Review *int `json:"review,omitempty"`
}

// TodayLimits override a deck's new and review limits for the study day
// Day (scheduler.DayNumber); on other days they are ignored. Nil keeps the
// deck's daily limit.
type TodayLimits struct {
	Day    int  `json:"day"`
	New    *int `json:"new,omitempty"`
	Review *int `json:"review,omitempty"`
}

// Preset is a named set of deck options, shared by the decks that use it.
type Preset struct {
	ID      ID          `json:"id"`
	Name    string      `json:"name"`
	Options DeckOptions `json:"options"`
}

// DefaultPresetID is the ID of the preset every collection has; it cannot
// be deleted.
const DefaultPresetID ID = 1

// DefaultPresetName is the name of a new collection's preset.
const DefaultPresetName = "Default"

// DeckOptions are the study settings of a preset; a deck's children do not
// inherit them.
type DeckOptions struct {
	NewPerDay        int     `json:"newPerDay"`
	ReviewsPerDay    int     `json:"reviewsPerDay"`
	DesiredRetention float64 `json:"desiredRetention"`
	// DisableAutoplay stops a side's sounds from playing when it is shown;
	// they play only on replay.
	DisableAutoplay bool `json:"disableAutoplay,omitempty"`
	// SkipQuestionOnReplay makes replaying the answer play only the
	// answer's sounds, not the question's first.
	SkipQuestionOnReplay bool `json:"skipQuestionOnReplay,omitempty"`
	// LearningSteps and RelearningSteps are the (re)learning step delays in
	// seconds. Empty lists leave the delays to FSRS; absent from stored
	// options they are the defaults (see UnmarshalJSON).
	LearningSteps   []int `json:"learningSteps"`
	RelearningSteps []int `json:"relearningSteps"`
	// InterdayLearningOrder places interday learning cards relative to
	// review cards: InterdayMix, InterdayBefore or InterdayAfter.
	InterdayLearningOrder string `json:"interdayLearningOrder"`
	// BuryNew and BuryReview bury the new and review cards of notes
	// studied today until the next day (sibling burying). Off by default,
	// but absent from stored options they are on, as they were by default
	// before snapshot version 3 (see UnmarshalJSON).
	BuryNew    bool `json:"buryNew"`
	BuryReview bool `json:"buryReview"`
	// BuryInterdayLearning buries interday learning cards of notes studied
	// today until the next day, as new and review siblings are.
	BuryInterdayLearning bool `json:"buryInterdayLearning,omitempty"`
	// LeechThreshold is the lapse count at which a card becomes a leech;
	// it is detected again every half threshold (rounded up) after that.
	LeechThreshold int `json:"leechThreshold"`
	// LeechAction is what happens to a leech besides the tag: LeechTag or
	// LeechSuspend.
	LeechAction string `json:"leechAction"`
	// NewIgnoresReviewLimit shows new cards regardless of the review limit;
	// when off, new cards only fill the review limit's remaining room.
	// Absent from stored options it is on (see UnmarshalJSON).
	NewIgnoresReviewLimit bool `json:"newIgnoresReviewLimit"`
	// LimitsStartFromTop applies the limits of the parent decks too when
	// this deck is studied.
	LimitsStartFromTop bool `json:"limitsStartFromTop,omitempty"`
	// MaxInterval caps review intervals and MinRelearnInterval is the
	// shortest interval after relearning, both in days. Absent from stored
	// options they are the defaults (see UnmarshalJSON).
	MaxInterval        int `json:"maxInterval"`
	MinRelearnInterval int `json:"minRelearnInterval"`
	// NewGatherOrder picks which new cards fill the daily limit (Gather*);
	// NewSortOrder orders them within the session (Sort*).
	NewGatherOrder string `json:"newGatherOrder"`
	NewSortOrder   string `json:"newSortOrder"`
	// NewReviewOrder places new cards relative to review cards: NewMix,
	// NewBeforeReviews or NewAfterReviews.
	NewReviewOrder string `json:"newReviewOrder"`
	// ReviewSortOrder orders review cards, and so picks which fill the
	// review limit (Review*).
	ReviewSortOrder string `json:"reviewSortOrder"`
	// InsertionOrder is how cards added to the deck are positioned among
	// its new cards: InsertSequential or InsertRandom.
	InsertionOrder string `json:"insertionOrder"`
	// MaxAnswerSeconds caps the answer time recorded in the review history.
	// Absent from stored options it is the default (see UnmarshalJSON).
	MaxAnswerSeconds int `json:"maxAnswerSeconds"`
	// ShowTimer shows the time spent on the card on the study screen, and
	// StopTimerOnAnswer stops it when the answer is shown; they only affect
	// the display. Absent from stored options StopTimerOnAnswer is on (see
	// UnmarshalJSON).
	ShowTimer         bool `json:"showTimer,omitempty"`
	StopTimerOnAnswer bool `json:"stopTimerOnAnswer"`
	// SecondsToShowQuestion and SecondsToShowAnswer are how long auto
	// advance shows a side before running QuestionAction (Question*) or
	// AnswerAction (Answer*); 0 turns auto advance off for that side.
	// Absent from stored options the actions are the defaults (see
	// UnmarshalJSON).
	SecondsToShowQuestion int    `json:"secondsToShowQuestion,omitempty"`
	SecondsToShowAnswer   int    `json:"secondsToShowAnswer,omitempty"`
	QuestionAction        string `json:"questionAction"`
	AnswerAction          string `json:"answerAction"`
	// Weights are the FSRS parameters; empty uses the defaults. Stored
	// parameters of another model version are dropped (see UnmarshalJSON).
	Weights []float64 `json:"weights,omitempty"`
	// OptimizeSearch selects the cards whose reviews optimizing learns
	// from; empty means the deck being optimized and its subdecks.
	OptimizeSearch string `json:"optimizeSearch,omitempty"`
	// IgnoreBefore, a date like "2026-01-31" or empty, makes optimizing
	// ignore reviews before the start of that study day.
	IgnoreBefore string `json:"ignoreBefore,omitempty"`
	// HistoricalRetention is the recall assumed for reviews missing from a
	// card's history. Absent from stored options it is the default (see
	// UnmarshalJSON).
	HistoricalRetention float64 `json:"historicalRetention"`
	// HealthCheck checks fitted parameters before using them (scheduling
	// spec, "Optimize parameters").
	HealthCheck bool `json:"healthCheck,omitempty"`
	// RescheduleOnChange recomputes the due dates of the review cards when
	// the desired retention or the parameters change.
	RescheduleOnChange bool `json:"rescheduleOnChange,omitempty"`
	// EasyDays is the Easy Days level of each weekday, Monday first:
	// EasyNormal, EasyReduced or EasyMinimum.
	EasyDays [7]int `json:"easyDays,omitzero"`
}

// Easy Days levels of a weekday (DeckOptions.EasyDays).
const (
	EasyNormal  = scheduler.EasyNormal
	EasyReduced = scheduler.EasyReduced
	EasyMinimum = scheduler.EasyMinimum
)

// MaxAnswerTime returns the longest answer time recorded (see
// MaxAnswerSeconds).
func (o DeckOptions) MaxAnswerTime() time.Duration {
	return time.Duration(o.MaxAnswerSeconds) * time.Second
}

// Insertion orders.
const (
	InsertSequential = "sequential"
	InsertRandom     = "random"
)

// Interday learning/review orders.
const (
	InterdayMix    = "mix"
	InterdayBefore = "before"
	InterdayAfter  = "after"
)

// New card gather orders.
const (
	GatherDeck            = "deck"
	GatherDeckRandomNotes = "deckRandomNotes"
	GatherAscending       = "ascending"
	GatherDescending      = "descending"
	GatherRandomNotes     = "randomNotes"
	GatherRandomCards     = "randomCards"
)

// New card sort orders.
const (
	SortGathered           = "gathered"
	SortTypeThenGathered   = "typeThenGathered"
	SortTypeThenRandom     = "typeThenRandom"
	SortRandomNoteThenType = "randomNoteThenType"
	SortRandom             = "random"
)

// New/review orders.
const (
	NewMix           = "mix"
	NewBeforeReviews = "before"
	NewAfterReviews  = "after"
)

// Review sort orders.
const (
	ReviewDueRandom               = "dueRandom"
	ReviewDueDeck                 = "dueDeck"
	ReviewDeckDue                 = "deckDue"
	ReviewIntervalsAscending      = "intervalsAscending"
	ReviewIntervalsDescending     = "intervalsDescending"
	ReviewDifficultyAscending     = "difficultyAscending"
	ReviewDifficultyDescending    = "difficultyDescending"
	ReviewRetrievabilityAscending = "retrievabilityAscending"
	ReviewRelativeOverdueness     = "relativeOverdueness"
)

// Leech actions and limits (collection spec, "Leech options").
const (
	LeechTag              = "tag"
	LeechSuspend          = "suspend"
	DefaultLeechThreshold = 8
	MaxLeechThreshold     = 99
	// LeechTagName is the tag a leech's note gets.
	LeechTagName = "leech"
)

// Step limits (collection spec, "Step deck options").
const (
	MaxSteps       = 20
	MaxStepSeconds = 365 * 24 * 3600
)

// Interval limits (collection spec, "Interval limit options").
const (
	MaxIntervalDays = 36500
)

// Answer time limits (collection spec, "Timer deck options").
const (
	DefaultMaxAnswerSeconds = 60
	MaxMaxAnswerSeconds     = 9999
)

// Auto advance actions and limits (collection spec, "Auto advance deck
// options").
const (
	QuestionShowAnswer    = "showAnswer"
	QuestionShowReminder  = "showReminder"
	AnswerBuryCard        = "buryCard"
	AnswerAgain           = "answerAgain"
	AnswerHard            = "answerHard"
	AnswerGood            = "answerGood"
	AnswerShowReminder    = "showReminder"
	MaxAutoAdvanceSeconds = 9999
)

// DefaultDeckOptions returns the options of a new deck.
func DefaultDeckOptions() DeckOptions {
	return DeckOptions{
		NewPerDay: DefaultNewPerDay, ReviewsPerDay: DefaultReviewsPerDay, DesiredRetention: DefaultRetention,
		LearningSteps: []int{60, 600}, RelearningSteps: []int{600}, InterdayLearningOrder: InterdayMix,
		LeechThreshold: DefaultLeechThreshold, LeechAction: LeechTag, NewIgnoresReviewLimit: true,
		MaxInterval: MaxIntervalDays, MinRelearnInterval: 1,
		NewGatherOrder: GatherAscending, NewSortOrder: SortGathered, NewReviewOrder: NewMix,
		ReviewSortOrder: ReviewDueRandom, InsertionOrder: InsertSequential,
		MaxAnswerSeconds: DefaultMaxAnswerSeconds, StopTimerOnAnswer: true,
		QuestionAction: QuestionShowAnswer, AnswerAction: AnswerBuryCard,
		HistoricalRetention: DefaultHistoricalRetention,
	}
}

// legacyDeckOptions returns the default options before snapshot version 3,
// which buried new and review siblings.
func legacyDeckOptions() DeckOptions {
	o := DefaultDeckOptions()
	o.BuryNew, o.BuryReview = true, true
	return o
}

// UnmarshalJSON fills options missing from stored data (written before
// they existed) with their defaults, and the burying options with their
// defaults before snapshot version 3 (see legacyDeckOptions).
func (o *DeckOptions) UnmarshalJSON(b []byte) error {
	type plain DeckOptions
	p := plain(legacyDeckOptions())
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	*o = DeckOptions(p)
	if len(o.Weights) != scheduler.NumWeights {
		o.Weights = nil
	}
	return nil
}

// Equal reports whether two sets of options are the same.
func (o DeckOptions) Equal(other DeckOptions) bool {
	a, b := o, other
	a.LearningSteps, a.RelearningSteps, b.LearningSteps, b.RelearningSteps = nil, nil, nil, nil
	a.Weights, b.Weights = nil, nil
	return reflect.DeepEqual(a, b) && slices.Equal(o.LearningSteps, other.LearningSteps) &&
		slices.Equal(o.RelearningSteps, other.RelearningSteps) && slices.Equal(o.Weights, other.Weights)
}

// NoteKind distinguishes standard note types from cloze note types.
type NoteKind string

const (
	Standard NoteKind = "standard"
	Cloze    NoteKind = "cloze"
)

// NoteType defines the fields of its notes and how their cards look.
type NoteType struct {
	ID     ID       `json:"id"`
	Name   string   `json:"name"`
	Kind   NoteKind `json:"kind"`
	Fields []string `json:"fields"`
	// FieldOptions holds per-field options, parallel to Fields. Missing
	// entries mean the defaults, so it is empty while all fields have them.
	FieldOptions []FieldOptions `json:"fieldOptions,omitempty"`
	// SortField is the index of the field the browser's Sort field column
	// shows and sorts by; the first field by default.
	SortField int            `json:"sortField,omitempty"`
	Templates []CardTemplate `json:"templates"`
	CSS       string         `json:"css"`
	// Source identifies where an imported note type came from (e.g. its
	// Anki ID), so importing the same package again reuses it.
	Source string `json:"source,omitempty"`
	// Modified is when the note type last changed; an imported one keeps
	// the package's time. Note types stored before it existed have none,
	// which counts as older than any time.
	Modified time.Time `json:"mod,omitzero"`
	// LatexSVG is set when the note's LaTeX images are SVG instead of PNG,
	// as in Anki, which names them by this setting.
	LatexSVG bool `json:"latexSvg,omitempty"`
}

// FieldOptions are the options of one note type field.
type FieldOptions struct {
	// HTMLByDefault opens the field in the HTML source editor instead of
	// the visual editor.
	HTMLByDefault bool `json:"htmlByDefault"`
	// Sticky keeps the field's content in the add screen after a note is
	// added.
	Sticky bool `json:"sticky"`
	// Font and Size are the font family and size (in logical pixels) the
	// editor shows the field in; empty and 0 mean the editor's defaults.
	// Cards are not affected.
	Font string `json:"font"`
	Size int    `json:"size"`
	// RTL lays the field out right to left in the editor.
	RTL bool `json:"rtl"`
	// Collapsed starts the field collapsed in the editor.
	Collapsed bool `json:"collapsed"`
	// ExcludeFromSearch leaves the field out of unqualified searches;
	// searches naming the field still look in it.
	ExcludeFromSearch bool `json:"excludeFromSearch"`
}

// cardType returns the card type that generates cards with ord, or nil.
func (nt *NoteType) cardType(ord int) *CardTemplate {
	if nt.Kind == Cloze {
		ord = 0
	}
	if ord < 0 || ord >= len(nt.Templates) {
		return nil
	}
	return &nt.Templates[ord]
}

// CardTemplate is one card a standard note type generates (cloze note types
// have exactly one, used for every cloze number).
type CardTemplate struct {
	Name string `json:"name"`
	QFmt string `json:"qfmt"`
	AFmt string `json:"afmt"`
	// DeckOverride, when set, is the deck that new cards of this card type
	// are added to instead of the note's deck. It is ignored while that
	// deck is missing or filtered.
	DeckOverride ID `json:"did,omitempty"`
	// BrowserQFmt and BrowserAFmt are simplified templates for the
	// browser's question and answer text; empty means QFmt and AFmt.
	BrowserQFmt string `json:"bqfmt,omitempty"`
	BrowserAFmt string `json:"bafmt,omitempty"`
}

// Note holds one value per field of its note type, plus tags.
type Note struct {
	ID       ID        `json:"id"`
	TypeID   ID        `json:"type"`
	GUID     string    `json:"guid"`
	Fields   []string  `json:"fields"`
	Tags     []string  `json:"tags,omitempty"`
	Modified time.Time `json:"mod"`
}

// Card is one thing to study, generated from a note: template Ord of a
// standard note type, or cloze number Ord+1 of a cloze note type.
type Card struct {
	ID     ID  `json:"id"`
	NoteID ID  `json:"note"`
	DeckID ID  `json:"deck"`
	Ord    int `json:"ord"`
	// NewPos orders New cards: they are introduced in increasing NewPos
	// (ties by ID). Cards get the next free position when they are added
	// (see newPos), or a random one in decks with random insertion order;
	// Reposition changes it.
	NewPos int64 `json:"pos,omitempty"`
	// OriginalPos is the card's NewPos when it was first answered (or, for
	// an imported card, its place among new cards where it came from), so
	// that a reset can restore it; nil while unknown. It is set once.
	OriginalPos *int64    `json:"opos,omitempty"`
	Suspended   bool      `json:"susp,omitempty"`
	BuriedUntil time.Time `json:"buried,omitzero"`
	// Flag is 0 for no flag, or one of the seven flags 1 (red) to 7
	// (purple).
	Flag int `json:"flag,omitempty"`
	// Modified is when the card last changed (see stampModified); zero for
	// cards stored before it was recorded.
	Modified time.Time `json:"mod,omitzero"`
	// HomeDeckID is the deck a card in a filtered deck (DeckID) came from
	// and returns to; 0 while the card is in its home deck. The home deck's
	// options schedule the card.
	HomeDeckID ID `json:"home,omitempty"`
	// FilteredPos is the card's place in its filtered deck's order, set when
	// the deck is built.
	FilteredPos int `json:"fpos,omitempty"`
	// PreviewDue is when a card answered Again or Hard in a filtered deck
	// that previews its cards is shown again; zero until then.
	PreviewDue time.Time `json:"pdue,omitzero"`
	scheduler.CardState
}

// HomeDeck returns the card's home deck: the deck it is in, or the one it
// returns to from a filtered deck.
func (c *Card) HomeDeck() ID {
	if c.HomeDeckID != 0 {
		return c.HomeDeckID
	}
	return c.DeckID
}

// returnHome returns a card in a filtered deck to its home deck.
func (c *Card) returnHome() {
	c.DeckID = c.HomeDeck()
	c.HomeDeckID, c.FilteredPos, c.PreviewDue = 0, 0, time.Time{}
}

// FilteredPlace is where a card was in a filtered deck before an answer,
// so that undoing the answer puts it back.
type FilteredPlace struct {
	Deck       ID        `json:"deck"`
	Home       ID        `json:"home"`
	Pos        int       `json:"pos,omitempty"`
	PreviewDue time.Time `json:"pdue,omitzero"`
}

// Theme modes.
const (
	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"
)

// Settings are app-wide settings. They live in the collection, so backups
// include them.
type Settings struct {
	Theme string `json:"theme"`
	// Language is the interface language as a language code, such as "de";
	// empty follows the device's languages.
	Language     string `json:"language,omitempty"`
	RolloverHour int    `json:"rolloverHour"`
	// LastNoteTypeID and LastDeckID remember the editor's pickers.
	LastNoteTypeID ID `json:"lastNoteType,omitempty"`
	LastDeckID     ID `json:"lastDeck,omitempty"`
	// LastDeckByNoteType remembers the deck of the last added note of each
	// note type. Settings values share it, so it is replaced, never changed
	// in place.
	LastDeckByNoteType map[ID]ID `json:"lastDeckByType,omitempty"`
	// DeckByNoteType makes the add screen pick the deck last used with the
	// chosen note type instead of the last used deck.
	DeckByNoteType bool `json:"deckByNoteType,omitempty"`
	// CurrentDeckID is the deck most recently studied, which deck:current
	// searches.
	CurrentDeckID ID `json:"currentDeck,omitempty"`
	// RecentDecks are the decks most recently studied, latest first, at
	// most MaxRecentDecks. They may include deleted decks; RecentDecks
	// leaves those out. Settings values share it, so it is replaced, never
	// changed in place.
	RecentDecks []ID `json:"recentDecks,omitempty"`
	// PersistenceNoticeShown records the one-time web storage notice.
	PersistenceNoticeShown bool `json:"persistNotice,omitempty"`
	// AutoContinueLearning makes a waiting study session show learning
	// cards as soon as they are due, instead of asking first.
	AutoContinueLearning bool `json:"autoContinueLearning,omitempty"`
	// LearnAheadMinutes is how far ahead of their due time a study session
	// shows learning cards when nothing else is due; nil means
	// DefaultLearnAheadMinutes, 0 shows them only once due.
	LearnAheadMinutes *int `json:"learnAheadMinutes,omitempty"`
	// HidePlayButtons hides the replay buttons on cards with audio.
	HidePlayButtons bool `json:"hidePlayButtons,omitempty"`
	// KeepAudioOnAnswer lets the question's sounds finish when the answer
	// is shown, instead of interrupting them.
	KeepAudioOnAnswer bool `json:"keepAudioOnAnswer,omitempty"`
	// FlagNames are the custom names of flags 1 to 7; an empty name means
	// the flag's colour name.
	FlagNames [MaxFlag]string `json:"flagNames,omitzero"`
	// StripPasteFormatting makes a normal paste into a field insert plain
	// text, and the plain-paste shortcut keep formatting.
	StripPasteFormatting bool `json:"stripPasteFormatting,omitempty"`
	// PasteImagesAsPng stores pasted images as PNG instead of JPEG.
	PasteImagesAsPng bool `json:"pasteImagesAsPng,omitempty"`
	// IgnoreAccents makes text and field searches ignore accents, as nc:
	// does.
	IgnoreAccents bool `json:"ignoreAccents,omitempty"`
	// Backup configures automatic backups.
	Backup BackupSettings `json:"backup"`
	// BrowserColumns are the browser's columns in order; empty means
	// DefaultColumns. Settings values share it, so it is replaced, never
	// changed in place.
	BrowserColumns []Column `json:"browserColumns,omitempty"`
	// BrowserMode is BrowserCards or BrowserNotes; empty means
	// BrowserCards.
	BrowserMode string `json:"browserMode,omitempty"`
	// BrowserSplit is the share of the width that the browser's list takes
	// next to the details on wide screens; 0 means DefaultBrowserSplit.
	BrowserSplit float64 `json:"browserSplit,omitempty"`
	// SavedSearches are the browser's saved searches, sorted by name.
	// Settings values share it, so it is replaced, never changed in place.
	SavedSearches []SavedSearch `json:"savedSearches,omitempty"`
	// DefaultSearch is the search the browser starts with; empty lists all
	// cards.
	DefaultSearch string `json:"defaultSearch,omitempty"`
	// LastDueDate is the last input of Set due date, which its dialog
	// starts with.
	LastDueDate string `json:"lastDueDate,omitempty"`
	// TimeboxMinutes is the length of a study timebox, after which the
	// study screen reports the cards studied in it; 0 turns it off.
	TimeboxMinutes int `json:"timeboxMinutes,omitempty"`
	// HideRemainingCounts hides the remaining new, learning and review
	// counts on the study screen.
	HideRemainingCounts bool `json:"hideRemainingCounts,omitempty"`
	// HideIntervals hides the next intervals on the answer buttons.
	HideIntervals bool `json:"hideIntervals,omitempty"`
	// SpaceOnlyReveals makes Space and Enter only reveal the answer instead
	// of also choosing Good.
	SpaceOnlyReveals bool `json:"spaceOnlyReveals,omitempty"`
	// HideBarsDuringReview hides the app bar and navigation while studying.
	HideBarsDuringReview bool `json:"hideBarsDuringReview,omitempty"`
	// QuestionFadeMs is how long each new question fades in, in
	// milliseconds; nil means DefaultQuestionFadeMs, 0 shows it at once.
	QuestionFadeMs *int `json:"questionFadeMs,omitempty"`
	// UIScale is the size of the interface in percent; 0 means
	// DefaultUIScale.
	UIScale int `json:"uiScale,omitempty"`
	// Minimalist shows the interface without shadows, elevation tints and
	// decorative fills.
	Minimalist bool `json:"minimalist,omitempty"`
	// ReduceMotion is ReduceMotionOn or ReduceMotionOff; empty follows the
	// operating system.
	ReduceMotion string `json:"reduceMotion,omitempty"`
	// DailyGoalKind is GoalCards or GoalMinutes; empty sets no daily goal.
	// DailyGoal is the number of cards or minutes.
	DailyGoalKind string `json:"dailyGoalKind,omitempty"`
	DailyGoal     int    `json:"dailyGoal,omitempty"`
	// RestDaysPerWeek is how many days without study, from 0 to
	// MaxRestDaysPerWeek, a streak survives in any 7 days in a row.
	RestDaysPerWeek int `json:"restDaysPerWeek,omitempty"`
	// HideProgress hides the streak and goal on the deck list and the study
	// summary; stored inverted so that they show by default.
	HideProgress bool `json:"hideProgress,omitempty"`
	// GoalSuggestionShown records the one-time suggestion to set a daily
	// goal.
	GoalSuggestionShown bool `json:"goalSuggestion,omitempty"`
	// SeenAchievements are the earned achievement tiers this device has
	// told the user about ("<id>:<threshold>"). Nil until achievements
	// were first evaluated, when everything earned until then is marked
	// seen; it is saved even when empty to tell the two apart.
	SeenAchievements []string `json:"seenAchievements"`
}

// UI size limits and default, in percent. The size changes in steps of
// UIScaleStep.
const (
	MinUIScale     = 75
	MaxUIScale     = 200
	DefaultUIScale = 100
	UIScaleStep    = 5
)

// Reduce motion choices.
const (
	ReduceMotionSystem = "system"
	ReduceMotionOn     = "on"
	ReduceMotionOff    = "off"
)

// Daily goal kinds.
const (
	GoalCards   = "cards"
	GoalMinutes = "minutes"
)

// MaxDailyGoal is the largest daily goal, in cards or minutes.
const MaxDailyGoal = 9999

// MaxRestDaysPerWeek is the most rest days per week a streak can allow.
const MaxRestDaysPerWeek = 2

// MaxRecentDecks is how many recently studied decks Settings.RecentDecks
// keeps.
const MaxRecentDecks = 10

// MaxTimeboxMinutes is the longest timebox.
const MaxTimeboxMinutes = 24 * 60

// Question fade default and maximum, in milliseconds.
const (
	DefaultQuestionFadeMs = 100
	MaxQuestionFadeMs     = 1000
)

// QuestionFade returns the question fade time (see QuestionFadeMs).
func (s Settings) QuestionFade() time.Duration {
	if s.QuestionFadeMs == nil {
		return DefaultQuestionFadeMs * time.Millisecond
	}
	return time.Duration(*s.QuestionFadeMs) * time.Millisecond
}

// SavedSearch is a browser search saved under a name.
type SavedSearch struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}

// Browser modes: one row per card, or one row per note.
const (
	BrowserCards = "cards"
	BrowserNotes = "notes"
)

// DefaultBrowserSplit is the share of the width the browser's list takes
// next to the details until the user drags the divider.
const DefaultBrowserSplit = 0.6

// BackupSettings configure automatic backups.
type BackupSettings struct {
	// IntervalMinutes is the time between periodic backups; 0 turns them
	// off. Backups before risky operations are made regardless.
	IntervalMinutes int `json:"intervalMinutes"`
	// Daily, Weekly and Monthly are how many backups to keep besides
	// today's: the latest of each of that many days, weeks and months.
	Daily   int `json:"daily"`
	Weekly  int `json:"weekly"`
	Monthly int `json:"monthly"`
}

// Backup settings defaults and limits.
const (
	DefaultBackupInterval = 30
	DefaultBackupsKept    = 8
	MaxBackupInterval     = 24 * 60
	MaxBackupsKept        = 100
)

// DefaultBackupSettings returns the backup settings of a new collection.
func DefaultBackupSettings() BackupSettings {
	return BackupSettings{IntervalMinutes: DefaultBackupInterval,
		Daily: DefaultBackupsKept, Weekly: DefaultBackupsKept, Monthly: DefaultBackupsKept}
}

// Learn ahead limit default and maximum, in minutes.
const (
	DefaultLearnAheadMinutes = 20
	MaxLearnAheadMinutes     = 24 * 60
)

// LearnAhead returns the learn ahead limit (see LearnAheadMinutes).
func (s Settings) LearnAhead() time.Duration {
	if s.LearnAheadMinutes == nil {
		return DefaultLearnAheadMinutes * time.Minute
	}
	return time.Duration(*s.LearnAheadMinutes) * time.Minute
}

// DefaultSettings returns the settings of a new collection.
func DefaultSettings() Settings {
	return Settings{Theme: ThemeSystem, RolloverHour: DefaultRolloverHour, Backup: DefaultBackupSettings()}
}

// equal reports whether s and o are the same settings. An empty and a
// missing LastDeckByNoteType are the same.
func (s Settings) equal(o Settings) bool {
	if !maps.Equal(s.LastDeckByNoteType, o.LastDeckByNoteType) {
		return false
	}
	s.LastDeckByNoteType, o.LastDeckByNoteType = nil, nil
	return reflect.DeepEqual(s, o)
}

// UnmarshalJSON reads settings, giving settings stored before they existed
// their defaults.
func (s *Settings) UnmarshalJSON(data []byte) error {
	type plain Settings
	p := plain{Backup: DefaultBackupSettings()}
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	*s = Settings(p)
	return nil
}

// ReviewEntry is one answer in the review log, or a manual entry (rating 0)
// for a change of scheduling. Before holds the card's complete scheduling
// state before the answer, so undo restores it exactly; it is empty for
// imported entries, which cannot be undone.
type ReviewEntry struct {
	ID          ID                  `json:"id"`
	CardID      ID                  `json:"card"`
	Time        time.Time           `json:"time"`
	Rating      scheduler.Rating    `json:"rating"`
	StateBefore scheduler.State     `json:"stateBefore"`
	IvlBefore   int                 `json:"ivlBefore"`
	IvlAfter    int                 `json:"ivlAfter"`
	DurationMs  int64               `json:"ms"`
	Before      scheduler.CardState `json:"before,omitzero"`
	// Source identifies an imported entry where it came from, e.g.
	// "anki:<revlog id>".
	Source string `json:"src,omitempty"`
	// Leech is set when the answer made the card a leech. LeechTagged and
	// LeechSuspended record what it changed besides the card's scheduling
	// (the note got the leech tag, the card was suspended), so undo reverts
	// exactly that.
	Leech          bool `json:"leech,omitempty"`
	LeechTagged    bool `json:"leechTag,omitempty"`
	LeechSuspended bool `json:"leechSusp,omitempty"`
	// Preview marks an answer in a filtered deck that previews its cards:
	// it did not change the card's scheduling, so it counts as no review
	// in the queue, searches and statistics.
	Preview bool `json:"preview,omitempty"`
	// Filtered is where the card was when it was answered in a filtered
	// deck.
	Filtered *FilteredPlace `json:"filtered,omitempty"`
}

// state is the snapshot: the whole collection except the review log and
// media.
type state struct {
	Version   int         `json:"version"`
	NextID    int64       `json:"nextId"`
	Settings  Settings    `json:"settings"`
	Presets   []*Preset   `json:"presets"`
	Decks     []*Deck     `json:"decks"`
	NoteTypes []*NoteType `json:"noteTypes"`
	Notes     []*Note     `json:"notes"`
	Cards     []*Card     `json:"cards"`
	// Normalized is set once the collection's text and media names are in
	// NFC (see migrateNFC); snapshots of earlier versions lack it.
	Normalized bool `json:"normalized,omitempty"`
	// Sync metadata (snapshot version 4, see sync_meta.go): the collection's
	// lineage, its change counter, the registers' and tombstones' stamps,
	// the review log's sequence numbers by month and the slot table.
	Epoch  string           `json:"epoch,omitempty"`
	Seq    int64            `json:"seq,omitempty"`
	Meta   metaTable        `json:"meta,omitempty"`
	Tombs  metaTable        `json:"tombs,omitempty"`
	RevSeq map[string]int64 `json:"revSeq,omitempty"`
	Slots  map[string]int   `json:"slots,omitempty"`
	// Media is the media manifest, once MediaBuilt.
	Media      map[string]mediaEntry `json:"media,omitempty"`
	MediaBuilt bool                  `json:"mediaBuilt,omitempty"`
}

// snapshotVersion 2 moved deck options into presets (migratePresets); 3
// turned sibling burying off in the Default preset (unburyDefault); 4 added
// the sync metadata (ensureMeta).
const snapshotVersion = 4
