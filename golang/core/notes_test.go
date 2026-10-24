package core

import (
	"cardflip/golang/internal/usermsg"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

func noteType(t *testing.T, c *Collection, name string) NoteType {
	t.Helper()
	for _, nt := range c.NoteTypes() {
		if nt.Name == name {
			return nt
		}
	}
	t.Fatalf("no note type %q", name)
	return NoteType{}
}

func defaultDeck(c *Collection) ID { return c.Decks()[0].ID }

func addNote(t *testing.T, c *Collection, typeName string, fields ...string) ID {
	t.Helper()
	id, err := c.AddNote(NoteInput{TypeID: noteType(t, c, typeName).ID, DeckID: defaultDeck(c), Fields: fields})
	if err != nil {
		t.Fatalf("add %s %v: %v", typeName, fields, err)
	}
	return id
}

func questions(t *testing.T, c *Collection, noteID ID) []string {
	t.Helper()
	var out []string
	for _, card := range c.CardsOfNote(noteID) {
		out = append(out, c.QuestionText(card.ID))
	}
	return out
}

func TestReversedCard(t *testing.T) {
	c, store, clk := testCollection(t)
	// Reversed card (spec): two cards, asking hola and hello.
	id := addNote(t, c, "Basic (and reversed card)", "hola", "hello")
	if got := questions(t, c, id); fmt.Sprint(got) != "[hola hello]" {
		t.Fatalf("questions = %v", got)
	}
	for _, card := range c.CardsOfNote(id) {
		// New card has no memory state (scheduling spec).
		if card.State != scheduler.New || card.Memory != nil || card.Reps != 0 || card.Lapses != 0 {
			t.Fatalf("new card = %+v", card)
		}
	}
	reopen(t, c, store, clk)
}

func TestClozeCards(t *testing.T) {
	c, store, clk := testCollection(t)
	// Cloze cards (spec): one card per cloze number.
	id := addNote(t, c, "Cloze", "{{c1::Paris}} is the capital of {{c2::France}}", "")
	cards := c.CardsOfNote(id)
	if len(cards) != 2 || cards[0].Ord != 0 || cards[1].Ord != 1 {
		t.Fatalf("cards = %+v", cards)
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[[...] is the capital of France Paris is the capital of [...]]" {
		t.Fatalf("questions = %q", got)
	}

	// Editing removes a cloze (spec): card 2 is kept as an empty card, card
	// 1 keeps its scheduling state.
	c1 := cards[0]
	c1.CardState = scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 10),
		Memory: &scheduler.Memory{Stability: 10, Difficulty: 5}, ScheduledDays: 10, Reps: 3, LastReview: clk.t}
	c.mu.Lock()
	if err := c.commit(&op{Kind: "test", Cards: []*Card{&c1}}); err != nil {
		t.Fatal(err)
	}
	c.mu.Unlock()
	clk.advance(time.Hour)
	if err := c.UpdateNote(id, []string{"{{c1::Paris}} is the capital of France", ""}, nil); err != nil {
		t.Fatal(err)
	}
	after := c.CardsOfNote(id)
	if len(after) != 2 || after[0].ID != c1.ID || after[1].ID != cards[1].ID {
		t.Fatalf("cards after edit = %+v", after)
	}
	if after[0].State != scheduler.Review || after[0].Reps != 3 || !after[0].Due.Equal(c1.Due) {
		t.Fatalf("card 1 lost its scheduling: %+v", after[0])
	}
	if empty := c.EmptyCards(); empty.Total != 1 || empty.Notes[0].Cards[0].CardID != cards[1].ID {
		t.Fatalf("empty cards = %+v", empty)
	}
	// Adding the deletion back uses card 2 again; a new number adds card 3.
	if err := c.UpdateNote(id, []string{"{{c1::Paris}} is the capital of {{c2::France}} {{c3::!}}", ""}, nil); err != nil {
		t.Fatal(err)
	}
	if after := c.CardsOfNote(id); len(after) != 3 || after[1].ID != cards[1].ID {
		t.Fatalf("cards = %+v", after)
	}
	if empty := c.EmptyCards(); empty.Total != 0 {
		t.Fatalf("empty cards = %+v", empty)
	}
	reopen(t, c, store, clk)
}

func TestMultiNumberClozeCards(t *testing.T) {
	c, store, clk := testCollection(t)
	// Two cards from one deletion (spec).
	id := addNote(t, c, "Cloze", "{{c1,2::Paris}} is the capital of France", "")
	cards := c.CardsOfNote(id)
	if len(cards) != 2 || cards[0].Ord != 0 || cards[1].Ord != 1 {
		t.Fatalf("cards = %+v", cards)
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[[...] is the capital of France [...] is the capital of France]" {
		t.Fatalf("questions = %q", got)
	}

	// Removing a number (spec): card 2 is kept as an empty card, card 1
	// keeps its scheduling state.
	c1 := cards[0]
	c1.CardState = scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 10),
		Memory: &scheduler.Memory{Stability: 10, Difficulty: 5}, ScheduledDays: 10, Reps: 3, LastReview: clk.t}
	c.mu.Lock()
	if err := c.commit(&op{Kind: "test", Cards: []*Card{&c1}}); err != nil {
		t.Fatal(err)
	}
	c.mu.Unlock()
	clk.advance(time.Hour)
	if err := c.UpdateNote(id, []string{"{{c1::Paris}} is the capital of France", ""}, nil); err != nil {
		t.Fatal(err)
	}
	after := c.CardsOfNote(id)
	if len(after) != 2 || after[0].ID != c1.ID || after[1].ID != cards[1].ID {
		t.Fatalf("cards after edit = %+v", after)
	}
	if after[0].State != scheduler.Review || after[0].Reps != 3 || !after[0].Due.Equal(c1.Due) {
		t.Fatalf("card 1 lost its scheduling: %+v", after[0])
	}
	if empty := c.EmptyCards(); empty.Total != 1 || empty.Notes[0].Cards[0].CardID != cards[1].ID {
		t.Fatalf("empty cards = %+v", empty)
	}
	reopen(t, c, store, clk)
}

func TestEmptyNoteRejected(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	// Empty front (spec): not added, the error says it would produce no cards.
	_, err := c.AddNote(NoteInput{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"", "back"}})
	mustUserErr(t, err, "would produce no cards")
	mustUserErr(t, err, "Front")
	_, err = c.AddNote(NoteInput{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{" <br> ", "back"}})
	mustUserErr(t, err, "would produce no cards")
	_, err = c.AddNote(NoteInput{TypeID: noteType(t, c, "Cloze").ID, DeckID: defaultDeck(c), Fields: []string{"no deletions", ""}})
	mustUserErr(t, err, "cloze deletion")
	if len(c.notes) != 0 || len(c.cards) != 0 {
		t.Fatal("something was added")
	}
	// An edit that empties the note is rejected as well.
	id := addNote(t, c, "Basic", "front", "back")
	mustUserErr(t, c.UpdateNote(id, []string{"", "back"}, nil), "would produce no cards")
	if n, _ := c.Note(id); n.Fields[0] != "front" {
		t.Fatal("rejected edit was applied")
	}
	// Too many fields.
	_, err = c.AddNote(NoteInput{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"a", "b", "c"}})
	mustUserErr(t, err, "fields")
}

func TestReversedCardFollowsEdits(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic (and reversed card)", "hola", "")
	if n := len(c.CardsOfNote(id)); n != 1 {
		t.Fatalf("cards with empty back = %d", n)
	}
	c.UpdateNote(id, []string{"hola", "hello"}, nil)
	if got := questions(t, c, id); fmt.Sprint(got) != "[hola hello]" {
		t.Fatalf("questions = %v", got)
	}
	// Emptying the back keeps the reverse card as an empty card.
	c.UpdateNote(id, []string{"hola", ""}, nil)
	if n := len(c.CardsOfNote(id)); n != 2 {
		t.Fatalf("cards after emptying back = %d", n)
	}
	if empty := c.EmptyCards(); empty.Total != 1 || empty.Notes[0].Cards[0].Name != "Card 2" {
		t.Fatalf("empty cards = %+v", empty)
	}
}

func TestTags(t *testing.T) {
	c, store, clk := testCollection(t)
	// Add tag with spaces (spec): two tags.
	id, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic").ID, DeckID: defaultDeck(c),
		Fields: []string{"a", "b"}, Tags: []string{"verb irregular", "Verb", " lang::es "}})
	if err != nil {
		t.Fatal(err)
	}
	n, _ := c.Note(id)
	if strings.Join(n.Tags, ",") != "verb,irregular,lang::es" {
		t.Fatalf("tags = %q", n.Tags)
	}
	id2 := addNote(t, c, "Basic", "c", "d")
	if err := c.AddTags([]ID{id, id2}, []string{"noun IRREGULAR"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(c.AllTags(), ","); got != "irregular,lang::es,noun,verb" {
		t.Fatalf("all tags = %q", got)
	}
	if err := c.RemoveTags([]ID{id}, []string{"Irregular"}); err != nil {
		t.Fatal(err)
	}
	n, _ = c.Note(id)
	if strings.Join(n.Tags, ",") != "verb,lang::es,noun" {
		t.Fatalf("tags after remove = %q", n.Tags)
	}
	reopen(t, c, store, clk)
}

func TestToggleMark(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic (and reversed card)", "a", "1")
	b := addNote(t, c, "Basic", "b", "2")
	marked := func(id ID) bool {
		n, _ := c.Note(id)
		return slices.Contains(n.Tags, "marked")
	}

	// Mark a note (spec), then toggle it off again.
	if err := c.ToggleMark([]ID{a}); err != nil {
		t.Fatal(err)
	}
	if !marked(a) || !c.IsMarked(a) || marked(b) {
		t.Fatal("note not marked")
	}
	cards := c.CardsOfNote(a)
	if rows := c.Rows([]ID{cards[0].ID, cards[1].ID}, ColSortField); len(rows) != 2 || !rows[0].Marked || !rows[1].Marked {
		t.Fatalf("rows = %+v", rows)
	}
	if rows := c.NoteRows([]ID{a, b}); !rows[0].Marked || rows[1].Marked {
		t.Fatalf("note rows = %+v", rows)
	}

	// Any note lacking the tag: all get it. All have it: all lose it.
	if err := c.ToggleMark([]ID{a, b}); err != nil {
		t.Fatal(err)
	}
	if !marked(a) || !marked(b) {
		t.Fatal("mixed selection not marked")
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Mark 1 note" {
		t.Fatalf("undo name = %q", undo)
	}
	if err := c.ToggleMark([]ID{a, b}); err != nil {
		t.Fatal(err)
	}
	if undo, _ := c.UndoStatus(); marked(a) || marked(b) || usermsg.Plain(undo) != "Unmark 2 notes" {
		t.Fatalf("still marked, undo name = %q", undo)
	}

	// The tag is an ordinary tag: one typed in another case counts and is
	// removed, and other tags are kept.
	if err := c.UpdateNote(b, []string{"b", "2"}, []string{"Marked", "verb"}); err != nil {
		t.Fatal(err)
	}
	if !c.IsMarked(b) {
		t.Fatal("Marked not recognised")
	}
	if err := c.ToggleMark([]ID{b}); err != nil {
		t.Fatal(err)
	}
	if n, _ := c.Note(b); strings.Join(n.Tags, ",") != "verb" {
		t.Fatalf("tags = %q", n.Tags)
	}
	reopen(t, c, store, clk)
}

func TestRememberedPickers(t *testing.T) {
	c, _, _ := testCollection(t)
	deck, _ := c.CreateDeck("Spanish")
	cloze := noteType(t, c, "Cloze")
	if _, err := c.AddNote(NoteInput{TypeID: cloze.ID, DeckID: deck, Fields: []string{"{{c1::x}}"}}); err != nil {
		t.Fatal(err)
	}
	if s := c.Settings(); s.LastDeckID != deck || s.LastNoteTypeID != cloze.ID {
		t.Fatalf("settings = %+v", s)
	}
}

func TestRemoveField(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	a := addNote(t, c, "Basic", "front a", "back a")
	b := addNote(t, c, "Basic", "front b", "back b")
	other := addNote(t, c, "Basic (and reversed card)", "x", "y")
	if n := c.NoteCount(basic.ID); n != 2 {
		t.Fatalf("note count = %d", n)
	}
	// Remove a field (spec): the field and its content are removed from all
	// notes of that type.
	if err := c.RemoveField(basic.ID, 1); err != nil {
		t.Fatal(err)
	}
	nt, _ := c.NoteType(basic.ID)
	if strings.Join(nt.Fields, ",") != "Front" || strings.Contains(nt.Templates[0].AFmt, "Back") {
		t.Fatalf("note type = %+v", nt)
	}
	for _, id := range []ID{a, b} {
		n, _ := c.Note(id)
		if len(n.Fields) != 1 || !strings.HasPrefix(n.Fields[0], "front") {
			t.Fatalf("note fields = %q", n.Fields)
		}
	}
	if n, _ := c.Note(other); len(n.Fields) != 2 {
		t.Fatal("other note type changed")
	}
	// Removing the only remaining field is refused.
	mustUserErr(t, c.RemoveField(basic.ID, 0), "at least one field")
	reopen(t, c, store, clk)
}

func TestRemoveFieldKeepsEmptyCards(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	id := addNote(t, c, "Basic", "front", "back")
	card := c.CardsOfNote(id)[0]
	if err := c.RemoveField(basic.ID, 0); err != nil {
		t.Fatal(err)
	}
	if after := c.CardsOfNote(id); len(after) != 1 || after[0].ID != card.ID {
		t.Fatalf("cards = %+v", after)
	}
}

func TestTemplateTypoKeepsCards(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	var ids []ID
	for i := 0; i < 1000; i++ {
		ids = append(ids, addNote(t, c, "Basic", fmt.Sprint("front ", i), ""))
	}
	before := len(c.cards)
	// Template typo (spec): no card is deleted, the 1,000 cards are empty.
	if err := c.UpdateTemplate(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Back}}", AFmt: "{{Back}}"}); err != nil {
		t.Fatal(err)
	}
	if len(c.cards) != before {
		t.Fatalf("cards = %d, want %d", len(c.cards), before)
	}
	empty := c.EmptyCards()
	if empty.Total != 1000 || len(empty.Notes) != 1000 || !empty.Notes[0].AllEmpty {
		t.Fatalf("empty cards: total %d, notes %d", empty.Total, len(empty.Notes))
	}
	// Fixing the typo makes them normal cards again.
	if err := c.UpdateTemplate(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Back}}"}); err != nil {
		t.Fatal(err)
	}
	if empty := c.EmptyCards(); empty.Total != 0 || len(c.cards) != before {
		t.Fatalf("empty after fix = %d, cards = %d", empty.Total, len(c.cards))
	}
	c = reopen(t, c, store, clk)
	if len(c.CardsOfNote(ids[0])) != 1 {
		t.Fatal("card lost after reopen")
	}
}

func TestTemplateChecks(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	reversed := noteType(t, c, "Basic (and reversed card)")
	cloze := noteType(t, c, "Cloze")
	if err := c.AddField(basic.ID, "Hint"); err != nil {
		t.Fatal(err)
	}
	id := addNote(t, c, "Basic (and reversed card)", "hola", "hello")
	cards := c.CardsOfNote(id)
	undo, _ := c.UndoStatus()

	refused := func(typeID ID, index int, tmpl CardTemplate, contains string) {
		t.Helper()
		before, _ := c.NoteType(typeID)
		mustUserErr(t, c.UpdateTemplate(typeID, index, tmpl), contains)
		// A refused save changes nothing.
		if after, _ := c.NoteType(typeID); !reflect.DeepEqual(after, before) {
			t.Fatalf("note type changed: %+v", after)
		}
		if got := c.CardsOfNote(id); !reflect.DeepEqual(got, cards) {
			t.Fatalf("cards changed: %+v", got)
		}
		if u, _ := c.UndoStatus(); u != undo {
			t.Fatalf("undo = %q, want %q", u, undo)
		}
	}
	// Unknown field (spec).
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Frnt}}", AFmt: "{{Back}}"}, `there is no field "Frnt"`)
	// Unclosed section (spec).
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}{{#Hint}}{{Hint}}", AFmt: "{{Back}}"}, "{{#Hint}} is not closed")
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Back}}{{/Hint}}"}, "{{/Hint}} has no matching {{#Hint}}")
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "Question", AFmt: "{{Back}}"}, "uses no field")
	// Unclosed tag, tag without a field name, crossed sections (spec).
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back"}, "{{Back has no closing }}")
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}{{}}", AFmt: "{{Back}}"}, "{{}} has no field name")
	refused(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}{{#Hint}}{{#Back}}{{/Hint}}{{/Back}}", AFmt: ""},
		"In the front template, {{/Hint}} closes {{#Hint}} before {{#Back}} is closed.")
	// Missing cloze filter (spec).
	refused(cloze.ID, 0, CardTemplate{Name: "Cloze", QFmt: "{{Text}}", AFmt: "{{Text}}"}, "must use {{cloze:Text}} or another cloze field")
	// Duplicate front (spec).
	refused(reversed.ID, 1, CardTemplate{Name: "Card 2", QFmt: "{{Front}}", AFmt: "{{Back}}"}, "would show the same question")
	mustUserErr(t, c.AddTemplate(reversed.ID, CardTemplate{Name: "Card 3", QFmt: " {{Back}}\n", AFmt: ""}), `Card type "Card 2" has the same front`)

	// Live feedback reports the same problems without saving.
	problems, err := c.CheckTemplate(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Frnt}}", AFmt: "{{Back}}"})
	if err != nil || len(problems) != 1 || !strings.Contains(problems[0].Message, `no field "Frnt"`) {
		t.Fatalf("problems = %v, %v", problems, err)
	}
	if problems, _ := c.CheckTemplate(reversed.ID, 1, reversed.Templates[1]); problems != nil {
		t.Fatalf("unchanged card type: problems = %v", problems)
	}
	if problems, _ := c.CheckTemplate(reversed.ID, -1, reversed.Templates[1]); len(problems) != 1 {
		t.Fatalf("new card type like Card 2: problems = %v", problems)
	}

	// Valid edits still save.
	if err := c.UpdateTemplate(basic.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}{{#Hint}}{{hint:Hint}}{{/Hint}}", AFmt: "{{FrontSide}}<hr id=answer>{{Back}} {{Tags}}"}); err != nil {
		t.Fatal(err)
	}
	reopen(t, c, store, clk)
}

// The card types of every stock note type pass the template checks.
func TestStockTemplatesPassChecks(t *testing.T) {
	for _, nt := range stockNoteTypes() {
		for i, tmpl := range nt.Templates {
			if p := templateProblems(nt, tmpl, i); p != nil {
				t.Errorf("%s card type %d: %v", nt.Name, i, p)
			}
		}
	}
}

func TestRenameAndAddFields(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	id := addNote(t, c, "Basic", "hola", "hello")
	if err := c.RenameField(basic.ID, 0, "Spanish"); err != nil {
		t.Fatal(err)
	}
	nt, _ := c.NoteType(basic.ID)
	if nt.Fields[0] != "Spanish" || nt.Templates[0].QFmt != "{{Spanish}}" {
		t.Fatalf("after rename: %+v", nt)
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[hola]" {
		t.Fatalf("question after rename = %v", got)
	}
	mustUserErr(t, c.RenameField(basic.ID, 1, "spanish"), "already has a field")
	mustUserErr(t, c.RenameField(basic.ID, 1, "a:b"), "cannot contain")
	if err := c.AddField(basic.ID, "Notes"); err != nil {
		t.Fatal(err)
	}
	if n, _ := c.Note(id); len(n.Fields) != 3 || n.Fields[2] != "" {
		t.Fatalf("fields after add = %q", n.Fields)
	}
	mustUserErr(t, c.RenameNoteType(basic.ID, "cloze"), "already exists")
	if err := c.RenameNoteType(basic.ID, "Vocabulary"); err != nil {
		t.Fatal(err)
	}
	reopen(t, c, store, clk)
}

func TestTemplates(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	id := addNote(t, c, "Basic", "hola", "hello")
	first := c.CardsOfNote(id)[0]

	// A new template creates cards for existing notes.
	if err := c.AddTemplate(basic.ID, CardTemplate{Name: "Reverse", QFmt: "{{Back}}", AFmt: "{{Front}}"}); err != nil {
		t.Fatal(err)
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[hola hello]" {
		t.Fatalf("questions = %v", got)
	}
	mustUserErr(t, c.AddTemplate(basic.ID, CardTemplate{Name: "reverse", QFmt: "x"}), "already has a card type")
	mustUserErr(t, c.AddTemplate(noteType(t, c, "Cloze").ID, CardTemplate{Name: "Two", QFmt: "{{Text}}"}), "exactly one")

	// Removing the first template deletes its cards and renumbers the rest.
	if n := c.TemplateCardCount(basic.ID, 0); n != 1 {
		t.Fatalf("template card count = %d", n)
	}
	if err := c.RemoveTemplate(basic.ID, 0); err != nil {
		t.Fatal(err)
	}
	cards := c.CardsOfNote(id)
	if len(cards) != 1 || cards[0].Ord != 0 || cards[0].ID == first.ID {
		t.Fatalf("cards after removing template 0: %+v", cards)
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[hello]" {
		t.Fatalf("questions = %v", got)
	}
	mustUserErr(t, c.RemoveTemplate(basic.ID, 0), "at least one card type")

	// Changing the question format regenerates cards.
	if err := c.UpdateTemplate(basic.ID, 0, CardTemplate{Name: "Card", QFmt: "Q: {{Front}}", AFmt: "{{Back}}"}); err != nil {
		t.Fatal(err)
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[Q: hola]" {
		t.Fatalf("questions = %v", got)
	}
	reopen(t, c, store, clk)
}

func TestDeleteNotes(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic (and reversed card)", "a", "b")
	b := addNote(t, c, "Basic", "c", "d")
	if err := c.DeleteNotes([]ID{a}); err != nil {
		t.Fatal(err)
	}
	if len(c.cards) != 1 || len(c.notes) != 1 {
		t.Fatalf("%d notes %d cards left", len(c.notes), len(c.cards))
	}
	if _, err := c.Note(b); err != nil {
		t.Fatal(err)
	}
	reopen(t, c, store, clk)
}

func TestFieldOptionsFollowFields(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	// A note type without stored options has the defaults for every field.
	if nt, _ := c.NoteType(basic.ID); nt.FieldOptions != nil || nt.FieldOption(1).HTMLByDefault {
		t.Fatalf("default options = %+v", nt.FieldOptions)
	}
	html := FieldOptions{HTMLByDefault: true}
	if err := c.SetFieldOptions(basic.ID, 1, html); err != nil {
		t.Fatal(err)
	}
	mustUserErr(t, c.SetFieldOptions(basic.ID, 2, html), "no longer exists")
	opts := func() []FieldOptions {
		nt, _ := c.NoteType(basic.ID)
		return nt.FieldOptions
	}
	if got := opts(); !reflect.DeepEqual(got, []FieldOptions{{}, html}) {
		t.Fatalf("options = %+v", got)
	}
	// Added fields get the defaults; renamed fields keep their options.
	if err := c.AddField(basic.ID, "Notes"); err != nil {
		t.Fatal(err)
	}
	if err := c.RenameField(basic.ID, 1, "English"); err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(basic.ID); !nt.FieldOption(1).HTMLByDefault || nt.FieldOption(2).HTMLByDefault {
		t.Fatalf("options after add and rename = %+v", nt.FieldOptions)
	}
	if err := c.SetFieldOptions(basic.ID, 2, html); err != nil {
		t.Fatal(err)
	}
	// Removing a field removes its options and shifts the others.
	if err := c.RemoveField(basic.ID, 0); err != nil {
		t.Fatal(err)
	}
	if got := opts(); !reflect.DeepEqual(got, []FieldOptions{html, html}) {
		t.Fatalf("options after remove = %+v", got)
	}
	// Turning the options off again leaves nothing stored.
	for i := range 2 {
		if err := c.SetFieldOptions(basic.ID, i, FieldOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	if got := opts(); got != nil {
		t.Fatalf("options after reset = %+v", got)
	}
	if err := c.SetFieldOptions(basic.ID, 0, html); err != nil {
		t.Fatal(err)
	}
	// Undo restores the previous options.
	if name, err := c.Undo(); err != nil || usermsg.Plain(name) != "Change field options" {
		t.Fatalf("undo = %q, %v", name, err)
	}
	if got := opts(); got != nil {
		t.Fatalf("options after undo = %+v", got)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	c = reopen(t, c, store, clk)
	if got := opts(); !reflect.DeepEqual(got, []FieldOptions{html}) {
		t.Fatalf("options after reopen = %+v", got)
	}
}

func TestStickyFieldOptions(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	sticky := FieldOptions{Sticky: true}
	if err := c.SetFieldOptions(basic.ID, 1, sticky); err != nil {
		t.Fatal(err)
	}
	// Pins follow their field through add, rename and remove.
	if err := c.AddField(basic.ID, "Source"); err != nil {
		t.Fatal(err)
	}
	if err := c.RenameField(basic.ID, 1, "English"); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveField(basic.ID, 0); err != nil {
		t.Fatal(err)
	}
	c = reopen(t, c, store, clk)
	nt, _ := c.NoteType(basic.ID)
	if !reflect.DeepEqual(nt.Fields, []string{"English", "Source"}) || !nt.FieldOption(0).Sticky || nt.FieldOption(1).Sticky {
		t.Fatalf("fields %v, options %+v", nt.Fields, nt.FieldOptions)
	}
}

func TestMoveField(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	id := addNote(t, c, "Basic", "front", "back")
	card := c.CardsOfNote(id)[0]
	opts := FieldOptions{Sticky: true, Font: "Noto Sans", Size: 24, RTL: true, Collapsed: true, ExcludeFromSearch: true}
	if err := c.SetFieldOptions(basic.ID, 1, opts); err != nil {
		t.Fatal(err)
	}
	// Move Back first (spec): the editor lists Back first, the note's
	// values stay in their fields and the card shows the same content.
	if err := c.MoveField(basic.ID, 1, 0); err != nil {
		t.Fatal(err)
	}
	nt, _ := c.NoteType(basic.ID)
	if !reflect.DeepEqual(nt.Fields, []string{"Back", "Front"}) || nt.FieldOption(0) != opts || nt.FieldOption(1) != (FieldOptions{}) {
		t.Fatalf("fields %v, options %+v", nt.Fields, nt.FieldOptions)
	}
	if n, _ := c.Note(id); !reflect.DeepEqual(n.Fields, []string{"back", "front"}) {
		t.Fatalf("note fields = %q", n.Fields)
	}
	if cards := c.CardsOfNote(id); len(cards) != 1 || cards[0].ID != card.ID || c.QuestionText(card.ID) != "front" {
		t.Fatalf("cards = %+v, question %q", cards, c.QuestionText(card.ID))
	}
	mustUserErr(t, c.MoveField(basic.ID, 0, 2), "no longer exists")
	mustUserErr(t, c.MoveField(basic.ID, -1, 0), "no longer exists")
	if name, err := c.Undo(); err != nil || usermsg.Plain(name) != "Move field" {
		t.Fatalf("undo = %q, %v", name, err)
	}
	if n, _ := c.Note(id); !reflect.DeepEqual(n.Fields, []string{"front", "back"}) {
		t.Fatalf("note fields after undo = %q", n.Fields)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	c = reopen(t, c, store, clk)
	if nt, _ := c.NoteType(basic.ID); nt.Fields[0] != "Back" || nt.FieldOption(0) != opts {
		t.Fatalf("after reopen: fields %v, options %+v", nt.Fields, nt.FieldOptions)
	}
}

func TestSortField(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := noteType(t, c, "Basic")
	if err := c.AddField(basic.ID, "Notes"); err != nil {
		t.Fatal(err)
	}
	sortField := func() int {
		t.Helper()
		nt, _ := c.NoteType(basic.ID)
		return nt.SortFieldIndex()
	}
	// The first field sorts by default.
	if got := sortField(); got != 0 {
		t.Fatalf("default sort field = %d", got)
	}
	// Sort field follows move (spec): Front, moved to the second position,
	// is still the sort field.
	if err := c.MoveField(basic.ID, 0, 1); err != nil {
		t.Fatal(err)
	}
	if got := sortField(); got != 1 {
		t.Fatalf("sort field after moving it = %d", got)
	}
	// Moving another field across it shifts it too.
	if err := c.MoveField(basic.ID, 2, 0); err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(basic.ID); nt.Fields[nt.SortFieldIndex()] != "Front" {
		t.Fatalf("fields %v, sort field %d", nt.Fields, nt.SortFieldIndex())
	}
	if err := c.SetSortField(basic.ID, 2); err != nil {
		t.Fatal(err)
	}
	mustUserErr(t, c.SetSortField(basic.ID, 3), "no longer exists")
	// Renaming keeps it; removing a field before it shifts it; removing it
	// makes the first field sort again.
	if err := c.RenameField(basic.ID, 2, "Answer"); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveField(basic.ID, 0); err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(basic.ID); nt.Fields[nt.SortFieldIndex()] != "Answer" {
		t.Fatalf("fields %v, sort field %d", nt.Fields, nt.SortFieldIndex())
	}
	c = reopen(t, c, store, clk)
	if got := sortField(); got != 1 {
		t.Fatalf("sort field after reopen = %d", got)
	}
	if err := c.RemoveField(basic.ID, 1); err != nil {
		t.Fatal(err)
	}
	if got := sortField(); got != 0 {
		t.Fatalf("sort field after removing it = %d", got)
	}
}

func TestFieldOptionsBackup(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	opts := FieldOptions{Font: "Amiri", Size: 30, RTL: true, Collapsed: true, ExcludeFromSearch: true}
	if err := c.SetFieldOptions(basic.ID, 1, opts); err != nil {
		t.Fatal(err)
	}
	if err := c.SetSortField(basic.ID, 1); err != nil {
		t.Fatal(err)
	}
	snap, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetFieldOptions(basic.ID, 1, FieldOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := c.SetSortField(basic.ID, 0); err != nil {
		t.Fatal(err)
	}
	if err := c.RestoreKeepingMedia(snap, nil, nil); err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(basic.ID); nt.FieldOption(1) != opts || nt.SortFieldIndex() != 1 {
		t.Fatalf("restored options %+v, sort field %d", nt.FieldOptions, nt.SortFieldIndex())
	}
	// Font and size are trimmed and never negative.
	if err := c.SetFieldOptions(basic.ID, 0, FieldOptions{Font: "  Amiri ", Size: -3}); err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(basic.ID); nt.FieldOption(0) != (FieldOptions{Font: "Amiri"}) {
		t.Fatalf("options = %+v", nt.FieldOptions)
	}
}

func TestLastDeckByNoteType(t *testing.T) {
	c, store, clk := testCollection(t)
	basic, cloze := noteType(t, c, "Basic"), noteType(t, c, "Cloze")
	bio, err := c.CreateDeck("Biology")
	if err != nil {
		t.Fatal(err)
	}
	spanish, err := c.CreateDeck("Spanish")
	if err != nil {
		t.Fatal(err)
	}
	add := func(typeID, deck ID, fields ...string) {
		t.Helper()
		if _, err := c.AddNote(NoteInput{TypeID: typeID, DeckID: deck, Fields: fields}); err != nil {
			t.Fatal(err)
		}
	}
	add(cloze.ID, bio, "{{c1::Mitochondria}}", "")
	add(basic.ID, spanish, "hola", "hello")
	want := map[ID]ID{cloze.ID: bio, basic.ID: spanish}
	if st := c.Settings(); !reflect.DeepEqual(st.LastDeckByNoteType, want) || st.LastDeckID != spanish {
		t.Fatalf("settings = %+v", st)
	}
	// The returned settings are a copy.
	c.Settings().LastDeckByNoteType[basic.ID] = bio
	if c.Settings().LastDeckByNoteType[basic.ID] != spanish {
		t.Fatal("settings map changed through a copy")
	}
	// Undoing an add restores the deck remembered before it.
	add(basic.ID, bio, "adiós", "bye")
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if st := c.Settings(); !reflect.DeepEqual(st.LastDeckByNoteType, want) {
		t.Fatalf("after undo = %+v", st.LastDeckByNoteType)
	}
	c = reopen(t, c, store, clk)
	if st := c.Settings(); !reflect.DeepEqual(st.LastDeckByNoteType, want) {
		t.Fatalf("after reopen = %+v", st.LastDeckByNoteType)
	}
	// Deleting a deck or a note type forgets it.
	if err := c.DeleteDeck(bio); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteNoteType(basic.ID); err != nil {
		t.Fatal(err)
	}
	if st := c.Settings(); len(st.LastDeckByNoteType) != 0 {
		t.Fatalf("after deletes = %+v", st.LastDeckByNoteType)
	}
}

func TestDuplicates(t *testing.T) {
	c, store, clk := testCollection(t)
	basic, cloze := noteType(t, c, "Basic").ID, noteType(t, c, "Cloze").ID
	perro := addNote(t, c, "Basic", "<b>perro</b>", "dog")
	addNote(t, c, "Cloze", "{{c1::gato}}", "")
	clozePerro := addNote(t, c, "Cloze", "perro {{c1::x}}", "")
	dups := func(c *Collection, typeID ID, first string, exclude ID) string {
		return fmt.Sprint(c.Duplicates(typeID, first, exclude))
	}
	want := fmt.Sprint([]ID{perro})

	// Duplicate word (spec): HTML, case and surrounding whitespace ignored.
	for _, first := range []string{"perro", "Perro ", " PERRO&nbsp;", "<i>perro</i><br>"} {
		if got := dups(c, basic, first, 0); got != want {
			t.Fatalf("Duplicates(%q) = %s, want %s", first, got, want)
		}
	}
	if got := dups(c, basic, "perros", 0); got != "[]" {
		t.Fatalf("different word: %s", got)
	}
	// Different note type (spec).
	if got := dups(c, cloze, "perro", 0); got != "[]" {
		t.Fatalf("cloze: %s", got)
	}
	if got := dups(c, cloze, "Perro {{c1::x}}", 0); got != fmt.Sprint([]ID{clozePerro}) {
		t.Fatalf("cloze match: %s", got)
	}
	// The note being edited is ignored; empty first fields never match.
	if got := dups(c, basic, "perro", perro); got != "[]" {
		t.Fatalf("excluded: %s", got)
	}
	if got := dups(c, basic, " <br> ", 0); got != "[]" {
		t.Fatalf("empty: %s", got)
	}

	// Add anyway (spec): a duplicate can be added; both are reported.
	perro2 := addNote(t, c, "Basic", "Perro", "dog (again)")
	if got := dups(c, basic, "perro", 0); got != fmt.Sprint([]ID{perro, perro2}) {
		t.Fatalf("after add: %s", got)
	}
	if got := dups(c, basic, "perro", perro2); got != want {
		t.Fatalf("excluding the second: %s", got)
	}

	// Updates move the note to its new first field.
	noErr(t, c.UpdateNote(perro2, []string{"gato", "cat"}, nil))
	if got := dups(c, basic, "perro", 0); got != want {
		t.Fatalf("after update: %s", got)
	}
	if got := dups(c, basic, "gato", 0); got != fmt.Sprint([]ID{perro2}) {
		t.Fatalf("updated note: %s", got)
	}
	// Undo and redo follow.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := dups(c, basic, "perro", 0); got != fmt.Sprint([]ID{perro, perro2}) {
		t.Fatalf("after undo: %s", got)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if got := dups(c, basic, "gato", 0); got != fmt.Sprint([]ID{perro2}) {
		t.Fatalf("after redo: %s", got)
	}
	// Deleting removes the note.
	noErr(t, c.DeleteNotes([]ID{perro}))
	if got := dups(c, basic, "perro", 0); got != "[]" {
		t.Fatalf("after delete: %s", got)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := dups(c, basic, "perro", 0); got != want {
		t.Fatalf("after undoing the delete: %s", got)
	}

	// Imported notes are indexed.
	sum, err := c.Import(&ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "b", ExistingID: basic}},
		Notes:     []ImportNote{{TypeKey: "b", GUID: "imp1", DeckName: "Imported", Fields: []string{"Caballo", "horse"}}},
	}, nil)
	if err != nil || sum.NotesAdded != 1 {
		t.Fatalf("import: %+v %v", sum, err)
	}
	horse := c.Duplicates(basic, "caballo", 0)
	if len(horse) != 1 {
		t.Fatalf("imported note: %v", horse)
	}

	// Renaming the first field keeps the index; removing it reindexes on
	// the new first field.
	noErr(t, c.RenameField(basic, 0, "Word"))
	if got := dups(c, basic, "perro", 0); got != want {
		t.Fatalf("after field rename: %s", got)
	}
	noErr(t, c.RemoveField(basic, 0))
	if got := dups(c, basic, "perro", 0); got != "[]" {
		t.Fatalf("after field removal: %s", got)
	}
	if got := dups(c, basic, "dog", 0); got != want {
		t.Fatalf("new first field: %s", got)
	}

	// The index survives a restart (journal replay and compaction).
	c2 := reopen(t, c, store, clk)
	if got := dups(c2, basic, "dog", 0); got != want {
		t.Fatalf("reopened: %s", got)
	}
	if got := dups(c2, basic, " HORSE", 0); got != fmt.Sprint(horse) {
		t.Fatalf("reopened import: %s", got)
	}
	if got := dups(c2, cloze, "perro {{c1::x}}", 0); got != fmt.Sprint([]ID{clozePerro}) {
		t.Fatalf("reopened cloze: %s", got)
	}

	// Deleting the note type removes its notes.
	noErr(t, c2.DeleteNoteType(cloze))
	if got := dups(c2, cloze, "perro {{c1::x}}", 0); got != "[]" {
		t.Fatalf("after deleting the note type: %s", got)
	}
}

func TestFindDuplicates(t *testing.T) {
	c, _, _ := testCollection(t)
	a, err := c.CreateDeck("A")
	noErr(t, err)
	b, err := c.CreateDeck("B")
	noErr(t, err)
	find := func(field, query string) string {
		t.Helper()
		sets, err := c.FindDuplicates(field, query)
		noErr(t, err)
		var out []string
		for _, s := range sets {
			out = append(out, fmt.Sprintf("%s=%v", s.Content, s.Notes))
		}
		return strings.Join(out, " ")
	}

	// Two duplicates (spec): HTML, case and surrounding whitespace ignored;
	// the oldest note's content is shown.
	perro := addNoteIn(t, c, a, "Basic", "perro", "dog")
	perro2 := addNoteIn(t, c, b, "Basic", "<b>Perro</b> ", "dog")
	gato := addNoteIn(t, c, a, "Basic", "gato", "cat")
	if got, want := find("Front", ""), fmt.Sprintf("perro=%v", []ID{perro, perro2}); got != want {
		t.Fatalf("find = %q, want %q", got, want)
	}
	// Other fields and unknown fields.
	if got := find("Back", ""); got != fmt.Sprintf("dog=%v", []ID{perro, perro2}) {
		t.Fatalf("back = %q", got)
	}
	if got := find("Nope", ""); got != "" {
		t.Fatalf("unknown field = %q", got)
	}

	// Limited by search (spec).
	perroA := addNoteIn(t, c, a, "Basic", "PERRO", "")
	if got := find("Front", "deck:A"); got != fmt.Sprintf("perro=%v", []ID{perro, perroA}) {
		t.Fatalf("deck:A = %q", got)
	}
	if got := find("Front", "deck:B"); got != "" {
		t.Fatalf("deck:B = %q", got)
	}
	if _, err := c.FindDuplicates("Front", "prop:nope>1"); err == nil {
		t.Fatal("invalid search accepted")
	}

	// Empty fields are not duplicates.
	addNoteIn(t, c, a, "Basic (and reversed card)", "x", " <br> ")
	addNoteIn(t, c, a, "Basic (and reversed card)", "y", "")
	if got := find("Back", ""); got != fmt.Sprintf("dog=%v", []ID{perro, perro2}) {
		t.Fatalf("empty fields = %q", got)
	}

	// A field name in two note types compares across them; note types
	// without it are left out. Larger sets come first.
	rev := addNoteIn(t, c, b, "Basic (and reversed card)", "gato", "cat")
	addNoteIn(t, c, a, "Cloze", "{{c1::gato}}", "gato")
	sets, err := c.FindDuplicates("Front", "")
	noErr(t, err)
	if len(sets) != 2 || sets[0].Content != "perro" || len(sets[0].Notes) != 3 ||
		sets[1].Content != "gato" || len(sets[1].Notes) != 2 || sets[1].Notes[0] != gato || sets[1].Notes[1] != rev {
		t.Fatalf("two note types = %+v", sets)
	}

	// Tagging reuses AddTags.
	var all []ID
	for _, s := range sets {
		all = append(all, s.Notes...)
	}
	noErr(t, c.AddTags(all, []string{"duplicate"}))
	ids, err := c.SearchNotes("tag:duplicate", ColCreated, false)
	noErr(t, err)
	if len(ids) != 5 {
		t.Fatalf("tagged = %v", ids)
	}
}

func TestDuplicatesLargeCollection(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	basic := noteType(t, c, "Basic").ID
	c.mu.Lock()
	o := &op{Kind: "bulk"}
	for i := range 50000 {
		n := &Note{ID: c.ids.next(), TypeID: basic, Fields: []string{fmt.Sprintf("<b>Word %d</b>", i%25000), "a"}}
		o.Notes = append(o.Notes, n)
		o.Cards = append(o.Cards, &Card{ID: c.ids.next(), NoteID: n.ID, DeckID: deck, NewPos: int64(i)})
	}
	noErr(t, c.commit(o))
	c.mu.Unlock()

	// One check per keystroke must not be noticeable.
	start := time.Now()
	for i := range 1000 {
		if got := c.Duplicates(basic, fmt.Sprintf("word %d ", i), 0); len(got) != 2 {
			t.Fatalf("word %d: %v", i, got)
		}
	}
	per := time.Since(start) / 1000
	t.Logf("duplicate check over 50,000 notes: %v", per)
	if per > 5*time.Millisecond {
		t.Fatal("too slow")
	}
}

func noErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReservedFieldNames(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	// Reserved name (spec).
	for _, name := range []string{"tags", "Type", "DECK", "card", "FrontSide"} {
		mustUserErr(t, c.AddField(basic.ID, name), "reserved name")
		mustUserErr(t, c.RenameField(basic.ID, 1, name), "reserved name")
	}
	if nt, _ := c.NoteType(basic.ID); strings.Join(nt.Fields, ",") != "Front,Back" {
		t.Fatalf("fields = %q", nt.Fields)
	}
	// Special fields that are not reserved can be field names.
	if err := c.AddField(basic.ID, "Subdeck"); err != nil {
		t.Fatal(err)
	}
}

// An imported note type with a field named like a reserved name still
// loads and renders; the field wins over the special field.
func TestImportedReservedFieldName(t *testing.T) {
	c, store, clk := testCollection(t)
	nt := NoteType{Name: "Imported", Kind: Standard, Fields: []string{"Word", "Type"},
		Templates: []CardTemplate{{Name: "Card 1", QFmt: "{{Word}} ({{Type}})", AFmt: "{{FrontSide}}"}}}
	_, err := c.Import(&ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "t", Type: nt}},
		Notes: []ImportNote{{TypeKey: "t", GUID: "g", DeckName: "Imported", Fields: []string{"laufen", "verb"},
			Cards: []ImportCard{{Ord: 0, DeckName: "Imported"}}}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	c = reopen(t, c, store, clk)
	card := importedCard(t, c, "g")
	if got := c.QuestionText(card); got != "laufen (verb)" {
		t.Fatalf("question = %q", got)
	}
	// Keeping the name when editing the note type is fine.
	typ := noteType(t, c, "Imported")
	if err := c.RenameField(typ.ID, 1, "type"); err != nil {
		t.Fatal(err)
	}
	if err := c.UpdateTemplate(typ.ID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Word}}: {{type}}", AFmt: "{{FrontSide}}"}); err != nil {
		t.Fatal(err)
	}
}

func fields(t *testing.T, c *Collection, id ID) string {
	t.Helper()
	n, err := c.Note(id)
	noErr(t, err)
	return strings.Join(n.Fields, "|")
}

// Fix a spelling (spec): only the notes that contain the text change, and
// the whole replacement is one undo step.
func TestFindReplaceSpelling(t *testing.T) {
	c, store, clk := testCollection(t)
	var ids []ID
	for i := 0; i < 200; i++ {
		front := fmt.Sprint("word ", i)
		if i%40 == 0 || i < 11 {
			front = fmt.Sprint("I Recieve ", i)
		}
		ids = append(ids, addNote(t, c, "Basic", front, "recieve, recieved"))
	}
	fr := FindReplace{Find: "recieve", Replace: "receive", IgnoreCase: true}
	// Every note's Back has it; restrict to the Front to get 15.
	fr.Field = "Front"
	if n, err := c.FindReplacePreview(ids, fr); err != nil || n != 15 {
		t.Fatalf("preview = %d, %v", n, err)
	}
	n, err := c.FindReplace(ids, fr)
	if err != nil || n != 15 {
		t.Fatalf("changed = %d, %v", n, err)
	}
	if got := fields(t, c, ids[0]); got != "I receive 0|recieve, recieved" {
		t.Fatalf("fields = %q", got)
	}
	if got := fields(t, c, ids[1]); !strings.HasPrefix(got, "I receive 1|") {
		t.Fatalf("fields = %q", got)
	}
	if got := fields(t, c, ids[11]); got != "word 11|recieve, recieved" {
		t.Fatalf("unchanged note = %q", got)
	}
	// All fields: every note changes, including both words in Back.
	fr.Field = ""
	if n, err := c.FindReplace(ids, fr); err != nil || n != 200 {
		t.Fatalf("all fields changed = %d, %v", n, err)
	}
	if got := fields(t, c, ids[11]); got != "word 11|receive, received" {
		t.Fatalf("fields = %q", got)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Find and replace" {
		t.Fatalf("undo = %q", undo)
	}
	_, err = c.Undo()
	noErr(t, err)
	if got := fields(t, c, ids[11]); got != "word 11|recieve, recieved" {
		t.Fatalf("after undo = %q", got)
	}
	reopen(t, c, store, clk)
}

func TestFindReplaceCase(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic", "Cat cat CAT", "")
	fr := FindReplace{Find: "cat", Replace: "dog"}
	n, err := c.FindReplace([]ID{id}, fr)
	if err != nil || n != 1 || fields(t, c, id) != "Cat dog CAT|" {
		t.Fatalf("case-sensitive: %d, %v, %q", n, err, fields(t, c, id))
	}
	fr.IgnoreCase = true
	_, err = c.FindReplace([]ID{id}, fr)
	noErr(t, err)
	if got := fields(t, c, id); got != "dog dog dog|" {
		t.Fatalf("ignoring case = %q", got)
	}
	// Literal text: regex characters and $ in the replacement are plain.
	id2 := addNote(t, c, "Basic", "a.b (x)", "")
	_, err = c.FindReplace([]ID{id2}, FindReplace{Find: "(x)", Replace: "$1"})
	noErr(t, err)
	if got := fields(t, c, id2); got != "a.b $1|" {
		t.Fatalf("literal = %q", got)
	}
	// No match: nothing changes and no undo step is added.
	before, _ := c.UndoStatus()
	if n, err := c.FindReplace([]ID{id2}, FindReplace{Find: "zzz", Replace: "y"}); err != nil || n != 0 {
		t.Fatalf("no match = %d, %v", n, err)
	}
	if after, _ := c.UndoStatus(); after != before {
		t.Fatalf("undo %q became %q", before, after)
	}
}

// Regex with group (spec): only the named field changes; fields of the same
// name in other note types change too, notes without it do not.
func TestFindReplaceRegexField(t *testing.T) {
	c, store, clk := testCollection(t)
	id := addNote(t, c, "Basic", "5 kg", "5 kg and 12 KG")
	rev := addNote(t, c, "Basic (and reversed card)", "1 kg", "2 kg")
	cloze := addNote(t, c, "Cloze", "{{c1::3 kg}}", "")
	fr := FindReplace{Find: `(\d+) kg`, Replace: "$1 kilograms", Field: "Back", Regex: true, IgnoreCase: true}
	n, err := c.FindReplace([]ID{id, rev, cloze}, fr)
	if err != nil || n != 2 {
		t.Fatalf("changed = %d, %v", n, err)
	}
	if got := fields(t, c, id); got != "5 kg|5 kilograms and 12 kilograms" {
		t.Fatalf("fields = %q", got)
	}
	if got := fields(t, c, rev); got != "1 kg|2 kilograms" {
		t.Fatalf("reversed = %q", got)
	}
	// Named groups.
	_, err = c.FindReplace([]ID{id}, FindReplace{Find: `(?P<n>\d+) kg`, Replace: "${n}kg", Regex: true})
	noErr(t, err)
	if got := fields(t, c, id); got != "5kg|5 kilograms and 12 kilograms" {
		t.Fatalf("named group = %q", got)
	}
	reopen(t, c, store, clk)
}

// Rename a tag pattern (spec).
func TestFindReplaceTags(t *testing.T) {
	c, store, clk := testCollection(t)
	id, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic").ID, DeckID: defaultDeck(c),
		Fields: []string{"ch0 text", ""}, Tags: []string{"ch01", "CH02", "other"}})
	noErr(t, err)
	n, err := c.FindReplace([]ID{id}, FindReplace{Find: "ch0", Replace: "chapter", Tags: true, IgnoreCase: true})
	if err != nil || n != 1 {
		t.Fatalf("changed = %d, %v", n, err)
	}
	note, _ := c.Note(id)
	if got := strings.Join(note.Tags, " "); got != "chapter1 chapter2 other" {
		t.Fatalf("tags = %q", got)
	}
	if note.Fields[0] != "ch0 text" {
		t.Fatalf("field changed: %q", note.Fields[0])
	}
	// A tag replaced by nothing is dropped; one replaced by two words splits.
	_, err = c.FindReplace([]ID{id}, FindReplace{Find: "other", Replace: "", Tags: true})
	noErr(t, err)
	_, err = c.FindReplace([]ID{id}, FindReplace{Find: "chapter1", Replace: "a b", Tags: true})
	noErr(t, err)
	note, _ = c.Note(id)
	if got := strings.Join(note.Tags, " "); got != "a b chapter2" {
		t.Fatalf("tags = %q", got)
	}
	reopen(t, c, store, clk)
}

// Invalid expression (spec), and an empty find text.
func TestFindReplaceInvalid(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic", "(a)", "")
	before, _ := c.UndoStatus()
	_, err := c.FindReplacePreview([]ID{id}, FindReplace{Find: "(", Regex: true})
	mustUserErr(t, err, "Invalid regular expression")
	_, err = c.FindReplace([]ID{id}, FindReplace{Find: "(", Replace: "x", Regex: true})
	mustUserErr(t, err, "Invalid regular expression")
	_, err = c.FindReplace([]ID{id}, FindReplace{Find: "", Replace: "x"})
	mustUserErr(t, err, "Enter the text to find")
	if got := fields(t, c, id); got != "(a)|" {
		t.Fatalf("fields = %q", got)
	}
	if after, _ := c.UndoStatus(); after != before {
		t.Fatalf("undo %q became %q", before, after)
	}
}

// Emptying a field regenerates cards as for an edit: cards whose question
// becomes empty are kept as empty cards, also when no card is left, and new
// cards are added for templates that now have content.
func TestFindReplaceEmptiesField(t *testing.T) {
	c, store, clk := testCollection(t)
	rev := addNote(t, c, "Basic (and reversed card)", "hola", "hello")
	basic := addNote(t, c, "Basic", "hello", "x")
	n, err := c.FindReplace([]ID{rev, basic}, FindReplace{Find: "hello", Replace: ""})
	if err != nil || n != 2 {
		t.Fatalf("changed = %d, %v", n, err)
	}
	if len(c.CardsOfNote(rev)) != 2 || len(c.CardsOfNote(basic)) != 1 {
		t.Fatal("cards were removed")
	}
	empty := c.EmptyCards()
	if empty.Total != 2 || len(empty.Notes) != 2 || empty.Notes[0].AllEmpty || !empty.Notes[1].AllEmpty {
		t.Fatalf("empty cards = %+v", empty)
	}
	// Filling it again brings the reverse card back into use, with no
	// extra card.
	_, err = c.FindReplace([]ID{rev}, FindReplace{Find: "^$", Replace: "hi", Field: "Back", Regex: true})
	noErr(t, err)
	if got := questions(t, c, rev); fmt.Sprint(got) != "[hola hi]" {
		t.Fatalf("questions = %v", got)
	}
	// A cloze note gains a card for a new number.
	cloze := addNote(t, c, "Cloze", "{{c1::a}} b", "")
	_, err = c.FindReplace([]ID{cloze}, FindReplace{Find: "b", Replace: "{{c2::b}}"})
	noErr(t, err)
	if len(c.CardsOfNote(cloze)) != 2 {
		t.Fatalf("cloze cards = %d", len(c.CardsOfNote(cloze)))
	}
	reopen(t, c, store, clk)
}

func TestAllNoteIDs(t *testing.T) {
	c, _, _ := testCollection(t)
	a := addNote(t, c, "Basic", "a", "")
	b := addNote(t, c, "Basic", "b", "")
	if got := c.AllNoteIDs(); fmt.Sprint(got) != fmt.Sprint([]ID{a, b}) {
		t.Fatalf("ids = %v", got)
	}
}

func addTagged(t *testing.T, c *Collection, front string, tags ...string) ID {
	t.Helper()
	id, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic").ID, DeckID: defaultDeck(c),
		Fields: []string{front, "back"}, Tags: tags})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func tagsOf(t *testing.T, c *Collection, id ID) string {
	t.Helper()
	n, err := c.Note(id)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(n.Tags, ",")
}

func TestRenameTagWithChildren(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addTagged(t, c, "a", "verb::irregular", "lang")
	b := addTagged(t, c, "b", "Verb", "verbose")
	n, err := c.RenameTag("verb", "verbs")
	if err != nil || n != 2 {
		t.Fatalf("renamed %d notes, %v", n, err)
	}
	if got := tagsOf(t, c, a); got != "verbs::irregular,lang" {
		t.Fatalf("a = %q", got)
	}
	// Matching ignores case; verbose is not a child of verb.
	if got := tagsOf(t, c, b); got != "verbs,verbose" {
		t.Fatalf("b = %q", got)
	}
	if name, _ := c.UndoStatus(); usermsg.Plain(name) != "Rename tag" {
		t.Fatalf("undo = %q", name)
	}
	c = reopen(t, c, store, clk)
	if got := strings.Join(c.AllTags(), ","); got != "lang,verbose,verbs,verbs::irregular" {
		t.Fatalf("all tags = %q", got)
	}
}

func TestRenameTagMerge(t *testing.T) {
	c, _, _ := testCollection(t)
	a := addTagged(t, c, "a", "noun", "nouns")
	b := addTagged(t, c, "b", "noun")
	if _, err := c.RenameTag("noun", "nouns"); err != nil {
		t.Fatal(err)
	}
	if got := tagsOf(t, c, a); got != "nouns" {
		t.Fatalf("a = %q", got)
	}
	if got := tagsOf(t, c, b); got != "nouns" {
		t.Fatalf("b = %q", got)
	}
	if got := strings.Join(c.AllTags(), ","); got != "nouns" {
		t.Fatalf("all tags = %q", got)
	}
}

func TestMoveTag(t *testing.T) {
	c, _, _ := testCollection(t)
	a := addTagged(t, c, "a", "irregular::strong")
	if _, err := c.RenameTag("irregular", "verb::irregular"); err != nil {
		t.Fatal(err)
	}
	if got := tagsOf(t, c, a); got != "verb::irregular::strong" {
		t.Fatalf("a = %q", got)
	}
	for _, bad := range []string{"", "::", "two words"} {
		if _, err := c.RenameTag("verb", bad); err == nil {
			t.Fatalf("rename to %q accepted", bad)
		}
	}
}

func TestReparentTags(t *testing.T) {
	c, _, _ := testCollection(t)
	a := addTagged(t, c, "a", "verb::irregular::strong", "noun", "grammar::noun")
	b := addTagged(t, c, "b", "Verb", "adj")
	// verb::irregular moves with verb; noun merges into grammar::noun.
	n, err := c.ReparentTags([]string{"verb", "verb::irregular", "noun"}, "grammar")
	if err != nil || n != 2 {
		t.Fatalf("moved on %d notes, %v", n, err)
	}
	if got := tagsOf(t, c, a); got != "grammar::verb::irregular::strong,grammar::noun" {
		t.Fatalf("a = %q", got)
	}
	// The moved tag is spelled as given.
	if got := tagsOf(t, c, b); got != "grammar::verb,adj" {
		t.Fatalf("b = %q", got)
	}
	if name, _ := c.UndoStatus(); usermsg.Plain(name) != "Move tags" {
		t.Fatalf("undo = %q", name)
	}
	// To the top level.
	if _, err := c.ReparentTags([]string{"grammar::verb::irregular"}, ""); err != nil {
		t.Fatal(err)
	}
	if got := tagsOf(t, c, a); got != "irregular::strong,grammar::noun" {
		t.Fatalf("a = %q", got)
	}
	if _, err := c.ReparentTags([]string{"grammar"}, "grammar::noun"); err == nil {
		t.Fatal("moved a tag into its child")
	}
}

func TestDeleteTag(t *testing.T) {
	c, _, _ := testCollection(t)
	var ids []ID
	for i := range 12 {
		ids = append(ids, addTagged(t, c, fmt.Sprint(i), "old", "keep"))
	}
	child := addTagged(t, c, "child", "OLD::sub")
	other := addTagged(t, c, "other", "older")
	if got := c.TagUsage([]string{"old"}); got != 13 {
		t.Fatalf("usage = %d", got)
	}
	// Notes with several of the tags count once.
	if got := c.TagUsage([]string{"old", "keep", "old::sub"}); got != 13 {
		t.Fatalf("usage of several = %d", got)
	}
	n, err := c.DeleteTags([]string{"old"})
	if err != nil || n != 13 {
		t.Fatalf("deleted from %d notes, %v", n, err)
	}
	for _, id := range ids {
		if got := tagsOf(t, c, id); got != "keep" {
			t.Fatalf("tags = %q", got)
		}
	}
	if got := tagsOf(t, c, child); got != "" {
		t.Fatalf("child = %q", got)
	}
	if got := tagsOf(t, c, other); got != "older" {
		t.Fatalf("other = %q", got)
	}
	if got := len(c.AllNoteIDs()); got != 14 {
		t.Fatalf("notes = %d", got)
	}
	if name, _ := c.UndoStatus(); usermsg.Plain(name) != "Delete tags" {
		t.Fatalf("undo = %q", name)
	}
	if _, err := c.DeleteTags([]string{"keep", "older"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(c.AllTags(), ","); got != "" {
		t.Fatalf("all tags = %q", got)
	}
}

func TestClearUnusedTags(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addTagged(t, c, "a", "draft", "verb")
	if err := c.RemoveTags([]ID{id}, []string{"draft"}); err != nil {
		t.Fatal(err)
	}
	if n := c.ClearUnusedTags(); n != 0 {
		t.Fatalf("cleared %d", n)
	}
	if got := strings.Join(c.AllTags(), ","); got != "verb" {
		t.Fatalf("all tags = %q", got)
	}
}

// threeFieldType adds a note type "Basic+Extra" with fields Front, Back and
// Extra and Basic's card type.
func threeFieldType(t *testing.T, c *Collection) NoteType {
	t.Helper()
	id, err := c.CloneNoteType(noteType(t, c, "Basic").ID, "Basic+Extra")
	noErr(t, err)
	noErr(t, c.AddField(id, "Extra"))
	nt, _ := c.NoteType(id)
	return nt
}

func TestChangeNoteTypePreview(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	rev := noteType(t, c, "Basic (and reversed card)")
	var basicNotes, revNotes []ID
	for i := range 10 {
		basicNotes = append(basicNotes, addNote(t, c, "Basic", fmt.Sprint("front ", i), "back"))
	}
	for i := range 5 {
		revNotes = append(revNotes, addNote(t, c, "Basic (and reversed card)", fmt.Sprint("rev ", i), "back"))
	}

	// Basic to reversed: 10 reverse cards are created, none deleted.
	info, err := c.ChangeNoteTypePreview(basicNotes, rev.ID, []int{0, 1}, []int{0, -1})
	noErr(t, err)
	if fmt.Sprintf("%+v", info) != "{Notes:10 CardsDeleted:0 CardsCreated:10 LostFields:[]}" {
		t.Fatalf("basic to reversed = %+v", info)
	}
	// Confirmation counts (spec): 5 reversed notes to Basic with Card 2
	// mapped to nothing delete 5 cards.
	info, err = c.ChangeNoteTypePreview(revNotes, basic.ID, []int{0, 1}, []int{0})
	noErr(t, err)
	if fmt.Sprintf("%+v", info) != "{Notes:5 CardsDeleted:5 CardsCreated:0 LostFields:[]}" {
		t.Fatalf("reversed to basic = %+v", info)
	}
	// Card 2 to Card 1 instead keeps the reverse cards.
	info, err = c.ChangeNoteTypePreview(revNotes, basic.ID, []int{0, 1}, []int{1})
	noErr(t, err)
	if info.CardsDeleted != 5 || info.CardsCreated != 0 {
		t.Fatalf("card 2 to card 1 = %+v", info)
	}

	// Unmapped field: Extra loses its content, listed by name; an unmapped
	// field without content is not listed.
	extra := threeFieldType(t, c)
	withExtra := addNote(t, c, "Basic+Extra", "a", "b", "c")
	withoutExtra := addNote(t, c, "Basic+Extra", "d", "e", "")
	info, err = c.ChangeNoteTypePreview([]ID{withExtra, withoutExtra}, basic.ID, []int{0, 1}, []int{0})
	noErr(t, err)
	if fmt.Sprintf("%+v", info) != "{Notes:2 CardsDeleted:0 CardsCreated:0 LostFields:[Extra]}" {
		t.Fatalf("unmapped field = %+v", info)
	}
	info, err = c.ChangeNoteTypePreview([]ID{withoutExtra}, basic.ID, []int{0, 1}, []int{0})
	noErr(t, err)
	if len(info.LostFields) != 0 {
		t.Fatalf("empty unmapped field = %+v", info)
	}

	// Notes of different note types are refused.
	_, err = c.ChangeNoteTypePreview([]ID{basicNotes[0], revNotes[0]}, extra.ID, []int{0, 1, -1}, []int{0})
	mustUserErr(t, err, "must use the same note type")
	// One source card type cannot go to two card types.
	_, err = c.ChangeNoteTypePreview(basicNotes, rev.ID, []int{0, 1}, []int{0, 0})
	mustUserErr(t, err, "only go to one card type")
	// Maps of the wrong length are refused.
	_, err = c.ChangeNoteTypePreview(basicNotes, rev.ID, []int{0}, []int{0, -1})
	mustUserErr(t, err, "each field")
	_, err = c.ChangeNoteTypePreview(basicNotes, rev.ID, []int{0, 1}, []int{0})
	mustUserErr(t, err, "each card type")
	// A change that leaves notes without cards is refused.
	_, err = c.ChangeNoteTypePreview(basicNotes, basic.ID, []int{-1, 1}, []int{-1})
	mustUserErr(t, err, "10 notes would have no cards")

	// Standard to cloze: Text without a cloze generates nothing.
	cloze := noteType(t, c, "Cloze")
	_, err = c.ChangeNoteTypePreview(basicNotes[:1], cloze.ID, []int{0, 1}, nil)
	mustUserErr(t, err, "1 note would have no cards")
	withCloze := addNote(t, c, "Basic", "{{c1::a}} {{c2::b}}", "x")
	info, err = c.ChangeNoteTypePreview([]ID{withCloze}, cloze.ID, []int{0, 1}, nil)
	noErr(t, err)
	// Card 1 keeps cloze 1, cloze 2 is created.
	if info.CardsDeleted != 0 || info.CardsCreated != 1 {
		t.Fatalf("basic to cloze = %+v", info)
	}
}

func TestChangeNoteTypeBasicToReversed(t *testing.T) {
	c, store, clk := testCollection(t)
	deck, err := c.CreateDeck("Spanish")
	noErr(t, err)
	var notes, forward []ID
	for i := range 10 {
		id := addNoteIn(t, c, deck, "Basic", fmt.Sprint("hola ", i), "hello")
		notes = append(notes, id)
		forward = append(forward, c.CardsOfNote(id)[0].ID)
	}
	if _, err := c.Answer(forward[0], scheduler.Good, 3*time.Second); err != nil {
		t.Fatal(err)
	}
	setState(t, c, forward[1:], scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 4),
		ScheduledDays: 4, Reps: 2, Memory: &scheduler.Memory{Stability: 4, Difficulty: 5}, LastReview: clk.t})
	answered := c.CardsOfNote(notes[0])[0]

	// Basic to reversed (spec): the forward cards keep their scheduling and
	// a New reverse card is created for each note, in the note's deck.
	rev := noteType(t, c, "Basic (and reversed card)")
	info, err := c.ChangeNoteType(notes, rev.ID, []int{0, 1}, []int{0, -1})
	noErr(t, err)
	if info.Notes != 10 || info.CardsCreated != 10 || info.CardsDeleted != 0 {
		t.Fatalf("info = %+v", info)
	}
	for i, id := range notes {
		n, _ := c.Note(id)
		if n.TypeID != rev.ID || n.Fields[0] != fmt.Sprint("hola ", i) || n.Fields[1] != "hello" {
			t.Fatalf("note = %+v", n)
		}
		cards := c.CardsOfNote(id)
		if len(cards) != 2 || cards[0].ID != forward[i] || cards[1].State != scheduler.New || cards[1].DeckID != deck {
			t.Fatalf("cards = %+v", cards)
		}
		if i > 0 && (cards[0].State != scheduler.Review || cards[0].ScheduledDays != 4) {
			t.Fatalf("forward card lost its scheduling: %+v", cards[0])
		}
	}
	if got := c.CardsOfNote(notes[0])[0]; !reflect.DeepEqual(got.CardState, answered.CardState) {
		t.Fatalf("answered card = %+v, want %+v", got.CardState, answered.CardState)
	}
	// The review history stays with the card.
	if got := c.ReviewsOfCard(forward[0]); len(got) != 1 || got[0].Rating != scheduler.Good {
		t.Fatalf("reviews = %+v", got)
	}
	if got := questions(t, c, notes[3]); fmt.Sprint(got) != "[hola 3 hello]" {
		t.Fatalf("questions = %v", got)
	}
	if name, _ := c.UndoStatus(); usermsg.Plain(name) != "Change note type" {
		t.Fatalf("undo = %q", name)
	}

	// Undo brings back the Basic notes with their one card.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if n, _ := c.Note(notes[0]); n.TypeID != noteType(t, c, "Basic").ID {
		t.Fatalf("undo left note = %+v", n)
	}
	if cards := c.CardsOfNote(notes[0]); len(cards) != 1 || cards[0].ID != forward[0] {
		t.Fatalf("undo left cards = %+v", cards)
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if cards := c.CardsOfNote(notes[0]); len(cards) != 2 || cards[0].ID != forward[0] {
		t.Fatalf("redo left cards = %+v", cards)
	}
	reopen(t, c, store, clk)
}

func TestChangeNoteTypeUnmappedField(t *testing.T) {
	c, store, clk := testCollection(t)
	threeFieldType(t, c)
	id := addNote(t, c, "Basic+Extra", "a", "b", "c")
	card := c.CardsOfNote(id)[0]

	// Unmapped field (spec): Front and Back keep a and b, c is gone.
	basic := noteType(t, c, "Basic")
	info, err := c.ChangeNoteType([]ID{id}, basic.ID, []int{0, 1}, []int{0})
	noErr(t, err)
	if strings.Join(info.LostFields, ",") != "Extra" {
		t.Fatalf("info = %+v", info)
	}
	if got := fields(t, c, id); got != "a|b" {
		t.Fatalf("fields = %q", got)
	}
	if cards := c.CardsOfNote(id); len(cards) != 1 || cards[0].ID != card.ID {
		t.Fatalf("cards = %+v", cards)
	}
	reopen(t, c, store, clk)
}

func TestChangeNoteTypeSwapsFields(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic", "front", "back")
	// One source field can fill several target fields.
	_, err := c.ChangeNoteType([]ID{id}, noteType(t, c, "Basic").ID, []int{1, 1}, []int{0})
	noErr(t, err)
	if got := fields(t, c, id); got != "back|back" {
		t.Fatalf("fields = %q", got)
	}
}

func TestChangeNoteTypeClozeNumbersKept(t *testing.T) {
	c, store, clk := testCollection(t)
	clozeType := noteType(t, c, "Cloze")
	otherID, err := c.CloneNoteType(clozeType.ID, "My cloze")
	noErr(t, err)
	id := addNote(t, c, "Cloze", "{{c1::Paris}} is in {{c2::France}}", "extra")
	before := c.CardsOfNote(id)
	setState(t, c, []ID{before[1].ID}, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 9),
		ScheduledDays: 9, Reps: 3, Memory: &scheduler.Memory{Stability: 9, Difficulty: 5}, LastReview: clk.t})
	before = c.CardsOfNote(id)

	// Cloze numbers kept (spec): each cloze number's card keeps its
	// scheduling.
	info, err := c.ChangeNoteType([]ID{id}, otherID, []int{0, 1}, nil)
	noErr(t, err)
	if info.CardsDeleted != 0 || info.CardsCreated != 0 {
		t.Fatalf("info = %+v", info)
	}
	if n, _ := c.Note(id); n.TypeID != otherID {
		t.Fatalf("note = %+v", n)
	}
	if after := c.CardsOfNote(id); !reflect.DeepEqual(after, before) {
		t.Fatalf("cards = %+v, want %+v", after, before)
	}

	// Cloze to standard: cloze 1 becomes Card 1 with its scheduling, cloze 2
	// has no card type and is deleted.
	info, err = c.ChangeNoteType([]ID{id}, noteType(t, c, "Basic").ID, []int{0, 1}, nil)
	noErr(t, err)
	if info.CardsDeleted != 1 || info.CardsCreated != 0 {
		t.Fatalf("info = %+v", info)
	}
	if after := c.CardsOfNote(id); len(after) != 1 || after[0].ID != before[0].ID {
		t.Fatalf("cards = %+v", after)
	}
	reopen(t, c, store, clk)
}

func TestChangeNoteTypeConfirmationCounts(t *testing.T) {
	c, store, clk := testCollection(t)
	var notes []ID
	for i := range 5 {
		notes = append(notes, addNote(t, c, "Basic (and reversed card)", fmt.Sprint("q", i), "a"))
	}
	basic := noteType(t, c, "Basic")
	// Confirmation counts (spec): Card 2 mapped to nothing deletes 5 cards.
	preview, err := c.ChangeNoteTypePreview(notes, basic.ID, []int{0, 1}, []int{0})
	noErr(t, err)
	info, err := c.ChangeNoteType(notes, basic.ID, []int{0, 1}, []int{0})
	noErr(t, err)
	if !reflect.DeepEqual(info, preview) || info.Notes != 5 || info.CardsDeleted != 5 {
		t.Fatalf("info = %+v, preview = %+v", info, preview)
	}
	for _, id := range notes {
		if cards := c.CardsOfNote(id); len(cards) != 1 || cards[0].Ord != 0 {
			t.Fatalf("cards = %+v", cards)
		}
	}
	// A refused change changes nothing.
	_, err = c.ChangeNoteType(append(notes, addNote(t, c, "Cloze", "{{c1::x}}", "")), basic.ID, []int{0, 1}, []int{0})
	mustUserErr(t, err, "same note type")
	reopen(t, c, store, clk)
}

// Card type options survive the journal and a snapshot.
func TestCardTypeOptionsRoundTrip(t *testing.T) {
	c, store, clk := testCollection(t)
	nt := noteType(t, c, "Basic (and reversed card)")
	deck, err := c.CreateDeck("Spanish::Production")
	noErr(t, err)
	tm := nt.Templates[1]
	tm.DeckOverride, tm.BrowserQFmt, tm.BrowserAFmt = deck, "{{Back}}", "{{Front}}"
	noErr(t, c.UpdateTemplate(nt.ID, 1, tm))
	c = reopen(t, c, store, clk)
	c.mu.Lock()
	err = c.compact()
	c.mu.Unlock()
	noErr(t, err)
	c = reopen(t, c, store, clk)
	if got := noteType(t, c, nt.Name).Templates[1]; got != tm {
		t.Fatalf("template = %+v, want %+v", got, tm)
	}
	// A filtered or missing deck cannot be the override.
	tm.DeckOverride = 12345
	mustUserErr(t, c.UpdateTemplate(nt.ID, 1, tm), "no longer exists")
}

// Swap two card types (spec): cards keep card type, deck and scheduling.
func TestMoveTemplate(t *testing.T) {
	c, store, clk := testCollection(t)
	nt := noteType(t, c, "Basic (and reversed card)")
	noErr(t, c.UpdateTemplate(nt.ID, 0, CardTemplate{Name: "Recognition", QFmt: "{{Front}}", AFmt: "{{Back}}"}))
	noErr(t, c.UpdateTemplate(nt.ID, 1, CardTemplate{Name: "Production", QFmt: "{{Back}}", AFmt: "{{Front}}"}))
	id := addNote(t, c, nt.Name, "hola", "hello")
	other, err := c.CreateDeck("Other")
	noErr(t, err)
	before := c.CardsOfNote(id)
	noErr(t, c.MoveCards([]ID{before[1].ID}, other))
	_, err = c.Answer(before[0].ID, scheduler.Good, time.Second)
	noErr(t, err)
	before = c.CardsOfNote(id)

	noErr(t, c.MoveTemplate(nt.ID, 1, 0))
	if got := noteType(t, c, nt.Name).Templates; got[0].Name != "Production" || got[1].Name != "Recognition" {
		t.Fatalf("templates = %+v", got)
	}
	after := c.CardsOfNote(id)
	if len(after) != 2 || after[0].ID != before[1].ID || after[1].ID != before[0].ID {
		t.Fatalf("cards = %+v", after)
	}
	for i, want := range []Card{before[1], before[0]} {
		got := after[i]
		if got.DeckID != want.DeckID || got.State != want.State || !got.Due.Equal(want.Due) || got.Reps != want.Reps ||
			got.Ord != i {
			t.Fatalf("card %d = %+v, want %+v with ord %d", i, got, want, i)
		}
	}
	if got := questions(t, c, id); fmt.Sprint(got) != "[hello hola]" {
		t.Fatalf("questions = %v", got)
	}

	// Undo puts the order back.
	_, err = c.Undo()
	noErr(t, err)
	if got := questions(t, c, id); fmt.Sprint(got) != "[hola hello]" {
		t.Fatalf("questions after undo = %v", got)
	}
	_, err = c.Redo()
	noErr(t, err)
	c = reopen(t, c, store, clk)
	if got := questions(t, c, id); fmt.Sprint(got) != "[hello hola]" {
		t.Fatalf("questions after reopen = %v", got)
	}
	mustUserErr(t, c.MoveTemplate(nt.ID, 0, 2), "no longer exists")
	mustUserErr(t, c.MoveTemplate(noteType(t, c, "Cloze").ID, 0, 1), "exactly one")
	reopen(t, c, store, clk)
}

// Moving a card type in a type with three keeps every card with its type.
func TestMoveTemplateThree(t *testing.T) {
	c, _, _ := testCollection(t)
	nt := noteType(t, c, "Basic (and reversed card)")
	noErr(t, c.AddTemplate(nt.ID, CardTemplate{Name: "Both", QFmt: "{{Front}} {{Back}}", AFmt: "x"}))
	id := addNote(t, c, nt.Name, "a", "b")
	noErr(t, c.MoveTemplate(nt.ID, 0, 2))
	if got := questions(t, c, id); fmt.Sprint(got) != "[b a b a]" {
		t.Fatalf("questions = %v", got)
	}
	noErr(t, c.MoveTemplate(nt.ID, 2, 0))
	if got := questions(t, c, id); fmt.Sprint(got) != "[a b a b]" {
		t.Fatalf("questions = %v", got)
	}
}

// Production cards to their own deck, and override deck deleted (spec).
func TestDeckOverride(t *testing.T) {
	c, store, clk := testCollection(t)
	if _, err := c.AddNoteType(StockOptionalReversed, "Optional"); err != nil {
		t.Fatal(err)
	}
	nt := noteType(t, c, "Optional")
	spanish, err := c.CreateDeck("Spanish")
	noErr(t, err)
	production, err := c.CreateDeck("Spanish::Production")
	noErr(t, err)
	tm := nt.Templates[1]
	tm.DeckOverride = production
	noErr(t, c.UpdateTemplate(nt.ID, 1, tm))

	add := func(fields ...string) ID {
		t.Helper()
		id, err := c.AddNote(NoteInput{TypeID: nt.ID, DeckID: spanish, Fields: fields})
		noErr(t, err)
		return id
	}
	decks := func(id ID) []ID {
		var out []ID
		for _, card := range c.CardsOfNote(id) {
			out = append(out, card.DeckID)
		}
		return out
	}
	both := add("hola", "hello", "y")
	if got := decks(both); !slices.Equal(got, []ID{spanish, production}) {
		t.Fatalf("decks = %v, want [%d %d]", got, spanish, production)
	}
	// Editing a note so that the card is generated uses the override too.
	one := add("adiós", "goodbye", "")
	noErr(t, c.UpdateNote(one, []string{"adiós", "goodbye", "y"}, nil))
	if got := decks(one); !slices.Equal(got, []ID{spanish, production}) {
		t.Fatalf("decks after edit = %v", got)
	}
	// Existing cards are not moved when the override changes.
	other, err := c.CreateDeck("Other")
	noErr(t, err)
	tm.DeckOverride = other
	noErr(t, c.UpdateTemplate(nt.ID, 1, tm))
	if got := decks(both); !slices.Equal(got, []ID{spanish, production}) {
		t.Fatalf("decks after changing the override = %v", got)
	}
	c = reopen(t, c, store, clk)

	noErr(t, c.DeleteDeck(other))
	id := add("gato", "cat", "y")
	if got := decks(id); !slices.Equal(got, []ID{spanish, spanish}) {
		t.Fatalf("decks with deleted override = %v", got)
	}
	reopen(t, c, store, clk)
}

func TestAddNotes(t *testing.T) {
	c, store, clk := testCollection(t)
	basic, cloze := noteType(t, c, "Basic"), noteType(t, c, "Cloze")
	spanish, err := c.CreateDeck("Spanish")
	if err != nil {
		t.Fatal(err)
	}
	addNote(t, c, "Basic", "hola", "hello") // remembers Basic and Default
	before := c.Settings()
	in := []NoteInput{
		{TypeID: basic.ID, DeckID: spanish, Fields: []string{"adiós", "goodbye"}},
		{TypeID: basic.ID, DeckID: spanish, Fields: []string{" <b>HOLA</b> ", "hi"}}, // duplicate of the existing note
		{TypeID: cloze.ID, DeckID: spanish, Fields: []string{"{{c1::Paris}} is the capital of {{c2::France}}"}},
		{TypeID: basic.ID, DeckID: spanish, Fields: []string{"Adiós", "bye"}}, // duplicate within the batch
	}
	got, err := c.AddNotes(in, AddNotesOptions{KeepDefaults: true, SkipDuplicates: true})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].NoteID == 0 || len(got[0].Cards) != 1 || got[1].NoteID != 0 || got[1].DuplicateOf == 0 ||
		got[2].NoteID == 0 || len(got[2].Cards) != 2 || got[3].NoteID != 0 || got[3].DuplicateOf != got[0].NoteID {
		t.Fatalf("results = %+v", got)
	}
	for _, id := range got[2].Cards {
		if card, _ := c.Card(id); card.DeckID != spanish {
			t.Fatalf("card %d in deck %d", id, card.DeckID)
		}
	}
	// Added notes keep the editor's choices (agent-access spec).
	if s := c.Settings(); !reflect.DeepEqual(s, before) {
		t.Fatalf("settings changed: %+v, was %+v", s, before)
	}
	// One undo step, named "Add 2 notes", reverts the whole batch.
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Add 2 notes" {
		t.Fatalf("undo = %q", usermsg.Plain(undo))
	}
	reopen(t, c, store, clk)
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if r.NoteID == 0 {
			continue
		}
		if _, err := c.Note(r.NoteID); err == nil {
			t.Fatalf("note %d still there after undo", r.NoteID)
		}
	}
	if n := len(c.AllNoteIDs()); n != 1 {
		t.Fatalf("%d notes after undo", n)
	}
}

func TestAddNotesRemembersLastWithoutKeepDefaults(t *testing.T) {
	c, _, _ := testCollection(t)
	basic, cloze := noteType(t, c, "Basic"), noteType(t, c, "Cloze")
	spanish, _ := c.CreateDeck("Spanish")
	_, err := c.AddNotes([]NoteInput{
		{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"a", "b"}},
		{TypeID: cloze.ID, DeckID: spanish, Fields: []string{"{{c1::x}}"}},
	}, AddNotesOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if s := c.Settings(); s.LastNoteTypeID != cloze.ID || s.LastDeckID != spanish {
		t.Fatalf("settings = %+v", s)
	}
}

func TestAddNotesKeepsDuplicatesWhenNotSkipping(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	addNote(t, c, "Basic", "hola", "hello")
	got, err := c.AddNotes([]NoteInput{
		{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"hola", "hi"}},
		{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"hola", "hey"}},
	}, AddNotesOptions{})
	if err != nil || got[0].NoteID == 0 || got[1].NoteID == 0 {
		t.Fatalf("results = %+v, %v", got, err)
	}
}

func TestAddNotesInvalidNoteAddsNothing(t *testing.T) {
	c, _, _ := testCollection(t)
	basic, cloze := noteType(t, c, "Basic"), noteType(t, c, "Cloze")
	undoBefore, _ := c.UndoStatus()
	_, err := c.AddNotes([]NoteInput{
		{TypeID: basic.ID, DeckID: defaultDeck(c), Fields: []string{"a", "b"}},
		{TypeID: cloze.ID, DeckID: defaultDeck(c), Fields: []string{"no deletions"}},
	}, AddNotesOptions{SkipDuplicates: true})
	var ae *AddNotesError
	if !errors.As(err, &ae) || ae.Index != 1 || !strings.Contains(usermsg.Plain(err.Error()), "cloze") {
		t.Fatalf("err = %v", err)
	}
	if n := len(c.AllNoteIDs()); n != 0 {
		t.Fatalf("%d notes added", n)
	}
	if undo, _ := c.UndoStatus(); undo != undoBefore {
		t.Fatalf("undo = %q", undo)
	}
}
