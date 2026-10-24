package core

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

// testCollection opens a fresh collection on an in-memory store. The clock
// starts at 2026-10-05 12:00 local time.
func testCollection(t *testing.T) (*Collection, *storage.MemoryStore, *clock) {
	t.Helper()
	store := storage.NewMemoryStore()
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	c, fresh, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil || !fresh {
		t.Fatalf("open: %v fresh=%v", err, fresh)
	}
	return c, store, clk
}

// reopen opens the collection again from the same store, as after a restart,
// and checks that it matches c exactly.
func reopen(t *testing.T, c *Collection, store *storage.MemoryStore, clk *clock) *Collection {
	t.Helper()
	c2, fresh, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil || fresh {
		t.Fatalf("reopen: %v fresh=%v", err, fresh)
	}
	c.mu.RLock()
	want := c.state()
	c.mu.RUnlock()
	if got := c2.state(); !reflect.DeepEqual(normalizeState(got), normalizeState(want)) {
		t.Fatalf("reopened collection differs:\n%s", jsonDiff(normalizeState(got), normalizeState(want)))
	}
	return c2
}

// normalizeState makes times comparable (monotonic clock readings and
// locations differ after a JSON round trip).
func normalizeState(s *state) *state {
	cp := *s
	cp.Notes = nil
	for _, n := range s.Notes {
		nn := *n
		nn.Modified = nn.Modified.UTC().Round(0)
		cp.Notes = append(cp.Notes, &nn)
	}
	cp.Cards = nil
	for _, card := range s.Cards {
		cc := *card
		cc.Due = cc.Due.UTC().Round(0)
		cc.LastReview = cc.LastReview.UTC().Round(0)
		cc.BuriedUntil = cc.BuriedUntil.UTC().Round(0)
		cp.Cards = append(cp.Cards, &cc)
	}
	cp.Decks = nil
	for _, d := range s.Decks {
		dd := *d
		dd.UnburiedAt = dd.UnburiedAt.UTC().Round(0)
		cp.Decks = append(cp.Decks, &dd)
	}
	return &cp
}

func deckNames(c *Collection) string {
	var names []string
	for _, d := range c.Decks() {
		names = append(names, d.Name)
	}
	return strings.Join(names, ", ")
}

func mustUserErr(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil || !IsUserError(err) || !strings.Contains(err.Error(), contains) {
		t.Fatalf("err = %v, want user error containing %q", err, contains)
	}
}

func TestFreshCollection(t *testing.T) {
	c, _, _ := testCollection(t)
	// Fresh collection (spec): exactly one deck named Default.
	if got := deckNames(c); got != "Default" {
		t.Fatalf("decks = %q", got)
	}
	// Built-ins available (spec).
	var names []string
	for _, nt := range c.NoteTypes() {
		names = append(names, nt.Name)
	}
	if strings.Join(names, ", ") != "Basic, Basic (and reversed card), Cloze, Image Occlusion" {
		t.Fatalf("note types = %v", names)
	}
	for _, nt := range c.NoteTypes() {
		switch nt.Name {
		case "Basic":
			if strings.Join(nt.Fields, ",") != "Front,Back" || len(nt.Templates) != 1 || nt.Kind != Standard {
				t.Errorf("Basic = %+v", nt)
			}
		case "Basic (and reversed card)":
			if strings.Join(nt.Fields, ",") != "Front,Back" || len(nt.Templates) != 2 {
				t.Errorf("reversed = %+v", nt)
			}
		case "Cloze":
			if strings.Join(nt.Fields, ",") != "Text,Back Extra" || len(nt.Templates) != 1 || nt.Kind != Cloze {
				t.Errorf("Cloze = %+v", nt)
			}
		}
	}
	if s := c.Settings(); s.Theme != ThemeSystem || s.RolloverHour != 4 {
		t.Fatalf("settings = %+v", s)
	}
	o := deckOptions(c, c.Decks()[0].ID)
	if want := (DeckOptions{NewPerDay: 20, ReviewsPerDay: 200, DesiredRetention: 0.9,
		LearningSteps: []int{60, 600}, RelearningSteps: []int{600}, InterdayLearningOrder: InterdayMix,
		LeechThreshold: 8, LeechAction: LeechTag, NewIgnoresReviewLimit: true,
		MaxInterval: 36500, MinRelearnInterval: 1, NewGatherOrder: GatherAscending, NewSortOrder: SortGathered,
		NewReviewOrder: NewMix, ReviewSortOrder: ReviewDueRandom, InsertionOrder: InsertSequential,
		MaxAnswerSeconds: 60, StopTimerOnAnswer: true,
		QuestionAction: QuestionShowAnswer, AnswerAction: AnswerBuryCard,
		HistoricalRetention: 0.9}); !o.Equal(want) {
		t.Fatalf("default options = %+v", o)
	}
	// A deck never optimized has no FSRS parameters: the defaults are used.
	if len(o.Weights) != 0 {
		t.Fatalf("default FSRS parameters = %v", o.Weights)
	}
}

func TestCreateNestedDeck(t *testing.T) {
	c, store, clk := testCollection(t)
	// Create nested deck (spec): the missing parent is created.
	id, err := c.CreateDeck("Languages::Spanish")
	if err != nil {
		t.Fatal(err)
	}
	if got := deckNames(c); got != "Default, Languages, Languages::Spanish" {
		t.Fatalf("decks = %q", got)
	}
	if d, _ := c.Deck(id); d.Name != "Languages::Spanish" {
		t.Fatalf("created %q", d.Name)
	}
	// Existing parents are matched ignoring case; parts are trimmed.
	if _, err := c.CreateDeck(" languages ::  French "); err != nil {
		t.Fatal(err)
	}
	if got := deckNames(c); got != "Default, Languages, Languages::French, Languages::Spanish" {
		t.Fatalf("decks = %q", got)
	}
	_, err = c.CreateDeck("LANGUAGES::spanish")
	mustUserErr(t, err, "already exists")
	_, err = c.CreateDeck("A::::B")
	mustUserErr(t, err, "empty part")
	_, err = c.CreateDeck("   ")
	mustUserErr(t, err, "enter a deck name")
	reopen(t, c, store, clk)
}

func TestRenameDeck(t *testing.T) {
	c, store, clk := testCollection(t)
	lang, _ := c.CreateDeck("Languages::Spanish::Verbs")
	c.CreateDeck("Languages::French")
	var languages ID
	for _, d := range c.Decks() {
		if d.Name == "Languages" {
			languages = d.ID
		}
	}
	spanish := c.deckByName("Languages::Spanish").ID

	// Duplicate sibling name rejected (spec), ignoring case; nothing changes.
	before := deckNames(c)
	mustUserErr(t, c.RenameDeck(spanish, "Languages::french"), "already taken")
	if deckNames(c) != before {
		t.Fatal("rejected rename changed decks")
	}

	// Children move with their parent.
	if err := c.RenameDeck(languages, "Lang"); err != nil {
		t.Fatal(err)
	}
	if got := deckNames(c); got != "Default, Lang, Lang::French, Lang::Spanish, Lang::Spanish::Verbs" {
		t.Fatalf("after rename: %q", got)
	}
	// Move under another (new) parent.
	if err := c.RenameDeck(spanish, "Archive::Spanish"); err != nil {
		t.Fatal(err)
	}
	if got := deckNames(c); got != "Archive, Archive::Spanish, Archive::Spanish::Verbs, Default, Lang, Lang::French" {
		t.Fatalf("after move: %q", got)
	}
	if d, _ := c.Deck(lang); d.Name != "Archive::Spanish::Verbs" {
		t.Fatalf("leaf = %q", d.Name)
	}
	// Not into its own subtree.
	mustUserErr(t, c.RenameDeck(spanish, "Archive::Spanish::Verbs::X"), "into itself")
	// Changing only the case is allowed.
	if err := c.RenameDeck(spanish, "Archive::SPANISH"); err != nil {
		t.Fatal(err)
	}
	reopen(t, c, store, clk)
}

func TestDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	parent, _ := c.CreateDeck("P")
	child, _ := c.CreateDeck("P::C")
	if _, err := c.AddPreset(child, "Child"); err != nil {
		t.Fatal(err)
	}
	opts := deckOpts(5, 50, 0.85)
	if err := c.SetDeckOptions(child, opts); err != nil {
		t.Fatal(err)
	}
	// Out-of-range retention (spec): rejected, previous value kept.
	bad := opts
	bad.DesiredRetention = 0.5
	mustUserErr(t, c.SetDeckOptions(child, bad), "between 0.70 and 0.99")
	for _, r := range []float64{0.6999, 0.991, 1} {
		bad.DesiredRetention = r
		if err := c.SetDeckOptions(child, bad); err == nil {
			t.Errorf("retention %v accepted", r)
		}
	}
	for _, r := range []float64{0.7, 0.99} {
		ok := opts
		ok.DesiredRetention = r
		if err := c.SetDeckOptions(parent, ok); err != nil {
			t.Errorf("retention %v rejected: %v", r, err)
		}
	}
	mustUserErr(t, c.SetDeckOptions(child, deckOpts(-1, 1, 0.9)), "New cards per day")
	// The child keeps its own preset, independent of the parent.
	if o := deckOptions(c, child); !o.Equal(opts) {
		t.Fatalf("child options = %+v", o)
	}
	reopen(t, c, store, clk)
}

// deckOpts returns the default deck options with these limits and
// retention.
// ownPreset gives deck a preset of its own, a copy of its current one.
func ownPreset(t *testing.T, c *Collection, deck ID) ID {
	t.Helper()
	d, _ := c.Deck(deck)
	id, err := c.ClonePreset(deck, "Own "+d.Name)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// deckOptions returns the options of deck id's preset.
func deckOptions(c *Collection, id ID) DeckOptions {
	o, _ := c.DeckOptions(id)
	return o
}

func deckOpts(newPerDay, reviewsPerDay int, retention float64) DeckOptions {
	o := DefaultDeckOptions()
	o.NewPerDay, o.ReviewsPerDay, o.DesiredRetention = newPerDay, reviewsPerDay, retention
	return o
}

func TestStepDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := defaultDeck(c)
	// Valid steps (spec): 30s 5m 1h 2d are saved and shown the same way.
	opts := DefaultDeckOptions()
	opts.LearningSteps = []int{30, 300, 3600, 2 * 86400}
	opts.RelearningSteps = []int{}
	opts.InterdayLearningOrder = InterdayBefore
	opts.BuryInterdayLearning = true
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	// Invalid steps are rejected and the previous steps kept.
	for name, mod := range map[string]func(o *DeckOptions){
		"zero step":     func(o *DeckOptions) { o.LearningSteps = []int{0} },
		"negative step": func(o *DeckOptions) { o.RelearningSteps = []int{-60} },
		"over a year":   func(o *DeckOptions) { o.LearningSteps = []int{MaxStepSeconds + 1} },
		"21 steps":      func(o *DeckOptions) { o.LearningSteps = make([]int, 21) },
		"unknown order": func(o *DeckOptions) { o.InterdayLearningOrder = "sideways" },
		"no order":      func(o *DeckOptions) { o.InterdayLearningOrder = "" },
	} {
		bad := DefaultDeckOptions()
		mod(&bad)
		if err := c.SetDeckOptions(deck, bad); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	mustUserErr(t, c.SetDeckOptions(deck, func() DeckOptions {
		o := DefaultDeckOptions()
		o.LearningSteps = []int{60, 0}
		return o
	}()), "Learning steps must each be longer than 0")
	check := func(c *Collection) {
		t.Helper()
		got := deckOptions(c, deck)
		if !got.Equal(opts) || got.RelearningSteps == nil {
			t.Fatalf("options = %+v", got)
		}
		// The steps drive the scheduler: a 2-day step, no relearning steps.
		card := &Card{DeckID: deck}
		p := c.params(card)
		if len(p.LearningSteps) != 4 || p.LearningSteps[3] != 48*time.Hour || len(p.RelearningSteps) != 0 || p.RolloverHour != 4 {
			t.Fatalf("params = %+v", p)
		}
	}
	check(c)
	// Empty relearning steps survive a restart: they are not the defaults.
	check(reopen(t, c, store, clk))
}

func TestDeckOptionsStoredBeforeSteps(t *testing.T) {
	// Options stored before the step options existed get the defaults.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5,"reviewsPerDay":50,"desiredRetention":0.85}`), &o); err != nil {
		t.Fatal(err)
	}
	want := deckOpts(5, 50, 0.85)
	want.BuryNew, want.BuryReview = true, true
	if !o.Equal(want) {
		t.Fatalf("options = %+v", o)
	}
}

func TestLeechDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := defaultDeck(c)
	// Options stored (spec): threshold 5, suspend.
	opts := DefaultDeckOptions()
	opts.LeechThreshold, opts.LeechAction = 5, LeechSuspend
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	for name, mod := range map[string]func(o *DeckOptions){
		"threshold 0":   func(o *DeckOptions) { o.LeechThreshold = 0 },
		"threshold 100": func(o *DeckOptions) { o.LeechThreshold = 100 },
		"no action":     func(o *DeckOptions) { o.LeechAction = "" },
		"unknown":       func(o *DeckOptions) { o.LeechAction = "delete" },
	} {
		bad := opts
		mod(&bad)
		if err := c.SetDeckOptions(deck, bad); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	for _, n := range []int{1, 99} {
		ok := opts
		ok.LeechThreshold = n
		if err := c.SetDeckOptions(deck, ok); err != nil {
			t.Errorf("threshold %d rejected: %v", n, err)
		}
	}
	mustUserErr(t, c.SetDeckOptions(deck, func() DeckOptions {
		o := opts
		o.LeechThreshold = 0
		return o
	}()), "between 1 and 99")
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	c2 := reopen(t, c, store, clk)
	if o := deckOptions(c2, deck); o.LeechThreshold != 5 || o.LeechAction != LeechSuspend {
		t.Fatalf("options after reopen = %+v", o)
	}
	// Options stored before the leech options existed get the defaults.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.LeechThreshold != 8 || o.LeechAction != LeechTag {
		t.Fatalf("old options = %+v", o)
	}
}

func TestTimerDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := defaultDeck(c)
	opts := DefaultDeckOptions()
	opts.MaxAnswerSeconds, opts.ShowTimer, opts.StopTimerOnAnswer = 90, true, false
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	// Out of range (spec): 0 is rejected and the previous value kept.
	for _, n := range []int{0, -1, 10000} {
		bad := opts
		bad.MaxAnswerSeconds = n
		mustUserErr(t, c.SetDeckOptions(deck, bad), "between 1 and 9999")
	}
	if o := deckOptions(c, deck); o.MaxAnswerSeconds != 90 {
		t.Fatalf("options after rejected change = %+v", o)
	}
	for _, n := range []int{1, 9999} {
		ok := opts
		ok.MaxAnswerSeconds = n
		if err := c.SetDeckOptions(deck, ok); err != nil {
			t.Errorf("maximum %d rejected: %v", n, err)
		}
	}
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	c2 := reopen(t, c, store, clk)
	if o := deckOptions(c2, deck); o.MaxAnswerSeconds != 90 || !o.ShowTimer || o.StopTimerOnAnswer {
		t.Fatalf("options after reopen = %+v", o)
	}
	// Options stored before the timer options existed get the defaults.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.MaxAnswerSeconds != 60 || o.ShowTimer || !o.StopTimerOnAnswer {
		t.Fatalf("old options = %+v", o)
	}
}

func TestAutoAdvanceDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := defaultDeck(c)
	opts := DefaultDeckOptions()
	opts.SecondsToShowQuestion, opts.SecondsToShowAnswer, opts.AnswerAction = 5, 3, AnswerGood
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{-1, 10000} {
		bad := opts
		bad.SecondsToShowQuestion = n
		mustUserErr(t, c.SetDeckOptions(deck, bad), "between 0 and 9999")
		bad = opts
		bad.SecondsToShowAnswer = n
		mustUserErr(t, c.SetDeckOptions(deck, bad), "between 0 and 9999")
	}
	bad := opts
	bad.QuestionAction = "skip"
	mustUserErr(t, c.SetDeckOptions(deck, bad), "Unknown question action")
	bad = opts
	bad.AnswerAction = "skip"
	mustUserErr(t, c.SetDeckOptions(deck, bad), "Unknown answer action")
	for _, a := range []string{AnswerBuryCard, AnswerAgain, AnswerHard, AnswerShowReminder} {
		ok := opts
		ok.QuestionAction, ok.AnswerAction = QuestionShowReminder, a
		if err := c.SetDeckOptions(deck, ok); err != nil {
			t.Errorf("answer action %q rejected: %v", a, err)
		}
	}
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	// Options stored (spec): the values persist.
	c2 := reopen(t, c, store, clk)
	if o := deckOptions(c2, deck); o.SecondsToShowQuestion != 5 || o.SecondsToShowAnswer != 3 ||
		o.QuestionAction != QuestionShowAnswer || o.AnswerAction != AnswerGood {
		t.Fatalf("options after reopen = %+v", o)
	}
	// Options stored before auto advance existed get the defaults.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.SecondsToShowQuestion != 0 || o.SecondsToShowAnswer != 0 ||
		o.QuestionAction != QuestionShowAnswer || o.AnswerAction != AnswerBuryCard {
		t.Fatalf("old options = %+v", o)
	}
}

func TestIntervalDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	deck := defaultDeck(c)
	opts := DefaultDeckOptions()
	if opts.MaxInterval != 36500 || opts.MinRelearnInterval != 1 {
		t.Fatalf("defaults = %+v", opts)
	}
	opts.MaxInterval, opts.MinRelearnInterval = 180, 3
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	for name, mod := range map[string]func(o *DeckOptions){
		"max 0":      func(o *DeckOptions) { o.MaxInterval = 0 },
		"max 36501":  func(o *DeckOptions) { o.MaxInterval = 36501 },
		"min 0":      func(o *DeckOptions) { o.MinRelearnInterval = 0 },
		"min 36501":  func(o *DeckOptions) { o.MaxInterval, o.MinRelearnInterval = 36500, 36501 },
		"min of 181": func(o *DeckOptions) { o.MinRelearnInterval = 181 },
	} {
		bad := opts
		mod(&bad)
		if err := c.SetDeckOptions(deck, bad); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	for _, n := range []int{1, 36500} {
		ok := opts
		ok.MaxInterval, ok.MinRelearnInterval = n, n
		if err := c.SetDeckOptions(deck, ok); err != nil {
			t.Errorf("%d days rejected: %v", n, err)
		}
	}
	// Minimum above maximum (spec): 30 and 60 are rejected with an
	// explanation.
	mustUserErr(t, c.SetDeckOptions(deck, func() DeckOptions {
		o := opts
		o.MaxInterval, o.MinRelearnInterval = 30, 60
		return o
	}()), "must not be above the maximum interval")
	if err := c.SetDeckOptions(deck, opts); err != nil {
		t.Fatal(err)
	}
	c2 := reopen(t, c, store, clk)
	if o := deckOptions(c2, deck); o.MaxInterval != 180 || o.MinRelearnInterval != 3 {
		t.Fatalf("options after reopen = %+v", o)
	}
	// The options drive the scheduler.
	if p := c2.params(&Card{DeckID: deck}); p.MaximumInterval != 180 || p.MinimumRelearnInterval != 3 {
		t.Fatalf("params = %+v", p)
	}
	// Options stored before the interval options existed get the defaults.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.MaxInterval != 36500 || o.MinRelearnInterval != 1 {
		t.Fatalf("old options = %+v", o)
	}
}

func TestDeleteLastDeckRejected(t *testing.T) {
	c, _, _ := testCollection(t)
	def := c.Decks()[0].ID
	// Deleting the last deck (spec): rejected, the deck remains.
	mustUserErr(t, c.DeleteDeck(def), "last deck")
	if deckNames(c) != "Default" {
		t.Fatal("deck removed")
	}
	// Deleting a parent that holds every deck is the same.
	c.RenameDeck(def, "Root::Default")
	root := c.deckByName("Root").ID
	mustUserErr(t, c.DeleteDeck(root), "last deck")
	// With another deck present, the Default deck may go.
	other, _ := c.CreateDeck("Other")
	if err := c.DeleteDeck(root); err != nil {
		t.Fatal(err)
	}
	if deckNames(c) != "Other" {
		t.Fatalf("decks = %q", deckNames(c))
	}
	mustUserErr(t, c.DeleteDeck(other), "last deck")
}

func TestSettings(t *testing.T) {
	c, store, clk := testCollection(t)
	s := c.Settings()
	s.Theme = ThemeDark
	s.RolloverHour = 6
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	bad := s
	bad.RolloverHour = 24
	mustUserErr(t, c.UpdateSettings(bad), "rollover")
	bad = s
	bad.Theme = "blue"
	mustUserErr(t, c.UpdateSettings(bad), "theme")
	c2 := reopen(t, c, store, clk)
	if got := c2.Settings(); got.Theme != ThemeDark || got.RolloverHour != 6 {
		t.Fatalf("settings after reopen = %+v", got)
	}
}

// A failed save changes nothing in memory.
func TestFailedCommitChangesNothing(t *testing.T) {
	store := storage.NewMemoryStore()
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	b := &failingBackend{Backend: store.Open()}
	c, _, err := Open(b, Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	b.fail = true
	if _, err := c.CreateDeck("X"); err == nil || IsUserError(err) {
		t.Fatalf("err = %v", err)
	}
	if deckNames(c) != "Default" {
		t.Fatalf("decks = %q", deckNames(c))
	}
}

type failingBackend struct {
	storage.Backend
	fail bool
}

var errDisk = errors.New("disk full")

func (f *failingBackend) Append(k string, r []byte) error {
	if f.fail {
		return errDisk
	}
	return f.Backend.Append(k, r)
}

func (f *failingBackend) Put(k string, v []byte) error {
	if f.fail {
		return errDisk
	}
	return f.Backend.Put(k, v)
}

// Corrupt stored data is reported, not replaced.
func TestOpenCorrupt(t *testing.T) {
	store := storage.NewMemoryStore()
	b := store.Open()
	b.Put("current", []byte(`{"v":1,"gen":1,"rl":1,"media":1}`))
	b.Put("g1/snapshot", []byte(`{"decks": "nope"`))
	if _, _, err := Open(store.Open(), Options{}); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("err = %v", err)
	}
	if v, _ := b.Get("g1/snapshot"); string(v) != `{"decks": "nope"` {
		t.Fatal("corrupt data modified")
	}
}

func testStore() *storage.MemoryStore { return storage.NewMemoryStore() }

func TestLimitOptionDefaults(t *testing.T) {
	c, _, _ := testCollection(t)
	// Defaults (spec): new cards ignore the review limit, limits do not
	// start from the top.
	id, _ := c.CreateDeck("Spanish")
	if d := deckOptions(c, id); !d.NewIgnoresReviewLimit || d.LimitsStartFromTop {
		t.Fatalf("options = %+v", d)
	}
	// Options stored before they existed get the defaults too.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5,"reviewsPerDay":50,"desiredRetention":0.9}`), &o); err != nil {
		t.Fatal(err)
	}
	if !o.NewIgnoresReviewLimit || o.LimitsStartFromTop {
		t.Fatalf("stored options = %+v", o)
	}
}

func TestBuryingOptionDefaults(t *testing.T) {
	c, _, _ := testCollection(t)
	// Defaults (spec): both burying options are off, in the Default preset
	// of a new collection and for a new deck.
	if p := c.preset(DefaultPresetID); p.Options.BuryNew || p.Options.BuryReview {
		t.Fatalf("Default preset = %+v", p.Options)
	}
	id, _ := c.CreateDeck("Spanish")
	if d := deckOptions(c, id); d.BuryNew || d.BuryReview {
		t.Fatalf("options = %+v", d)
	}
	// New preset (spec): a preset starting with the defaults buries nothing.
	added, err := c.AddPreset(id, "Languages")
	if err != nil {
		t.Fatal(err)
	}
	if p := c.preset(added); p.ID != added || p.Options.BuryNew || p.Options.BuryReview {
		t.Fatalf("added preset = %+v", p)
	}
	// Options stored without them (before they existed) bury siblings, as
	// the defaults of that time did.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5,"reviewsPerDay":50,"desiredRetention":0.9}`), &o); err != nil {
		t.Fatal(err)
	}
	if !o.BuryNew || !o.BuryReview {
		t.Fatalf("stored options = %+v", o)
	}
}

func TestTodayLimitsExpireAtRollover(t *testing.T) {
	c, store, clk := testCollection(t)
	id, _ := c.CreateDeck("Spanish")
	limits := func() (int, int) {
		n, r, _ := c.EffectiveLimits(id)
		return n, r
	}
	if err := c.IncreaseTodayLimits(id, 10, 0); err != nil {
		t.Fatal(err)
	}
	if n, r := limits(); n != 30 || r != 200 || deckOptions(c, id).NewPerDay != 20 {
		t.Fatalf("limits = %d/%d, options %+v", n, r, deckOptions(c, id))
	}
	// A second increase adds to the effective limit.
	if err := c.IncreaseTodayLimits(id, 5, 50); err != nil {
		t.Fatal(err)
	}
	if n, r := limits(); n != 35 || r != 250 {
		t.Fatalf("limits = %d/%d", n, r)
	}
	// The override persists ...
	c = reopen(t, c, store, clk)
	// ... and expires with the next study day (spec, "Override expires").
	clk.t = scheduler.NextDayStart(clk.t, 4)
	if n, r := limits(); n != 20 || r != 200 {
		t.Fatalf("limits after rollover = %d/%d", n, r)
	}
	// An increase on the new day starts from the daily limits.
	if err := c.IncreaseTodayLimits(id, 0, 1); err != nil {
		t.Fatal(err)
	}
	if n, r := limits(); n != 20 || r != 201 {
		t.Fatalf("limits = %d/%d", n, r)
	}
	if err := c.IncreaseTodayLimits(id, -1, 0); err == nil {
		t.Fatal("negative increase accepted")
	}
	// Undo reverts the increase.
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Custom study" {
		t.Fatalf("undo = %q", undo)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, r := limits(); r != 200 {
		t.Fatalf("review limit after undo = %d", r)
	}
}

func TestDisplayOrderOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	// New card defaults and review defaults (spec).
	id, _ := c.CreateDeck("Spanish")
	if o := deckOptions(c, id); o.NewGatherOrder != GatherAscending || o.NewSortOrder != SortGathered ||
		o.NewReviewOrder != NewMix || o.ReviewSortOrder != ReviewDueRandom {
		t.Fatalf("options = %+v", o)
	}
	// Options stored before they existed get the defaults too.
	var o DeckOptions
	if err := json.Unmarshal([]byte(`{"newPerDay":5,"reviewsPerDay":50,"desiredRetention":0.9}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.NewGatherOrder != GatherAscending || o.NewSortOrder != SortGathered ||
		o.NewReviewOrder != NewMix || o.ReviewSortOrder != ReviewDueRandom {
		t.Fatalf("stored options = %+v", o)
	}
	for _, tc := range []struct {
		set  func(*DeckOptions)
		want string
	}{
		{func(o *DeckOptions) { o.NewGatherOrder = "shuffle" }, "new card gather order"},
		{func(o *DeckOptions) { o.NewSortOrder = "" }, "new card sort order"},
		{func(o *DeckOptions) { o.NewReviewOrder = "first" }, "new/review order"},
		{func(o *DeckOptions) { o.ReviewSortOrder = "ease" }, "review sort order"},
	} {
		bad := DefaultDeckOptions()
		tc.set(&bad)
		mustUserErr(t, c.SetDeckOptions(id, bad), tc.want)
	}
	opts := DefaultDeckOptions()
	opts.NewGatherOrder, opts.NewSortOrder = GatherRandomNotes, SortTypeThenRandom
	opts.NewReviewOrder, opts.ReviewSortOrder = NewAfterReviews, ReviewRelativeOverdueness
	if err := c.SetDeckOptions(id, opts); err != nil {
		t.Fatal(err)
	}
	c = reopen(t, c, store, clk)
	if o := deckOptions(c, id); !o.Equal(opts) {
		t.Fatalf("options after reopen = %+v", o)
	}
}

// presetNamed returns the ID of the preset with this name.
func presetNamed(t *testing.T, c *Collection, name string) ID {
	t.Helper()
	for _, p := range c.Presets() {
		if p.Name == name {
			return p.ID
		}
	}
	t.Fatalf("no preset %q", name)
	return 0
}

func TestNewCollectionPresets(t *testing.T) {
	c, _, _ := testCollection(t)
	// A new collection has the preset Default, used by every deck, new
	// ones included.
	spanish, _ := c.CreateDeck("Languages::Spanish")
	ps := c.Presets()
	if len(ps) != 1 || ps[0].ID != DefaultPresetID || ps[0].Name != "Default" || !ps[0].Options.Equal(DefaultDeckOptions()) {
		t.Fatalf("presets = %+v", ps)
	}
	for _, d := range c.Decks() {
		if d.PresetID != DefaultPresetID {
			t.Errorf("%s uses preset %d", d.Name, d.PresetID)
		}
	}
	if u := c.PresetUsage(); u[DefaultPresetID] != 3 {
		t.Fatalf("usage = %v", u)
	}
	if d, _ := c.Deck(spanish); d.PresetID != DefaultPresetID {
		t.Fatalf("new deck preset = %d", d.PresetID)
	}
}

func TestSharedPreset(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	french, _ := c.CreateDeck("French")
	german, _ := c.CreateDeck("German")
	languages, err := c.AddPreset(spanish, "Languages")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetDeckPreset(french, languages); err != nil {
		t.Fatal(err)
	}
	// Shared preset (spec): setting retention from Spanish changes French.
	o := deckOptions(c, spanish)
	o.DesiredRetention = 0.85
	if err := c.SetDeckOptions(spanish, o); err != nil {
		t.Fatal(err)
	}
	if got := deckOptions(c, french).DesiredRetention; got != 0.85 {
		t.Fatalf("French retention = %v", got)
	}
	if got := deckOptions(c, german).DesiredRetention; got != DefaultRetention {
		t.Fatalf("German retention = %v", got)
	}
	if u := c.PresetUsage(); u[languages] != 2 {
		t.Fatalf("usage = %v", u)
	}
	reopen(t, c, store, clk)
}

func TestManagePresets(t *testing.T) {
	c, store, clk := testCollection(t)
	deck, _ := c.CreateDeck("Spanish")
	o := deckOpts(7, 70, 0.8)
	if err := c.SetDeckOptions(deck, o); err != nil {
		t.Fatal(err)
	}
	// Clone copies the deck's preset and switches the deck to the copy.
	clone, err := c.ClonePreset(deck, "  Spanish   copy ")
	if err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(deck); d.PresetID != clone || !deckOptions(c, deck).Equal(o) {
		t.Fatalf("after clone: deck %+v, options %+v", d, deckOptions(c, deck))
	}
	if presetNamed(t, c, "Spanish copy") != clone {
		t.Fatal("clone name not normalized")
	}
	// Add starts from the defaults.
	added, err := c.AddPreset(deck, "Fresh")
	if err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(deck); d.PresetID != added || !deckOptions(c, deck).Equal(DefaultDeckOptions()) {
		t.Fatalf("after add: deck %+v", d)
	}
	// Names are unique, ignoring case.
	mustUserErr(t, func() error { _, err := c.AddPreset(deck, "fresh"); return err }(), "already exists")
	mustUserErr(t, func() error { _, err := c.ClonePreset(deck, "DEFAULT"); return err }(), "already exists")
	mustUserErr(t, func() error { _, err := c.AddPreset(deck, "  "); return err }(), "enter a preset name")
	mustUserErr(t, c.RenamePreset(added, "spanish COPY"), "already exists")
	if err := c.RenamePreset(added, "Fresh"); err != nil {
		t.Fatalf("renaming to the same name: %v", err)
	}
	if err := c.RenamePreset(added, "fresh"); err != nil {
		t.Fatalf("changing case: %v", err)
	}
	if err := c.RenamePreset(DefaultPresetID, "Standard"); err != nil {
		t.Fatal(err)
	}
	// Default cannot be deleted.
	mustUserErr(t, c.DeletePreset(DefaultPresetID), "cannot be deleted")
	mustUserErr(t, c.SetDeckPreset(deck, 12345), "no longer exists")
	if got := names(c.Presets()); got != "fresh, Spanish copy, Standard" {
		t.Fatalf("presets = %s", got)
	}
	reopen(t, c, store, clk)
}

func names(ps []Preset) string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Name)
	}
	return strings.Join(out, ", ")
}

func TestDeletePreset(t *testing.T) {
	c, store, clk := testCollection(t)
	var decks []ID
	for _, name := range []string{"A", "B", "C"} {
		id, _ := c.CreateDeck(name)
		decks = append(decks, id)
	}
	other, _ := c.CreateDeck("Other")
	exam, _ := c.AddPreset(decks[0], "Exam")
	for _, d := range decks[1:] {
		if err := c.SetDeckPreset(d, exam); err != nil {
			t.Fatal(err)
		}
	}
	keep, _ := c.AddPreset(other, "Keep")
	// Delete a preset (spec): its 3 decks use Default.
	if u := c.PresetUsage(); u[exam] != 3 {
		t.Fatalf("usage = %v", u)
	}
	if err := c.DeletePreset(exam); err != nil {
		t.Fatal(err)
	}
	for _, id := range decks {
		if d, _ := c.Deck(id); d.PresetID != DefaultPresetID {
			t.Fatalf("%s uses %d", d.Name, d.PresetID)
		}
	}
	if d, _ := c.Deck(other); d.PresetID != keep {
		t.Fatal("other deck changed preset")
	}
	if got := names(c.Presets()); got != "Default, Keep" {
		t.Fatalf("presets = %s", got)
	}
	// Undo brings the preset back, with its decks.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if u := c.PresetUsage(); u[exam] != 3 {
		t.Fatalf("usage after undo = %v", u)
	}
	reopen(t, c, store, clk)
}

func TestSavePresetToSubdecks(t *testing.T) {
	c, store, clk := testCollection(t)
	languages, _ := c.CreateDeck("Languages")
	spanish, _ := c.CreateDeck("Languages::Spanish")
	verbs, _ := c.CreateDeck("Languages::Spanish::Verbs")
	other, _ := c.CreateDeck("Languages2")
	preset, _ := c.AddPreset(languages, "Languages")
	// A child deck uses its own preset, independent of its parent.
	if d, _ := c.Deck(spanish); d.PresetID != DefaultPresetID {
		t.Fatalf("child preset = %d", d.PresetID)
	}
	// Save to all subdecks (spec): every descendant uses the preset.
	if err := c.SavePresetToSubdecks(languages); err != nil {
		t.Fatal(err)
	}
	for _, id := range []ID{spanish, verbs} {
		if d, _ := c.Deck(id); d.PresetID != preset {
			t.Fatalf("%s uses %d", d.Name, d.PresetID)
		}
	}
	if d, _ := c.Deck(other); d.PresetID != DefaultPresetID {
		t.Fatal("a deck outside the tree changed")
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Save to all subdecks" {
		t.Fatalf("undo = %q", undo)
	}
	reopen(t, c, store, clk)
}

func TestPresetSwitchRepositionsRandomDecks(t *testing.T) {
	c, _, _ := testCollection(t)
	words, _ := c.CreateDeck("Words")
	ids := addCards(t, c, words, 50)
	random := DefaultDeckOptions()
	random.InsertionOrder = InsertRandom
	shuffled, _ := c.AddPreset(words, "Random")
	if err := c.SetPresetOptions(shuffled, random); err != nil {
		t.Fatal(err)
	}
	// Using a preset with another insertion order repositions the deck's
	// new cards, as changing the option does.
	if slices.Equal(byPosition(t, c, ids), ids) {
		t.Fatal("random preset kept the order")
	}
	if err := c.SetDeckPreset(words, DefaultPresetID); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(byPosition(t, c, ids), ids) {
		t.Fatal("sequential preset kept the random order")
	}
}

func TestDeckLimits(t *testing.T) {
	c, store, clk := testCollection(t)
	deck, _ := c.CreateDeck("French")
	five, ten := 5, 10
	if err := c.SetDeckLimits(deck, DeckLimits{New: &five, Review: &ten}); err != nil {
		t.Fatal(err)
	}
	if n, r, _ := c.EffectiveLimits(deck); n != 5 || r != 10 {
		t.Fatalf("limits = %d/%d", n, r)
	}
	// The preset's options are unchanged.
	if o := deckOptions(c, deck); o.NewPerDay != DefaultNewPerDay || o.ReviewsPerDay != DefaultReviewsPerDay {
		t.Fatalf("preset options = %+v", o)
	}
	bad := -1
	mustUserErr(t, c.SetDeckLimits(deck, DeckLimits{New: &bad}), "New cards per day")
	mustUserErr(t, c.SetDeckLimits(deck, DeckLimits{Review: &bad}), "Maximum reviews per day")
	// A custom study increase adds to the deck's own limit.
	if err := c.IncreaseTodayLimits(deck, 3, 0); err != nil {
		t.Fatal(err)
	}
	if n, _, _ := c.EffectiveLimits(deck); n != 8 {
		t.Fatalf("new limit after increase = %d", n)
	}
	c = reopen(t, c, store, clk)
	if err := c.SetDeckLimits(deck, DeckLimits{Review: &ten}); err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(deck); d.Limits.New != nil || *d.Limits.Review != 10 {
		t.Fatalf("limits = %+v", d.Limits)
	}
}

func TestSaveDeckOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	french, _ := c.CreateDeck("French")
	ids := addCards(t, c, french, 30)
	languages, _ := c.AddPreset(spanish, "Languages")
	// Saving from French: it switches to Languages, whose options change
	// for Spanish too, and gets its own new limit; one undo step.
	o := deckOpts(15, 150, 0.85)
	o.InsertionOrder = InsertRandom
	five := 5
	if err := c.SaveDeckOptions(french, languages, o, DeckLimits{New: &five}); err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(french); d.PresetID != languages || *d.Limits.New != 5 {
		t.Fatalf("French = %+v", d)
	}
	if got := deckOptions(c, spanish); !got.Equal(o) {
		t.Fatalf("Spanish options = %+v", got)
	}
	if n, _, _ := c.EffectiveLimits(french); n != 5 {
		t.Fatalf("French new limit = %d", n)
	}
	// French's insertion order changed to random: its cards were shuffled.
	if slices.Equal(byPosition(t, c, ids), ids) {
		t.Fatal("French kept its order")
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Change deck options" {
		t.Fatalf("undo = %q", undo)
	}
	// Invalid values change nothing.
	bad := -1
	mustUserErr(t, c.SaveDeckOptions(french, DefaultPresetID, o, DeckLimits{Review: &bad}), "Maximum reviews")
	o.DesiredRetention = 2
	mustUserErr(t, c.SaveDeckOptions(french, DefaultPresetID, o, DeckLimits{}), "Desired retention")
	if d, _ := c.Deck(french); d.PresetID != languages {
		t.Fatal("preset changed")
	}
	reopen(t, c, store, clk)
}

// Deck descriptions (collection spec, "Edit description").
func TestDeckDescription(t *testing.T) {
	c, store, clk := testCollection(t)
	id, _ := c.CreateDeck("Spanish")
	if err := c.SetDeckDescription(id, "  Vocabulary from <b>Book 1</b>\n"); err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(id); d.Description != "Vocabulary from <b>Book 1</b>" {
		t.Fatalf("description = %q", d.Description)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Edit deck description" {
		t.Fatalf("undo = %q", undo)
	}
	c = reopen(t, c, store, clk)
	// The description moves with a renamed deck.
	if err := c.RenameDeck(id, "Languages::Spanish"); err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(id); d.Description != "Vocabulary from <b>Book 1</b>" {
		t.Fatalf("description after rename = %q", d.Description)
	}
	if err := c.SetDeckDescription(id, ""); err != nil {
		t.Fatal(err)
	}
	if d, _ := c.Deck(id); d.Description != "" {
		t.Fatalf("cleared description = %q", d.Description)
	}
	if err := c.SetDeckDescription(12345, "x"); err == nil {
		t.Fatal("description of a missing deck accepted")
	}
}

// Deck descriptions on Anki import (data-import spec): new decks get the
// batch's description, existing decks keep theirs.
func TestImportDeckDescriptions(t *testing.T) {
	c, _, _ := testCollection(t)
	existing, _ := c.CreateDeck("Old")
	if err := c.SetDeckDescription(existing, "mine"); err != nil {
		t.Fatal(err)
	}
	b := &ImportBatch{
		Decks:            []string{"Old", "Shared::Sub"},
		DeckDescriptions: map[string]string{"old": "theirs", "Shared::Sub": "Words from <i>Book 2</i>", "Shared": " Top "},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, d := range c.Decks() {
		got[d.Name] = d.Description
	}
	want := map[string]string{"Default": "", "Old": "mine", "Shared": "Top", "Shared::Sub": "Words from <i>Book 2</i>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("descriptions = %v, want %v", got, want)
	}
}
