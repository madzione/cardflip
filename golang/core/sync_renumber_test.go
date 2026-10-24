package core

import (
	"encoding/json"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// collidingPair returns two collections created separately at the same
// moment, so that they use the same IDs for different objects.
func collidingPair(t *testing.T) (a, b *device) {
	t.Helper()
	ca, sa, clka := testCollection(t)
	cb, sb, clkb := testCollection(t)
	ca.SetDevice("LAPTOP")
	cb.SetDevice("PHONE")
	return &device{name: "laptop", c: ca, store: sa, clk: clka, recv: -1},
		&device{name: "phone", c: cb, store: sb, clk: clkb, recv: -1}
}

func linkSlots(t *testing.T, a, b *device) {
	t.Helper()
	slots := map[string]int{"LAPTOP": 0, "PHONE": 1}
	noErr(t, a.c.SetSlots(slots))
	noErr(t, b.c.SetSlots(slots))
}

// Notes, cards and review log entries that use the same IDs for different
// things are all kept, with their cards and history, after the phone
// renumbers its own.
func TestSyncRenumberCollisions(t *testing.T) {
	a, b := collidingPair(t)
	na := addNote(t, a.c, "Basic", "laptop note", "")
	nb := addNote(t, b.c, "Basic", "phone note", "")
	if na != nb {
		t.Fatalf("setup: note IDs %d and %d do not collide", na, nb)
	}
	ca, cb := a.c.CardsOfNote(na)[0].ID, b.c.CardsOfNote(nb)[0].ID
	ea, err := a.c.Answer(ca, scheduler.Good, time.Second)
	noErr(t, err)
	eb, err := b.c.Answer(cb, scheduler.Again, time.Second)
	noErr(t, err)
	if ca != cb || ea.ID != eb.ID {
		t.Fatalf("setup: cards %d/%d, entries %d/%d do not collide", ca, cb, ea.ID, eb.ID)
	}
	linkSlots(t, a, b)
	ids, err := a.c.Identities()
	noErr(t, err)
	n, err := b.c.Renumber(ids)
	noErr(t, err)
	if n != 3 {
		t.Fatalf("renumbered %d objects, want 3 (note, card, entry)", n)
	}
	if _, ok := tombOf(b.c, kindNote, nb); ok {
		t.Fatal("renumbering left a tombstone for the old note ID")
	}
	if log, _ := b.c.DeletionLog(); len(log) != 0 {
		t.Fatalf("renumbering wrote the deletion log: %+v", log)
	}
	syncPair(t, a, b)
	syncPair(t, a, b)
	for _, d := range []*device{a, b} {
		var texts []string
		for _, id := range d.c.AllNoteIDs() {
			n, _ := d.c.Note(id)
			texts = append(texts, n.Fields[0])
			cards := d.c.CardsOfNote(id)
			if len(cards) != 1 {
				t.Fatalf("%s: note %q has %d cards", d.name, n.Fields[0], len(cards))
			}
			h, err := d.c.CardHistory(cards[0].ID)
			noErr(t, err)
			want := scheduler.Good
			if n.Fields[0] == "phone note" {
				want = scheduler.Again
			}
			if len(h) != 1 || h[0].Rating != want {
				t.Fatalf("%s: history of %q = %+v", d.name, n.Fields[0], h)
			}
		}
		if len(texts) != 2 {
			t.Fatalf("%s: notes %q", d.name, texts)
		}
	}
	sameContent(t, a, b)
	reopen(t, b.c, b.store, b.clk)
}

// A collection restored from another's backup has the same objects under
// the same IDs: nothing is renumbered.
func TestSyncRenumberBackupCopy(t *testing.T) {
	ca, _, _ := testCollection(t)
	n := addNote(t, ca, "Basic", "a", "")
	if _, err := ca.Answer(ca.CardsOfNote(n)[0].ID, scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	cb, _, clkb := testCollection(t)
	clkb.advance(time.Hour)
	snap, _ := ca.Snapshot()
	var revlog []json.RawMessage
	months, _ := ca.RevlogMonths()
	for _, m := range months {
		e, _ := ca.ReadRevlog(m)
		revlog = append(revlog, e...)
	}
	noErr(t, cb.Restore(snap, revlog, nil, nil))
	ids, err := ca.Identities()
	noErr(t, err)
	if got, err := cb.Renumber(ids); err != nil || got != 0 {
		t.Fatalf("Renumber = %d, %v", got, err)
	}
}

// Tombstones stay until every link has received them, or for at most 180
// days.
func TestSyncPruneTombstones(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic", "a", "")
	b := addNote(t, c, "Basic", "b", "")
	noErr(t, c.DeleteNotes([]ID{a}))
	ta, _ := tombOf(c, kindNote, a)
	clk.advance(time.Hour)
	noErr(t, c.DeleteNotes([]ID{b}))
	tb, _ := tombOf(c, kindNote, b)
	if n, err := c.PruneTombstones(ta.seq-1, TombstoneRetention); err != nil || n != 0 {
		t.Fatalf("pruned %d (%v) before any link received them", n, err)
	}
	if n, err := c.PruneTombstones(ta.seq, TombstoneRetention); err != nil || n != 1 {
		t.Fatalf("pruned %d (%v) once the first was received", n, err)
	}
	if _, ok := tombOf(c, kindNote, a); ok {
		t.Fatal("received tombstone kept")
	}
	if _, ok := tombOf(c, kindNote, b); !ok {
		t.Fatal("tombstone not received yet was dropped")
	}
	c = reopen(t, c, store, clk)
	// Without links, only age counts.
	clk.advance(TombstoneRetention - 2*time.Hour)
	if n, _ := c.PruneTombstones(-1, TombstoneRetention); n != 0 {
		t.Fatalf("pruned %d young tombstones", n)
	}
	clk.advance(2*time.Hour + time.Minute)
	if n, _ := c.PruneTombstones(-1, TombstoneRetention); n != 1 {
		t.Fatalf("pruned %d old tombstones, want 1", n)
	}
	if _, ok := tombOf(c, kindNote, b); ok {
		t.Fatalf("tombstone %+v older than 180 days kept", tb)
	}
	reopen(t, c, store, clk)
}
