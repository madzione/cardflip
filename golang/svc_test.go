package main

import (
	"bytes"
	"cardflip/golang/internal/usermsg"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"cardflip/golang/backup"
	"cardflip/golang/core"
	"cardflip/golang/importer/apkg"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
	"cardflip/golang/stats"
	"cardflip/golang/storage"
)

// services is the full set of bound services on one app.
type services struct {
	app      *AppService
	decks    *DecksService
	notes    *NotesService
	study    *StudyService
	browse   *BrowseService
	transfer *TransferService
	media    *MediaService
	stats    *StatsService
	progress *ProgressService
	mcp      *McpService
	clock    *time.Time
	// stores are the profiles' data by data directory; the default
	// profile's is the store newServices was given.
	stores   map[string]*storage.MemoryStore
	registry []byte // the stored profile registry
	mcpFile  []byte // the stored MCP settings
}

// profileStore returns the data of profile id of the app opened on /data.
func (s *services) profileStore(id string) *storage.MemoryStore {
	return s.stores[storage.ProfileDataDir("/data", id)]
}

// newServices builds the services on an in-memory store, like fluffy.Start.
func newServices(t testing.TB, store *storage.MemoryStore) *services {
	t.Helper()
	a := newApp()
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.Local)
	s := &services{clock: &now}
	a.now = func() time.Time { return *s.clock }
	s.stores = map[string]*storage.MemoryStore{"/data": store}
	a.openBackend = func(dir string) (storage.Backend, error) {
		if s.stores[dir] == nil {
			s.stores[dir] = storage.NewMemoryStore()
		}
		return s.stores[dir].Open(), nil
	}
	a.removeProfile = func(root, id string) error {
		delete(s.stores, storage.ProfileDataDir(root, id))
		return nil
	}
	a.readProfiles = func(string) ([]byte, error) {
		if s.registry == nil {
			return nil, storage.ErrNotFound
		}
		return s.registry, nil
	}
	a.writeProfiles = func(_ string, data []byte) error {
		s.registry = data
		return nil
	}
	a.persist = func() bool { return true }
	a.backups.tick = 0 // tests call periodicBackup themselves
	s.app, s.decks, s.notes = &AppService{a: a}, &DecksService{a: a}, &NotesService{a: a}
	s.study, s.browse = &StudyService{a: a}, &BrowseService{a: a}
	s.transfer, s.media = &TransferService{a: a}, &MediaService{a: a}
	s.stats, s.progress = &StatsService{a: a}, &ProgressService{a: a}
	a.resets = []func(){s.study.reset, s.transfer.reset, s.progress.reset}
	s.mcp = newMcpService(a)
	s.mcp.read = func(string) ([]byte, error) {
		if s.mcpFile == nil {
			return nil, storage.ErrNotFound
		}
		return s.mcpFile, nil
	}
	s.mcp.write = func(_ string, data []byte) error {
		s.mcpFile = data
		return nil
	}
	t.Cleanup(func() {
		s.mcp.mu.Lock()
		defer s.mcp.mu.Unlock()
		s.mcp.stopLocked()
	})
	return s
}

func opened(t testing.TB) (*services, *storage.MemoryStore) {
	t.Helper()
	store := storage.NewMemoryStore()
	s := newServices(t, store)
	st, err := s.app.Open("/data")
	if err != nil || st.State != StatusFresh {
		t.Fatalf("open: %+v %v", st, err)
	}
	return s, store
}

// must returns v, failing the test (by panicking) when err is set.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func check(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestOpenGatingAndIdempotence(t *testing.T) {
	store := storage.NewMemoryStore()
	s := newServices(t, store)
	if _, err := s.decks.Tree(); !errors.Is(err, errNotOpened) {
		t.Fatalf("before open: %v", err)
	}
	if _, err := s.study.Start(1); err == nil {
		t.Fatal("study before open")
	}
	st := must(s.app.Open("/data"))
	if st.State != StatusFresh || st.Version != appVersion {
		t.Fatalf("status = %+v", st)
	}
	// A hot restart calls Open again: same collection, same status.
	again := must(s.app.Open("/data"))
	if again.State != StatusFresh {
		t.Fatalf("second open = %+v", again)
	}
	tree := must(s.decks.Tree())
	if len(tree) != 1 || tree[0].Name != "Default" {
		t.Fatalf("tree = %+v", tree)
	}
	// Restart: the existing collection opens as ok.
	if st := must(newServices(t, store).app.Open("/data")); st.State != StatusOK {
		t.Fatalf("restart = %+v", st)
	}
}

func TestReloadRereadsStorage(t *testing.T) {
	s, store := opened(t)
	basicNote(t, s, "saved", "card")
	st := must(s.app.Reload())
	if st.State != StatusOK {
		t.Fatalf("status = %+v", st)
	}
	if rows := must(s.browse.Rows(must(s.browse.Search("", "created", false)), nil)); len(rows) != 1 || rows[0].Question != "saved" {
		t.Fatalf("rows after reload = %+v", rows)
	}
	// Unreadable storage after a reload is reported, not replaced.
	store.Open().Put("current", []byte("garbage"))
	if st := must(s.app.Reload()); st.State != StatusCorrupt {
		t.Fatalf("status = %+v", st)
	}
}

func TestCorruptStartup(t *testing.T) {
	store := storage.NewMemoryStore()
	b := store.Open()
	b.Put("current", []byte(`{"v":1,"gen":1,"rl":1,"media":1}`))
	b.Put("g1/snapshot", []byte(`garbage`))
	s := newServices(t, store)
	st := must(s.app.Open("/data"))
	if st.State != StatusCorrupt || !strings.Contains(st.Message, "could not be read") {
		t.Fatalf("status = %+v", st)
	}
	if _, err := s.decks.Tree(); err == nil || !strings.Contains(err.Error(), "could not be read") {
		t.Fatalf("tree on corrupt: %v", err)
	}
	if v, _ := b.Get("g1/snapshot"); string(v) != "garbage" {
		t.Fatal("unreadable data was modified before the user chose")
	}
	st = must(s.app.StartFresh())
	if st.State != StatusFresh {
		t.Fatalf("after start fresh = %+v", st)
	}
	if tree := must(s.decks.Tree()); len(tree) != 1 {
		t.Fatalf("tree = %+v", tree)
	}
	if _, err := s.app.StartFresh(); err == nil {
		t.Fatal("start fresh over a readable collection")
	}
}

func TestSettingsAndChanges(t *testing.T) {
	s, store := opened(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := s.app.Changes(ctx)
	if err := s.app.SetTheme("dark"); err != nil {
		t.Fatal(err)
	}
	if ch := <-changes; !ch.Settings {
		t.Fatalf("change = %+v", ch)
	}
	if err := s.app.SetRolloverHour(6); err != nil {
		t.Fatal(err)
	}
	<-changes
	if err := s.app.SetTheme("purple"); err == nil {
		t.Fatal("bad theme accepted")
	}
	if st := must(s.app.Settings()); st.Language != "" {
		t.Fatalf("default language = %q, want the device's", st.Language)
	}
	check(t, s.app.SetLanguage("de"))
	<-changes
	if err := s.app.SetLanguage("German!"); err == nil {
		t.Fatal("bad language accepted")
	}
	// Theme and language persist (app-shell spec).
	restarted := newServices(t, store)
	must(restarted.app.Open("/data"))
	if st := must(restarted.app.Settings()); st.Theme != "dark" || st.RolloverHour != 6 || st.Language != "de" {
		t.Fatalf("settings after restart = %+v", st)
	}
	cancel()
	for range changes {
	}
}

// Cards render for the environment Dart sets; changing it tells open card
// views to render again (card-rendering spec: root classes, media queries).
func TestBuiltinNames(t *testing.T) {
	s := newServices(t, storage.NewMemoryStore())
	german := core.BuiltinNames{Deck: "Standard", Preset: "Standard", NoteTypes: []core.StockNames{
		{Stock: core.StockBasic, Name: "Einfach", Fields: []string{"Vorderseite", "Rückseite"}, CardTypes: []string{"Karte 1"}},
		{Stock: core.StockImageOcclusion, Name: "Bildverdeckung",
			Fields:    []string{"Verdeckung", "Bild", "Kopfzeile", "Rückseite Extra", "Kommentare"},
			CardTypes: []string{"Bildverdeckung"}},
	}}
	// Before Open, so that a new collection gets them.
	s.app.SetBuiltinNames(german)
	if st := must(s.app.Open("/data")); st.State != StatusFresh {
		t.Fatalf("status = %+v", st)
	}
	if decks := must(s.decks.Tree()); len(decks) != 1 || decks[0].Name != "Standard" {
		t.Fatalf("decks = %+v", decks)
	}
	types := must(s.notes.Types())
	d := must(s.notes.Defaults(0))
	var occlusion NoteType
	for _, nt := range types {
		if nt.ID == d.TypeID && nt.Name != "Einfach" {
			t.Fatalf("defaults pick %q", nt.Name)
		}
		if nt.ImageOcclusion {
			occlusion = nt
		}
	}
	if occlusion.Name != "Bildverdeckung" || occlusion.OcclusionField != 0 || occlusion.ImageField != 1 {
		t.Fatalf("occlusion = %+v", occlusion)
	}
	// Later changes apply to the open collection.
	s.app.SetBuiltinNames(core.BuiltinNames{})
	id := must(s.notes.AddType(core.StockBasic, "Basic"))
	for _, nt := range must(s.notes.Types()) {
		if nt.ID == id && nt.Fields[0] != "Front" {
			t.Fatalf("added %+v", nt)
		}
	}
}

func TestRenderEnvironment(t *testing.T) {
	s, _ := opened(t)
	basicNote(t, s, "<b>hola</b>", "hello")
	d := must(s.notes.Defaults(0))
	check(t, s.notes.SetCSS(d.TypeID, `.nightMode b { color: lightblue }
.card.nightMode { background: #222 }
@media (max-width: 400px) { .card { font-size: 30px } }`))
	id := must(s.browse.Search("", "created", false))[0]
	light := must(s.browse.Card(id))
	if strings.Contains(light.Question, "lightblue") || light.Background != "" ||
		!strings.HasPrefix(light.Question, `<div class="card card1">`) {
		t.Fatalf("light: %q background %q", light.Question, light.Background)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := s.app.Changes(ctx)
	s.app.SetRenderEnvironment(render.Env{Night: true, Platform: "android mobile", WidthPx: 350, HeightPx: 700})
	if ch := <-changes; ch.Kind != "environment" {
		t.Fatalf("change = %+v", ch)
	}
	dark := must(s.browse.Card(id))
	if !strings.Contains(dark.Question, `<b style="color:lightblue">hola</b>`) || dark.Background != "#222" ||
		!strings.HasPrefix(dark.Question, `<div class="card card1 nightMode night_mode android mobile" style="font-size:30px">`) {
		t.Fatalf("dark: %q background %q", dark.Question, dark.Background)
	}
	st := must(s.study.Start(d.DeckID))
	if st.Card == nil || !strings.Contains(st.Card.Question, "lightblue") {
		t.Fatalf("study card = %+v", st.Card)
	}
	cancel()
	for range changes {
	}
}

// The note type's media fonts reach the study screen, the browser and the
// template preview (card-rendering spec: fonts from media).
func TestCardFonts(t *testing.T) {
	s, _ := opened(t)
	basicNote(t, s, "hola", "hello")
	d := must(s.notes.Defaults(0))
	css := `@font-face { font-family: jp; src: url("_jp.ttf") } .card { font-family: jp }`
	check(t, s.notes.SetCSS(d.TypeID, css))
	want := []render.FontFace{{Family: "jp", File: "_jp.ttf"}}
	id := must(s.browse.Search("", "created", false))[0]
	if got := must(s.browse.Card(id)).Fonts; !reflect.DeepEqual(got, want) {
		t.Fatalf("browser fonts = %+v", got)
	}
	st := must(s.study.Start(d.DeckID))
	if st.Card == nil || !reflect.DeepEqual(st.Card.Fonts, want) {
		t.Fatalf("study card = %+v", st.Card)
	}
	nt := must(s.notes.Type(d.TypeID))
	p := must(s.notes.PreviewTemplate(d.TypeID, nt.Templates[0], css, 0, 0, false))
	if !reflect.DeepEqual(p.Fonts, want) {
		t.Fatalf("preview fonts = %+v", p.Fonts)
	}
}

func TestAutoContinueLearningSetting(t *testing.T) {
	s, store := opened(t)
	// A collection without the field reads false.
	if st := must(s.app.Settings()); st.AutoContinueLearning {
		t.Fatalf("default settings = %+v", st)
	}
	restarted := newServices(t, store)
	must(restarted.app.Open("/data"))
	if st := must(restarted.app.Settings()); st.AutoContinueLearning {
		t.Fatalf("settings without the field = %+v", st)
	}
	// Set it, reopen, read it back.
	check(t, restarted.app.SetAutoContinueLearning(true))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); !st.AutoContinueLearning {
		t.Fatalf("settings after restart = %+v", st)
	}
}

func TestLearnAheadSetting(t *testing.T) {
	s, store := opened(t)
	// A collection without the field gets Anki's default.
	if st := must(s.app.Settings()); st.LearnAheadMinutes != 20 {
		t.Fatalf("default settings = %+v", st)
	}
	restarted := newServices(t, store)
	must(restarted.app.Open("/data"))
	if st := must(restarted.app.Settings()); st.LearnAheadMinutes != 20 {
		t.Fatalf("settings without the field = %+v", st)
	}
	for _, bad := range []int{-1, 1441} {
		if err := restarted.app.SetLearnAheadMinutes(bad); err == nil {
			t.Fatalf("limit %d accepted", bad)
		}
	}
	// 0 turns it off and is kept, not replaced by the default.
	check(t, restarted.app.SetLearnAheadMinutes(0))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); st.LearnAheadMinutes != 0 {
		t.Fatalf("settings after restart = %+v", st)
	}
	check(t, again.app.SetLearnAheadMinutes(1440))
	if st := must(again.app.Settings()); st.LearnAheadMinutes != 1440 {
		t.Fatalf("settings = %+v", st)
	}
}

func TestTimeboxSetting(t *testing.T) {
	s, store := opened(t)
	if st := must(s.app.Settings()); st.TimeboxMinutes != 0 {
		t.Fatalf("default settings = %+v", st)
	}
	for _, bad := range []int{-1, 1441} {
		if err := s.app.SetTimeboxMinutes(bad); err == nil {
			t.Fatalf("limit %d accepted", bad)
		}
	}
	// Enable timebox (spec): still 15 minutes after a restart ...
	check(t, s.app.SetTimeboxMinutes(15))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); st.TimeboxMinutes != 15 {
		t.Fatalf("settings after restart = %+v", st)
	}
	// ... and included in backups.
	file := must(again.transfer.ExportBackup())
	other, _ := opened(t)
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	check(t, other.transfer.Restore(info.ID))
	if st := must(other.app.Settings()); st.TimeboxMinutes != 15 {
		t.Fatalf("settings after restore = %+v", st)
	}
	check(t, other.app.SetTimeboxMinutes(0))
	if st := must(other.app.Settings()); st.TimeboxMinutes != 0 {
		t.Fatalf("settings = %+v", st)
	}
}

func TestTimerDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	options := func(s *services) DeckOptions {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n.Options
			}
		}
		t.Fatal("deck not found")
		return DeckOptions{}
	}
	if o := options(s); o.MaxAnswerSeconds != 60 || o.ShowTimer || !o.StopTimerOnAnswer {
		t.Fatalf("default options = %+v", o)
	}
	o := options(s)
	o.MaxAnswerSeconds, o.ShowTimer, o.StopTimerOnAnswer = 30, true, false
	check(t, s.decks.SetOptions(id, o))
	bad := o
	bad.MaxAnswerSeconds = 0
	if err := s.decks.SetOptions(id, bad); err == nil {
		t.Fatal("maximum 0 accepted")
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if o := options(again); o.MaxAnswerSeconds != 30 || !o.ShowTimer || o.StopTimerOnAnswer {
		t.Fatalf("options after restart = %+v", o)
	}
}

// Weekend reduced (spec): the setting persists for the deck.
func TestEasyDaysDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	options := func(s *services) DeckOptions {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n.Options
			}
		}
		t.Fatal("deck not found")
		return DeckOptions{}
	}
	if o := options(s); o.EasyDays != [7]int{} {
		t.Fatalf("default easy days = %v", o.EasyDays)
	}
	o := options(s)
	o.EasyDays[5], o.EasyDays[6] = core.EasyReduced, core.EasyReduced
	check(t, s.decks.SetOptions(id, o))
	bad := o
	bad.EasyDays[0] = -1
	if err := s.decks.SetOptions(id, bad); err == nil {
		t.Fatal("level -1 accepted")
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if got := options(again).EasyDays; got != [7]int{0, 0, 0, 0, 0, 1, 1} {
		t.Fatalf("easy days after restart = %v", got)
	}
}

func TestAutoAdvanceDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	options := func(s *services) DeckOptions {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n.Options
			}
		}
		t.Fatal("deck not found")
		return DeckOptions{}
	}
	if o := options(s); o.SecondsToShowQuestion != 0 || o.SecondsToShowAnswer != 0 ||
		o.QuestionAction != "showAnswer" || o.AnswerAction != "buryCard" {
		t.Fatalf("default options = %+v", o)
	}
	// Options stored (spec): 5 seconds for the question, 3 for the answer,
	// answer Good.
	o := options(s)
	o.SecondsToShowQuestion, o.SecondsToShowAnswer, o.AnswerAction = 5, 3, "answerGood"
	check(t, s.decks.SetOptions(id, o))
	bad := o
	bad.SecondsToShowAnswer = 10000
	if err := s.decks.SetOptions(id, bad); err == nil {
		t.Fatal("10000 seconds accepted")
	}
	bad = o
	bad.AnswerAction = "answerEasy"
	if err := s.decks.SetOptions(id, bad); err == nil {
		t.Fatal("unknown answer action accepted")
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if o := options(again); o.SecondsToShowQuestion != 5 || o.SecondsToShowAnswer != 3 || o.AnswerAction != "answerGood" {
		t.Fatalf("options after restart = %+v", o)
	}
	// ... and included in backups.
	file := must(again.transfer.ExportBackup())
	other, _ := opened(t)
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	check(t, other.transfer.Restore(info.ID))
	for _, n := range must(other.decks.Tree()) {
		if n.FullName == "Spanish" {
			if o := n.Options; o.SecondsToShowQuestion != 5 || o.SecondsToShowAnswer != 3 || o.AnswerAction != "answerGood" {
				t.Fatalf("options after restore = %+v", o)
			}
		}
	}
	// The study card carries its deck's options.
	basicNoteIn(t, again, id, "uno", "one")
	st := must(again.study.Start(id))
	if c := st.Card; c == nil || c.SecondsToShowQuestion != 5 || c.SecondsToShowAnswer != 3 ||
		c.QuestionAction != "showAnswer" || c.AnswerAction != "answerGood" {
		t.Fatalf("card = %+v", st.Card)
	}
}

func TestStudyCardTimerOptions(t *testing.T) {
	s, _ := opened(t)
	parent := must(s.decks.Create("Spanish"))
	child := must(s.decks.Create("Spanish::Verbs"))
	// The subdeck gets its own preset, so its options differ from the
	// parent's.
	must(s.decks.AddPreset(child, "Timed"))
	var o DeckOptions
	for _, n := range must(s.decks.Tree()) {
		if n.ID == child {
			o = n.Options
		}
	}
	o.ShowTimer, o.StopTimerOnAnswer, o.MaxAnswerSeconds = true, false, 20
	check(t, s.decks.SetOptions(child, o))
	basicNoteIn(t, s, child, "ser", "to be")
	// The card shows its own deck's options, also when the parent is
	// studied ...
	st := must(s.study.Start(parent))
	if c := st.Card; c == nil || !c.ShowTimer || c.StopTimerOnAnswer || c.MaxAnswerSeconds != 20 {
		t.Fatalf("card = %+v", st.Card)
	}
	// ... and the answer time is capped at that deck's maximum.
	if r := must(s.study.Answer(st.Card.CardID, 3, 45000)); r.TimeSpent != 20*time.Second {
		t.Fatalf("time spent = %v", r.TimeSpent)
	}
	basicNoteIn(t, s, parent, "uno", "one")
	st = must(s.study.Start(parent))
	if c := st.Card; c == nil || c.ShowTimer || !c.StopTimerOnAnswer || c.MaxAnswerSeconds != 60 {
		t.Fatalf("parent card = %+v", st.Card)
	}
}

func TestReviewSettings(t *testing.T) {
	s, store := opened(t)
	// Defaults, also in collections without the fields.
	if st := must(s.app.Settings()); !st.ShowRemainingCounts || !st.ShowIntervals || !st.SpaceAnswersGood ||
		st.HideBarsDuringReview || st.QuestionFadeMs != 100 {
		t.Fatalf("default settings = %+v", st)
	}
	for _, bad := range []int{-1, 1001} {
		if err := s.app.SetQuestionFadeMs(bad); err == nil {
			t.Fatalf("fade %d accepted", bad)
		}
	}
	// Setting persists (spec): still off after a restart ...
	check(t, s.app.SetShowRemainingCounts(false))
	check(t, s.app.SetShowIntervals(false))
	check(t, s.app.SetSpaceAnswersGood(false))
	check(t, s.app.SetHideBarsDuringReview(true))
	check(t, s.app.SetQuestionFadeMs(0))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	want := func(st Settings) bool {
		return !st.ShowRemainingCounts && !st.ShowIntervals && !st.SpaceAnswersGood &&
			st.HideBarsDuringReview && st.QuestionFadeMs == 0
	}
	if st := must(again.app.Settings()); !want(st) {
		t.Fatalf("settings after restart = %+v", st)
	}
	// ... and included in backups.
	file := must(again.transfer.ExportBackup())
	other, _ := opened(t)
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	check(t, other.transfer.Restore(info.ID))
	if st := must(other.app.Settings()); !want(st) {
		t.Fatalf("settings after restore = %+v", st)
	}
	check(t, other.app.SetShowRemainingCounts(true))
	check(t, other.app.SetQuestionFadeMs(1000))
	if st := must(other.app.Settings()); !st.ShowRemainingCounts || st.QuestionFadeMs != 1000 {
		t.Fatalf("settings = %+v", st)
	}
}

func TestAppearanceSettings(t *testing.T) {
	s, store := opened(t)
	// Defaults, also in collections without the fields.
	if st := must(s.app.Settings()); st.UIScale != 100 || st.Minimalist || st.ReduceMotion != "system" {
		t.Fatalf("default settings = %+v", st)
	}
	for _, bad := range []int{0, 70, 205, 101, -100} {
		if err := s.app.SetUIScale(bad); err == nil {
			t.Fatalf("UI size %d accepted", bad)
		}
	}
	if err := s.app.SetReduceMotion("sometimes"); err == nil {
		t.Fatal("bad reduce motion accepted")
	}
	// Changes persist (spec): still set after a restart ...
	check(t, s.app.SetUIScale(150))
	check(t, s.app.SetMinimalist(true))
	check(t, s.app.SetReduceMotion("on"))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	want := func(st Settings) bool { return st.UIScale == 150 && st.Minimalist && st.ReduceMotion == "on" }
	if st := must(again.app.Settings()); !want(st) {
		t.Fatalf("settings after restart = %+v", st)
	}
	// ... and included in backups.
	file := must(again.transfer.ExportBackup())
	other, _ := opened(t)
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	check(t, other.transfer.Restore(info.ID))
	if st := must(other.app.Settings()); !want(st) {
		t.Fatalf("settings after restore = %+v", st)
	}
	check(t, other.app.SetUIScale(75))
	check(t, other.app.SetUIScale(200))
	check(t, other.app.SetReduceMotion("off"))
	if st := must(other.app.Settings()); st.UIScale != 200 || st.ReduceMotion != "off" {
		t.Fatalf("settings = %+v", st)
	}
	check(t, other.app.SetReduceMotion("system"))
	if st := must(other.app.Settings()); st.ReduceMotion != "system" {
		t.Fatalf("settings = %+v", st)
	}
}

func TestGoalSettings(t *testing.T) {
	s, _ := opened(t)
	// Defaults, also in collections without the fields: no goal, streaks
	// shown.
	if st := must(s.app.Settings()); !st.ShowProgress || st.DailyGoalKind != "" || st.DailyGoal != 0 || st.RestDaysPerWeek != 0 ||
		st.GoalSuggestionShown {
		t.Fatalf("default settings = %+v", st)
	}
	for _, bad := range []struct {
		kind  string
		value int
	}{{"cards", 0}, {"minutes", -5}, {"cards", 10000}, {"reviews", 10}} {
		if err := s.app.SetDailyGoal(bad.kind, bad.value); !core.IsUserError(err) {
			t.Fatalf("goal %s %d: err = %v", bad.kind, bad.value, err)
		}
	}
	check(t, s.app.SetDailyGoal("minutes", 9999))
	check(t, s.app.SetDailyGoal("", 0))
	if st := must(s.app.Settings()); st.DailyGoalKind != "" || st.DailyGoal != 0 {
		t.Fatalf("no goal: %+v", st)
	}
	check(t, s.app.SetDailyGoal("", 25)) // the value is dropped with the kind
	if st := must(s.app.Settings()); st.DailyGoal != 0 {
		t.Fatalf("no goal with a value: %+v", st)
	}
	for _, bad := range []int{-1, 3} {
		if err := s.app.SetRestDaysPerWeek(bad); !core.IsUserError(err) {
			t.Fatalf("%d rest days: err = %v", bad, err)
		}
	}
	check(t, s.app.SetShowProgress(false))
	check(t, s.app.GoalSuggestionShown())
	// Persisted in backups (spec): a goal of 30 cards and 2 rest days per
	// week survive a restore on another installation.
	check(t, s.app.SetDailyGoal("cards", 30))
	check(t, s.app.SetRestDaysPerWeek(2))
	file := must(s.transfer.ExportBackup())
	other, _ := opened(t)
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	check(t, other.transfer.Restore(info.ID))
	if st := must(other.app.Settings()); st.DailyGoalKind != "cards" || st.DailyGoal != 30 || st.RestDaysPerWeek != 2 {
		t.Fatalf("settings after restore = %+v", st)
	}
}

func TestAudioSettings(t *testing.T) {
	s, store := opened(t)
	// Both are on by default, also in collections without the fields.
	if st := must(s.app.Settings()); !st.ShowPlayButtons || !st.InterruptAudio {
		t.Fatalf("default settings = %+v", st)
	}
	check(t, s.app.SetShowPlayButtons(false))
	check(t, s.app.SetInterruptAudio(false))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); st.ShowPlayButtons || st.InterruptAudio {
		t.Fatalf("settings after restart = %+v", st)
	}
	check(t, again.app.SetInterruptAudio(true))
	if st := must(again.app.Settings()); st.ShowPlayButtons || !st.InterruptAudio {
		t.Fatalf("settings = %+v", st)
	}
}

func TestAudioDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	options := func(s *services) DeckOptions {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n.Options
			}
		}
		t.Fatal("deck not found")
		return DeckOptions{}
	}
	if o := options(s); o.DisableAutoplay || o.SkipQuestionOnReplay {
		t.Fatalf("default options = %+v", o)
	}
	o := options(s)
	o.DisableAutoplay = true
	check(t, s.decks.SetOptions(id, o))
	// Options persist (collection spec): still on after a restart ...
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if o := options(again); !o.DisableAutoplay || o.SkipQuestionOnReplay {
		t.Fatalf("options after restart = %+v", o)
	}
	// ... and included in backups.
	file := must(again.transfer.ExportBackup())
	other, _ := opened(t)
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	check(t, other.transfer.Restore(info.ID))
	if o := options(other); !o.DisableAutoplay {
		t.Fatalf("options after restore = %+v", o)
	}
}

func TestStepDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	options := func(s *services) DeckOptions {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n.Options
			}
		}
		t.Fatal("deck not found")
		return DeckOptions{}
	}
	o := options(s)
	if !slices.Equal(o.LearningSteps, []int{60, 600}) || !slices.Equal(o.RelearningSteps, []int{600}) ||
		o.InterdayLearningOrder != "mix" || o.BuryInterdayLearning || o.BuryNew || o.BuryReview {
		t.Fatalf("default options = %+v", o)
	}
	o.LearningSteps = []int{3600}
	o.RelearningSteps = []int{}
	o.InterdayLearningOrder = "after"
	o.BuryInterdayLearning = true
	o.BuryNew, o.BuryReview = true, true
	check(t, s.decks.SetOptions(id, o))
	bad := o
	bad.LearningSteps = []int{-1}
	if err := s.decks.SetOptions(id, bad); err == nil {
		t.Fatal("negative step accepted")
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if got := options(again); !reflect.DeepEqual(got, o) {
		t.Fatalf("options after restart = %+v, want %+v", got, o)
	}
	// The deck's steps drive the rating buttons: Again repeats the 1h step.
	basicNoteIn(t, again, id, "uno", "one")
	st := must(again.study.Start(id))
	if again := core.FormatInterval(st.Card.Previews[0].Interval); again != "1h" && again != "1.1h" {
		t.Fatalf("Again label = %q", again)
	}
}

func TestIntervalDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	var o DeckOptions
	for _, n := range must(s.decks.Tree()) {
		if n.ID == id {
			o = n.Options
		}
	}
	if o.MaxInterval != 36500 || o.MinRelearnInterval != 1 {
		t.Fatalf("default options = %+v", o)
	}
	// Minimum above maximum (spec): rejected with an explanation.
	bad := o
	bad.MaxInterval, bad.MinRelearnInterval = 30, 60
	if err := s.decks.SetOptions(id, bad); err == nil || !strings.Contains(err.Error(), "maximum interval") {
		t.Fatalf("err = %v", err)
	}
	o.MaxInterval = 1
	check(t, s.decks.SetOptions(id, o))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	for _, n := range must(again.decks.Tree()) {
		if n.ID == id && n.Options.MaxInterval != 1 {
			t.Fatalf("options after restart = %+v", n.Options)
		}
	}
	// The interval preview reflects the cap: Easy on a new card is 1 day.
	basicNoteIn(t, again, id, "uno", "one")
	st := must(again.study.Start(id))
	if easy := core.FormatInterval(st.Card.Previews[3].Interval); easy != "1d" {
		t.Fatalf("Easy label = %q", easy)
	}
}

func TestStudyInterdayStepAcrossRollover(t *testing.T) {
	// Studying with steps 1m 1d across the 04:00 rollover.
	s, _ := opened(t)
	id := must(s.decks.Create("Spanish"))
	o := DeckOptions{}
	for _, n := range must(s.decks.Tree()) {
		if n.ID == id {
			o = n.Options
		}
	}
	o.LearningSteps = []int{60, 86400}
	check(t, s.decks.SetOptions(id, o))
	basicNoteIn(t, s, id, "uno", "one")
	st := must(s.study.Start(id))
	if st.Card == nil || core.FormatInterval(st.Card.Previews[2].Interval) != "1d" {
		t.Fatalf("start = %+v", st.Card)
	}
	st = must(s.study.Answer(st.Card.CardID, 3, 1000))
	tomorrow := time.Date(2026, 10, 14, 4, 0, 0, 0, time.Local)
	if st.Card != nil || st.Counts != (Counts{}) || st.NextDue == nil || !st.NextDue.Equal(tomorrow) {
		t.Fatalf("after Good = %+v", st)
	}
	*s.clock = tomorrow.Add(-time.Minute)
	if st := must(s.study.Start(id)); st.Card != nil {
		t.Fatalf("before rollover = %+v", st)
	}
	*s.clock = tomorrow
	st = must(s.study.Start(id))
	if st.Card == nil || st.Card.Kind != "learning" || st.Counts != (Counts{Learning: 1}) {
		t.Fatalf("after rollover = %+v", st)
	}
	st = must(s.study.Answer(st.Card.CardID, 3, 1000))
	if st.Card != nil || st.Counts != (Counts{}) {
		t.Fatalf("after graduating = %+v", st)
	}
}

func TestStudyNewAfterReviews(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	var o DeckOptions
	for _, n := range must(s.decks.Tree()) {
		if n.ID == id {
			o = n.Options
		}
	}
	if o.NewGatherOrder != "ascending" || o.NewSortOrder != "gathered" || o.NewReviewOrder != "mix" ||
		o.ReviewSortOrder != "dueRandom" {
		t.Fatalf("default options = %+v", o)
	}
	// Two reviews due a month later, and two new cards.
	basicNoteIn(t, s, id, "uno", "one")
	basicNoteIn(t, s, id, "dos", "two")
	for st := must(s.study.Start(id)); st.Card != nil; {
		st = must(s.study.Answer(st.Card.CardID, 4, 1000))
	}
	*s.clock = s.clock.AddDate(0, 1, 0)
	basicNoteIn(t, s, id, "tres", "three")
	basicNoteIn(t, s, id, "cuatro", "four")
	bad := o
	bad.NewReviewOrder = "first"
	if err := s.decks.SetOptions(id, bad); err == nil || !strings.Contains(err.Error(), "new/review order") {
		t.Fatalf("err = %v", err)
	}
	o.NewReviewOrder = "after"
	check(t, s.decks.SetOptions(id, o))
	again := newServices(t, store)
	*again.clock = *s.clock
	must(again.app.Open("/data"))
	var kinds []string
	for st := must(again.study.Start(id)); st.Card != nil; {
		kinds = append(kinds, st.Card.Kind)
		st = must(again.study.Answer(st.Card.CardID, 4, 1000))
	}
	if strings.Join(kinds, " ") != "review review new new" {
		t.Fatalf("kinds = %v", kinds)
	}
}

func TestDecksService(t *testing.T) {
	s, _ := opened(t)
	id := must(s.decks.Create("Languages::Spanish"))
	must(s.decks.Create("Languages::French"))
	tree := must(s.decks.Tree())
	var names []string
	for _, n := range tree {
		names = append(names, strings.Repeat("  ", n.Depth)+n.Name)
	}
	if strings.Join(names, "|") != "Default|Languages|  French|  Spanish" {
		t.Fatalf("tree = %q", names)
	}
	if !tree[1].HasChildren || tree[3].ParentID != tree[1].ID || tree[3].ID != id {
		t.Fatalf("tree links = %+v", tree)
	}
	if err := s.decks.SetOptions(id, DeckOptions{NewPerDay: 5, ReviewsPerDay: 9, DesiredRetention: 0.5}); err == nil {
		t.Fatal("retention 0.5 accepted")
	}
	if err := s.decks.SetCollapsed(tree[1].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := s.decks.Rename(id, "Languages::Español"); err != nil {
		t.Fatal(err)
	}
	basicNoteIn(t, s, id, "uno", "one")
	if tree := must(s.decks.Tree()); tree[1].Cards != 1 || tree[2].Cards != 1 || tree[3].Cards != 0 || tree[0].Cards != 0 {
		t.Fatalf("card totals = %+v", tree)
	}
	if info := must(s.decks.DeletionInfo(tree[1].ID)); info.Decks != 3 || info.Cards != 1 {
		t.Fatalf("info = %+v", info)
	}
	if err := s.decks.Delete(tree[1].ID); err != nil {
		t.Fatal(err)
	}
	if n := len(must(s.decks.Tree())); n != 1 {
		t.Fatalf("decks left = %d", n)
	}
}

func TestDecksNextChange(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	st := must(s.study.Start(deck))
	id := st.Card.CardID
	must(s.study.Answer(id, 1, 1000))
	due := must(s.browse.Card(id)).Due
	if next := must(s.decks.NextChange()); due == nil || !next.Equal(*due) {
		t.Fatalf("next change = %v, want card due %v", next, due)
	}
}

func basicNoteIn(t *testing.T, s *services, deck int64, front, back string) int64 {
	t.Helper()
	d := must(s.notes.Defaults(0))
	return must(s.notes.Add(NewNote{TypeID: d.TypeID, DeckID: deck, Fields: []string{front, back}}))
}

func basicNote(t *testing.T, s *services, front, back string, tags ...string) int64 {
	t.Helper()
	d := must(s.notes.Defaults(0))
	return must(s.notes.Add(NewNote{TypeID: d.TypeID, DeckID: d.DeckID, Fields: []string{front, back}, Tags: tags}))
}

func TestNotesService(t *testing.T) {
	s, _ := opened(t)
	types := must(s.notes.Types())
	if len(types) != 4 || types[0].Name != "Basic" || !types[2].Cloze || types[2].ImageOcclusion ||
		types[3].Name != "Image Occlusion" || !types[3].Cloze || !types[3].ImageOcclusion {
		t.Fatalf("types = %+v", types)
	}
	if id := must(s.notes.ImageOcclusionType()); id != types[3].ID {
		t.Fatalf("ImageOcclusionType = %d, want %d", id, types[3].ID)
	}
	d := must(s.notes.Defaults(0))
	if d.TypeID != types[0].ID {
		t.Fatalf("defaults = %+v", d)
	}
	id := basicNote(t, s, "hola", "hello", "verb irregular")
	n := must(s.notes.Get(id))
	if strings.Join(n.Tags, ",") != "verb,irregular" || len(n.CardIDs) != 1 {
		t.Fatalf("note = %+v", n)
	}
	if _, err := s.notes.Add(NewNote{TypeID: d.TypeID, DeckID: d.DeckID, Fields: []string{"", "x"}}); err == nil ||
		!strings.Contains(err.Error(), "no cards") {
		t.Fatalf("empty note: %v", err)
	}
	if err := s.notes.Update(id, []string{"hola!", "hello"}, nil); err != nil {
		t.Fatal(err)
	}
	if tags := must(s.notes.Tags()); len(tags) != 0 {
		t.Fatalf("tags = %v", tags)
	}
	// Tag management.
	if err := s.notes.Update(id, []string{"hola!", "hello"}, []string{"verb::irregular", "old"}); err != nil {
		t.Fatal(err)
	}
	if n := must(s.notes.RenameTag("verb", "verbs")); n != 1 {
		t.Fatalf("renamed %d", n)
	}
	if n := must(s.notes.MoveTags([]string{"verbs::irregular"}, "")); n != 1 {
		t.Fatalf("moved %d", n)
	}
	if n := must(s.notes.TagUsage([]string{"old"})); n != 1 {
		t.Fatalf("usage = %d", n)
	}
	if n := must(s.notes.DeleteTags([]string{"old"})); n != 1 {
		t.Fatalf("deleted %d", n)
	}
	if n := must(s.notes.ClearUnusedTags()); n != 0 {
		t.Fatalf("cleared %d", n)
	}
	if tags := must(s.notes.Tags()); strings.Join(tags, ",") != "irregular" {
		t.Fatalf("tags = %v", tags)
	}
	if err := s.notes.Update(id, []string{"hola!", "hello"}, nil); err != nil {
		t.Fatal(err)
	}
	// Note type manager.
	if err := s.notes.AddField(d.TypeID, "Notes"); err != nil {
		t.Fatal(err)
	}
	if err := s.notes.RenameField(d.TypeID, 2, "Extra"); err != nil {
		t.Fatal(err)
	}
	if err := s.notes.AddTemplate(d.TypeID, CardTemplate{Name: "Reverse", QFmt: "{{Back}}", AFmt: "{{Front}}"}); err != nil {
		t.Fatal(err)
	}
	if n := must(s.notes.TemplateCardCount(d.TypeID, 1)); n != 1 {
		t.Fatalf("template cards = %d", n)
	}
	if err := s.notes.RemoveTemplate(d.TypeID, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.notes.RemoveField(d.TypeID, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.notes.SetCSS(d.TypeID, ".card{}"); err != nil {
		t.Fatal(err)
	}
	if err := s.notes.RenameType(d.TypeID, "Vocab"); err != nil {
		t.Fatal(err)
	}
	if nt := must(s.notes.Type(d.TypeID)); nt.Name != "Vocab" || len(nt.Fields) != 2 || nt.NoteCount != 1 {
		t.Fatalf("type = %+v", nt)
	}
	if err := s.notes.Delete([]int64{id}); err != nil {
		t.Fatal(err)
	}
}

func TestNotesDuplicates(t *testing.T) {
	s, _ := opened(t)
	types := must(s.notes.Types())
	basic, cloze := types[0].ID, types[2].ID
	deck := must(s.notes.Defaults(basic)).DeckID
	add := func(typeID int64, fields ...string) int64 {
		return must(s.notes.Add(NewNote{TypeID: typeID, DeckID: deck, Fields: fields}))
	}
	perro := add(basic, "perro", "dog")
	add(cloze, "gato {{c1::x}}")

	// Duplicate word (spec).
	if got := must(s.notes.Duplicates(basic, "Perro ", 0)); !reflect.DeepEqual(got, []int64{perro}) {
		t.Fatalf("duplicates = %v", got)
	}
	// Different note type (spec).
	if got := must(s.notes.Duplicates(basic, "gato {{c1::x}}", 0)); len(got) != 0 {
		t.Fatalf("other note type = %v", got)
	}
	// The edited note and empty fields.
	if got := must(s.notes.Duplicates(basic, "perro", perro)); len(got) != 0 {
		t.Fatalf("excluded = %v", got)
	}
	if got := must(s.notes.Duplicates(basic, "", 0)); len(got) != 0 {
		t.Fatalf("empty = %v", got)
	}
	// Add anyway (spec).
	again := add(basic, "Perro", "dog")
	if got := must(s.notes.Duplicates(basic, "perro", 0)); !reflect.DeepEqual(got, []int64{perro, again}) {
		t.Fatalf("after adding anyway = %v", got)
	}
}

func TestOcclusionNoteCards(t *testing.T) {
	s, _ := opened(t)
	d := must(s.notes.Defaults(0))
	typeID := must(s.notes.ImageOcclusionType())
	// An occlusion note's cards by number, which the mask editor does not
	// reuse: a removed shape leaves an empty card.
	rect := func(n string) string { return "{{c" + n + "::image-occlusion:rect:left=.1:top=.1:width=.1:height=.1}}" }
	io := must(s.notes.Add(NewNote{TypeID: typeID, DeckID: d.DeckID, Fields: []string{rect("1") + rect("3"), `<img src="a.png">`}}))
	if n := must(s.notes.Get(io)); !slices.Equal(n.CardOrds, []int{0, 2}) || len(n.CardIDs) != 2 {
		t.Fatalf("occlusion note = %+v", n)
	}
	check(t, s.notes.Update(io, []string{rect("3"), `<img src="a.png">`}, nil))
	if n := must(s.notes.Get(io)); !slices.Equal(n.CardOrds, []int{0, 2}) {
		t.Fatalf("after removing shape 1: %+v", n)
	}
}

func TestNoteTypeManagementService(t *testing.T) {
	s, _ := opened(t)
	if got := strings.Join(s.notes.StockTypes(), ", "); got !=
		"Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer), Cloze, Image Occlusion" {
		t.Fatalf("stock = %q", got)
	}
	vocab := must(s.notes.AddType("Basic (type in the answer)", "Vocabulary"))
	if nt := must(s.notes.Type(vocab)); nt.Name != "Vocabulary" || nt.NoteCount != 0 ||
		nt.Templates[0].QFmt != "{{Front}}\n\n{{type:Back}}" {
		t.Fatalf("added = %+v", nt)
	}
	if _, err := s.notes.AddType("Basic", "vocabulary"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("taken name: %v", err)
	}
	clone := must(s.notes.CloneType(vocab, "Vocabulary copy"))
	if nt := must(s.notes.Type(clone)); nt.Name != "Vocabulary copy" || len(nt.Fields) != 2 {
		t.Fatalf("clone = %+v", nt)
	}
	d := must(s.notes.Defaults(0))
	must(s.notes.Add(NewNote{TypeID: clone, DeckID: d.DeckID, Fields: []string{"perro", "dog"}}))
	if info := must(s.notes.TypeDeletionInfo(clone)); info != (NoteTypeDeletionInfo{Notes: 1, Cards: 1}) {
		t.Fatalf("deletion info = %+v", info)
	}
	check(t, s.notes.DeleteType(clone))
	if len(must(s.notes.Types())) != 5 {
		t.Fatalf("types = %+v", must(s.notes.Types()))
	}
	// The remembered note type was deleted: the editor falls back to Basic.
	if d := must(s.notes.Defaults(0)); d.TypeID != must(s.notes.Types())[0].ID {
		t.Fatalf("defaults = %+v", d)
	}
	if _, err := s.notes.TypeDeletionInfo(clone); err == nil {
		t.Fatal("deletion info of a deleted note type")
	}
}

func TestStudyService(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	// Nothing due: no card, and nothing scheduled.
	st := must(s.study.Start(deck))
	if st.Card != nil || st.NextDue != nil || st.Studied != 0 {
		t.Fatalf("empty deck = %+v", st)
	}
	basicNote(t, s, "uno", "one")
	basicNote(t, s, "dos", "two")
	st = must(s.study.Start(deck))
	if st.Card == nil || st.Counts.New != 2 || len(st.Card.Previews) != 4 || st.Card.Kind != "new" {
		t.Fatalf("start = %+v", st)
	}
	if good := core.FormatInterval(st.Card.Previews[2].Interval); !slices.Contains([]string{"<10m", "<11m", "<12m", "<13m"}, good) || !strings.Contains(st.Card.Question, "uno") {
		t.Fatalf("card = %+v", st.Card)
	}
	first := st.Card.CardID

	// Answer Easy: next card, counts decrease.
	st = must(s.study.Answer(first, 4, 3000))
	if st.Card == nil || st.Card.CardID == first || st.Counts.New != 1 || st.Studied != 1 || !st.CanUndo {
		t.Fatalf("after answer = %+v", st)
	}
	// Undo: the first card is shown again.
	st = must(s.study.Undo())
	if st.Card == nil || st.Card.CardID != first || st.Studied != 0 || st.CanUndo || st.Counts.New != 2 {
		t.Fatalf("after undo = %+v", st)
	}
	if _, err := s.study.Undo(); err == nil {
		t.Fatal("undo with empty stack")
	}
	// Edit while studying: the corrected content is shown.
	card := must(s.browse.Card(first))
	if err := s.notes.Update(card.Note.ID, []string{"uno (1)", "one"}, nil); err != nil {
		t.Fatal(err)
	}
	if st = must(s.study.Current()); st.Card.CardID != first || !strings.Contains(st.Card.Question, "uno (1)") {
		t.Fatalf("after edit = %+v", st.Card)
	}
	// Flag the current card: it stays on screen; the same flag toggles
	// it off, another flag replaces it.
	if st = must(s.study.SetFlag(first, 1)); st.Card.CardID != first || st.Card.Flag != 1 {
		t.Fatalf("after flag = %+v", st.Card)
	}
	if st = must(s.study.SetFlag(first, 3)); st.Card.CardID != first || st.Card.Flag != 3 {
		t.Fatalf("after change flag = %+v", st.Card)
	}
	if st = must(s.study.SetFlag(first, 3)); st.Card.CardID != first || st.Card.Flag != 0 {
		t.Fatalf("after toggle = %+v", st.Card)
	}
	if _, err := s.study.SetFlag(first, 9); err == nil {
		t.Fatal("flag 9 accepted")
	}
	// Mark the current card's note: it stays on screen with the star; the
	// same toggle unmarks it.
	if st = must(s.study.ToggleMark(first)); st.Card.CardID != first || !st.Card.Marked {
		t.Fatalf("after mark = %+v", st.Card)
	}
	if n := len(must(s.browse.Search("tag:marked", "created", false))); n != 1 {
		t.Fatalf("marked = %d", n)
	}
	if st = must(s.study.ToggleMark(first)); st.Card.CardID != first || st.Card.Marked {
		t.Fatalf("after unmark = %+v", st.Card)
	}
	// Bury and suspend move on.
	st = must(s.study.Bury(first))
	if st.Card == nil || st.Card.CardID == first {
		t.Fatalf("after bury = %+v", st)
	}
	st = must(s.study.Suspend(st.Card.CardID))
	if st.Card != nil {
		t.Fatalf("after suspend = %+v", st)
	}
	// Nothing due now; the buried card returns tomorrow.
	if st.NextDue == nil {
		t.Fatalf("next due missing: %+v", st)
	}
	// Learning card due later today, with learning ahead off: the summary
	// says when, and the screen checks again then.
	check(t, s.app.SetLearnAheadMinutes(0))
	basicNote(t, s, "tres", "three")
	st = must(s.study.Start(deck))
	st = must(s.study.Answer(st.Card.CardID, 3, 1000))
	if st.Card != nil || st.LearningDue == nil || st.Studied != 1 || st.TimeSpent != time.Second {
		t.Fatalf("summary = %+v", st)
	}
	if st.CheckAt == nil || !st.CheckAt.Equal(*st.LearningDue) {
		t.Fatalf("check at = %v, learning due = %v", st.CheckAt, st.LearningDue)
	}
}

func TestRecentDecks(t *testing.T) {
	s, store := opened(t)
	if r := must(s.app.Settings()).RecentDecks; len(r) != 0 {
		t.Fatalf("new collection = %v", r)
	}
	var decks []int64
	for i := range core.MaxRecentDecks + 2 {
		decks = append(decks, must(s.decks.Create(fmt.Sprintf("Deck %d", i))))
	}
	// Starting a session puts the deck first; studying it again moves it
	// to the front instead of listing it twice.
	must(s.study.Start(decks[0]))
	must(s.study.Start(decks[1]))
	must(s.study.Start(decks[0]))
	if r := must(s.app.Settings()).RecentDecks; !slices.Equal(r, []int64{decks[0], decks[1]}) {
		t.Fatalf("recent = %v", r)
	}
	// At most MaxRecentDecks, latest first.
	for _, id := range decks {
		must(s.study.Start(id))
	}
	want := slices.Clone(decks[2:])
	slices.Reverse(want)
	if r := must(s.app.Settings()).RecentDecks; !slices.Equal(r, want) {
		t.Fatalf("recent = %v, want %v", r, want)
	}
	// Deleted decks are left out, and come back when the deletion is
	// undone.
	last := decks[len(decks)-1]
	check(t, s.decks.Delete(last))
	if r := must(s.app.Settings()).RecentDecks; !slices.Equal(r, want[1:]) {
		t.Fatalf("after delete = %v", r)
	}
	must(s.app.Undo())
	if r := must(s.app.Settings()).RecentDecks; !slices.Equal(r, want) {
		t.Fatalf("after undo = %v", r)
	}
	// Saved with the collection.
	check(t, s.decks.Delete(last))
	must(s.study.Start(decks[2]))
	s2 := newServices(t, store)
	must(s2.app.Open("/data"))
	want = append([]int64{decks[2]}, slices.DeleteFunc(want[1:], func(id int64) bool { return id == decks[2] })...)
	if r := must(s2.app.Settings()).RecentDecks; !slices.Equal(r, want) {
		t.Fatalf("after reopen = %v, want %v", r, want)
	}
}

func TestStudyNoteActions(t *testing.T) {
	s, _ := opened(t)
	types := must(s.notes.Types()) // Basic, Basic (and reversed card), Cloze
	deck := must(s.notes.Defaults(0)).DeckID
	reversed := func(front, back string) []int64 {
		id := must(s.notes.Add(NewNote{TypeID: types[1].ID, DeckID: deck, Fields: []string{front, back}}))
		return must(s.notes.Get(id)).CardIDs
	}
	a := reversed("uno", "one")
	b := reversed("dos", "two")
	c := reversed("tres", "three")
	must(s.notes.Add(NewNote{TypeID: types[0].ID, DeckID: deck, Fields: []string{"cuatro", "four"}}))
	st := must(s.study.Start(deck))
	if st.Card == nil || st.Card.CardID != a[0] || st.Counts.New != 7 {
		t.Fatalf("start = %+v", st)
	}

	// Suspend note on the forward card: both cards are suspended, the next
	// card is shown and the new count drops by two.
	st = must(s.study.SuspendNote(a[0]))
	if st.Card == nil || st.Card.NoteID == must(s.browse.Card(a[0])).Note.ID || st.Counts.New != 5 {
		t.Fatalf("after suspend note = %+v", st)
	}
	for _, id := range a {
		if card := must(s.browse.Card(id)); !card.Suspended {
			t.Fatalf("card %d not suspended: %+v", id, card)
		}
	}

	// Bury note: both cards are buried until tomorrow.
	st = must(s.study.BuryNote(b[0]))
	if st.Card == nil || st.Card.CardID == b[0] || st.Card.CardID == b[1] || st.Counts.New != 3 {
		t.Fatalf("after bury note = %+v", st)
	}
	if ids := must(s.browse.Search("is:buried", "created", false)); !slices.Equal(ids, b) {
		t.Fatalf("buried = %v, want %v", ids, b)
	}

	// Delete note: the note and both its cards are gone.
	note := must(s.browse.Card(c[0])).Note.ID
	st = must(s.study.DeleteNote(c[1]))
	if st.Card == nil || st.Card.NoteID == note || st.Counts.New != 1 {
		t.Fatalf("after delete note = %+v", st)
	}
	if _, err := s.notes.Get(note); err == nil {
		t.Fatal("deleted note still exists")
	}
	for _, id := range c {
		if _, err := s.browse.Card(id); err == nil {
			t.Fatalf("card %d still exists", id)
		}
	}
	if _, err := s.study.DeleteNote(c[0]); err == nil {
		t.Fatal("deleting a gone card's note succeeded")
	}
	if _, err := s.study.SuspendNote(c[0]); err == nil {
		t.Fatal("suspending a gone card's note succeeded")
	}
}

// Reset from study (study-session spec): the card is New again in its
// original position, and the study screen moves on.
func TestStudyReset(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	var ids []int64
	for _, w := range []string{"uno", "dos", "tres"} {
		ids = append(ids, must(s.notes.Get(basicNote(t, s, w, w))).CardIDs[0])
	}
	st := must(s.study.Start(deck))
	st = must(s.study.Answer(ids[0], 4, 1000))
	st = must(s.study.Answer(ids[1], 4, 1000))
	if st.Card == nil || st.Card.CardID != ids[2] || st.Counts.New != 1 {
		t.Fatalf("before reset = %+v", st)
	}
	// Reset the second card: it is New again before the third.
	st = must(s.study.Reset(ids[1], ResetOptions{RestorePosition: true}))
	if st.Card == nil || st.Card.CardID != ids[1] || st.Counts.New != 2 {
		t.Fatalf("after reset = %+v", st)
	}
	if got := must(s.browse.Search("is:new", "created", false)); !slices.Equal(got, ids[1:]) {
		t.Fatalf("new cards = %v", got)
	}
}

// Deck options changed during a session apply to the rest of it.
func TestStudyOptionsDuringSession(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	for _, w := range []string{"uno", "dos", "tres"} {
		basicNote(t, s, w, w)
	}
	st := must(s.study.Start(deck))
	st = must(s.study.Answer(st.Card.CardID, 4, 1000))
	if st.Card == nil || st.Counts.New != 2 {
		t.Fatalf("after one answer = %+v", st)
	}
	// Lower new cards per day below the one already introduced today.
	set := must(s.decks.Settings(deck))
	o := set.Presets[0].Options
	o.NewPerDay = 0
	check(t, s.decks.SaveSettings(deck, set.PresetID, o, set.Limits))
	if st = must(s.study.Current()); st.Card != nil || st.Counts.New != 0 {
		t.Fatalf("after lowering the limit = %+v", st)
	}
}

func TestStudyStartSetsCurrentDeck(t *testing.T) {
	s, _ := opened(t)
	french := must(s.decks.Create("French"))
	verbs := must(s.decks.Create("French::Verbs"))
	spanish := must(s.decks.Create("Spanish"))
	basicNoteIn(t, s, french, "chat", "cat")
	basicNoteIn(t, s, verbs, "aller", "to go")
	basicNoteIn(t, s, spanish, "gato", "cat")
	// No deck studied yet: deck:current matches nothing.
	if ids := must(s.browse.Search("deck:current", "created", false)); len(ids) != 0 {
		t.Fatalf("before studying = %v", ids)
	}
	must(s.study.Start(french))
	if ids := must(s.browse.Search("deck:current", "created", false)); len(ids) != 2 {
		t.Fatalf("after studying French = %v", ids)
	}
	must(s.study.Start(spanish))
	if ids := must(s.browse.Search("deck:current", "created", false)); len(ids) != 1 {
		t.Fatalf("after studying Spanish = %v", ids)
	}
}

func TestStudyTypeAnswer(t *testing.T) {
	s, _ := opened(t)
	d := must(s.notes.Defaults(0))
	basicNote(t, s, "dog", "perro")
	st := must(s.study.Start(d.DeckID))
	if st.Card == nil || st.Card.TypeAnswer {
		t.Fatalf("plain Basic card = %+v", st.Card)
	}
	check(t, s.notes.UpdateTemplate(d.TypeID, 0, CardTemplate{Name: "Card 1",
		QFmt: "{{Front}} {{type:Back}}", AFmt: "{{FrontSide}}<hr id=answer>{{type:Back}}"}))
	st = must(s.study.Current())
	if !st.Card.TypeAnswer || !strings.Contains(st.Card.Question, "<cf-typeans></cf-typeans>") {
		t.Fatalf("type-in card = %+v", st.Card)
	}
	// Without Reveal the answer shows the expected text.
	if !strings.Contains(st.Card.Answer, "perro") || strings.Contains(st.Card.Answer, "typeans") {
		t.Fatalf("answer = %q", st.Card.Answer)
	}
	card := must(s.study.Reveal(st.Card.CardID, "pero"))
	if card.CardID != st.Card.CardID || !card.TypeAnswer || len(card.Previews) != 4 ||
		!strings.Contains(card.Answer, `<span class="typeMissed">r</span>`) || strings.Contains(card.Answer, "cf-typeans") {
		t.Fatalf("revealed = %+v", card)
	}
	card = must(s.study.Reveal(st.Card.CardID, "perro"))
	if !strings.Contains(card.Answer, `<code id="typeans"><span class="typeGood">perro</span></code>`) {
		t.Fatalf("exact match = %q", card.Answer)
	}
	// Revealing does not answer: the card is still on screen.
	if st = must(s.study.Current()); st.Card.CardID != card.CardID || st.Studied != 0 {
		t.Fatalf("after reveal = %+v", st)
	}
}

func TestStudyLearningDueWithCard(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	st := must(s.study.Start(deck))
	review := st.Card.CardID
	must(s.study.Answer(review, 4, 1000)) // Easy: a review card, due in days
	*s.clock = s.clock.AddDate(0, 2, 0)
	basicNote(t, s, "dos", "two")
	st = must(s.study.Start(deck))
	if st.Card == nil || st.Card.Kind != "new" {
		t.Fatalf("start = %+v", st.Card)
	}
	// Again: a learning card pending within the learn ahead limit, the
	// review card on screen. The screen checks again when the learning
	// card becomes due, to count it.
	learning := st.Card.CardID
	st = must(s.study.Answer(learning, 1, 1000))
	if st.Card == nil || st.Card.CardID != review || st.Card.Kind != "review" {
		t.Fatalf("card = %+v", st.Card)
	}
	if st.LearningDue != nil || st.CheckAt == nil || !st.CheckAt.After(*s.clock) {
		t.Fatalf("learning due = %v, check at = %v", st.LearningDue, st.CheckAt)
	}
	// Learning ahead off: the same card is now beyond the limit.
	check(t, s.app.SetLearnAheadMinutes(0))
	st = must(s.study.Current())
	if st.LearningDue == nil || !st.LearningDue.After(*s.clock) || !st.CheckAt.Equal(*st.LearningDue) {
		t.Fatalf("learning due = %v, check at = %v", st.LearningDue, st.CheckAt)
	}
}

func TestStudyLearnAhead(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	opts := DeckOptions(core.DefaultDeckOptions())
	opts.LearningSteps = []int{30 * 60}
	check(t, s.decks.SetOptions(deck, opts))
	basicNote(t, s, "uno", "one")
	basicNote(t, s, "dos", "two")
	st := must(s.study.Start(deck))
	first := st.Card.CardID
	// Again: due in 30 minutes, on the only step.
	st = must(s.study.Answer(first, 1, 1000))
	second := st.Card.CardID
	*s.clock = s.clock.Add(15 * time.Minute)
	// Finish without waiting (study-session spec): after the last due card
	// (Easy, now a review card), the first card is due in 15 minutes,
	// within the limit of 20, and is shown right away.
	st = must(s.study.Answer(second, 4, 1000))
	if st.Card == nil || st.Card.CardID != first || st.Card.Kind != "learning" {
		t.Fatalf("after the last due card = %+v", st)
	}
	// Not due yet: not counted, and nothing beyond the limit.
	if st.Counts.Learning != 0 || st.LearningDue != nil || st.CheckAt == nil || !st.CheckAt.After(*s.clock) {
		t.Fatalf("early card state = %+v", st)
	}
	answered := *s.clock

	// Reacting at the limit (study-session spec): Again schedules the step
	// from the actual time of the answer, 30 minutes from now, beyond the
	// limit. The summary says when it is due, and the screen checks again
	// 20 minutes before.
	st = must(s.study.Answer(first, 1, 1000))
	if st.Card != nil || st.LearningDue == nil || st.CheckAt == nil {
		t.Fatalf("summary = %+v", st)
	}
	if !st.LearningDue.After(answered.Add(20 * time.Minute)) {
		t.Fatalf("learning due = %v, answered at %v", st.LearningDue, answered)
	}
	if want := st.LearningDue.Add(-20 * time.Minute); !st.CheckAt.Equal(want) {
		t.Fatalf("check at = %v, want %v", st.CheckAt, want)
	}
	checkAt := *st.CheckAt
	*s.clock = checkAt.Add(-time.Second)
	if st = must(s.study.Current()); st.Card != nil {
		t.Fatalf("before the limit = %+v", st.Card)
	}
	*s.clock = checkAt
	if st = must(s.study.Current()); st.Card == nil || st.Card.CardID != first {
		t.Fatalf("at the limit = %+v", st)
	}
}

func TestStudyDueEarly(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	st := must(s.study.Start(deck))
	if st.Card == nil || st.Card.DueEarly != nil {
		t.Fatalf("new card = %+v", st.Card)
	}
	// Good: due in 10 minutes on the second step, and with nothing else due
	// shown again right away (Card shown again, study-session spec).
	card := st.Card.CardID
	answered := *s.clock
	st = must(s.study.Answer(card, 3, 1000))
	if st.Card == nil || st.Card.CardID != card || st.Card.DueEarly == nil {
		t.Fatalf("after Good = %+v", st.Card)
	}
	// The step is fuzzed by up to 25 %.
	due := must(s.browse.CardInfo(card)).Due
	if left := st.Card.DueEarly.Sub(answered); due == nil || !st.Card.DueEarly.Equal(*due) || left < 10*time.Minute || left > 12*time.Minute+30*time.Second {
		t.Fatalf("due early = %v, card due %v, answered at %v", st.Card.DueEarly, due, answered)
	}
	// Once it is due, it is not shown early.
	*s.clock = st.Card.DueEarly.Add(time.Second)
	if st = must(s.study.Current()); st.Card == nil || st.Card.CardID != card || st.Card.DueEarly != nil {
		t.Fatalf("due card = %+v", st.Card)
	}
}

func TestStudyLeech(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	opts := DeckOptions(core.DefaultDeckOptions())
	opts.LeechThreshold, opts.LeechAction = 1, "suspend"
	check(t, s.decks.SetOptions(deck, opts))
	basicNote(t, s, "uno", "one")
	st := must(s.study.Start(deck))
	id := st.Card.CardID
	if st = must(s.study.Answer(id, 4, 1000)); st.Leech {
		t.Fatalf("Easy on a new card = %+v", st)
	}
	*s.clock = s.clock.AddDate(0, 2, 0)
	st = must(s.study.Start(deck))
	if st.Card == nil || st.Card.CardID != id || st.Card.Kind != "review" {
		t.Fatalf("review = %+v", st.Card)
	}
	// Message (spec): the answer made the card a leech and suspended it.
	st = must(s.study.Answer(id, 1, 1000))
	if !st.Leech || !st.LeechSuspended || st.Card != nil {
		t.Fatalf("after lapse = %+v", st)
	}
	card := must(s.browse.Card(id))
	if !card.Suspended || !slices.Equal(card.Note.Tags, []string{"leech"}) {
		t.Fatalf("leech card = %+v", card)
	}
	// Only the answer's own state reports it.
	if st = must(s.study.Current()); st.Leech || st.LeechSuspended {
		t.Fatalf("current = %+v", st)
	}
	// Undo leech (spec): unsuspended, no tag, shown again.
	st = must(s.study.Undo())
	if st.Card == nil || st.Card.CardID != id || st.Leech {
		t.Fatalf("after undo = %+v", st)
	}
	card = must(s.browse.Card(id))
	if card.Suspended || len(card.Note.Tags) != 0 {
		t.Fatalf("after undo = %+v", card)
	}
	// Tag only: reported, not suspended.
	opts.LeechAction = "tag"
	check(t, s.decks.SetOptions(deck, opts))
	st = must(s.study.Answer(id, 1, 1000))
	if !st.Leech || st.LeechSuspended {
		t.Fatalf("tag only = %+v", st)
	}
	if card = must(s.browse.Card(id)); card.Suspended || !slices.Equal(card.Note.Tags, []string{"leech"}) {
		t.Fatalf("tag only = %+v", card)
	}
}

func TestBrowseService(t *testing.T) {
	s, _ := opened(t)
	basicNote(t, s, "comer", "to eat", "verb")
	basicNote(t, s, "perro", "dog", "noun")
	ids := must(s.browse.Search("tag:verb", "created", false))
	if len(ids) != 1 {
		t.Fatalf("search = %v", ids)
	}
	if _, err := s.browse.Search("is:bogus", "created", false); err == nil || !strings.Contains(err.Error(), "is:bogus") {
		t.Fatalf("invalid search: %v", err)
	}
	all := must(s.browse.Search("", "question", false))
	rows := must(s.browse.Rows(all, nil))
	if len(rows) != 2 || rows[0].Question != "comer" || rows[0].State != "new" || rows[0].Due != nil {
		t.Fatalf("rows = %+v", rows)
	}
	d := must(s.browse.Card(ids[0]))
	if d.State != "new" || d.Stability != nil || d.Note.Fields[0] != "comer" || d.NoteType.Name != "Basic" || d.CardName != "Card 1" {
		t.Fatalf("details = %+v", d)
	}
	archive := must(s.decks.Create("Archive"))
	if err := s.browse.Move(all, archive); err != nil {
		t.Fatal(err)
	}
	if rows := must(s.browse.Rows(all, nil)); rows[0].Deck != "Archive" || rows[1].Deck != "Archive" {
		t.Fatalf("rows after move = %+v", rows)
	}
	if err := s.browse.AddTags(all, []string{"x"}); err != nil {
		t.Fatal(err)
	}
	if n := len(must(s.browse.Search("tag:x", "created", false))); n != 2 {
		t.Fatalf("tagged = %d", n)
	}
	check(t, s.browse.RemoveTags(all, []string{"x"}))
	check(t, s.browse.SetSuspended(all, true))
	if n := len(must(s.browse.Search("is:suspended", "created", false))); n != 2 {
		t.Fatalf("suspended = %d", n)
	}
	// Flag selection (spec), then clear flags.
	check(t, s.browse.SetFlag(all, 4))
	if rows := must(s.browse.Rows(all, nil)); rows[0].Flag != 4 || rows[1].Flag != 4 {
		t.Fatalf("rows after flag = %+v", rows)
	}
	if d := must(s.browse.Card(all[0])); d.Flag != 4 {
		t.Fatalf("details flag = %d", d.Flag)
	}
	if n := len(must(s.browse.Search("flag:4", "created", false))); n != 2 {
		t.Fatalf("flagged = %d", n)
	}
	check(t, s.browse.SetFlag(all, 0))
	if rows := must(s.browse.Rows(all, nil)); rows[0].Flag != 0 || rows[1].Flag != 0 {
		t.Fatalf("rows after clear = %+v", rows)
	}
	// Mark the selection's notes, then toggle the mark off.
	check(t, s.browse.ToggleMark(all))
	if rows := must(s.browse.Rows(all, nil)); !rows[0].Marked || !rows[1].Marked {
		t.Fatalf("rows after mark = %+v", rows)
	}
	if n := len(must(s.browse.Search("tag:marked", "created", false))); n != 2 {
		t.Fatalf("marked = %d", n)
	}
	check(t, s.browse.ToggleMark(all))
	if rows := must(s.browse.Rows(all, nil)); rows[0].Marked || rows[1].Marked {
		t.Fatalf("rows after unmark = %+v", rows)
	}
	check(t, s.browse.ResetToNew(all, ResetOptions{RestorePosition: true}))
	if n := must(s.browse.DeleteImpact(all)); n != (DeleteImpact{Notes: 2, Cards: 2}) {
		t.Fatalf("delete impact = %+v", n)
	}
	if d := must(s.browse.Card(ids[0])); d.NoteCards != 1 {
		t.Fatalf("basic note cards = %d", d.NoteCards)
	}
	check(t, s.browse.DeleteNotes(all[:1]))
	if n := len(must(s.browse.Search("", "created", false))); n != 1 {
		t.Fatalf("cards left = %d", n)
	}
	// A reversed note: one picked card reaches its sibling too.
	reversed := must(s.notes.Types())[1]
	note := must(s.notes.Add(NewNote{TypeID: reversed.ID, DeckID: archive, Fields: []string{"gato", "cat"}}))
	cards := must(s.notes.Get(note)).CardIDs
	if d := must(s.browse.Card(cards[0])); d.NoteCards != 2 {
		t.Fatalf("reversed note cards = %d", d.NoteCards)
	}
	if n := must(s.browse.DeleteImpact(cards[:1])); n != (DeleteImpact{Notes: 1, Cards: 2}) {
		t.Fatalf("reversed delete impact = %+v", n)
	}
}

func TestBrowseColumns(t *testing.T) {
	s, _ := opened(t)
	spanish, animals := must(s.decks.Create("Spanish")), must(s.decks.Create("Animals"))
	comer := must(s.notes.Add(NewNote{TypeID: must(s.notes.Defaults(0)).TypeID, DeckID: spanish,
		Fields: []string{"comer", "to eat"}, Tags: []string{"verb"}}))
	perro := must(s.notes.Add(NewNote{TypeID: must(s.notes.Defaults(0)).TypeID, DeckID: animals,
		Fields: []string{"perro", "<b>dog</b>"}, Tags: []string{"noun"}}))

	// Rows carry only the requested columns.
	all := must(s.browse.Search("", "created", false))
	rows := must(s.browse.Rows(all, []string{"answer", "tags", "interval", "created", "due"}))
	if len(rows) != 2 || rows[1].Answer != "dog" || rows[1].Tags != "noun" || rows[1].Question != "" ||
		rows[1].Deck != "" || rows[1].Interval != nil || rows[1].Created == nil || rows[1].Position == nil ||
		rows[1].Due != nil || rows[1].CardModified != nil || rows[1].Difficulty != nil {
		t.Fatalf("rows = %+v", rows)
	}

	// Sort by lapses (spec): the card with the most lapses first.
	study(t, s, spanish, 3, 3, 3, 1, 3)
	study(t, s, animals, 3, 3, 3, 1, 3, 1, 3)
	ids := must(s.browse.Search("", "lapses", true))
	rows = must(s.browse.Rows(ids, []string{"lapses", "interval", "stability"}))
	if rows[0].NoteID != perro || rows[0].Lapses != 2 || rows[1].NoteID != comer || rows[1].Lapses != 1 ||
		rows[0].Interval == nil || *rows[0].Interval < 1 || rows[0].Stability == nil {
		t.Fatalf("rows by lapses = %+v", rows)
	}

	for _, err := range []error{
		func() error { _, err := s.browse.Search("", "bogus", false); return err }(),
		func() error { _, err := s.browse.Rows(ids, []string{"deck", "bogus"}); return err }(),
	} {
		if err == nil || !strings.Contains(err.Error(), "bogus") {
			t.Fatalf("unknown column: %v", err)
		}
	}
}

// study answers the first card of deck with the ratings, each when it is
// due.
func study(t *testing.T, s *services, deck int64, ratings ...int) {
	t.Helper()
	card := must(s.study.Start(deck)).Card.CardID
	for _, r := range ratings {
		must(s.study.Answer(card, r, 1000))
		*s.clock = must(s.browse.CardInfo(card)).Due.Add(time.Minute)
	}
}

func TestBrowserColumnsSetting(t *testing.T) {
	s, store := opened(t)
	if cols := must(s.app.Settings()).BrowserColumns; !slices.Equal(cols, []string{"question", "deck", "due", "noteType"}) {
		t.Fatalf("default columns = %v", cols)
	}
	// Layout remembered (spec): hide Note type, add Interval, restart.
	check(t, s.app.SetBrowserColumns([]string{"question", "deck", "due", "interval"}))
	// Changing columns is not an undo step.
	if u := must(s.app.UndoInfo()); u.Undo != "" {
		t.Fatalf("undo after setting columns = %+v", u)
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if cols := must(again.app.Settings()).BrowserColumns; !slices.Equal(cols, []string{"question", "deck", "due", "interval"}) {
		t.Fatalf("columns after restart = %v", cols)
	}
	for _, bad := range [][]string{nil, {"question", "bogus"}, {"deck", "question", "deck"}} {
		if err := again.app.SetBrowserColumns(bad); err == nil {
			t.Fatalf("columns %v accepted", bad)
		}
	}
}

func TestBrowserNotesMode(t *testing.T) {
	s, store := opened(t)
	if mode := must(s.app.Settings()).BrowserMode; mode != "cards" {
		t.Fatalf("default mode = %q", mode)
	}
	reversed := must(s.notes.Types())[1]
	d := must(s.notes.Defaults(0))
	var notes []int64
	for _, front := range []string{"gato", "perro", "pájaro"} {
		notes = append(notes, must(s.notes.Add(NewNote{TypeID: reversed.ID, DeckID: d.DeckID, Fields: []string{front, "x"}})))
	}
	basicNote(t, s, "uno", "one")

	// Reversed note once (spec): one row per note, showing 2 cards.
	ids := must(s.browse.SearchNotes("gato", "created", false))
	if !slices.Equal(ids, notes[:1]) {
		t.Fatalf("notes = %v", ids)
	}
	rows := must(s.browse.NoteRows(ids))
	if len(rows) != 1 || rows[0].Cards != 2 || rows[0].SortField != "gato" || rows[0].Deck != "Default" ||
		rows[0].State != "new" || rows[0].Position == nil || rows[0].Due != nil {
		t.Fatalf("rows = %+v", rows)
	}
	if _, err := s.browse.SearchNotes("", "question", false); err == nil {
		t.Fatal("notes sorted by question")
	}

	// Suspend notes (spec): all 6 cards of 3 reversed notes.
	cards := must(s.browse.NoteCards(notes))
	if len(cards) != 6 {
		t.Fatalf("cards of notes = %v", cards)
	}
	check(t, s.browse.SetSuspended(cards, true))
	if got := must(s.browse.Search("is:suspended", "created", false)); len(got) != 6 {
		t.Fatalf("suspended cards = %v", got)
	}

	// Match through one card (spec).
	check(t, s.browse.SetSuspended(cards[1:], false))
	if got := must(s.browse.SearchNotes("is:suspended", "created", false)); !slices.Equal(got, notes[:1]) {
		t.Fatalf("is:suspended notes = %v", got)
	}

	// The mode is remembered across restarts.
	check(t, s.app.SetBrowserMode("notes"))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if mode := must(again.app.Settings()).BrowserMode; mode != "notes" {
		t.Fatalf("mode after restart = %q", mode)
	}
	if err := again.app.SetBrowserMode("rows"); err == nil {
		t.Fatal("unknown mode accepted")
	}
}

func TestBrowserSelectNotes(t *testing.T) {
	s, _ := opened(t)
	reversed := must(s.notes.Types())[1]
	d := must(s.notes.Defaults(0))
	note := must(s.notes.Add(NewNote{TypeID: reversed.ID, DeckID: d.DeckID, Fields: []string{"gato", "cat"}}))
	basicNote(t, s, "uno", "one")
	both := must(s.browse.NoteCards([]int64{note}))
	if len(both) != 2 {
		t.Fatalf("cards of reversed note = %v", both)
	}

	// Select notes (spec): the forward card gives both cards of its note.
	got := must(s.browse.CardsOfNotes(both[:1]))
	if !slices.Equal(got, both) {
		t.Fatalf("cards of notes = %v, want %v", got, both)
	}
	// Both cards of one note count the note once; unknown cards are skipped.
	if got := must(s.browse.CardsOfNotes([]int64{both[1], both[0], 999})); !slices.Equal(got, both) {
		t.Fatalf("cards of notes from both cards = %v", got)
	}
}

func TestBrowserSplitSetting(t *testing.T) {
	s, store := opened(t)
	if split := must(s.app.Settings()).BrowserSplit; split != 0.6 {
		t.Fatalf("default split = %v", split)
	}
	// Remembered width (spec): the list is made wider, and the browser
	// reopened after a restart.
	check(t, s.app.SetBrowserSplit(0.7))
	if u := must(s.app.UndoInfo()); u.Undo != "" {
		t.Fatalf("undo after setting split = %+v", u)
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if split := must(again.app.Settings()).BrowserSplit; split != 0.7 {
		t.Fatalf("split after restart = %v", split)
	}
	for _, bad := range []float64{0, -0.2, 1, 1.5} {
		if err := again.app.SetBrowserSplit(bad); err == nil {
			t.Fatalf("split %v accepted", bad)
		}
	}
}

func TestSavedSearches(t *testing.T) {
	s, store := opened(t)
	if saved := must(s.browse.SavedSearches()); len(saved) != 0 {
		t.Fatalf("saved searches at start = %+v", saved)
	}
	// Save and run (spec).
	check(t, s.browse.SaveSearch(" Spanish leeches ", "deck:Spanish tag:leech"))
	check(t, s.browse.SaveSearch("added today", "added:1"))
	want := []core.SavedSearch{{Name: "added today", Query: "added:1"}, {Name: "Spanish leeches", Query: "deck:Spanish tag:leech"}}
	if saved := must(s.browse.SavedSearches()); !slices.Equal(saved, want) {
		t.Fatalf("saved searches = %+v", saved)
	}
	// Saving is not an undo step.
	if u := must(s.app.UndoInfo()); u.Undo != "" {
		t.Fatalf("undo after saving a search = %+v", u)
	}

	// Invalid search (spec): refused with the search error.
	_, searchErr := s.browse.Search("is:bogus", "created", false)
	if err := s.browse.SaveSearch("bogus", "is:bogus"); err == nil || searchErr == nil || err.Error() != searchErr.Error() {
		t.Fatalf("saving is:bogus = %v, searching = %v", err, searchErr)
	}
	// Names are unique, ignoring case, and not empty.
	for _, name := range []string{"spanish LEECHES", "  "} {
		if err := s.browse.SaveSearch(name, "is:new"); err == nil {
			t.Fatalf("saved under %q", name)
		}
	}

	// Rename and delete.
	check(t, s.browse.RenameSavedSearch("added today", "today"))
	if err := s.browse.RenameSavedSearch("today", "Spanish Leeches"); err == nil {
		t.Fatal("renamed to a taken name")
	}
	if err := s.browse.RenameSavedSearch("missing", "x"); err == nil {
		t.Fatal("renamed a missing search")
	}
	check(t, s.browse.DeleteSavedSearch("Spanish leeches"))
	if err := s.browse.DeleteSavedSearch("Spanish leeches"); err == nil {
		t.Fatal("deleted a missing search")
	}

	// Kept across restarts, and in backups.
	backup := must(s.app.CreateBackup())
	check(t, s.browse.SaveSearch("later", "is:due"))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	want = []core.SavedSearch{{Name: "later", Query: "is:due"}, {Name: "today", Query: "added:1"}}
	if saved := must(again.browse.SavedSearches()); !slices.Equal(saved, want) {
		t.Fatalf("saved searches after restart = %+v", saved)
	}
	check(t, again.app.RestoreBackup(backup.Name))
	if saved := must(again.browse.SavedSearches()); !slices.Equal(saved, want[1:]) {
		t.Fatalf("saved searches after restoring the backup = %+v", saved)
	}
}

func TestDefaultSearch(t *testing.T) {
	s, store := opened(t)
	if q := must(s.app.Settings()).DefaultSearch; q != "" {
		t.Fatalf("default search at start = %q", q)
	}
	check(t, s.app.SetDefaultSearch(" is:due "))
	if err := s.app.SetDefaultSearch("is:bogus"); err == nil {
		t.Fatal("invalid default search accepted")
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if q := must(again.app.Settings()).DefaultSearch; q != "is:due" {
		t.Fatalf("default search after restart = %q", q)
	}
	check(t, again.app.SetDefaultSearch(""))
	if q := must(again.app.Settings()).DefaultSearch; q != "" {
		t.Fatalf("cleared default search = %q", q)
	}
}

func TestCardInfoNewCard(t *testing.T) {
	// New card (spec): its position, no stability, difficulty or
	// retrievability.
	s, _ := opened(t)
	basicNote(t, s, "uno", "one")
	basicNote(t, s, "dos", "two")
	ids := must(s.browse.Search("", "created", false))
	info := must(s.browse.CardInfo(ids[1]))
	if info.State != "new" || info.Position == nil || *info.Position != 2 || info.Due != nil || info.Interval != nil {
		t.Fatalf("scheduling = %+v", info)
	}
	if info.Stability != nil || info.Difficulty != nil || info.Retrievability != nil {
		t.Fatalf("memory state = %+v", info)
	}
	if info.Reviews != 0 || len(info.History) != 0 || info.FirstReview != nil || info.AverageTime != nil {
		t.Fatalf("history = %+v", info)
	}
	if info.CardType != "Card 1" || info.NoteType != "Basic" || info.Deck != "Default" || info.Added.IsZero() {
		t.Fatalf("identity = %+v", info)
	}
}

func TestCardInfoHistoryAfterStudying(t *testing.T) {
	// History after studying (spec): Good, then Good 10 minutes later.
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	card := must(s.study.Start(deck)).Card.CardID
	must(s.study.Answer(card, 3, 4000))
	*s.clock = s.clock.Add(10 * time.Minute)
	must(s.study.Answer(card, 3, 2000))
	info := must(s.browse.CardInfo(card))
	h := info.History
	if len(h) != 2 || h[0].Kind != "learning" || h[1].Kind != "learning" || h[0].Rating != 3 || h[1].Rating != 3 {
		t.Fatalf("history = %+v", h)
	}
	if !h[0].Time.After(h[1].Time) {
		t.Fatalf("not newest first: %+v", h)
	}
	if h[0].Interval == nil || *h[0].Interval < 24*time.Hour {
		t.Fatalf("second interval = %v", h[0].Interval)
	}
	// The first answer's step: 10 minutes plus step fuzz.
	if h[1].Interval == nil || *h[1].Interval < 10*time.Minute || *h[1].Interval > 15*time.Minute {
		t.Fatalf("first interval = %v", h[1].Interval)
	}
	if h[0].AnswerTime != 2*time.Second || *info.TotalTime != 6*time.Second || *info.AverageTime != 3*time.Second {
		t.Fatalf("times = %+v", info)
	}
	if !info.FirstReview.Equal(h[1].Time) || !info.LatestReview.Equal(h[0].Time) {
		t.Fatalf("first %v latest %v", info.FirstReview, info.LatestReview)
	}
}

func TestCardInfoReviewCard(t *testing.T) {
	// Review card (spec): reviewed 5 times with 1 lapse.
	s, store := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	card := must(s.study.Start(deck)).Card.CardID
	for _, r := range []int{3, 3, 3, 1, 3} {
		must(s.study.Answer(card, r, 1000))
		d := must(s.browse.CardInfo(card))
		*s.clock = d.Due.Add(time.Minute)
	}
	info := must(s.browse.CardInfo(card))
	if info.State != "review" || info.Reviews != 5 || info.Lapses != 1 || info.Due == nil || info.Position != nil {
		t.Fatalf("info = %+v", info)
	}
	if info.Stability == nil || info.Difficulty == nil || info.Retrievability == nil || *info.Retrievability >= 1 {
		t.Fatalf("memory state = %+v", info)
	}
	if info.Interval == nil || *info.Interval < 24*time.Hour {
		t.Fatalf("interval = %v", info.Interval)
	}
	var kinds []string
	for _, h := range info.History {
		kinds = append(kinds, h.Kind)
	}
	if want := []string{"relearning", "review", "review", "learning", "learning"}; !slices.Equal(kinds, want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	// Months later, after a restart, the old months are not in memory; card
	// info still has the whole history.
	*s.clock = s.clock.AddDate(0, 4, 0)
	s2 := newServices(t, store)
	*s2.clock = *s.clock
	must(s2.app.Open("/data"))
	if h := must(s2.browse.CardInfo(card)).History; len(h) != 5 {
		t.Fatalf("history after restart = %d entries", len(h))
	}
}

func TestCardInfoManualEntry(t *testing.T) {
	// Imported manual entries (rating 0) have no state before; their
	// interval comes from the stored days.
	e := core.ReviewEntry{Rating: 0, StateBefore: scheduler.Review, IvlAfter: 7}
	if k := historyKind(e); k != "manual" {
		t.Fatalf("kind = %q", k)
	}
	if d := intervalAfter(scheduler.CardState{}, time.Now(), e.IvlAfter); d == nil || *d != 7*24*time.Hour {
		t.Fatalf("interval = %v", d)
	}
	if d := intervalAfter(scheduler.CardState{}, time.Now(), 0); d != nil {
		t.Fatalf("unknown interval = %v", d)
	}
}

func collectImport(t *testing.T, events <-chan ImportEvent) (progress int, last ImportEvent) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case e, ok := <-events:
			if !ok {
				return progress, last
			}
			if e.Summary == nil && e.Error == "" {
				progress++
			}
			last = e
		case <-timeout:
			t.Fatal("import stream did not finish")
		}
	}
}

func TestTransferCSVFormat(t *testing.T) {
	s, _ := opened(t)
	text := "#html:false\n#tags:chapter1 vocab\na < b, x|1 & 2\n# chapter 2\nc, y|d\n"
	info := must(s.transfer.StageBytes("words.txt", []byte(text)))
	p := must(s.transfer.Inspect(info.ID)).Csv
	if p == nil || p.Delimiter != "comma" || p.AllowHTML || strings.Join(p.Tags, " ") != "chapter1 vocab" {
		t.Fatalf("preview = %+v", p)
	}
	// Override detection (spec): the preview is split at "|".
	piped := must(s.transfer.PreviewCsv(info.ID, "pipe"))
	if piped.Delimiter != "pipe" || piped.Columns != 2 || piped.TotalRows != 2 || piped.Rows[0][0] != "a < b, x" {
		t.Fatalf("pipe preview = %+v", piped)
	}
	if again := must(s.transfer.PreviewCsv(info.ID, "")); again.Delimiter != "comma" {
		t.Fatalf("default preview = %+v", again)
	}
	if _, err := s.transfer.PreviewCsv(info.ID, "slash"); err == nil {
		t.Fatal("unknown separator accepted")
	}
	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{
		TypeID: piped.TypeID, DeckID: piped.DeckID, Mapping: piped.Mapping,
		Separator: "pipe", AllowHTML: piped.AllowHTML, Tags: piped.Tags}))
	if last.Summary == nil || last.Summary.NotesAdded != 2 || last.Summary.CommentLines != 1 {
		t.Fatalf("import = %+v", last)
	}
	// Literal text (spec) and header tags (spec).
	ids := must(s.browse.Search("tag:chapter1 tag:vocab", "created", false))
	if len(ids) != 2 {
		t.Fatalf("tagged = %v", ids)
	}
	n := must(s.notes.Get(must(s.browse.Rows(ids, nil))[0].NoteID))
	if n.Fields[0] != "a &lt; b, x" || n.Fields[1] != "1 &amp; 2" {
		t.Fatalf("note = %+v", n)
	}
}

func TestTransferCSV(t *testing.T) {
	s, _ := opened(t)
	info := must(s.transfer.StageBytes("words.txt", []byte("Front\tBack\nhola\thello\nperro\tdog\n")))
	if info.Kind != "csv" {
		t.Fatalf("stage = %+v", info)
	}
	in := must(s.transfer.Inspect(info.ID))
	p := in.Csv
	if p == nil || p.Delimiter != "tab" || p.Columns != 2 || !p.HeaderRow || p.Mapping[0] != 0 || p.Mapping[1] != 1 || p.TotalRows != 3 {
		t.Fatalf("preview = %+v", p)
	}
	types := must(s.notes.Types())
	if p.TypeID != types[0].ID {
		t.Fatalf("default note type %d, want Basic", p.TypeID)
	}
	progress, last := collectImport(t, s.transfer.Import(context.Background(), info.ID,
		ImportOptions{TypeID: p.TypeID, DeckID: p.DeckID, Mapping: p.Mapping, HasHeader: true, AllowHTML: p.AllowHTML}))
	if last.Summary == nil || last.Summary.NotesAdded != 2 || progress == 0 {
		t.Fatalf("import: progress %d, last %+v", progress, last)
	}
	// The stage is gone after a successful import.
	if _, err := s.transfer.Inspect(info.ID); err == nil {
		t.Fatal("stage kept")
	}
	// A failing import reports its error in the stream.
	bad := must(s.transfer.StageBytes("x.csv", []byte("a,b\n")))
	_, last = collectImport(t, s.transfer.Import(context.Background(), bad.ID, ImportOptions{TypeID: 1, DeckID: 1}))
	if last.Error == "" {
		t.Fatalf("last = %+v", last)
	}
}

func TestTransferCSVRoutingColumns(t *testing.T) {
	s, _ := opened(t)
	text := "#separator:tab\n#html:true\n#notetype:Missing\n#deck:Imported::Words\n" +
		"#guid column:1\n#notetype column:2\n#deck column:3\n#tags column:6\n" +
		"g1\tBasic\tSpanish\thola\thello\tverb\n" +
		"g2\tCloze\t\t{{c1::Paris}} is in France\tgeo\tgeo\n" +
		"g3\tVocab\tSpanish\tgato\tcat\t\n"
	info := must(s.transfer.StageBytes("notes.txt", []byte(text)))
	p := must(s.transfer.Inspect(info.ID)).Csv
	if p == nil || fmt.Sprint(p.Mapping) != "[-5 -3 -4 0 1 -2]" || p.HeaderRow {
		t.Fatalf("preview = %+v", p)
	}
	// The file's deck does not exist yet; its note type does not exist.
	if p.DeckName != "Imported::Words" || len(p.Warnings) != 1 || !strings.Contains(p.Warnings[0], `"Missing"`) {
		t.Fatalf("preview defaults = %+v", p)
	}
	if fmt.Sprint(p.NoteTypes) != "{[Vocab] 1 [Field 1 Field 2]}" {
		t.Fatalf("note types = %+v", p.NoteTypes)
	}
	// Another mapping, without a note type column: every row has the
	// default note type.
	o := ImportOptions{TypeID: p.TypeID, Mapping: []int{-1, -1, -1, 0, 1, -2}, AllowHTML: true}
	if nt := must(s.transfer.CsvNoteTypes(info.ID, o)); len(nt.Unknown) != 0 || strings.Join(nt.Fields, ",") != "Front,Back" {
		t.Fatalf("without a note type column = %+v", nt)
	}

	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{
		TypeID: p.TypeID, DeckID: p.DeckID, DeckName: p.DeckName, Mapping: p.Mapping, AllowHTML: true}))
	sum := last.Summary
	if sum == nil || sum.NotesAdded != 2 || sum.NotesSkipped != 1 || len(sum.Warnings) != 1 || !strings.Contains(sum.Warnings[0], `"Vocab"`) {
		t.Fatalf("import = %+v", last)
	}
	for q, want := range map[string]int{"deck:Spanish": 1, "deck:Imported::Words": 1} {
		if ids := must(s.browse.Search(q, "created", false)); len(ids) != want {
			t.Fatalf("%s: %v", q, ids)
		}
	}

	// Imported again with an edit: the GUID matches, so the note is updated
	// though duplicates are skipped, within the deck.
	info = must(s.transfer.StageBytes("notes.txt", []byte(strings.Replace(text, "hello", "hi", 1))))
	_, last = collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{
		TypeID: p.TypeID, DeckID: p.DeckID, DeckName: p.DeckName, Mapping: p.Mapping, AllowHTML: true,
		Duplicates: "skip", DuplicateScope: "deck"}))
	if last.Summary == nil || last.Summary.NotesUpdated != 1 || last.Summary.NotesAdded != 0 || last.Summary.NotesSkipped != 2 {
		t.Fatalf("second import = %+v", last)
	}
	ids := must(s.browse.Search("deck:Spanish", "created", false))
	if n := must(s.notes.Get(must(s.browse.Rows(ids, nil))[0].NoteID)); n.Fields[1] != "hi" {
		t.Fatalf("updated note = %+v", n)
	}
}

func TestTransferApkgAndMedia(t *testing.T) {
	s, _ := opened(t)
	dir := t.TempDir()
	data, _ := os.ReadFile("importer/apkg/testdata/modern.apkg")
	path := filepath.Join(dir, "deck.apkg")
	os.WriteFile(path, data, 0o644)
	info := must(s.transfer.StagePath(path))
	in := must(s.transfer.Inspect(info.ID))
	if in.Apkg == nil || in.Apkg.Notes != 8 || in.Apkg.Cards != 12 || in.Apkg.Media != 4 || in.Apkg.Unsupported != 0 ||
		strings.Join(in.Apkg.Decks, ",") != "Fixture::Geo,Fixture::Spanish" {
		t.Fatalf("apkg info = %+v", in.Apkg)
	}
	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{KeepScheduling: true}))
	if last.Summary == nil || last.Summary.NotesAdded != 8 || last.Summary.Reviews != 24 || len(last.Summary.Warnings) != 0 {
		t.Fatalf("summary = %+v", last)
	}
	if img := must(s.media.Get("cat.jpg")); len(img) == 0 {
		t.Fatal("no image")
	}
	if snd := must(s.media.Get("hola.mp3")); len(snd) == 0 {
		t.Fatal("no audio")
	}
	if _, err := s.media.Get("missing.png"); err == nil {
		t.Fatal("missing media returned")
	}
	bad := must(s.transfer.StageBytes("t.apkg", data[:len(data)/2]))
	_, last = collectImport(t, s.transfer.Import(context.Background(), bad.ID, ImportOptions{}))
	if !strings.Contains(last.Error, "damaged") {
		t.Fatalf("corrupt import = %+v", last)
	}
}

func TestBackupExportRestore(t *testing.T) {
	s, _ := opened(t)
	basicNote(t, s, "hola", "hello")
	file := must(s.transfer.ExportBackup())
	if file.Name != "cardflip-2026-10-13.cardflip" || len(file.Data) == 0 {
		t.Fatalf("backup = %q %d", file.Name, len(file.Data))
	}
	// Restore into another installation that has other content.
	other, _ := opened(t)
	basicNote(t, other, "other", "content")
	info := must(other.transfer.StageBytes(file.Name, file.Data))
	in := must(other.transfer.Inspect(info.ID))
	if in.Backup == nil || in.Backup.Format != 2 || in.Backup.AppVersion != appVersion {
		t.Fatalf("inspect = %+v", in)
	}
	if err := other.transfer.Restore(info.ID); err != nil {
		t.Fatal(err)
	}
	rows := must(other.browse.Rows(must(other.browse.Search("", "created", false)), nil))
	if len(rows) != 1 || rows[0].Question != "hola" {
		t.Fatalf("rows after restore = %+v", rows)
	}
	// A non-backup zip is refused and nothing changes.
	notBackup := must(other.transfer.StageBytes("x.cardflip", []byte("not a zip")))
	if err := other.transfer.Restore(notBackup.ID); err == nil || !strings.Contains(err.Error(), "not a Cardflip backup") {
		t.Fatalf("restore of junk: %v", err)
	}
}

func TestRestoreFromCorruptScreen(t *testing.T) {
	good, _ := opened(t)
	basicNote(t, good, "saved", "card")
	file := must(good.transfer.ExportBackup())

	store := storage.NewMemoryStore()
	b := store.Open()
	b.Put("current", []byte(`{"v":1,"gen":1,"rl":1,"media":1}`))
	b.Put("g1/snapshot", []byte(`garbage`))
	s := newServices(t, store)
	if st := must(s.app.Open("/data")); st.State != StatusCorrupt {
		t.Fatal("not corrupt")
	}
	info := must(s.transfer.StageBytes(file.Name, file.Data))
	if err := s.transfer.Restore(info.ID); err != nil {
		t.Fatal(err)
	}
	if st := s.app.Status(); st.State != StatusOK {
		t.Fatalf("status = %+v", st)
	}
	if rows := must(s.browse.Rows(must(s.browse.Search("", "created", false)), nil)); len(rows) != 1 || rows[0].Question != "saved" {
		t.Fatalf("rows = %+v", rows)
	}
}

// deckTree returns the deck list as "<full name> <new>/<learning>/<review>".
func deckTree(t *testing.T, s *services) []string {
	t.Helper()
	var out []string
	for _, d := range must(s.decks.Tree()) {
		out = append(out, fmt.Sprintf("%s %d/%d/%d", d.FullName, d.Counts.New, d.Counts.Learning, d.Counts.Review))
	}
	return out
}

// Move from Anki (spec): a collection package replaces the collection,
// after a backup of the current one; the app settings and the deletion log
// are kept.
func TestTransferColpkgReplacesCollection(t *testing.T) {
	s, store := opened(t)
	check(t, s.app.SetTheme("dark"))
	check(t, s.app.SetRolloverHour(6))
	gone := basicNote(t, s, "deleted before", "")
	check(t, s.notes.Delete([]int64{gone}))
	basicNote(t, s, "replaced", "")

	data, _ := os.ReadFile("importer/apkg/testdata/modern.colpkg")
	path := filepath.Join(t.TempDir(), "collection-2026-10-01.colpkg")
	os.WriteFile(path, data, 0o644)
	info := must(s.transfer.StagePath(path))
	if info.Kind != "colpkg" {
		t.Fatalf("stage = %+v", info)
	}
	in := must(s.transfer.Inspect(info.ID))
	if in.Apkg == nil || in.Apkg.Notes != 8 || in.Apkg.Cards != 12 || in.Apkg.Media != 4 ||
		strings.Join(in.Apkg.Decks, ", ") != "Default, Fixture, Fixture::Geo, Fixture::Spanish, Inbox, Inbox::Later" {
		t.Fatalf("colpkg info = %+v", in.Apkg)
	}

	progress, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{}))
	if last.Summary == nil || last.Summary.NotesAdded != 8 || last.Summary.Reviews != 24 || last.Summary.Media != 4 || progress == 0 {
		t.Fatalf("import: progress %d, last %+v", progress, last)
	}
	// The deck list shows the Anki collection's decks, empty ones included,
	// with their due counts; the filtered deck's card is back home.
	want := []string{"Default 0/0/0", "Fixture 9/1/0", "Fixture::Geo 4/1/0", "Fixture::Spanish 5/0/0",
		"Inbox 0/0/0", "Inbox::Later 0/0/0"}
	if got := deckTree(t, s); !reflect.DeepEqual(got, want) {
		t.Fatalf("decks = %v, want %v", got, want)
	}
	if rows := must(s.browse.Rows(must(s.browse.Search("replaced", "created", false)), nil)); len(rows) != 0 {
		t.Fatal("old note kept")
	}
	if st := must(s.app.Settings()); st.Theme != "dark" || st.RolloverHour != 6 {
		t.Fatalf("settings = %+v", st)
	}
	if log := must(s.app.ExportDeletionLog()); log.Notes != 1 {
		t.Fatalf("deletion log has %d notes", log.Notes)
	}
	if len(must(s.media.Get("dog.png"))) == 0 {
		t.Fatal("no media")
	}
	list := autoBackupsIn(t, store)
	if len(list) != 1 {
		t.Fatalf("backups = %d, want 1", len(list))
	}
	if a := must(backup.Read(must(store.Open().Backups().Read(list[0].Name)))); !strings.Contains(string(a.Collection), "replaced") {
		t.Fatal("the backup before the import misses the replaced content")
	}
	// Defaults for the editor point into the new collection.
	d := must(s.notes.Defaults(0))
	if _, err := s.decks.Settings(d.DeckID); err != nil {
		t.Fatalf("default deck: %v", err)
	}
}

// Cancel (spec): picking and inspecting a collection package changes
// nothing; only Import replaces the collection.
func TestTransferColpkgCancel(t *testing.T) {
	s, store := opened(t)
	basicNote(t, s, "kept", "")
	data, _ := os.ReadFile("importer/apkg/testdata/modern.colpkg")
	info := must(s.transfer.StageBytes("collection.colpkg", data))
	must(s.transfer.Inspect(info.ID))
	s.transfer.Discard(info.ID)
	if got := deckTree(t, s); !reflect.DeepEqual(got, []string{"Default 1/0/0"}) {
		t.Fatalf("decks = %v", got)
	}
	if len(autoBackupsIn(t, store)) != 0 {
		t.Fatal("backup made")
	}
}

// Corrupt package (spec): an error explains the file could not be read,
// and the collection is unchanged.
func TestTransferColpkgCorrupt(t *testing.T) {
	s, store := opened(t)
	basicNote(t, s, "kept", "")
	data, _ := os.ReadFile("importer/apkg/testdata/truncated.colpkg")
	info := must(s.transfer.StageBytes("collection.colpkg", data))
	if _, err := s.transfer.Inspect(info.ID); err == nil || !strings.Contains(err.Error(), "damaged") {
		t.Fatalf("inspect = %v", err)
	}
	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{}))
	if !strings.Contains(last.Error, "damaged") {
		t.Fatalf("import = %+v", last)
	}
	if rows := must(s.browse.Rows(must(s.browse.Search("kept", "created", false)), nil)); len(rows) != 1 {
		t.Fatal("collection changed")
	}
	if len(autoBackupsIn(t, store)) != 0 {
		t.Fatal("backup made for a package that cannot be read")
	}
}

// A collection package also replaces a collection that could not be read.
func TestTransferColpkgFromCorruptScreen(t *testing.T) {
	store := storage.NewMemoryStore()
	b := store.Open()
	b.Put("current", []byte(`{"v":1,"gen":1,"rl":1,"media":1}`))
	b.Put("g1/snapshot", []byte(`garbage`))
	s := newServices(t, store)
	if st := must(s.app.Open("/data")); st.State != StatusCorrupt {
		t.Fatal("not corrupt")
	}
	data, _ := os.ReadFile("importer/apkg/testdata/legacy.colpkg")
	info := must(s.transfer.StageBytes("collection.colpkg", data))
	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{}))
	if last.Summary == nil || last.Summary.NotesAdded != 8 {
		t.Fatalf("import = %+v", last)
	}
	if st := s.app.Status(); st.State != StatusOK {
		t.Fatalf("status = %+v", st)
	}
}

func TestKindOfFile(t *testing.T) {
	for name, want := range map[string]string{"a.apkg": "apkg", "b.COLPKG": "colpkg", "c.cardflip": "backup", "d.tsv": "csv"} {
		if got, err := kindOfFile(name, nil); err != nil || got != want {
			t.Errorf("%s: %q %v, want %q", name, got, err, want)
		}
	}
}

func TestCheckStorageNative(t *testing.T) {
	a := newApp()
	a.persist = func() bool { return true }
	a.backups.tick = 0
	s := &AppService{a: a}
	dir := t.TempDir()
	if st := must(s.Open(dir)); st.State != StatusFresh {
		t.Fatalf("status = %+v", st)
	}
	if failures := must(s.CheckStorage()); len(failures) != 0 {
		t.Fatalf("contract failures: %v", failures)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "selftest") {
			t.Fatalf("scratch store left behind: %s", e.Name())
		}
	}
}

var _ = core.Change{}

// autoBackupsIn lists the automatic backups in store, newest first.
func autoBackupsIn(t *testing.T, store *storage.MemoryStore) []backup.Auto {
	t.Helper()
	return must(backup.ListAuto(store.Open().Backups()))
}

// tick advances the clock by d and runs the periodic backup check.
func (s *services) tick(t *testing.T, d time.Duration) {
	t.Helper()
	*s.clock = s.clock.Add(d)
	if _, err := s.app.a.periodicBackup(); err != nil {
		t.Fatal(err)
	}
}

func setBackupSettings(t *testing.T, s *services, edit func(*core.BackupSettings)) {
	t.Helper()
	c := must(s.app.a.collection())
	st := c.Settings()
	edit(&st.Backup)
	check(t, c.UpdateSettings(st))
}

// Automatic backups (spec): periodic backup.
func TestPeriodicBackupWhileStudying(t *testing.T) {
	s, store := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	for i := range 20 {
		basicNote(t, s, fmt.Sprint("question ", i), "answer")
	}
	start := *s.clock
	st := must(s.study.Start(deck))
	for range 40 {
		s.tick(t, time.Minute)
		if st.Card != nil {
			st = must(s.study.Answer(st.Card.CardID, 3, 5000))
		}
	}
	list := autoBackupsIn(t, store)
	if len(list) == 0 {
		t.Fatal("no automatic backup after 40 minutes of studying")
	}
	for i, b := range list {
		if b.CreatedAt.Before(start) || b.CreatedAt.After(s.clock.Add(time.Second)) {
			t.Errorf("backup %s outside the session", b.Name)
		}
		if i > 0 && list[i-1].CreatedAt.Sub(b.CreatedAt) < 30*time.Minute {
			t.Errorf("backups %s and %s closer than the interval", b.Name, list[i-1].Name)
		}
	}
	// The backup holds the collection without media and can be read back.
	a := must(backup.Read(must(store.Open().Backups().Read(list[0].Name))))
	if !strings.Contains(string(a.Collection), "question 19") || len(a.Revlog) == 0 {
		t.Fatal("backup misses notes or reviews")
	}
}

// Automatic backups (spec): nothing changed.
func TestNoPeriodicBackupWithoutChanges(t *testing.T) {
	s, store := opened(t)
	for range 60 {
		s.tick(t, time.Minute)
	}
	if list := autoBackupsIn(t, store); len(list) != 0 {
		t.Fatalf("fresh collection, no change: %d backups", len(list))
	}
	basicNote(t, s, "one change", "")
	s.tick(t, time.Minute)
	if list := autoBackupsIn(t, store); len(list) != 1 {
		t.Fatalf("after a change: %d backups, want 1", len(list))
	}
	for range 60 {
		s.tick(t, time.Minute)
	}
	if list := autoBackupsIn(t, store); len(list) != 1 {
		t.Fatalf("an hour without change: %d backups, want 1", len(list))
	}
	// After a restart, the stored collection is the baseline and the
	// interval counts from the newest stored backup.
	s2 := newServices(t, store)
	*s2.clock = *s.clock
	must(s2.app.Open("/data"))
	for range 60 {
		s2.tick(t, time.Minute)
	}
	if list := autoBackupsIn(t, store); len(list) != 1 {
		t.Fatalf("restart without change: %d backups, want 1", len(list))
	}
	basicNote(t, s2, "after restart", "")
	s2.tick(t, time.Minute)
	if list := autoBackupsIn(t, store); len(list) != 2 {
		t.Fatalf("change after restart: %d backups, want 2", len(list))
	}
}

// Backup settings (spec): periodic backups off; a backup is still made
// before a risky operation (restore).
func TestBackupBeforeRestoreWithPeriodicBackupsOff(t *testing.T) {
	s, store := opened(t)
	basicNote(t, s, "from the file", "")
	file := must(s.transfer.ExportBackup())
	setBackupSettings(t, s, func(b *core.BackupSettings) { b.IntervalMinutes = 0 })
	basicNote(t, s, "made today", "")
	for range 120 {
		s.tick(t, time.Minute)
	}
	if list := autoBackupsIn(t, store); len(list) != 0 {
		t.Fatalf("interval 0: %d periodic backups", len(list))
	}

	info := must(s.transfer.StageBytes(file.Name, file.Data))
	check(t, s.transfer.Restore(info.ID))
	list := autoBackupsIn(t, store)
	if len(list) != 1 {
		t.Fatalf("backups after restore = %d, want 1", len(list))
	}
	a := must(backup.Read(must(store.Open().Backups().Read(list[0].Name))))
	if !strings.Contains(string(a.Collection), "made today") {
		t.Fatal("the backup before the restore misses the replaced content")
	}
}

// A risky operation does not start when the backup before it fails.
func TestRestoreStopsWhenBackupFails(t *testing.T) {
	s, store := opened(t)
	basicNote(t, s, "from the file", "")
	file := must(s.transfer.ExportBackup())
	basicNote(t, s, "keep me", "")
	store.FailPut = func(key string) bool { return strings.HasPrefix(key, "backups/") }
	info := must(s.transfer.StageBytes(file.Name, file.Data))
	err := s.transfer.Restore(info.ID)
	if err == nil || !strings.Contains(err.Error(), "could not be backed up") {
		t.Fatalf("restore = %v", err)
	}
	if rows := must(s.browse.Rows(must(s.browse.Search("keep", "created", false)), nil)); len(rows) != 1 {
		t.Fatal("collection changed although the backup failed")
	}
}

// Restore from automatic backup (spec): undo a bad day. Media stay as they
// are, also those added after the backup.
func TestRestoreAutomaticBackup(t *testing.T) {
	s, store := opened(t)
	*s.clock = time.Date(2026, 10, 13, 8, 0, 0, 0, time.Local)
	basicNote(t, s, "learned this morning", "")
	morning := must(s.app.CreateBackup())
	if morning.Size == 0 || !morning.CreatedAt.Equal(*s.clock) {
		t.Fatalf("created = %+v", morning)
	}

	*s.clock = s.clock.Add(6 * time.Hour)
	basicNote(t, s, "bad change", "")
	data, _ := os.ReadFile("importer/apkg/testdata/modern.apkg")
	info := must(s.transfer.StageBytes("deck.apkg", data))
	collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{}))

	check(t, s.app.RestoreBackup(morning.Name))
	rows := must(s.browse.Rows(must(s.browse.Search("", "created", false)), nil))
	if len(rows) != 1 || rows[0].Question != "learned this morning" {
		t.Fatalf("rows after restore = %+v", rows)
	}
	if img, err := s.media.Get("cat.jpg"); err != nil || len(img) == 0 {
		t.Fatalf("media changed by the restore: %v", err)
	}
	// The backup made before the restore is at the top of the list.
	list := must(s.app.Backups())
	if len(list.Backups) != 2 || list.Backups[1].Name != morning.Name || !list.Backups[0].CreatedAt.Equal(*s.clock) {
		t.Fatalf("list = %+v", list)
	}
	if list.Bytes != list.Backups[0].Size+list.Backups[1].Size {
		t.Fatalf("bytes = %d", list.Bytes)
	}
	// It undoes the restore, and the backup made before that does not
	// replace it, although the clock stood still.
	check(t, s.app.RestoreBackup(list.Backups[0].Name))
	if rows := must(s.browse.Rows(must(s.browse.Search("bad", "created", false)), nil)); len(rows) != 1 {
		t.Fatal("restoring the backup made before the restore did not bring the day back")
	}
	if again := must(s.app.Backups()); len(again.Backups) != 3 || again.Backups[1].Name != list.Backups[0].Name {
		t.Fatalf("list after second restore = %+v", again)
	}

	for _, name := range []string{"backup-2020-01-01-000000-000.cardflip", "notes.txt", "../x"} {
		if err := s.app.RestoreBackup(name); err == nil {
			t.Errorf("RestoreBackup(%q) succeeded", name)
		}
	}
	// A damaged backup is refused and nothing changes.
	check(t, store.Open().Backups().Write("backup-2026-10-13-000000-000.cardflip", []byte("junk")))
	if err := s.app.RestoreBackup("backup-2026-10-13-000000-000.cardflip"); err == nil {
		t.Fatal("damaged backup restored")
	}
	if rows := must(s.browse.Rows(must(s.browse.Search("bad", "created", false)), nil)); len(rows) != 1 {
		t.Fatal("a refused restore changed the collection")
	}
}

func TestBackupSettingsService(t *testing.T) {
	s, _ := opened(t)
	if got := must(s.app.Settings()).Backup; got != core.DefaultBackupSettings() {
		t.Fatalf("defaults = %+v", got)
	}
	want := core.BackupSettings{IntervalMinutes: 0, Daily: 3, Weekly: 2, Monthly: 1}
	check(t, s.app.SetBackupSettings(want))
	if got := must(s.app.Settings()).Backup; got != want {
		t.Fatalf("saved = %+v", got)
	}
	if err := s.app.SetBackupSettings(core.BackupSettings{IntervalMinutes: -5}); err == nil {
		t.Fatal("negative interval accepted")
	}
	if list := must(s.app.Backups()); list.Backups == nil || len(list.Backups) != 0 || list.Bytes != 0 {
		t.Fatalf("empty list = %+v", list)
	}
}

func TestEmptyCardsService(t *testing.T) {
	s, _ := opened(t)
	// Nothing to remove (app-shell spec).
	if list := must(s.app.EmptyCards()); list.Total != 0 || list.Notes == nil {
		t.Fatalf("empty list = %+v", list)
	}
	d := must(s.notes.Defaults(0))
	cloze := must(s.notes.Types())[2]
	id := must(s.notes.Add(NewNote{TypeID: cloze.ID, DeckID: d.DeckID, Fields: []string{"{{c1::Paris}} is the capital of {{c2::France}}", ""}}))
	check(t, s.notes.Update(id, []string{"{{c1::Paris}} is the capital of France", ""}, nil))
	list := must(s.app.EmptyCards())
	if list.Total != 1 || list.Deletable != 1 || len(list.Notes) != 1 || list.Notes[0].AllEmpty ||
		list.Notes[0].Cards[0].Name != "Cloze 2" || list.Notes[0].Text != "{{c1::Paris}} is the capital of France" {
		t.Fatalf("list = %+v", list)
	}
	emptyID := list.Notes[0].Cards[0].CardID

	// Empty cloze card (study-session spec): the card is marked empty and
	// can still be rated.
	st := must(s.study.Start(d.DeckID))
	first := st.Card.CardID
	if first == emptyID || st.Card.Empty {
		t.Fatalf("cloze 1 = %+v", st.Card)
	}
	st = must(s.study.Bury(first))
	if st.Card == nil || st.Card.CardID != emptyID {
		t.Fatalf("next card = %+v", st.Card)
	}
	if st.Card == nil || !st.Card.Empty || len(st.Card.Previews) != 4 {
		t.Fatalf("empty card in study = %+v", st.Card)
	}
	st = must(s.study.Answer(emptyID, 3, 1000))

	check(t, s.app.DeleteEmptyCards([]int64{emptyID}))
	if list := must(s.app.EmptyCards()); list.Total != 0 {
		t.Fatalf("after delete = %+v", list)
	}
	if card := must(s.browse.Card(first)); card.Note.ID != id {
		t.Fatalf("remaining card = %+v", card)
	}
	if _, err := s.browse.Card(emptyID); err == nil {
		t.Fatal("empty card not deleted")
	}
}

// Undo and redo through the app service; reloading storage clears the
// history.
func TestAppUndoRedo(t *testing.T) {
	s, _ := opened(t)
	var ids []int64
	for _, q := range []string{"a", "b", "c"} {
		ids = append(ids, basicNote(t, s, q, "x"))
	}
	check(t, s.notes.Delete(ids))
	if info := must(s.app.UndoInfo()); usermsg.Plain(info.Undo) != "Delete 3 notes" || usermsg.Plain(info.Redo) != "" {
		t.Fatalf("info = %+v", info)
	}
	if name := must(s.app.Undo()); usermsg.Plain(name) != "Delete 3 notes" {
		t.Fatalf("undo = %q", name)
	}
	if rows := must(s.browse.Search("", "created", false)); len(rows) != 3 {
		t.Fatalf("after undo: %d cards", len(rows))
	}
	if info := must(s.app.UndoInfo()); usermsg.Plain(info.Undo) != "Add note" || usermsg.Plain(info.Redo) != "Delete 3 notes" {
		t.Fatalf("info = %+v", info)
	}
	if name := must(s.app.Redo()); usermsg.Plain(name) != "Delete 3 notes" {
		t.Fatalf("redo = %q", name)
	}
	if rows := must(s.browse.Search("", "created", false)); len(rows) != 0 {
		t.Fatalf("after redo: %d cards", len(rows))
	}
	must(s.app.Reload())
	if info := must(s.app.UndoInfo()); info != (UndoInfo{}) {
		t.Fatalf("after reload: %+v", info)
	}
	if _, err := s.app.Undo(); err == nil {
		t.Fatal("undo after reload")
	}
}

// The study session's undo is a view of the history: an answer undone
// through the app's undo no longer counts, and redo brings it back.
func TestStudyFollowsAppUndo(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	basicNote(t, s, "dos", "two")
	st := must(s.study.Start(deck))
	st = must(s.study.Answer(st.Card.CardID, 3, 2000))
	if st.Studied != 1 || !st.CanUndo {
		t.Fatalf("after answer = %+v", st)
	}
	if name := must(s.app.Undo()); usermsg.Plain(name) != "Answer card" {
		t.Fatalf("undo = %q", name)
	}
	if st = must(s.study.Current()); st.Studied != 0 || st.CanUndo || st.Counts.New != 2 || st.TimeSpent != 0 {
		t.Fatalf("after app undo = %+v", st)
	}
	must(s.app.Redo())
	if st = must(s.study.Current()); st.Studied != 1 || !st.CanUndo || st.Counts.New != 1 {
		t.Fatalf("after app redo = %+v", st)
	}
	// Study undo of the latest answer can be redone from the app.
	must(s.study.Undo())
	if info := must(s.app.UndoInfo()); usermsg.Plain(info.Redo) != "Answer card" {
		t.Fatalf("info = %+v", info)
	}
}

func TestStudyPreviousCard(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.notes.Defaults(0)).DeckID
	basicNote(t, s, "uno", "one")
	basicNote(t, s, "dos", "two")
	st := must(s.study.Start(deck))
	if prev := must(s.study.PreviousCardID()); prev != 0 {
		t.Fatalf("before answering = %d", prev)
	}
	first := st.Card.CardID
	st = must(s.study.Answer(first, 1, 1000))
	second := st.Card.CardID
	if prev := must(s.study.PreviousCardID()); prev != first {
		t.Fatalf("after first answer = %d, want %d", prev, first)
	}
	// Previous card (spec): its history includes the Again rating.
	if h := must(s.browse.CardInfo(first)).History; len(h) != 1 || h[0].Rating != 1 {
		t.Fatalf("history = %+v", h)
	}
	must(s.study.Answer(second, 3, 1000))
	if prev := must(s.study.PreviousCardID()); prev != second {
		t.Fatalf("after second answer = %d, want %d", prev, second)
	}
	// Undo, in study or in the app, makes the answer before it the previous
	// one.
	must(s.study.Undo())
	if prev := must(s.study.PreviousCardID()); prev != first {
		t.Fatalf("after study undo = %d, want %d", prev, first)
	}
	must(s.app.Undo())
	if prev := must(s.study.PreviousCardID()); prev != 0 {
		t.Fatalf("after app undo = %d", prev)
	}
	// A new session starts without a previous card.
	must(s.study.Answer(must(s.study.Current()).Card.CardID, 3, 1000))
	must(s.study.Start(deck))
	if prev := must(s.study.PreviousCardID()); prev != 0 {
		t.Fatalf("new session = %d", prev)
	}
}

func TestDeletionLogExportAndClear(t *testing.T) {
	s, _ := opened(t)
	empty := must(s.app.ExportDeletionLog())
	if empty.Notes != 0 || empty.Name != "cardflip-deleted-2026-10-13.txt" {
		t.Fatalf("empty log = %+v", empty)
	}
	gone := basicNote(t, s, `<b>"quoted"</b>`+"\ttab", "line\nbreak", "kitchen", "food")
	hash := basicNote(t, s, "#hash:start", "kept")
	check(t, s.notes.Delete([]int64{gone}))
	*s.clock = s.clock.Add(time.Minute)
	check(t, s.browse.DeleteNotes(must(s.notes.Get(hash)).CardIDs))

	file := must(s.app.ExportDeletionLog())
	if file.Notes != 2 {
		t.Fatalf("notes = %d", file.Notes)
	}
	want := "#separator:tab\n#html:true\n#columns:Front\tBack\tTags\tNote type\tDeleted\n" +
		"#tags column:3\n#notetype column:4\n" +
		"\"<b>\"\"quoted\"\"</b>\ttab\"\tline break\tkitchen food\tBasic\t2026-10-13 12:00:00\n" +
		"\"#hash:start\"\tkept\t\tBasic\t2026-10-13 12:01:00\n"
	if string(file.Data) != want {
		t.Fatalf("file =\n%s\nwant\n%s", file.Data, want)
	}

	// Recover a note (spec): importing the file with the note type brings
	// back the fields and tags.
	info := must(s.transfer.StageBytes(file.Name, file.Data))
	p := must(s.transfer.Inspect(info.ID)).Csv
	if p == nil || p.Columns != 5 || p.TotalRows != 2 || p.ColumnNames[2] != "Tags" {
		t.Fatalf("preview = %+v", p)
	}
	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID,
		ImportOptions{TypeID: p.TypeID, DeckID: p.DeckID, Mapping: []int{0, 1, -2, -1, -1}, Duplicates: "skip", AllowHTML: p.AllowHTML}))
	if last.Summary == nil || last.Summary.NotesAdded != 2 {
		t.Fatalf("import = %+v", last)
	}
	ids := must(s.browse.Search("tag:kitchen", "created", false))
	if len(ids) != 1 {
		t.Fatalf("recovered = %v", ids)
	}
	n := must(s.notes.Get(must(s.browse.Rows(ids, nil))[0].NoteID))
	if n.Fields[0] != `<b>"quoted"</b>`+"\ttab" || n.Fields[1] != "line break" || strings.Join(n.Tags, " ") != "kitchen food" {
		t.Fatalf("recovered note = %+v", n)
	}

	check(t, s.app.ClearDeletionLog())
	if f := must(s.app.ExportDeletionLog()); f.Notes != 0 {
		t.Fatalf("after clearing: %+v", f)
	}
	// The log is included in backups.
	check(t, s.notes.Delete([]int64{n.ID}))
	backupFile := must(s.transfer.ExportBackup())
	other, _ := opened(t)
	check(t, other.transfer.Restore(must(other.transfer.StageBytes(backupFile.Name, backupFile.Data)).ID))
	if f := must(other.app.ExportDeletionLog()); f.Notes != 1 {
		t.Fatalf("restored log = %+v", f)
	}
}

// Rows of note types with different field counts keep the tags in one
// column; the columns get generic names.
func TestDeletionLogMixedNoteTypes(t *testing.T) {
	log := []core.DeletedNote{
		{NoteType: "Basic", Fields: []string{"a", "b"}, Tags: []string{"x"}, Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local)},
		{NoteType: "Three, fields", Fields: []string{"1", "2", "3"}, Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.Local)},
	}
	got := string(deletionLogText(log, nil))
	want := "#separator:tab\n#html:true\n#columns:Field 1\tField 2\tField 3\tTags\tNote type\tDeleted\n" +
		"#tags column:4\n#notetype column:5\n" +
		"a\tb\t\tx\tBasic\t2026-01-02 03:04:05\n" +
		"1\t2\t3\t\tThree, fields\t2026-01-02 03:04:05\n"
	if got != want {
		t.Fatalf("file =\n%s\nwant\n%s", got, want)
	}
}

func TestFlagNames(t *testing.T) {
	s, store := opened(t)
	if st := must(s.app.Settings()); len(st.FlagNames) != 7 || strings.Join(st.FlagNames, "") != "" {
		t.Fatalf("default flag names = %q", st.FlagNames)
	}
	check(t, s.app.SetFlagName(1, "  Fix later "))
	if s.app.SetFlagName(8, "x") == nil || s.app.SetFlagName(0, "x") == nil {
		t.Fatal("flag outside 1–7 accepted")
	}
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); st.FlagNames[0] != "Fix later" || st.FlagNames[1] != "" {
		t.Fatalf("flag names after restart = %q", st.FlagNames)
	}
	// Clearing the name restores the colour name.
	check(t, again.app.SetFlagName(1, ""))
	if st := must(again.app.Settings()); st.FlagNames[0] != "" {
		t.Fatalf("flag names = %q", st.FlagNames)
	}
}

func TestStripPasteFormattingSetting(t *testing.T) {
	s, store := opened(t)
	if st := must(s.app.Settings()); st.StripPasteFormatting {
		t.Fatalf("default settings = %+v", st)
	}
	check(t, s.app.SetStripPasteFormatting(true))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); !st.StripPasteFormatting {
		t.Fatalf("settings after restart = %+v", st)
	}
}

// Paste images as PNG (app-shell spec): off by default, persisted.
func TestPasteImagesAsPngSetting(t *testing.T) {
	s, store := opened(t)
	if st := must(s.app.Settings()); st.PasteImagesAsPng {
		t.Fatalf("default settings = %+v", st)
	}
	check(t, s.app.SetPasteImagesAsPng(true))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); !st.PasteImagesAsPng {
		t.Fatalf("settings after restart = %+v", st)
	}
}

// Ignore accents setting (app-shell spec): off by default, persisted, and
// used by searches.
func TestIgnoreAccentsSetting(t *testing.T) {
	s, store := opened(t)
	if st := must(s.app.Settings()); st.IgnoreAccents {
		t.Fatalf("default settings = %+v", st)
	}
	d := must(s.notes.Defaults(0))
	must(s.notes.Add(NewNote{TypeID: d.TypeID, DeckID: d.DeckID, Fields: []string{"café", "x"}}))
	check(t, s.app.SetIgnoreAccents(true))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); !st.IgnoreAccents {
		t.Fatalf("settings after restart = %+v", st)
	}
	if ids := must(again.browse.Search("cafe", "created", false)); len(ids) != 1 {
		t.Fatalf("search cafe = %v", ids)
	}
}

func TestFieldOptions(t *testing.T) {
	s, store := opened(t)
	d := must(s.notes.Defaults(0))
	// The DTO has one entry per field, also without stored options, and
	// the first field sorts.
	if nt := must(s.notes.Type(d.TypeID)); len(nt.FieldOptions) != 2 || nt.FieldOptions[1] != (core.FieldOptions{}) || nt.SortField != 0 {
		t.Fatalf("options = %+v, sort field %d", nt.FieldOptions, nt.SortField)
	}
	opts := core.FieldOptions{HTMLByDefault: true, Font: "Amiri", Size: 28, RTL: true, Collapsed: true, ExcludeFromSearch: true}
	check(t, s.notes.SetFieldOptions(d.TypeID, 1, opts))
	if s.notes.SetFieldOptions(d.TypeID, 5, opts) == nil {
		t.Fatal("option of a missing field accepted")
	}
	check(t, s.notes.SetSortField(d.TypeID, 1))
	check(t, s.notes.AddField(d.TypeID, "Notes"))
	// Moving a field takes its options and the sort field along.
	check(t, s.notes.MoveField(d.TypeID, 1, 0))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	nt := must(again.notes.Type(d.TypeID))
	if !slices.Equal(nt.Fields, []string{"Back", "Front", "Notes"}) || len(nt.FieldOptions) != 3 || nt.FieldOptions[0] != opts ||
		nt.FieldOptions[1] != (core.FieldOptions{}) || nt.FieldOptions[2] != (core.FieldOptions{}) || nt.SortField != 0 {
		t.Fatalf("after restart: fields %v, options %+v, sort field %d", nt.Fields, nt.FieldOptions, nt.SortField)
	}
	check(t, again.notes.SetFieldOptions(d.TypeID, 0, core.FieldOptions{}))
	if nt := must(again.notes.Type(d.TypeID)); nt.FieldOptions[0] != (core.FieldOptions{}) {
		t.Fatalf("options = %+v", nt.FieldOptions)
	}
}

func TestFieldSticky(t *testing.T) {
	s, _ := opened(t)
	d := must(s.notes.Defaults(0))
	if nt := must(s.notes.Type(d.TypeID)); nt.FieldOptions[0].Sticky || nt.FieldOptions[1].Sticky {
		t.Fatalf("options = %+v", nt.FieldOptions)
	}
	check(t, s.notes.SetFieldSticky(d.TypeID, 1, true))
	check(t, s.notes.SetFieldOptions(d.TypeID, 1, core.FieldOptions{Sticky: true, HTMLByDefault: true}))
	if s.notes.SetFieldSticky(d.TypeID, 5, true) == nil {
		t.Fatal("pin of a missing field accepted")
	}
	// Pins are stored with the note type, so backups include them.
	file := must(s.transfer.ExportBackup())
	other, _ := opened(t)
	check(t, other.transfer.Restore(must(other.transfer.StageBytes(file.Name, file.Data)).ID))
	nt := must(other.notes.Type(d.TypeID))
	if nt.FieldOptions[0].Sticky || !nt.FieldOptions[1].Sticky || !nt.FieldOptions[1].HTMLByDefault {
		t.Fatalf("options after restore = %+v", nt.FieldOptions)
	}
	check(t, other.notes.SetFieldSticky(d.TypeID, 1, false))
	if nt := must(other.notes.Type(d.TypeID)); nt.FieldOptions[1].Sticky || !nt.FieldOptions[1].HTMLByDefault {
		t.Fatalf("options = %+v", nt.FieldOptions)
	}
}

func TestDefaultsDeckByNoteType(t *testing.T) {
	s, store := opened(t)
	types := must(s.notes.Types())
	basic, cloze := types[0].ID, types[2].ID
	bio := must(s.decks.Create("Biology"))
	spanish := must(s.decks.Create("Spanish"))
	must(s.notes.Add(NewNote{TypeID: cloze, DeckID: bio, Fields: []string{"{{c1::ATP}}", ""}}))
	must(s.notes.Add(NewNote{TypeID: basic, DeckID: spanish, Fields: []string{"hola", "hello"}}))

	// "Current deck" (the default): one remembered deck for every note type.
	if st := must(s.app.Settings()); st.DeckByNoteType {
		t.Fatalf("settings = %+v", st)
	}
	if d := must(s.notes.Defaults(0)); d != (EditorDefaults{TypeID: basic, DeckID: spanish}) {
		t.Fatalf("defaults = %+v", d)
	}
	if d := must(s.notes.Defaults(cloze)); d != (EditorDefaults{TypeID: cloze, DeckID: spanish}) {
		t.Fatalf("cloze defaults = %+v", d)
	}

	// "Change deck depending on note type": the deck last used with each.
	check(t, s.app.SetDeckByNoteType(true))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if st := must(again.app.Settings()); !st.DeckByNoteType {
		t.Fatalf("settings after restart = %+v", st)
	}
	if d := must(again.notes.Defaults(cloze)); d != (EditorDefaults{TypeID: cloze, DeckID: bio}) {
		t.Fatalf("cloze defaults = %+v", d)
	}
	if d := must(again.notes.Defaults(0)); d != (EditorDefaults{TypeID: basic, DeckID: spanish}) {
		t.Fatalf("defaults = %+v", d)
	}
	// A note type without a remembered deck gets the last used deck.
	if d := must(again.notes.Defaults(types[1].ID)); d.DeckID != spanish {
		t.Fatalf("unused type defaults = %+v", d)
	}
	// A deleted deck is forgotten.
	check(t, again.decks.Delete(bio))
	if d := must(again.notes.Defaults(cloze)); d.DeckID != spanish {
		t.Fatalf("defaults after delete = %+v", d)
	}
}

func TestPreviewTemplate(t *testing.T) {
	s, store := opened(t)
	types := must(s.notes.Types())
	basic, cloze := types[0], types[2]

	// No notes yet: placeholder field text.
	p := must(s.notes.PreviewTemplate(basic.ID, basic.Templates[0], basic.CSS, 0, 0, false))
	if !strings.Contains(p.Question, "(Front)") || !strings.Contains(p.Answer, "(Back)") || len(p.ClozeNumbers) != 0 {
		t.Fatalf("placeholder = %+v", p)
	}
	if len(must(s.notes.PreviewNotes(basic.ID))) != 0 {
		t.Fatal("notes of an empty note type")
	}

	id := basicNote(t, s, "hola", "hello")
	if notes := must(s.notes.PreviewNotes(basic.ID)); len(notes) != 1 || notes[0] != (NoteChoice{ID: id, Label: "hola"}) {
		t.Fatalf("preview notes = %+v", notes)
	}
	cardID := must(s.notes.Get(id)).CardIDs[0]
	before := store.Clone()
	undo := must(s.app.UndoInfo())

	// Live update: the unsaved front template and CSS are used.
	tmpl := basic.Templates[0]
	tmpl.QFmt = "<b>{{Front}}</b>"
	p = must(s.notes.PreviewTemplate(basic.ID, tmpl, ".card { color: red }", id, 0, false))
	if !strings.Contains(p.Question, "<b>hola</b>") || p.Color != "red" {
		t.Fatalf("preview = %+v", p)
	}
	// The dark variant gets the night mode classes.
	if p := must(s.notes.PreviewTemplate(basic.ID, tmpl, "", id, 0, true)); !strings.Contains(p.Question, "nightMode") {
		t.Fatalf("night preview = %+v", p)
	}
	// Template errors show inline, as on cards.
	tmpl.QFmt = "{{#Front}}"
	if p := must(s.notes.PreviewTemplate(basic.ID, tmpl, "", id, 0, false)); !strings.Contains(p.Question, "is not closed") {
		t.Fatalf("error preview = %+v", p)
	}
	// Cards in the collection are unchanged.
	if d := must(s.browse.Card(cardID)); strings.Contains(d.Question, "<b>") {
		t.Fatalf("card changed: %+v", d)
	}
	if must(s.app.UndoInfo()) != undo || !reflect.DeepEqual(before, store.Clone()) {
		t.Fatal("preview changed the collection")
	}

	// Cloze: the note's cloze numbers, and the chosen cloze card.
	deck := must(s.notes.Defaults(0)).DeckID
	cid := must(s.notes.Add(NewNote{TypeID: cloze.ID, DeckID: deck, Fields: []string{"{{c1::uno}} {{c2::dos}}", ""}}))
	p = must(s.notes.PreviewTemplate(cloze.ID, cloze.Templates[0], cloze.CSS, cid, 2, false))
	if fmt.Sprint(p.ClozeNumbers) != "[1 2]" || !strings.Contains(p.Question, "uno") ||
		strings.Contains(p.Question, "dos") || !strings.Contains(p.Question, "card2") {
		t.Fatalf("cloze 2 = %+v", p)
	}
	// 0 and unknown numbers preview the first cloze card.
	for _, ord := range []int{0, 7} {
		p = must(s.notes.PreviewTemplate(cloze.ID, cloze.Templates[0], cloze.CSS, cid, ord, false))
		if strings.Contains(p.Question, "uno") || !strings.Contains(p.Question, "dos") {
			t.Fatalf("cloze %d = %+v", ord, p)
		}
	}
	// A cloze note type without a note previews cloze card 1.
	p = must(s.notes.PreviewTemplate(cloze.ID, cloze.Templates[0], cloze.CSS, 0, 0, false))
	if fmt.Sprint(p.ClozeNumbers) != "[1]" {
		t.Fatalf("cloze placeholder = %+v", p)
	}

	// A note of another note type is refused.
	if _, err := s.notes.PreviewTemplate(cloze.ID, cloze.Templates[0], "", id, 0, false); err == nil {
		t.Fatal("previewed a note of another note type")
	}
}

// Media file naming (card-browser spec): the same file twice is stored once,
// a different file under a taken name gets a new name, the extension stays.
func TestMediaAddNaming(t *testing.T) {
	s, _ := opened(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := s.app.Changes(ctx)

	if name := must(s.media.Add("cat.jpg", []byte("tabby"))); name != "cat.jpg" {
		t.Fatalf("first add = %q", name)
	}
	if ch := <-changes; !ch.Media {
		t.Fatalf("change = %+v", ch)
	}
	// Same image twice: one copy.
	if name := must(s.media.Add("cat.jpg", []byte("tabby"))); name != "cat.jpg" {
		t.Fatalf("same file again = %q", name)
	}
	// Same content under another name reuses the stored file.
	if name := must(s.media.Add("Downloads/kitty.jpg", []byte("tabby"))); name != "cat.jpg" {
		t.Fatalf("same content, other name = %q", name)
	}
	// Name collision: the existing file is unchanged.
	if name := must(s.media.Add("cat.jpg", []byte("siamese"))); name != "cat-1.jpg" {
		t.Fatalf("collision = %q", name)
	}
	if name := must(s.media.Add("cat.jpg", []byte("persian"))); name != "cat-2.jpg" {
		t.Fatalf("second collision = %q", name)
	}
	if img := must(s.media.Get("cat.jpg")); string(img) != "tabby" {
		t.Fatalf("existing file = %q", img)
	}
	if img := must(s.media.Get("cat-1.jpg")); string(img) != "siamese" {
		t.Fatalf("renamed file = %q", img)
	}
	// Extension kept; directories and characters that break markup dropped.
	if name := must(s.media.Add(`C:\Users\me\[hola]<1>.MP3`, []byte("hola"))); name != "hola1.MP3" {
		t.Fatalf("cleaned name = %q", name)
	}
	// No stem (pasted images, recordings): named after the content.
	png := must(s.media.Add(".png", []byte("screenshot")))
	if !strings.HasSuffix(png, ".png") || len(png) != len("0123456789abcdef.png") {
		t.Fatalf("content name = %q", png)
	}
	if again := must(s.media.Add(".png", []byte("screenshot"))); again != png {
		t.Fatalf("content name again = %q, want %q", again, png)
	}
}

// A media file no note uses is removed by compaction in a later session;
// files that notes use are kept, also those referenced by [sound:].
func TestMediaAddUnreferencedRemovedByCompaction(t *testing.T) {
	s, store := opened(t)
	img := must(s.media.Add("cat.jpg", []byte("tabby")))
	snd := must(s.media.Add("hola.mp3", []byte("hola")))
	must(s.media.Add("unsaved.png", []byte("never used")))
	basicNote(t, s, `<img src="`+img+`">`, "[sound:"+snd+"]")

	// The journal has ops, so opening compacts.
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if _, err := again.media.Get("unsaved.png"); err == nil {
		t.Fatal("unreferenced media kept")
	}
	for _, name := range []string{img, snd} {
		if _, err := again.media.Get(name); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	// The removed file's content can be added again.
	if name := must(again.media.Add("unsaved.png", []byte("never used"))); name != "unsaved.png" {
		t.Fatalf("re-added = %q", name)
	}
}

// Paste a screenshot (card-browser spec): stored as a new JPEG, or as PNG
// with the setting; the same image pasted twice is stored once.
func TestMediaAddPastedImage(t *testing.T) {
	s, _ := opened(t)
	shot := image.NewNRGBA(image.Rect(0, 0, 32, 24))
	shot.Set(1, 1, color.NRGBA{R: 255, A: 255}) // the rest is transparent
	var raw bytes.Buffer
	check(t, png.Encode(&raw, shot))

	name := must(s.media.AddPastedImage(raw.Bytes(), false))
	if !strings.HasSuffix(name, ".jpg") {
		t.Fatalf("JPEG paste stored as %q", name)
	}
	stored, err := jpeg.Decode(bytes.NewReader(must(s.media.Get(name))))
	if err != nil || stored.Bounds() != shot.Bounds() {
		t.Fatalf("stored JPEG: %v %v", stored, err)
	}
	if r, g, b, _ := stored.At(28, 20).RGBA(); r>>8 < 250 || g>>8 < 250 || b>>8 < 250 {
		t.Fatalf("transparent pixel = %v, want white", stored.At(28, 20))
	}
	if again := must(s.media.AddPastedImage(raw.Bytes(), false)); again != name {
		t.Fatalf("pasted twice: %q and %q", name, again)
	}

	// PNG paste: a PNG is stored unchanged.
	pngName := must(s.media.AddPastedImage(raw.Bytes(), true))
	if !strings.HasSuffix(pngName, ".png") || !bytes.Equal(must(s.media.Get(pngName)), raw.Bytes()) {
		t.Fatalf("PNG paste stored as %q", pngName)
	}
	// A JPEG on the clipboard becomes a PNG with the setting.
	jpg := must(s.media.Get(name))
	converted := must(s.media.AddPastedImage(jpg, true))
	if _, err := png.Decode(bytes.NewReader(must(s.media.Get(converted)))); err != nil {
		t.Fatalf("JPEG converted to PNG: %v", err)
	}
	if _, err := s.media.AddPastedImage([]byte("not an image"), false); err == nil {
		t.Fatal("garbage accepted")
	}
}

func TestCustomStudyIncreaseToday(t *testing.T) {
	s, _ := opened(t)
	id := must(s.decks.Create("Spanish"))
	node := func() DeckNode {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n
			}
		}
		t.Fatal("deck not found")
		return DeckNode{}
	}
	if o := node().Options; !o.NewIgnoresReviewLimit || o.LimitsStartFromTop {
		t.Fatalf("default options = %+v", o)
	}
	for i := range 25 {
		basicNoteIn(t, s, id, fmt.Sprint("q", i), "a")
	}
	if info := must(s.decks.CustomStudyInfo(id)); info.NewAvailable != 5 || info.ReviewAvailable != 0 {
		t.Fatalf("info = %+v", info)
	}
	// Ten more new cards (spec): the deck shows up to 10 more new cards
	// today (5 are left) and its options are unchanged.
	before := node().Options
	check(t, s.decks.IncreaseToday(id, 10, 0))
	if n := node(); n.Counts.New != 25 || !reflect.DeepEqual(n.Options, before) {
		t.Fatalf("after increase: counts %+v, options %+v", n.Counts, n.Options)
	}
	if info := must(s.decks.CustomStudyInfo(id)); info.NewAvailable != 0 {
		t.Fatalf("info = %+v", info)
	}
	// The increase lasts until the next study day.
	*s.clock = s.clock.Add(24 * time.Hour)
	if n := node(); n.Counts.New != 20 {
		t.Fatalf("next day new = %d", n.Counts.New)
	}
	if err := s.decks.IncreaseToday(id, -5, 0); err == nil {
		t.Fatal("negative increase accepted")
	}
}

func TestBrowseReposition(t *testing.T) {
	s, _ := opened(t)
	for _, w := range []string{"uno", "dos", "tres", "cuatro"} {
		basicNote(t, s, w, w)
	}
	all := must(s.browse.Search("", "created", false))
	rows := must(s.browse.Rows(all, nil))
	for i, r := range rows {
		if r.Position == nil || *r.Position != int64(i+1) {
			t.Fatalf("row %d position = %v", i, r.Position)
		}
	}
	// The last two to the front, making room.
	if skipped := must(s.browse.Reposition(all[2:], 0, 1, false, true)); skipped != 0 {
		t.Fatalf("skipped = %d", skipped)
	}
	byDue := must(s.browse.Search("", "due", false))
	if want := []int64{all[2], all[3], all[0], all[1]}; !slices.Equal(byDue, want) {
		t.Fatalf("due order = %v, want %v", byDue, want)
	}
	if d := must(s.browse.Card(all[2])); d.Position == nil || *d.Position != 0 {
		t.Fatalf("details position = %v", d.Position)
	}
	// A card that is not new has no position and is skipped.
	st := must(s.study.Start(must(s.notes.Defaults(0)).DeckID))
	must(s.study.Answer(st.Card.CardID, 4, 1000))
	if skipped := must(s.browse.Reposition(all, 10, 1, false, false)); skipped != 1 {
		t.Fatalf("skipped = %d", skipped)
	}
	if d := must(s.browse.Card(st.Card.CardID)); d.Position != nil {
		t.Fatalf("reviewed card position = %v", *d.Position)
	}
	if _, err := s.browse.Reposition(all, 0, 0, false, false); err == nil || !strings.Contains(err.Error(), "step") {
		t.Fatalf("err = %v", err)
	}
}

func TestInsertionOrderOption(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Words"))
	var o DeckOptions
	for _, n := range must(s.decks.Tree()) {
		if n.ID == id {
			o = n.Options
		}
	}
	if o.InsertionOrder != "sequential" {
		t.Fatalf("default insertion order = %q", o.InsertionOrder)
	}
	bad := o
	bad.InsertionOrder = "sideways"
	if err := s.decks.SetOptions(id, bad); err == nil || !strings.Contains(err.Error(), "insertion order") {
		t.Fatalf("err = %v", err)
	}
	o.InsertionOrder = "random"
	check(t, s.decks.SetOptions(id, o))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	for _, n := range must(again.decks.Tree()) {
		if n.ID == id && n.Options.InsertionOrder != "random" {
			t.Fatalf("options after restart = %+v", n.Options)
		}
	}
}

func TestDeckPresetsService(t *testing.T) {
	s, store := opened(t)
	spanish := must(s.decks.Create("Languages::Spanish"))
	french := must(s.decks.Create("Languages::French"))
	var parent int64
	for _, n := range must(s.decks.Tree()) {
		if n.FullName == "Languages" {
			parent = n.ID
		}
	}
	st := must(s.decks.Settings(spanish))
	if len(st.Presets) != 1 || st.PresetID != st.DefaultPresetID || st.Presets[0].Name != "Default" ||
		st.Presets[0].Decks != 4 || st.Limits.NewPerDay != nil {
		t.Fatalf("settings = %+v", st)
	}
	// A new preset for the parent, saved to its subdecks.
	id := must(s.decks.AddPreset(parent, "Languages"))
	check(t, s.decks.SavePresetToSubdecks(parent))
	// Saving from Spanish changes the shared preset; French gets it too.
	o := must(s.decks.Settings(spanish)).Presets[1].Options
	o.DesiredRetention = 0.85
	five := 5
	check(t, s.decks.SaveSettings(spanish, id, o, DeckLimits{NewPerDay: &five}))
	st = must(s.decks.Settings(french))
	if st.PresetID != id || st.Limits.NewPerDay != nil || st.Presets[1].Options.DesiredRetention != 0.85 || st.Presets[1].Decks != 3 {
		t.Fatalf("French settings = %+v", st)
	}
	if st := must(s.decks.Settings(spanish)); *st.Limits.NewPerDay != 5 {
		t.Fatalf("Spanish limits = %+v", st.Limits)
	}
	check(t, s.decks.RenamePreset(id, "Romance"))
	clone := must(s.decks.ClonePreset(french, "French only"))
	if err := s.decks.DeletePreset(st.DefaultPresetID); err == nil {
		t.Fatal("Default deleted")
	}
	check(t, s.decks.DeletePreset(clone))
	// The presets survive a restart.
	again := newServices(t, store)
	must(again.app.Open("/data"))
	st = must(again.decks.Settings(french))
	if len(st.Presets) != 2 || st.PresetID != st.DefaultPresetID || st.Presets[1].Name != "Romance" || st.Presets[1].Decks != 2 {
		t.Fatalf("after restart = %+v", st)
	}
}

func collectExport(t *testing.T, events <-chan ExportEvent) (progress int, last ExportEvent) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case e, ok := <-events:
			if !ok {
				return progress, last
			}
			if e.File == nil && e.Error == "" {
				progress++
			}
			last = e
		case <-timeout:
			t.Fatal("export stream did not finish")
		}
	}
}

func TestExportDeckAndReimport(t *testing.T) {
	s, _ := opened(t)
	basicNote(t, s, "elsewhere", "not exported")
	spanish := must(s.decks.Create("Languages::Spanish"))
	basicNoteIn(t, s, spanish, "hola", "hello")
	basicNoteIn(t, s, spanish, `perro <img src="dog.png">`, "dog")
	png := must(s.media.Add("dog.png", []byte("\x89PNG fake")))
	if png != "dog.png" {
		t.Fatalf("media stored as %s", png)
	}
	types := len(must(s.notes.Types()))

	progress, last := collectExport(t, s.transfer.ExportDeck(context.Background(), spanish, ExportOptions{Media: true}))
	if last.File == nil || last.File.Name != "Spanish.apkg" || last.Stage != "done" || progress == 0 {
		t.Fatalf("export: progress %d, last %+v", progress, last)
	}
	info := must(s.transfer.StageBytes(last.File.Name, last.File.Data))
	in := must(s.transfer.Inspect(info.ID))
	if in.Apkg == nil || in.Apkg.Notes != 2 || in.Apkg.Cards != 2 || in.Apkg.Media != 1 ||
		strings.Join(in.Apkg.Decks, ",") != "Languages::Spanish" {
		t.Fatalf("exported package = %+v", in.Apkg)
	}

	// Importing the export into the same collection duplicates nothing:
	// notes keep their GUIDs, and the note type is recognised.
	_, done := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{KeepScheduling: true}))
	if done.Summary == nil || done.Summary.NotesAdded != 0 || done.Summary.NotesSkipped != 2 || done.Summary.CardsAdded != 0 {
		t.Fatalf("re-import = %+v", done)
	}
	if n := len(must(s.browse.Search("", "created", false))); n != 3 {
		t.Fatalf("cards after re-import = %d, want 3", n)
	}
	if n := len(must(s.notes.Types())); n != types {
		t.Fatalf("note types after re-import = %d, want %d", n, types)
	}

	// A missing deck fails in the stream.
	if _, last := collectExport(t, s.transfer.ExportDeck(context.Background(), 12345, ExportOptions{})); last.Error == "" {
		t.Fatalf("missing deck = %+v", last)
	}
}

func TestExportSelectedNotes(t *testing.T) {
	s, _ := opened(t)
	reversed := must(s.notes.Types())[1]
	d := must(s.notes.Defaults(0))
	var picked []int64
	for i := range 8 {
		note := must(s.notes.Add(NewNote{TypeID: reversed.ID, DeckID: d.DeckID, Fields: []string{fmt.Sprint("front ", i), "back"}}))
		cards := must(s.notes.Get(note)).CardIDs
		// Ten cards of eight notes: both cards of the first two.
		picked = append(picked, cards[0])
		if i < 2 {
			picked = append(picked, cards[1])
		}
	}
	basicNote(t, s, "not", "selected")
	if len(picked) != 10 {
		t.Fatalf("picked %d cards", len(picked))
	}
	_, last := collectExport(t, s.transfer.ExportNotes(context.Background(), picked, ExportOptions{}))
	if last.File == nil || last.File.Name != "Selected notes.apkg" {
		t.Fatalf("export = %+v", last)
	}
	in := must(s.transfer.Inspect(must(s.transfer.StageBytes(last.File.Name, last.File.Data)).ID))
	if in.Apkg.Notes != 8 || in.Apkg.Cards != 16 {
		t.Fatalf("package has %d notes and %d cards, want 8 and 16", in.Apkg.Notes, in.Apkg.Cards)
	}
	if _, last := collectExport(t, s.transfer.ExportNotes(context.Background(), nil, ExportOptions{})); last.Error == "" {
		t.Fatalf("empty selection = %+v", last)
	}
}

func TestExportCollection(t *testing.T) {
	s, _ := opened(t)
	basicNote(t, s, "hola", "hello")
	d := must(s.notes.Defaults(0))
	card := must(s.browse.Search("", "created", false))[0]
	must(s.study.Start(d.DeckID))
	must(s.study.Answer(card, 4, 4000))
	_, last := collectExport(t, s.transfer.ExportCollection(context.Background(), ExportOptions{}))
	if last.File == nil || last.File.Name != "collection-2026-10-13.colpkg" {
		t.Fatalf("export = %+v", last)
	}
	p, err := apkg.Read(last.File.Data, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Scheduling and history are always included.
	if len(p.Cards) != 1 || p.Cards[0].Type == 0 || len(p.Revlog) != 1 {
		t.Fatalf("collection package: cards %+v, %d reviews", p.Cards, len(p.Revlog))
	}
}

func TestExportNotesTextAndReimport(t *testing.T) {
	// Edit outside and re-import (spec): export a deck's notes with tags,
	// change a Back field, import with "update": the note is updated and
	// nothing is duplicated.
	s, _ := opened(t)
	basicNote(t, s, "elsewhere", "not exported")
	spanish := must(s.decks.Create("Languages::Spanish"))
	hola := basicNoteIn(t, s, spanish, "hola", "hello")
	basicNoteIn(t, s, spanish, "perro", "dog\nhound")
	check(t, s.notes.Update(hola, []string{"hola", "hello"}, []string{"greeting"}))

	progress, last := collectExport(t, s.transfer.ExportNotesText(context.Background(), TextScope{DeckID: spanish}, TextOptions{Tags: true}))
	if last.File == nil || last.File.Name != "Spanish.txt" || progress == 0 {
		t.Fatalf("export: progress %d, last %+v", progress, last)
	}
	want := "#separator:tab\n#html:true\n#tags column:3\nhola\thello\tgreeting\nperro\t\"dog\nhound\"\t\n"
	if got := string(last.File.Data); got != want {
		t.Fatalf("file =\n%q\nwant\n%q", got, want)
	}

	edited := strings.Replace(want, "hola\thello", "hola\thi there", 1)
	info := must(s.transfer.StageBytes(last.File.Name, []byte(edited)))
	in := must(s.transfer.Inspect(info.ID))
	if in.Csv == nil || in.Csv.Delimiter != "tab" || in.Csv.TotalRows != 2 {
		t.Fatalf("preview = %+v", in.Csv)
	}
	_, done := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{
		TypeID: in.Csv.TypeID, DeckID: spanish, Mapping: in.Csv.Mapping, Duplicates: "update", AllowHTML: in.Csv.AllowHTML}))
	if done.Summary == nil || done.Summary.NotesAdded != 0 || done.Summary.NotesUpdated != 1 || done.Summary.NotesSkipped != 1 {
		t.Fatalf("re-import = %+v", done)
	}
	if n := must(s.notes.Get(hola)); n.Fields[1] != "hi there" {
		t.Fatalf("note = %+v", n)
	}
	if n := len(must(s.browse.Search("", "created", false))); n != 3 {
		t.Fatalf("cards after re-import = %d, want 3", n)
	}
}

func TestExportTextSelectionAndCollection(t *testing.T) {
	s, _ := opened(t)
	d := must(s.notes.Defaults(0))
	var picked []int64
	for i := range 6 {
		note := basicNote(t, s, fmt.Sprint("front ", i), fmt.Sprint("<b>back</b> ", i))
		if i < 5 {
			picked = append(picked, must(s.notes.Get(note)).CardIDs...)
		}
	}

	// Export selected notes (spec): 5 cards of 5 notes give 5 note lines.
	_, last := collectExport(t, s.transfer.ExportNotesText(context.Background(), TextScope{CardIDs: picked}, TextOptions{}))
	if last.File == nil || last.File.Name != "Selected notes.txt" {
		t.Fatalf("notes export = %+v", last)
	}
	if lines := strings.Split(strings.TrimSpace(string(last.File.Data)), "\n"); len(lines) != 2+5 || lines[2] != "front 0\t<b>back</b> 0" {
		t.Fatalf("notes file = %q", last.File.Data)
	}

	// Selected cards as plain text, without HTML.
	_, last = collectExport(t, s.transfer.ExportCardsText(context.Background(), TextScope{CardIDs: picked[:2]}))
	if last.File == nil || last.File.Name != "Selected cards.txt" ||
		string(last.File.Data) != "#separator:tab\n#html:false\nfront 0\tback 0\nfront 1\tback 1\n" {
		t.Fatalf("cards export = %+v (%q)", last, last.File.Data)
	}

	// A deck's cards, and the whole collection.
	progress, last := collectExport(t, s.transfer.ExportCardsText(context.Background(), TextScope{DeckID: d.DeckID}))
	if last.File == nil || last.File.Name != "Default.txt" || strings.Count(string(last.File.Data), "\n") != 2+6 || progress == 0 {
		t.Fatalf("deck cards export: progress %d, %+v", progress, last)
	}
	_, last = collectExport(t, s.transfer.ExportNotesText(context.Background(), TextScope{All: true}, TextOptions{Deck: true}))
	if last.File == nil || last.File.Name != "notes-2026-10-13.txt" || !strings.Contains(string(last.File.Data), "#deck column:1\nDefault\tfront 0\t") {
		t.Fatalf("collection export = %+v (%q)", last, last.File.Data)
	}

	// Nothing selected fails in the stream.
	if _, last := collectExport(t, s.transfer.ExportNotesText(context.Background(), TextScope{}, TextOptions{})); last.Error == "" {
		t.Fatalf("empty notes selection = %+v", last)
	}
	if _, last := collectExport(t, s.transfer.ExportCardsText(context.Background(), TextScope{})); last.Error == "" {
		t.Fatalf("empty cards selection = %+v", last)
	}
}

// The update options reach the import, and the summary reports what
// happened to notes and note types.
func TestTransferApkgUpdatePolicy(t *testing.T) {
	s, _ := opened(t)
	importFile := func(name string, o ImportOptions) ImportSummary {
		t.Helper()
		data, err := os.ReadFile("importer/apkg/testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		info := must(s.transfer.StageBytes(name, data))
		_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, o))
		if last.Summary == nil {
			t.Fatalf("%s: %+v", name, last)
		}
		return *last.Summary
	}
	if sum := importFile("update-v1.apkg", ImportOptions{}); sum.NotesAdded != 2 || sum.NoteTypesAdded != 1 {
		t.Fatalf("v1 = %+v", sum)
	}
	if sum := importFile("update-v2.apkg", ImportOptions{UpdateNotes: "never", UpdateNoteTypes: "never"}); sum.NotesUpdated != 0 || sum.NoteTypesUpdated != 0 {
		t.Fatalf("v2 never = %+v", sum)
	}
	if sum := importFile("update-v1.apkg", ImportOptions{UpdateNotes: "always"}); sum.NotesUpdated != 0 || sum.NotesSkipped != 2 {
		t.Fatalf("v1 always = %+v", sum)
	}
	sum := importFile("update-v3.apkg", ImportOptions{MergeNoteTypes: true})
	if sum.NoteTypesMerged != 1 || sum.NoteTypesAdded != 0 || sum.NotesUpdated != 1 || sum.NotesAdded != 1 {
		t.Fatalf("v3 merged = %+v", sum)
	}
	// Without merging, version 2's note type no longer matches the merged
	// one: it is added separately, and the existing notes stay.
	if sum := importFile("update-v2.apkg", ImportOptions{}); sum.NoteTypesAdded != 1 || sum.NotesUpdated != 0 || sum.NotesSkipped != 2 {
		t.Fatalf("v2 after v3 = %+v", sum)
	}
	data, _ := os.ReadFile("importer/apkg/testdata/update-v1.apkg")
	info := must(s.transfer.StageBytes("update-v1.apkg", data))
	_, last := collectImport(t, s.transfer.Import(context.Background(), info.ID, ImportOptions{UpdateNotes: "sometimes"}))
	if !strings.Contains(last.Error, "unknown update policy") {
		t.Fatalf("bad policy = %+v", last)
	}
}

// Check media (persistence spec): unused and missing files, a static
// template file, deleting into the trash and restoring, tagging notes with
// missing media.
func TestCheckMediaService(t *testing.T) {
	s, store := opened(t)
	if r := must(s.app.CheckMedia()); r.Unused == nil || r.Missing == nil || r.Trash == nil || len(r.Unused)+len(r.Missing)+len(r.Trash) != 0 {
		t.Fatalf("empty report = %+v", r)
	}
	old := must(s.media.Add("old.jpg", []byte("old picture")))
	logo := must(s.media.Add("_logo.png", []byte("logo")))
	id := basicNote(t, s, `<img src="`+old+`">`, "")

	// In a later session the note stops using old.jpg; nothing uses
	// _logo.png. Three notes refer to missing files.
	s = newServices(t, store)
	must(s.app.Open("/data"))
	check(t, s.notes.Update(id, []string{"no picture", ""}, nil))
	var broken []int64
	for _, f := range []string{"[sound:gone.mp3]", `<img src="gone.png">`, "[sound:gone.mp3] again"} {
		broken = append(broken, basicNote(t, s, f, ""))
	}
	r := must(s.app.CheckMedia())
	if len(r.Unused) != 1 || r.Unused[0].Name != old || r.Unused[0].Size != int64(len("old picture")) || r.UnusedBytes != r.Unused[0].Size {
		t.Fatalf("unused = %+v (%d bytes)", r.Unused, r.UnusedBytes)
	}
	if strings.Join(r.Missing, ",") != "gone.mp3,gone.png" || r.MissingNotes != 3 {
		t.Fatalf("missing = %q in %d notes", r.Missing, r.MissingNotes)
	}

	// Tag missing (spec: find notes to fix).
	if n := must(s.app.TagMissingMedia()); n != 3 {
		t.Fatalf("tagged %d notes, want 3", n)
	}
	for _, id := range broken {
		if tags := must(s.notes.Get(id)).Tags; strings.Join(tags, " ") != "missing-media" {
			t.Fatalf("note %d tags = %q", id, tags)
		}
	}

	// Delete unused, then restore (spec: restore deleted files).
	if n := must(s.app.DeleteUnusedMedia()); n != 1 {
		t.Fatalf("deleted %d files, want 1", n)
	}
	if _, err := s.media.Get(old); err == nil {
		t.Fatal("deleted file still in the media")
	}
	if _, err := s.media.Get(logo); err != nil {
		t.Fatalf("static file deleted: %v", err)
	}
	r = must(s.app.CheckMedia())
	if len(r.Unused) != 0 || len(r.Trash) != 1 || r.Trash[0].Name != old || r.TrashBytes != int64(len("old picture")) {
		t.Fatalf("after delete: %+v", r)
	}
	if n := must(s.app.RestoreMediaTrash()); n != 1 {
		t.Fatalf("restored %d files, want 1", n)
	}
	if data := must(s.media.Get(old)); string(data) != "old picture" {
		t.Fatalf("restored %q", data)
	}

	// Empty trash deletes for good.
	must(s.app.DeleteUnusedMedia())
	check(t, s.app.EmptyMediaTrash())
	if r := must(s.app.CheckMedia()); len(r.Trash) != 0 || r.TrashBytes != 0 {
		t.Fatalf("after empty: %+v", r)
	}
	if n := must(s.app.RestoreMediaTrash()); n != 0 {
		t.Fatalf("restored %d files from an empty trash", n)
	}
}

// Compaction moves unreferenced media to the trash instead of deleting
// them, so they can be restored (persistence spec: media trash).
func TestCompactionMovesMediaToTrash(t *testing.T) {
	s, store := opened(t)
	must(s.media.Add("unsaved.png", []byte("never used")))
	basicNote(t, s, "x", "")
	// The journal has ops, so opening compacts.
	s = newServices(t, store)
	must(s.app.Open("/data"))
	r := must(s.app.CheckMedia())
	if len(r.Trash) != 1 || r.Trash[0].Name != "unsaved.png" {
		t.Fatalf("trash = %+v", r.Trash)
	}
	must(s.app.RestoreMediaTrash())
	if data := must(s.media.Get("unsaved.png")); string(data) != "never used" {
		t.Fatalf("restored %q", data)
	}
}

// corruptStored appends an op that bypasses validation to the stored
// journal, as a bug would; the next Open replays it.
func corruptStored(t *testing.T, store *storage.MemoryStore, op string) {
	t.Helper()
	b := store.Open()
	var ptr struct {
		Gen int `json:"gen"`
	}
	check(t, json.Unmarshal(must(b.Get("current")), &ptr))
	check(t, b.Append(fmt.Sprintf("g%d/journal", ptr.Gen), []byte(`{"op":`+op+`}`)))
}

// cardCount counts the cards in the open collection; searching would
// fail on a card without a note.
func cardCount(t *testing.T, s *services) int {
	t.Helper()
	n := 0
	for _, count := range must(s.app.a.collection()).CardsPerDeck() {
		n += count
	}
	return n
}

func collectCheck(t *testing.T, events <-chan CheckDatabaseEvent) (stages []string, last CheckDatabaseEvent) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case e, ok := <-events:
			if !ok {
				return stages, last
			}
			if !slices.Contains(stages, e.Stage) {
				stages = append(stages, e.Stage)
			}
			last = e
		case <-timeout:
			t.Fatal("check stream did not finish")
		}
	}
}

// failingAppend fails journal appends while fail is set, as a full disk
// would.
type failingAppend struct {
	storage.Backend
	fail *bool
}

func (f failingAppend) Append(key string, record []byte) error {
	if *f.fail {
		return errors.New("disk full")
	}
	return f.Backend.Append(key, record)
}

// Check database (persistence spec): healthy collection, with a backup first
// and progress.
func TestCheckDatabaseServiceHealthy(t *testing.T) {
	s, store := opened(t)
	basicNote(t, s, "hola", "hello")
	stages, last := collectCheck(t, s.app.CheckDatabase(context.Background()))
	if last.Error != "" || last.Report == nil || last.Report.Problems == nil || len(last.Report.Problems) != 0 {
		t.Fatalf("last = %+v", last)
	}
	if !slices.Contains(stages, "backup") || stages[len(stages)-1] != "done" {
		t.Fatalf("stages = %v", stages)
	}
	if list := autoBackupsIn(t, store); len(list) != 1 {
		t.Fatalf("backups = %d, want 1", len(list))
	}
}

// Check database (persistence spec): orphaned card.
func TestCheckDatabaseServiceOrphanedCard(t *testing.T) {
	s, store := opened(t)
	id := basicNote(t, s, "hola", "hello")
	corruptStored(t, store, fmt.Sprintf(`{"k":"corrupt","dn":[%d]}`, id))
	s = newServices(t, store)
	must(s.app.Open("/data"))
	if n := cardCount(t, s); n != 1 {
		t.Fatalf("cards before = %d", n)
	}
	_, last := collectCheck(t, s.app.CheckDatabase(context.Background()))
	if last.Report == nil || len(last.Report.Problems) != 1 ||
		usermsg.Plain(last.Report.Problems[0].Message) != "1 card without a note was removed." {
		t.Fatalf("last = %+v", last)
	}
	if n := cardCount(t, s); n != 0 {
		t.Fatalf("cards after = %d", n)
	}
	// The repair was saved.
	s = newServices(t, store)
	must(s.app.Open("/data"))
	if n := cardCount(t, s); n != 0 {
		t.Fatalf("cards after restart = %d", n)
	}
}

// Check database (persistence spec): if the repair cannot be saved, the
// collection stays unchanged.
func TestCheckDatabaseServiceFailedSave(t *testing.T) {
	store := storage.NewMemoryStore()
	s := newServices(t, store)
	must(s.app.Open("/data"))
	id := basicNote(t, s, "hola", "hello")
	corruptStored(t, store, fmt.Sprintf(`{"k":"corrupt","dn":[%d]}`, id))
	s = newServices(t, store)
	fail := false
	s.app.a.openBackend = func(string) (storage.Backend, error) { return failingAppend{store.Open(), &fail}, nil }
	must(s.app.Open("/data"))
	fail = true
	_, last := collectCheck(t, s.app.CheckDatabase(context.Background()))
	if last.Error == "" || last.Report != nil {
		t.Fatalf("last = %+v", last)
	}
	if n := cardCount(t, s); n != 1 {
		t.Fatalf("cards = %d, want the orphaned card kept", n)
	}
	fail = false
	s = newServices(t, store)
	must(s.app.Open("/data"))
	if n := cardCount(t, s); n != 1 {
		t.Fatalf("cards after restart = %d", n)
	}
}

// The check does not start when the backup before it fails.
func TestCheckDatabaseServiceStopsWhenBackupFails(t *testing.T) {
	s, store := opened(t)
	id := basicNote(t, s, "hola", "hello")
	corruptStored(t, store, fmt.Sprintf(`{"k":"corrupt","dn":[%d]}`, id))
	s = newServices(t, store)
	must(s.app.Open("/data"))
	store.FailPut = func(key string) bool { return strings.HasPrefix(key, "backups/") }
	_, last := collectCheck(t, s.app.CheckDatabase(context.Background()))
	if !strings.Contains(last.Error, "could not be backed up") {
		t.Fatalf("last = %+v", last)
	}
	if n := cardCount(t, s); n != 1 {
		t.Fatal("collection changed although the backup failed")
	}
}

func TestStatsService(t *testing.T) {
	s, _ := opened(t)
	spanish := must(s.decks.Create("Spanish"))
	verbs := must(s.decks.Create("Spanish::Verbs"))
	french := must(s.decks.Create("French"))
	basicNoteIn(t, s, spanish, "uno", "one")
	basicNoteIn(t, s, verbs, "ser", "to be")
	frNote := basicNoteIn(t, s, french, "un", "one")
	for _, deck := range []int64{spanish, french} {
		for {
			st := must(s.study.Start(deck))
			if st.Card == nil {
				break
			}
			must(s.study.Answer(st.Card.CardID, 4, 5000))
		}
	}
	check(t, s.notes.Delete([]int64{frNote}))

	collection := stats.Scope{Kind: stats.ScopeCollection}
	// Collection statistics include the deleted French card's answer.
	if today := must(s.stats.Today(collection, stats.PeriodMonth)); today.Cards != 3 || today.Time != 15*time.Second || today.Learning != 3 {
		t.Fatalf("collection today = %+v", today)
	}
	// Deck scope (spec): Spanish with its subdeck.
	deck := stats.Scope{Kind: stats.ScopeDeck, DeckID: spanish}
	if today := must(s.stats.Today(deck, stats.PeriodMonth)); today.Cards != 2 {
		t.Fatalf("Spanish today = %+v", today)
	}
	if counts := must(s.stats.Counts(deck, stats.PeriodMonth)); counts.Total != 2 || counts.Young != 2 {
		t.Fatalf("Spanish counts = %+v", counts)
	}
	search := stats.Scope{Kind: stats.ScopeSearch, Query: "ser"}
	if counts := must(s.stats.Counts(search, stats.PeriodMonth)); counts.Total != 1 {
		t.Fatalf("search counts = %+v", counts)
	}
	reviews := must(s.stats.Reviews(collection, stats.PeriodMonth))
	if reviews.Unit != stats.UnitDay || len(reviews.Buckets) != 30 || reviews.Buckets[29].Learning != 3 {
		t.Fatalf("reviews = %+v", reviews)
	}
	if rt := must(s.stats.ReviewTime(collection, stats.PeriodYear)); rt.Unit != stats.UnitWeek || rt.Total != 15000 {
		t.Fatalf("review time = %+v", rt)
	}
	if f := must(s.stats.FutureDue(deck, stats.Period3Months)); len(f.Days) != 90 || f.Total != 2 {
		t.Fatalf("future due = %+v", f)
	}
	if c := must(s.stats.Calendar(collection, stats.PeriodMonth)); len(c.Days) != 1 || c.Days[0].Count != 3 || c.Today != "2026-10-13" {
		t.Fatalf("calendar = %+v", c)
	}
	if h := must(s.stats.Hourly(collection, stats.PeriodMonth)); len(h) != 24 || h[12].Reviews != 3 {
		t.Fatalf("hourly = %+v", h)
	}
	if b := must(s.stats.Buttons(collection, stats.PeriodMonth)); b.Learning.Counts[3] != 3 || b.Learning.Correct != 1 {
		t.Fatalf("buttons = %+v", b)
	}
	if h := must(s.stats.Intervals(deck, stats.PeriodMonth)); h.Count != 2 {
		t.Fatalf("intervals = %+v", h)
	}
	for name, fn := range map[string]func(stats.Scope, string) (stats.Histogram, error){
		"stability": s.stats.Stability, "difficulty": s.stats.Difficulty, "retrievability": s.stats.Retrievability,
	} {
		if h := must(fn(deck, stats.PeriodMonth)); h.Count != 2 {
			t.Errorf("%s = %+v", name, h)
		}
	}
	if r := must(s.stats.TrueRetention(deck, stats.PeriodMonth)); r.Desired != core.DefaultRetention || len(r.Rows) != 6 {
		t.Fatalf("true retention = %+v", r)
	}
	if _, err := s.stats.Today(collection, "week"); err == nil {
		t.Fatal("unknown period accepted")
	}
	if _, err := s.stats.Today(stats.Scope{Kind: stats.ScopeSearch, Query: "is:bogus"}, stats.PeriodMonth); err == nil {
		t.Fatal("invalid search accepted")
	}

	// A new answer shows up at once.
	basicNoteIn(t, s, spanish, "dos", "two")
	st := must(s.study.Start(spanish))
	must(s.study.Answer(st.Card.CardID, 1, 1000))
	if today := must(s.stats.Today(deck, stats.PeriodMonth)); today.Cards != 3 || today.Again != 1 {
		t.Fatalf("after another answer = %+v", today)
	}
}

func TestBrowseFindDuplicates(t *testing.T) {
	s, _ := opened(t)
	a := basicNote(t, s, "perro", "dog")
	b := basicNote(t, s, "Perro ", "dog")
	basicNote(t, s, "gato", "cat")
	sets := must(s.browse.FindDuplicates("Front", ""))
	if len(sets) != 1 || sets[0].Content != "perro" || fmt.Sprint(sets[0].NoteIDs) != fmt.Sprint([]int64{a, b}) {
		t.Fatalf("sets = %+v", sets)
	}
	if sets := must(s.browse.FindDuplicates("Front", "deck:nope")); len(sets) != 0 {
		t.Fatalf("limited = %+v", sets)
	}
	if _, err := s.browse.FindDuplicates("Front", "prop:nope>1"); err == nil {
		t.Fatal("invalid search accepted")
	}
	// Tag duplicates: the notes' cards get the tag through AddTags.
	cards := must(s.browse.NoteCards(sets[0].NoteIDs))
	must(0, s.browse.AddTags(cards, []string{"duplicate"}))
	if n := len(must(s.browse.SearchNotes("tag:duplicate", "created", false))); n != 2 {
		t.Fatalf("tagged = %d", n)
	}
}

func TestBrowseFindReplace(t *testing.T) {
	s, _ := opened(t)
	a := basicNote(t, s, "recieve", "5 kg")
	b := basicNote(t, s, "other", "recieve it", "ch01")
	basicNote(t, s, "recieve too", "")
	picked := must(s.notes.Get(a)).CardIDs
	picked = append(picked, must(s.notes.Get(b)).CardIDs...)
	fix := FindReplaceOptions{Find: "recieve", Replace: "receive", IgnoreCase: true}
	// Selected notes only, then every note.
	if n := must(s.browse.FindReplacePreview(picked, false, fix)); n != 2 {
		t.Fatalf("preview = %d", n)
	}
	if n := must(s.browse.FindReplacePreview(nil, true, fix)); n != 3 {
		t.Fatalf("preview all = %d", n)
	}
	if n := must(s.browse.FindReplace(picked, false, fix)); n != 2 {
		t.Fatalf("changed = %d", n)
	}
	if n := len(must(s.browse.Search("recieve", "created", false))); n != 1 {
		t.Fatalf("left = %d", n)
	}
	kg := FindReplaceOptions{Find: `(\d+) kg`, Replace: "$1 kilograms", Field: "Back", Regex: true}
	must(s.browse.FindReplace(nil, true, kg))
	if f := must(s.notes.Get(a)).Fields; f[0] != "receive" || f[1] != "5 kilograms" {
		t.Fatalf("fields = %q", f)
	}
	must(s.browse.FindReplace(picked, false, FindReplaceOptions{Find: "ch0", Replace: "chapter", Tags: true}))
	if tags := must(s.notes.Get(b)).Tags; len(tags) != 1 || tags[0] != "chapter1" {
		t.Fatalf("tags = %q", tags)
	}
	if _, err := s.browse.FindReplacePreview(picked, false, FindReplaceOptions{Find: "(", Regex: true}); err == nil ||
		!strings.Contains(err.Error(), "Invalid regular expression") {
		t.Fatalf("err = %v", err)
	}
	if names := must(s.browse.FieldNames(picked, false)); strings.Join(names, ",") != "Front,Back" {
		t.Fatalf("fields = %q", names)
	}
	if names := must(s.browse.FieldNames(nil, false)); len(names) != 0 {
		t.Fatalf("no selection fields = %q", names)
	}
}

func TestBrowseChangeNoteType(t *testing.T) {
	s, _ := opened(t)
	types := must(s.notes.Types()) // Basic, Basic (and reversed card), Cloze
	a := basicNote(t, s, "hola", "hello")
	b := basicNote(t, s, "gato", "cat")
	picked := append(must(s.notes.Get(a)).CardIDs, must(s.notes.Get(b)).CardIDs...)
	if id := must(s.browse.NoteTypeOfCards(picked)); id != types[0].ID {
		t.Fatalf("note type = %d", id)
	}
	preview := must(s.browse.ChangeNoteTypePreview(picked, types[1].ID, []int{0, 1}, []int{0, -1}))
	if preview.Notes != 2 || preview.CardsCreated != 2 || preview.CardsDeleted != 0 {
		t.Fatalf("preview = %+v", preview)
	}
	done := must(s.browse.ChangeNoteType(picked, types[1].ID, []int{0, 1}, []int{0, -1}))
	if !reflect.DeepEqual(done, preview) {
		t.Fatalf("done = %+v, preview = %+v", done, preview)
	}
	n := must(s.notes.Get(a))
	if n.TypeID != types[1].ID || len(n.CardIDs) != 2 || n.CardIDs[0] != picked[0] {
		t.Fatalf("note = %+v", n)
	}
	// Mixed selection: the notes must share one note type.
	cloze := must(s.notes.Add(NewNote{TypeID: types[2].ID, DeckID: must(s.notes.Defaults(0)).DeckID, Fields: []string{"{{c1::x}}", ""}}))
	mixed := append(n.CardIDs, must(s.notes.Get(cloze)).CardIDs...)
	if _, err := s.browse.NoteTypeOfCards(mixed); err == nil || !strings.Contains(err.Error(), "same note type") {
		t.Fatalf("err = %v", err)
	}
	if _, err := s.browse.ChangeNoteType(mixed, types[0].ID, []int{0, 1}, []int{0}); err == nil {
		t.Fatal("mixed selection changed")
	}
}

// Deck overview (app-shell spec): counts, buried cards, description and,
// when nothing is due, the next due time; Unbury makes buried cards due.
func TestDeckOverview(t *testing.T) {
	s, _ := opened(t)
	deck := must(s.decks.Create("Spanish"))
	ov := must(s.decks.Overview(deck))
	if ov.Counts != (Counts{}) || ov.Buried != 0 || ov.NextDue != nil || ov.DescriptionHTML != "" {
		t.Fatalf("empty deck overview = %+v", ov)
	}
	for _, w := range []string{"uno", "dos", "tres"} {
		basicNoteIn(t, s, deck, w, "x")
	}
	if ov := must(s.decks.Overview(deck)); ov.Counts.New != 3 || ov.NextDue != nil {
		t.Fatalf("overview = %+v", ov)
	}
	// Studying from the overview buries a card; Unbury brings it back.
	st := must(s.study.Start(deck))
	must(s.study.Bury(st.Card.CardID))
	if ov := must(s.decks.Overview(deck)); ov.Counts.New != 2 || ov.Buried != 1 {
		t.Fatalf("overview after bury = %+v", ov)
	}
	check(t, s.decks.Unbury(deck))
	if ov := must(s.decks.Overview(deck)); ov.Counts.New != 3 || ov.Buried != 0 {
		t.Fatalf("overview after unbury = %+v", ov)
	}

	// Edit description (collection spec).
	check(t, s.decks.SetDescription(deck, "Vocabulary from <b>Book 1</b><script>x()</script>"))
	if ov := must(s.decks.Overview(deck)); ov.Description != "Vocabulary from <b>Book 1</b><script>x()</script>" ||
		ov.DescriptionHTML != "Vocabulary from <b>Book 1</b>" {
		t.Fatalf("description = %q / %q", ov.Description, ov.DescriptionHTML)
	}
	check(t, s.decks.SetDescription(deck, "Line 1 > 0\nLine 2"))
	if ov := must(s.decks.Overview(deck)); ov.DescriptionHTML != "Line 1 &gt; 0<br>Line 2" {
		t.Fatalf("plain description = %q", ov.DescriptionHTML)
	}

	// Nothing due (app-shell spec): the overview says when the next card is
	// due.
	empty := must(s.decks.Create("Empty"))
	basicNoteIn(t, s, empty, "cuatro", "x")
	check(t, s.decks.SetOptions(empty, func() DeckOptions {
		o := DeckOptions(core.DefaultDeckOptions())
		o.NewPerDay = 0
		return o
	}()))
	ov = must(s.decks.Overview(empty))
	tomorrow := scheduler.NextDayStart(*s.clock, core.DefaultRolloverHour)
	if ov.Counts != (Counts{}) || ov.NextDue == nil || !ov.NextDue.Equal(tomorrow) {
		t.Fatalf("nothing due overview = %+v, want next due %v", ov, tomorrow)
	}
	if _, err := s.decks.Overview(12345); err == nil {
		t.Fatal("overview of a missing deck")
	}
}

// Set due date in the browser (bulk postpone) and while studying (push a
// card to next week), with the input remembered.
func TestSetDueDate(t *testing.T) {
	s, _ := opened(t)
	for i := 0; i < 20; i++ {
		basicNote(t, s, fmt.Sprint("q", i), "a")
	}
	ids := must(s.browse.Search("", "created", false))
	if len(ids) != 20 {
		t.Fatalf("cards = %d", len(ids))
	}
	if err := s.browse.SetDueDate(ids, "10-5"); err == nil || !core.IsUserError(err) {
		t.Fatalf("invalid input: err = %v", err)
	}
	check(t, s.browse.SetDueDate(ids, "3"))
	if n := len(must(s.browse.Search("prop:due=3", "created", false))); n != 20 {
		t.Fatalf("due in 3 days = %d", n)
	}
	if st := must(s.app.Settings()); st.LastDueDate != "3" {
		t.Fatalf("last due date = %q", st.LastDueDate)
	}
	// Two of them due today: setting the current one to next week shows
	// the other and lowers the review count.
	check(t, s.browse.SetDueDate(ids[:2], "0"))
	deck := must(s.notes.Defaults(0)).DeckID
	st := must(s.study.Start(deck))
	if st.Card == nil || st.Counts.Review != 2 {
		t.Fatalf("start = %+v", st)
	}
	first := st.Card.CardID
	st = must(s.study.SetDueDate(first, "7"))
	if st.Card == nil || st.Card.CardID == first || st.Counts.Review != 1 {
		t.Fatalf("after set due date = %+v", st)
	}
	if got := must(s.browse.Search("prop:due=7", "created", false)); len(got) != 1 || got[0] != first {
		t.Fatalf("due in 7 days = %v, want %d", got, first)
	}
	if set := must(s.app.Settings()); set.LastDueDate != "7" {
		t.Fatalf("last due date = %q", set.LastDueDate)
	}
	if _, err := s.study.SetDueDate(st.Card.CardID, "x"); err == nil {
		t.Fatal("invalid input accepted while studying")
	}
}

// Filtered decks through the service: create from options, study in
// preview mode, rebuild, empty, custom study and the tree's marking.
func TestFilteredDecksService(t *testing.T) {
	s, _ := opened(t)
	spanish := must(s.decks.Create("Spanish"))
	for i := range 3 {
		basicNoteIn(t, s, spanish, fmt.Sprint("q", i), "a")
	}
	opts := must(s.decks.NewFilteredOptions("deck:Spanish"))
	if opts.Name != "Filtered Deck 1" || len(opts.Terms) != 1 || opts.Terms[0].Limit != 100 ||
		!opts.Reschedule || fmt.Sprint(opts.PreviewDelays) != "[60 600 0]" || opts.SecondLimit != 20 {
		t.Fatalf("new options = %+v", opts)
	}
	opts.Reschedule = false
	built := must(s.decks.CreateFiltered(opts))
	if built.Cards != 3 {
		t.Fatalf("built = %+v", built)
	}
	if again := must(s.decks.NewFilteredOptions("")); again.Name != "Filtered Deck 2" {
		t.Fatalf("second name %q", again.Name)
	}
	if got := must(s.decks.FilteredOptions(built.DeckID)); got.Reschedule || got.Terms[0].Search != "deck:Spanish" {
		t.Fatalf("options = %+v", got)
	}
	var node DeckNode
	for _, n := range must(s.decks.Tree()) {
		if n.ID == built.DeckID {
			node = n
		}
	}
	if !node.Filtered || node.Counts.New != 3 || node.Cards != 3 {
		t.Fatalf("node = %+v", node)
	}

	// Preview: the rating buttons show the delays; Good returns the card.
	st := must(s.study.Start(built.DeckID))
	var labels []string
	for _, p := range st.Card.Previews {
		if p.Ends {
			labels = append(labels, "End")
		} else {
			labels = append(labels, core.FormatInterval(p.Interval))
		}
	}
	if strings.Join(labels, " ") != "<1m <10m End End" {
		t.Fatalf("labels %v", labels)
	}
	st = must(s.study.Answer(st.Card.CardID, 3, 1000))
	if st.Counts.New != 2 {
		t.Fatalf("counts %+v", st.Counts)
	}
	if b := must(s.decks.Rebuild(built.DeckID)); b.Cards != 3 {
		t.Fatalf("rebuild = %+v", b)
	}
	check(t, s.decks.Empty(built.DeckID))
	if info := must(s.decks.DeletionInfo(spanish)); info.Cards != 3 || info.InFiltered != 0 {
		t.Fatalf("info = %+v", info)
	}
	opts.Terms[0].Search = "tag:nothing"
	if _, err := s.decks.UpdateFiltered(built.DeckID, opts); err == nil || !strings.Contains(err.Error(), "No cards matched") {
		t.Fatalf("err = %v", err)
	}

	session := must(s.decks.CustomStudy(spanish, CustomStudyChoice{Kind: "byState", State: "new", Cards: 2}))
	if session.Cards != 2 {
		t.Fatalf("session = %+v", session)
	}
	if info := must(s.decks.DeletionInfo(spanish)); info.InFiltered != 2 {
		t.Fatalf("info = %+v", info)
	}

	// The editor never defaults to a filtered deck, though they sort first
	// once Default and the last used deck are gone.
	temp := must(s.decks.Create("Temp"))
	basicNoteIn(t, s, temp, "t", "t")
	check(t, s.decks.Delete(temp))
	for _, n := range must(s.decks.Tree()) {
		if n.FullName == "Default" {
			check(t, s.decks.Delete(n.ID))
		}
	}
	if d := must(s.notes.Defaults(0)); d.DeckID != spanish {
		t.Fatalf("editor deck %d", d.DeckID)
	}
}

func TestFSRSDeckOptions(t *testing.T) {
	s, store := opened(t)
	id := must(s.decks.Create("Spanish"))
	options := func(s *services) DeckOptions {
		for _, n := range must(s.decks.Tree()) {
			if n.ID == id {
				return n.Options
			}
		}
		t.Fatal("deck not found")
		return DeckOptions{}
	}
	if o := options(s); len(o.Weights) != 0 || o.HistoricalRetention != 0.9 || o.RescheduleOnChange {
		t.Fatalf("default options = %+v", o)
	}
	o := options(s)
	o.Weights = scheduler.DefaultWeights()
	o.Weights[0] = 0.5
	o.OptimizeSearch, o.IgnoreBefore, o.HistoricalRetention = "deck:Spanish", "2026-01-31", 0.8
	o.HealthCheck, o.RescheduleOnChange = true, true
	check(t, s.decks.SetOptions(id, o))
	again := newServices(t, store)
	must(again.app.Open("/data"))
	if got := options(again); got.Weights[0] != 0.5 || got.OptimizeSearch != "deck:Spanish" ||
		got.IgnoreBefore != "2026-01-31" || got.HistoricalRetention != 0.8 || !got.HealthCheck || !got.RescheduleOnChange {
		t.Fatalf("options after restart = %+v", got)
	}
	// No review cards, nothing to reschedule.
	settings := must(s.decks.Settings(id))
	o.DesiredRetention = 0.85
	if n := must(s.decks.RescheduleCount(id, settings.PresetID, o)); n != 0 {
		t.Errorf("reschedule count = %d", n)
	}
	if n := must(s.decks.RescheduleCountAll()); n != 0 {
		t.Errorf("reschedule count of all = %d", n)
	}
}

func TestOptimizeService(t *testing.T) {
	s, _ := opened(t)
	id := must(s.decks.Create("Spanish"))
	basicNoteIn(t, s, id, "uno", "one")
	o := must(s.decks.Settings(id)).Presets[0].Options
	// Too few reviews: the result says so, with the minimum.
	var last OptimizeEvent
	for ev := range s.decks.Optimize(context.Background(), id, o) {
		last = ev
	}
	if last.Stage != "done" || last.Result == nil || !last.Result.NotEnoughReviews ||
		last.Result.Reviews != 0 || last.Result.MinReviews != 400 {
		t.Fatalf("last event = %+v", last)
	}
	// Optimize all reports every preset and applies nothing.
	last = OptimizeEvent{}
	for ev := range s.decks.OptimizeAll(context.Background()) {
		last = ev
	}
	if last.Stage != "done" || len(last.Presets) != 1 || last.Presets[0].Applied || !last.Presets[0].Result.NotEnoughReviews {
		t.Fatalf("last event = %+v", last)
	}
	// A cancelled stream ends without a result.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for ev := range s.decks.Optimize(ctx, id, o) {
		if ev.Stage == "done" || ev.Error != "" {
			t.Fatalf("cancelled stream sent %+v", ev)
		}
	}
	// An invalid search is an error.
	o.OptimizeSearch = "deck:("
	last = OptimizeEvent{}
	for ev := range s.decks.Optimize(context.Background(), id, o) {
		last = ev
	}
	if !strings.Contains(last.Error, "search") {
		t.Fatalf("last event = %+v", last)
	}
}

// storeDump is everything stored, to check that nothing changed.
func storeDump(t *testing.T, store *storage.MemoryStore) map[string]string {
	t.Helper()
	b := store.Open()
	out := map[string]string{}
	for _, k := range must(b.List("")) {
		v, err := b.Get(k)
		if errors.Is(err, storage.ErrNotFound) {
			v = bytes.Join(must(b.Records(k)), []byte("\n"))
		} else {
			check(t, err)
		}
		out[k] = string(v)
	}
	return out
}

func collectSimulation(t *testing.T, events <-chan SimulationEvent) (stages []string, last SimulationEvent) {
	t.Helper()
	for ev := range events {
		stages = append(stages, ev.Stage)
		last = ev
	}
	return stages, last
}

// simulationDeck makes a deck with a subdeck, 60 notes and a few days of
// answers.
func simulationDeck(t *testing.T, s *services) (int64, DeckOptions) {
	t.Helper()
	id := must(s.decks.Create("Spanish"))
	sub := must(s.decks.Create("Spanish::Verbs"))
	for i := range 60 {
		deck := id
		if i%3 == 0 {
			deck = sub
		}
		basicNoteIn(t, s, deck, fmt.Sprint("palabra ", i), fmt.Sprint("word ", i))
	}
	for range 3 {
		st := must(s.study.Start(id))
		for n := 0; st.Card != nil && n < 200; n++ {
			rating := 3
			if n%5 == 0 {
				rating = 1
			}
			st = must(s.study.Answer(st.Card.CardID, rating, 4000))
		}
		*s.clock = s.clock.AddDate(0, 0, 1)
	}
	return id, must(s.decks.Settings(id)).Presets[0].Options
}

// Compare retentions (spec): both curves, more reviews at the higher
// retention; simulating changes nothing and is reproducible.
func TestSimulateService(t *testing.T) {
	s, store := opened(t)
	id, o := simulationDeck(t, s)
	before, undo := storeDump(t, store), must(s.app.UndoInfo())
	in := SimulationInput{Days: 365, NewPerDay: 20, ReviewsPerDay: 9999, Retention: 0.90, CompareRetention: 0.95}
	stages, last := collectSimulation(t, s.decks.Simulate(context.Background(), id, o, in))
	if last.Stage != "done" || last.Error != "" || len(last.Series) != 2 {
		t.Fatalf("last event = %+v", last)
	}
	if !slices.Contains(stages, "reading") {
		t.Errorf("stages = %v", stages)
	}
	a, b := last.Series[0], last.Series[1]
	if a.Retention != 0.90 || b.Retention != 0.95 || len(a.Reviews) != 365 || len(b.Memorized) != 365 {
		t.Fatalf("series = %v %v, %d days", a.Retention, b.Retention, len(a.Reviews))
	}
	if b.TotalReviews <= a.TotalReviews || a.TotalSeconds <= 0 {
		t.Fatalf("reviews %d at 0.90, %d at 0.95; %v s", a.TotalReviews, b.TotalReviews, a.TotalSeconds)
	}
	// 40 new cards were never answered, 20 were.
	if a.TotalNew > 40 || a.Memorized[0] < 10 {
		t.Fatalf("new %d, memorised on day 0 %.1f", a.TotalNew, a.Memorized[0])
	}
	_, again := collectSimulation(t, s.decks.Simulate(context.Background(), id, o, in))
	if !reflect.DeepEqual(last, again) {
		t.Fatal("same inputs, different results")
	}
	// Without a second retention there is one series.
	in.CompareRetention = 0
	if _, one := collectSimulation(t, s.decks.Simulate(context.Background(), id, o, in)); len(one.Series) != 1 {
		t.Fatalf("series = %d", len(one.Series))
	}
	if !reflect.DeepEqual(before, storeDump(t, store)) || must(s.app.UndoInfo()) != undo {
		t.Fatal("simulating changed the collection")
	}

	// Invalid inputs and decks are errors; a cancelled stream ends silently.
	for _, bad := range []SimulationInput{{Days: 0, Retention: 0.9}, {Days: 10, Retention: 0.5},
		{Days: 10, Retention: 0.9, CompareRetention: 1.2}, {Days: 10, NewPerDay: -1, Retention: 0.9}} {
		if _, ev := collectSimulation(t, s.decks.Simulate(context.Background(), id, o, bad)); ev.Error == "" {
			t.Errorf("%+v: %+v", bad, ev)
		}
	}
	if _, ev := collectSimulation(t, s.decks.Simulate(context.Background(), 12345, o, in)); !strings.Contains(ev.Error, "no longer exists") {
		t.Errorf("missing deck: %+v", ev)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for ev := range s.decks.Simulate(ctx, id, o, in) {
		if ev.Stage == "done" || ev.Error != "" {
			t.Fatalf("cancelled stream sent %+v", ev)
		}
	}
}

// Help me decide (spec): a suggestion and its cost curve; the deck's
// retention is unchanged.
func TestSuggestRetentionService(t *testing.T) {
	s, store := opened(t)
	id, o := simulationDeck(t, s)
	before := storeDump(t, store)
	in := SimulationInput{Days: 365, NewPerDay: 20, ReviewsPerDay: 9999}
	stages, last := collectSimulation(t, s.decks.SuggestRetention(context.Background(), id, o, in))
	if last.Stage != "done" || last.Suggestion == nil {
		t.Fatalf("last event = %+v", last)
	}
	if !slices.Contains(stages, "simulating") {
		t.Errorf("stages = %v", stages)
	}
	sug := last.Suggestion
	if sug.Retention < 0.70 || sug.Retention > 0.99 || len(sug.Points) != 30 {
		t.Fatalf("suggestion = %+v", sug)
	}
	for _, p := range sug.Points {
		if p.SecondsPerCard <= 0 {
			t.Fatalf("point %+v", p)
		}
	}
	if !reflect.DeepEqual(before, storeDump(t, store)) {
		t.Fatal("suggesting changed the collection")
	}
	if got := must(s.decks.Settings(id)).Presets[0].Options.DesiredRetention; got != o.DesiredRetention {
		t.Fatalf("retention = %v", got)
	}
	// An empty deck has nothing to suggest from.
	empty := must(s.decks.Create("Empty"))
	if _, ev := collectSimulation(t, s.decks.SuggestRetention(context.Background(), empty, o, in)); !strings.Contains(ev.Error, "no cards") {
		t.Fatalf("empty deck: %+v", ev)
	}
}

// Card type options pass through the service: the deck override, browser
// templates and the card type order.
func TestCardTypeOptionsService(t *testing.T) {
	s, _ := opened(t)
	typeID := must(s.notes.AddType("Basic (and reversed card)", "Vocab"))
	production := must(s.decks.Create("Spanish::Production"))
	tm := must(s.notes.Type(typeID)).Templates[1]
	tm.DeckOverride, tm.BrowserQFmt = production, "{{Back}}"
	check(t, s.notes.UpdateTemplate(typeID, 1, tm))
	if got := must(s.notes.Type(typeID)).Templates[1]; got != tm {
		t.Fatalf("template = %+v, want %+v", got, tm)
	}
	check(t, s.notes.MoveTemplate(typeID, 1, 0))
	if got := must(s.notes.Type(typeID)).Templates; got[0] != tm || got[1].Name != "Card 1" {
		t.Fatalf("templates after move = %+v", got)
	}
}

// restarted is the app started again on the data s left behind.
func restarted(t *testing.T, s *services) *services {
	t.Helper()
	again := newServices(t, s.stores["/data"])
	again.stores, again.registry = s.stores, s.registry
	return again
}

func deckNames(t *testing.T, s *services) []string {
	t.Helper()
	var names []string
	for _, d := range must(s.decks.Tree()) {
		names = append(names, d.Name)
	}
	return names
}

func profileNames(t *testing.T, s *services) []string {
	t.Helper()
	var names []string
	for _, p := range must(s.app.Profiles()).Profiles {
		names = append(names, p.Name)
	}
	return names
}

func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || usermsg.Code(err.Error()) != code {
		t.Fatalf("error = %v, want code %s", err, code)
	}
}

func TestProfilesUpgrade(t *testing.T) {
	// A collection from before profiles: no registry.
	s, store := opened(t)
	must(s.decks.Create("Spanish"))
	basicNote(t, s, "hola", "hello")

	again := newServices(t, store)
	list := must(again.app.LoadProfiles("/data"))
	want := []Profile{{ID: storage.DefaultProfile, Name: "Default"}}
	if !reflect.DeepEqual(list.Profiles, want) || list.Current != "" || list.ShowAtStartup {
		t.Fatalf("profiles before open = %+v", list)
	}
	if st := must(again.app.Open("/data")); st.State != StatusOK {
		t.Fatalf("open = %+v", st)
	}
	if names := deckNames(t, again); !slices.Contains(names, "Spanish") {
		t.Fatalf("decks = %v", names)
	}
	if n := must(again.app.ProfileCardCount(storage.DefaultProfile)); n != 1 {
		t.Fatalf("cards = %d", n)
	}
	if list := must(again.app.Profiles()); list.Current != storage.DefaultProfile {
		t.Fatalf("current = %q", list.Current)
	}
	// Opening writes nothing: the data stays where it was.
	if again.registry != nil || len(again.stores) != 1 {
		t.Fatalf("registry %s, stores %v", again.registry, again.stores)
	}
}

func TestProfilesSeparate(t *testing.T) {
	s, _ := opened(t)
	check(t, s.app.RenameProfile(storage.DefaultProfile, "Anna"))
	must(s.decks.Create("Spanish"))
	spanish := basicNote(t, s, "hola", "hello")
	check(t, s.app.SetTheme("dark"))
	must(s.app.CreateBackup())
	must(s.study.Start(must(s.decks.Tree())[0].ID))
	staged := must(s.transfer.StageBytes("deck.txt", []byte("a\tb\n")))
	changes := s.app.Changes(t.Context())

	ben := must(s.app.AddProfile("Ben"))
	st := must(s.app.SwitchProfile(ben.ID))
	if st.State != StatusFresh {
		t.Fatalf("switch = %+v", st)
	}
	if names := deckNames(t, s); !reflect.DeepEqual(names, []string{"Default"}) {
		t.Fatalf("Ben's decks = %v", names)
	}
	if _, err := s.notes.Get(spanish); err == nil {
		t.Fatal("Anna's note visible in Ben")
	}
	if st := must(s.app.Settings()); st.Theme != "system" {
		t.Fatalf("Ben's theme = %q", st.Theme)
	}
	if list := must(s.app.Backups()); len(list.Backups) != 0 {
		t.Fatalf("Ben's backups = %+v", list)
	}
	// The study session and staged files belonged to Anna.
	if _, err := s.study.Current(); !errors.Is(err, errNoSession) {
		t.Fatalf("study after switch: %v", err)
	}
	if _, err := s.transfer.Inspect(staged.ID); err == nil {
		t.Fatal("staged file kept after switch")
	}
	// Open screens learn that everything changed.
	for ch := range changes {
		if ch.Kind == "replaced" && ch.All {
			break
		}
	}

	basicNote(t, s, "ben", "card")
	must(s.app.SwitchProfile(storage.DefaultProfile))
	if names := deckNames(t, s); !slices.Contains(names, "Spanish") {
		t.Fatalf("Anna's decks = %v", names)
	}
	if n := must(s.app.ProfileCardCount(ben.ID)); n != 1 {
		t.Fatalf("Ben's cards = %d", n)
	}
	if n := must(s.app.ProfileCardCount(storage.DefaultProfile)); n != 1 {
		t.Fatalf("Anna's cards = %d", n)
	}
	if st := must(s.app.Settings()); st.Theme != "dark" {
		t.Fatalf("Anna's theme = %q", st.Theme)
	}
	if list := must(s.app.Backups()); len(list.Backups) != 1 {
		t.Fatalf("Anna's backups = %+v", list)
	}
}

func TestProfilesManage(t *testing.T) {
	s, _ := opened(t)
	ben := must(s.app.AddProfile("  Ben "))
	if ben.Name != "Ben" || ben.ID == storage.DefaultProfile {
		t.Fatalf("added %+v", ben)
	}
	_, err := s.app.AddProfile("ben")
	wantCode(t, err, "profileNameTaken")
	_, err = s.app.AddProfile(" ")
	wantCode(t, err, "profileNameEmpty")
	wantCode(t, s.app.RenameProfile(ben.ID, "default"), "profileNameTaken")
	check(t, s.app.RenameProfile(ben.ID, "BEN"))
	check(t, s.app.RenameProfile(storage.DefaultProfile, "anna"))
	if names := profileNames(t, s); !reflect.DeepEqual(names, []string{"anna", "BEN"}) {
		t.Fatalf("profiles = %v", names)
	}

	// A profile that was never opened has no cards and no data.
	if n := must(s.app.ProfileCardCount(ben.ID)); n != 0 {
		t.Fatalf("cards = %d", n)
	}

	// The last used profile opens on start.
	must(s.app.SwitchProfile(ben.ID))
	for range 3 {
		basicNote(t, s, "ben", "card")
	}
	check(t, s.app.SetShowProfilesAtStartup(true))
	again := restarted(t, s)
	list := must(again.app.LoadProfiles("/data"))
	if !list.ShowAtStartup || list.Current != "" {
		t.Fatalf("profiles on start = %+v", list)
	}
	must(again.app.Open("/data"))
	if n := len(must(again.browse.Search("", "sortField", false))); n != 3 {
		t.Fatalf("Ben opened with %d cards", n)
	}

	// The open profile cannot be deleted; another one can, with its data.
	wantCode(t, again.app.DeleteProfile(ben.ID), "profileOpen")
	must(again.app.SwitchProfile(storage.DefaultProfile))
	if n := must(again.app.ProfileCardCount(ben.ID)); n != 3 {
		t.Fatalf("cards = %d", n)
	}
	check(t, again.app.DeleteProfile(ben.ID))
	if again.profileStore(ben.ID) != nil {
		t.Fatal("Ben's data kept")
	}
	if names := profileNames(t, again); !reflect.DeepEqual(names, []string{"anna"}) {
		t.Fatalf("profiles = %v", names)
	}
	_, err = again.app.SwitchProfile(ben.ID)
	wantCode(t, err, "profileGone")
	// A new profile never gets a deleted one's ID.
	if p := must(again.app.AddProfile("Ben")); p.ID == ben.ID {
		t.Fatalf("ID %s reused", p.ID)
	}

	// On the start list, before any profile is open, the only one stays.
	only := restarted(t, again)
	must(only.app.LoadProfiles("/data"))
	for _, p := range must(only.app.Profiles()).Profiles {
		if p.Name == "Ben" {
			check(t, only.app.DeleteProfile(p.ID))
		}
	}
	wantCode(t, only.app.DeleteProfile(storage.DefaultProfile), "profileOnly")
	if st := must(only.app.Open("/data")); st.State != StatusOK {
		t.Fatalf("open = %+v", st)
	}
}

func TestProfilesBackups(t *testing.T) {
	s, _ := opened(t)
	must(s.decks.Create("Spanish"))
	file := must(s.transfer.ExportBackup())
	ben := must(s.app.AddProfile("Ben"))
	must(s.app.SwitchProfile(ben.ID))
	must(s.decks.Create("German"))

	// Restoring replaces only the open profile's collection.
	info := must(s.transfer.StageBytes(file.Name, file.Data))
	check(t, s.transfer.Restore(info.ID))
	if names := deckNames(t, s); !slices.Contains(names, "Spanish") || slices.Contains(names, "German") {
		t.Fatalf("Ben after restore = %v", names)
	}
	// The backup made before the restore is Ben's.
	if list := must(s.app.Backups()); len(list.Backups) != 1 {
		t.Fatalf("Ben's backups = %+v", list)
	}
	must(s.app.SwitchProfile(storage.DefaultProfile))
	if names := deckNames(t, s); !reflect.DeepEqual(names, []string{"Default", "Spanish"}) {
		t.Fatalf("Default after Ben's restore = %v", names)
	}
	if list := must(s.app.Backups()); len(list.Backups) != 0 {
		t.Fatalf("Default's backups = %+v", list)
	}
}

// freePort returns a loopback port nothing listens on.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// mcpPost sends an MCP initialize request to the server at url and returns
// the HTTP status; 0 means the connection failed.
func mcpPost(t *testing.T, url, token, origin string) int {
	t.Helper()
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// mcpOn turns the MCP server on at a free port.
func mcpOn(t *testing.T, s *services) McpStatus {
	t.Helper()
	must(s.mcp.SetPort(freePort(t)))
	st := must(s.mcp.SetEnabled(true))
	if !st.Running || st.Error != "" {
		t.Fatalf("status = %+v", st)
	}
	return st
}

func TestMcpOffByDefault(t *testing.T) {
	s, _ := opened(t)
	if !s.mcp.Supported() {
		t.Fatal("not supported on " + runtime.GOOS)
	}
	st := s.mcp.Status()
	if st.Enabled || st.Running || st.Token != "" || st.Port != DefaultMcpPort || st.URL != "http://127.0.0.1:47811/mcp" {
		t.Fatalf("status = %+v", st)
	}
	if s.mcpFile != nil {
		t.Fatalf("settings written: %s", s.mcpFile)
	}
}

func TestMcpEnableDisable(t *testing.T) {
	s, _ := opened(t)
	st := mcpOn(t, s)
	if st.URL != fmt.Sprintf("http://127.0.0.1:%d/mcp", st.Port) || len(st.Token) < 22 {
		t.Fatalf("status = %+v", st)
	}
	if code := mcpPost(t, st.URL, st.Token, ""); code != http.StatusOK {
		t.Fatalf("initialize = %d", code)
	}
	st = must(s.mcp.SetEnabled(false))
	if st.Enabled || st.Running {
		t.Fatalf("status = %+v", st)
	}
	if code := mcpPost(t, st.URL, st.Token, ""); code != 0 {
		t.Fatalf("after turning off = %d, want no connection", code)
	}
}

func TestMcpTokenKeptAndRegenerated(t *testing.T) {
	s, store := opened(t)
	first := mcpOn(t, s)
	must(s.mcp.SetEnabled(false))
	again := must(s.mcp.SetEnabled(true))
	if again.Token != first.Token {
		t.Fatal("turning the server on again changed the token")
	}
	// Next start (agent-access spec): the server runs again with the same
	// URL and token.
	file := s.mcpFile
	s.mcp.mu.Lock()
	s.mcp.stopLocked()
	s.mcp.mu.Unlock()
	s2 := newServices(t, store)
	s2.mcpFile = file
	must(s2.app.Open("/data"))
	if st := s2.mcp.Status(); !st.Running || st.Token != first.Token || st.URL != first.URL {
		t.Fatalf("after restart = %+v, want running at %s", st, first.URL)
	}
	// Regenerating makes the old token invalid at once.
	st := must(s2.mcp.RegenerateToken())
	if st.Token == first.Token {
		t.Fatal("token not regenerated")
	}
	if code := mcpPost(t, st.URL, first.Token, ""); code != http.StatusUnauthorized {
		t.Fatalf("old token = %d", code)
	}
	if code := mcpPost(t, st.URL, st.Token, ""); code != http.StatusOK {
		t.Fatalf("new token = %d", code)
	}
}

func TestMcpRequestChecks(t *testing.T) {
	s, _ := opened(t)
	st := mcpOn(t, s)
	if code := mcpPost(t, st.URL, "", ""); code != http.StatusUnauthorized {
		t.Fatalf("missing token = %d", code)
	}
	if code := mcpPost(t, st.URL, "wrong", ""); code != http.StatusUnauthorized {
		t.Fatalf("wrong token = %d", code)
	}
	if code := mcpPost(t, st.URL, st.Token, "https://example.com"); code != http.StatusForbidden {
		t.Fatalf("web page = %d", code)
	}
	for _, origin := range []string{"http://localhost:3000", "http://127.0.0.1", "https://[::1]:8080"} {
		if code := mcpPost(t, st.URL, st.Token, origin); code != http.StatusOK {
			t.Fatalf("origin %s = %d", origin, code)
		}
	}
	// Only the loopback interface listens.
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			ip, _, _ := net.ParseCIDR(a.String())
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			if conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, st.Port), time.Second); err == nil {
				conn.Close()
				t.Fatalf("server reachable at %s", ip)
			}
		}
	}
}

func TestMcpPortInUseAndPortChange(t *testing.T) {
	s, _ := opened(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	busy := ln.Addr().(*net.TCPAddr).Port
	must(s.mcp.SetPort(busy))
	st := must(s.mcp.SetEnabled(true))
	if !st.Enabled || st.Running || usermsg.Plain(st.Error) != fmt.Sprintf("Port %d is already in use.", busy) {
		t.Fatalf("status = %+v", st)
	}
	// Another port starts the server; the shown URL follows.
	port := freePort(t)
	st = must(s.mcp.SetPort(port))
	if !st.Running || st.Error != "" || st.URL != fmt.Sprintf("http://127.0.0.1:%d/mcp", port) {
		t.Fatalf("status = %+v", st)
	}
	// Moving again leaves the old port.
	old := st
	st = must(s.mcp.SetPort(freePort(t)))
	if code := mcpPost(t, old.URL, st.Token, ""); code != 0 {
		t.Fatalf("old port = %d", code)
	}
	if code := mcpPost(t, st.URL, st.Token, ""); code != http.StatusOK {
		t.Fatalf("new port = %d", code)
	}
	if _, err := s.mcp.SetPort(80); usermsg.Plain(fmt.Sprint(err)) != "The port must be between 1024 and 65535." {
		t.Fatalf("port 80: %v", err)
	}
	if s.mcp.Status().Port != st.Port {
		t.Fatal("invalid port saved")
	}
}

func TestMcpAcrossProfiles(t *testing.T) {
	s, _ := opened(t)
	st := mcpOn(t, s)
	p := must(s.app.AddProfile("Ben"))
	must(s.app.SwitchProfile(p.ID))
	if got := s.mcp.Status(); !got.Running || got.Token != st.Token {
		t.Fatalf("after switch = %+v", got)
	}
	if code := mcpPost(t, st.URL, st.Token, ""); code != http.StatusOK {
		t.Fatalf("after switch = %d", code)
	}
}

func TestMcpStartsWithProfileList(t *testing.T) {
	// The profile list at startup: the server runs although no collection
	// is open.
	store := storage.NewMemoryStore()
	s := newServices(t, store)
	port := freePort(t)
	s.mcpFile = []byte(fmt.Sprintf(`{"enabled":true,"port":%d,"token":"secret"}`, port))
	must(s.app.LoadProfiles("/data"))
	if st := s.mcp.Status(); !st.Running {
		t.Fatalf("status = %+v", st)
	}
	// A hot restart does not start it again.
	must(s.app.LoadProfiles("/data"))
	must(s.app.Open("/data"))
	if st := s.mcp.Status(); !st.Running || st.Error != "" {
		t.Fatalf("status = %+v", st)
	}
}

func TestMcpStatusChanges(t *testing.T) {
	s, _ := opened(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := s.mcp.StatusChanges(ctx)
	if st := <-ch; st.Enabled {
		t.Fatalf("first = %+v", st)
	}
	mcpOn(t, s)
	for st := range ch {
		if st.Running {
			break
		}
	}
	s.mcp.recordActivity("x")
	if st := <-ch; !st.Running {
		t.Fatalf("after activity = %+v", st)
	}
	if a := s.mcp.Activity(); len(a) != 1 || a[0].Step != "x" {
		t.Fatalf("activity = %+v", a)
	}
}
