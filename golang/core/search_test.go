package core

import (
	"cardflip/golang/internal/usermsg"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cardflip/golang/render"
	"cardflip/golang/scheduler"
	"cardflip/golang/search"
	"cardflip/golang/storage"
)

func TestSearchCombined(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	irregular, _ := c.CreateDeck("Spanish::Irregular")
	french, _ := c.CreateDeck("French")
	basic := noteType(t, c, "Basic").ID
	add := func(deck ID, front string, tags ...string) ID {
		id, err := c.AddNote(NoteInput{TypeID: basic, DeckID: deck, Fields: []string{front, "x"}, Tags: tags})
		if err != nil {
			t.Fatal(err)
		}
		return c.CardsOfNote(id)[0].ID
	}
	comer := add(spanish, "<b>comer</b>", "verb")
	ir := add(irregular, "ir", "verb")
	ser := add(spanish, "ser", "verb")
	perro := add(spanish, "perro", "noun")
	manger := add(french, "manger", "verb")
	c.SetSuspended([]ID{ser}, true)

	// Combined search (spec).
	ids, err := c.Search("deck:Spanish tag:verb -is:suspended", ColCreated, false)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(ids) != fmt.Sprint([]ID{comer, ir}) {
		t.Fatalf("ids = %v, want %v", ids, []ID{comer, ir})
	}
	// Flag search (spec).
	c.SetFlag([]ID{ir}, 1)
	if ids, _ = c.Search("flag:1", ColCreated, false); fmt.Sprint(ids) != fmt.Sprint([]ID{ir}) {
		t.Fatalf("flag:1 = %v", ids)
	}
	if ids, _ = c.Search("deck:Spanish -flag:0 -is:suspended", ColCreated, false); fmt.Sprint(ids) != fmt.Sprint([]ID{ir}) {
		t.Fatalf("-flag:0 = %v", ids)
	}
	// Free text ignores HTML and case.
	ids, _ = c.Search("COMER", ColCreated, false)
	if len(ids) != 1 || ids[0] != comer {
		t.Fatalf("text search = %v", ids)
	}
	// Search by preset (spec): the cards of every deck using it.
	languages, err := c.AddPreset(spanish, "Languages")
	if err != nil {
		t.Fatal(err)
	}
	c.SetDeckPreset(french, languages)
	ids, _ = c.Search("preset:languages", ColCreated, false)
	if want := []ID{comer, ser, perro, manger}; fmt.Sprint(ids) != fmt.Sprint(want) {
		t.Fatalf("preset:languages = %v, want %v", ids, want)
	}
	if ids, _ = c.Search(`preset:default`, ColCreated, false); fmt.Sprint(ids) != fmt.Sprint([]ID{ir}) {
		t.Fatalf("preset:default = %v", ids)
	}
	// Either deck (spec), and a negated group.
	if ids, _ = c.Search("(deck:Spanish or deck:French) -is:suspended tag:verb", ColCreated, false); fmt.Sprint(ids) != fmt.Sprint([]ID{comer, ir, manger}) {
		t.Fatalf("or = %v", ids)
	}
	if ids, _ = c.Search("-(tag:verb or deck:French)", ColCreated, false); fmt.Sprint(ids) != fmt.Sprint([]ID{perro}) {
		t.Fatalf("negated group = %v", ids)
	}
	// Invalid qualifier (spec): an error naming the term.
	_, err = c.Search("verb is:bogus", ColCreated, false)
	var se *search.Error
	if !errors.As(err, &se) || se.Term != "is:bogus" {
		t.Fatalf("err = %v", err)
	}
	// Unclosed group (spec).
	_, err = c.Search("(deck:Spanish or deck:French", ColCreated, false)
	if !errors.As(err, &se) || usermsg.Plain(se.Error()) != "a parenthesis is not closed (character 1)" {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchTextAndFields(t *testing.T) {
	c, _, _ := testCollection(t)
	dog := addNote(t, c, "Basic", "<b>dog</b>", "chien")
	aDog := addNote(t, c, "Basic", "a dog", "")
	digits := addNote(t, c, "Basic", "123", "über")
	salt := addNote(t, c, "Basic", "salt &amp; pepper", "café")
	cloze := addNote(t, c, "Cloze", "{{c1::capital}} of France", "")
	check := func(query string, want ...ID) {
		t.Helper()
		ids, err := c.Search(query, ColCreated, false)
		if err != nil {
			t.Fatalf("%q: %v", query, err)
		}
		var notes []ID
		for _, id := range ids {
			if n := c.mustCard(t, id).NoteID; len(notes) == 0 || notes[len(notes)-1] != n {
				notes = append(notes, n)
			}
		}
		if fmt.Sprint(notes) != fmt.Sprint(want) {
			t.Fatalf("%q = %v, want %v", query, notes, want)
		}
	}
	check("front:dog", dog)
	check("Front:*DOG*", dog, aDog)
	check("back:", aDog)
	check("back:_*", dog, digits, salt)
	check(`front:re:^\d+$`, digits)
	check("re:<b>", dog)
	check("d_g", dog, aDog)
	check("w:dog", dog, aDog)
	check(`"salt & pepper"`, salt)
	check("nc:uber", digits)
	check(`"sc:capital of France"`, cloze)
	check("text:*france", cloze)
	check("cafe")
	// Setting on (spec): plain searches ignore accents.
	st := c.Settings()
	st.IgnoreAccents = true
	if err := c.UpdateSettings(st); err != nil {
		t.Fatal(err)
	}
	check("cafe", salt)
	// The per-field text follows edits.
	c.UpdateNote(dog, []string{"cat", "chat"}, nil)
	check("front:dog")
	check("front:cat", dog)
	_, err := c.Search("re:[", ColCreated, false)
	var se *search.Error
	if !errors.As(err, &se) || se.Term != "re:[" {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchSortAndRows(t *testing.T) {
	c, _, clk := testCollection(t)
	deck := defaultDeck(c)
	b := addCards(t, c, deck, 1)[0]
	c.UpdateNote(c.CardsOfNote(c.mustCard(t, b).NoteID)[0].NoteID, []string{"banana", "x"}, nil)
	a := c.CardsOfNote(addNote(t, c, "Basic", "apple", "x"))[0].ID
	cc := c.CardsOfNote(addNote(t, c, "Basic", "Cherry", "x"))[0].ID
	setState(t, c, []ID{cc}, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 1), ScheduledDays: 1, Memory: &scheduler.Memory{Stability: 1, Difficulty: 5}})
	setState(t, c, []ID{b}, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 5), ScheduledDays: 5, Memory: &scheduler.Memory{Stability: 5, Difficulty: 5}})

	check := func(order Column, desc bool, want ...ID) {
		t.Helper()
		ids, err := c.Search("", order, desc)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(ids) != fmt.Sprint(want) {
			t.Fatalf("%s desc=%v: %v, want %v", order, desc, ids, want)
		}
	}
	check(ColCreated, false, b, a, cc)
	check(ColCreated, true, cc, a, b)
	check(ColQuestion, false, a, b, cc)
	check(ColDue, false, cc, b, a)

	rows := c.Rows([]ID{cc, 999, a})
	if len(rows) != 2 || rows[0].Question != "Cherry" || rows[0].State != scheduler.Review || rows[0].Deck != "Default" ||
		rows[0].NoteType != "Basic" || rows[0].CardName != "Card 1" || rows[1].Question != "apple" || !rows[1].Due.IsZero() {
		t.Fatalf("rows = %+v", rows)
	}
	// Edit from browser (spec): the row shows the new content.
	c.UpdateNote(rows[1].NoteID, []string{"apricot", "x"}, nil)
	if got := c.Rows([]ID{a})[0].Question; got != "apricot" {
		t.Fatalf("row after edit = %q", got)
	}
	ids, _ := c.Search("apricot", ColCreated, false)
	if len(ids) != 1 {
		t.Fatalf("search after edit = %v", ids)
	}
}

func (c *Collection) mustCard(t *testing.T, id ID) Card {
	t.Helper()
	card, err := c.Card(id)
	if err != nil {
		t.Fatal(err)
	}
	return card
}

func largeCollection(t testing.TB, n int) *Collection {
	store := testStore()
	c, _, err := Open(store.Open(), Options{Now: func() time.Time { return time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local) }})
	if err != nil {
		t.Fatal(err)
	}
	var basic ID
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Basic" {
			basic = nt.ID
		}
	}
	deck := c.Decks()[0].ID
	c.mu.Lock()
	o := &op{Kind: "bulk"}
	for i := 0; i < n; i++ {
		note := &Note{ID: c.ids.next(), TypeID: basic, Fields: []string{
			fmt.Sprintf("<div>Question number %d about <b>topic %d</b></div>", i, i%97),
			fmt.Sprintf("The answer %d is somewhat longer than the question, with words", i)},
			Tags: []string{"tag" + fmt.Sprint(i%13)}}
		o.Notes = append(o.Notes, note)
		o.Cards = append(o.Cards, &Card{ID: c.ids.next(), NoteID: note.ID, DeckID: deck, NewPos: int64(i)})
	}
	if err := c.commit(o); err != nil {
		t.Fatal(err)
	}
	c.mu.Unlock()
	return c
}

// Search over 50,000 cards stays under 50 ms (design D10).
func TestSearch50kUnder50ms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	for _, q := range []string{"topic 42", "tag:tag3 -is:suspended", "deck:default answer", "nothing-matches-this", "added:1", "rated:7 or edited:3"} {
		start := time.Now()
		ids, err := c.Search(q, ColCreated, false)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%q: %d cards in %v", q, len(ids), elapsed)
		if elapsed > 50*time.Millisecond && !raceDetector {
			t.Fatalf("%q took %v", q, elapsed)
		}
	}
}

// Field, wildcard, accent and regex searches over 50,000 cards show their
// first rows within a second.
func TestFieldAndRegexSearch50k(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	queries := []string{"front:*x*", "front:*u*", `"front:*topic 42*"`, `"re:number \d+7 "`, `"back:re:^the answer \d+ is"`,
		"w:topic", "nc:answer", `"sc:question number"`, "*42*about"}
	for _, ignoreAccents := range []bool{false, true} {
		st := c.Settings()
		st.IgnoreAccents = ignoreAccents
		if err := c.UpdateSettings(st); err != nil {
			t.Fatal(err)
		}
		for _, q := range queries {
			start := time.Now()
			ids, err := c.Search(q, ColCreated, false)
			if err != nil {
				t.Fatal(err)
			}
			c.Rows(ids[:min(len(ids), 100)])
			elapsed := time.Since(start)
			t.Logf("%q (ignore accents %v): %d cards in %v", q, ignoreAccents, len(ids), elapsed)
			if elapsed > time.Second && !raceDetector {
				t.Fatalf("%q took %v", q, elapsed)
			}
		}
	}
}

func BenchmarkSearch50k(b *testing.B) {
	c := largeCollection(b, 50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Search("topic 42 tag:tag3", ColCreated, false)
	}
}

// Filling the per-field search text of 50,000 notes stays under 200 ms.
func TestSearchFieldsFill50k(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	start := time.Now()
	fillSearchFields(c)
	elapsed := time.Since(start)
	t.Logf("filled %d notes in %v", len(c.notes), elapsed)
	if elapsed > 200*time.Millisecond && !raceDetector {
		t.Fatalf("filling took %v", elapsed)
	}
}

func BenchmarkSearchFieldsFill50k(b *testing.B) {
	c := largeCollection(b, 50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fillSearchFields(c)
	}
}

func fillSearchFields(c *Collection) {
	for id, n := range c.notes {
		c.noteFields[id] = searchFields(n.Fields)
	}
}

// searchIDs runs a query that must be valid.
func searchIDs(t *testing.T, c *Collection, query string) []ID {
	t.Helper()
	ids, err := c.Search(query, ColCreated, false)
	if err != nil {
		t.Fatalf("%q: %v", query, err)
	}
	return ids
}

// collectionAt opens a fresh collection with the clock at start, so that
// IDs, which never decrease, can be creation times from then on.
func collectionAt(t *testing.T, start time.Time) (*Collection, *storage.MemoryStore, *clock) {
	t.Helper()
	store := storage.NewMemoryStore()
	clk := &clock{t: start}
	c, _, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	return c, store, clk
}

func TestSearchAddedAndEdited(t *testing.T) {
	c, _, clk := collectionAt(t, time.Date(2026, 10, 1, 12, 0, 0, 0, time.Local)) // rollover 04:00
	at := func(month time.Month, day, hour, min int) {
		clk.t = time.Date(2026, month, day, hour, min, 0, 0, time.Local)
	}
	card := func(front string) ID { return c.CardsOfNote(addNote(t, c, "Basic", front, "x"))[0].ID }

	at(10, 4, 2, 0) // yesterday 02:00: the study day of 3 October
	early := card("early")
	at(10, 5, 2, 0) // 02:00 today: the study day of 4 October
	late := card("late")
	at(10, 5, 10, 0)
	var today []ID
	for i := range 12 {
		today = append(today, card(fmt.Sprint("today ", i)))
	}

	// Added today (spec), and before rollover (spec): neither the card
	// added yesterday at 02:00 nor the one added today at 02:00 counts.
	if got := searchIDs(t, c, "added:1"); fmt.Sprint(got) != fmt.Sprint(today) {
		t.Fatalf("added:1 = %v, want %v", got, today)
	}
	if got := searchIDs(t, c, "added:2"); fmt.Sprint(got) != fmt.Sprint(append([]ID{late}, today...)) {
		t.Fatalf("added:2 = %v", got)
	}
	if got := searchIDs(t, c, "added:3"); len(got) != 14 || got[0] != early {
		t.Fatalf("added:3 = %v", got)
	}
	// The study day ends at the next rollover.
	at(10, 6, 3, 59)
	if got := searchIDs(t, c, "added:1"); len(got) != 12 {
		t.Fatalf("added:1 at 03:59 = %d cards", len(got))
	}
	at(10, 6, 4, 0)
	if got := searchIDs(t, c, "added:1"); len(got) != 0 {
		t.Fatalf("added:1 after rollover = %v", got)
	}
	// A later rollover hour moves the boundary.
	st := c.Settings()
	st.RolloverHour = 5
	if err := c.UpdateSettings(st); err != nil {
		t.Fatal(err)
	}
	if got := searchIDs(t, c, "added:1"); len(got) != 12 {
		t.Fatalf("added:1 at 04:00 with rollover 05:00 = %d cards", len(got))
	}

	// edited: follows the note's modification time.
	at(10, 6, 12, 0)
	n, _ := c.Note(c.mustCard(t, early).NoteID)
	if err := c.UpdateNote(n.ID, []string{"early!", "x"}, nil); err != nil {
		t.Fatal(err)
	}
	if got := searchIDs(t, c, "edited:1"); fmt.Sprint(got) != fmt.Sprint([]ID{early}) {
		t.Fatalf("edited:1 = %v", got)
	}
	if got := searchIDs(t, c, "edited:2"); len(got) != 13 {
		t.Fatalf("edited:2 = %d cards", len(got))
	}
}

func TestSearchReviewEvents(t *testing.T) {
	c, store, clk := collectionAt(t, time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local))
	at := func(month time.Month, day, hour int) { clk.t = time.Date(2026, month, day, hour, 0, 0, 0, time.Local) }
	card := func(front string) ID { return c.CardsOfNote(addNote(t, c, "Basic", front, "x"))[0].ID }
	a, b, cc, d := card("a"), card("b"), card("c"), card("d")
	answer := func(id ID, r scheduler.Rating) ID {
		e, err := c.Answer(id, r, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		return e.ID
	}
	at(9, 25, 12)
	answer(b, scheduler.Good)
	at(10, 1, 12)
	answer(a, scheduler.Again)
	at(10, 4, 12)
	answer(a, scheduler.Good)
	at(10, 5, 10)
	answer(cc, scheduler.Again)
	at(10, 5, 12)

	check := func(query string, want ...ID) {
		t.Helper()
		if got := searchIDs(t, c, query); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("%q = %v, want %v", query, got, want)
		}
	}
	check("rated:1", cc)
	check("rated:2", a, cc)
	// Failed this week (spec).
	check("rated:7:1", a, cc)
	check("rated:7:3", a)
	check("rated:30", a, b, cc)
	check("-rated:30", d)
	check("rated:365 or is:new", a, b, cc, d)
	check("introduced:1", cc)
	check("introduced:4", cc)
	check("introduced:5", a, cc)
	check("prop:rated=0", cc)
	check("prop:rated=-1", a)
	check("prop:rated<-1", b)
	check("prop:rated>=-1", a, cc)
	check("prop:rated!=0", a, b)
	check("resched:30")

	// The index follows undo and redo.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	check("rated:1")
	check("introduced:30", a, b)
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	check("rated:1", cc)

	// Imported history: a manual entry two days ago, an answer a year ago
	// and an Again a minute ago, added while the index exists.
	basic := noteType(t, c, "Basic").ID
	now := clk.now()
	if _, err := c.Import(&ImportBatch{
		NoteTypes: []ImportNoteType{{Key: "basic", ExistingID: basic}},
		Notes: []ImportNote{{TypeKey: "basic", DeckName: "Imported", Fields: []string{"gato", "cat"},
			Cards: []ImportCard{{Ord: 0, State: scheduler.CardState{State: scheduler.Learning, Due: now},
				History: []ImportReview{
					{Time: now.AddDate(-1, 0, 0), Rating: scheduler.Good, Source: "test:1"},
					{Time: now.AddDate(0, 0, -2), Rating: 0, Source: "test:2"},
					{Time: now.Add(-time.Minute), Rating: scheduler.Again, StateBefore: scheduler.Review, Source: "test:3"},
				}}}}},
	}, nil); err != nil {
		t.Fatal(err)
	}
	gato := searchIDs(t, c, "gato")[0]
	check("rated:1:1", cc, gato)
	check("introduced:30", a, b, cc)
	check("resched:3", gato)
	check("resched:2")
	check("prop:resched=-2", gato)
	check("rated:3 deck:imported", gato)

	// After a restart months later, the index is read from the stored log,
	// beyond the months loaded for the study day.
	at(12, 20, 12)
	c = reopen(t, c, store, clk)
	check("rated:365", a, b, cc, gato)
	check("rated:30")
	check("introduced:365", a, b, cc)
	check("prop:rated=-76", cc, gato)
	check("prop:resched<-70", gato)
}

func TestSearchEventErrors(t *testing.T) {
	c, _, _ := testCollection(t)
	for query, msg := range map[string]string{
		// Out of range (spec).
		"rated:400":   "the number of days must be at most 365",
		"rated:0":     "needs a number of days",
		"rated:7:5":   "the rating must be 1 (Again)",
		"added:x":     "needs a number of days",
		"added:1:1":   "needs a number of days",
		"prop:rated":  "needs a comparison",
		"prop:foo=1":  "unknown property",
		"prop:rated=": "needs a comparison",
	} {
		_, err := c.Search(query, ColCreated, false)
		var se *search.Error
		if !errors.As(err, &se) || se.Term != query || !strings.Contains(se.Msg, msg) {
			t.Errorf("%q: err = %v, want %q", query, err, msg)
		}
	}
	if _, err := c.Search("rated:365", ColCreated, false); err != nil {
		t.Fatal(err)
	}
}

func TestSearchCardProperties(t *testing.T) {
	c, _, clk := testCollection(t)
	card := func(front string) ID { return c.CardsOfNote(addNote(t, c, "Basic", front, "x"))[0].ID }
	reviewed, learning, fresh := card("reviewed"), card("learning"), card("new")
	if _, err := c.Answer(reviewed, scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Answer(learning, scheduler.Again, time.Second); err != nil {
		t.Fatal(err)
	}
	r := c.mustCard(t, reviewed)
	if r.State != scheduler.Review || r.ScheduledDays < 2 {
		t.Fatalf("reviewed = %+v", r.CardState)
	}
	check := func(query string, want ...ID) {
		t.Helper()
		if got := searchIDs(t, c, query); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("%q = %v, want %v", query, got, want)
		}
	}
	// Only Review cards have an interval; the learning card has a memory
	// state but no interval.
	check(fmt.Sprintf("prop:ivl=%d", r.ScheduledDays), reviewed)
	check("prop:ivl>=1", reviewed)
	check("prop:reps=1", reviewed, learning)
	check("prop:reps=0", fresh)
	check("prop:lapses=0", reviewed, learning, fresh)
	check(fmt.Sprintf("prop:pos=%d", c.mustCard(t, fresh).NewPos), fresh)
	check("prop:s>0", reviewed, learning)
	check("prop:d>=0 prop:d<=1", reviewed, learning)
	// Just answered: everything is remembered.
	check("prop:r>0.99", reviewed, learning)
	check("prop:due=0", learning)
	check(fmt.Sprintf("prop:due=%d", r.ScheduledDays), reviewed)
	// Overdue (spec) and at risk (spec), counted at search time: two
	// weeks past the due date.
	clk.t = r.Due.Add(14 * 24 * time.Hour)
	check("prop:due<0", reviewed, learning)
	check("prop:due=-14", reviewed)
	check("prop:r<0.9", reviewed, learning)
	if _, err := c.Search("prop:ease<2.5", ColCreated, false); err == nil || !strings.Contains(err.Error(), "prop:d") {
		t.Fatalf("prop:ease err = %v", err)
	}
}

func TestSearchBurialAndRelearning(t *testing.T) {
	c, _, clk := testCollection(t)
	burySiblings(t, c, defaultDeck(c))
	pair := func(front string) (ID, ID) {
		cards := c.CardsOfNote(addNote(t, c, StockReversed, front, "x"))
		return cards[0].ID, cards[1].ID
	}
	answer := func(id ID, r scheduler.Rating) {
		t.Helper()
		if _, err := c.Answer(id, r, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	check := func(query string, want ...ID) {
		t.Helper()
		if got := searchIDs(t, c, query); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("%q = %v, want %v", query, got, want)
		}
	}
	front, back := pair("chat")
	otherFront, suspendedBack := pair("chien")
	lone := c.CardsOfNote(addNote(t, c, "Basic", "lone", "x"))[0].ID
	if err := c.SetSuspended([]ID{suspendedBack}, true); err != nil {
		t.Fatal(err)
	}
	answer(front, scheduler.Good)
	answer(otherFront, scheduler.Good)
	if err := c.BuryCards([]ID{lone}); err != nil {
		t.Fatal(err)
	}
	// The reverse card waits for tomorrow; the suspended one is not buried.
	check("is:buried", back, lone)
	check("is:buried-sibling", back)
	check("is:buried-manually", lone)
	check("-is:buried is:new", suspendedBack)
	// The next study day: nothing is buried.
	clk.t = clk.t.Add(24 * time.Hour)
	check("is:buried")

	// Lapsed cards (spec): a review card answered Again is relearning.
	lapsed := c.CardsOfNote(addNote(t, c, "Basic", "lapsed", "x"))[0].ID
	answer(lapsed, scheduler.Easy)
	clk.t = c.mustCard(t, lapsed).Due.Add(time.Hour)
	answer(lapsed, scheduler.Again)
	if st := c.mustCard(t, lapsed).State; st != scheduler.Relearning {
		t.Fatalf("lapsed state = %v", st)
	}
	check("is:learn is:review", lapsed)
	check("is:learn", front, otherFront, lapsed)
	check("is:review -is:learn")
}

func TestSearchCardTypesDecksAndIDs(t *testing.T) {
	c, _, _ := testCollection(t)
	french, _ := c.CreateDeck("French")
	verbs, _ := c.CreateDeck("French::Verbs")
	extra, _ := c.CreateDeck("French Extra")
	reversed := addNoteIn(t, c, french, StockReversed, "chat", "cat")
	forward, reverse := c.CardsOfNote(reversed)[0].ID, c.CardsOfNote(reversed)[1].ID
	basic := addNoteIn(t, c, verbs, "Basic", "aller", "to go")
	verb := c.CardsOfNote(basic)[0].ID
	clozeNote := addNoteIn(t, c, extra, "Cloze", "{{c1::Paris}} is in {{c2::France}}", "")
	cloze1, cloze2 := c.CardsOfNote(clozeNote)[0].ID, c.CardsOfNote(clozeNote)[1].ID
	if err := c.UpdateNote(basic, []string{"aller", "to go"}, []string{"verb"}); err != nil {
		t.Fatal(err)
	}
	check := func(query string, want ...ID) {
		t.Helper()
		if got := searchIDs(t, c, query); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("%q = %v, want %v", query, got, want)
		}
	}
	// Reverse cards (spec).
	check("card:2", reverse, cloze2)
	check("card:1", forward, verb, cloze1)
	check(`"card:card 2"`, reverse)
	check("card:cloze", cloze1, cloze2)
	check("tag:none", forward, reverse, cloze1, cloze2)
	// Without subdecks (spec).
	check("deck:French -deck:French::*", forward, reverse)
	// By card IDs (spec), and note IDs.
	check(fmt.Sprintf("cid:%d,%d", forward, cloze2), forward, cloze2)
	check(fmt.Sprintf("nid:%d", clozeNote), cloze1, cloze2)
	// deck:current follows the deck most recently studied, renamed or not.
	check("deck:current")
	if err := c.SetCurrentDeck(french); err != nil {
		t.Fatal(err)
	}
	check("deck:current", forward, reverse, verb)
	if err := c.RenameDeck(french, "Français"); err != nil {
		t.Fatal(err)
	}
	check("deck:current", forward, reverse, verb)
	if err := c.SetCurrentDeck(extra); err != nil {
		t.Fatal(err)
	}
	check("deck:current", cloze1, cloze2)
	if err := c.SetCurrentDeck(ID(12345)); err == nil {
		t.Fatal("unknown deck accepted")
	}
}

func TestSearchNotes(t *testing.T) {
	c, _, clk := testCollection(t)
	other, _ := c.CreateDeck("Other")
	reversed := addNote(t, c, "Basic (and reversed card)", "gato", "cat")
	apple := addNote(t, c, "Basic", "Apple", "x")
	banana := addNote(t, c, "Basic (and reversed card)", "banana", "x")

	// Reversed note once (spec): one row showing 2 cards.
	ids, err := c.SearchNotes("", ColCreated, false)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(ids) != fmt.Sprint([]ID{reversed, apple, banana}) {
		t.Fatalf("notes = %v", ids)
	}
	rows := c.NoteRows([]ID{reversed, 999})
	if len(rows) != 1 || rows[0].Cards != 2 || rows[0].SortField != "gato" || rows[0].NoteType != "Basic (and reversed card)" ||
		rows[0].Deck != "Default" || rows[0].MultipleDecks || rows[0].CardID != c.CardsOfNote(reversed)[0].ID ||
		rows[0].DueState != scheduler.New || rows[0].Suspended {
		t.Fatalf("rows = %+v", rows)
	}

	// Match through one card (spec): one suspended card lists the note.
	gato := c.CardsOfNote(reversed)
	c.SetSuspended([]ID{gato[1].ID}, true)
	if ids, _ = c.SearchNotes("is:suspended", ColCreated, false); fmt.Sprint(ids) != fmt.Sprint([]ID{reversed}) {
		t.Fatalf("is:suspended = %v", ids)
	}
	if row := c.NoteRows([]ID{reversed})[0]; row.Suspended {
		t.Fatal("note suspended with one card unsuspended")
	}

	// Cards in two decks show as multiple.
	c.MoveCards([]ID{gato[1].ID}, other)
	if row := c.NoteRows([]ID{reversed})[0]; row.Deck != "" || !row.MultipleDecks {
		t.Fatalf("deck = %q multiple=%v", row.Deck, row.MultipleDecks)
	}

	// Sorting: first field ignoring case, and earliest due of all cards.
	check := func(order Column, desc bool, want ...ID) {
		t.Helper()
		ids, err := c.SearchNotes("", order, desc)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(ids) != fmt.Sprint(want) {
			t.Fatalf("%s desc=%v: %v, want %v", order, desc, ids, want)
		}
	}
	check(ColCreated, true, banana, apple, reversed)
	check(ColSortField, false, apple, banana, reversed)
	bananaCards := c.CardsOfNote(banana)
	setState(t, c, []ID{bananaCards[1].ID}, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 3), ScheduledDays: 3})
	setState(t, c, []ID{c.CardsOfNote(apple)[0].ID}, scheduler.CardState{State: scheduler.Review, Due: clk.t.AddDate(0, 0, 5), ScheduledDays: 5})
	check(ColDue, false, banana, apple, reversed)
	check(ColDue, true, reversed, apple, banana)
	row := c.NoteRows([]ID{banana})[0]
	if row.DueState != scheduler.Review || !row.Due.Equal(clk.t.AddDate(0, 0, 3)) {
		t.Fatalf("banana due = %v %v", row.DueState, row.Due)
	}

	if cards := c.CardsOfNotes([]ID{reversed, 999, apple}); len(cards) != 3 {
		t.Fatalf("cards of notes = %v", cards)
	}
	if _, err := c.SearchNotes("", ColQuestion, false); err == nil {
		t.Fatal("notes sorted by question")
	}
	if _, err := c.SearchNotes("is:bogus", ColCreated, false); err == nil {
		t.Fatal("invalid term accepted")
	}
}

// Notes mode over 50,000 cards stays as fast as card search.
func TestSearchNotes50k(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	for _, order := range NoteColumns {
		start := time.Now()
		ids, err := c.SearchNotes("", order, false)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s: %d notes in %v", order, len(ids), elapsed)
		if len(ids) != 50000 {
			t.Fatalf("%d notes", len(ids))
		}
	}
	start := time.Now()
	ids, _ := c.SearchNotes("", ColCreated, false)
	c.NoteRows(ids[:100])
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond && !raceDetector {
		t.Fatalf("search and first page took %v", elapsed)
	}
}

// Simplified question (spec): the browser shows the browser templates,
// studying the normal ones.
func TestBrowserAppearanceTemplates(t *testing.T) {
	c, _, _ := testCollection(t)
	id, err := c.AddNoteType(StockBasic, "Vocab")
	noErr(t, err)
	noErr(t, c.RenameField(id, 0, "Word"))
	noErr(t, c.RenameField(id, 1, "Meaning"))
	noErr(t, c.AddField(id, "Hint"))
	tm := CardTemplate{Name: "Card 1", QFmt: "<div class=hint>{{Hint}}</div>{{Word}}", AFmt: "{{FrontSide}}<hr id=answer>{{Meaning}}"}
	noErr(t, c.UpdateTemplate(id, 0, tm))
	card := c.CardsOfNote(addNote(t, c, "Vocab", "gato", "cat", "an animal"))[0].ID
	row := func() Row {
		t.Helper()
		return c.Rows([]ID{card}, ColQuestion, ColAnswer)[0]
	}
	if r := row(); r.Question != "an animal gato" || r.Answer != "cat" {
		t.Fatalf("row with normal templates = %q / %q", r.Question, r.Answer)
	}
	tm.BrowserQFmt, tm.BrowserAFmt = "{{Word}}", "{{Meaning}} (noun)"
	noErr(t, c.UpdateTemplate(id, 0, tm))
	if r := row(); r.Question != "gato" || r.Answer != "cat (noun)" {
		t.Fatalf("row with browser templates = %q / %q", r.Question, r.Answer)
	}
	ids, err := c.Search("", ColQuestion, false)
	noErr(t, err)
	if len(ids) != 1 {
		t.Fatalf("search = %v", ids)
	}
	r, err := c.RenderCard(card, render.Env{})
	noErr(t, err)
	if !strings.Contains(r.Question, "an animal") {
		t.Fatalf("study question = %q", r.Question)
	}
}

func TestSearchExcludedField(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	if err := c.AddField(basic.ID, "Notes"); err != nil {
		t.Fatal(err)
	}
	gone := addNote(t, c, "Basic", "go", "went", "irregular")
	be := addNote(t, c, "Basic", "be", "irregular verb", "irregular")
	cloze := addNote(t, c, "Cloze", "irregular {{c1::verbs}}", "")
	check := func(query string, want ...ID) {
		t.Helper()
		ids, err := c.SearchNotes(query, ColCreated, false)
		if err != nil {
			t.Fatalf("%q: %v", query, err)
		}
		if fmt.Sprint(ids) != fmt.Sprint(want) {
			t.Fatalf("%q = %v, want %v", query, ids, want)
		}
	}
	check("irregular", gone, be, cloze)
	if err := c.SetFieldOptions(basic.ID, 2, FieldOptions{ExcludeFromSearch: true}); err != nil {
		t.Fatal(err)
	}
	// Notes field excluded (spec): the card is listed only when another
	// field contains the word; other note types are unaffected.
	check("irregular", be, cloze)
	check("notes:irregular", gone, be)
	check("irregular -notes:irregular", cloze)
	// The option follows the field when it moves.
	if err := c.MoveField(basic.ID, 2, 0); err != nil {
		t.Fatal(err)
	}
	check("irregular", be, cloze)
	check("go", gone)
}

func TestSortByConfiguredSortField(t *testing.T) {
	c, _, _ := testCollection(t)
	basic := noteType(t, c, "Basic")
	water := addNote(t, c, "Basic", "Agua", "<b>water</b>")
	dog := addNote(t, c, "Basic", "Perro", "Dog")
	cat := addNote(t, c, "Basic", "gato", "cat")
	cards := func(want ...ID) []ID {
		var out []ID
		for _, n := range want {
			out = append(out, c.CardsOfNote(n)[0].ID)
		}
		return out
	}
	check := func(want ...ID) {
		t.Helper()
		ids, err := c.Search("", ColSortField, false)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(ids) != fmt.Sprint(cards(want...)) {
			t.Fatalf("cards = %v, want notes %v", ids, want)
		}
		notes, err := c.SearchNotes("", ColSortField, false)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(notes) != fmt.Sprint(want) {
			t.Fatalf("notes = %v, want %v", notes, want)
		}
	}
	check(water, cat, dog)
	// Sort by second field (spec): the cards are in order of Back, with
	// HTML stripped and ignoring case, and the column shows it.
	if err := c.SetSortField(basic.ID, 1); err != nil {
		t.Fatal(err)
	}
	check(cat, dog, water)
	if row := c.Rows(cards(water), ColSortField)[0]; row.SortField != "water" {
		t.Fatalf("sort field column = %q", row.SortField)
	}
}
