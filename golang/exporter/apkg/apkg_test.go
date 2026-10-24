package apkg

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"cardflip/golang/core"
	importer "cardflip/golang/importer/apkg"
	"cardflip/golang/importer/apkg/sqlite"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// The clock is the importer tests' (2026-10-13 12:00 UTC), so the
// fixtures' review cards are due in the future.
var testNow = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC).In(time.Local)

func newCollection(t *testing.T) *core.Collection {
	t.Helper()
	c, _, err := core.Open(storage.NewMemory(), core.Options{Now: func() time.Time { return testNow }})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// fixtureCollection is a collection with the importer's modern.apkg
// fixture imported with its scheduling: decks with presets, every card
// state, flags, history and media.
func fixtureCollection(t *testing.T) *core.Collection {
	t.Helper()
	data, err := os.ReadFile("../../importer/apkg/testdata/modern.apkg")
	if err != nil {
		t.Fatal(err)
	}
	c := newCollection(t)
	if _, err := importer.Import(c, data, importer.Options{KeepScheduling: true}, nil, nil); err != nil {
		t.Fatal(err)
	}
	return c
}

func export(t *testing.T, c *core.Collection, scope core.ExportScope, o Options) []byte {
	t.Helper()
	set, err := c.Export(scope)
	if err != nil {
		t.Fatal(err)
	}
	reviews, err := c.ExportReviews(set.Cards)
	if err != nil {
		t.Fatal(err)
	}
	data, err := Write(Input{Set: set, Reviews: reviews, Media: c.Media, Now: testNow}, o, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func deckID(t *testing.T, c *core.Collection, name string) core.ID {
	t.Helper()
	for _, d := range c.Decks() {
		if d.Name == name {
			return d.ID
		}
	}
	t.Fatalf("no deck %s", name)
	return 0
}

type cardView struct {
	Deck                    string
	State                   scheduler.State
	Due                     string
	Ivl, Step, Reps, Lapses int
	Interday, Susp          bool
	Flag                    int
	S, D                    float64
	Reviews                 []string
}

type noteView struct {
	Type, CSS string
	Fields    []string
	Tags      []string
	Cards     map[int]cardView
}

// view summarises a collection by note GUID, with what an export keeps.
func view(t *testing.T, c *core.Collection) map[string]noteView {
	t.Helper()
	set, err := c.Export(core.ExportScope{All: true})
	if err != nil {
		t.Fatal(err)
	}
	reviews, err := c.ExportReviews(set.Cards)
	if err != nil {
		t.Fatal(err)
	}
	decks := map[core.ID]string{}
	for _, d := range set.Decks {
		decks[d.ID] = d.Name
	}
	types := map[core.ID]core.NoteType{}
	for _, nt := range set.NoteTypes {
		types[nt.ID] = nt
	}
	byCard := map[core.ID][]string{}
	for _, r := range reviews {
		byCard[r.CardID] = append(byCard[r.CardID], fmt.Sprintf("%s r%d %v %d→%d %dms %s",
			r.Time.UTC().Format(time.DateTime), r.Rating, r.StateBefore, r.IvlBefore, r.IvlAfter, r.DurationMs, r.Source))
	}
	notes := map[core.ID]string{}
	out := map[string]noteView{}
	for _, n := range set.Notes {
		nt := types[n.TypeID]
		out[n.GUID] = noteView{Type: nt.Name, CSS: nt.CSS, Fields: n.Fields, Tags: n.Tags, Cards: map[int]cardView{}}
		notes[n.ID] = n.GUID
	}
	for _, card := range set.Cards {
		cv := cardView{Deck: decks[card.DeckID], State: card.State, Ivl: card.ScheduledDays, Step: card.Step,
			Reps: card.Reps, Lapses: card.Lapses, Interday: card.Interday, Susp: card.Suspended, Flag: card.Flag,
			Reviews: byCard[card.ID]}
		switch {
		case card.Due.IsZero():
		case card.State == scheduler.Review || card.Interday:
			// Anki stores the day, which the importer turns into a time
			// in it.
			cv.Due = fmt.Sprint("day ", scheduler.DayNumber(card.Due, set.Settings.RolloverHour))
		default:
			cv.Due = card.Due.UTC().Format(time.DateTime)
		}
		if card.Memory != nil {
			cv.S, cv.D = math.Round(card.Memory.Stability*100)/100, math.Round(card.Memory.Difficulty*100)/100
		}
		out[notes[card.NoteID]].Cards[card.Ord] = cv
	}
	return out
}

func diffViews(t *testing.T, got, want map[string]noteView) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%d notes, want %d", len(got), len(want))
	}
	for guid, w := range want {
		g, ok := got[guid]
		if !ok {
			t.Errorf("note %s missing", guid)
			continue
		}
		gj, _ := json.MarshalIndent(g, "", " ")
		wj, _ := json.MarshalIndent(w, "", " ")
		if !bytes.Equal(gj, wj) {
			t.Errorf("note %s:\ngot  %s\nwant %s", guid, gj, wj)
		}
	}
}

func reimport(t *testing.T, data []byte, keep bool) *core.Collection {
	t.Helper()
	c := newCollection(t)
	if _, err := importer.Import(c, data, importer.Options{KeepScheduling: keep}, nil, nil); err != nil {
		t.Fatalf("import of the export: %v", err)
	}
	return c
}

func TestRoundTripWithScheduling(t *testing.T) {
	src := fixtureCollection(t)
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy=%v", legacy), func(t *testing.T) {
			data := export(t, src, core.ExportScope{All: true}, Options{Scheduling: true, Presets: true, Media: true, Legacy: legacy})
			dst := reimport(t, data, true)
			diffViews(t, view(t, dst), view(t, src))

			// Presets and media come along.
			spanish, _ := dst.DeckOptions(deckID(t, dst, "Fixture::Spanish"))
			want, _ := src.DeckOptions(deckID(t, src, "Fixture::Spanish"))
			if !spanish.Equal(want) {
				t.Errorf("Spanish options = %+v, want %+v", spanish, want)
			}
			for _, name := range []string{"dog.png", "cat.jpg", "hola.mp3", "adios.mp3"} {
				a, _ := src.Media(name)
				b, err := dst.Media(name)
				if err != nil || !bytes.Equal(a, b) {
					t.Errorf("media %s: %v", name, err)
				}
			}
		})
	}
}

func TestRoundTripWithoutScheduling(t *testing.T) {
	src := fixtureCollection(t)
	data := export(t, src, core.ExportScope{DeckID: deckID(t, src, "Fixture")}, Options{Media: true})
	p, err := importer.Read(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Revlog) != 0 {
		t.Errorf("%d review log entries exported without scheduling", len(p.Revlog))
	}
	for _, c := range p.Cards {
		if c.Type != 0 || c.Queue != 0 || c.Flag != 0 || c.Reps != 0 {
			t.Errorf("card %d is not new: %+v", c.ID, c)
		}
	}
	for _, d := range p.Decks {
		if d.PresetID != 1 {
			t.Errorf("deck %s uses preset %d without presets", d.Name, d.PresetID)
		}
	}
	// Imported with scheduling, every card is still new.
	dst := reimport(t, data, true)
	want := view(t, src)
	for guid, n := range want {
		for ord, cv := range n.Cards {
			n.Cards[ord] = cardView{Deck: cv.Deck}
		}
		want[guid] = n
	}
	diffViews(t, view(t, dst), want)
}

func TestDeckExportContainsTheDeckTree(t *testing.T) {
	src := fixtureCollection(t)
	data := export(t, src, core.ExportScope{DeckID: deckID(t, src, "Fixture::Spanish")}, Options{Media: true})
	p, err := importer.Read(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	decks := map[int64]string{}
	for id, d := range p.Decks {
		decks[id] = d.Name
	}
	for _, c := range p.Cards {
		if decks[c.DeckID] != "Fixture::Spanish" {
			t.Errorf("card %d in deck %q", c.ID, decks[c.DeckID])
		}
	}
	var media []string
	for name := range p.Media {
		media = append(media, name)
	}
	sort.Strings(media)
	// Spanish has the image and audio notes, Geo none of them.
	if strings.Join(media, ",") != "adios.mp3,cat.jpg,dog.png,hola.mp3" {
		t.Errorf("media = %v", media)
	}
	if data := export(t, src, core.ExportScope{DeckID: deckID(t, src, "Fixture::Spanish")}, Options{}); len(readZip(t, data)["media"]) != 2 {
		t.Errorf("media map without media = %q", readZip(t, data)["media"])
	}
}

func TestSelectedNotesWithAllTheirCards(t *testing.T) {
	src := fixtureCollection(t)
	ids, _ := src.Search("", core.ColCreated, false)
	var noteIDs []core.ID
	seen := map[core.ID]bool{}
	for _, row := range src.Rows(ids) {
		if n, _ := src.Note(row.NoteID); !seen[n.ID] && strings.HasPrefix(n.GUID, "cf-cloze") {
			seen[n.ID] = true
			noteIDs = append(noteIDs, n.ID)
		}
	}
	p, err := importer.Read(export(t, src, core.ExportScope{NoteIDs: noteIDs}, Options{}), nil)
	if err != nil {
		t.Fatal(err)
	}
	// cf-cloze-paris has 2 cards, cf-cloze-multi 3.
	if len(p.Notes) != 2 || len(p.Cards) != 5 {
		t.Errorf("%d notes and %d cards, want 2 and 5", len(p.Notes), len(p.Cards))
	}
}

func readZip(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	for _, f := range zr.File {
		rc, _ := f.Open()
		out[f.Name], _ = io.ReadAll(rc)
		rc.Close()
	}
	return out
}

func TestPackageLayout(t *testing.T) {
	src := fixtureCollection(t)
	modern := readZip(t, export(t, src, core.ExportScope{All: true}, Options{Media: true}))
	if modern["meta"] == nil || !bytes.Equal(modern["meta"], []byte{0x08, 2}) {
		t.Errorf("meta = %x", modern["meta"])
	}
	if modern["collection.anki21"] == nil || modern["collection.anki2"] == nil {
		t.Fatal("collection.anki21 and the dummy collection.anki2 expected")
	}
	db, err := sqlite.Open(modern["collection.anki2"])
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Scan("notes", func(r sqlite.Row) error {
		if !strings.HasPrefix(r.Str("flds"), "Please update") {
			t.Errorf("dummy note = %q", r.Str("flds"))
		}
		return nil
	})
	var m map[string]string
	if err := json.Unmarshal(modern["media"], &m); err != nil || len(m) != 4 {
		t.Errorf("media map = %s", modern["media"])
	}
	for key, name := range m {
		if data, _ := src.Media(name); !bytes.Equal(modern[key], data) {
			t.Errorf("zip entry %s is not %s", key, name)
		}
	}

	legacy := readZip(t, export(t, src, core.ExportScope{All: true}, Options{Legacy: true}))
	if legacy["collection.anki21"] != nil || legacy["collection.anki2"] == nil || !bytes.Equal(legacy["meta"], []byte{0x08, 1}) {
		t.Errorf("legacy package has %v", keys(legacy))
	}
}

func keys[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// jsonKeys returns the keys of the JSON objects in a col column (a map of
// ID to object), merged over all objects.
func jsonKeys(t *testing.T, raw string, nested bool) []string {
	t.Helper()
	var objs map[string]map[string]any
	if !nested {
		var obj map[string]any
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			t.Fatal(err)
		}
		objs = map[string]map[string]any{"": obj}
	} else if err := json.Unmarshal([]byte(raw), &objs); err != nil {
		t.Fatal(err)
	}
	set := map[string]bool{}
	for _, o := range objs {
		for k := range o {
			set[k] = true
		}
	}
	return keys(set)
}

func colRow(t *testing.T, collection []byte) sqlite.Row {
	t.Helper()
	db, err := sqlite.Open(collection)
	if err != nil {
		t.Fatal(err)
	}
	var row sqlite.Row
	_ = db.Scan("col", func(r sqlite.Row) error { row = r; return nil })
	return row
}

// The col JSON has the keys of an Anki-exported package's (Anki 26.09
// legacy export), except Anki-only ones that are optional when read.
func TestColumnJSONMatchesAnki(t *testing.T) {
	ref, err := os.ReadFile("../../importer/apkg/testdata/legacy.apkg")
	if err != nil {
		t.Fatal(err)
	}
	want := colRow(t, readZip(t, ref)["collection.anki21"])
	got := colRow(t, readZip(t, export(t, fixtureCollection(t), core.ExportScope{All: true}, Options{Presets: true}))["collection.anki21"])
	optional := map[string]bool{
		// Random IDs of fields/templates and stock kinds are filled in by
		// Anki when missing.
		"originalStockKind": true,
	}
	for _, col := range []struct {
		name   string
		nested bool
	}{{"conf", false}, {"models", true}, {"decks", true}, {"dconf", true}} {
		g, w := jsonKeys(t, got.Str(col.name), col.nested), jsonKeys(t, want.Str(col.name), col.nested)
		gs := map[string]bool{}
		for _, k := range g {
			gs[k] = true
		}
		for _, k := range w {
			if !gs[k] && !optional[k] {
				t.Errorf("col.%s lacks key %q", col.name, k)
			}
		}
	}
	if got.Int("ver") != 11 {
		t.Errorf("ver = %d", got.Int("ver"))
	}
}

func TestFileName(t *testing.T) {
	for in, want := range map[string]string{
		"Spanish":            "Spanish.apkg",
		"Languages::Spanish": "Spanish.apkg",
		`a/b:c?`:             "a_b_c_.apkg",
		"  ":                 "export.apkg",
		"Ends with a dot.":   "Ends with a dot.apkg",
	} {
		if got := FileName(in, "apkg"); got != want {
			t.Errorf("FileName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestWriteSamples writes sample packages of the fixture collection into
// the directory $CARDFLIP_EXPORT_SAMPLES, for tool/check_apkg_export.py to
// import into Anki. It does nothing without the variable.
func TestWriteSamples(t *testing.T) {
	dir := os.Getenv("CARDFLIP_EXPORT_SAMPLES")
	if dir == "" {
		t.Skip("CARDFLIP_EXPORT_SAMPLES is not set")
	}
	src := fixtureCollection(t)
	fixture := core.ExportScope{DeckID: deckID(t, src, "Fixture")}
	all := core.ExportScope{All: true}
	full := Options{Scheduling: true, Presets: true, Media: true}
	legacy := full
	legacy.Legacy = true
	for name, s := range map[string]struct {
		scope core.ExportScope
		o     Options
	}{
		"deck.apkg":                {fixture, Options{Media: true}},
		"deck-scheduling.apkg":     {fixture, full},
		"deck-legacy.apkg":         {fixture, legacy},
		"collection.colpkg":        {all, full},
		"collection-legacy.colpkg": {all, legacy},
	} {
		if err := os.WriteFile(dir+"/"+name, export(t, src, s.scope, s.o), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// A note type set to SVG LaTeX images stays so through an export, in both
// formats.
func TestLatexSVGRoundTrip(t *testing.T) {
	data, err := os.ReadFile("../../importer/apkg/testdata/latex.apkg")
	if err != nil {
		t.Fatal(err)
	}
	src := newCollection(t)
	if _, err := importer.Import(src, data, importer.Options{}, nil, nil); err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []bool{false, true} {
		p, err := importer.Read(export(t, src, core.ExportScope{All: true}, Options{Media: true, Legacy: legacy}), nil)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]bool{}
		for _, nt := range p.NoteTypes {
			got[nt.Name] = nt.LatexSVG
		}
		if want := map[string]bool{"Latex PNG": false, "Latex SVG": true}; !reflect.DeepEqual(got, want) {
			t.Errorf("legacy %v: LatexSVG by note type = %v, want %v", legacy, got, want)
		}
	}
}

// Field options and the sort field survive an export, in both formats;
// the editor's default font and size are written as Anki's.
func TestFieldOptionsRoundTrip(t *testing.T) {
	src := newCollection(t)
	var basic core.NoteType
	for _, nt := range src.NoteTypes() {
		if nt.Name == core.StockBasic {
			basic = nt
		}
	}
	if _, err := src.AddNote(core.NoteInput{TypeID: basic.ID, DeckID: src.Decks()[0].ID, Fields: []string{"book", "كتاب"}}); err != nil {
		t.Fatal(err)
	}
	opts := core.FieldOptions{Font: "Amiri", Size: 28, RTL: true, Collapsed: true, ExcludeFromSearch: true, Sticky: true}
	if err := src.SetFieldOptions(basic.ID, 1, opts); err != nil {
		t.Fatal(err)
	}
	if err := src.SetSortField(basic.ID, 1); err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []bool{false, true} {
		p, err := importer.Read(export(t, src, core.ExportScope{All: true}, Options{Legacy: legacy}), nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, nt := range p.NoteTypes {
			if nt.Name != core.StockBasic {
				continue
			}
			want := []importer.FieldConfig{{Font: "Arial", Size: 20},
				{Font: "Amiri", Size: 28, RTL: true, Collapsed: true, ExcludeFromSearch: true, Sticky: true}}
			if nt.SortField != 1 || !reflect.DeepEqual(nt.FieldConfigs, want) {
				t.Errorf("legacy %v: sort field %d, fields %+v", legacy, nt.SortField, nt.FieldConfigs)
			}
		}
	}
}

// Image Occlusion note types are exported with Anki's stock kind, so Anki
// edits their notes with its mask editor.
func TestImageOcclusionStockKind(t *testing.T) {
	src := newCollection(t)
	for _, nt := range src.NoteTypes() {
		fields := map[string][]string{
			core.StockBasic:          {"front", "back"},
			core.StockImageOcclusion: {"{{c1::image-occlusion:rect:left=.1:top=.1:width=.2:height=.2:oi=1}}", `<img src="a.png">`},
		}[nt.Name]
		if fields == nil {
			continue
		}
		if _, err := src.AddNote(core.NoteInput{TypeID: nt.ID, DeckID: src.Decks()[0].ID, Fields: fields}); err != nil {
			t.Fatal(err)
		}
	}
	models := map[string]map[string]any{}
	row := colRow(t, readZip(t, export(t, src, core.ExportScope{All: true}, Options{}))["collection.anki21"])
	if err := json.Unmarshal([]byte(row.Str("models")), &models); err != nil {
		t.Fatal(err)
	}
	kinds := map[string]any{}
	for _, m := range models {
		kinds[m["name"].(string)] = m["originalStockKind"]
	}
	if want := map[string]any{"Basic": nil, "Image Occlusion": 6.0}; !reflect.DeepEqual(kinds, want) {
		t.Errorf("originalStockKind by note type = %v, want %v", kinds, want)
	}
}

// cardTypeCollection is a collection with a "Basic (and reversed card)"
// note added to deck Spanish, whose second card type, Production, has the
// browser question template {{Back}} and a deck override to the deck
// named override, created when missing.
func cardTypeCollection(t *testing.T, override string) (*core.Collection, core.ID) {
	t.Helper()
	c := newCollection(t)
	spanish, err := c.CreateDeck("Spanish")
	if err != nil {
		t.Fatal(err)
	}
	var typeID core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == core.StockReversed {
			typeID = nt.ID
		}
	}
	var target core.ID
	for _, d := range c.Decks() {
		if d.Name == override {
			target = d.ID
		}
	}
	if target == 0 {
		if target, err = c.CreateDeck(override); err != nil {
			t.Fatal(err)
		}
	}
	nt, _ := c.NoteType(typeID)
	tm := nt.Templates[1]
	tm.Name, tm.BrowserQFmt, tm.DeckOverride = "Production", "{{Back}}", target
	if err := c.UpdateTemplate(typeID, 1, tm); err != nil {
		t.Fatal(err)
	}
	// The note's Production card goes to the override deck.
	if _, err := c.AddNote(core.NoteInput{TypeID: typeID, DeckID: spanish, Fields: []string{"hola", "hello"}}); err != nil {
		t.Fatal(err)
	}
	return c, typeID
}

// packageTemplate returns card type name of the package's note type name.
func packageTemplate(t *testing.T, p *importer.Package, typeName, name string) importer.Template {
	t.Helper()
	for _, nt := range p.NoteTypes {
		if nt.Name != typeName {
			continue
		}
		for _, tm := range nt.Templates {
			if tm.Name == name {
				return tm
			}
		}
	}
	t.Fatalf("package has no card type %s of %s", name, typeName)
	return importer.Template{}
}

// A deck export keeps a card type's browser appearance and its deck
// override to a deck of the export, in both formats.
func TestCardTypeOptionsRoundTrip(t *testing.T) {
	src, _ := cardTypeCollection(t, "Spanish::Production")
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy=%v", legacy), func(t *testing.T) {
			data := export(t, src, core.ExportScope{DeckID: deckID(t, src, "Spanish")}, Options{Legacy: legacy})
			dst := reimport(t, data, false)
			// The import adds the note type next to the new collection's
			// stock one, under another name.
			var tm *core.CardTemplate
			for _, nt := range dst.NoteTypes() {
				for i := range nt.Templates {
					if nt.Templates[i].Name == "Production" {
						tm = &nt.Templates[i]
					}
				}
			}
			if tm == nil {
				t.Fatal("no Production card type after the import")
			}
			if want := deckID(t, dst, "Spanish::Production"); tm.DeckOverride != want {
				t.Errorf("deck override = %d, want Spanish::Production (%d)", tm.DeckOverride, want)
			}
			if tm.BrowserQFmt != "{{Back}}" || tm.BrowserAFmt != "" {
				t.Errorf("browser templates = %q, %q", tm.BrowserQFmt, tm.BrowserAFmt)
			}
		})
	}
}

// An override to a deck the export does not contain is left out; the
// browser appearance is kept.
func TestCardTypeOverrideOutsideExport(t *testing.T) {
	src, _ := cardTypeCollection(t, "Production")
	for _, legacy := range []bool{false, true} {
		p, err := importer.Read(export(t, src, core.ExportScope{DeckID: deckID(t, src, "Spanish")}, Options{Legacy: legacy}), nil)
		if err != nil {
			t.Fatal(err)
		}
		tm := packageTemplate(t, p, core.StockReversed, "Production")
		if tm.DeckID != 0 || tm.BrowserQFmt != "{{Back}}" {
			t.Errorf("legacy %v: Production = %+v, want no override and browser question {{Back}}", legacy, tm)
		}
	}
}

// An override to the Default deck names Anki's default deck.
func TestCardTypeOverrideToDefaultDeck(t *testing.T) {
	src, _ := cardTypeCollection(t, core.DefaultDeckName)
	p, err := importer.Read(export(t, src, core.ExportScope{All: true}, Options{}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if tm := packageTemplate(t, p, core.StockReversed, "Production"); tm.DeckID != ankiDefaultID {
		t.Errorf("Production deck override = %d, want %d", tm.DeckID, ankiDefaultID)
	}
	if tm := packageTemplate(t, p, core.StockReversed, "Card 1"); tm.DeckID != 0 || tm.BrowserQFmt != "" {
		t.Errorf("Card 1 = %+v, want no options", tm)
	}
}
