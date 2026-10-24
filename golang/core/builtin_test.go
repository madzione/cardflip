package core

import (
	"slices"
	"strings"
	"testing"

	"cardflip/golang/storage"
)

// germanNames are built-in names as the app sends them in German.
func germanNames() BuiltinNames {
	return BuiltinNames{
		Deck: "Standard", Preset: "Standard",
		NoteTypes: []StockNames{
			{Stock: StockBasic, Name: "Einfach", Fields: []string{"Vorderseite", "Rückseite"}, CardTypes: []string{"Karte 1"}},
			{Stock: StockReversed, Name: "Einfach (mit umgekehrter Karte)", Fields: []string{"Vorderseite", "Rückseite"},
				CardTypes: []string{"Karte 1", "Karte 2"}},
			{Stock: StockCloze, Name: "Lückentext", Fields: []string{"Text", "Rückseite Extra"}, CardTypes: []string{"Lückentext"}},
			{Stock: StockImageOcclusion, Name: "Bildverdeckung",
				Fields:    []string{"Verdeckung", "Bild", "Kopfzeile", "Rückseite Extra", "Kommentare"},
				CardTypes: []string{"Bildverdeckung"}},
		},
	}
}

func openNamed(t *testing.T, names BuiltinNames) *Collection {
	t.Helper()
	c, fresh, err := Open(storage.NewMemory(), Options{Names: names})
	if err != nil || !fresh {
		t.Fatalf("open: %v fresh=%v", err, fresh)
	}
	return c
}

func TestNewCollectionInGerman(t *testing.T) {
	c := openNamed(t, germanNames())
	if decks := c.Decks(); len(decks) != 1 || decks[0].Name != "Standard" {
		t.Fatalf("decks = %+v", decks)
	}
	if p := c.presets[DefaultPresetID]; p.Name != "Standard" {
		t.Fatalf("preset = %q", p.Name)
	}
	basic := noteType(t, c, "Einfach")
	if !slices.Equal(basic.Fields, []string{"Vorderseite", "Rückseite"}) || basic.Templates[0].Name != "Karte 1" ||
		basic.Templates[0].QFmt != "{{Vorderseite}}" || !strings.HasSuffix(basic.Templates[0].AFmt, "{{Rückseite}}") {
		t.Fatalf("Einfach = %+v", basic)
	}
	noteType(t, c, "Einfach (mit umgekehrter Karte)")
	noteType(t, c, "Lückentext")
	occlusion := noteType(t, c, "Bildverdeckung")
	if shapes, image, ok := occlusion.OcclusionFields(); !ok || shapes != 0 || image != 1 {
		t.Fatalf("occlusion fields = %d, %d, %v", shapes, image, ok)
	}
	if !strings.Contains(occlusion.Templates[0].QFmt, "{{cloze:Verdeckung}}") ||
		!strings.Contains(occlusion.Templates[0].AFmt, "{{#Rückseite Extra}}") {
		t.Fatalf("occlusion templates = %+v", occlusion.Templates[0])
	}
	// A note of the translated note type gets its card.
	id := addNote(t, c, "Einfach", "hola", "hallo")
	if len(c.CardsOfNote(id)) != 1 {
		t.Fatal("no card")
	}
	if id, ok := c.BasicNoteType(); !ok || id != basic.ID {
		t.Fatalf("BasicNoteType = %v, %v", id, ok)
	}
}

func TestBuiltinNamesFallBackToEnglish(t *testing.T) {
	names := BuiltinNames{Deck: "A::B", Preset: " ", NoteTypes: []StockNames{
		// Two fields with the same name.
		{Stock: StockBasic, Name: "Einfach", Fields: []string{"Seite", "seite"}, CardTypes: []string{"Karte 1"}},
		// A name that templates cannot use, and one missing.
		{Stock: StockCloze, Name: "Lückentext", Fields: []string{"Te:xt", "Extra"}, CardTypes: []string{"Lückentext"}},
		{Stock: StockReversed, Name: "Umgekehrt", Fields: []string{"Vorne"}, CardTypes: []string{"1", "2"}},
	}}
	c := openNamed(t, names)
	if c.Decks()[0].Name != DefaultDeckName || c.presets[DefaultPresetID].Name != DefaultPresetName {
		t.Fatalf("deck %q, preset %q", c.Decks()[0].Name, c.presets[DefaultPresetID].Name)
	}
	for _, name := range []string{StockBasic, StockReversed, StockCloze, StockImageOcclusion} {
		noteType(t, c, name)
	}
}

func TestAddStockNoteTypeInGerman(t *testing.T) {
	c, _, _ := testCollection(t)
	c.SetBuiltinNames(germanNames())
	// Existing note types keep their names.
	noteType(t, c, StockBasic)
	id, err := c.AddNoteType(StockCloze, "Lückentext")
	if err != nil {
		t.Fatal(err)
	}
	nt, _ := c.NoteType(id)
	if nt.Name != "Lückentext" || !slices.Equal(nt.Fields, []string{"Text", "Rückseite Extra"}) ||
		nt.Templates[0].AFmt != "{{cloze:Text}}<br>\n{{Rückseite Extra}}" {
		t.Fatalf("added %+v", nt)
	}
	// The English Basic is still found while there is no Einfach.
	if got, ok := c.BasicNoteType(); !ok || got != noteType(t, c, StockBasic).ID {
		t.Fatalf("BasicNoteType = %v, %v", got, ok)
	}
}

func TestImageOcclusionNoteTypeInGerman(t *testing.T) {
	c, _, _ := testCollection(t)
	english := noteType(t, c, StockImageOcclusion)
	c.SetBuiltinNames(germanNames())
	// An existing English one is used.
	if id, err := c.ImageOcclusionNoteType(); err != nil || id != english.ID {
		t.Fatalf("existing: %v, %v", id, err)
	}
	if err := c.DeleteNoteType(english.ID); err != nil {
		t.Fatal(err)
	}
	id, err := c.ImageOcclusionNoteType()
	if err != nil {
		t.Fatal(err)
	}
	nt, _ := c.NoteType(id)
	if nt.Name != "Bildverdeckung" || !nt.IsImageOcclusion() || nt.Fields[0] != "Verdeckung" {
		t.Fatalf("created %+v", nt)
	}
}

func TestOcclusionFieldsFollowTheTemplate(t *testing.T) {
	c := openNamed(t, germanNames())
	nt := noteType(t, c, "Bildverdeckung")
	// Reordered and renamed fields are still found.
	nt.Fields = []string{"Bild", "Kopfzeile", "Masken", "Rückseite Extra", "Kommentare"}
	nt.Templates[0].QFmt = strings.ReplaceAll(nt.Templates[0].QFmt, "Verdeckung", "Masken")
	if shapes, image, ok := nt.OcclusionFields(); !ok || shapes != 2 || image != 0 {
		t.Fatalf("occlusion fields = %d, %d, %v", shapes, image, ok)
	}
	// Without the image container it is a plain cloze note type.
	nt.Templates[0].QFmt = "{{cloze:Masken}}{{Bild}}"
	if nt.IsImageOcclusion() {
		t.Fatal("a plain cloze note type counts as Image Occlusion")
	}
	if cloze := noteType(t, c, "Lückentext"); cloze.IsImageOcclusion() {
		t.Fatal("Lückentext counts as Image Occlusion")
	}
}

func TestCheckDatabaseDefaultDeckInGerman(t *testing.T) {
	c := openNamed(t, germanNames())
	id := addNote(t, c, "Einfach", "hola", "hallo")
	card := rawCard(c, firstCard(t, c, id).ID)
	card.DeckID = 424242
	corrupt(t, c, &op{Cards: []*Card{card}})
	r := checkDB(t, c)
	if got := message(t, r, ProblemMissingDeck); got != "1 card in a missing deck was moved to Standard." {
		t.Fatalf("message = %q", got)
	}
	if got := c.mustCard(t, card.ID); got.DeckID != c.Decks()[0].ID || len(c.Decks()) != 1 {
		t.Fatalf("card in deck %v, decks %+v", got.DeckID, c.Decks())
	}
}
