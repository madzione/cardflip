package core

import "sync"

// Change describes what an op changed, so screens can refresh what they show.
type Change struct {
	Kind      string `json:"kind"`
	Settings  bool   `json:"settings"`
	Decks     bool   `json:"decks"`
	NoteTypes bool   `json:"noteTypes"`
	Notes     bool   `json:"notes"`
	Cards     bool   `json:"cards"`
	Reviews   bool   `json:"reviews"`
	// Media is set when a media file was added.
	Media bool `json:"media"`
	// All is set when the whole collection was replaced (restore).
	All bool `json:"all"`
}

func changeOf(o *op) Change {
	return Change{
		Kind:      o.Kind,
		Settings:  o.Settings != nil,
		Decks:     len(o.Presets)+len(o.DelPresets)+len(o.Decks)+len(o.DelDecks) > 0,
		NoteTypes: len(o.NoteTypes)+len(o.DelNoteTypes) > 0,
		Notes:     len(o.Notes)+len(o.DelNotes) > 0,
		Cards:     len(o.Cards)+len(o.DelCards) > 0,
		Reviews:   o.Review != nil || o.ReviewUndone != 0 || len(o.AddReviews)+len(o.DelReviews) > 0,
		Media:     len(o.Media) > 0,
	}
}

func (a Change) merge(b Change) Change {
	return Change{
		Kind:      b.Kind,
		Settings:  a.Settings || b.Settings,
		Decks:     a.Decks || b.Decks,
		NoteTypes: a.NoteTypes || b.NoteTypes,
		Notes:     a.Notes || b.Notes,
		Cards:     a.Cards || b.Cards,
		Reviews:   a.Reviews || b.Reviews,
		Media:     a.Media || b.Media,
		All:       a.All || b.All,
	}
}

// broadcaster delivers changes to subscribers. A slow subscriber never blocks
// a mutation: changes it has not read yet are merged into one.
type broadcaster struct {
	mu     sync.Mutex
	subs   map[chan Change]bool
	closed bool
}

func newBroadcaster() *broadcaster { return &broadcaster{subs: map[chan Change]bool{}} }

// Subscribe returns a channel of changes and a function that ends the
// subscription and closes the channel.
func (c *Collection) Subscribe() (<-chan Change, func()) {
	return c.events.subscribe()
}

func (b *broadcaster) subscribe() (<-chan Change, func()) {
	ch := make(chan Change, 1)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		close(ch)
		return ch, func() {}
	}
	b.subs[ch] = true
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			if b.subs[ch] {
				delete(b.subs, ch)
				close(ch)
			}
		})
	}
}

func (b *broadcaster) publish(c Change) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- c:
		default:
			select {
			case old := <-ch:
				ch <- old.merge(c)
			default:
				ch <- c
			}
		}
	}
}

func (b *broadcaster) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	for ch := range b.subs {
		close(ch)
		delete(b.subs, ch)
	}
}
