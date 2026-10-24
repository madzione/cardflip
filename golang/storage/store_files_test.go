//go:build !js

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Kill during save with real files: the journal is cut at every byte of its
// last record, as a torn write would leave it.
func TestKillDuringJournalAppendWithFiles(t *testing.T) {
	dir := t.TempDir()
	b, _ := OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	m, _ := openModel(t, b)
	m.commit("a")
	m.commit("b")
	before, _ := os.ReadFile(filepath.Join(dir, "g1", "journal"))
	m.answer(7)
	b.Close()
	full, _ := os.ReadFile(filepath.Join(dir, "g1", "journal"))
	segment := filepath.Join(dir, "rl1", "2026-10")

	for cut := len(before); cut < len(full); cut++ {
		// The segment is written only after the journal record is synced, so
		// a torn journal record means the segment write never started.
		os.Remove(segment)
		os.WriteFile(filepath.Join(dir, "g1", "journal"), full[:cut], 0o644)
		b, _ := OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
		m, err := openModel(t, b)
		if err != nil {
			t.Fatalf("cut at %d: %v", cut, err)
		}
		if strings.Join(m.ops, ",") != "a,b" || len(m.revs) != 0 {
			t.Fatalf("cut at %d: ops %v revs %v", cut, m.ops, m.revs)
		}
		// The app keeps working after recovery.
		if err := m.commit("c"); err != nil {
			t.Fatal(err)
		}
		b.Close()
		b, _ = OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
		m, _ = openModel(t, b)
		if strings.Join(m.ops, ",") != "a,b,c" {
			t.Fatalf("cut at %d: after new commit ops = %v", cut, m.ops)
		}
		b.Close()
	}
}

// Opening a 50,000-card collection with 500,000 review log entries takes
// less than 3 seconds (persistence spec, "Scale").
func TestOpenLargeCollectionUnder3s(t *testing.T) {
	if testing.Short() {
		t.Skip("large collection test")
	}
	dir := t.TempDir()
	b, _ := OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))

	type card struct {
		ID, NoteID, DeckID int64
		Ord, State, Reps   int
		Due, Last          time.Time
		Stability, Diff    float64
		IntervalDays       int
		Lapses             int
	}
	type note struct {
		ID, TypeID int64
		GUID       string
		Fields     []string
		Tags       []string
	}
	type snapshot struct {
		Notes []note
		Cards []card
	}
	var snap snapshot
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 50000; i++ {
		id := int64(1_700_000_000_000 + i)
		if i%2 == 0 {
			snap.Notes = append(snap.Notes, note{ID: id, TypeID: 1, GUID: fmt.Sprintf("g%08d", i),
				Fields: []string{fmt.Sprintf("<b>front %d</b> some longer question text", i), fmt.Sprintf("back %d answer text", i)},
				Tags:   []string{"spanish", "verb"}})
		}
		snap.Cards = append(snap.Cards, card{ID: id, NoteID: snap.Notes[len(snap.Notes)-1].ID, DeckID: 1, Ord: i % 2,
			State: 2, Reps: 10, Due: now.AddDate(0, 0, i%90), Last: now.AddDate(0, 0, -i%30),
			Stability: 12.5, Diff: 5.25, IntervalDays: i % 90})
	}
	raw, _ := json.Marshal(snap)

	// 500,000 entries over 25 months, written as framed segment files.
	type entry struct {
		ID, CardID          int64
		Time                time.Time
		Rating, State       int
		IvlBefore, IvlAfter int
		Stability, Diff     float64
		DurationMs          int64
	}
	for mth := 0; mth < 25; mth++ {
		month := time.Date(2024, time.Month(10+mth), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		var seg []byte
		for i := 0; i < 20000; i++ {
			e, _ := json.Marshal(entry{ID: int64(mth*20000 + i), CardID: int64(i), Time: now, Rating: 3, State: 2,
				IvlBefore: 5, IvlAfter: 12, Stability: 12.5, Diff: 5.25, DurationMs: 6500})
			rec, _ := json.Marshal(segmentRecord{ID: int64(mth*20000 + i), Entry: e})
			seg = append(seg, frame(rec)...)
		}
		if err := b.Put(segKey(1, month), seg); err != nil {
			t.Fatal(err)
		}
	}
	b.Put("g1/snapshot", raw)
	b.Put("current", []byte(`{"v":1,"gen":1,"rl":1,"media":1}`))
	for i := 0; i < 500; i++ {
		b.Append("g1/journal", []byte(`{"op":"x"}`))
	}
	b.Close()
	runtime.GC()

	start := time.Now()
	b, _ = OpenFiles(dir, filepath.Join(t.TempDir(), "backups"), filepath.Join(t.TempDir(), "trash"))
	st, loaded, err := Open(b)
	if err != nil {
		t.Fatal(err)
	}
	var got snapshot
	if err := json.Unmarshal(loaded.Snapshot, &got); err != nil {
		t.Fatal(err)
	}
	// Current and previous month, as the scheduler loads them.
	var revlog []entry
	for _, month := range []string{"2026-09", "2026-10"} {
		entries, err := st.ReadRevlog(month)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			var en entry
			json.Unmarshal(e, &en)
			revlog = append(revlog, en)
		}
	}
	elapsed := time.Since(start)
	t.Logf("opened %d cards, %d notes, %d journal ops, %d revlog entries loaded, snapshot %.1f MB, in %v",
		len(got.Cards), len(got.Notes), len(loaded.Ops), len(revlog), float64(len(raw))/1e6, elapsed)
	if len(got.Cards) != 50000 || len(revlog) != 40000 {
		t.Fatalf("loaded %d cards, %d entries", len(got.Cards), len(revlog))
	}
	if elapsed > 3*time.Second {
		t.Fatalf("open took %v, want < 3s", elapsed)
	}
}
