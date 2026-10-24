package main

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// BrowseService searches, inspects and bulk-edits cards.
type BrowseService struct{ a *app }

// Search returns the IDs of the cards matching query, sorted by a browser
// column (see Rows), ascending or descending. An invalid term fails with a
// message naming it.
func (s *BrowseService) Search(query string, column string, descending bool) ([]int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	col, err := columnOf(column)
	if err != nil {
		return nil, err
	}
	ids, err := c.Search(query, col, descending)
	if err != nil {
		return nil, err
	}
	return fromIDs(ids), nil
}

func columnOf(name string) (core.Column, error) {
	col := core.Column(name)
	if !col.Valid() {
		return "", core.NewError("unknownBrowserColumn", "Unknown browser column %q.", name)
	}
	return col, nil
}

// SavedSearches returns the saved searches, sorted by name.
func (s *BrowseService) SavedSearches() ([]core.SavedSearch, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	return c.Settings().SavedSearches, nil
}

// SaveSearch saves query under name, which no other saved search may have
// (ignoring case). An invalid term fails with a message naming it.
func (s *BrowseService) SaveSearch(name, query string) error {
	return s.editSavedSearches(func(c *core.Collection, saved []core.SavedSearch) ([]core.SavedSearch, error) {
		if err := c.CheckSearch(query); err != nil {
			return nil, err
		}
		return append(saved, core.SavedSearch{Name: strings.TrimSpace(name), Query: strings.TrimSpace(query)}), nil
	})
}

// RenameSavedSearch renames the saved search name to newName, which no
// other saved search may have (ignoring case).
func (s *BrowseService) RenameSavedSearch(name, newName string) error {
	return s.editSavedSearches(func(_ *core.Collection, saved []core.SavedSearch) ([]core.SavedSearch, error) {
		i, err := savedSearch(saved, name)
		if err != nil {
			return nil, err
		}
		saved[i].Name = strings.TrimSpace(newName)
		return saved, nil
	})
}

// DeleteSavedSearch deletes the saved search name.
func (s *BrowseService) DeleteSavedSearch(name string) error {
	return s.editSavedSearches(func(_ *core.Collection, saved []core.SavedSearch) ([]core.SavedSearch, error) {
		i, err := savedSearch(saved, name)
		if err != nil {
			return nil, err
		}
		return slices.Delete(saved, i, i+1), nil
	})
}

// editSavedSearches stores the saved searches edit returns for a copy of
// the current ones, sorted by name.
func (s *BrowseService) editSavedSearches(edit func(*core.Collection, []core.SavedSearch) ([]core.SavedSearch, error)) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	st := c.Settings()
	saved, err := edit(c, slices.Clone(st.SavedSearches))
	if err != nil {
		return err
	}
	slices.SortStableFunc(saved, func(a, b core.SavedSearch) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.Name, b.Name))
	})
	st.SavedSearches = saved
	return c.UpdateSettings(st)
}

func savedSearch(saved []core.SavedSearch, name string) (int, error) {
	i := slices.IndexFunc(saved, func(s core.SavedSearch) bool { return s.Name == name })
	if i < 0 {
		return 0, core.NewError("savedSearchMissing", "there is no saved search named %q", name)
	}
	return i, nil
}

// Row is one card in the browser list. Only the requested columns' values
// are set; the others are empty or nil.
type Row struct {
	CardID int64 `json:"cardId"`
	NoteID int64 `json:"noteId"`
	// State is "new", "learning", "review" or "relearning".
	State     string `json:"state"`
	Suspended bool   `json:"suspended"`
	// Flag is 0 for no flag, or 1 to 7.
	Flag int `json:"flag"`
	// Marked is set when the note has the tag "marked".
	Marked bool `json:"marked"`
	// Question, Answer and SortField are plain text; the answer leaves out
	// the question it repeats.
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	SortField string `json:"sortField"`
	CardType  string `json:"cardType"`
	NoteType  string `json:"noteType"`
	Deck      string `json:"deck"`
	// Tags are the note's tags, separated by spaces.
	Tags string `json:"tags"`
	// Due is a scheduled card's due time, Position a New card's position
	// in the new card queue.
	Due      *time.Time `json:"due"`
	Position *int64     `json:"position"`
	// Interval is a Review card's interval in days; nil for other cards.
	Interval *int `json:"interval"`
	// Difficulty (0 to 1), Stability (days) and Retrievability (0 to 1)
	// are nil when the card has no memory state.
	Difficulty     *float64   `json:"difficulty"`
	Stability      *float64   `json:"stability"`
	Retrievability *float64   `json:"retrievability"`
	Reviews        int        `json:"reviews"`
	Lapses         int        `json:"lapses"`
	Created        *time.Time `json:"created"`
	CardModified   *time.Time `json:"cardModified"`
	NoteModified   *time.Time `json:"noteModified"`
}

// Rows returns the rows for a page of card IDs with the values of the
// given columns: "question", "answer", "sortField", "cardType", "noteType",
// "deck", "due", "interval", "difficulty", "stability", "retrievability",
// "reviews", "lapses", "tags", "created", "cardModified" and
// "noteModified". No columns means all of them.
func (s *BrowseService) Rows(ids []int64, columns []string) ([]Row, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	cols := make([]core.Column, len(columns))
	for i, name := range columns {
		if cols[i], err = columnOf(name); err != nil {
			return nil, err
		}
	}
	if len(cols) == 0 {
		cols = core.Columns
	}
	rows := c.Rows(toIDs(ids), cols...)
	out := make([]Row, len(rows))
	for i, r := range rows {
		o := Row{CardID: int64(r.CardID), NoteID: int64(r.NoteID), State: r.State.String(), Suspended: r.Suspended,
			Flag: r.Flag, Marked: r.Marked, Question: r.Question, Answer: r.Answer, SortField: r.SortField, CardType: r.CardName,
			NoteType: r.NoteType, Deck: r.Deck, Tags: r.Tags, Reviews: r.Reviews, Lapses: r.Lapses}
		for _, col := range cols {
			switch col {
			case core.ColDue:
				if r.State == scheduler.New {
					o.Position = &r.Position
				} else {
					o.Due = &r.Due
				}
			case core.ColInterval:
				if !math.IsNaN(r.Interval) {
					days := int(r.Interval)
					o.Interval = &days
				}
			case core.ColDifficulty:
				o.Difficulty = number(r.Difficulty)
			case core.ColStability:
				o.Stability = number(r.Stability)
			case core.ColRetrievability:
				o.Retrievability = number(r.Retrievability)
			case core.ColCreated:
				o.Created = &r.Created
			case core.ColCardModified:
				o.CardModified = &r.CardModified
			case core.ColNoteModified:
				o.NoteModified = &r.NoteModified
			}
		}
		out[i] = o
	}
	return out, nil
}

// number is v, or nil when it is NaN.
func number(v float64) *float64 {
	if math.IsNaN(v) {
		return nil
	}
	return &v
}

// SearchNotes returns the IDs of the notes with a card matching query, for
// the browser's Notes mode, sorted by "created", "sortField" (the note
// type's sort field) or "due" (the earliest due of the note's cards),
// ascending or descending. An invalid term fails with a message naming it.
func (s *BrowseService) SearchNotes(query string, column string, descending bool) ([]int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	ids, err := c.SearchNotes(query, core.Column(column), descending)
	if err != nil {
		return nil, err
	}
	return fromIDs(ids), nil
}

// NoteRow is one note in the browser's Notes mode.
type NoteRow struct {
	NoteID int64 `json:"noteId"`
	// CardID is the note's first card, whose details the browser shows.
	CardID int64 `json:"cardId"`
	// SortField is the note type's sort field as plain text.
	SortField string `json:"sortField"`
	NoteType  string `json:"noteType"`
	Cards     int    `json:"cards"`
	// Deck is the deck of the note's cards; empty when they are in several
	// decks (MultipleDecks).
	Deck          string `json:"deck"`
	MultipleDecks bool   `json:"multipleDecks"`
	// Tags are the note's tags, separated by spaces.
	Tags string `json:"tags"`
	// State is the state of the card due first: "new", "learning",
	// "review" or "relearning". Due is its due time, or Position its
	// position in the new card queue when it is new.
	State    string     `json:"state"`
	Due      *time.Time `json:"due"`
	Position *int64     `json:"position"`
	// Suspended is set when all of the note's cards are suspended.
	Suspended bool `json:"suspended"`
	// Marked is set when the note has the tag "marked".
	Marked  bool      `json:"marked"`
	Created time.Time `json:"created"`
}

// NoteRows returns the Notes mode rows for a page of note IDs.
func (s *BrowseService) NoteRows(ids []int64) ([]NoteRow, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	rows := c.NoteRows(toIDs(ids))
	out := make([]NoteRow, len(rows))
	for i, r := range rows {
		o := NoteRow{NoteID: int64(r.NoteID), CardID: int64(r.CardID), SortField: r.SortField, NoteType: r.NoteType,
			Cards: r.Cards, Deck: r.Deck, MultipleDecks: r.MultipleDecks, Tags: r.Tags, State: r.DueState.String(),
			Suspended: r.Suspended, Marked: r.Marked, Created: r.Created}
		if r.DueState == scheduler.New {
			o.Position = &r.Position
		} else {
			o.Due = &r.Due
		}
		out[i] = o
	}
	return out, nil
}

// NoteCards returns the IDs of all cards of the notes, which Notes mode
// passes to the bulk actions.
func (s *BrowseService) NoteCards(noteIDs []int64) ([]int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	return fromIDs(c.CardsOfNotes(toIDs(noteIDs))), nil
}

// CardsOfNotes returns the IDs of all cards of the notes of the cards, for
// the browser's Select notes.
func (s *BrowseService) CardsOfNotes(cardIDs []int64) ([]int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	return fromIDs(c.CardsOfNotes(notesOf(c, cardIDs))), nil
}

// CardDetails is everything the browser's details pane shows.
type CardDetails struct {
	CardID int64 `json:"cardId"`
	// DeckID is the card's home deck, where a copy of its note goes; Deck
	// names the deck it is in, with the home deck when that is a filtered
	// deck.
	DeckID     int64  `json:"deckId"`
	Deck       string `json:"deck"`
	CardName   string `json:"cardName"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	Background string `json:"background,omitempty"`
	Color      string `json:"color,omitempty"`
	// QuestionMedia and AnswerMedia are each side's sounds and videos, and
	// QuestionTTS and AnswerTTS its text to speech. Fonts are the note
	// type's fonts from the media.
	QuestionMedia []render.MediaRef  `json:"questionMedia"`
	AnswerMedia   []render.MediaRef  `json:"answerMedia"`
	QuestionTTS   []render.Utterance `json:"questionTts"`
	AnswerTTS     []render.Utterance `json:"answerTts"`
	Fonts         []render.FontFace  `json:"fonts"`
	// Scheduling information.
	State string     `json:"state"`
	Due   *time.Time `json:"due"`
	// Position is a New card's position in the new card queue; nil for
	// other cards.
	Position   *int64   `json:"position"`
	Interval   int      `json:"intervalDays"`
	Stability  *float64 `json:"stability"`
	Difficulty *float64 `json:"difficulty"`
	Reviews    int      `json:"reviews"`
	Lapses     int      `json:"lapses"`
	Suspended  bool     `json:"suspended"`
	Flag       int      `json:"flag"`
	// The note, for the editor.
	Note     Note     `json:"note"`
	NoteType NoteType `json:"noteType"`
	// NoteCards is how many cards the note has, this one included.
	NoteCards int `json:"noteCards"`
}

// Card returns a card's details.
func (s *BrowseService) Card(id int64) (CardDetails, error) {
	c, err := s.a.collection()
	if err != nil {
		return CardDetails{}, err
	}
	card, err := c.Card(core.ID(id))
	if err != nil {
		return CardDetails{}, err
	}
	r, err := c.RenderCard(card.ID, s.a.renderEnv())
	if err != nil {
		return CardDetails{}, err
	}
	n, err := c.Note(card.NoteID)
	if err != nil {
		return CardDetails{}, err
	}
	nt, err := c.NoteType(n.TypeID)
	if err != nil {
		return CardDetails{}, err
	}
	row := c.Rows([]core.ID{card.ID}, core.ColCardType)
	d := CardDetails{
		CardID: int64(card.ID), DeckID: int64(card.HomeDeck()), Deck: c.DeckLabel(card.ID),
		Question: r.Question, Answer: r.Answer, Background: r.Background, Color: r.Color,
		QuestionMedia: r.QuestionMedia, AnswerMedia: r.AnswerMedia, QuestionTTS: r.QuestionTTS, AnswerTTS: r.AnswerTTS, Fonts: r.Fonts,
		State: card.State.String(), Interval: card.ScheduledDays, Reviews: card.Reps, Lapses: card.Lapses,
		Suspended: card.Suspended, Flag: card.Flag, NoteType: noteTypeOf(nt, 0),
		Note: Note{ID: int64(n.ID), TypeID: int64(n.TypeID), Fields: n.Fields, Tags: n.Tags},
	}
	if len(row) == 1 {
		d.CardName = row[0].CardName
	}
	if card.State != scheduler.New {
		due := card.Due
		d.Due = &due
	} else {
		pos := card.NewPos
		d.Position = &pos
	}
	if card.Memory != nil {
		st, df := card.Memory.Stability, card.Memory.Difficulty
		d.Stability, d.Difficulty = &st, &df
	}
	for _, cc := range c.CardsOfNote(n.ID) {
		d.Note.CardIDs = append(d.Note.CardIDs, int64(cc.ID))
	}
	d.NoteCards = len(d.Note.CardIDs)
	return d, nil
}

// CardInfo is a card's identity, scheduling, memory state and review
// history. Values that do not apply to the card are nil.
type CardInfo struct {
	CardID       int64      `json:"cardId"`
	NoteID       int64      `json:"noteId"`
	Added        time.Time  `json:"added"`
	FirstReview  *time.Time `json:"firstReview"`
	LatestReview *time.Time `json:"latestReview"`
	// State is "new", "learning", "review" or "relearning".
	State string     `json:"state"`
	Due   *time.Time `json:"due"`
	// Position is a New card's position in the new card queue.
	Position *int64 `json:"position"`
	// Interval is the time from the latest review to the due date.
	Interval       *time.Duration `json:"interval"`
	Stability      *float64       `json:"stability"`
	Difficulty     *float64       `json:"difficulty"`
	Retrievability *float64       `json:"retrievability"`
	Reviews        int            `json:"reviews"`
	Lapses         int            `json:"lapses"`
	// AverageTime and TotalTime are the answer times of the history's
	// answers.
	AverageTime *time.Duration `json:"averageTime"`
	TotalTime   *time.Duration `json:"totalTime"`
	CardType    string         `json:"cardType"`
	NoteType    string         `json:"noteType"`
	Deck        string         `json:"deck"`
	// History is the review log, newest first.
	History []HistoryEntry `json:"history"`
}

// HistoryEntry is one review log entry in card info.
type HistoryEntry struct {
	Time time.Time `json:"time"`
	// Kind is "learning", "review", "relearning" or "manual".
	Kind string `json:"kind"`
	// Rating is 1 (Again) to 4 (Easy), 0 for a manual entry.
	Rating int `json:"rating"`
	// Interval is the time until the card was due after the entry; nil
	// when unknown.
	Interval   *time.Duration `json:"interval"`
	AnswerTime time.Duration  `json:"answerTime"`
}

// CardInfo returns a card's info with its whole review history.
func (s *BrowseService) CardInfo(id int64) (CardInfo, error) {
	c, err := s.a.collection()
	if err != nil {
		return CardInfo{}, err
	}
	card, err := c.Card(core.ID(id))
	if err != nil {
		return CardInfo{}, err
	}
	hist, err := c.CardHistory(card.ID)
	if err != nil {
		return CardInfo{}, err
	}
	info := CardInfo{CardID: int64(card.ID), NoteID: int64(card.NoteID), Added: time.UnixMilli(int64(card.ID)),
		State: card.State.String(), Reviews: card.Reps, Lapses: card.Lapses, History: []HistoryEntry{}}
	if row := c.Rows([]core.ID{card.ID}, core.ColCardType, core.ColNoteType, core.ColDeck); len(row) == 1 {
		info.CardType, info.NoteType, info.Deck = row[0].CardName, row[0].NoteType, row[0].Deck
	}
	if card.State == scheduler.New {
		pos := card.NewPos
		info.Position = &pos
	} else {
		due := card.Due
		info.Due = &due
		info.Interval = intervalAfter(card.CardState, card.LastReview, card.ScheduledDays)
	}
	if card.Memory != nil {
		st, df := card.Memory.Stability, card.Memory.Difficulty
		info.Stability, info.Difficulty = &st, &df
		if r, ok := scheduler.Retrievability(card.CardState, c.Now()); ok {
			info.Retrievability = &r
		}
	}

	var total time.Duration
	answers := 0
	for i, e := range hist {
		h := HistoryEntry{Time: e.Time, Kind: historyKind(e), Rating: int(e.Rating),
			AnswerTime: time.Duration(e.DurationMs) * time.Millisecond}
		// Answers in Cardflip record intraday steps as 0 days; the state
		// after them is the next entry's state before, or the card's.
		after := card.CardState
		if i+1 < len(hist) {
			after = hist[i+1].Before
		}
		h.Interval = intervalAfter(after, e.Time, e.IvlAfter)
		info.History = append(info.History, h)
		if e.Rating == 0 {
			continue
		}
		answers++
		total += h.AnswerTime
		t := e.Time
		if info.FirstReview == nil {
			info.FirstReview = &t
		}
		info.LatestReview = &t
	}
	slices.Reverse(info.History)
	if answers > 0 {
		avg := total / time.Duration(answers)
		info.TotalTime, info.AverageTime = &total, &avg
	}
	return info, nil
}

// intervalAfter is the interval a card got at time at: days when it is
// one or more, otherwise the time until the due date of after, the card's
// state that followed. It is nil when neither is known.
func intervalAfter(after scheduler.CardState, at time.Time, days int) *time.Duration {
	var d time.Duration
	switch {
	case days > 0:
		d = time.Duration(days) * 24 * time.Hour
	case !at.IsZero() && after.State != scheduler.New && after.Due.After(at):
		d = after.Due.Sub(at)
	default:
		return nil
	}
	return &d
}

func historyKind(e core.ReviewEntry) string {
	switch {
	case e.Rating == 0:
		return "manual"
	case e.StateBefore == scheduler.Review:
		return "review"
	case e.StateBefore == scheduler.Relearning:
		return "relearning"
	}
	return "learning"
}

// SetSuspended suspends or unsuspends cards.
func (s *BrowseService) SetSuspended(cardIDs []int64, suspended bool) error {
	return s.edit(func(c *core.Collection) error { return c.SetSuspended(toIDs(cardIDs), suspended) })
}

// SetFlag sets flag 1 to 7 on the cards, or clears it when every one of them
// already has it. Flag 0 clears their flags.
func (s *BrowseService) SetFlag(cardIDs []int64, flag int) error {
	return s.edit(func(c *core.Collection) error { return c.SetFlag(toIDs(cardIDs), flag) })
}

// ToggleMark adds the tag "marked" to the notes of the cards when any of
// them lacks it, and otherwise removes it from all of them.
func (s *BrowseService) ToggleMark(cardIDs []int64) error {
	return s.edit(func(c *core.Collection) error { return c.ToggleMark(notesOf(c, cardIDs)) })
}

// Move moves cards to a deck; their scheduling is unchanged.
func (s *BrowseService) Move(cardIDs []int64, deckID int64) error {
	return s.edit(func(c *core.Collection) error { return c.MoveCards(toIDs(cardIDs), core.ID(deckID)) })
}

// AddTags adds tags to the notes of the cards.
func (s *BrowseService) AddTags(cardIDs []int64, tags []string) error {
	return s.edit(func(c *core.Collection) error { return c.AddTags(notesOf(c, cardIDs), tags) })
}

// RemoveTags removes tags from the notes of the cards.
func (s *BrowseService) RemoveTags(cardIDs []int64, tags []string) error {
	return s.edit(func(c *core.Collection) error { return c.RemoveTags(notesOf(c, cardIDs), tags) })
}

// ResetOptions say how Reset makes cards New again. The reset dialog starts
// with RestorePosition on and ResetCounts off.
type ResetOptions struct {
	// RestorePosition puts each card back at its original position among
	// new cards, where it is known; other cards go after all new cards.
	RestorePosition bool `json:"restorePosition"`
	// ResetCounts sets the review and lapse counts to zero.
	ResetCounts bool `json:"resetCounts"`
}

func (o ResetOptions) core() core.ResetOptions {
	return core.ResetOptions{RestorePosition: o.RestorePosition, ResetCounts: o.ResetCounts}
}

// ResetToNew makes cards New again (see ResetOptions); their review history
// is kept and gains a manual entry per card.
func (s *BrowseService) ResetToNew(cardIDs []int64, opts ResetOptions) error {
	return s.edit(func(c *core.Collection) error { return c.ResetCards(toIDs(cardIDs), opts.core()) })
}

// SetDueDate makes the cards Review cards due on a day from spec: "N" days
// from today (0 is today), "N-M" a random day in that range per card, and a
// trailing "!" sets the interval too. Invalid input is rejected and changes
// nothing. Settings.LastDueDate remembers the input.
func (s *BrowseService) SetDueDate(cardIDs []int64, spec string) error {
	return s.edit(func(c *core.Collection) error { return c.SetDueDate(toIDs(cardIDs), spec) })
}

// Reposition gives the New cards among cardIDs the positions start,
// start+step, … in their current order, or in random order with randomize.
// With shift, other new cards at or after start move back to make room.
// Cards that are not New are left unchanged; it returns how many.
func (s *BrowseService) Reposition(cardIDs []int64, start, step int64, randomize, shift bool) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.Reposition(toIDs(cardIDs), start, step, randomize, shift)
}

// DeleteNotes deletes the notes of the cards, with all their cards.
func (s *BrowseService) DeleteNotes(cardIDs []int64) error {
	return s.edit(func(c *core.Collection) error { return c.DeleteNotes(notesOf(c, cardIDs)) })
}

// DeleteImpact is what deleting the notes of some cards would remove.
type DeleteImpact struct {
	Notes int `json:"notes"`
	// Cards counts every card of those notes, including ones not picked.
	Cards int `json:"cards"`
}

// DeleteImpact returns how many notes and cards DeleteNotes would delete,
// for confirmations.
func (s *BrowseService) DeleteImpact(cardIDs []int64) (DeleteImpact, error) {
	c, err := s.a.collection()
	if err != nil {
		return DeleteImpact{}, err
	}
	notes := notesOf(c, cardIDs)
	out := DeleteImpact{Notes: len(notes)}
	for _, n := range notes {
		out.Cards += len(c.CardsOfNote(n))
	}
	return out, nil
}

// FindReplaceOptions say what Find and Replace changes.
type FindReplaceOptions struct {
	Find    string `json:"find"`
	Replace string `json:"replace"`
	// Field is the one field to change, by name; "" changes all fields.
	Field string `json:"field"`
	// Tags changes the tags instead of the fields.
	Tags bool `json:"tags"`
	// Regex reads Find as a regular expression (RE2); Replace can then use
	// $1 or ${name}.
	Regex      bool `json:"regex"`
	IgnoreCase bool `json:"ignoreCase"`
}

func (o FindReplaceOptions) core() core.FindReplace {
	return core.FindReplace{Find: o.Find, Replace: o.Replace, Field: o.Field, Tags: o.Tags, Regex: o.Regex, IgnoreCase: o.IgnoreCase}
}

// scope returns the notes of the cards, or every note with all.
func scope(c *core.Collection, cardIDs []int64, all bool) []core.ID {
	if all {
		return c.AllNoteIDs()
	}
	return notesOf(c, cardIDs)
}

// FindReplacePreview returns how many notes FindReplace would change. An
// invalid regular expression fails with a message.
func (s *BrowseService) FindReplacePreview(cardIDs []int64, all bool, opts FindReplaceOptions) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.FindReplacePreview(scope(c, cardIDs, all), opts.core())
}

// FindReplace replaces text in the notes of the cards, or in every note
// with all, as one undo step. It returns how many notes changed.
func (s *BrowseService) FindReplace(cardIDs []int64, all bool, opts FindReplaceOptions) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.FindReplace(scope(c, cardIDs, all), opts.core())
}

// FieldNames lists the fields of the note types of the cards' notes, or of
// every note with all, each name once, in note type and field order.
func (s *BrowseService) FieldNames(cardIDs []int64, all bool) ([]string, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	types := map[core.ID]bool{}
	for _, id := range scope(c, cardIDs, all) {
		if n, err := c.Note(id); err == nil {
			types[n.TypeID] = true
		}
	}
	out := []string{}
	for _, nt := range c.NoteTypes() {
		if !types[nt.ID] {
			continue
		}
		for _, f := range nt.Fields {
			if !slices.Contains(out, f) {
				out = append(out, f)
			}
		}
	}
	return out, nil
}

// NoteTypeOfCards returns the note type the notes of the cards share, for
// Change note type. Cards of several note types fail with a message.
func (s *BrowseService) NoteTypeOfCards(cardIDs []int64) (int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	var typeID core.ID
	for _, nid := range notesOf(c, cardIDs) {
		n, err := c.Note(nid)
		if err != nil {
			continue
		}
		if typeID != 0 && n.TypeID != typeID {
			return 0, core.NewError("sameNoteType", "All selected cards must use the same note type.")
		}
		typeID = n.TypeID
	}
	if typeID == 0 {
		return 0, core.NewError("selectCardsToChange", "Select the cards of the notes to change.")
	}
	return int64(typeID), nil
}

// ChangeNoteTypePreview returns what ChangeNoteType would do: how many
// notes change, how many cards are deleted and created, and which source
// fields lose their content.
func (s *BrowseService) ChangeNoteTypePreview(cardIDs []int64, targetTypeID int64, fieldMap, templateMap []int) (core.NoteTypeChange, error) {
	c, err := s.a.collection()
	if err != nil {
		return core.NoteTypeChange{}, err
	}
	return c.ChangeNoteTypePreview(notesOf(c, cardIDs), core.ID(targetTypeID), fieldMap, templateMap)
}

// ChangeNoteType changes the notes of the cards, which must share one note
// type, to note type targetTypeID, as one undo step. fieldMap gives for each
// target field the index of the source field it takes, or -1 for none;
// templateMap does the same for card types, between standard note types
// only. Cards whose card type is mapped keep their scheduling and review
// history; the others are deleted, and missing cards are created New.
func (s *BrowseService) ChangeNoteType(cardIDs []int64, targetTypeID int64, fieldMap, templateMap []int) (core.NoteTypeChange, error) {
	c, err := s.a.collection()
	if err != nil {
		return core.NoteTypeChange{}, err
	}
	return c.ChangeNoteType(notesOf(c, cardIDs), core.ID(targetTypeID), fieldMap, templateMap)
}

// DuplicateSet is a set of notes whose chosen field has the same content.
type DuplicateSet struct {
	// Content is the field as plain text, from the oldest note.
	Content string  `json:"content"`
	NoteIDs []int64 `json:"noteIds"`
}

// FindDuplicates returns the sets of notes whose field fieldName has the
// same content, ignoring HTML, case and surrounding whitespace, among the
// notes of note types with that field and, unless query is empty, a card
// matching query. Empty fields are never duplicates. Larger sets come
// first. An invalid term fails with a message naming it.
func (s *BrowseService) FindDuplicates(fieldName, query string) ([]DuplicateSet, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	sets, err := c.FindDuplicates(fieldName, query)
	if err != nil {
		return nil, err
	}
	out := make([]DuplicateSet, len(sets))
	for i, set := range sets {
		out[i] = DuplicateSet{Content: set.Content, NoteIDs: fromIDs(set.Notes)}
	}
	return out, nil
}

func (s *BrowseService) edit(fn func(c *core.Collection) error) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return fn(c)
}

func notesOf(c *core.Collection, cardIDs []int64) []core.ID {
	seen := map[core.ID]bool{}
	var out []core.ID
	for _, id := range cardIDs {
		if card, err := c.Card(core.ID(id)); err == nil && !seen[card.NoteID] {
			seen[card.NoteID] = true
			out = append(out, card.NoteID)
		}
	}
	return out
}
