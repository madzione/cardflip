package scheduler

import "time"

// DayStart returns when the study day containing t began: the most recent
// rollover hour, in t's location. A review at 02:30 with rollover 04:00
// belongs to the previous calendar day.
func DayStart(t time.Time, rolloverHour int) time.Time {
	y, m, d := t.Date()
	start := time.Date(y, m, d, rolloverHour, 0, 0, 0, t.Location())
	if t.Before(start) {
		start = time.Date(y, m, d-1, rolloverHour, 0, 0, 0, t.Location())
	}
	return start
}

// NextDayStart returns when the study day containing t ends.
func NextDayStart(t time.Time, rolloverHour int) time.Time {
	y, m, d := DayStart(t, rolloverHour).Date()
	return time.Date(y, m, d+1, rolloverHour, 0, 0, 0, t.Location())
}

// DayNumber identifies the study day containing t; consecutive days have
// consecutive numbers, independent of daylight saving changes.
func DayNumber(t time.Time, rolloverHour int) int {
	y, m, d := DayStart(t, rolloverHour).Date()
	return int(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400)
}
