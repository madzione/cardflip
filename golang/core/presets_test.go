package core

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"cardflip/golang/storage"
)

// v1Deck is a deck as snapshot version 1 stored it: with its own options.
type v1Deck struct {
	ID      ID          `json:"id"`
	Name    string      `json:"name"`
	Options DeckOptions `json:"options"`
}

// v1Store returns a store holding a version 1 collection: c's settings and
// note types with the decks given (their options included), and journal ops
// changing decks after the snapshot. Decks are v1Decks or their JSON.
func v1Store(t *testing.T, c *Collection, decks any, journal ...[]v1Deck) *storage.MemoryStore {
	t.Helper()
	c.mu.RLock()
	snap, err := json.Marshal(c.state())
	c.mu.RUnlock()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(snap, &m); err != nil {
		t.Fatal(err)
	}
	m["version"] = 1
	m["decks"] = decks
	delete(m, "presets")
	delete(m, "notes")
	delete(m, "cards")
	if snap, err = json.Marshal(m); err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemoryStore()
	st, _, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Init(snap); err != nil {
		t.Fatal(err)
	}
	for _, ds := range journal {
		raw, _ := json.Marshal(map[string]any{"k": "deckOptions", "d": ds})
		if err := st.Commit(raw, nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	return store
}

func TestMigrateVersion1Collection(t *testing.T) {
	src, _, clk := testCollection(t)
	def := legacyDeckOptions()       // the defaults of version 1
	biology := deckOpts(5, 200, 0.9) // only the new card limit differs
	exam := deckOpts(50, 500, 0.95)
	exam.LearningSteps = []int{60}
	store := v1Store(t, src, []v1Deck{
		{ID: 1000, Name: "Default", Options: exam},
		{ID: 1001, Name: "Biology", Options: biology},
		{ID: 1002, Name: "Chemistry", Options: biology},
		{ID: 1003, Name: "Art", Options: def},
		{ID: 1004, Name: "Physics", Options: def},
	}, []v1Deck{{ID: 1004, Name: "Physics", Options: exam}}, []v1Deck{{ID: 1005, Name: "Zoology", Options: biology}})

	c, fresh, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil || fresh {
		t.Fatalf("open: %v fresh=%v", err, fresh)
	}
	// Existing options kept (spec): every deck studies with the options it
	// had, the journal's changes included, except that the Default preset
	// no longer buries siblings (snapshot version 3).
	want := map[string]DeckOptions{"Default": exam, "Biology": biology, "Chemistry": biology,
		"Art": DefaultDeckOptions(), "Physics": exam, "Zoology": biology}
	presetName := map[string]string{"Default": "Default (2)", "Biology": "Biology", "Chemistry": "Biology",
		"Art": "Default", "Physics": "Default (2)", "Zoology": "Biology"}
	names := map[ID]string{}
	for _, p := range c.Presets() {
		names[p.ID] = p.Name
	}
	for _, d := range c.Decks() {
		if got := deckOptions(c, d.ID); !got.Equal(want[d.Name]) {
			t.Errorf("%s: options = %+v, want %+v", d.Name, got, want[d.Name])
		}
		if names[d.PresetID] != presetName[d.Name] {
			t.Errorf("%s: preset %q, want %q", d.Name, names[d.PresetID], presetName[d.Name])
		}
	}
	if n := len(c.Presets()); n != 3 {
		t.Errorf("%d presets, want 3", n)
	}
	// The migration is saved as a current snapshot.
	_, loaded, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(loaded.Snapshot, &s); err != nil || s.Version != snapshotVersion || len(loaded.Ops) != 0 {
		t.Fatalf("snapshot version %d, %d ops (%v)", s.Version, len(loaded.Ops), err)
	}
	reopen(t, c, store, clk)
}

// Property (design, Risks): for random collections, each deck's options
// after the migration equal its options before, decks with equal options
// share a preset, and decks with the default options of version 1 use
// Default.
func TestMigratePresetsKeepsOptions(t *testing.T) {
	def := legacyDeckOptions()
	pool := []DeckOptions{def, deckOpts(5, 200, 0.9), deckOpts(20, 100, 0.85), deckOpts(0, 0, 0.7)}
	steps := def
	steps.LearningSteps = nil
	pool = append(pool, steps)
	random := def
	random.InsertionOrder = InsertRandom
	pool = append(pool, random)
	names := []string{"Default", "Default (2)", "A", "a::b", "B", "C::D", "C"}
	for seed := range uint64(500) {
		r := rand.New(rand.NewPCG(seed, 1))
		var decks []*Deck
		before := map[ID]DeckOptions{}
		used := map[string]bool{}
		for i := range 1 + r.IntN(len(names)) {
			name := names[r.IntN(len(names))]
			if used[name] {
				name = fmt.Sprintf("%s %d", name, i)
			}
			used[name] = true
			opts := pool[r.IntN(len(pool))]
			d := &Deck{ID: ID(100 + i), Name: name, legacy: &opts}
			decks = append(decks, d)
			before[d.ID] = opts
		}
		next := ID(1000)
		presets, migrated := migratePresets(decks, false, func() ID { next++; return next })
		c := &Collection{presets: map[ID]*Preset{}, decks: map[ID]*Deck{}}
		for _, p := range presets {
			if c.presets[p.ID] != nil {
				t.Fatalf("seed %d: duplicate preset ID %d", seed, p.ID)
			}
			c.presets[p.ID] = p
		}
		if c.presets[DefaultPresetID] == nil || !c.presets[DefaultPresetID].Options.Equal(def) {
			t.Fatalf("seed %d: no Default preset", seed)
		}
		seen := map[string]bool{}
		for _, p := range presets {
			if seen[p.Name] {
				t.Fatalf("seed %d: preset name %q twice", seed, p.Name)
			}
			seen[p.Name] = true
		}
		for _, d := range migrated {
			if d.legacy != nil || c.presets[d.PresetID] == nil {
				t.Fatalf("seed %d: deck %s not migrated", seed, d.Name)
			}
			c.decks[d.ID] = d
		}
		if len(migrated) != len(decks) {
			t.Fatalf("seed %d: %d of %d decks migrated", seed, len(migrated), len(decks))
		}
		for _, a := range migrated {
			if got := c.optionsOf(a.ID); !got.Equal(before[a.ID]) {
				t.Fatalf("seed %d: %s options = %+v, want %+v", seed, a.Name, got, before[a.ID])
			}
			if before[a.ID].Equal(def) != (a.PresetID == DefaultPresetID) {
				t.Fatalf("seed %d: %s uses preset %d", seed, a.Name, a.PresetID)
			}
			for _, b := range migrated {
				if before[a.ID].Equal(before[b.ID]) != (a.PresetID == b.PresetID) {
					t.Fatalf("seed %d: %s and %s share presets wrongly", seed, a.Name, b.Name)
				}
			}
		}
		for _, d := range decks {
			if d.legacy == nil || d.PresetID != 0 {
				t.Fatalf("seed %d: input deck changed", seed)
			}
		}
	}
}

// A version 1 deck stored without the burying options buried siblings, as
// the defaults of that time did: it keeps burying in a preset of its own,
// while decks with the old defaults use Default, which no longer buries.
func TestMigrateVersion1BuryingOptions(t *testing.T) {
	src, _, clk := testCollection(t)
	store := v1Store(t, src, []json.RawMessage{
		json.RawMessage(`{"id":1000,"name":"Default"}`),
		json.RawMessage(`{"id":1001,"name":"Art","options":{"newPerDay":20,"reviewsPerDay":200,"desiredRetention":0.9}}`),
		json.RawMessage(`{"id":1002,"name":"Biology","options":{"newPerDay":5,"reviewsPerDay":200,"desiredRetention":0.9}}`),
	})
	c, _, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	names := map[ID]string{}
	for _, p := range c.Presets() {
		names[p.ID] = p.Name
	}
	want := map[string]struct {
		preset string
		bury   bool
	}{"Default": {"Default", false}, "Art": {"Default", false}, "Biology": {"Biology", true}}
	for _, d := range c.Decks() {
		o := deckOptions(c, d.ID)
		if w := want[d.Name]; names[d.PresetID] != w.preset || o.BuryNew != w.bury || o.BuryReview != w.bury {
			t.Errorf("%s: preset %q, options %+v", d.Name, names[d.PresetID], o)
		}
	}
	reopen(t, c, store, clk)
}

// v2Snapshot returns a snapshot of a version 2 collection whose Default
// preset buries new and review siblings, as it did by default then, and
// whose preset "Own Exam", used by deck Exam, buries them too.
func v2Snapshot(t *testing.T) []byte {
	t.Helper()
	c, _, _ := testCollection(t)
	exam, _ := c.CreateDeck("Exam")
	ownPreset(t, c, exam)
	burySiblings(t, c, defaultDeck(c))
	burySiblings(t, c, exam)
	snap, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(snap, &m); err != nil {
		t.Fatal(err)
	}
	m["version"] = 2
	if snap, err = json.Marshal(m); err != nil {
		t.Fatal(err)
	}
	return snap
}

// checkV2Burying checks c, opened or restored from v2Snapshot: Default
// buries nothing and "Own Exam" still buries both kinds of siblings.
func checkV2Burying(t *testing.T, c *Collection) {
	t.Helper()
	presets := c.Presets()
	if len(presets) != 2 {
		t.Fatalf("presets = %+v", presets)
	}
	for _, p := range presets {
		bury := p.ID != DefaultPresetID
		if p.Options.BuryNew != bury || p.Options.BuryReview != bury {
			t.Errorf("%s: options %+v", p.Name, p.Options)
		}
	}
}

// Existing collection (collection spec): opening a version 2 collection
// turns burying off in its Default preset once; other presets keep it.
func TestMigrateVersion2Burying(t *testing.T) {
	_, _, clk := testCollection(t)
	store := storage.NewMemoryStore()
	st, _, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Init(v2Snapshot(t)); err != nil {
		t.Fatal(err)
	}
	c, fresh, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil || fresh {
		t.Fatalf("open: %v fresh=%v", err, fresh)
	}
	checkV2Burying(t, c)
	// The migration is saved as a current snapshot, so it runs once: burying
	// turned back on in Default stays on.
	_, loaded, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(loaded.Snapshot, &s); err != nil || s.Version != snapshotVersion || len(loaded.Ops) != 0 {
		t.Fatalf("snapshot version %d, %d ops (%v)", s.Version, len(loaded.Ops), err)
	}
	burySiblings(t, c, defaultDeck(c))
	c = reopen(t, c, store, clk)
	if o := c.preset(DefaultPresetID).Options; !o.BuryNew || !o.BuryReview {
		t.Fatalf("Default after reopen = %+v", o)
	}
}

// Older backup (collection spec): restoring a version 2 backup turns
// burying off in its Default preset; other presets keep it.
func TestRestoreVersion2Burying(t *testing.T) {
	c, store, clk := testCollection(t)
	if err := c.RestoreKeepingMedia(v2Snapshot(t), nil, nil); err != nil {
		t.Fatal(err)
	}
	checkV2Burying(t, c)
	checkV2Burying(t, reopen(t, c, store, clk))
}
