package backup

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/importer/apkg"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

var now = time.Date(2026, 10, 13, 12, 0, 0, 0, time.Local)

func open(t *testing.T, store *storage.MemoryStore) *core.Collection {
	t.Helper()
	c, _, err := core.Open(store.Open(), core.Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// populated returns a collection with decks, notes of every built-in type,
// imported media, settings, answers, an undone answer and a deleted note.
func populated(t *testing.T) *core.Collection {
	t.Helper()
	c := open(t, storage.NewMemoryStore())
	data, err := os.ReadFile("../importer/apkg/testdata/modern.apkg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := apkg.Import(c, data, apkg.Options{KeepScheduling: true}, nil, nil); err != nil {
		t.Fatal(err)
	}
	deck, _ := c.CreateDeck("Languages::Spanish")
	c.SetDeckOptions(deck, core.DeckOptions{NewPerDay: 7, ReviewsPerDay: 70, DesiredRetention: 0.85})
	var basic, cloze core.ID
	for _, nt := range c.NoteTypes() {
		switch nt.Name {
		case "Basic":
			basic = nt.ID
		case "Cloze":
			cloze = nt.ID
		}
	}
	n1, _ := c.AddNote(core.NoteInput{TypeID: basic, DeckID: deck, Fields: []string{"hola", "hello"}, Tags: []string{"greeting"}})
	cz, _ := c.AddNote(core.NoteInput{TypeID: cloze, DeckID: deck, Fields: []string{"{{c1::Madrid}} is in {{c2::Spain}}", ""}})
	czCards := c.CardsOfNote(cz)
	c.SetFlag([]core.ID{czCards[0].ID}, 3)
	c.SetFlag([]core.ID{czCards[1].ID}, 7)
	gone, _ := c.AddNote(core.NoteInput{TypeID: basic, DeckID: deck, Fields: []string{"adiós", "bye"}, Tags: []string{"farewell"}})
	c.DeleteNotes([]core.ID{gone})
	card := c.CardsOfNote(n1)[0]
	c.Answer(card.ID, scheduler.Good, 4*time.Second)
	e, _ := c.Answer(card.ID, scheduler.Again, 2*time.Second)
	c.UndoReview(e.ID)
	s := c.Settings()
	s.Theme = core.ThemeDark
	s.RolloverHour = 6
	s.FlagNames[0] = "Fix later"
	c.UpdateSettings(s)
	return c
}

func revlogOf(t *testing.T, c *core.Collection) []string {
	t.Helper()
	months, _ := c.RevlogMonths()
	var out []string
	for _, m := range months {
		entries, _ := c.ReadRevlog(m)
		for _, e := range entries {
			out = append(out, string(e))
		}
	}
	return out
}

func deletionsOf(t *testing.T, c *core.Collection) []string {
	t.Helper()
	entries, err := c.ReadDeletionLog()
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, string(e))
	}
	return out
}

func mediaOf(t *testing.T, c *core.Collection) map[string]string {
	t.Helper()
	out := map[string]string{}
	c.EachMedia(func(name string, data []byte) error {
		out[name] = string(data)
		return nil
	})
	return out
}

func TestRoundTrip(t *testing.T) {
	orig := populated(t)
	data, err := Export(orig, "0.1.0", now, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Round trip (spec): restore on another device; everything matches.
	other := open(t, storage.NewMemoryStore())
	other.AddNote(core.NoteInput{TypeID: other.NoteTypes()[0].ID, DeckID: other.Decks()[0].ID, Fields: []string{"to be replaced"}})
	changes, cancel := other.Subscribe()
	defer cancel()
	a, err := Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if a.Manifest.Format != Format || a.Manifest.AppVersion != "0.1.0" || !a.Manifest.CreatedAt.Equal(now) {
		t.Fatalf("manifest = %+v", a.Manifest)
	}
	if err := other.Restore(a.Collection, a.Revlog, a.Deletions, a.Media); err != nil {
		t.Fatal(err)
	}
	if ch := <-changes; !ch.All {
		t.Fatalf("change = %+v", ch)
	}
	want, _ := orig.Snapshot()
	got, _ := other.Snapshot()
	if !restoredEqual(t, want, got) {
		t.Fatalf("collection differs after restore:\n%s\n%s", want, got)
	}
	// One answer, and the 24 entries of history the package import added.
	if !reflect.DeepEqual(revlogOf(t, orig), revlogOf(t, other)) || len(revlogOf(t, other)) != 25 {
		t.Fatalf("review log differs: %v vs %v", revlogOf(t, orig), revlogOf(t, other))
	}
	if !reflect.DeepEqual(mediaOf(t, orig), mediaOf(t, other)) || len(mediaOf(t, other)) != 4 {
		t.Fatal("media differs")
	}
	if log, _ := other.DeletionLog(); !reflect.DeepEqual(deletionsOf(t, orig), deletionsOf(t, other)) ||
		len(log) != 1 || log[0].Fields[0] != "adiós" || log[0].Tags[0] != "farewell" {
		t.Fatalf("deletion log differs: %v vs %v", deletionsOf(t, orig), deletionsOf(t, other))
	}
	if !reflect.DeepEqual(orig.DeckCounts(), other.DeckCounts()) {
		t.Fatal("due counts differ")
	}
	if s := other.Settings(); s.Theme != core.ThemeDark || s.RolloverHour != 6 || s.FlagNames[0] != "Fix later" {
		t.Fatalf("settings = %+v", s)
	}
	ids, _ := other.Search("", core.ColCreated, false)
	// Flag kept across restore (spec).
	flagged := 0
	for _, id := range ids {
		a, _ := orig.Card(id)
		b, _ := other.Card(id)
		if a.Flag != b.Flag {
			t.Fatalf("card %d has flag %d, want %d", id, b.Flag, a.Flag)
		}
		if b.Flag != 0 {
			flagged++
		}
	}
	if flagged != 3 { // the two cloze cards and the imported green card
		t.Fatalf("%d flagged cards after restore, want 3", flagged)
	}
	for _, id := range ids {
		a, _ := orig.RenderCard(id, render.Env{})
		b, _ := other.RenderCard(id, render.Env{})
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("card %d renders differently", id)
		}
	}
	// The restored collection survives a restart.
	if _, err := other.Snapshot(); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreSurvivesRestart(t *testing.T) {
	orig := populated(t)
	data, _ := Export(orig, "0.1.0", now, nil)
	store := storage.NewMemoryStore()
	c := open(t, store)
	a, _ := Read(data)
	if err := c.Restore(a.Collection, a.Revlog, a.Deletions, a.Media); err != nil {
		t.Fatal(err)
	}
	reopened := open(t, store)
	want, _ := orig.Snapshot()
	got, _ := reopened.Snapshot()
	if !restoredEqual(t, want, got) || !reflect.DeepEqual(mediaOf(t, orig), mediaOf(t, reopened)) ||
		!reflect.DeepEqual(deletionsOf(t, orig), deletionsOf(t, reopened)) {
		t.Fatal("restored collection not persisted")
	}
}

func jsonEqual(t *testing.T, a, b []byte) bool {
	t.Helper()
	var x, y any
	json.Unmarshal(a, &x)
	json.Unmarshal(b, &y)
	return reflect.DeepEqual(x, y)
}

// restoredEqual compares a snapshot with one restored from it, apart from
// the lineage: a restored collection starts a new one (sync design decision
// 7).
func restoredEqual(t *testing.T, orig, restored []byte) bool {
	t.Helper()
	var x, y map[string]any
	json.Unmarshal(orig, &x)
	json.Unmarshal(restored, &y)
	if x["epoch"] == nil || x["epoch"] == y["epoch"] {
		t.Errorf("restored collection has epoch %v, the original %v", y["epoch"], x["epoch"])
	}
	delete(x, "epoch")
	delete(y, "epoch")
	return reflect.DeepEqual(x, y)
}

func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, _ := zw.Create(name)
		w.Write([]byte(content))
	}
	zw.Close()
	return buf.Bytes()
}

func TestRestoreFailuresKeepCollection(t *testing.T) {
	good, _ := Export(populated(t), "0.1.0", now, nil)
	a, _ := Read(good)
	cases := []struct {
		name string
		data []byte
		want error
	}{
		// Wrong file (spec): a zip that is not a Cardflip backup.
		{"other zip", zipOf(t, map[string]string{"photo.jpg": "x"}), ErrNotBackup},
		{"not a zip", []byte("hello"), ErrNotBackup},
		// Newer backup (spec): refused, asking to update Cardflip.
		{"newer format", zipOf(t, map[string]string{"manifest.json": `{"format":3,"appVersion":"9.0"}`,
			"collection.json": string(a.Collection), "revlog.jsonl": ""}), ErrNewerFormat},
		{"missing collection", zipOf(t, map[string]string{"manifest.json": `{"format":1}`, "revlog.jsonl": ""}), ErrDamaged},
		{"truncated", good[:len(good)/2], ErrNotBackup},
	}
	for _, c := range cases {
		_, err := Read(c.data)
		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
	if !strings.Contains(ErrNotBackup.Error(), "not a Cardflip backup") || !strings.Contains(ErrNewerFormat.Error(), "update Cardflip") {
		t.Fatal("error messages do not say why")
	}

	// Content that passes Read but not validation leaves the collection as
	// it was.
	store := storage.NewMemoryStore()
	c := open(t, store)
	c.CreateDeck("Mine")
	before, _ := c.Snapshot()
	for name, collection := range map[string]string{
		"bad json":       `{"version":1,"decks":`,
		"no decks":       `{"version":1,"decks":[]}`,
		"newer snapshot": `{"version":99,"decks":[{"id":1,"name":"D"}]}`,
		"dangling card":  `{"version":1,"decks":[{"id":1,"name":"D"}],"cards":[{"id":5,"note":77,"deck":1}]}`,
	} {
		err := c.Restore([]byte(collection), nil, nil, nil)
		var re *core.RestoreError
		if !errors.As(err, &re) {
			t.Fatalf("%s: err = %v", name, err)
		}
		after, _ := c.Snapshot()
		if !bytes.Equal(before, after) {
			t.Fatalf("%s: collection changed", name)
		}
	}
	if err := c.Restore(a.Collection, []json.RawMessage{json.RawMessage(`{"nope":1}`)}, nil, nil); err == nil {
		t.Fatal("bad review log accepted")
	}
	if err := c.Restore(a.Collection, nil, []json.RawMessage{json.RawMessage(`{"nope":1}`)}, nil); err == nil {
		t.Fatal("bad deletion log accepted")
	}
	bad := zipOf(t, map[string]string{"manifest.json": `{"format":1}`, "collection.json": string(a.Collection),
		"revlog.jsonl": "", "deleted.jsonl": "{nope\n"})
	if _, err := Read(bad); !errors.Is(err, ErrDamaged) {
		t.Fatalf("damaged deleted.jsonl: err = %v", err)
	}
	reopened := open(t, store)
	after, _ := reopened.Snapshot()
	if !jsonEqual(t, before, after) {
		t.Fatal("failed restores changed stored data")
	}
}

func TestFileName(t *testing.T) {
	if got := FileName(time.Date(2026, 10, 5, 23, 0, 0, 0, time.Local)); got != "cardflip-2026-10-05.cardflip" {
		t.Fatalf("name = %q", got)
	}
}

// Backups made before the deletion log have no deleted.jsonl; restoring one
// leaves an empty log.
func TestRestoreBackupWithoutDeletionLog(t *testing.T) {
	orig := populated(t)
	a0, _ := Read(mustExport(t, orig))
	old := zipOf(t, map[string]string{"manifest.json": `{"format":1}`, "collection.json": string(a0.Collection), "revlog.jsonl": ""})
	a, err := Read(old)
	if err != nil {
		t.Fatal(err)
	}
	c := open(t, storage.NewMemoryStore())
	c.DeleteNotes([]core.ID{addTo(t, c)})
	if err := c.Restore(a.Collection, a.Revlog, a.Deletions, a.Media); err != nil {
		t.Fatal(err)
	}
	if got := deletionsOf(t, c); len(got) != 0 {
		t.Fatalf("deletion log = %v", got)
	}
}

func mustExport(t *testing.T, c *core.Collection) []byte {
	t.Helper()
	data, err := Export(c, "0.1.0", now, nil)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func addTo(t *testing.T, c *core.Collection) core.ID {
	t.Helper()
	id, err := c.AddNote(core.NoteInput{TypeID: c.NoteTypes()[0].ID, DeckID: c.Decks()[0].ID, Fields: []string{"x"}})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// Backups made before presets (format 1) store options per deck; restoring
// one keeps every deck's options, now in presets.
func TestRestoreVersion1Backup(t *testing.T) {
	collection := `{"version":1,"nextId":1700000000010,
		"settings":{"theme":"system","rolloverHour":4},
		"decks":[
			{"id":1700000000001,"name":"Default","options":{"newPerDay":20,"reviewsPerDay":200,"desiredRetention":0.9}},
			{"id":1700000000002,"name":"Biology","options":{"newPerDay":5,"reviewsPerDay":200,"desiredRetention":0.9}},
			{"id":1700000000003,"name":"Biology::Cells","options":{"newPerDay":5,"reviewsPerDay":200,"desiredRetention":0.9}},
			{"id":1700000000004,"name":"History","options":{"newPerDay":20,"reviewsPerDay":200,"desiredRetention":0.85,"maxInterval":365}}
		],
		"noteTypes":[{"id":1700000000005,"name":"Basic","kind":"standard","fields":["Front","Back"],
			"templates":[{"name":"Card 1","qfmt":"{{Front}}","afmt":"{{Back}}"}],"css":""}],
		"notes":[{"id":1700000000006,"type":1700000000005,"guid":"g1","fields":["mitochondria","powerhouse"],"mod":"2026-01-01T00:00:00Z"}],
		"cards":[{"id":1700000000007,"note":1700000000006,"deck":1700000000003,"ord":0,"pos":1,"state":0}]}`
	a, err := Read(zipOf(t, map[string]string{"manifest.json": `{"format":1}`, "collection.json": collection, "revlog.jsonl": ""}))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.NewMemoryStore()
	c := open(t, store)
	if err := c.Restore(a.Collection, a.Revlog, a.Deletions, a.Media); err != nil {
		t.Fatal(err)
	}
	check := func(c *core.Collection) {
		t.Helper()
		presets := map[core.ID]string{}
		for _, p := range c.Presets() {
			presets[p.ID] = p.Name
		}
		want := map[string]struct {
			newPerDay, maxInterval int
			retention              float64
			preset                 string
		}{
			"Default":        {20, core.MaxIntervalDays, 0.9, "Default"},
			"Biology":        {5, core.MaxIntervalDays, 0.9, "Biology"},
			"Biology::Cells": {5, core.MaxIntervalDays, 0.9, "Biology"},
			"History":        {20, 365, 0.85, "History"},
		}
		for _, d := range c.Decks() {
			o, _ := c.DeckOptions(d.ID)
			w := want[d.Name]
			if o.NewPerDay != w.newPerDay || o.MaxInterval != w.maxInterval || o.DesiredRetention != w.retention || presets[d.PresetID] != w.preset {
				t.Errorf("%s: options %+v, preset %q", d.Name, o, presets[d.PresetID])
			}
		}
		if len(presets) != 3 {
			t.Errorf("presets = %v", presets)
		}
	}
	check(c)
	// The restored collection is saved migrated.
	check(open(t, store))
}
