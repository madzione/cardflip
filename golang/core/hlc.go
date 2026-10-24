package core

import (
	"sync"
	"time"
)

// Stamp is a hybrid logical clock reading (sync design decision 1): wall
// clock milliseconds, a counter for readings within the same millisecond,
// and the slot of the device that made it as tie-breaker. Stamps compare as
// integers. They exceed 2^53, so they never reach Dart.
type Stamp uint64

// Stamp layout: [milliseconds:42][counter:17][slot:5]. 42 bits of
// milliseconds last until the year 2109.
const (
	stampSlotBits    = 5
	stampCounterBits = 17
	stampCounterMax  = 1<<stampCounterBits - 1
	// MaxSlot is the highest device slot (see Collection.Slot).
	MaxSlot = 1<<stampSlotBits - 1
)

func makeStamp(ms int64, counter uint64, slot int) Stamp {
	if ms < 0 {
		ms = 0
	}
	return Stamp(uint64(ms)<<(stampCounterBits+stampSlotBits) | counter<<stampSlotBits | uint64(slot&MaxSlot))
}

// Millis returns the stamp's wall clock milliseconds.
func (s Stamp) Millis() int64 { return int64(s >> (stampCounterBits + stampSlotBits)) }

func (s Stamp) counter() uint64 { return uint64(s>>stampSlotBits) & stampCounterMax }

// Time returns the stamp's wall clock time.
func (s Stamp) Time() time.Time { return time.UnixMilli(s.Millis()) }

// stampAt is the stamp of an object last changed at t that has no recorded
// stamp (migration from snapshot version 3): counter 0, no slot.
func stampAt(t time.Time) Stamp {
	if t.IsZero() {
		return 0
	}
	return makeStamp(t.UnixMilli(), 0, 0)
}

// hlc is the collection's hybrid logical clock. It never goes backwards and
// stays above every stamp it observed, local or received, so a change made
// after seeing another one gets a larger stamp even if the wall clock is
// behind.
type hlc struct {
	mu   sync.Mutex
	last Stamp
	slot int
	wall func() time.Time
}

// now returns a new stamp, larger than every stamp so far.
func (h *hlc) now() Stamp {
	h.mu.Lock()
	defer h.mu.Unlock()
	ms, ctr := h.last.Millis(), h.last.counter()
	if pt := h.wall().UnixMilli(); pt > ms {
		ms, ctr = pt, 0
	} else if ctr++; ctr > stampCounterMax {
		ms, ctr = ms+1, 0
	}
	h.last = makeStamp(ms, ctr, h.slot)
	return h.last
}

// observe makes sure later stamps are above s.
func (h *hlc) observe(s Stamp) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s > h.last {
		h.last = s
	}
}

func (h *hlc) setSlot(slot int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.slot = slot
}
