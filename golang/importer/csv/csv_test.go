package csv

import (
	"cardflip/golang/internal/usermsg"
	"errors"
	"fmt"
	"strings"
	"testing"

	"cardflip/golang/core"
	"cardflip/golang/storage"
)

func openCollection(t *testing.T, b storage.Backend) (*core.Collection, core.ID, core.ID) {
	t.Helper()
	c, _, err := core.Open(b, core.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var basic core.ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	return c, basic, c.Decks()[0].ID
}

func importText(t *testing.T, c *core.Collection, typeID, deck core.ID, text string, dup core.DupMode) core.ImportSummary {
	t.Helper()
	f, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build(f, Options{TypeID: typeID, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: DefaultMapping(f.Columns, 2), Duplicates: dup,
		AllowHTML: !f.NoHTML, Tags: f.Tags})
	if err != nil {
		t.Fatal(err)
	}
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	return sum
}

func notes(t *testing.T, c *core.Collection) []string {
	t.Helper()
	ids, _ := c.Search("", core.ColCreated, false)
	var out []string
	for _, row := range c.Rows(ids) {
		n, _ := c.Note(row.NoteID)
		out = append(out, strings.Join(n.Fields, "|"))
	}
	return out
}

func TestTabSeparated(t *testing.T) {
	// Tab-separated file (spec): two columns, mapped to Front and Back.
	f, err := Parse([]byte("hola\thello\nperro\tdog\n"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Delimiter != '\t' || f.Columns != 2 || len(f.Rows) != 2 {
		t.Fatalf("file = %+v", f)
	}
	if m := DefaultMapping(f.Columns, 2); m[0] != 0 || m[1] != 1 {
		t.Fatalf("mapping = %v", m)
	}
	c, basic, deck := openCollection(t, storage.NewMemory())
	sum := importText(t, c, basic, deck, "hola\thello\nperro\tdog\n", core.DupSkip)
	if sum.NotesAdded != 2 || sum.CardsAdded != 2 {
		t.Fatalf("summary = %+v", sum)
	}
	if got := strings.Join(notes(t, c), ","); got != "hola|hello,perro|dog" {
		t.Fatalf("notes = %q", got)
	}
}

func TestQuotedFields(t *testing.T) {
	// Quoted fields (spec): a comma and a newline inside quotes.
	text := "\"a, b\",\"line1\nline2\"\nc,d\n"
	f, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if f.Delimiter != ',' || len(f.Rows) != 2 || f.Rows[0][0] != "a, b" || f.Rows[0][1] != "line1\nline2" {
		t.Fatalf("rows = %q (delimiter %q)", f.Rows, f.Delimiter)
	}
	c, basic, deck := openCollection(t, storage.NewMemory())
	importText(t, c, basic, deck, text, core.DupSkip)
	if got := notes(t, c)[0]; got != "a, b|line1<br>line2" {
		t.Fatalf("note = %q", got)
	}
}

func TestHTMLHeaderKeepsLineBreaks(t *testing.T) {
	// A file declaring HTML fields (Cardflip and Anki text exports) keeps
	// line breaks as they are, so the fields come back as exported.
	text := "#separator:tab\n#html:true\n\"<i>a</i>\nb\"\tc\n"
	f, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if !f.HTML || len(f.Rows) != 1 {
		t.Fatalf("file = %+v", f)
	}
	c, basic, deck := openCollection(t, storage.NewMemory())
	importText(t, c, basic, deck, text, core.DupSkip)
	if got := notes(t, c)[0]; got != "<i>a</i>\nb|c" {
		t.Fatalf("note = %q", got)
	}
}

func TestDelimiterDetection(t *testing.T) {
	cases := map[string]rune{
		"a;b;c\nd;e;f\n":                     ';',
		"a,b\nc,d\n":                         ',',
		"a\tb, c\nd\te, f\n":                 '\t',
		"\xef\xbb\xbfa;b\nc;d\n":             ';',
		"word, with comma;other\nx, y;z\n":   ';',
		"#separator:Pipe\n#html:true\na|b\n": '|',
	}
	for text, want := range cases {
		f, err := Parse([]byte(text))
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if f.Delimiter != want {
			t.Errorf("%q: delimiter %q, want %q", text, f.Delimiter, want)
		}
		if strings.HasPrefix(f.Rows[0][0], "\xef\xbb\xbf") {
			t.Errorf("%q: BOM not stripped", text)
		}
	}
	f, _ := Parse([]byte("#separator:tab\n#columns:Front\tBack\nhola\thello\n"))
	if !f.Header || len(f.Rows) != 1 || strings.Join(f.ColumnNames, ",") != "Front,Back" {
		t.Fatalf("anki header: %+v", f)
	}
	if _, err := Parse([]byte("a,b\n\xff\xfe")); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if _, err := Parse([]byte("\n\n")); err == nil {
		t.Fatal("empty file accepted")
	}
}

func TestSeparatorHeader(t *testing.T) {
	// Names in any case and the characters themselves.
	cases := map[string]rune{
		"Comma": ',', "semicolon": ';', "TAB": '\t', "Space": ' ', "pipe": '|', "Colon": ':',
		",": ',', ";": ';', "\t": '\t', " ": ' ', "|": '|', ":": ':',
	}
	for value, want := range cases {
		row := strings.Join([]string{"a", "b"}, string(want))
		f, err := Parse([]byte("#separator:" + value + "\n" + row + "\n"))
		if err != nil {
			t.Fatalf("%q: %v", value, err)
		}
		if f.Delimiter != want || f.Columns != 2 {
			t.Errorf("#separator:%q: delimiter %q, %d columns", value, f.Delimiter, f.Columns)
		}
		if name := SeparatorName(want); name == "" || len(name) == 1 {
			t.Errorf("%q has no name: %q", want, name)
		}
		if r, ok := ParseSeparator(SeparatorName(want)); !ok || r != want {
			t.Errorf("name of %q does not parse back", want)
		}
	}
	if _, ok := ParseSeparator("slash"); ok {
		t.Error("unknown separator accepted")
	}
	// Column names are split at the declared separator.
	f, _ := Parse([]byte("#separator:Pipe\n#columns:Front|Back\na|b\n"))
	if strings.Join(f.ColumnNames, ",") != "Front,Back" {
		t.Fatalf("column names = %q", f.ColumnNames)
	}
}

func TestSeparatorOverride(t *testing.T) {
	// Override detection (spec): detected as comma-separated, split at "|".
	text := "a, x|b\nc, y|d\n"
	f, _ := Parse([]byte(text))
	if f.Delimiter != ',' {
		t.Fatalf("detected %q", f.Delimiter)
	}
	f, err := ParseWith([]byte(text), '|')
	if err != nil {
		t.Fatal(err)
	}
	if f.Delimiter != '|' || f.Columns != 2 || f.Rows[0][0] != "a, x" || f.Rows[1][1] != "d" {
		t.Fatalf("file = %+v", f)
	}
	// The override also wins over a declared separator.
	f, _ = ParseWith([]byte("#separator:tab\na;b\n"), ';')
	if f.Delimiter != ';' || f.Columns != 2 {
		t.Fatalf("file = %+v", f)
	}
}

func TestCommentLines(t *testing.T) {
	// Comment in the middle (spec): two notes, one comment line.
	c, basic, deck := openCollection(t, storage.NewMemory())
	f, err := Parse([]byte("a,b\n# chapter 2\nc,d\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Rows) != 2 || f.Comments != 1 {
		t.Fatalf("file = %+v", f)
	}
	sum := importText(t, c, basic, deck, "a,b\n# chapter 2\nc,d\n", core.DupSkip)
	if sum.NotesAdded != 2 {
		t.Fatalf("summary = %+v", sum)
	}

	// Comments after the header lines, between blank lines and at the end
	// count; a "#" inside a quoted field or after a space does not.
	text := "#separator:tab\n# words\n\na\t\"1\n# not a comment\"\n\n# two\n#three\n #x\tb\n# end"
	f, err = Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if f.Comments != 4 || len(f.Rows) != 2 || f.Rows[0][1] != "1\n# not a comment" || f.Rows[1][0] != " #x" {
		t.Fatalf("rows = %q, %d comments", f.Rows, f.Comments)
	}
	// Comment lines do not confuse the separator detection.
	f, _ = Parse([]byte("# a, b, c, d\n# e, f, g\nx;y\nz;w\n"))
	if f.Delimiter != ';' || f.Comments != 2 {
		t.Fatalf("file = %+v", f)
	}
}

func TestHTMLAndTagsHeaders(t *testing.T) {
	// Literal text (spec): with HTML off, "<", ">" and "&" show as typed and
	// line breaks are kept.
	text := "#separator:tab\n#html:false\n#tags:chapter1 vocab\na < b & c\t\"x>y\nz\"\n"
	f, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if f.HTML || !f.NoHTML || strings.Join(f.Tags, " ") != "chapter1 vocab" {
		t.Fatalf("file = %+v", f)
	}
	c, basic, deck := openCollection(t, storage.NewMemory())
	importText(t, c, basic, deck, text, core.DupSkip)
	if got := notes(t, c)[0]; got != "a &lt; b &amp; c|x&gt;y<br>z" {
		t.Fatalf("note = %q", got)
	}
	// Header tags (spec): every note has them.
	importText(t, c, basic, deck, "#tags:chapter1 vocab\nhola,hello\nperro,dog\n", core.DupSkip)
	ids, _ := c.Search("tag:chapter1 tag:vocab", core.ColCreated, false)
	if len(ids) != 3 {
		t.Fatalf("tagged notes = %d", len(ids))
	}
	// With HTML allowed, the same text is markup.
	f, _ = Parse([]byte("<b>x</b> &amp; y\tz\n"))
	b, _ := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: DefaultMapping(2, 2), AllowHTML: true})
	if got := b.Notes[0].Fields[0]; got != "<b>x</b> &amp; y" {
		t.Fatalf("html field = %q", got)
	}
	if f.NoHTML || f.HTML {
		t.Fatalf("no html header: %+v", f)
	}
}

func TestHeaderGuess(t *testing.T) {
	if !LooksLikeHeader([]string{"front", "Back"}, []string{"Front", "Back"}) {
		t.Fatal("header not recognised")
	}
	if LooksLikeHeader([]string{"hola", "hello"}, []string{"Front", "Back"}) {
		t.Fatal("data row taken for a header")
	}
}

func TestDuplicates(t *testing.T) {
	c, basic, deck := openCollection(t, storage.NewMemory())
	importText(t, c, basic, deck, "hola\thello\n", core.DupSkip)

	// Skip duplicates (spec): no new note, counted as skipped.
	sum := importText(t, c, basic, deck, "hola\thi\n<b>hola</b>\thi again\nadiós\tbye\n", core.DupSkip)
	if sum.NotesAdded != 1 || sum.NotesSkipped != 2 {
		t.Fatalf("skip summary = %+v", sum)
	}
	if got := strings.Join(notes(t, c), ","); got != "hola|hello,adiós|bye" {
		t.Fatalf("notes = %q", got)
	}
	// Update changes the existing note instead.
	sum = importText(t, c, basic, deck, "hola\thi\n", core.DupUpdate)
	if sum.NotesUpdated != 1 || sum.NotesAdded != 0 {
		t.Fatalf("update summary = %+v", sum)
	}
	if got := notes(t, c)[0]; got != "hola|hi" {
		t.Fatalf("updated note = %q", got)
	}
	// Duplicate adds it again.
	sum = importText(t, c, basic, deck, "hola\tyet another\n", core.DupDuplicate)
	if sum.NotesAdded != 1 || len(notes(t, c)) != 3 {
		t.Fatalf("duplicate summary = %+v", sum)
	}
	// Rows within one file count as duplicates of each other.
	sum = importText(t, c, basic, deck, "x\t1\nx\t2\n", core.DupSkip)
	if sum.NotesAdded != 1 || sum.NotesSkipped != 1 {
		t.Fatalf("in-file duplicate summary = %+v", sum)
	}
}

func TestRowsWithoutCardsAreWarnings(t *testing.T) {
	c, basic, deck := openCollection(t, storage.NewMemory())
	sum := importText(t, c, basic, deck, "\tonly back\nfront\tback\n", core.DupDuplicate)
	if sum.NotesAdded != 1 || sum.NotesSkipped != 1 || len(sum.Warnings) != 1 {
		t.Fatalf("summary = %+v", sum)
	}
}

func TestTagsColumnAndMapping(t *testing.T) {
	c, basic, deck := openCollection(t, storage.NewMemory())
	f, _ := Parse([]byte("Front,Back,Tags\nhola,hello,verb spanish\n"))
	if !LooksLikeHeader(f.Rows[0], []string{"Front", "Back"}) {
		t.Fatal("header not guessed")
	}
	b, err := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: []int{1, 0, Tags}, HasHeader: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	ids, _ := c.Search("tag:spanish tag:verb", core.ColCreated, false)
	if len(ids) != 1 {
		t.Fatalf("tagged notes = %v", ids)
	}
	n, _ := c.Note(c.Rows(ids)[0].NoteID)
	if n.Fields[0] != "hello" || n.Fields[1] != "hola" {
		t.Fatalf("mapped fields = %q", n.Fields)
	}
	if _, err := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: []int{Ignore, Ignore, Tags}}); err == nil {
		t.Fatal("mapping without fields accepted")
	}
}

type failingAppend struct{ storage.Backend }

func (failingAppend) Append(string, []byte) error { return errors.New("disk full") }

// An import is one journal entry: if saving fails nothing is added.
func TestImportIsAllOrNothing(t *testing.T) {
	store := storage.NewMemoryStore()
	c, basic, deck := openCollection(t, store.Open())
	f, _ := Parse([]byte("a\t1\nb\t2\nc\t3\n"))
	b, _ := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: DefaultMapping(2, 2)})

	failing, _, _ := core.Open(failingAppend{store.Open()}, core.Options{})
	if _, err := failing.Import(b, nil); err == nil {
		t.Fatal("import succeeded on a failing store")
	}
	if n := len(notes(t, failing)); n != 0 {
		t.Fatalf("%d notes after failed import", n)
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	reopened, _, _ := core.Open(store.Open(), core.Options{})
	if n := len(notes(t, reopened)); n != 3 {
		t.Fatalf("%d notes after reopen", n)
	}
	_ = basic
}

// notesByGUID returns the collection's notes by GUID.
func notesByGUID(t *testing.T, c *core.Collection) map[string]core.Note {
	t.Helper()
	ids, _ := c.Search("", core.ColCreated, false)
	out := map[string]core.Note{}
	for _, row := range c.Rows(ids) {
		n, _ := c.Note(row.NoteID)
		out[n.GUID] = n
	}
	return out
}

func TestRoleHeaders(t *testing.T) {
	text := "#separator:tab\n#html:true\n#notetype:Cloze\n#deck:Spanish::Verbs\n" +
		"#guid column:1\n#notetype column:2\n#deck column:3\n#tags column:6\n#junk column:x\n" +
		"g\tBasic\tD\ta\tb\tt\n"
	f, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if f.NoteType != "Cloze" || f.Deck != "Spanish::Verbs" || len(f.Rows) != 1 {
		t.Fatalf("file = %+v", f)
	}
	if m := f.Mapping(2); fmt.Sprint(m) != fmt.Sprint([]int{GUID, NoteType, Deck, 0, 1, Tags}) {
		t.Fatalf("mapping = %v", m)
	}
	// Columns the rows do not have are not roles; with fewer fields, extra
	// columns are ignored.
	f, _ = Parse([]byte("#tags column:9\n#guid column:2\na\tg\tb\tc\n"))
	if m := f.Mapping(2); fmt.Sprint(m) != fmt.Sprint([]int{0, GUID, 1, Ignore}) {
		t.Fatalf("mapping = %v (roles %v)", m, f.Roles)
	}
}

func TestAnkiNotesExport(t *testing.T) {
	// Anki notes export (spec): each note in its own deck, with its own
	// note type and GUID. Fields are padded to the largest note type, and
	// the tags are the last column.
	c, basic, deck := openCollection(t, storage.NewMemory())
	if _, err := c.AddNoteType(core.StockOptionalReversed, core.StockOptionalReversed); err != nil {
		t.Fatal(err)
	}
	text := "#separator:tab\n#html:true\n#guid column:1\n#notetype column:2\n#deck column:3\n#tags column:7\n" +
		"g1\tBasic\tSpanish\thola\thello\t\tverb\n" +
		"g2\tcloze\tScience::Bio\t{{c1::Mitochondria}} make energy\tcells\t\tbio cell\n" +
		"g3\t" + core.StockOptionalReversed + "\tSpanish\tperro\tdog\ty\tnoun\n" +
		"g4\tBasic\t\tsolo\tmio\tleft over\t\n"
	f, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: f.Mapping(3), AllowHTML: true})
	if err != nil {
		t.Fatal(err)
	}
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if sum.NotesAdded != 4 || sum.CardsAdded != 5 || len(sum.Warnings) != 1 || !strings.Contains(sum.Warnings[0], "1 row had more fields") {
		t.Fatalf("summary = %+v", sum)
	}
	types := map[core.ID]string{}
	for _, nt := range c.NoteTypes() {
		types[nt.ID] = nt.Name
	}
	got := notesByGUID(t, c)
	want := map[string]struct{ typ, deck, fields, tags string }{
		"g1": {"Basic", "Spanish", "hola|hello", "verb"},
		"g2": {"Cloze", "Science::Bio", "{{c1::Mitochondria}} make energy|cells", "bio cell"},
		"g3": {core.StockOptionalReversed, "Spanish", "perro|dog|y", "noun"},
		// Without a deck, the default deck; the extra field is left out.
		"g4": {"Basic", "Default", "solo|mio", ""},
	}
	for guid, w := range want {
		n, ok := got[guid]
		if !ok {
			t.Fatalf("no note %s", guid)
		}
		d, _ := c.Deck(c.CardsOfNote(n.ID)[0].DeckID)
		if types[n.TypeID] != w.typ || d.Name != w.deck || strings.Join(n.Fields, "|") != w.fields || strings.Join(n.Tags, " ") != w.tags {
			t.Errorf("%s: %s in %s, %q, tags %q; want %+v", guid, types[n.TypeID], d.Name, n.Fields, n.Tags, w)
		}
	}
}

func TestUnknownNoteType(t *testing.T) {
	// Unknown note type (spec): the row is skipped and the summary names
	// the note type.
	c, basic, deck := openCollection(t, storage.NewMemory())
	f, _ := Parse([]byte("#notetype column:1\nBasic\ta\tb\nVocab\tc\td\nVocab\te\tf\n\tg\th\n"))
	b, err := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: f.Mapping(2)})
	if err != nil {
		t.Fatal(err)
	}
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if sum.NotesAdded != 2 || sum.NotesSkipped != 2 || len(sum.Warnings) != 1 || usermsg.Plain(sum.Warnings[0]) != `Skipped 2 rows: there is no note type named "Vocab".` {
		t.Fatalf("summary = %+v", sum)
	}
	if got := strings.Join(notes(t, c), ","); got != "a|b,g|h" {
		t.Fatalf("notes = %q", got)
	}
	// One column per role.
	if _, err := Build(f, Options{TypeID: basic, DeckID: deck, NoteTypes: c.NoteTypes(), Mapping: []int{NoteType, NoteType, 0}}); err == nil {
		t.Fatal("two note type columns accepted")
	}
}

func TestDeckColumnAndDefaultDeckName(t *testing.T) {
	c, basic, _ := openCollection(t, storage.NewMemory())
	f, _ := Parse([]byte("a\tb\tLang::French\nc\td\t\ne\tf\tbad::::name\n"))
	b, err := Build(f, Options{TypeID: basic, DeckName: "Inbox", NoteTypes: c.NoteTypes(), Mapping: []int{0, 1, Deck}})
	if err != nil {
		t.Fatal(err)
	}
	sum, err := c.Import(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	// A deck name that is not valid skips its row, not the file.
	if sum.NotesAdded != 2 || sum.NotesSkipped != 1 || len(sum.Warnings) != 1 {
		t.Fatalf("summary = %+v", sum)
	}
	for q, want := range map[string]int{`deck:"Lang::French"`: 1, "deck:Inbox": 1} {
		if ids, _ := c.Search(q, core.ColCreated, false); len(ids) != want {
			t.Errorf("%s: %d cards", q, len(ids))
		}
	}
}
