package main

import (
	"fmt"
	"testing"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
)

// Results on a Ryzen 5 5600X (2026-10-25), native, median of 10 runs:
//
//	card totals only (before deck progress)  16.2 ms
//	with deck progress                       16.7 ms

// BenchmarkDeckTree is DecksService.Tree, run on every deck list reload,
// on 100,000 cards spread over 20 decks with two subdecks each, a fifth
// each new, learning, young, mature and suspended.
func BenchmarkDeckTree(b *testing.B) {
	s, _ := opened(b)
	c := must(s.app.a.collection())
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	due := s.clock.AddDate(0, 0, 1)
	batch := &core.ImportBatch{NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}}}
	for i := range 100_000 {
		deck := fmt.Sprintf("Deck %d", i%20)
		if sub := i / 20 % 3; sub > 0 {
			deck += fmt.Sprintf("::Sub %d", sub)
		}
		ic := core.ImportCard{DeckName: deck}
		switch i / 60 % 5 {
		case 1:
			ic.State = scheduler.CardState{State: scheduler.Learning, Due: due}
		case 2:
			ic.State = scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: 5}
		case 3:
			ic.State = scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: 30}
		case 4:
			ic.State = scheduler.CardState{State: scheduler.Review, Due: due, ScheduledDays: 30}
			ic.Suspended = true
		}
		batch.Notes = append(batch.Notes, core.ImportNote{TypeKey: "basic", GUID: fmt.Sprint(i), DeckName: deck,
			Fields: []string{fmt.Sprint("q", i), "a"}, Cards: []core.ImportCard{ic}})
	}
	if _, err := c.Import(batch, nil); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		must(s.decks.Tree())
	}
}
