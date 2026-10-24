package core

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"

	"cardflip/golang/storage"
)

// cafe is "café" with the accent as a combining character (NFD), café the
// same text as one code point (NFC).
const (
	cafe = "café"
	café = "café"
)

func TestNFCForms(t *testing.T) {
	if cafe == café || norm.NFC.String(cafe) != café {
		t.Fatal("test constants are not two forms of the same text")
	}
}

// Decomposed input (spec): fields and tags of added and edited notes are
// stored in NFC, and the duplicate check and find and replace match either
// form.
func TestNFCNotes(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	id, err := c.AddNote(NoteInput{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"é", cafe}, Tags: []string{cafe}})
	if err != nil {
		t.Fatal(err)
	}
	n, _ := c.Note(id)
	if n.Fields[0] != "é" || n.Fields[1] != café || n.Tags[0] != café {
		t.Fatalf("added note = %q %q", n.Fields, n.Tags)
	}
	if err := c.UpdateNote(id, []string{"un " + cafe, cafe + " au lait"}, []string{cafe + "::noir", café}); err != nil {
		t.Fatal(err)
	}
	n, _ = c.Note(id)
	if n.Fields[0] != "un "+café || n.Fields[1] != café+" au lait" || strings.Join(n.Tags, " ") != café+"::noir "+café {
		t.Fatalf("updated note = %q %q", n.Fields, n.Tags)
	}
	if got := c.Duplicates(basic.ID, "Un "+cafe, 0); !slices.Equal(got, []ID{id}) {
		t.Fatalf("duplicates of the decomposed first field = %v", got)
	}

	count, err := c.FindReplace([]ID{id}, FindReplace{Find: cafe, Replace: "the" + "́", Field: "Back"})
	if err != nil || count != 1 {
		t.Fatalf("find and replace = %d, %v", count, err)
	}
	n, _ = c.Note(id)
	if n.Fields[0] != "un "+café || n.Fields[1] != "thé au lait" {
		t.Fatalf("after find and replace = %q", n.Fields)
	}
	reopen(t, c, store, clk)
}

func TestNFCTags(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic", "a", "b")
	if err := c.AddTags([]ID{id}, []string{cafe + "::x", café}); err != nil {
		t.Fatal(err)
	}
	if got := c.AllTags(); strings.Join(got, " ") != café+" "+café+"::x" {
		t.Fatalf("tags = %q", got)
	}
	// Renaming, moving and deleting find the NFC tags by either form.
	if n, err := c.RenameTag(cafe, "thé"); err != nil || n != 1 {
		t.Fatalf("rename = %d, %v", n, err)
	}
	if n, err := c.ReparentTags([]string{"thé::x"}, cafe); err != nil || n != 1 {
		t.Fatalf("reparent = %d, %v", n, err)
	}
	if got := c.AllTags(); strings.Join(got, " ") != café+"::x thé" {
		t.Fatalf("tags after rename and move = %q", got)
	}
	if n, err := c.DeleteTags([]string{cafe}); err != nil || n != 1 {
		t.Fatalf("delete = %d, %v", n, err)
	}
	if err := c.RemoveTags([]ID{id}, []string{"thé"}); err != nil {
		t.Fatal(err)
	}
	if got := c.AllTags(); len(got) != 0 {
		t.Fatalf("tags after delete and remove = %q", got)
	}
}

func TestNFCDecks(t *testing.T) {
	c, store, clk := testCollection(t)
	id, err := c.CreateDeck("Langues::" + cafe)
	if err != nil {
		t.Fatal(err)
	}
	if got := deckNames(c); got != "Default, Langues, Langues::"+café {
		t.Fatalf("decks = %q", got)
	}
	_, err = c.CreateDeck("Langues::" + café)
	mustUserErr(t, err, "already exists")
	if err := c.RenameDeck(id, "Langues::"+cafe+" noir"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetDeckDescription(id, "<b>"+cafe+"</b>"); err != nil {
		t.Fatal(err)
	}
	d, _ := c.Deck(id)
	if d.Name != "Langues::"+café+" noir" || d.Description != "<b>"+café+"</b>" {
		t.Fatalf("deck = %q %q", d.Name, d.Description)
	}
	addNote(t, c, "Basic", "a", "b")
	fid, _, err := c.CreateFilteredDeck(cafe+" filtered", DefaultFilterConfig("deck:Default"))
	if err != nil {
		t.Fatal(err)
	}
	if f, _ := c.Deck(fid); f.Name != café+" filtered" {
		t.Fatalf("filtered deck = %q", f.Name)
	}
	reopen(t, c, store, clk)
}

func TestNFCNoteTypes(t *testing.T) {
	c, store, clk := testCollection(t)
	id, err := c.AddNoteType(StockBasic, cafe)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.AddNoteType(StockBasic, café)
	mustUserErr(t, err, "already exists")
	if err := c.RenameNoteType(id, cafe+" 2"); err != nil {
		t.Fatal(err)
	}
	if err := c.AddField(id, "Extra"); err != nil {
		t.Fatal(err)
	}
	if err := c.RenameField(id, 2, cafe); err != nil {
		t.Fatal(err)
	}
	err = c.AddField(id, café)
	mustUserErr(t, err, "already has a field")
	// A template that refers to the field in decomposed form still finds it.
	nt, _ := c.NoteType(id)
	tmpl := nt.Templates[0]
	tmpl.Name = "Carte " + cafe
	tmpl.AFmt = "{{" + cafe + "}}"
	if problems, err := c.CheckTemplate(id, 0, tmpl); err != nil || len(problems) != 0 {
		t.Fatalf("problems = %v, %v", problems, err)
	}
	if err := c.UpdateTemplate(id, 0, tmpl); err != nil {
		t.Fatal(err)
	}
	if err := c.SetNoteTypeCSS(id, `.card::after { content: "`+cafe+`" }`); err != nil {
		t.Fatal(err)
	}
	nt, _ = c.NoteType(id)
	if nt.Name != café+" 2" || nt.Fields[2] != café || nt.Templates[0].Name != "Carte "+café ||
		nt.Templates[0].AFmt != "{{"+café+"}}" || nt.CSS != `.card::after { content: "`+café+`" }` {
		t.Fatalf("note type = %q %q %+v %q", nt.Name, nt.Fields, nt.Templates[0], nt.CSS)
	}
	reopen(t, c, store, clk)
}

func TestNFCMediaNames(t *testing.T) {
	c, _, _ := testCollection(t)
	name, err := c.AddMedia(cafe+".jpg", []byte("x"))
	if err != nil || name != café+".jpg" || !c.HasMedia(café+".jpg") {
		t.Fatalf("added media = %q, %v", name, err)
	}
}

// An import stores its text and media names in NFC. A decomposed media name
// whose NFC form names another file of the batch gets a hashed name.
func TestNFCImport(t *testing.T) {
	c, store, clk := testCollection(t)
	nt := NoteType{Name: cafe, Kind: Standard, Fields: []string{cafe, "Back"}, CSS: `@font-face { src: url("` + cafe + `.ttf") }`,
		Templates: []CardTemplate{{Name: "Card 1", QFmt: "{{" + cafe + "}}", AFmt: "{{Back}}"}}}
	b := &ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "k", Type: nt}},
		Notes: []ImportNote{
			{TypeKey: "k", GUID: "g1", DeckName: "Langues::" + cafe, Fields: []string{cafe, `<img src="` + cafe + `.jpg">`}, Tags: []string{cafe}},
			{TypeKey: "k", GUID: "g2", DeckName: "Langues::" + café, Fields: []string{"x", `<img src="` + café + `.jpg">`}},
		},
		Media: map[string][]byte{cafe + ".jpg": []byte("decomposed"), café + ".jpg": []byte("composed"), cafe + ".ttf": []byte("font")},
		Decks: []string{cafe + " vide"}, DeckDescriptions: map[string]string{"Langues::" + cafe: cafe},
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	if b.Notes[0].Fields[0] != cafe || b.Media[cafe+".jpg"] == nil || b.NoteTypes[0].Type.Name != cafe {
		t.Fatal("the batch was changed")
	}
	if got := deckNames(c); got != café+" vide, Default, Langues, Langues::"+café {
		t.Fatalf("decks = %q", got)
	}
	typ := noteType(t, c, café)
	if typ.Fields[0] != café || typ.Templates[0].QFmt != "{{"+café+"}}" || !strings.Contains(typ.CSS, café+".ttf") {
		t.Fatalf("note type = %q %+v %q", typ.Fields, typ.Templates, typ.CSS)
	}
	n1, n2 := noteByGUID(t, c, "g1"), noteByGUID(t, c, "g2")
	hashed := hashedMediaName(café+".jpg", []byte("decomposed"))
	if n1.Fields[0] != café || n1.Tags[0] != café || n1.Fields[1] != `<img src="`+hashed+`">` || n2.Fields[1] != `<img src="`+café+`.jpg">` {
		t.Fatalf("notes = %q %q, %q", n1.Fields, n1.Tags, n2.Fields)
	}
	for name, want := range map[string]string{hashed: "decomposed", café + ".jpg": "composed", café + ".ttf": "font"} {
		if data, err := c.Media(name); err != nil || string(data) != want {
			t.Fatalf("media %q = %q, %v", name, data, err)
		}
	}
	if c.HasMedia(cafe+".jpg") || c.HasMedia(cafe+".ttf") {
		t.Fatal("media stored under decomposed names")
	}
	for _, d := range c.Decks() {
		if d.Name == "Langues::"+café && d.Description != café {
			t.Fatalf("description = %q", d.Description)
		}
	}
	reopen(t, c, store, clk)
}

// Replacing the collection with a package stores its media in NFC too.
func TestNFCReplaceWithImport(t *testing.T) {
	c, _, _ := testCollection(t)
	b := &ImportBatch{Media: map[string][]byte{cafe + ".jpg": []byte("x")}}
	if _, err := c.ReplaceWithImport(b, nil); err != nil {
		t.Fatal(err)
	}
	if !c.HasMedia(café+".jpg") || c.HasMedia(cafe+".jpg") {
		t.Fatal("media not stored under its NFC name")
	}
}

// oldNFCState returns a snapshot as an earlier version saved it: decomposed
// text, no normalised flag. The first note shows cafe.jpg and cafe.png,
// the second café.png, a different file than cafe.png; the note type's CSS
// uses the font _cafe.ttf. The media map holds the files.
func oldNFCState(t *testing.T) (*state, map[string][]byte, *clock) {
	t.Helper()
	src, _, clk := testCollection(t)
	n1 := addNote(t, src, "Basic", "a", "b")
	n2 := addNote(t, src, "Basic", "c", "d")
	src.mu.RLock()
	s := src.state()
	src.mu.RUnlock()
	s.Normalized = false
	for _, n := range s.Notes {
		switch n.ID {
		case n1:
			n.Fields = []string{cafe, `<img src="` + cafe + `.jpg"><img src="` + cafe + `.png">`}
			n.Tags = []string{cafe}
		case n2:
			n.Fields = []string{"c", `<img src="` + café + `.png">`}
		}
	}
	for _, d := range s.Decks {
		d.Name, d.Description = cafe, cafe
	}
	for _, nt := range s.NoteTypes {
		if nt.Name == "Basic" {
			nt.Name, nt.Fields[0] = cafe, cafe
			nt.Templates[0].QFmt = "{{" + cafe + "}}"
			nt.CSS = `@font-face { font-family: f; src: url("_` + cafe + `.ttf") }`
		}
	}
	media := map[string][]byte{cafe + ".jpg": []byte("jpg"), cafe + ".png": []byte("decomposed png"),
		café + ".png": []byte("composed png"), "_" + cafe + ".ttf": []byte("font")}
	return s, media, clk
}

// checkNFCState checks the result of normalising oldNFCState.
func checkNFCState(t *testing.T, c *Collection, old *state) {
	t.Helper()
	c.mu.RLock()
	defer c.mu.RUnlock()
	hashed := hashedMediaName(café+".png", []byte("decomposed png"))
	for _, n := range old.Notes {
		got := c.notes[n.ID]
		var want []string
		for _, f := range n.Fields {
			want = append(want, norm.NFC.String(strings.ReplaceAll(f, `"`+cafe+`.png"`, `"`+hashed+`"`)))
		}
		if !slices.Equal(got.Fields, want) || !slices.Equal(got.Tags, nfcAll(n.Tags)) || !got.Modified.Equal(n.Modified) {
			t.Errorf("note %d = %q %q %v, want %q %q %v", n.ID, got.Fields, got.Tags, got.Modified, want, nfcAll(n.Tags), n.Modified)
		}
	}
	for _, d := range c.decks {
		if d.Name != café || d.Description != café {
			t.Errorf("deck = %q %q", d.Name, d.Description)
		}
	}
	var nt *NoteType
	for _, x := range c.noteTypes {
		if x.Name == café {
			nt = x
		}
	}
	if nt == nil || nt.Fields[0] != café || nt.Templates[0].QFmt != "{{"+café+"}}" || !strings.Contains(nt.CSS, "_"+café+".ttf") {
		t.Fatalf("note type = %+v", nt)
	}
	for name, want := range map[string]string{café + ".jpg": "jpg", hashed: "decomposed png", café + ".png": "composed png", "_" + café + ".ttf": "font"} {
		if data, err := c.store.GetMedia(name); err != nil || string(data) != want {
			t.Errorf("media %q = %q, %v", name, data, err)
		}
	}
	for _, name := range []string{cafe + ".jpg", cafe + ".png", "_" + cafe + ".ttf"} {
		if c.store.HasMedia(name) {
			t.Errorf("media %q kept under its decomposed name", name)
		}
	}
}

// A collection saved by an earlier version is normalised once, when it is
// first opened: its text, and the names of its media files, which the
// references keep finding. A second open writes nothing.
func TestMigrateNFC(t *testing.T) {
	s, media, clk := oldNFCState(t)
	snap, err := json.Marshal(s)
	if err != nil {
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
	for name, data := range media {
		if err := st.PutMedia(name, data); err != nil {
			t.Fatal(err)
		}
	}
	// The journal of the earlier version is normalised too.
	tag := *s.Notes[1]
	tag.Tags = []string{"thé"}
	raw, _ := json.Marshal(&op{Kind: "addTags", Notes: []*Note{&tag}})
	if err := st.Commit(raw, nil, nil); err != nil {
		t.Fatal(err)
	}
	s.Notes[1] = &tag

	c, fresh, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil || fresh {
		t.Fatalf("open: %v fresh=%v", err, fresh)
	}
	checkNFCState(t, c, s)
	if trashed, err := c.store.TrashedMedia(); err != nil || len(trashed) != 3 {
		t.Errorf("trash = %+v, %v; want the 3 decomposed names", trashed, err)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "" {
		t.Errorf("the migration is an undo step: %q", undo)
	}
	_, loaded, err := storage.Open(store.Open())
	if err != nil {
		t.Fatal(err)
	}
	var saved state
	if err := json.Unmarshal(loaded.Snapshot, &saved); err != nil || !saved.Normalized || len(loaded.Ops) != 0 {
		t.Fatalf("saved snapshot normalised %v, %d ops (%v)", saved.Normalized, len(loaded.Ops), err)
	}

	var writes []string
	store.FailPut = func(key string) bool {
		writes = append(writes, key)
		return false
	}
	reopen(t, c, store, clk)
	if len(writes) != 0 {
		t.Fatalf("second open wrote %q", writes)
	}
}

// Restoring a backup of an earlier version normalises it in the same way.
func TestRestoreNFC(t *testing.T) {
	s, media, _ := oldNFCState(t)
	snap, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	c, store, clk := testCollection(t)
	if err := c.Restore(snap, nil, nil, media); err != nil {
		t.Fatal(err)
	}
	checkNFCState(t, c, s)
	reopen(t, c, store, clk)
}
