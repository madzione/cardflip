package core

import (
	"sync"
	"time"
)

// ID identifies a deck, note type, note or card. IDs are epoch milliseconds
// made strictly increasing, so they stay below 2^53 and are exact in
// JavaScript (design D9), and newer objects have larger IDs.
type ID int64

type idGen struct {
	mu   sync.Mutex
	last int64
	now  func() time.Time
	// slotted is set once the collection has a device slot (sync design
	// decision 4): IDs are then congruent to slot modulo slotCount, so
	// devices that sync never create the same ID.
	slotted bool
	slot    int64
}

// slotCount is the number of device slots.
const slotCount = MaxSlot + 1

func (g *idGen) next() ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.now().UnixMilli()
	if n <= g.last {
		n = g.last + 1
	}
	if g.slotted {
		// The smallest such ID at or above n: at most 31 ms ahead.
		n += ((g.slot-n)%slotCount + slotCount) % slotCount
	}
	g.last = n
	return ID(n)
}

// setSlot makes IDs congruent to slot, or allocates them without a slot
// when ok is false.
func (g *idGen) setSlot(slot int, ok bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.slotted, g.slot = ok, int64(slot)
}

// observe makes sure future IDs are above id (after loading or replay).
func (g *idGen) observe(id int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if id > g.last {
		g.last = id
	}
}

func (g *idGen) high() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.last
}
