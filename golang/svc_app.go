package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"cardflip/golang/backup"
	"cardflip/golang/core"
	"cardflip/golang/render"
	"cardflip/golang/storage"
	"cardflip/golang/storage/storagetest"
)

// AppService opens the collection and holds app-wide settings.
type AppService struct{ a *app }

// Open opens the collection in dataDir (empty on the web), creating it on
// first launch. It must succeed before any other call. Calling it again
// (e.g. after a hot restart) returns the current status.
func (s *AppService) Open(dataDir string) (Status, error) { return s.a.open(dataDir) }

// Status returns the current status.
func (s *AppService) Status() Status {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	return s.a.statusLocked()
}

// StartFresh deletes the unreadable collection and starts a new one. The
// app only calls it after the user explicitly confirmed.
func (s *AppService) StartFresh() (Status, error) { return s.a.startFresh() }

// Reload drops the collection from memory and reads it again from storage,
// as a restart of the app would. The integration tests use it to check that
// everything was persisted.
func (s *AppService) Reload() (Status, error) { return s.a.reload() }

// Settings are the app settings the UI shows.
type Settings struct {
	// Theme is "system", "light" or "dark".
	Theme string `json:"theme"`
	// Language is the interface language as a language code, such as
	// "de"; empty follows the device's languages.
	Language     string `json:"language"`
	RolloverHour int    `json:"rolloverHour"`
	// AutoContinueLearning shows learning cards in a waiting study session
	// as soon as they are due, without asking.
	AutoContinueLearning bool `json:"autoContinueLearning"`
	// LearnAheadMinutes is how far ahead of their due time a study session
	// shows learning cards when nothing else is due (0 to 1440; 0 shows
	// them only once due).
	LearnAheadMinutes int `json:"learnAheadMinutes"`
	// ShowPlayButtons shows replay buttons on cards with audio.
	ShowPlayButtons bool `json:"showPlayButtons"`
	// InterruptAudio stops the question's sounds when the answer is shown.
	InterruptAudio bool `json:"interruptAudio"`
	// FlagNames are the custom names of flags 1 to 7 (index 0 is flag 1);
	// an empty name means the flag's colour name.
	FlagNames []string `json:"flagNames"`
	// StripPasteFormatting makes a normal paste insert plain text and the
	// plain-paste shortcut keep formatting.
	StripPasteFormatting bool `json:"stripPasteFormatting"`
	// PasteImagesAsPng stores pasted images as PNG (lossless) instead of
	// JPEG (smaller).
	PasteImagesAsPng bool `json:"pasteImagesAsPng"`
	// Backup configures automatic backups.
	Backup core.BackupSettings `json:"backup"`
	// DeckByNoteType makes the add screen switch to the deck last used with
	// the chosen note type ("Deck for new cards").
	DeckByNoteType bool `json:"deckByNoteType"`
	// IgnoreAccents makes text and field searches ignore accents ("Ignore
	// accents in search").
	IgnoreAccents bool `json:"ignoreAccents"`
	// BrowserColumns are the browser's columns in order (see
	// BrowseService.Rows).
	BrowserColumns []string `json:"browserColumns"`
	// BrowserMode is "cards" (one row per card) or "notes" (one row per
	// note).
	BrowserMode string `json:"browserMode"`
	// BrowserSplit is the share of the width that the list takes next to
	// the details on wide screens (0 to 1, exclusive).
	BrowserSplit float64 `json:"browserSplit"`
	// DefaultSearch is the search the browser starts with ("Default
	// browser search"); empty lists all cards.
	DefaultSearch string `json:"defaultSearch"`
	// LastDueDate is the last input of Set due date, which its dialog
	// starts with.
	LastDueDate string `json:"lastDueDate"`
	// TimeboxMinutes is the timebox time limit in minutes (0 to 1440; 0
	// turns it off).
	TimeboxMinutes int `json:"timeboxMinutes"`
	// ShowRemainingCounts shows the remaining new, learning and review
	// counts on the study screen.
	ShowRemainingCounts bool `json:"showRemainingCounts"`
	// ShowIntervals shows the next review time on the answer buttons.
	ShowIntervals bool `json:"showIntervals"`
	// SpaceAnswersGood makes Space and Enter choose Good once the answer is
	// shown; when off, they only reveal the answer.
	SpaceAnswersGood bool `json:"spaceAnswersGood"`
	// HideBarsDuringReview hides the app bar and navigation while studying.
	HideBarsDuringReview bool `json:"hideBarsDuringReview"`
	// QuestionFadeMs is how long each new question fades in, in
	// milliseconds (0 to 1000; 0 shows it at once).
	QuestionFadeMs int `json:"questionFadeMs"`
	// RecentDecks are the IDs of the decks most recently studied, latest
	// first (at most 10), which the deck picker lists first.
	RecentDecks []int64 `json:"recentDecks"`
	// UIScale is the size of the interface in percent (75 to 200 in steps
	// of 5), on top of the operating system's text scale.
	UIScale int `json:"uiScale"`
	// Minimalist shows the interface without shadows, elevation tints and
	// decorative fills.
	Minimalist bool `json:"minimalist"`
	// ReduceMotion is "system" (follow the operating system), "on" or
	// "off": whether transitions and animations are replaced by immediate
	// changes.
	ReduceMotion string `json:"reduceMotion"`
	// ShowProgress shows the streak and daily goal on the deck list and the
	// study summary ("Show streaks and goals").
	ShowProgress bool `json:"showProgress"`
	// DailyGoalKind is "cards" or "minutes", or empty for no daily goal;
	// DailyGoal is the number of cards or minutes.
	DailyGoalKind string `json:"dailyGoalKind"`
	DailyGoal     int    `json:"dailyGoal"`
	// RestDaysPerWeek is how many days without study, 0 to 2, a streak
	// survives in any 7 days in a row.
	RestDaysPerWeek int `json:"restDaysPerWeek"`
	// GoalSuggestionShown records that the user opened or dismissed the
	// one-time suggestion to set a daily goal.
	GoalSuggestionShown bool `json:"goalSuggestionShown"`
	// DeckProgressInList shows a thin progress bar under each deck of the
	// deck list ("Show deck progress in the deck list").
	DeckProgressInList bool `json:"deckProgressInList"`
	// RecapDismissedWeek is the start of the week whose recap was
	// dismissed (Recap.WeekStart), or 0.
	RecapDismissedWeek int `json:"recapDismissedWeek"`
}

// Settings returns the app settings.
func (s *AppService) Settings() (Settings, error) {
	c, err := s.a.collection()
	if err != nil {
		return Settings{}, err
	}
	st := c.Settings()
	return Settings{Theme: st.Theme, Language: st.Language, RolloverHour: st.RolloverHour, AutoContinueLearning: st.AutoContinueLearning,
		ShowPlayButtons: !st.HidePlayButtons, InterruptAudio: !st.KeepAudioOnAnswer, FlagNames: st.FlagNames[:],
		StripPasteFormatting: st.StripPasteFormatting, PasteImagesAsPng: st.PasteImagesAsPng, Backup: st.Backup,
		DeckByNoteType: st.DeckByNoteType, IgnoreAccents: st.IgnoreAccents, BrowserColumns: browserColumns(st),
		BrowserMode: cmp.Or(st.BrowserMode, core.BrowserCards), BrowserSplit: cmp.Or(st.BrowserSplit, core.DefaultBrowserSplit),
		DefaultSearch: st.DefaultSearch, LastDueDate: st.LastDueDate, LearnAheadMinutes: int(st.LearnAhead() / time.Minute),
		TimeboxMinutes: st.TimeboxMinutes, ShowRemainingCounts: !st.HideRemainingCounts, ShowIntervals: !st.HideIntervals,
		SpaceAnswersGood: !st.SpaceOnlyReveals, HideBarsDuringReview: st.HideBarsDuringReview,
		QuestionFadeMs: int(st.QuestionFade() / time.Millisecond), RecentDecks: fromIDs(c.RecentDecks()),
		UIScale: cmp.Or(st.UIScale, core.DefaultUIScale), Minimalist: st.Minimalist,
		ReduceMotion: cmp.Or(st.ReduceMotion, core.ReduceMotionSystem), ShowProgress: !st.HideProgress,
		DailyGoalKind: st.DailyGoalKind, DailyGoal: st.DailyGoal, RestDaysPerWeek: st.RestDaysPerWeek,
		GoalSuggestionShown: st.GoalSuggestionShown, DeckProgressInList: st.DeckProgressInList,
		RecapDismissedWeek: st.RecapDismissedWeek}, nil
}

func browserColumns(st core.Settings) []string {
	cols := st.BrowserColumns
	if len(cols) == 0 {
		cols = core.DefaultColumns
	}
	out := make([]string, len(cols))
	for i, col := range cols {
		out[i] = string(col)
	}
	return out
}

// SetTheme stores the theme mode: "system", "light" or "dark".
func (s *AppService) SetTheme(mode string) error {
	return s.updateSettings(func(st *core.Settings) { st.Theme = mode })
}

// SetLanguage stores the interface language as a language code, such as
// "de"; empty follows the device's languages.
func (s *AppService) SetLanguage(code string) error {
	return s.updateSettings(func(st *core.Settings) { st.Language = code })
}

// SetRolloverHour sets the hour (0–23) at which a new study day starts.
func (s *AppService) SetRolloverHour(hour int) error {
	return s.updateSettings(func(st *core.Settings) { st.RolloverHour = hour })
}

// SetAutoContinueLearning sets whether a waiting study session shows
// learning cards as soon as they are due, without asking.
func (s *AppService) SetAutoContinueLearning(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.AutoContinueLearning = on })
}

// SetLearnAheadMinutes sets how far ahead of their due time, in minutes
// (0 to 1440), a study session shows learning cards when nothing else is
// due; 0 shows them only once due.
func (s *AppService) SetLearnAheadMinutes(minutes int) error {
	if minutes < 0 || minutes > core.MaxLearnAheadMinutes {
		return core.NewError("learnAheadRange", "the learn ahead limit must be between 0 and %d minutes", core.MaxLearnAheadMinutes)
	}
	return s.updateSettings(func(st *core.Settings) { st.LearnAheadMinutes = &minutes })
}

// SetTimeboxMinutes sets the timebox time limit in minutes (0 to 1440):
// each time it passes in a study session, the study screen reports the
// cards studied in it. 0 turns it off.
func (s *AppService) SetTimeboxMinutes(minutes int) error {
	return s.updateSettings(func(st *core.Settings) { st.TimeboxMinutes = minutes })
}

// SetShowRemainingCounts sets whether the study screen shows the remaining
// new, learning and review counts.
func (s *AppService) SetShowRemainingCounts(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.HideRemainingCounts = !on })
}

// SetShowIntervals sets whether the answer buttons show the next review
// time.
func (s *AppService) SetShowIntervals(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.HideIntervals = !on })
}

// SetSpaceAnswersGood sets whether Space and Enter choose Good once the
// answer is shown; when off, they only reveal the answer.
func (s *AppService) SetSpaceAnswersGood(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.SpaceOnlyReveals = !on })
}

// SetHideBarsDuringReview sets whether the app bar and navigation are
// hidden while studying.
func (s *AppService) SetHideBarsDuringReview(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.HideBarsDuringReview = on })
}

// SetQuestionFadeMs sets how long each new question fades in, in
// milliseconds (0 to 1000); 0 shows it at once.
func (s *AppService) SetQuestionFadeMs(ms int) error {
	return s.updateSettings(func(st *core.Settings) { st.QuestionFadeMs = &ms })
}

// SetUIScale sets the size of the interface in percent: 75 to 200 in
// steps of 5.
func (s *AppService) SetUIScale(percent int) error {
	if err := core.CheckUIScale(percent); err != nil {
		return err
	}
	return s.updateSettings(func(st *core.Settings) { st.UIScale = percent })
}

// SetMinimalist sets whether the interface is shown without shadows,
// elevation tints and decorative fills.
func (s *AppService) SetMinimalist(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.Minimalist = on })
}

// SetReduceMotion sets whether transitions and animations are replaced by
// immediate changes: "system" follows the operating system, "on" or "off".
func (s *AppService) SetReduceMotion(mode string) error {
	if mode == core.ReduceMotionSystem {
		mode = ""
	}
	return s.updateSettings(func(st *core.Settings) { st.ReduceMotion = mode })
}

// SetShowProgress sets whether the deck list and the study summary show
// the streak and daily goal. It belongs to the device.
func (s *AppService) SetShowProgress(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.HideProgress = !on })
}

// SetDailyGoal sets the daily goal: kind "cards" or "minutes" with a value
// from 1 to 9,999, or kind "" for no goal (the value is then ignored).
func (s *AppService) SetDailyGoal(kind string, value int) error {
	if kind == "" {
		value = 0
	}
	return s.updateSettings(func(st *core.Settings) { st.DailyGoalKind, st.DailyGoal = kind, value })
}

// SetRestDaysPerWeek sets how many days without study, 0 to 2, a streak
// survives in any 7 days in a row. It belongs to the collection.
func (s *AppService) SetRestDaysPerWeek(n int) error {
	return s.updateSettings(func(st *core.Settings) { st.RestDaysPerWeek = n })
}

// GoalSuggestionShown records that the user opened or dismissed the
// one-time suggestion to set a daily goal.
func (s *AppService) GoalSuggestionShown() error {
	return s.updateSettings(func(st *core.Settings) { st.GoalSuggestionShown = true })
}

// SetDeckProgressInList sets whether the deck list shows a thin progress
// bar under each deck. It belongs to the device.
func (s *AppService) SetDeckProgressInList(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.DeckProgressInList = on })
}

// DismissRecap hides the weekly recap of the week starting on study day
// weekStart (Recap.WeekStart) on this device.
func (s *AppService) DismissRecap(weekStart int) error {
	return s.updateSettings(func(st *core.Settings) { st.RecapDismissedWeek = weekStart })
}

// SetShowPlayButtons sets whether cards with audio show replay buttons.
func (s *AppService) SetShowPlayButtons(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.HidePlayButtons = !on })
}

// SetInterruptAudio sets whether showing the answer stops the question's
// sounds.
func (s *AppService) SetInterruptAudio(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.KeepAudioOnAnswer = !on })
}

// SetStripPasteFormatting sets whether a normal paste into a field inserts
// plain text.
func (s *AppService) SetStripPasteFormatting(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.StripPasteFormatting = on })
}

// SetPasteImagesAsPng sets whether pasted images are stored as PNG
// instead of JPEG.
func (s *AppService) SetPasteImagesAsPng(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.PasteImagesAsPng = on })
}

// SetIgnoreAccents sets whether text and field searches ignore accents.
func (s *AppService) SetIgnoreAccents(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.IgnoreAccents = on })
}

// SetDeckByNoteType sets whether the add screen picks the deck last used
// with the chosen note type instead of the last used deck.
func (s *AppService) SetDeckByNoteType(on bool) error {
	return s.updateSettings(func(st *core.Settings) { st.DeckByNoteType = on })
}

// SetBrowserColumns sets the browser's columns, in order. At least one
// column is needed; an unknown or repeated column fails with a message.
func (s *AppService) SetBrowserColumns(columns []string) error {
	if len(columns) == 0 {
		return core.NewError("browserNeedsColumn", "the browser needs at least one column")
	}
	cols := make([]core.Column, len(columns))
	for i, name := range columns {
		cols[i] = core.Column(name)
	}
	return s.updateSettings(func(st *core.Settings) { st.BrowserColumns = cols })
}

// SetBrowserMode sets whether the browser lists cards ("cards") or notes
// ("notes").
func (s *AppService) SetBrowserMode(mode string) error {
	return s.updateSettings(func(st *core.Settings) { st.BrowserMode = mode })
}

// SetBrowserSplit sets the share of the width that the browser's list takes
// next to the details on wide screens; it must be between 0 and 1.
func (s *AppService) SetBrowserSplit(split float64) error {
	if split <= 0 {
		return core.NewError("browserSplitRange", "The browser split must be between 0 and 1.")
	}
	return s.updateSettings(func(st *core.Settings) { st.BrowserSplit = split })
}

// SetDefaultSearch sets the search the browser starts with; empty lists all
// cards. An invalid term fails with a message naming it.
func (s *AppService) SetDefaultSearch(query string) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	if err := c.CheckSearch(query); err != nil {
		return err
	}
	return s.updateSettings(func(st *core.Settings) { st.DefaultSearch = strings.TrimSpace(query) })
}

// SetBackupSettings sets the backup interval in minutes (0 turns periodic
// backups off) and how many daily, weekly and monthly backups to keep.
func (s *AppService) SetBackupSettings(backup core.BackupSettings) error {
	return s.updateSettings(func(st *core.Settings) { st.Backup = backup })
}

// SetFlagName renames flag 1 to 7. An empty name restores the flag's colour
// name.
func (s *AppService) SetFlagName(flag int, name string) error {
	if flag < 1 || flag > core.MaxFlag {
		return core.NewError("unknownFlag", "There is no flag %d.", flag)
	}
	return s.updateSettings(func(st *core.Settings) { st.FlagNames[flag-1] = strings.TrimSpace(name) })
}

// PersistenceNoticeShown records that the web storage notice was shown.
func (s *AppService) PersistenceNoticeShown() error {
	return s.updateSettings(func(st *core.Settings) { st.PersistenceNoticeShown = true })
}

func (s *AppService) updateSettings(edit func(*core.Settings)) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	st := c.Settings()
	edit(&st)
	return c.UpdateSettings(st)
}

// SetRenderEnvironment sets where cards are shown: theme, platform and the
// card's display size. Cards are rendered for it from then on; when it
// changes, Changes sends a change of kind "environment" so open cards
// render again.
func (s *AppService) SetRenderEnvironment(env render.Env) { s.a.setRenderEnv(env) }

// SetBuiltinNames sets the names, in the interface language, that a new
// collection's deck and preset and stock note types are created with. The
// app calls it before Open and whenever the interface language changes.
func (s *AppService) SetBuiltinNames(names core.BuiltinNames) { s.a.setBuiltinNames(names) }

// Changes streams a notification whenever the collection or the render
// environment changes, so open screens can refresh.
func (s *AppService) Changes(ctx context.Context) <-chan core.Change { return s.a.changes(ctx) }

var selfTestRuns atomic.Int64

// CheckStorage runs the storage backend contract against scratch stores
// next to the collection and returns one message per failed check. The
// integration tests use it to check IndexedDB in real browsers.
func (s *AppService) CheckStorage() ([]string, error) {
	s.a.mu.Lock()
	dir := s.a.dataDir
	s.a.mu.Unlock()
	run := selfTestRuns.Add(1)
	var scratches []string
	defer func() {
		for _, d := range scratches {
			s.a.removeBackend(d)
			if dir != "" {
				os.RemoveAll(d)
			}
		}
	}()
	n := 0
	errs := storagetest.RunAll(func() (storagetest.Opener, error) {
		n++
		scratch := fmt.Sprintf("selftest-%d-%d", run, n)
		if dir != "" {
			scratch = filepath.Join(dir, scratch)
		}
		scratches = append(scratches, scratch)
		if err := s.a.removeBackend(scratch); err != nil {
			return nil, err
		}
		return func() (storage.Backend, error) { return s.a.openBackend(scratch) }, nil
	})
	out := []string{}
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return out, nil
}

// AutoBackup is an automatic backup.
type AutoBackup struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Size      int64     `json:"size"`
}

// BackupList lists the automatic backups, newest first, and the storage
// space they use in bytes.
type BackupList struct {
	Backups []AutoBackup `json:"backups"`
	Bytes   int64        `json:"bytes"`
}

func (s *AppService) backupStore() (storage.BackupStore, error) {
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	if s.a.backend == nil {
		if s.a.opened {
			return nil, errors.New(s.a.status.Message)
		}
		return nil, errNotOpened
	}
	return s.a.backend.Backups(), nil
}

// Backups lists the automatic backups.
func (s *AppService) Backups() (BackupList, error) {
	store, err := s.backupStore()
	if err != nil {
		return BackupList{}, err
	}
	list, err := backup.ListAuto(store)
	if err != nil {
		return BackupList{}, err
	}
	out := BackupList{Backups: []AutoBackup{}}
	for _, b := range list {
		out.Backups = append(out.Backups, AutoBackup(b))
		out.Bytes += b.Size
	}
	return out, nil
}

// CreateBackup makes an automatic backup now.
func (s *AppService) CreateBackup() (AutoBackup, error) {
	c, err := s.a.collection()
	if err != nil {
		return AutoBackup{}, err
	}
	_, b := s.a.backupTarget()
	s.a.backups.mu.Lock()
	defer s.a.backups.mu.Unlock()
	made, err := s.a.writeBackup(c, b)
	return AutoBackup(made), err
}

// RestoreBackup replaces the collection with the automatic backup name,
// after backing up the current one. Media files stay as they are. The
// current collection stays unchanged if the backup is invalid.
func (s *AppService) RestoreBackup(name string) error {
	store, err := s.backupStore()
	if err != nil {
		return err
	}
	if _, ok := backup.ParseAutoName(name); !ok {
		return core.NewError("notAutomaticBackup", "%q is not an automatic backup", name)
	}
	data, err := store.Read(name)
	if errors.Is(err, storage.ErrNotFound) {
		return core.NewError("backupGone", "this backup no longer exists")
	}
	if err != nil {
		return err
	}
	a, err := backup.Read(data)
	if err != nil {
		return err
	}
	return s.a.restoreWith(func(c *core.Collection) error {
		return c.RestoreKeepingMedia(a.Collection, a.Revlog, a.Deletions)
	})
}

// UndoInfo names the steps Undo and Redo act on, e.g. "Delete 3 notes";
// empty when there is none.
type UndoInfo struct {
	Undo string `json:"undo"`
	Redo string `json:"redo"`
}

// UndoInfo returns what Undo and Redo would act on.
func (s *AppService) UndoInfo() (UndoInfo, error) {
	c, err := s.a.collection()
	if err != nil {
		return UndoInfo{}, err
	}
	undo, redo := c.UndoStatus()
	return UndoInfo{Undo: undo, Redo: redo}, nil
}

// Undo reverts the most recent change of this app session and returns its
// name.
func (s *AppService) Undo() (string, error) {
	c, err := s.a.collection()
	if err != nil {
		return "", err
	}
	return c.Undo()
}

// Redo re-applies the most recently undone change and returns its name.
func (s *AppService) Redo() (string, error) {
	c, err := s.a.collection()
	if err != nil {
		return "", err
	}
	return c.Redo()
}

// EmptyCard is a card whose question is empty.
type EmptyCard struct {
	CardID int64 `json:"cardId"`
	// Name is the card type name, or "Cloze N".
	Name string `json:"name"`
}

// EmptyNote is a note with empty cards. AllEmpty is set when every card of
// the note is empty: deleting empty cards keeps one of them.
type EmptyNote struct {
	NoteID   int64       `json:"noteId"`
	Text     string      `json:"text"`
	Cards    []EmptyCard `json:"cards"`
	AllEmpty bool        `json:"allEmpty"`
}

// EmptyCardList lists the empty cards by note. Total counts them, Deletable
// those DeleteEmptyCards deletes.
type EmptyCardList struct {
	Notes     []EmptyNote `json:"notes"`
	Total     int         `json:"total"`
	Deletable int         `json:"deletable"`
}

// EmptyCards lists the cards whose question is empty.
func (s *AppService) EmptyCards() (EmptyCardList, error) {
	c, err := s.a.collection()
	if err != nil {
		return EmptyCardList{}, err
	}
	list := c.EmptyCards()
	out := EmptyCardList{Notes: []EmptyNote{}, Total: list.Total, Deletable: list.Deletable}
	for _, n := range list.Notes {
		en := EmptyNote{NoteID: int64(n.NoteID), Text: n.Text, AllEmpty: n.AllEmpty}
		for _, card := range n.Cards {
			en.Cards = append(en.Cards, EmptyCard{CardID: int64(card.CardID), Name: card.Name})
		}
		out.Notes = append(out.Notes, en)
	}
	return out, nil
}

// DeleteEmptyCards deletes those of the cards that are still empty, in one
// change, keeping one card of each note whose cards are all empty.
func (s *AppService) DeleteEmptyCards(cardIDs []int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.DeleteEmptyCards(toIDs(cardIDs))
}

// DeletionLogFile is the exported deletion log.
type DeletionLogFile struct {
	// Name is the suggested file name, cardflip-deleted-<YYYY-MM-DD>.txt.
	Name string `json:"name"`
	Data []byte `json:"data"`
	// Notes is the number of deleted notes in the file; 0 means the log is
	// empty.
	Notes int `json:"notes"`
}

// ExportDeletionLog writes the deletion log as tab-separated text that the
// text import reads: one deleted note per line, its fields, then its tags,
// note type and deletion time.
func (s *AppService) ExportDeletionLog() (DeletionLogFile, error) {
	c, err := s.a.collection()
	if err != nil {
		return DeletionLogFile{}, err
	}
	log, err := c.DeletionLog()
	if err != nil {
		return DeletionLogFile{}, err
	}
	name := "cardflip-deleted-" + s.a.now().Format("2006-01-02") + ".txt"
	return DeletionLogFile{Name: name, Data: deletionLogText(log, c.NoteTypes()), Notes: len(log)}, nil
}

// ClearDeletionLog empties the deletion log. The app only calls it after
// the user confirmed.
func (s *AppService) ClearDeletionLog() error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.ClearDeletionLog()
}

// CheckMedia reports the media files no note, template or styling refers
// to, the file names notes refer to that are missing, and the media trash.
func (s *AppService) CheckMedia() (core.MediaCheck, error) {
	c, err := s.a.collection()
	if err != nil {
		return core.MediaCheck{}, err
	}
	r, err := c.CheckMedia()
	if err != nil {
		return core.MediaCheck{}, err
	}
	return *r, nil
}

// CheckDatabaseEvent is progress, then the final report or an error.
type CheckDatabaseEvent struct {
	// Stage is "backup", "notes", "cards", "history", "saving" or "done".
	Stage  string            `json:"stage"`
	Done   int               `json:"done"`
	Total  int               `json:"total"`
	Report *core.CheckReport `json:"report"`
	// Error is set when the check failed; the collection is unchanged.
	Error string `json:"error"`
}

// CheckDatabase checks the collection for inconsistencies and repairs them
// in one change, after a backup of the collection, then compacts the
// storage. It streams progress and ends with the report or an error.
func (s *AppService) CheckDatabase(ctx context.Context) <-chan CheckDatabaseEvent {
	out := make(chan CheckDatabaseEvent, 16)
	go func() {
		defer close(out)
		send := func(e CheckDatabaseEvent) {
			select {
			case out <- e:
			case <-ctx.Done():
			}
		}
		r, err := s.checkDatabase(func(stage string, done, total int) {
			// Progress is best effort: never block the check on a slow UI.
			select {
			case out <- CheckDatabaseEvent{Stage: stage, Done: done, Total: total}:
			default:
			}
		})
		if err != nil {
			send(CheckDatabaseEvent{Error: err.Error()})
			return
		}
		send(CheckDatabaseEvent{Stage: "done", Done: 1, Total: 1, Report: &r})
	}()
	return out
}

func (s *AppService) checkDatabase(progress core.Progress) (core.CheckReport, error) {
	c, err := s.a.collection()
	if err != nil {
		return core.CheckReport{}, err
	}
	progress("backup", 0, 1)
	if err := s.a.backupBeforeOperation(); err != nil {
		return core.CheckReport{}, err
	}
	return c.CheckDatabase(progress)
}

// DeleteUnusedMedia moves the unused media files to the media trash and
// returns how many it moved.
func (s *AppService) DeleteUnusedMedia() (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.DeleteUnusedMedia()
}

// TagMissingMedia adds the tag missing-media to the notes that refer to
// missing media files and returns how many notes it tagged.
func (s *AppService) TagMissingMedia() (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.TagMissingMedia()
}

// EmptyMediaTrash deletes the files in the media trash for good. The app
// only calls it after the user confirmed.
func (s *AppService) EmptyMediaTrash() error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.EmptyMediaTrash()
}

// RestoreMediaTrash moves the files in the media trash back into the media
// and returns how many it restored. A file whose name a different file has
// taken since stays in the trash.
func (s *AppService) RestoreMediaTrash() (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.RestoreMediaTrash()
}

// deletionLogText formats the deletion log with the headers of Anki's text
// export. Every row has as many field columns as the note with the most
// fields, so that the tags are always in the same column.
func deletionLogText(log []core.DeletedNote, types []core.NoteType) []byte {
	fields := 0
	sameType := true
	for _, e := range log {
		fields = max(fields, len(e.Fields))
		sameType = sameType && e.NoteType == log[0].NoteType
	}
	// Columns are named after the note type when all notes have the same
	// one and it still has those fields.
	names := make([]string, fields)
	for i := range names {
		names[i] = fmt.Sprintf("Field %d", i+1)
	}
	for _, nt := range types {
		if sameType && len(log) > 0 && nt.Name == log[0].NoteType && len(nt.Fields) == fields {
			copy(names, nt.Fields)
		}
	}
	names = append(names, "Tags", "Note type", "Deleted")
	for i, n := range names {
		// The importer splits column names at these.
		names[i] = strings.Map(func(r rune) rune {
			if r == '\t' || r == ',' || r == ';' {
				return ' '
			}
			return r
		}, n)
	}
	var b strings.Builder
	b.WriteString("#separator:tab\n#html:true\n")
	b.WriteString("#columns:" + strings.Join(names, "\t") + "\n")
	fmt.Fprintf(&b, "#tags column:%d\n#notetype column:%d\n", fields+1, fields+2)
	for _, e := range log {
		row := make([]string, 0, fields+3)
		for i := range fields {
			v := ""
			if i < len(e.Fields) {
				v = e.Fields[i]
			}
			row = append(row, v)
		}
		row = append(row, strings.Join(e.Tags, " "), e.NoteType, e.Time.Local().Format("2006-01-02 15:04:05"))
		for i, v := range row {
			row[i] = tsvCell(v)
		}
		b.WriteString(strings.Join(row, "\t") + "\n")
	}
	return []byte(b.String())
}

// tsvCell quotes a value when the importer would otherwise misread it. Line
// breaks become spaces: in HTML they display as one, and the importer would
// turn them into visible <br> breaks.
func tsvCell(v string) string {
	v = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(v)
	if strings.ContainsAny(v, "\t\"") || strings.HasPrefix(v, "#") {
		return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
	}
	return v
}
