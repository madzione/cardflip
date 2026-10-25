package core

import (
	"slices"
	"testing"
	"time"

	"cardflip/golang/internal/usermsg"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// device is a collection with what its link to one peer remembers.
type device struct {
	name  string
	c     *Collection
	store *storage.MemoryStore
	clk   *clock
	// recv is the peer's sequence number received so far, echo the local
	// sequence numbers of the changes received from it, and last the local
	// sequence number at the end of the last sync.
	recv int64
	echo SeqRanges
	last int64
}

// pull applies everything src has that dst has not received, in small
// pages.
func pull(t *testing.T, dst, src *device) {
	t.Helper()
	req := DeltaRequest{Since: dst.recv, Skip: src.echo, MaxItems: 7}
	watermark := int64(-1)
	for {
		page, err := src.c.Delta(req)
		if err != nil {
			t.Fatal(err)
		}
		if watermark < 0 {
			watermark = page.Seq
		}
		res, err := dst.c.ApplyBatch(page.Items, BatchSource{Link: src.name, Received: dst.echo, LastSync: dst.last})
		if err != nil {
			t.Fatal(err)
		}
		if res.Last > 0 {
			dst.echo = dst.echo.Add(res.First, res.Last)
		}
		if page.Done {
			break
		}
		req.After = page.Items[len(page.Items)-1].Cursor()
	}
	dst.recv = watermark
}

// syncPair syncs a and b in both directions.
func syncPair(t *testing.T, a, b *device) {
	t.Helper()
	pull(t, a, b)
	pull(t, b, a)
	a.last, b.last = a.c.seq, b.c.seq
}

// linkedPair returns two devices with the same collection, as after a
// backup was restored on the second, already synced once.
func linkedPair(t *testing.T, setup func(*Collection)) (a, b *device) {
	t.Helper()
	ca, sa, clka := testCollection(t)
	if setup != nil {
		setup(ca)
	}
	cb, sb, clkb := testCollection(t)
	clkb.t = clka.t
	snap, err := ca.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := cb.Restore(snap, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	slots := map[string]int{"LAPTOP": 0, "PHONE": 1}
	ca.SetDevice("LAPTOP")
	cb.SetDevice("PHONE")
	noErr(t, ca.SetSlots(slots))
	noErr(t, cb.SetSlots(slots))
	a = &device{name: "laptop", c: ca, store: sa, clk: clka, recv: -1}
	b = &device{name: "phone", c: cb, store: sb, clk: clkb, recv: -1}
	syncPair(t, a, b)
	// Pairing is long after the setup: an app session of its own.
	for _, c := range []*Collection{ca, cb} {
		c.mu.Lock()
		c.hist.clear()
		c.mu.Unlock()
	}
	return a, b
}

// advance moves both clocks.
func advance(d time.Duration, devs ...*device) {
	for _, dev := range devs {
		dev.clk.advance(d)
	}
}

// sameContent fails unless both collections hold the same objects and
// synced settings.
func sameContent(t *testing.T, a, b *device) {
	t.Helper()
	sa, sb := collectionState(a.c), collectionState(b.c)
	sa.NextID, sb.NextID = 0, 0
	sa.Settings, sb.Settings = withSyncedSettings(Settings{}, sa.Settings), withSyncedSettings(Settings{}, sb.Settings)
	if d := jsonDiff(sa, sb); d != "" {
		t.Fatalf("%s and %s differ:\n%s", a.name, b.name, d)
	}
}

func fieldsOf(t *testing.T, c *Collection, id ID) []string {
	t.Helper()
	n, err := c.Note(id)
	if err != nil {
		t.Fatal(err)
	}
	return n.Fields
}

func TestSyncDifferentNotesEdited(t *testing.T) {
	var na, nb ID
	a, b := linkedPair(t, func(c *Collection) {
		na = addNote(t, c, "Basic", "a", "1")
		nb = addNote(t, c, "Basic", "b", "2")
	})
	advance(time.Minute, a, b)
	noErr(t, a.c.UpdateNote(na, []string{"a", "laptop"}, nil))
	noErr(t, b.c.UpdateNote(nb, []string{"b", "phone"}, nil))
	syncPair(t, a, b)
	for _, d := range []*device{a, b} {
		if f := fieldsOf(t, d.c, na); f[1] != "laptop" {
			t.Errorf("%s: note A = %q", d.name, f)
		}
		if f := fieldsOf(t, d.c, nb); f[1] != "phone" {
			t.Errorf("%s: note B = %q", d.name, f)
		}
	}
	sameContent(t, a, b)
}

func TestSyncSameNoteEdited(t *testing.T) {
	var n ID
	a, b := linkedPair(t, func(c *Collection) { n = addNote(t, c, "Basic", "a", "1") })
	advance(time.Minute, a, b)
	noErr(t, a.c.UpdateNote(n, []string{"a", "laptop at 10:00"}, nil))
	advance(5*time.Minute, a, b)
	noErr(t, b.c.UpdateNote(n, []string{"a", "phone at 10:05"}, nil))
	syncPair(t, a, b)
	for _, d := range []*device{a, b} {
		if f := fieldsOf(t, d.c, n); f[1] != "phone at 10:05" {
			t.Errorf("%s: note = %q", d.name, f)
		}
	}
	sameContent(t, a, b)
}

// A card's scheduling and its flag are merged separately.
func TestSyncReviewAndFlag(t *testing.T) {
	var card ID
	a, b := linkedPair(t, func(c *Collection) { card = c.CardsOfNote(addNote(t, c, "Basic", "a", "1"))[0].ID })
	advance(time.Minute, a, b)
	if _, err := b.c.Answer(card, scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	advance(time.Minute, a, b)
	noErr(t, a.c.SetFlag([]ID{card}, 1))
	syncPair(t, a, b)
	answered, _ := b.c.Card(card)
	for _, d := range []*device{a, b} {
		got, _ := d.c.Card(card)
		if got.Flag != 1 || got.State != answered.State || !got.Due.Equal(answered.Due) || got.Reps != 1 {
			t.Errorf("%s: card = %+v", d.name, got)
		}
		if h, _ := d.c.CardHistory(card); len(h) != 1 {
			t.Errorf("%s: history = %+v", d.name, h)
		}
	}
	sameContent(t, a, b)
}

// Answers on both devices end up in both review logs, once each.
func TestSyncReviewHistoriesCombined(t *testing.T) {
	var notes []ID
	a, b := linkedPair(t, func(c *Collection) {
		for i := range 50 {
			notes = append(notes, addNote(t, c, "Basic", "q", string(rune('a'+i%26))+string(rune('a'+i/26))))
		}
	})
	advance(time.Minute, a, b)
	for i, n := range notes {
		d := a
		if i >= 30 {
			d = b
		}
		if _, err := d.c.Answer(d.c.CardsOfNote(n)[0].ID, scheduler.Good, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	syncPair(t, a, b)
	syncPair(t, a, b)
	for _, d := range []*device{a, b} {
		d.c.mu.RLock()
		n := len(d.c.todayEntries(d.c.now()))
		d.c.mu.RUnlock()
		if n != 50 {
			t.Errorf("%s: %d answers today, want 50", d.name, n)
		}
		months, _ := d.c.RevlogMonths()
		total := 0
		for _, m := range months {
			e, _ := d.c.ReadRevlog(m)
			total += len(e)
		}
		if total != 50 {
			t.Errorf("%s: %d stored entries, want 50", d.name, total)
		}
	}
	sameContent(t, a, b)
}

func TestSyncDeletePropagates(t *testing.T) {
	var ids []ID
	a, b := linkedPair(t, func(c *Collection) {
		for _, f := range []string{"uno", "dos", "tres", "cuatro"} {
			ids = append(ids, addNote(t, c, "Basic (and reversed card)", f, "x"))
		}
	})
	advance(time.Minute, a, b)
	noErr(t, a.c.DeleteNotes(ids[:3]))
	syncPair(t, a, b)
	for _, id := range ids[:3] {
		if _, err := b.c.Note(id); err == nil {
			t.Errorf("note %d still on the phone", id)
		}
		if cards := b.c.CardsOfNote(id); len(cards) != 0 {
			t.Errorf("note %d's cards still on the phone", id)
		}
	}
	if len(b.c.CardsOfNote(ids[3])) != 2 {
		t.Error("the kept note lost its cards")
	}
	log, err := b.c.DeletionLog()
	if err != nil || len(log) != 3 {
		t.Fatalf("phone's deletion log = %+v, %v", log, err)
	}
	sameContent(t, a, b)
	reopen(t, b.c, b.store, b.clk)
}

// A note edited after it was deleted elsewhere comes back with its cards.
func TestSyncEditedAfterDeletion(t *testing.T) {
	var n ID
	a, b := linkedPair(t, func(c *Collection) { n = addNote(t, c, "Basic (and reversed card)", "a", "1") })
	advance(time.Minute, a, b)
	noErr(t, a.c.DeleteNotes([]ID{n}))
	advance(10*time.Minute, a, b)
	noErr(t, b.c.UpdateNote(n, []string{"a", "edited on the phone"}, nil))
	syncPair(t, a, b)
	syncPair(t, a, b)
	for _, d := range []*device{a, b} {
		if f := fieldsOf(t, d.c, n); f[1] != "edited on the phone" {
			t.Errorf("%s: note = %q", d.name, f)
		}
		if cards := d.c.CardsOfNote(n); len(cards) != 2 {
			t.Errorf("%s: %d cards", d.name, len(cards))
		}
	}
	sameContent(t, a, b)
}

func TestSyncSettings(t *testing.T) {
	a, b := linkedPair(t, nil)
	advance(time.Minute, a, b)
	s := a.c.Settings()
	s.RolloverHour, s.Theme, s.Language, s.RestDaysPerWeek = 6, ThemeDark, "de", 1
	noErr(t, a.c.UpdateSettings(s))
	s = b.c.Settings()
	s.Theme, s.Language = ThemeLight, "en"
	noErr(t, b.c.UpdateSettings(s))
	syncPair(t, a, b)
	// Rest days synced (spec).
	if got := b.c.Settings(); got.RolloverHour != 6 || got.Theme != ThemeLight || got.Language != "en" || got.RestDaysPerWeek != 1 {
		t.Fatalf("phone's settings = %+v", got)
	}
	if got := a.c.Settings(); got.RolloverHour != 6 || got.Theme != ThemeDark || got.Language != "de" {
		t.Fatalf("laptop's settings = %+v", got)
	}
}

// Undo keeps working across a sync that changed something else, and has
// nothing left once a sync changed what it would revert.
func TestSyncUndoHistory(t *testing.T) {
	var na, nb ID
	a, b := linkedPair(t, func(c *Collection) {
		na = addNote(t, c, "Basic", "a", "1")
		nb = addNote(t, c, "Basic", "b", "2")
	})
	advance(time.Minute, a, b)
	noErr(t, a.c.UpdateNote(na, []string{"a", "laptop"}, nil))
	noErr(t, b.c.UpdateNote(nb, []string{"b", "phone"}, nil))
	syncPair(t, a, b)
	if name, err := a.c.Undo(); err != nil || usermsg.Plain(name) != "Edit note" {
		t.Fatalf("undo after an unrelated sync = %q, %v", name, err)
	}
	if f := fieldsOf(t, a.c, na); f[1] != "1" {
		t.Fatalf("note A after undo = %q", f)
	}
	if f := fieldsOf(t, a.c, nb); f[1] != "phone" {
		t.Fatalf("note B after undo = %q", f)
	}

	// A later edit of the same note from the phone clears the history.
	noErr(t, a.c.UpdateNote(na, []string{"a", "laptop again"}, nil))
	advance(time.Minute, a, b)
	noErr(t, b.c.UpdateNote(na, []string{"a", "phone, later"}, nil))
	syncPair(t, a, b)
	if undo, redo := a.c.UndoStatus(); undo != "" || redo != "" {
		t.Fatalf("history after a related sync: %q, %q", undo, redo)
	}
	sameContent(t, a, b)
}

// Undoing a deletion that was synced brings the note back everywhere.
func TestSyncUndoAfterSync(t *testing.T) {
	var n ID
	a, b := linkedPair(t, func(c *Collection) { n = addNote(t, c, "Basic", "a", "1") })
	advance(time.Minute, a, b)
	noErr(t, a.c.DeleteNotes([]ID{n}))
	syncPair(t, a, b)
	if _, err := b.c.Note(n); err == nil {
		t.Fatal("deletion not synced")
	}
	advance(time.Minute, a, b)
	if _, err := a.c.Undo(); err != nil {
		t.Fatal(err)
	}
	syncPair(t, a, b)
	if f := fieldsOf(t, b.c, n); f[0] != "a" || len(b.c.CardsOfNote(n)) != 1 {
		t.Fatalf("note on the phone after undo = %q", f)
	}
	sameContent(t, a, b)
}

// Syncing again sends nothing back and changes nothing.
func TestSyncIdempotent(t *testing.T) {
	var n ID
	a, b := linkedPair(t, func(c *Collection) { n = addNote(t, c, "Basic", "a", "1") })
	advance(time.Minute, a, b)
	noErr(t, a.c.UpdateNote(n, []string{"a", "x"}, nil))
	syncPair(t, a, b)
	revA, revB := a.c.Revision(), b.c.Revision()
	syncPair(t, a, b)
	if a.c.Revision() != revA || b.c.Revision() != revB {
		t.Fatal("a second sync changed something")
	}
	// Applying the same items twice is harmless.
	page, err := a.c.Delta(DeltaRequest{Since: -1})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := b.c.ApplyBatch(page.Items, BatchSource{}); err != nil {
			t.Fatal(err)
		}
	}
	sameContent(t, a, b)
}

func TestSeqRanges(t *testing.T) {
	var r SeqRanges
	r = r.Add(5, 7).Add(10, 12).Add(8, 9).Add(20, 20)
	if !slices.Equal(r, SeqRanges{{5, 12}, {20, 20}}) {
		t.Fatalf("ranges = %v", r)
	}
	for seq, want := range map[int64]bool{4: false, 5: true, 9: true, 12: true, 13: false, 20: true, 21: false} {
		if r.Contains(seq) != want {
			t.Errorf("Contains(%d) = %v", seq, !want)
		}
	}
	if p := r.Prune(10); !slices.Equal(p, SeqRanges{{11, 12}, {20, 20}}) {
		t.Fatalf("pruned = %v", p)
	}
}
