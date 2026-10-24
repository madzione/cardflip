package main

import (
	"slices"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
)

// StudyService runs study sessions: one deck at a time, one card at a time.
type StudyService struct {
	a  *app
	mu sync.Mutex
	s  *session
}

type session struct {
	deck    core.ID
	current core.ID // card on screen
	// answers are the session's answers, oldest first. Those whose review
	// log entry is gone were undone, here or through the app's undo
	// history, and do not count; a redo brings them back.
	answers []answer
}

type answer struct {
	entry    core.ID
	card     core.ID
	wasNew   bool
	wasRev   bool
	duration time.Duration
}

// sessionView is the session as of now: the answers that were not undone.
type sessionView struct {
	answers     []answer
	studied     int
	newShown    int
	reviewShown int
	spent       time.Duration
}

func (sess *session) view(c *core.Collection) sessionView {
	ids := make([]core.ID, len(sess.answers))
	for i, a := range sess.answers {
		ids[i] = a.entry
	}
	exist := c.ReviewsExist(ids)
	var v sessionView
	for _, a := range sess.answers {
		if !exist[a.entry] {
			continue
		}
		v.answers = append(v.answers, a)
		v.studied++
		v.spent += a.duration
		if a.wasNew {
			v.newShown++
		}
		if a.wasRev {
			v.reviewShown++
		}
	}
	return v
}

// Preview is a rating button's interval, which the app shows short, e.g.
// "<10m", "3d", "1.2mo" (see core.FormatInterval).
type Preview struct {
	Rating   int           `json:"rating"`
	Interval time.Duration `json:"interval"`
	// Ends is set in a filtered deck's preview mode when the rating ends
	// the preview of the card, which the button says instead of an
	// interval.
	Ends bool `json:"ends"`
}

// StudyCard is the card to show.
type StudyCard struct {
	CardID int64 `json:"cardId"`
	NoteID int64 `json:"noteId"`
	// Question and Answer are sanitized HTML with styles inlined.
	Question string `json:"question"`
	Answer   string `json:"answer"`
	// Background and Color come from the note type's CSS; empty means the
	// app theme decides.
	Background string    `json:"background,omitempty"`
	Color      string    `json:"color,omitempty"`
	Previews   []Preview `json:"previews"`
	// Kind is "new", "learning" or "review".
	Kind string `json:"kind"`
	// Empty is set when the card's question is empty, so Empty Cards would
	// remove it.
	Empty bool `json:"empty"`
	// QuestionMedia and AnswerMedia are each side's sounds and videos in
	// order; the <cf-sound idx=N> and <cf-video idx=N> placeholders in the
	// HTML index them.
	QuestionMedia []render.MediaRef `json:"questionMedia"`
	AnswerMedia   []render.MediaRef `json:"answerMedia"`
	// QuestionTTS and AnswerTTS are each side's text to speech in order;
	// the <cf-tts idx=N> placeholders in the HTML index them.
	QuestionTTS []render.Utterance `json:"questionTts"`
	AnswerTTS   []render.Utterance `json:"answerTts"`
	// Fonts are the note type's fonts from the media, for its font-family
	// declarations.
	Fonts []render.FontFace `json:"fonts"`
	// Autoplay and ReplayQuestion come from the card's deck options: play
	// a side's sounds when it is shown, and replay the question's sounds
	// before the answer's.
	Autoplay       bool `json:"autoplay"`
	ReplayQuestion bool `json:"replayQuestion"`
	// ShowTimer, StopTimerOnAnswer and MaxAnswerSeconds come from the
	// card's deck options: show the time spent on the card, stop it when
	// the answer is shown, and the cap on the recorded time, at which the
	// timer changes colour.
	ShowTimer         bool `json:"showTimer"`
	StopTimerOnAnswer bool `json:"stopTimerOnAnswer"`
	MaxAnswerSeconds  int  `json:"maxAnswerSeconds"`
	// SecondsToShowQuestion, SecondsToShowAnswer, QuestionAction and
	// AnswerAction are the card's deck's auto advance options, used while
	// the user has auto advance on.
	SecondsToShowQuestion int    `json:"secondsToShowQuestion"`
	SecondsToShowAnswer   int    `json:"secondsToShowAnswer"`
	QuestionAction        string `json:"questionAction"`
	AnswerAction          string `json:"answerAction"`
	// TypeAnswer is set when the question has an input for typing the
	// answer (a <cf-typeans> placeholder). The answer is then shown with
	// Reveal, which compares the typed text.
	TypeAnswer bool `json:"typeAnswer"`
	// Flag is 0 for no flag, or 1 to 7.
	Flag int `json:"flag"`
	// Marked is set when the card's note has the tag "marked".
	Marked bool `json:"marked"`
	// DueEarly is when a learning card is due, when it is shown before then
	// because nothing else is due (learn ahead); nil otherwise.
	DueEarly *time.Time `json:"dueEarly,omitempty"`
}

// StudyState is what the study screen shows.
type StudyState struct {
	DeckID   int64  `json:"deckId"`
	DeckName string `json:"deckName"`
	// Card is the card to study, or nil when nothing is due now.
	Card    *StudyCard `json:"card"`
	Counts  Counts     `json:"counts"`
	CanUndo bool       `json:"canUndo"`
	// Studied and TimeSpent summarise the session.
	Studied   int           `json:"studied"`
	TimeSpent time.Duration `json:"timeSpent"`
	// NextDue is when something in the deck can next be studied, when
	// nothing is due now (nil when the deck has no cards to study at all).
	NextDue *time.Time `json:"nextDue"`
	// LearningDue is when the next learning card beyond the learn ahead
	// limit becomes due today, also while a card is shown (nil if none).
	// Learning cards within the limit are shown when nothing else is due.
	LearningDue *time.Time `json:"learningDue"`
	// CheckAt is when the study screen fetches the state again on its own
	// (nil if not today): while a card is shown, when the next learning card
	// becomes due, so the counts include it; otherwise, when the next
	// learning card comes within the learn ahead limit, so it can be shown.
	CheckAt *time.Time `json:"checkAt"`
	// Leech is set in the state Answer returns when that answer made the
	// card a leech; LeechSuspended when it also suspended the card.
	Leech          bool `json:"leech"`
	LeechSuspended bool `json:"leechSuspended"`
}

var errNoSession = core.NewError("noStudySession", "no study session; start one first")

// reset ends the session; another profile was opened.
func (s *StudyService) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = nil
}

// Start begins a study session for a deck and its subdecks.
func (s *StudyService) Start(deckID int64) (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	if _, err := c.Deck(core.ID(deckID)); err != nil {
		return StudyState{}, err
	}
	if err := c.SetCurrentDeck(core.ID(deckID)); err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = &session{deck: core.ID(deckID)}
	return s.state(c)
}

// Current returns the session's state; the card on screen stays the same,
// re-rendered (e.g. after its note was edited), while the queue still has it
// (e.g. not after the deck's options lowered the new card limit).
func (s *StudyService) Current() (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyState{}, errNoSession
	}
	return s.state(c)
}

// Answer rates the card on screen and moves to the next card.
func (s *StudyService) Answer(cardID int64, rating int, durationMs int64) (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyState{}, errNoSession
	}
	card, err := c.Card(core.ID(cardID))
	if err != nil {
		return StudyState{}, err
	}
	d := time.Duration(durationMs) * time.Millisecond
	entry, err := c.Answer(card.ID, scheduler.Rating(rating), d)
	if err != nil {
		return StudyState{}, err
	}
	s.s.answers = append(s.s.answers, answer{entry: entry.ID, card: card.ID, wasNew: card.State == scheduler.New,
		wasRev: card.State == scheduler.Review || card.Interday, duration: time.Duration(entry.DurationMs) * time.Millisecond})
	s.s.current = 0
	st, err := s.state(c)
	st.Leech, st.LeechSuspended = entry.Leech, entry.LeechSuspended
	return st, err
}

// Reveal renders the card on screen with its answer comparing typed, what
// the user typed into the question's input, with the expected text. Only
// cards with TypeAnswer need it; Answer of StudyCard is the answer for the
// others.
func (s *StudyService) Reveal(cardID int64, typed string) (StudyCard, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyCard{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyCard{}, errNoSession
	}
	card, err := c.Card(core.ID(cardID))
	if err != nil {
		return StudyCard{}, err
	}
	r, err := c.RenderCardTyped(card.ID, typed, s.a.renderEnv())
	if err != nil {
		return StudyCard{}, err
	}
	return studyCard(c, card, r)
}

// Undo undoes the session's most recent answer that was not undone yet and
// shows that card again.
func (s *StudyService) Undo() (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyState{}, errNoSession
	}
	live := s.s.view(c).answers
	if len(live) == 0 {
		return StudyState{}, core.NewError("nothingToUndo", "There is nothing to undo.")
	}
	cardID, err := c.UndoReview(live[len(live)-1].entry)
	if err != nil {
		return StudyState{}, err
	}
	s.s.current = cardID
	return s.state(c)
}

// PreviousCardID returns the card of the session's most recent answer that
// was not undone, or 0 when there is none.
func (s *StudyService) PreviousCardID() (int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return 0, errNoSession
	}
	live := s.s.view(c).answers
	if len(live) == 0 {
		return 0, nil
	}
	return int64(live[len(live)-1].card), nil
}

// Bury hides the card until tomorrow and moves on.
func (s *StudyService) Bury(cardID int64) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error { return c.BuryCards([]core.ID{core.ID(cardID)}) })
}

// Suspend suspends the card and moves on.
func (s *StudyService) Suspend(cardID int64) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error { return c.SetSuspended([]core.ID{core.ID(cardID)}, true) })
}

// BuryNote hides all cards of the card's note until tomorrow and moves on.
func (s *StudyService) BuryNote(cardID int64) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error {
		ids, err := siblingsOf(c, cardID)
		if err != nil {
			return err
		}
		return c.BuryCards(ids)
	})
}

// SuspendNote suspends all cards of the card's note and moves on.
func (s *StudyService) SuspendNote(cardID int64) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error {
		ids, err := siblingsOf(c, cardID)
		if err != nil {
			return err
		}
		return c.SetSuspended(ids, true)
	})
}

// DeleteNote deletes the card's note with all its cards and moves on.
func (s *StudyService) DeleteNote(cardID int64) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error {
		card, err := c.Card(core.ID(cardID))
		if err != nil {
			return err
		}
		return c.DeleteNotes([]core.ID{card.NoteID})
	})
}

// siblingsOf returns all cards of the card's note, the card included.
func siblingsOf(c *core.Collection, cardID int64) ([]core.ID, error) {
	card, err := c.Card(core.ID(cardID))
	if err != nil {
		return nil, err
	}
	var out []core.ID
	for _, sib := range c.CardsOfNote(card.NoteID) {
		out = append(out, sib.ID)
	}
	return out, nil
}

// SetDueDate sets the card's due date from spec (see BrowseService.SetDueDate)
// and moves on; the card stays in the session only while it is due today.
func (s *StudyService) SetDueDate(cardID int64, spec string) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error { return c.SetDueDate([]core.ID{core.ID(cardID)}, spec) })
}

// Reset makes the card New again (see BrowseService.ResetToNew) and moves
// on.
func (s *StudyService) Reset(cardID int64, opts ResetOptions) (StudyState, error) {
	return s.cardAction(func(c *core.Collection) error { return c.ResetCards([]core.ID{core.ID(cardID)}, opts.core()) })
}

// SetFlag sets flag 1 to 7 on the card, or clears it when the card already
// has it; flag 0 clears it. The card stays on screen.
func (s *StudyService) SetFlag(cardID int64, flag int) (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyState{}, errNoSession
	}
	if err := c.SetFlag([]core.ID{core.ID(cardID)}, flag); err != nil {
		return StudyState{}, err
	}
	return s.state(c)
}

// ToggleMark adds the tag "marked" to the card's note, or removes it when
// the note has it. The card stays on screen.
func (s *StudyService) ToggleMark(cardID int64) (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyState{}, errNoSession
	}
	card, err := c.Card(core.ID(cardID))
	if err != nil {
		return StudyState{}, err
	}
	if err := c.ToggleMark([]core.ID{card.NoteID}); err != nil {
		return StudyState{}, err
	}
	return s.state(c)
}

func (s *StudyService) cardAction(fn func(c *core.Collection) error) (StudyState, error) {
	c, err := s.a.collection()
	if err != nil {
		return StudyState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.s == nil {
		return StudyState{}, errNoSession
	}
	if err := fn(c); err != nil {
		return StudyState{}, err
	}
	s.s.current = 0
	return s.state(c)
}

// state picks the card to show and describes the session. Callers hold s.mu.
func (s *StudyService) state(c *core.Collection) (StudyState, error) {
	sess := s.s
	deck, err := c.Deck(sess.deck)
	if err != nil {
		return StudyState{}, err
	}
	q, err := c.Queue(sess.deck)
	if err != nil {
		return StudyState{}, err
	}
	v := sess.view(c)
	st := StudyState{DeckID: int64(deck.ID), DeckName: deck.Name, Counts: countsOf(q.Counts()),
		CanUndo: len(v.answers) > 0, Studied: v.studied, TimeSpent: v.spent}

	// The card on screen stays while the queue still has it: not once it is
	// suspended, buried or deleted, or the deck's options changed the limits.
	if sess.current != 0 && !inQueue(q, sess.current) {
		sess.current = 0
	}
	if sess.current == 0 {
		if id, ok := core.PickNext(q, v.newShown, v.reviewShown); ok {
			sess.current = id
		}
	}
	if !q.LearningLater.IsZero() {
		t := q.LearningLater
		st.LearningDue = &t
	}
	if sess.current == 0 {
		if st.LearningDue != nil {
			t := st.LearningDue.Add(-c.Settings().LearnAhead())
			st.CheckAt = &t
		}
		if next := c.NextDue(sess.deck); !next.IsZero() {
			st.NextDue = &next
		}
		return st, nil
	}
	card, err := c.Card(sess.current)
	if err != nil {
		return StudyState{}, err
	}
	r, err := c.RenderCard(card.ID, s.a.renderEnv())
	if err != nil {
		return StudyState{}, err
	}
	sc, err := studyCard(c, card, r)
	if err != nil {
		return StudyState{}, err
	}
	st.Card = &sc
	st.CheckAt = st.LearningDue
	if len(q.LearningAhead) > 0 {
		if ahead, err := c.Card(q.LearningAhead[0]); err == nil && (st.CheckAt == nil || ahead.Due.Before(*st.CheckAt)) {
			st.CheckAt = &ahead.Due
		}
	}
	return st, nil
}

// inQueue reports whether the queue has card id.
func inQueue(q core.Queue, id core.ID) bool {
	return slices.Contains(q.Learning, id) || slices.Contains(q.Review, id) ||
		slices.Contains(q.New, id) || slices.Contains(q.LearningAhead, id)
}

// studyCard describes a card rendered as r for the study screen.
func studyCard(c *core.Collection, card core.Card, r render.Card) (StudyCard, error) {
	previews, err := c.Previews(card.ID)
	if err != nil {
		return StudyCard{}, err
	}
	sc := StudyCard{CardID: int64(card.ID), NoteID: int64(card.NoteID), Question: r.Question, Answer: r.Answer,
		Background: r.Background, Color: r.Color, Kind: kindOf(card.State), Empty: c.IsEmptyCard(card.ID),
		QuestionMedia: r.QuestionMedia, AnswerMedia: r.AnswerMedia, QuestionTTS: r.QuestionTTS, AnswerTTS: r.AnswerTTS, Fonts: r.Fonts, Autoplay: true, ReplayQuestion: true,
		StopTimerOnAnswer: true, MaxAnswerSeconds: core.DefaultMaxAnswerSeconds,
		QuestionAction: core.QuestionShowAnswer, AnswerAction: core.AnswerBuryCard,
		TypeAnswer: r.TypeAnswer, Flag: card.Flag, Marked: c.IsMarked(card.NoteID)}
	if o, err := c.DeckOptions(card.HomeDeck()); err == nil {
		sc.Autoplay, sc.ReplayQuestion = !o.DisableAutoplay, !o.SkipQuestionOnReplay
		sc.ShowTimer, sc.StopTimerOnAnswer, sc.MaxAnswerSeconds = o.ShowTimer, o.StopTimerOnAnswer, o.MaxAnswerSeconds
		sc.SecondsToShowQuestion, sc.SecondsToShowAnswer = o.SecondsToShowQuestion, o.SecondsToShowAnswer
		sc.QuestionAction, sc.AnswerAction = o.QuestionAction, o.AnswerAction
	}
	for _, p := range previews {
		sc.Previews = append(sc.Previews, Preview{Rating: int(p.Rating), Interval: p.Interval, Ends: p.Label == core.PreviewEndLabel})
	}
	if due, ok := learningDue(card); ok && due.After(c.Now()) {
		sc.DueEarly = &due
	}
	return sc, nil
}

// learningDue is when a card waiting like a learning card is due: a card
// answered Again or Hard in preview mode, or a learning card within the
// day. Other cards are not waiting.
func learningDue(card core.Card) (time.Time, bool) {
	switch {
	case !card.PreviewDue.IsZero():
		return card.PreviewDue, true
	case (card.State == scheduler.Learning || card.State == scheduler.Relearning) && !card.Interday:
		return card.Due, true
	}
	return time.Time{}, false
}

func kindOf(s scheduler.State) string {
	switch s {
	case scheduler.New:
		return "new"
	case scheduler.Review:
		return "review"
	}
	return "learning"
}
