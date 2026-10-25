package main

import (
	"slices"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/stats"
)

// ProgressService reports the collection's streak and today's progress
// toward the daily goal, for the deck list and the study summary. It is
// cheap to ask after every answer: the study days of the whole review log
// are read once and then only extended by new answers.
type ProgressService struct {
	a  *app
	mu sync.Mutex
	// days are the study days of col, read at review generation gen with
	// rollover in loc, and extended by the answers up to last.
	col      *core.Collection
	gen      uint64
	rollover int
	loc      *time.Location
	days     []int
	last     core.ID
	// result is the last result, for collection revision rev on study day
	// today.
	result Progress
	rev    uint64
	today  int
}

// Progress is the streak and today's progress toward the daily goal.
type Progress struct {
	// Current is the current streak in days: the consecutive study days
	// ending today, or yesterday when today has no answers yet, which
	// AtRisk then reports (with a streak of at least one day).
	Current      int  `json:"current"`
	AtRisk       bool `json:"atRisk"`
	TodayStudied bool `json:"todayStudied"`
	Longest      int  `json:"longest"`
	// TodayCards and TodayMinutes are today's answers and the whole
	// minutes spent on them.
	TodayCards   int `json:"todayCards"`
	TodayMinutes int `json:"todayMinutes"`
	// GoalKind is "cards" or "minutes", or empty without a daily goal;
	// Goal is its number of cards or minutes, GoalReached whether today's
	// progress meets it.
	GoalKind    string `json:"goalKind"`
	Goal        int    `json:"goal"`
	GoalReached bool   `json:"goalReached"`
}

// Today returns the streak and today's progress. The first call reads
// the whole review log.
func (s *ProgressService) Today() (Progress, error) {
	c, err := s.a.collection()
	if err != nil {
		return Progress{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st := c.Settings()
	now := c.Now()
	today := scheduler.DayNumber(now, st.RolloverHour)
	rev := c.Revision()
	fresh := s.col == c && s.rollover == st.RolloverHour && s.loc == now.Location()
	if fresh && s.rev == rev && s.today == today {
		return s.result, nil
	}
	answers, gen := c.RecentAnswers(s.last, scheduler.DayStart(now, st.RolloverHour))
	if !fresh || gen != s.gen {
		sd, err := c.StudyDays(st.RolloverHour, now.Location())
		if err != nil {
			return Progress{}, err
		}
		s.col, s.gen, s.rollover, s.loc, s.days, s.last = c, sd.Gen, st.RolloverHour, now.Location(), sd.Days, sd.Last
		answers, gen = c.RecentAnswers(s.last, scheduler.DayStart(now, st.RolloverHour))
		if gen != s.gen {
			s.col = nil // changed in between: read again next time
		}
	}
	var cards int
	var ms int64
	after := s.last
	for _, e := range answers {
		day := scheduler.DayNumber(e.Time.In(s.loc), s.rollover)
		if e.ID > after {
			if i, found := slices.BinarySearch(s.days, day); !found {
				s.days = slices.Insert(s.days, i, day)
			}
			s.last = max(s.last, e.ID)
		}
		if day == today {
			cards++
			ms += e.DurationMs
		}
	}
	streak := stats.Streaks(s.days, today)
	goal := goalOf(st)
	s.result = Progress{Current: streak.Current, AtRisk: streak.AtRisk, TodayStudied: streak.TodayStudied,
		Longest: streak.Longest, TodayCards: cards, TodayMinutes: int(ms / time.Minute.Milliseconds()),
		GoalKind: goal.Kind, Goal: goal.Value, GoalReached: goal.Met(cards, ms)}
	s.rev, s.today = rev, today
	return s.result, nil
}

// reset forgets the study days, when the profile is closed.
func (s *ProgressService) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.col, s.days = nil, nil
}
