package text

import (
	"strings"
	"testing"

	"cardflip/golang/core"
	"cardflip/golang/importer/csv"
	"cardflip/golang/render"
	"cardflip/golang/storage"
)

func newCollection(t *testing.T) (*core.Collection, map[string]core.ID, core.ID) {
	t.Helper()
	c, _, err := core.Open(storage.NewMemory(), core.Options{})
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]core.ID{}
	for _, nt := range c.NoteTypes() {
		types[nt.Name] = nt.ID
	}
	if _, ok := types[core.StockOptionalReversed]; !ok {
		id, err := c.AddNoteType(core.StockOptionalReversed, core.StockOptionalReversed)
		if err != nil {
			t.Fatal(err)
		}
		types[core.StockOptionalReversed] = id
	}
	deck, err := c.CreateDeck("Spanish")
	if err != nil {
		t.Fatal(err)
	}
	return c, types, deck
}

func addNote(t *testing.T, c *core.Collection, typeID, deck core.ID, tags []string, fields ...string) core.ID {
	t.Helper()
	id, err := c.AddNote(core.NoteInput{TypeID: typeID, DeckID: deck, Fields: fields, Tags: tags})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func exportNotes(t *testing.T, c *core.Collection, scope core.ExportScope, o NoteOptions) string {
	t.Helper()
	set, err := c.Export(scope)
	if err != nil {
		t.Fatal(err)
	}
	return string(WriteNotes(set, o, nil))
}

// reimport imports a notes export into the Basic note type with the
// "update" duplicate option and the file's default mapping.
func reimport(t *testing.T, c *core.Collection, basic, deck core.ID, text string) core.ImportSummary {
	t.Helper()
	f, err := csv.Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	b, err := csv.Build(f, csv.Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(),
		Mapping: f.Mapping(2), Duplicates: core.DupUpdate, AllowHTML: !f.NoHTML, Tags: f.Tags})
	if err != nil {
		t.Fatal(err)
	}
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sum
}

func TestNotesWithTags(t *testing.T) {
	// Export with tags (spec): fields followed by the tags, and a header
	// naming the tags column.
	c, types, deck := newCollection(t)
	addNote(t, c, types[core.StockBasic], deck, []string{"food", "verb"}, "<b>comer</b>", "to eat")
	addNote(t, c, types[core.StockBasic], deck, nil, "beber", "to drink")
	got := exportNotes(t, c, core.ExportScope{DeckID: deck}, NoteOptions{Tags: true})
	want := "#separator:tab\n#html:true\n#tags column:3\n" +
		"<b>comer</b>\tto eat\tfood verb\n" +
		"beber\tto drink\t\n"
	if got != want {
		t.Fatalf("export =\n%q\nwant\n%q", got, want)
	}
}

func TestNotesAllColumns(t *testing.T) {
	// GUID, note type and deck come before the fields; fields are padded to
	// the largest note type so that the tags stay in one column.
	c, types, deck := newCollection(t)
	sub, err := c.CreateDeck("Spanish::Verbs")
	if err != nil {
		t.Fatal(err)
	}
	b := addNote(t, c, types[core.StockBasic], sub, []string{"a"}, "comer", "to eat")
	r := addNote(t, c, types[core.StockOptionalReversed], deck, []string{"b"}, "casa", "house", "y")
	got := exportNotes(t, c, core.ExportScope{DeckID: deck}, NoteOptions{Tags: true, Deck: true, NoteType: true, GUID: true})
	nb, _ := c.Note(b)
	nr, _ := c.Note(r)
	want := "#separator:tab\n#html:true\n#guid column:1\n#notetype column:2\n#deck column:3\n#tags column:7\n" +
		nb.GUID + "\tBasic\tSpanish::Verbs\tcomer\tto eat\t\ta\n" +
		nr.GUID + "\t" + core.StockOptionalReversed + "\tSpanish\tcasa\thouse\ty\tb\n"
	if got != want {
		t.Fatalf("export =\n%q\nwant\n%q", got, want)
	}
}

func TestNotesQuoting(t *testing.T) {
	// Fields with tabs, quotes or line breaks are quoted, as is a first
	// field starting with "#"; the importer reads them back unchanged.
	c, types, deck := newCollection(t)
	basic := types[core.StockBasic]
	fields := [][2]string{
		{"line 1\nline 2", "a\tb"},
		{`say "hi"`, "plain"},
		{"#1 rule", "crlf\r\nend"},
		{"x<br>y", "  leading space"},
	}
	for _, f := range fields {
		addNote(t, c, basic, deck, nil, f[0], f[1])
	}
	text := exportNotes(t, c, core.ExportScope{DeckID: deck}, NoteOptions{})
	if !strings.Contains(text, "\"line 1\nline 2\"\t\"a\tb\"\n") || !strings.Contains(text, "\"say \"\"hi\"\"\"\tplain\n") ||
		!strings.Contains(text, "\"#1 rule\"\t") {
		t.Fatalf("export = %q", text)
	}
	f, err := csv.Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if !f.HTML || f.Delimiter != '\t' || len(f.Rows) != len(fields) {
		t.Fatalf("parsed = %+v", f)
	}
	for i, row := range f.Rows {
		want := fields[i]
		if i == 2 {
			want[1] = "crlf\nend" // the importer reads CRLF as LF
		}
		if len(row) != 2 || row[0] != want[0] || row[1] != want[1] {
			t.Fatalf("row %d = %q, want %q", i, row, want)
		}
	}
}

func TestNotesRoundTrip(t *testing.T) {
	// Edit outside and re-import (spec): the edited note is updated, line
	// breaks are kept and nothing is duplicated.
	c, types, deck := newCollection(t)
	basic := types[core.StockBasic]
	id := addNote(t, c, basic, deck, nil, "perro", "dog")
	addNote(t, c, basic, deck, nil, "gato", "cat\nfeline")
	text := exportNotes(t, c, core.ExportScope{DeckID: deck}, NoteOptions{Tags: true})
	edited := strings.Replace(text, "perro\tdog", "perro\thound", 1)
	sum := reimport(t, c, basic, deck, edited)
	if sum.NotesAdded != 0 || sum.NotesUpdated != 1 || sum.NotesSkipped != 1 {
		t.Fatalf("summary = %+v", sum)
	}
	n, _ := c.Note(id)
	if n.Fields[1] != "hound" {
		t.Fatalf("fields = %q", n.Fields)
	}
	if got := exportNotes(t, c, core.ExportScope{DeckID: deck}, NoteOptions{Tags: true}); got != edited {
		t.Fatalf("after re-import =\n%q\nwant\n%q", got, edited)
	}
}

func TestNotesSelection(t *testing.T) {
	// Export selected notes: one line per note, also for a note selected by
	// two of its cards.
	c, types, deck := newCollection(t)
	var cards []core.ID
	for _, w := range []string{"uno", "dos", "tres"} {
		id := addNote(t, c, types[core.StockOptionalReversed], deck, nil, w, w+"!", "y")
		for _, card := range c.CardsOfNote(id) {
			cards = append(cards, card.ID)
		}
	}
	set, err := c.Export(core.ExportScope{CardIDs: cards[:3]})
	if err != nil {
		t.Fatal(err)
	}
	got := string(WriteNotes(set, NoteOptions{}, nil))
	if want := "#separator:tab\n#html:true\nuno\tuno!\ty\ndos\tdos!\ty\n"; got != want {
		t.Fatalf("export = %q, want %q", got, want)
	}
}

func TestCards(t *testing.T) {
	// Printable list (spec): question and answer without HTML; the answer
	// leaves out the repeated question.
	cards := []render.Card{
		{Question: `<div class="card"><b>hola</b>&nbsp;&amp; adiós</div>`,
			Answer: `<div class="card"><b>hola</b>&nbsp;&amp; adiós<hr id="answer">hello<br>hi</div>`},
		{Question: `<div class="card">x &lt; y</div>`, Answer: `<div class="card">only "answer"</div>`},
	}
	got := string(WriteCards(cards, nil))
	want := "#separator:tab\n#html:false\nhola & adiós\thello hi\nx < y\t\"only \"\"answer\"\"\"\n"
	if got != want {
		t.Fatalf("export = %q, want %q", got, want)
	}
}

func TestCardsRendered(t *testing.T) {
	// Cards of a deck, rendered by the collection: a cloze shows its blank
	// on the question and the text on the answer.
	c, types, deck := newCollection(t)
	addNote(t, c, types[core.StockCloze], deck, nil, "{{c1::Madrid}} is in Spain", "")
	set, err := c.Export(core.ExportScope{DeckID: deck})
	if err != nil {
		t.Fatal(err)
	}
	var rendered []render.Card
	for _, card := range set.Cards {
		r, err := c.RenderCard(card.ID, render.Env{})
		if err != nil {
			t.Fatal(err)
		}
		rendered = append(rendered, r)
	}
	got := string(WriteCards(rendered, nil))
	if want := "#separator:tab\n#html:false\n[...] is in Spain\tMadrid is in Spain\n"; got != want {
		t.Fatalf("export = %q, want %q", got, want)
	}
}
