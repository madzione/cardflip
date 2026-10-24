package core

import (
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// withoutSyncMeta returns a copy of s without the sync metadata.
func withoutSyncMeta(s *state) *state {
	cp := *s
	cp.Epoch, cp.Seq, cp.Meta, cp.Tombs, cp.RevSeq, cp.Slots = "", 0, nil, nil, nil, nil
	return &cp
}

func metaOf(t *testing.T, c *Collection, kind syncKind, id ID) objMeta {
	t.Helper()
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.meta[metaKey{kind, id}]
	if !ok {
		t.Fatalf("no metadata for kind %d, ID %d", kind, id)
	}
	return m
}

func tombOf(c *Collection, kind syncKind, id ID) (objMeta, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.tombs[metaKey{kind, id}]
	return m, ok
}

// syncState returns copies of the sync metadata.
func syncState(c *Collection) (meta, tombs map[metaKey]objMeta, seq int64, epoch string, revSeq map[string]int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return maps.Clone(c.meta), maps.Clone(c.tombs), c.seq, c.epoch, maps.Clone(c.revSeq)
}

func sameSyncState(t *testing.T, what string, got, want *Collection) {
	t.Helper()
	gm, gt, gs, ge, gr := syncState(got)
	wm, wt, ws, we, wr := syncState(want)
	if !maps.Equal(gm, wm) {
		t.Errorf("%s: registers differ: %d vs %d entries", what, len(gm), len(wm))
		for k, v := range wm {
			if gm[k] != v {
				t.Errorf("  %v: got %+v, want %+v", k, gm[k], v)
			}
		}
	}
	if !maps.Equal(gt, wt) {
		t.Errorf("%s: tombstones differ:\n got %v\nwant %v", what, gt, wt)
	}
	if gs != ws || ge != we || !maps.Equal(gr, wr) {
		t.Errorf("%s: seq %d, epoch %s, revSeq %v; want %d, %s, %v", what, gs, ge, gr, ws, we, wr)
	}
}

func TestSyncStampsIncrease(t *testing.T) {
	c, _, clk := testCollection(t)
	a := addNote(t, c, "Basic", "a", "1")
	ma := metaOf(t, c, kindNote, a)
	b := addNote(t, c, "Basic", "b", "2")
	mb := metaOf(t, c, kindNote, b)
	if mb.stamp <= ma.stamp || mb.seq <= ma.seq {
		t.Fatalf("second note %+v not after first %+v", mb, ma)
	}
	// The wall clock going backwards does not make stamps go backwards.
	clk.advance(-time.Hour)
	if err := c.UpdateNote(a, []string{"a", "one"}, nil); err != nil {
		t.Fatal(err)
	}
	ma2 := metaOf(t, c, kindNote, a)
	if ma2.stamp <= mb.stamp || ma2.seq <= mb.seq {
		t.Fatalf("edit %+v not after %+v", ma2, mb)
	}
	// Unchanged objects keep their metadata.
	if got := metaOf(t, c, kindNote, b); got != mb {
		t.Fatalf("unchanged note restamped: %+v, was %+v", got, mb)
	}
}

func TestSyncCardRegisters(t *testing.T) {
	c, _, _ := testCollection(t)
	id := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	sched, place := metaOf(t, c, kindCardSched, id), metaOf(t, c, kindCardPlace, id)
	if _, err := c.Answer(id, scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	if got := metaOf(t, c, kindCardSched, id); got.stamp <= sched.stamp {
		t.Fatalf("answer did not stamp scheduling: %+v", got)
	}
	if got := metaOf(t, c, kindCardPlace, id); got != place {
		t.Fatalf("answer stamped placement: %+v, was %+v", got, place)
	}
	sched = metaOf(t, c, kindCardSched, id)
	if err := c.SetFlag([]ID{id}, 1); err != nil {
		t.Fatal(err)
	}
	if got := metaOf(t, c, kindCardSched, id); got != sched {
		t.Fatalf("flag stamped scheduling: %+v, was %+v", got, sched)
	}
	if got := metaOf(t, c, kindCardPlace, id); got.stamp <= place.stamp {
		t.Fatalf("flag did not stamp placement: %+v", got)
	}
	// The answer's review log entry is stored with its sequence number.
	recs, err := c.store.ReadRevlogRecords(monthKey(c.Now()))
	if err != nil || len(recs) != 1 || recs[0].Seq == 0 || recs[0].Seq >= place.seq+10 {
		t.Fatalf("records = %+v, %v", recs, err)
	}
	if c.revSeq[monthKey(c.Now())] != recs[0].Seq {
		t.Fatalf("revSeq = %v, record seq %d", c.revSeq, recs[0].Seq)
	}
}

// Every Card field is in exactly one of the two registers, or is the ID or
// the modification time, which merging takes as the later of both.
func TestSyncCardFieldsClassified(t *testing.T) {
	sched := []string{"NewPos", "OriginalPos", "BuriedUntil", "PreviewDue", "CardState"}
	place := []string{"NoteID", "Ord", "DeckID", "HomeDeckID", "FilteredPos", "Flag", "Suspended"}
	other := []string{"ID", "Modified"}
	typ := reflect.TypeFor[Card]()
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		n := 0
		for _, list := range [][]string{sched, place, other} {
			if slices.Contains(list, name) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("Card field %s is classified %d times", name, n)
		}
	}
	// And the comparisons look at them.
	base := &Card{ID: 1, NoteID: 2, DeckID: 3}
	for _, edit := range []func(*Card){
		func(c *Card) { c.NewPos = 5 }, func(c *Card) { p := int64(1); c.OriginalPos = &p },
		func(c *Card) { c.BuriedUntil = time.Unix(1, 0) }, func(c *Card) { c.PreviewDue = time.Unix(1, 0) },
		func(c *Card) { c.State = scheduler.Review }, func(c *Card) { c.Due = time.Unix(1, 0) },
		func(c *Card) { c.Memory = &scheduler.Memory{Stability: 1} }, func(c *Card) { c.Step = 1 },
		func(c *Card) { c.ScheduledDays = 1 }, func(c *Card) { c.Interday = true }, func(c *Card) { c.Reps = 1 },
		func(c *Card) { c.Lapses = 1 }, func(c *Card) { c.LastReview = time.Unix(1, 0) },
	} {
		cp := *base
		edit(&cp)
		if sameSchedule(base, &cp) || !samePlacement(base, &cp) {
			t.Errorf("scheduling change %+v not seen as one", cp)
		}
	}
	for _, edit := range []func(*Card){
		func(c *Card) { c.NoteID = 9 }, func(c *Card) { c.Ord = 1 }, func(c *Card) { c.DeckID = 9 },
		func(c *Card) { c.HomeDeckID = 9 }, func(c *Card) { c.FilteredPos = 1 }, func(c *Card) { c.Flag = 1 },
		func(c *Card) { c.Suspended = true },
	} {
		cp := *base
		edit(&cp)
		if samePlacement(base, &cp) || !sameSchedule(base, &cp) {
			t.Errorf("placement change %+v not seen as one", cp)
		}
	}
}

func TestSyncUndoGetsNewStamp(t *testing.T) {
	c, _, _ := testCollection(t)
	a := addNote(t, c, "Basic", "a", "1")
	if err := c.UpdateNote(a, []string{"a", "one"}, nil); err != nil {
		t.Fatal(err)
	}
	edited := metaOf(t, c, kindNote, a)
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	undone := metaOf(t, c, kindNote, a)
	if undone.stamp <= edited.stamp || undone.seq <= edited.seq {
		t.Fatalf("undo %+v not after edit %+v", undone, edited)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if redone := metaOf(t, c, kindNote, a); redone.stamp <= undone.stamp || redone.seq <= undone.seq {
		t.Fatalf("redo %+v not after undo %+v", redone, undone)
	}

	// Deleting leaves a tombstone for the note, which deletes its cards
	// too; undoing the deletion brings back registers newer than it.
	card := c.CardsOfNote(a)[0].ID
	if err := c.DeleteNotes([]ID{a}); err != nil {
		t.Fatal(err)
	}
	nt, ok := tombOf(c, kindNote, a)
	if !ok {
		t.Fatal("no tombstone after deleting")
	}
	if _, ok := tombOf(c, kindCard, card); ok {
		t.Fatal("card deleted with its note has a tombstone of its own")
	}
	c.mu.RLock()
	_, stale := c.meta[metaKey{kindCardSched, card}]
	c.mu.RUnlock()
	if stale {
		t.Fatal("deleted card keeps its register")
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, ok := tombOf(c, kindNote, a); ok {
		t.Fatal("note tombstone left after undoing the deletion")
	}
	if m := metaOf(t, c, kindNote, a); m.stamp <= nt.stamp {
		t.Fatalf("restored note %+v not after tombstone %+v", m, nt)
	}
	if m := metaOf(t, c, kindCardSched, card); m.stamp <= nt.stamp {
		t.Fatalf("restored card %+v not after tombstone %+v", m, nt)
	}
}

func TestSyncUndoAnswer(t *testing.T) {
	c, _, _ := testCollection(t)
	card := c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID
	e, err := c.Answer(card, scheduler.Good, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	tomb, ok := tombOf(c, kindReview, e.ID)
	if !ok || tomb.aux != monthNumber(monthKey(e.Time)) {
		t.Fatalf("review tombstone = %+v, %v", tomb, ok)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if _, ok := tombOf(c, kindReview, e.ID); ok {
		t.Fatal("review tombstone left after redo")
	}
	if m := metaOf(t, c, kindReview, e.ID); m.stamp <= tomb.stamp {
		t.Fatalf("redone entry %+v not after its tombstone %+v", m, tomb)
	}
}

// Replaying the journal, and loading a compacted snapshot, rebuild exactly
// the same metadata.
func TestSyncMetaReplay(t *testing.T) {
	f := newUndoFixture(t)
	c := f.c
	if err := c.DeleteNotes([]ID{f.uno}); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFlag([]ID{f.reviewed}, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UndoReview(f.entry); err != nil {
		t.Fatal(err)
	}
	s := c.Settings()
	s.RolloverHour = 6
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	if c.store.JournalEntries() == 0 {
		t.Fatal("nothing in the journal")
	}
	replayed, _, err := Open(f.store.Clone().Open(), Options{Now: f.clk.now})
	if err != nil {
		t.Fatal(err)
	}
	sameSyncState(t, "replayed", replayed, c)
	// Open compacted the journal into a snapshot; open that.
	compacted := reopen(t, c, f.store, f.clk)
	sameSyncState(t, "compacted", compacted, c)
}

func TestSyncSettingsStampedOnlyWhenSynced(t *testing.T) {
	c, _, _ := testCollection(t)
	before := metaOf(t, c, kindSettings, 0)
	s := c.Settings()
	s.Theme = ThemeDark
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	if got := metaOf(t, c, kindSettings, 0); got != before {
		t.Fatalf("theme change stamped the synced settings: %+v", got)
	}
	s.RolloverHour = 7
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	if got := metaOf(t, c, kindSettings, 0); got.stamp <= before.stamp {
		t.Fatalf("rollover change not stamped: %+v", got)
	}
}

func TestMetaTableRoundTrip(t *testing.T) {
	m := metaTable{
		{kindNote, 1700000000000}:   {stamp: makeStamp(1700000000000, 0, 0)},
		{kindNote, 1700000000005}:   {stamp: makeStamp(1700000100000, 3, 7), seq: 12},
		{kindCardSched, 17}:         {stamp: makeStamp(5, 0, 1), seq: 1},
		{kindSettings, 0}:           {},
		{kindDeck, 1700000000009}:   {stamp: makeStamp(1700000000001, 0, 0), aux: 1700000000003},
		{kindReview, 1800000000000}: {stamp: makeStamp(1799999999990, stampCounterMax, MaxSlot), seq: 99},
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var got metaTable
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(got, m) {
		t.Fatalf("round trip:\n got %v\nwant %v\n%s", got, m, raw)
	}
	var e syncEntry
	want := syncEntry{Kind: kindReview, ID: 5, Stamp: makeStamp(1<<41, 5, 31), Seq: 9, Aux: 202610}
	raw, _ = json.Marshal(want)
	if err := json.Unmarshal(raw, &e); err != nil || e != want {
		t.Fatalf("entry round trip: %+v, %v (%s)", e, err, raw)
	}
}

// v3Store returns a store holding c as snapshot version 3 would have saved
// it, with a review log entry stored without a sequence number.
func v3Store(t *testing.T, c *Collection) *storage.MemoryStore {
	t.Helper()
	c.mu.RLock()
	s := withoutSyncMeta(c.state())
	c.mu.RUnlock()
	s.Version = 3
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemoryStore()
	st, _, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Init(raw); err != nil {
		t.Fatal(err)
	}
	e := ReviewEntry{ID: 42, CardID: 1, Time: c.Now(), Rating: scheduler.Good}
	entry, _ := json.Marshal(e)
	if err := st.Commit(json.RawMessage(`{"k":"answer"}`), []storage.RevlogWrite{{Month: monthKey(e.Time), ID: 42, Entry: entry}}, nil); err != nil {
		t.Fatal(err)
	}
	return store
}

// A version 3 collection opens with stamps from its objects' modification
// or creation times and a new epoch, and is saved as version 4.
func TestSyncMigrateV3(t *testing.T) {
	src, _, clk := testCollection(t)
	note := addNote(t, src, "Basic", "a", "1")
	clk.advance(time.Hour)
	card := src.CardsOfNote(note)[0].ID
	if _, err := src.Answer(card, scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	deck := src.Decks()[0].ID
	store := v3Store(t, src)

	c, fresh, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil || fresh {
		t.Fatalf("open: %v fresh=%v", err, fresh)
	}
	n, _ := c.Note(note)
	cd, _ := c.Card(card)
	for _, tc := range []struct {
		kind syncKind
		id   ID
		want Stamp
	}{
		{kindNote, note, stampAt(n.Modified)},
		{kindCardSched, card, stampAt(cd.Modified)},
		{kindCardPlace, card, stampAt(cd.Modified)},
		{kindDeck, deck, makeStamp(int64(deck), 0, 0)},
		{kindPreset, DefaultPresetID, makeStamp(int64(DefaultPresetID), 0, 0)},
	} {
		if m := metaOf(t, c, tc.kind, tc.id); m.stamp != tc.want || m.seq != 0 {
			t.Errorf("kind %d: %+v, want stamp %d and seq 0", tc.kind, m, tc.want)
		}
	}
	_, _, seq, epoch, _ := syncState(c)
	_, _, _, srcEpoch, _ := syncState(src)
	if epoch == "" || epoch == srcEpoch || seq != 0 {
		t.Fatalf("epoch %q (source %q), seq %d", epoch, srcEpoch, seq)
	}
	// The stored entry has no sequence number.
	if recs, err := c.store.ReadRevlogRecords(monthKey(c.Now())); err != nil || len(recs) != 1 || recs[0].Seq != 0 {
		t.Fatalf("records = %+v, %v", recs, err)
	}
	// Saved as version 4, so the epoch stays.
	_, loaded, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Version int    `json:"version"`
		Epoch   string `json:"epoch"`
	}
	if err := json.Unmarshal(loaded.Snapshot, &s); err != nil || s.Version != 4 || s.Epoch != epoch {
		t.Fatalf("snapshot version %d, epoch %q (%v)", s.Version, s.Epoch, err)
	}
	c2 := reopen(t, c, store, clk)
	sameSyncState(t, "reopened", c2, c)
	// Later changes are numbered from 1.
	if err := c2.UpdateNote(note, []string{"a", "one"}, nil); err != nil {
		t.Fatal(err)
	}
	if m := metaOf(t, c2, kindNote, note); m.seq != 1 || m.stamp <= stampAt(n.Modified) {
		t.Fatalf("edited note %+v", m)
	}
}

// Every way of replacing the collection starts a new lineage, and sequence
// numbers never go back.
func TestSyncEpochChanges(t *testing.T) {
	c, store, clk := testCollection(t)
	other, _, _ := testCollection(t)
	_, _, _, e1, _ := syncState(c)
	_, _, _, e2, _ := syncState(other)
	if e1 == "" || e1 == e2 {
		t.Fatalf("new collections have epochs %q and %q", e1, e2)
	}
	snap, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	addNote(t, c, "Basic", "a", "1")
	addNote(t, c, "Basic", "b", "2")
	_, _, before, _, _ := syncState(c)
	if err := c.Restore(snap, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	_, _, seq, restored, _ := syncState(c)
	if restored == "" || restored == e1 || seq < before {
		t.Fatalf("after restore: epoch %q (was %q), seq %d (was %d)", restored, e1, seq, before)
	}
	c = reopen(t, c, store, clk)
	if _, err := c.ReplaceWithImport(&ImportBatch{}, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, imported, _ := syncState(c); imported == restored || imported == "" {
		t.Fatalf("after importing a collection package: epoch %q", imported)
	}
	reopen(t, c, store, clk)
}

// A snapshot of a newer version is refused, whether opened or restored.
func TestSyncNewerSnapshotRefused(t *testing.T) {
	src, _, clk := testCollection(t)
	src.mu.RLock()
	s := src.state()
	src.mu.RUnlock()
	s.Version = snapshotVersion + 1
	raw, _ := json.Marshal(s)
	store := storage.NewMemoryStore()
	st, _, _ := storage.Open(store.Open())
	if err := st.Init(raw); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(store.Open(), Options{Now: clk.now}); !errors.Is(err, ErrCorrupt) || !strings.Contains(err.Error(), "update Cardflip") {
		t.Fatalf("open = %v", err)
	}
	var re *RestoreError
	if err := src.Restore(raw, nil, nil, nil); !errors.As(err, &re) || !strings.Contains(re.Msg, "newer version") {
		t.Fatalf("restore = %v", err)
	}
}
