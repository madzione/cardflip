package core

import (
	"math/rand/v2"
	"testing"
	"time"
)

func TestIDsStrictlyIncreasing(t *testing.T) {
	now := time.UnixMilli(1_000)
	g := &idGen{now: func() time.Time { return now }}
	a, b := g.next(), g.next()
	if a != 1000 || b != 1001 {
		t.Fatalf("ids %d %d", a, b)
	}
	now = time.UnixMilli(500) // clock went backwards
	if c := g.next(); c != 1002 {
		t.Fatalf("id after clock step back = %d", c)
	}
	g.observe(5000)
	if d := g.next(); d != 5001 {
		t.Fatalf("id after observe = %d", d)
	}
	now = time.UnixMilli(9000)
	if e := g.next(); e != 9000 {
		t.Fatalf("id = %d", e)
	}
	if e := g.next(); e >= 1<<53 {
		t.Fatal("id not exact in JavaScript")
	}
}

// With a device slot, IDs are congruent to the slot, increase, stay exact
// in JavaScript and stay within 31 ms of the clock for single additions.
func TestIDsWithSlot(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	g := &idGen{now: func() time.Time { return now }}
	g.setSlot(7, true)
	last := ID(0)
	r := rand.New(rand.NewPCG(1, 2))
	for i := range 2000 {
		if i%3 != 0 {
			now = now.Add(time.Duration(r.IntN(200)) * time.Millisecond)
		}
		id := g.next()
		if id%slotCount != 7 || id <= last || id >= 1<<53 {
			t.Fatalf("id %d after %d (mod %d = %d)", id, last, slotCount, id%slotCount)
		}
		if ahead := int64(id) - now.UnixMilli(); i%3 != 0 && int64(last) < now.UnixMilli() && (ahead < 0 || ahead > 31) {
			t.Fatalf("id %d is %d ms from the clock", id, ahead)
		}
		last = id
	}
	// Many additions in one millisecond are 32 apart.
	a, b := g.next(), g.next()
	if b-a != slotCount {
		t.Fatalf("ids %d, %d", a, b)
	}
	// Back to unslotted allocation.
	g.setSlot(0, false)
	if c := g.next(); c != b+1 {
		t.Fatalf("unslotted id %d after %d", c, b)
	}
}

// A collection allocates by slot once the slot table names its device, and
// as before otherwise.
func TestCollectionIDsUseSlot(t *testing.T) {
	c, store, clk := testCollection(t)
	c.SetDevice("DEVA")
	plain := []ID{addNote(t, c, "Basic", "a", ""), addNote(t, c, "Basic", "b", "")}
	if plain[1] != plain[0]+2 { // note and card
		t.Fatalf("unlinked ids %v", plain)
	}
	steps, _ := c.LastStep()
	if err := c.SetSlots(map[string]int{"DEVB": 0, "DEVA": 5}); err != nil {
		t.Fatal(err)
	}
	if after, _ := c.LastStep(); after != steps {
		t.Fatal("slot table change is an undo step")
	}
	for range 5 {
		clk.advance(3 * time.Millisecond)
		n := addNote(t, c, "Basic", "x", "")
		card := c.CardsOfNote(n)[0].ID
		if n%slotCount != 5 || card%slotCount != 5 || card <= n {
			t.Fatalf("note %d, card %d not in slot 5", n, card)
		}
		if st := metaOf(t, c, kindNote, n).stamp; int(st&MaxSlot) != 5 {
			t.Fatalf("stamp %d has slot %d", st, st&MaxSlot)
		}
	}
	// The slot table is journaled; the device ID is not stored.
	c2, _, err := Open(store.Clone().Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c2.Slot(); ok {
		t.Fatal("slot without the device ID")
	}
	c2.SetDevice("DEVA")
	if slot, ok := c2.Slot(); !ok || slot != 5 {
		t.Fatalf("slot = %d, %v", slot, ok)
	}
	if n := addNote(t, c2, "Basic", "y", ""); n%slotCount != 5 {
		t.Fatalf("note %d after reopening", n)
	}
	if err := c2.SetSlots(map[string]int{"DEVA": 32}); err == nil {
		t.Fatal("slot 32 accepted")
	}
}
