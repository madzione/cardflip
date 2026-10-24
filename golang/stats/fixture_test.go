package stats

import (
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// fixtureNow is the fixture's clock: Tuesday 2026-10-13 12:00, rollover at
// 04:00, so the study day began at 04:00 today.
var fixtureNow = time.Date(2026, 10, 13, 12, 0, 0, 0, time.Local)

func at(month time.Month, day, hour, minute int) time.Time {
	return time.Date(2026, month, day, hour, minute, 0, 0, time.Local)
}

// fixture is a collection with a hand-written review log:
//
//	Spanish        S1  review, ivl 30, due tomorrow
//	                   Sep 13 10:00  new        Good  5 s   learning
//	                   Sep 13 10:10  learning   Good  4 s   learning  (ivl 3)
//	                   Sep 20 09:00  manual (rating 0, skipped)
//	                   Oct 13 02:30  review     Good  6 s   young     (study day Oct 12)
//	Spanish::Verbs V1  relearning
//	                   Oct 13 09:00  review     Again 8 s   mature    (ivl 25)
//	                   Oct 13 09:10  relearning Good  3 s   relearning
//	French         F1  review, ivl 5, suspended
//	                   Oct 12 20:00  new        Good  2 s   learning
//	                   Oct 13 08:00  review     Good  1 s   filtered  (cram: no ivl after)
//	French         D1  deleted after one answer
//	                   Oct 01 10:00  new        Again 9 s   learning
//	French         N1  new
type fixture struct {
	c                  *core.Collection
	spanish, verbs, fr core.ID
	s1, v1, f1, n1     core.ID
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	clk := fixtureNow
	c, _, err := core.Open(storage.NewMemory(), core.Options{Now: func() time.Time { return clk }})
	if err != nil {
		t.Fatal(err)
	}
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	mem := func(s, d float64) *scheduler.Memory { return &scheduler.Memory{Stability: s, Difficulty: d} }
	note := func(guid, deck string, ic core.ImportCard) core.ImportNote {
		ic.DeckName = deck
		return core.ImportNote{TypeKey: "basic", GUID: guid, DeckName: deck, Fields: []string{guid, "a"}, Cards: []core.ImportCard{ic}}
	}
	b := &core.ImportBatch{
		NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []core.ImportNote{
			note("s1", "Spanish", core.ImportCard{
				State: scheduler.CardState{State: scheduler.Review, Due: at(10, 14, 2, 30), ScheduledDays: 30,
					Memory: mem(30, 5), Reps: 3, LastReview: at(10, 13, 2, 30)},
				History: []core.ImportReview{
					{Time: at(9, 13, 10, 0), Rating: scheduler.Good, StateBefore: scheduler.New, DurationMs: 5000},
					{Time: at(9, 13, 10, 10), Rating: scheduler.Good, StateBefore: scheduler.Learning, IvlAfter: 3, DurationMs: 4000},
					{Time: at(9, 20, 9, 0), Rating: 0, StateBefore: scheduler.Review, IvlBefore: 3, IvlAfter: 3},
					{Time: at(10, 13, 2, 30), Rating: scheduler.Good, StateBefore: scheduler.Review, IvlBefore: 3, IvlAfter: 30, DurationMs: 6000},
				}}),
			note("v1", "Spanish::Verbs", core.ImportCard{
				State: scheduler.CardState{State: scheduler.Relearning, Due: at(10, 13, 9, 20), Memory: mem(4, 7),
					Reps: 8, Lapses: 1, LastReview: at(10, 13, 9, 10)},
				History: []core.ImportReview{
					{Time: at(10, 13, 9, 0), Rating: scheduler.Again, StateBefore: scheduler.Review, IvlBefore: 25, DurationMs: 8000},
					{Time: at(10, 13, 9, 10), Rating: scheduler.Good, StateBefore: scheduler.Relearning, DurationMs: 3000},
				}}),
			note("f1", "French", core.ImportCard{
				Suspended: true,
				State: scheduler.CardState{State: scheduler.Review, Due: at(10, 17, 20, 0), ScheduledDays: 5,
					Memory: mem(5, 3), Reps: 2, LastReview: at(10, 13, 8, 0)},
				History: []core.ImportReview{
					{Time: at(10, 12, 20, 0), Rating: scheduler.Good, StateBefore: scheduler.New, IvlAfter: 5, DurationMs: 2000},
					{Time: at(10, 13, 8, 0), Rating: scheduler.Good, StateBefore: scheduler.Review, IvlBefore: 5, DurationMs: 1000},
				}}),
			note("d1", "French", core.ImportCard{
				State: scheduler.CardState{State: scheduler.Learning, Due: at(10, 1, 10, 1), Memory: mem(0.2, 8),
					Reps: 1, LastReview: at(10, 1, 10, 0)},
				History: []core.ImportReview{
					{Time: at(10, 1, 10, 0), Rating: scheduler.Again, StateBefore: scheduler.New, DurationMs: 9000},
				}}),
			note("n1", "French", core.ImportCard{}),
		},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	f := &fixture{c: c}
	for _, d := range c.Decks() {
		switch d.Name {
		case "Spanish":
			f.spanish = d.ID
		case "Spanish::Verbs":
			f.verbs = d.ID
		case "French":
			f.fr = d.ID
		}
	}
	var deleted core.ID
	for _, card := range c.AllCards() {
		n, err := c.Note(card.NoteID)
		if err != nil {
			t.Fatal(err)
		}
		switch n.GUID {
		case "s1":
			f.s1 = card.ID
		case "v1":
			f.v1 = card.ID
		case "f1":
			f.f1 = card.ID
		case "n1":
			f.n1 = card.ID
		case "d1":
			deleted = n.ID
		}
	}
	if err := c.DeleteNotes([]core.ID{deleted}); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) load(t *testing.T, scope Scope) *Data {
	t.Helper()
	var l Loader
	d, err := l.Load(f.c, scope)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// day returns the study day of t in the fixture (rollover 04:00).
func day(t time.Time) int { return scheduler.DayNumber(t, core.DefaultRolloverHour) }
