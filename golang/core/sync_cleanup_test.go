package core

import (
	"testing"
	"time"

	"cardflip/golang/storage"
)

// separatePair returns two collections created separately (their clocks a
// few seconds apart, so that their IDs differ), with device slots, not yet
// synced.
func separatePair(t *testing.T) (a, b *device) {
	t.Helper()
	ca, sa, clka := testCollection(t)
	clkb := &clock{t: clka.t.Add(7 * time.Second)}
	sb := storage.NewMemoryStore()
	cb, _, err := Open(sb.Open(), Options{Now: clkb.now})
	if err != nil {
		t.Fatal(err)
	}
	slots := map[string]int{"LAPTOP": 0, "PHONE": 1}
	ca.SetDevice("LAPTOP")
	cb.SetDevice("PHONE")
	noErr(t, ca.SetSlots(slots))
	noErr(t, cb.SetSlots(slots))
	clka.t = clkb.t
	return &device{name: "laptop", c: ca, store: sa, clk: clka, recv: -1},
		&device{name: "phone", c: cb, store: sb, clk: clkb, recv: -1}
}

func decksNamed(c *Collection, name string) []Deck {
	var out []Deck
	for _, d := range c.Decks() {
		if d.Name == name {
			out = append(out, d)
		}
	}
	return out
}

// Two separately created collections each have a Default deck: merged,
// there is one, holding the cards of both, and one of each built-in note
// type.
func TestSyncTwoDefaultDecks(t *testing.T) {
	a, b := separatePair(t)
	na := addNote(t, a.c, "Basic", "from the laptop", "")
	nb := addNote(t, b.c, "Basic", "from the phone", "")
	syncPair(t, a, b)
	syncPair(t, a, b)
	for _, d := range []*device{a, b} {
		defaults := decksNamed(d.c, DefaultDeckName)
		if len(defaults) != 1 {
			t.Fatalf("%s has %d Default decks: %+v", d.name, len(defaults), d.c.Decks())
		}
		for _, n := range []ID{na, nb} {
			cards := d.c.CardsOfNote(n)
			if len(cards) != 1 || cards[0].DeckID != defaults[0].ID {
				t.Errorf("%s: cards of note %d = %+v, want in deck %d", d.name, n, cards, defaults[0].ID)
			}
		}
		basics := 0
		for _, nt := range d.c.NoteTypes() {
			if nt.Name == "Basic" {
				basics++
			}
		}
		if basics != 1 {
			t.Errorf("%s has %d Basic note types", d.name, basics)
		}
	}
	sameContent(t, a, b)
}

// Note types of the same name with different fields are both kept: the
// newer one is renamed the same way on both devices.
func TestSyncNoteTypeNameClash(t *testing.T) {
	a, b := separatePair(t)
	va, err := a.c.CloneNoteType(noteType(t, a.c, "Basic").ID, "Vocab")
	noErr(t, err)
	noErr(t, a.c.AddField(va, "Example"))
	vb, err := b.c.CloneNoteType(noteType(t, b.c, "Basic").ID, "Vocab")
	noErr(t, err)
	noErr(t, b.c.AddField(vb, "Plural"))
	syncPair(t, a, b)
	syncPair(t, a, b)
	older, newer := min(va, vb), max(va, vb)
	for _, d := range []*device{a, b} {
		o, err := d.c.NoteType(older)
		noErr(t, err)
		n, err := d.c.NoteType(newer)
		noErr(t, err)
		if o.Name != "Vocab" || n.Name != "Vocab (2)" {
			t.Errorf("%s: %q and %q", d.name, o.Name, n.Name)
		}
	}
	sameContent(t, a, b)
}

// Presets of the same name are both kept, the newer one renamed.
func TestSyncPresetNameClash(t *testing.T) {
	a, b := separatePair(t)
	pa, err := a.c.ClonePreset(defaultDeck(a.c), "Exam")
	noErr(t, err)
	pb, err := b.c.ClonePreset(defaultDeck(b.c), "Exam")
	noErr(t, err)
	syncPair(t, a, b)
	syncPair(t, a, b)
	names := map[ID]string{}
	for _, p := range a.c.Presets() {
		names[p.ID] = p.Name
	}
	if names[min(pa, pb)] != "Exam" || names[max(pa, pb)] != "Exam (2)" {
		t.Fatalf("presets = %v", names)
	}
	sameContent(t, a, b)
}

// A card that arrives for a deck that was combined into another lands in
// the kept deck.
func TestSyncCardForRedirectedDeck(t *testing.T) {
	a, b := separatePair(t)
	addNote(t, a.c, "Basic", "laptop", "")
	addNote(t, b.c, "Basic", "phone", "")
	pull(t, a, b) // the laptop combines the two Default decks
	if len(decksNamed(a.c, DefaultDeckName)) != 1 {
		t.Fatal("Default decks not combined")
	}
	kept := decksNamed(a.c, DefaultDeckName)[0].ID
	gone := decksNamed(b.c, DefaultDeckName)[0].ID
	if kept == gone {
		t.Fatal("the phone's Default deck was kept")
	}
	advance(time.Minute, a, b)
	late := addNote(t, b.c, "Basic", "added before the phone heard", "")
	if b.c.CardsOfNote(late)[0].DeckID != gone {
		t.Fatal("setup: the late card is not in the phone's Default deck")
	}
	pull(t, a, b)
	if cards := a.c.CardsOfNote(late); len(cards) != 1 || cards[0].DeckID != kept {
		t.Fatalf("late card on the laptop = %+v, want in deck %d", cards, kept)
	}
	pull(t, b, a)
	if cards := b.c.CardsOfNote(late); len(cards) != 1 || cards[0].DeckID != kept {
		t.Fatalf("late card on the phone = %+v, want in deck %d", cards, kept)
	}
	// The phone's own clean-up moved its cards too, with later times.
	syncPair(t, a, b)
	sameContent(t, a, b)
}
