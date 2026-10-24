package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// The tests model the collection as a list of ops: the snapshot is the JSON
// list of ops applied so far, and loading means snapshot + journal.

type model struct {
	t    *testing.T
	st   *Store
	ops  []string
	revs map[int64]bool // review entries that exist (not undone)
	dels map[int64]bool // deletion log entries
}

func openModel(t *testing.T, b Backend) (*model, error) {
	t.Helper()
	st, loaded, err := Open(b)
	if err != nil {
		return nil, err
	}
	m := &model{t: t, st: st, revs: map[int64]bool{}, dels: map[int64]bool{}}
	if loaded.Fresh {
		if err := st.Init([]byte("[]")); err != nil {
			return nil, err
		}
		return m, nil
	}
	if err := json.Unmarshal(loaded.Snapshot, &m.ops); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	for _, op := range loaded.Ops {
		var s string
		if err := json.Unmarshal(op, &s); err != nil {
			t.Fatalf("op: %v", err)
		}
		m.ops = append(m.ops, s)
	}
	months, err := st.RevlogMonths()
	if err != nil {
		return nil, err
	}
	for _, month := range months {
		entries, err := st.ReadRevlog(month)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			var id int64
			json.Unmarshal(e, &id)
			m.revs[id] = true
		}
	}
	dels, err := st.ReadDeletions()
	if err != nil {
		return nil, err
	}
	for _, e := range dels {
		var id int64
		json.Unmarshal(e, &id)
		if m.dels[id] {
			t.Fatalf("deletion log entry %d is listed twice", id)
		}
		m.dels[id] = true
	}
	return m, nil
}

func (m *model) commit(op string, rev ...RevlogWrite) error {
	return m.commitDel(op, rev, nil)
}

func (m *model) commitDel(op string, rev []RevlogWrite, del []Deletion) error {
	raw, _ := json.Marshal(op)
	if err := m.st.Commit(raw, rev, del); err != nil {
		return err
	}
	m.ops = append(m.ops, op)
	for _, d := range del {
		m.dels[d.ID] = true
	}
	for _, w := range rev {
		m.revs[w.ID] = w.Entry != nil
		if w.Entry == nil {
			delete(m.revs, w.ID)
		}
	}
	return nil
}

func (m *model) answer(id int64) error {
	return m.commit(fmt.Sprintf("answer-%d", id), RevlogWrite{Month: "2026-10", ID: id, Entry: json.RawMessage(fmt.Sprint(id))})
}

func (m *model) undo(id int64) error {
	return m.commit(fmt.Sprintf("undo-%d", id), RevlogWrite{Month: "2026-10", ID: id})
}

// del logs deletion log entry id.
func (m *model) del(id int64) error {
	return m.commitDel(fmt.Sprintf("delete-%d", id), nil, []Deletion{{ID: id, Entry: json.RawMessage(fmt.Sprint(id))}})
}

func (m *model) snapshot() []byte {
	b, _ := json.Marshal(m.ops)
	return b
}

// compactTime is when model.compact compacts.
var compactTime = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

func (m *model) compact(keep ...string) error { return m.compactAt(compactTime, keep...) }

func (m *model) compactAt(now time.Time, keep ...string) error {
	return m.st.Compact(m.snapshot(), keep, nil, now)
}

// expectedRevs derives which review entries must exist from the ops.
func expectedRevs(ops []string) map[int64]bool {
	revs := map[int64]bool{}
	for _, op := range ops {
		var id int64
		if _, err := fmt.Sscanf(op, "answer-%d", &id); err == nil {
			revs[id] = true
		} else if _, err := fmt.Sscanf(op, "undo-%d", &id); err == nil {
			delete(revs, id)
		}
	}
	return revs
}

// expectedDels derives which deletion log entries must exist from the ops.
func expectedDels(ops []string) map[int64]bool {
	dels := map[int64]bool{}
	for _, op := range ops {
		var id int64
		if _, err := fmt.Sscanf(op, "delete-%d", &id); err == nil {
			dels[id] = true
		}
	}
	return dels
}

func TestFreshInitCommitReopen(t *testing.T) {
	store := NewMemoryStore()
	m, err := openModel(t, store.Open())
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{"a", "b", "c"} {
		if err := m.commit(op); err != nil {
			t.Fatal(err)
		}
	}
	m2, err := openModel(t, store.Open())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(m2.ops, ",") != "a,b,c" || m2.st.JournalEntries() != 3 {
		t.Fatalf("ops = %v, journal = %d", m2.ops, m2.st.JournalEntries())
	}
}

func TestCompactionStartsNewGeneration(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.commit("a")
	m.commit("b")
	if err := m.compact(); err != nil {
		t.Fatal(err)
	}
	m.commit("c")

	keys, _ := store.Open().List("")
	if strings.Join(keys, ",") != "current,g2/journal,g2/snapshot" {
		t.Fatalf("keys = %v", keys)
	}
	m2, _ := openModel(t, store.Open())
	if strings.Join(m2.ops, ",") != "a,b,c" || m2.st.JournalEntries() != 1 {
		t.Fatalf("ops = %v, journal = %d", m2.ops, m2.st.JournalEntries())
	}
}

func TestCompactionThresholds(t *testing.T) {
	m, _ := openModel(t, NewMemory())
	for i := 0; i < CompactEntries-1; i++ {
		m.commit("x")
	}
	if m.st.NeedsCompaction() {
		t.Fatal("compaction needed before the entry threshold")
	}
	m.commit("x")
	if !m.st.NeedsCompaction() {
		t.Fatal("compaction not needed at the entry threshold")
	}
	m.compact()
	if m.st.NeedsCompaction() {
		t.Fatal("compaction still needed after compacting")
	}
	big := strings.Repeat("y", 1<<20)
	for i := 0; i < 3; i++ {
		m.commit(big)
	}
	if m.st.NeedsCompaction() {
		t.Fatal("compaction needed below 4 MB")
	}
	m.commit(big)
	if !m.st.NeedsCompaction() {
		t.Fatal("compaction not needed at 4 MB")
	}
}

func TestRevlogSegmentsAndUndo(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.answer(1)
	m.answer(2)
	m.answer(3)
	m.undo(2)
	m.commit("other", RevlogWrite{Month: "2026-09", ID: 9, Entry: json.RawMessage("9")})

	entries, err := m.st.ReadRevlog("2026-10")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(rawStrings(entries)) != "[1 3]" {
		t.Fatalf("2026-10 = %v", rawStrings(entries))
	}
	months, _ := m.st.RevlogMonths()
	if strings.Join(months, ",") != "2026-09,2026-10" {
		t.Fatalf("months = %v", months)
	}
	// Compaction keeps the review log: it is not part of the generation.
	m.compact()
	m2, _ := openModel(t, store.Open())
	if !reflect.DeepEqual(m2.revs, map[int64]bool{1: true, 3: true, 9: true}) {
		t.Fatalf("revs after reopen = %v", m2.revs)
	}
}

// batching counts batched segment writes, and can fail them.
type batching struct {
	*Memory
	batches int
	fail    bool
}

func (b *batching) AppendAll(key string, records [][]byte) error {
	if strings.HasPrefix(key, "rl") {
		if b.fail {
			return errDead
		}
		b.batches++
	}
	return b.Memory.AppendAll(key, records)
}

// The review log writes of one commit go to each segment in one write, in
// order; when that fails after the journal record, the next open repairs
// all of them, once.
func TestRevlogWritesBatchedPerSegment(t *testing.T) {
	var writes []RevlogWrite
	want := map[string][]string{}
	for i := range 300 {
		month := fmt.Sprintf("2026-0%d", 7+i%3)
		writes = append(writes, RevlogWrite{Month: month, ID: int64(1000 + i), Entry: json.RawMessage(fmt.Sprint(1000 + i))})
		want[month] = append(want[month], fmt.Sprint(1000+i))
	}
	check := func(st *Store, label string) {
		t.Helper()
		for month, w := range want {
			entries, err := st.ReadRevlog(month)
			if err != nil {
				t.Fatal(err)
			}
			if got := rawStrings(entries); !reflect.DeepEqual(got, w) {
				t.Fatalf("%s: %s = %d entries, want %d", label, month, len(got), len(w))
			}
		}
	}

	store := NewMemoryStore()
	b := &batching{Memory: store.Open()}
	m, _ := openModel(t, b)
	if err := m.commit("import", writes...); err != nil {
		t.Fatal(err)
	}
	if b.batches != 3 {
		t.Fatalf("segment writes = %d, want 3", b.batches)
	}
	check(m.st, "committed")
	// Undoing deletes them in one write per segment too.
	var undo []RevlogWrite
	for _, w := range writes[:30] {
		undo = append(undo, RevlogWrite{Month: w.Month, ID: w.ID})
	}
	if err := m.commit("undo", undo...); err != nil {
		t.Fatal(err)
	}
	if b.batches != 6 {
		t.Fatalf("segment writes = %d, want 6", b.batches)
	}
	if err := m.commit("redo", writes[:30]...); err != nil {
		t.Fatal(err)
	}

	store = NewMemoryStore()
	b = &batching{Memory: store.Open(), fail: true}
	m, _ = openModel(t, b)
	if err := m.commit("import", writes...); err != nil {
		t.Fatalf("commit must succeed once the journal is durable: %v", err)
	}
	for range 2 {
		m2, err := openModel(t, store.Open())
		if err != nil {
			t.Fatal(err)
		}
		check(m2.st, "repaired")
	}
	for month := range want {
		if recs, _ := store.Open().Records(segKey(1, month)); len(recs) != 100 {
			t.Fatalf("%s has %d records after two repairs, want 100", month, len(recs))
		}
	}
}

func rawStrings(rs []json.RawMessage) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = string(r)
	}
	return out
}

// failing wraps a backend so that writes fail once a condition is met: the
// process "dies", nothing more is written.
type failing struct {
	Backend
	writes int
	limit  int               // fail the write with this index and all after it (-1: never)
	match  func(string) bool // only count/fail writes to matching keys (nil: all)
	dead   bool
}

var errDead = errors.New("simulated crash")

func (f *failing) write(key string) error {
	if f.match != nil && !f.match(key) {
		return nil
	}
	if f.dead || (f.limit >= 0 && f.writes >= f.limit) {
		f.dead = true
		return errDead
	}
	f.writes++
	return nil
}

func (f *failing) Put(k string, v []byte) error {
	if err := f.write(k); err != nil {
		return err
	}
	return f.Backend.Put(k, v)
}

func (f *failing) Append(k string, r []byte) error {
	if err := f.write(k); err != nil {
		return err
	}
	return f.Backend.Append(k, r)
}

func (f *failing) Delete(k string) error {
	if err := f.write(k); err != nil {
		return err
	}
	return f.Backend.Delete(k)
}

// A failed segment write after a durable journal record is repaired at the
// next open, for answers and for undos.
func TestRevlogRepairedFromJournal(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.answer(1)

	f := &failing{Backend: store.Open(), limit: 0, match: func(k string) bool { return strings.HasPrefix(k, "rl") }}
	m.st.b = f
	if err := m.answer(2); err != nil {
		t.Fatalf("commit must succeed once the journal is durable: %v", err)
	}
	if err := m.undo(1); err != nil {
		t.Fatal(err)
	}
	if err := m.compact(); err == nil {
		t.Fatal("compaction must refuse while segment writes are pending")
	}

	m2, err := openModel(t, store.Open())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m2.revs, map[int64]bool{2: true}) {
		t.Fatalf("revs = %v, want only 2", m2.revs)
	}
	// Repair is idempotent: opening again changes nothing.
	m3, _ := openModel(t, store.Open())
	if !reflect.DeepEqual(m3.revs, map[int64]bool{2: true}) {
		t.Fatalf("revs on second open = %v", m3.revs)
	}
	if err := m3.compact(); err != nil {
		t.Fatal(err)
	}
}

// scenario runs a fixed sequence of saves. It stops at the first error,
// like an app process that was killed, and returns the ops whose save
// returned successfully.
func scenario(t *testing.T, b Backend) (committed []string, attempted []string) {
	m, err := openModel(t, b)
	if err != nil {
		return nil, nil
	}
	committed = append(committed, m.ops...)
	attempted = append(attempted, m.ops...)
	steps := []func() (string, error){
		func() (string, error) { return "answer-1", m.answer(1) },
		func() (string, error) { return "b", m.commit("b") },
		func() (string, error) {
			if err := m.st.PutMedia("cat.jpg", []byte("meow")); err != nil {
				return "", err
			}
			return "media", m.commit("media")
		},
		func() (string, error) { return "", m.compact("cat.jpg") },
		func() (string, error) { return "delete-5", m.del(5) },
		func() (string, error) { return "answer-2", m.answer(2) },
		func() (string, error) { return "undo-1", m.undo(1) },
		func() (string, error) { return "", m.compact("cat.jpg") },
		func() (string, error) { return "c", m.commit("c") },
		func() (string, error) { return "delete-6", m.del(6) },
	}
	for _, step := range steps {
		op, err := step()
		if op != "" {
			attempted = append(attempted, op)
		}
		if err != nil {
			return committed, attempted
		}
		if op != "" {
			committed = append(committed, op)
		}
	}
	return committed, attempted
}

func checkRecovered(t *testing.T, b Backend, committed, attempted []string, label string) {
	t.Helper()
	m, err := openModel(t, b)
	if err != nil {
		t.Fatalf("%s: reopen failed: %v", label, err)
	}
	// Every successful save is there; at most the save in progress is lost
	// or kept, and nothing else appears.
	ok := reflect.DeepEqual(m.ops, committed) || reflect.DeepEqual(m.ops, attempted)
	if !ok && !(len(committed) == 0 && len(m.ops) == 0) {
		t.Fatalf("%s: recovered ops %v, want %v or %v", label, m.ops, committed, attempted)
	}
	if want := expectedRevs(m.ops); !reflect.DeepEqual(m.revs, want) {
		t.Fatalf("%s: review log %v does not match ops %v", label, m.revs, m.ops)
	}
	if want := expectedDels(m.ops); !reflect.DeepEqual(m.dels, want) {
		t.Fatalf("%s: deletion log %v does not match ops %v", label, m.dels, m.ops)
	}
	for _, op := range m.ops {
		if op == "media" && !m.st.HasMedia("cat.jpg") {
			t.Fatalf("%s: media referenced by a committed op is missing", label)
		}
	}
}

// Kill during save: the process dies at every possible write, then the app
// starts again on what was stored.
func TestKillDuringSaveAtEveryWrite(t *testing.T) {
	total := func() int {
		f := &failing{Backend: NewMemory(), limit: -1}
		scenario(t, f)
		return f.writes
	}()
	if total < 15 {
		t.Fatalf("scenario only does %d writes", total)
	}
	for limit := 0; limit <= total; limit++ {
		store := NewMemoryStore()
		committed, attempted := scenario(t, &failing{Backend: store.Open(), limit: limit})
		checkRecovered(t, store.Open(), committed, attempted, fmt.Sprintf("crash at write %d", limit))
	}
}

func fillOld(t *testing.T, store *MemoryStore) *model {
	m, _ := openModel(t, store.Open())
	m.st.PutMedia("old.png", []byte("old"))
	m.commit("old-media")
	m.answer(1)
	m.del(1)
	return m
}

// A replacement (restore) that fails at any point leaves the old collection,
// review log and media exactly as they were.
func TestFailedReplacementKeepsOldGeneration(t *testing.T) {
	// The replacement writes media, a segment record, a deletion log record,
	// the snapshot and the pointer: a crash before any of these five keeps
	// the old state.
	for limit := 0; limit < 5; limit++ {
		store := NewMemoryStore()
		fillOld(t, store)
		before := dump(store)

		st, _, _ := Open(&failing{Backend: store.Open(), limit: limit})
		err := func() error {
			r, err := st.Replace()
			if err != nil {
				return err
			}
			if err := r.PutMedia("new.png", []byte("new")); err != nil {
				return err
			}
			if err := r.AppendRevlog("2026-01", 50, json.RawMessage("50")); err != nil {
				return err
			}
			if err := r.AppendDeletion(60, json.RawMessage("60")); err != nil {
				return err
			}
			return r.Commit([]byte(`["new"]`))
		}()
		if err == nil {
			t.Fatalf("limit %d: replacement unexpectedly succeeded", limit)
		}
		m, err := openModel(t, store.Open())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(m.ops, ",") != "old-media,answer-1,delete-1" || !m.st.HasMedia("old.png") || m.st.HasMedia("new.png") {
			t.Fatalf("limit %d: state changed: ops %v", limit, m.ops)
		}
		if !reflect.DeepEqual(m.revs, map[int64]bool{1: true}) || !reflect.DeepEqual(m.dels, map[int64]bool{1: true}) {
			t.Fatalf("limit %d: revs %v dels %v", limit, m.revs, m.dels)
		}
		if after := dump(store); !reflect.DeepEqual(before, after) {
			t.Fatalf("limit %d: leftovers not cleaned:\nbefore %v\nafter  %v", limit, keysOf(before), keysOf(after))
		}
	}
}

func TestReplacementCommitAndAbort(t *testing.T) {
	store := NewMemoryStore()
	m := fillOld(t, store)

	r, _ := m.st.Replace()
	r.PutMedia("x.png", []byte("x"))
	r.Abort()
	if m.st.HasMedia("x.png") || !m.st.HasMedia("old.png") {
		t.Fatal("abort changed the live media")
	}

	r, _ = m.st.Replace()
	r.PutMedia("new.png", []byte("new"))
	r.AppendRevlog("2025-12", 77, json.RawMessage("77"))
	r.AppendDeletion(78, json.RawMessage("78"))
	if err := r.Commit([]byte(`["restored"]`)); err != nil {
		t.Fatal(err)
	}
	m2, _ := openModel(t, store.Open())
	if strings.Join(m2.ops, ",") != "restored" || !reflect.DeepEqual(m2.revs, map[int64]bool{77: true}) ||
		!reflect.DeepEqual(m2.dels, map[int64]bool{78: true}) {
		t.Fatalf("ops %v revs %v dels %v", m2.ops, m2.revs, m2.dels)
	}
	if m2.st.HasMedia("old.png") || !m2.st.HasMedia("new.png") {
		t.Fatal("media not replaced")
	}
	keys, _ := store.Open().List("")
	for _, k := range keys {
		if strings.HasPrefix(k, "g1/") || strings.HasPrefix(k, "rl1/") || strings.HasPrefix(k, "m1/") || strings.HasPrefix(k, "dl1/") {
			t.Fatalf("old set not removed: %v", keys)
		}
	}
}

func TestReplacementKeepingMedia(t *testing.T) {
	store := NewMemoryStore()
	m := fillOld(t, store)
	r, err := m.st.ReplaceKeepingMedia()
	if err != nil {
		t.Fatal(err)
	}
	r.AppendRevlog("2025-12", 77, json.RawMessage("77"))
	if err := r.Commit([]byte(`["restored"]`)); err != nil {
		t.Fatal(err)
	}
	m2, _ := openModel(t, store.Open())
	if strings.Join(m2.ops, ",") != "restored" || !reflect.DeepEqual(m2.revs, map[int64]bool{77: true}) {
		t.Fatalf("ops %v revs %v", m2.ops, m2.revs)
	}
	if !m2.st.HasMedia("old.png") {
		t.Fatal("live media not kept")
	}
}

func TestOrphanMediaCollected(t *testing.T) {
	m, _ := openModel(t, NewMemory())
	m.st.PutMedia("keep.png", []byte("k"))
	m.st.PutMedia("Keep.png", []byte("K"))
	m.st.PutMedia("orphan.png", []byte("o"))
	if err := m.compact("keep.png", "Keep.png"); err != nil {
		t.Fatal(err)
	}
	var names []string
	m.st.EachMedia(func(name string, data []byte) error {
		names = append(names, name+"="+string(data))
		return nil
	})
	if fmt.Sprint(names) != fmt.Sprint([]string{"Keep.png=K", "keep.png=k"}) && fmt.Sprint(names) != fmt.Sprint([]string{"keep.png=k", "Keep.png=K"}) {
		t.Fatalf("media after GC = %v", names)
	}
	if _, err := m.st.GetMedia("orphan.png"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("orphan: %v", err)
	}
}

func trashNames(t *testing.T, st *Store) string {
	t.Helper()
	list, err := st.TrashedMedia()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range list {
		names = append(names, f.Name)
	}
	return strings.Join(names, ",")
}

// Compaction moves unreferenced media to the trash instead of deleting them
// (spec: media trash), and deletes files trashed more than 30 days ago.
func TestCompactionTrashesOrphanMedia(t *testing.T) {
	m, _ := openModel(t, NewMemory())
	m.st.PutMedia("keep.png", []byte("k"))
	m.st.PutMedia("orphan.png", []byte("o"))
	if err := m.compact("keep.png"); err != nil {
		t.Fatal(err)
	}
	list, err := m.st.TrashedMedia()
	if err != nil || len(list) != 1 || list[0].Name != "orphan.png" || list[0].Size != 1 || !list[0].Trashed.Equal(compactTime) {
		t.Fatalf("trash = %v, %v", list, err)
	}

	m.st.PutMedia("later.png", []byte("l"))
	day := 24 * time.Hour
	if err := m.compactAt(compactTime.Add(30*day), "keep.png"); err != nil {
		t.Fatal(err)
	}
	if got := trashNames(t, m.st); got != "later.png,orphan.png" {
		t.Fatalf("trash after 30 days = %q, want both files", got)
	}
	if err := m.compactAt(compactTime.Add(30*day+time.Minute), "keep.png"); err != nil {
		t.Fatal(err)
	}
	if got := trashNames(t, m.st); got != "later.png" {
		t.Fatalf("trash after 30 days and a minute = %q, want only later.png", got)
	}
	if !m.st.HasMedia("keep.png") {
		t.Fatal("referenced media removed")
	}
}

// Trashed files come back with Restore (spec: restore deleted files). A
// file whose name was taken by different content stays in the trash.
func TestTrashAndRestoreMedia(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	for name, data := range map[string]string{"a.png": "a", "b.mp3": "b", "c.jpg": "c", "same.png": "s"} {
		m.st.PutMedia(name, []byte(data))
	}
	if err := m.st.TrashMedia([]string{"a.png", "b.mp3", "c.jpg", "same.png", "never-stored.png"}, compactTime); err != nil {
		t.Fatal(err)
	}
	if got := trashNames(t, m.st); got != "a.png,b.mp3,c.jpg,same.png" {
		t.Fatalf("trash = %q", got)
	}
	if m.st.HasMedia("a.png") || m.st.HasMedia("b.mp3") {
		t.Fatal("trashed files still in the media")
	}
	m.st.PutMedia("c.jpg", []byte("a different c"))
	m.st.PutMedia("same.png", []byte("s"))

	// The trash survives a restart.
	m2, err := openModel(t, store.Open())
	if err != nil {
		t.Fatal(err)
	}
	names, err := m2.st.RestoreTrash()
	if sort.Strings(names); err != nil || strings.Join(names, " ") != "a.png b.mp3" {
		t.Fatalf("RestoreTrash = %q, %v; want a.png and b.mp3", names, err)
	}
	for name, want := range map[string]string{"a.png": "a", "b.mp3": "b", "c.jpg": "a different c", "same.png": "s"} {
		if data, err := m2.st.GetMedia(name); err != nil || string(data) != want {
			t.Fatalf("%s = %q, %v; want %q", name, data, err, want)
		}
	}
	if got := trashNames(t, m2.st); got != "c.jpg" {
		t.Fatalf("trash after restore = %q, want the conflicting c.jpg", got)
	}
}

func TestEmptyAndPurgeTrash(t *testing.T) {
	m, _ := openModel(t, NewMemory())
	m.st.PutMedia("old.png", []byte("o"))
	m.st.PutMedia("new.png", []byte("n"))
	m.st.TrashMedia([]string{"old.png"}, compactTime)
	m.st.TrashMedia([]string{"new.png"}, compactTime.Add(time.Hour))
	if err := m.st.PurgeTrash(compactTime.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got := trashNames(t, m.st); got != "new.png" {
		t.Fatalf("trash after purge = %q, want new.png", got)
	}
	if err := m.st.EmptyTrash(); err != nil {
		t.Fatal(err)
	}
	if got := trashNames(t, m.st); got != "" {
		t.Fatalf("trash after empty = %q", got)
	}
}

// A failed trash write leaves the file in the media.
func TestTrashFailureKeepsMedia(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.st.PutMedia("orphan.png", []byte("o"))
	store.FailPut = func(key string) bool { return strings.HasPrefix(key, "trash/") }
	if err := m.compact(); err == nil {
		t.Fatal("compaction succeeded although the trash could not be written")
	}
	if err := m.st.TrashMedia([]string{"orphan.png"}, compactTime); err == nil {
		t.Fatal("TrashMedia succeeded although the trash could not be written")
	}
	if !m.st.HasMedia("orphan.png") {
		t.Fatal("file lost")
	}
}

func dump(store *MemoryStore) map[string]string {
	out := map[string]string{}
	b := store.Open()
	keys, _ := b.List("")
	for _, k := range keys {
		v, err := b.Get(k)
		if err != nil {
			recs, _ := b.Records(k)
			v = bytes.Join(recs, []byte("|"))
		}
		out[k] = string(v)
	}
	return out
}

func keysOf(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

// Unreadable data is reported as corrupt and left exactly as it was.
func TestCorruptDataReportedAndUntouched(t *testing.T) {
	cases := map[string]func(b Backend){
		"garbage pointer":  func(b Backend) { b.Put("current", []byte("{nope")) },
		"missing snapshot": func(b Backend) { b.Delete("g1/snapshot") },
		"bad journal record": func(b Backend) {
			b.Append("g1/journal", []byte("not json"))
		},
		"bad segment record": func(b Backend) {
			b.Append("rl1/2026-10", []byte("not json"))
		},
		"bad deletion log record": func(b Backend) {
			b.Append("dl1/log", []byte("not json"))
			// A journal entry with a deletion makes open read the log.
			b.Append("g1/journal", []byte(`{"op":"\"delete-2\"","del":[{"id":2,"e":2}]}`))
		},
		"newer format": func(b Backend) { b.Put("current", []byte(`{"v":2,"gen":1,"rl":1,"media":1}`)) },
	}
	for name, damage := range cases {
		store := NewMemoryStore()
		fillOld(t, store)
		// An orphan set that a successful open would clean up.
		store.Open().Put("g9/snapshot", []byte("leftover"))
		damage(store.Open())
		before := dump(store)

		_, _, err := Open(store.Open())
		if !errors.Is(err, ErrCorrupt) {
			t.Fatalf("%s: err = %v, want ErrCorrupt", name, err)
		}
		if after := dump(store); !reflect.DeepEqual(before, after) {
			t.Fatalf("%s: data was modified", name)
		}
	}
}

func TestResetStartsFresh(t *testing.T) {
	store := NewMemoryStore()
	fillOld(t, store)
	store.Open().Put("current", []byte("{nope"))
	st := &Store{b: store.Open()}
	if err := st.Reset(); err != nil {
		t.Fatal(err)
	}
	_, loaded, err := Open(store.Open())
	if err != nil || !loaded.Fresh {
		t.Fatalf("after reset: %v fresh=%v", err, loaded != nil && loaded.Fresh)
	}
}

// The deletion log is append-only and survives compaction, which only
// replaces the generation.
func TestDeletionLogAppendAndCompact(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.del(1)
	m.del(2)
	m.commitDel("both", nil, []Deletion{{ID: 3, Entry: json.RawMessage("3")}, {ID: 4, Entry: json.RawMessage("4")}})
	entries, err := m.st.ReadDeletions()
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(rawStrings(entries)) != "[1 2 3 4]" {
		t.Fatalf("deletions = %v", rawStrings(entries))
	}
	if err := m.compact(); err != nil {
		t.Fatal(err)
	}
	m2, _ := openModel(t, store.Open())
	if !reflect.DeepEqual(m2.dels, map[int64]bool{1: true, 2: true, 3: true, 4: true}) {
		t.Fatalf("deletions after compaction = %v", m2.dels)
	}
}

// A failed deletion log write after a durable journal record is repaired at
// the next open, once.
func TestDeletionLogRepairedFromJournal(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.del(1)

	m.st.b = &failing{Backend: store.Open(), limit: 0, match: func(k string) bool { return strings.HasPrefix(k, "dl") }}
	if err := m.del(2); err != nil {
		t.Fatalf("commit must succeed once the journal is durable: %v", err)
	}
	if err := m.compact(); err == nil {
		t.Fatal("compaction must refuse while deletion log writes are pending")
	}
	for range 2 {
		m2, err := openModel(t, store.Open())
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(m2.dels, map[int64]bool{1: true, 2: true}) {
			t.Fatalf("deletions = %v", m2.dels)
		}
	}
}

// Clearing empties the log for good: the journal that recorded the cleared
// entries is gone, so the next open does not repair them back. A crash
// during the clear leaves the log either full or empty.
func TestClearDeletions(t *testing.T) {
	fill := func(store *MemoryStore) *model {
		m, _ := openModel(t, store.Open())
		m.del(1)
		m.commit("a")
		m.del(2)
		return m
	}
	store := NewMemoryStore()
	m := fill(store)
	if err := m.st.ClearDeletions(m.snapshot()); err != nil {
		t.Fatal(err)
	}
	m.del(3)
	m2, _ := openModel(t, store.Open())
	if strings.Join(m2.ops, ",") != "delete-1,a,delete-2,delete-3" || !reflect.DeepEqual(m2.dels, map[int64]bool{3: true}) {
		t.Fatalf("ops %v dels %v", m2.ops, m2.dels)
	}

	for limit := 0; limit < 4; limit++ {
		store := NewMemoryStore()
		m := fill(store)
		m.st.b = &failing{Backend: store.Open(), limit: limit}
		err := m.st.ClearDeletions(m.snapshot())
		m2, oerr := openModel(t, store.Open())
		if oerr != nil {
			t.Fatal(oerr)
		}
		if strings.Join(m2.ops, ",") != "delete-1,a,delete-2" {
			t.Fatalf("limit %d: ops %v", limit, m2.ops)
		}
		full := reflect.DeepEqual(m2.dels, map[int64]bool{1: true, 2: true})
		if !full && len(m2.dels) != 0 || err != nil && !full {
			t.Fatalf("limit %d (err %v): dels %v", limit, err, m2.dels)
		}
	}
}

// Collections created before the deletion log have no dl set in the pointer;
// they start an empty log.
func TestDeletionLogInOlderCollection(t *testing.T) {
	store := NewMemoryStore()
	m, _ := openModel(t, store.Open())
	m.commit("a")
	store.Open().Put("current", []byte(`{"v":1,"gen":1,"rl":1,"media":1}`))
	m2, err := openModel(t, store.Open())
	if err != nil {
		t.Fatal(err)
	}
	if err := m2.del(1); err != nil {
		t.Fatal(err)
	}
	m3, _ := openModel(t, store.Open())
	if strings.Join(m3.ops, ",") != "a,delete-1" || !reflect.DeepEqual(m3.dels, map[int64]bool{1: true}) {
		t.Fatalf("ops %v dels %v", m3.ops, m3.dels)
	}
}
