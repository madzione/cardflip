package main

import (
	"fmt"

	"cardflip/golang/stats"
)

// StatsService computes the statistics screen's sections. Each section is
// its own call, so the screen shows sections as they are ready; they share
// one read of the review log. Every call takes the scope and the period
// ("month", "3months", "year" or "all"); sections that are not time-based
// ignore the period.
type StatsService struct {
	a      *app
	loader stats.Loader
}

func (s *StatsService) data(scope stats.Scope, period string) (*stats.Data, error) {
	if !stats.ValidPeriod(period) {
		return nil, fmt.Errorf("stats: unknown period %q", period)
	}
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	return s.loader.Load(c, scope)
}

func section[T any](s *StatsService, scope stats.Scope, period string, fn func(*stats.Data) T) (T, error) {
	d, err := s.data(scope, period)
	if err != nil {
		var zero T
		return zero, err
	}
	return fn(d), nil
}

// Today summarises the current study day.
func (s *StatsService) Today(scope stats.Scope, period string) (stats.TodayStats, error) {
	return section(s, scope, period, stats.TodaySummary)
}

// FutureDue is the forecast of due cards over the period.
func (s *StatsService) FutureDue(scope stats.Scope, period string) (stats.Forecast, error) {
	return section(s, scope, period, func(d *stats.Data) stats.Forecast { return stats.FutureDue(d, period) })
}

// Calendar is the number of answers per day over the last year.
func (s *StatsService) Calendar(scope stats.Scope, period string) (stats.Calendar, error) {
	return section(s, scope, period, stats.ReviewCalendar)
}

// Reviews is the number of answers per day, week or month of the period.
func (s *StatsService) Reviews(scope stats.Scope, period string) (stats.ReviewHistory, error) {
	return section(s, scope, period, func(d *stats.Data) stats.ReviewHistory { return stats.Reviews(d, period) })
}

// ReviewTime is the time spent answering per day, week or month of the
// period, in milliseconds.
func (s *StatsService) ReviewTime(scope stats.Scope, period string) (stats.ReviewHistory, error) {
	return section(s, scope, period, func(d *stats.Data) stats.ReviewHistory { return stats.ReviewTime(d, period) })
}

// Counts are the cards by state.
func (s *StatsService) Counts(scope stats.Scope, period string) (stats.StateCounts, error) {
	return section(s, scope, period, stats.CardCounts)
}

// Intervals is the distribution of review intervals, in days.
func (s *StatsService) Intervals(scope stats.Scope, period string) (stats.Histogram, error) {
	return section(s, scope, period, stats.Intervals)
}

// Stability is the distribution of FSRS stability, in days.
func (s *StatsService) Stability(scope stats.Scope, period string) (stats.Histogram, error) {
	return section(s, scope, period, stats.Stability)
}

// Difficulty is the distribution of FSRS difficulty, from 0 to 1.
func (s *StatsService) Difficulty(scope stats.Scope, period string) (stats.Histogram, error) {
	return section(s, scope, period, stats.Difficulty)
}

// Retrievability is the distribution of the current probability of
// recall; its average is the average retrievability and its sum the
// estimated knowledge.
func (s *StatsService) Retrievability(scope stats.Scope, period string) (stats.Histogram, error) {
	return section(s, scope, period, stats.Retrievability)
}

// Hourly is the answers of the period by hour of the day, 0 to 23.
func (s *StatsService) Hourly(scope stats.Scope, period string) ([]stats.HourBucket, error) {
	return section(s, scope, period, func(d *stats.Data) []stats.HourBucket { return stats.Hourly(d, period) })
}

// Buttons are the answer button counts of the period.
func (s *StatsService) Buttons(scope stats.Scope, period string) (stats.ButtonStats, error) {
	return section(s, scope, period, func(d *stats.Data) stats.ButtonStats { return stats.AnswerButtons(d, period) })
}

// TrueRetention is the true retention table.
func (s *StatsService) TrueRetention(scope stats.Scope, period string) (stats.RetentionTable, error) {
	return section(s, scope, period, stats.TrueRetention)
}
