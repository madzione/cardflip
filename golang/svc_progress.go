package main

import (
	"slices"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/stats"
)

// ProgressService reports the collection's streak, today's progress
// toward the daily goal and the achievements, for the deck list and the
// study summary. It is cheap to ask after every answer: the study days of
// the whole review log are read once and then only extended by new
// answers.
type ProgressService struct {
	a  *app
	mu sync.Mutex
	// days are the study days of col, read at review generation gen with
	// rollover in loc, and extended by the answers up to last. answers[i]
	// and ms[i] are the answers on days[i] and the milliseconds spent on
	// them; mature are the sorted days cards became mature.
	col      *core.Collection
	gen      uint64
	rollover int
	loc      *time.Location
	days     []int
	answers  []int
	ms       []int64
	mature   []int
	last     core.ID
	// result and achievements are the last results, for collection
	// revision rev on study day today.
	result       Progress
	achievements []stats.AchievementStatus
	rev          uint64
	today        int
	// recap is the last weekly recap, for recapKey.
	recap    *Recap
	recapKey recapKey
}

// recapKey is what a weekly recap depends on: the study days read at
// generation gen with rollover in loc, the week start and the first day of
// the week.
type recapKey struct {
	col          *core.Collection
	gen          uint64
	rollover     int
	loc          *time.Location
	start        int
	firstWeekday int
}

// Progress is the streak and today's progress toward the daily goal.
type Progress struct {
	// Current is the current streak in days: the consecutive study days
	// ending today, or yesterday when today has no answers yet, which
	// AtRisk then reports (with a streak of at least one day). Rest days
	// per week let it go on over days without study.
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
	// NewAchievements are the earned achievement tiers this device has not
	// shown yet, in the order of the achievements; MarkSeen records them
	// as shown.
	NewAchievements []NewAchievement `json:"newAchievements"`
}

// NewAchievement is an earned achievement tier not shown yet.
type NewAchievement struct {
	// ID is the achievement's, Threshold the tier's and TierName its name
	// (see stats.AchievementStatus).
	ID        string `json:"id"`
	Threshold int    `json:"threshold"`
	TierName  string `json:"tierName"`
	// Key is what MarkSeen takes.
	Key string `json:"key"`
}

// Today returns the streak, today's progress and the achievements not
// shown yet. The first call reads the whole review log.
func (s *ProgressService) Today() (Progress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Progress{}, err
	}
	return s.result, nil
}

// Achievements returns every achievement with its progress and the dates
// its tiers were earned.
func (s *ProgressService) Achievements() ([]stats.AchievementStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	return s.achievements, nil
}

// MarkSeen records achievement tiers (NewAchievement.Key) as shown on this
// device.
func (s *ProgressService) MarkSeen(keys []string) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	st := c.Settings()
	seen := slices.Clone(st.SeenAchievements)
	if seen == nil {
		seen = []string{}
	}
	for _, k := range keys {
		if !slices.Contains(seen, k) {
			seen = append(seen, k)
		}
	}
	st.SeenAchievements = seen
	return c.UpdateSettings(st)
}

// Recap is the weekly recap: the previous week and the week before it.
type Recap struct {
	// WeekStart is the first study day (scheduler.DayNumber) of the current
	// week; AppService.DismissRecap takes it.
	WeekStart int       `json:"weekStart"`
	Last      RecapWeek `json:"last"`
	Before    RecapWeek `json:"before"`
}

// RecapWeek sums up one week of study.
type RecapWeek struct {
	StudyDays int `json:"studyDays"`
	Answers   int `json:"answers"`
	// Minutes are the whole minutes spent answering.
	Minutes int `json:"minutes"`
	// Retention is the true retention of the week's young and mature
	// reviews, as in the statistics; nil without such reviews.
	Retention *float64 `json:"retention"`
}

// WeeklyRecap returns the previous week's study days, answers, minutes and
// true retention and those of the week before, or nil when the previous
// week has no study days. Weeks start on firstWeekday (0 Sunday, 1 Monday,
// up to 6 Saturday, as in Dart's MaterialLocalizations) and follow the
// study day boundary. The result is kept until the week or the review log
// (other than by answering) changes.
func (s *ProgressService) WeeklyRecap(firstWeekday int) (*Recap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return nil, err
	}
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	weekday := int(time.Unix(int64(s.today)*86400, 0).UTC().Weekday())
	start := s.today - (weekday-firstWeekday+7)%7
	key := recapKey{col: s.col, gen: s.gen, rollover: s.rollover, loc: s.loc, start: start, firstWeekday: firstWeekday}
	if s.col != nil && key == s.recapKey {
		return s.recap, nil
	}
	recap := &Recap{WeekStart: start, Last: s.week(start - 7), Before: s.week(start - 14)}
	if recap.Last.StudyDays == 0 {
		recap = nil
	} else {
		d, err := stats.LoadDays(c, start-14)
		if err != nil {
			return nil, err
		}
		recap.Last.Retention = stats.RetentionOfDays(d, start-7, start-1).Total.Retention
		recap.Before.Retention = stats.RetentionOfDays(d, start-14, start-8).Total.Retention
	}
	s.recap, s.recapKey = recap, key
	return recap, nil
}

// week sums the study days from first on for 7 days. Callers hold s.mu.
func (s *ProgressService) week(first int) RecapWeek {
	var w RecapWeek
	var ms int64
	i, _ := slices.BinarySearch(s.days, first)
	for ; i < len(s.days) && s.days[i] < first+7; i++ {
		w.StudyDays++
		w.Answers += s.answers[i]
		ms += s.ms[i]
	}
	w.Minutes = int(ms / time.Minute.Milliseconds())
	return w
}

// refresh brings s.result and s.achievements up to date. Callers hold
// s.mu.
func (s *ProgressService) refresh() error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	st := c.Settings()
	now := c.Now()
	today := scheduler.DayNumber(now, st.RolloverHour)
	rev := c.Revision()
	fresh := s.col == c && s.rollover == st.RolloverHour && s.loc == now.Location()
	if fresh && s.rev == rev && s.today == today {
		return nil
	}
	answers, gen := c.RecentAnswers(s.last, scheduler.DayStart(now, st.RolloverHour))
	if !fresh || gen != s.gen {
		sd, err := c.StudyDays(st.RolloverHour, now.Location())
		if err != nil {
			return err
		}
		s.col, s.gen, s.rollover, s.loc, s.last = c, sd.Gen, st.RolloverHour, now.Location(), sd.Last
		s.days, s.answers, s.ms, s.mature = sd.Days, sd.Answers, sd.Ms, sd.Mature
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
			i, found := slices.BinarySearch(s.days, day)
			if !found {
				s.days = slices.Insert(s.days, i, day)
				s.answers = slices.Insert(s.answers, i, 0)
				s.ms = slices.Insert(s.ms, i, 0)
			}
			s.answers[i]++
			s.ms[i] += e.DurationMs
			if e.FirstMature {
				j, _ := slices.BinarySearch(s.mature, day)
				s.mature = slices.Insert(s.mature, j, day)
			}
			s.last = max(s.last, e.ID)
		}
		if day == today {
			cards++
			ms += e.DurationMs
		}
	}
	streak := stats.Streaks(s.days, today, st.RestDaysPerWeek)
	goal := goalOf(st)
	s.achievements = stats.EvalAchievements(stats.History{Days: s.days, Answers: s.answers, Ms: s.ms, Mature: s.mature,
		Goal: goal, Rest: st.RestDaysPerWeek})
	var unseen []NewAchievement
	var earned []string
	for _, a := range s.achievements {
		for i := range a.Earned {
			k := a.Key(i)
			earned = append(earned, k)
			if !slices.Contains(st.SeenAchievements, k) {
				unseen = append(unseen, NewAchievement{ID: a.ID, Threshold: a.Tiers[i], TierName: a.TierNames[i], Key: k})
			}
		}
	}
	if st.SeenAchievements == nil {
		// The first evaluation, also after an upgrade: what was earned
		// before is no news.
		if err := s.MarkSeen(earned); err != nil {
			return err
		}
		unseen, rev = nil, c.Revision()
	}
	s.result = Progress{Current: streak.Current, AtRisk: streak.AtRisk, TodayStudied: streak.TodayStudied,
		Longest: streak.Longest, TodayCards: cards, TodayMinutes: int(ms / time.Minute.Milliseconds()),
		GoalKind: goal.Kind, Goal: goal.Value, GoalReached: goal.Met(cards, ms), NewAchievements: unseen}
	s.rev, s.today = rev, today
	return nil
}

// reset forgets the study days, when the profile is closed.
func (s *ProgressService) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.col, s.days, s.answers, s.ms, s.mature = nil, nil, nil, nil, nil
	s.recap, s.recapKey = nil, recapKey{}
}
