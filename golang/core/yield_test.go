package core

import (
	"fmt"
	"testing"
	"time"

	"cardflip/golang/storage"
)

// Long loops (import, journal replay, restore) call the yield function at a
// steady cadence, so the web UI keeps running during them.
func TestYieldCadence(t *testing.T) {
	yields := 0
	store := storage.NewMemoryStore()
	opts := Options{Now: func() time.Time { return time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local) }, Yield: func() { yields++ }}
	c, _, err := Open(store.Open(), opts)
	if err != nil {
		t.Fatal(err)
	}
	var basic ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	b := &ImportBatch{NoteTypes: []ImportNoteType{{Key: "b", ExistingID: basic}}, Media: map[string][]byte{}}
	for i := 0; i < 1000; i++ {
		b.Notes = append(b.Notes, ImportNote{TypeKey: "b", DeckID: c.Decks()[0].ID, Fields: []string{fmt.Sprint("q", i), "a"}})
	}
	for i := 0; i < 3; i++ {
		b.Media[fmt.Sprintf("m%d.png", i)] = []byte{byte(i)}
	}
	yields = 0
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	// One yield per 200 notes and one per media file.
	if want := 1000/200 + 3; yields != want {
		t.Fatalf("import yielded %d times, want %d", yields, want)
	}

	// Journal replay at open: once per 500 ops.
	for i := 0; i < 1000; i++ {
		c.SetDeckCollapsed(c.Decks()[0].ID, i%2 == 0)
	}
	yields = 0
	if _, _, err := Open(store.Open(), opts); err != nil {
		t.Fatal(err)
	}
	if yields < 2 {
		t.Fatalf("replay of 1000 ops yielded %d times", yields)
	}
}
