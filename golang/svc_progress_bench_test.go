package main

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
)

// progressCollection opens services on a collection of 20,000 review cards
// with 10 answers each over two years, 200,000 in total, and 100 new cards
// to answer.
func progressCollection(b *testing.B) *services {
	b.Helper()
	s, _ := opened(b)
	c := must(s.app.a.collection())
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	rng := rand.New(rand.NewPCG(1, 2))
	batch := &core.ImportBatch{NoteTypes: []core.ImportNoteType{{Key: "basic", ExistingID: basic}}}
	for i := range 20_100 {
		ic := core.ImportCard{DeckName: "Default"}
		if i >= 100 {
			ic.State = scheduler.CardState{State: scheduler.Review, Due: s.clock.AddDate(0, 0, 30), ScheduledDays: 30, Reps: 10}
			for range 10 {
				ic.History = append(ic.History, core.ImportReview{
					Time:   s.clock.Add(-time.Duration(1+rng.IntN(730*24*60)) * time.Minute),
					Rating: scheduler.Rating(1 + rng.IntN(4)), StateBefore: scheduler.Review,
					IvlBefore: 1 + rng.IntN(100), IvlAfter: 1 + rng.IntN(100), DurationMs: int64(rng.IntN(20000)),
				})
			}
		}
		batch.Notes = append(batch.Notes, core.ImportNote{TypeKey: "basic", GUID: fmt.Sprint(i), DeckName: "Default",
			Fields: []string{fmt.Sprint("q", i), "a"}, Cards: []core.ImportCard{ic}})
	}
	if _, err := c.Import(batch, nil); err != nil {
		b.Fatal(err)
	}
	return s
}

// Results on a Ryzen 5 5600X (2026-10-25), native and WebAssembly (Node),
// with the achievements' history aggregates:
//
//	first call       450 ms   1.8 s    (decoding the review log)
//	study days       2.9 ms   6.2 ms   (after undo, imports and sync)
//	after an answer  0.07 ms  0.25 ms  (with the achievements)
//	memory kept      12.6 MB  12.6 MB  (9.2 MB without the aggregates)
//
// WebAssembly: PATH=$(go env GOROOT)/lib/wasm:$PATH GOOS=js GOARCH=wasm
// go test -run x -bench Progress -benchtime 5x

// BenchmarkProgressFirstCall is ProgressService.Today on a collection just
// opened: it reads the whole review log of 200,000 answers.
func BenchmarkProgressFirstCall(b *testing.B) {
	s := progressCollection(b)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		must(s.app.Reload())
		s.progress = &ProgressService{a: s.app.a}
		b.StartTimer()
		must(s.progress.Today())
	}
}

// BenchmarkProgressStudyDays is the part of the first call after the
// review index is loaded (also after undo, imports and sync): mapping
// 200,000 answers to study days.
func BenchmarkProgressStudyDays(b *testing.B) {
	s := progressCollection(b)
	c := must(s.app.a.collection())
	must(c.StudyDays(core.DefaultRolloverHour, time.Local)) // loads the review index
	for b.Loop() {
		must(c.StudyDays(core.DefaultRolloverHour, time.Local))
	}
}

// BenchmarkProgressAfterAnswer is ProgressService.Today after one answer,
// as the deck list and the summary ask after every answer. It must stay
// within a few milliseconds.
func BenchmarkProgressAfterAnswer(b *testing.B) {
	s := progressCollection(b)
	must(s.progress.Today())
	st := must(s.study.Start(must(s.notes.Defaults(0)).DeckID))
	n := 0
	for b.Loop() {
		n++
		b.StopTimer()
		if st.Card == nil {
			check(b, s.app.SetDailyGoal(core.GoalCards, 1+st.Studied)) // a change without an answer
		} else {
			st = must(s.study.Answer(st.Card.CardID, int(scheduler.Good), 1000))
		}
		b.StartTimer()
		must(s.progress.Today())
	}
	if per := b.Elapsed() / time.Duration(n); per > 5*time.Millisecond {
		b.Errorf("Today after an answer took %v, want under 5ms", per)
	}
}

// BenchmarkProgressMemory reports the heap that the first call keeps: the
// review index of 200,000 answers with their progress history, and the
// progress service's study days.
func BenchmarkProgressMemory(b *testing.B) {
	s := progressCollection(b)
	var before, after runtime.MemStats
	for range b.N {
		b.StopTimer()
		must(s.app.Reload())
		s.progress = &ProgressService{a: s.app.a}
		runtime.GC()
		runtime.ReadMemStats(&before)
		b.StartTimer()
		must(s.progress.Today())
		b.StopTimer()
		runtime.GC()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(int64(after.HeapAlloc)-int64(before.HeapAlloc))/1e6, "MB")
		b.StartTimer()
	}
}
