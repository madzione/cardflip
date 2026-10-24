package core

import (
	"cardflip/golang/internal/usermsg"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// corrupt applies a raw op that bypasses validation, as a bug or a damaged
// import would.
func corrupt(t *testing.T, c *Collection, o *op) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if o.Kind == "" {
		o.Kind = "corrupt"
	}
	if err := c.commit(o); err != nil {
		t.Fatal(err)
	}
}

func rawCard(c *Collection, id ID) *Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := *c.cards[id]
	return &cp
}

func rawNote(c *Collection, id ID) *Note {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := cloneNote(c.notes[id])
	return &cp
}

func checkDB(t *testing.T, c *Collection) CheckReport {
	t.Helper()
	r, err := c.CheckDatabase(func(string, int, int) {})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// problems renders a report's problems as "kind=count" strings.
func problems(r CheckReport) string {
	var out []string
	for _, p := range r.Problems {
		out = append(out, fmt.Sprintf("%s=%d", p.Kind, p.Count))
	}
	return strings.Join(out, " ")
}

func message(t *testing.T, r CheckReport, kind string) string {
	t.Helper()
	for _, p := range r.Problems {
		if p.Kind == kind {
			return usermsg.Plain(p.Message)
		}
	}
	t.Fatalf("no %s problem in %s", kind, problems(r))
	return ""
}

// collectionState returns the collection's objects and settings, without
// the sync metadata, which every change updates, undo included.
func collectionState(c *Collection) *state {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return withoutSyncMeta(normalizeState(c.state()))
}

func firstCard(t *testing.T, c *Collection, noteID ID) Card {
	t.Helper()
	cards := c.CardsOfNote(noteID)
	if len(cards) == 0 {
		t.Fatalf("note %d has no cards", noteID)
	}
	return cards[0]
}

// answerUntilReview answers a new card Good until it is a review card.
func answerUntilReview(t *testing.T, c *Collection, clk *clock, id ID) {
	t.Helper()
	for range 10 {
		if c.mustCard(t, id).State == scheduler.Review {
			return
		}
		if _, err := c.Answer(id, scheduler.Good, time.Second); err != nil {
			t.Fatal(err)
		}
		clk.advance(24 * time.Hour)
	}
	t.Fatalf("card %d did not become a review card", id)
}

// Healthy collection (spec): no problems, nothing changes. Review log
// entries of deleted cards are counted, not reported as problems.
func TestCheckDatabaseHealthy(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic", "hola", "hello")
	addNote(t, c, "Basic (and reversed card)", "gato", "cat")
	addNote(t, c, "Cloze", "{{c1::Paris}} is in {{c2::France}}", "")
	gone := addNote(t, c, "Basic", "adiós", "bye")
	if err := c.AddTags([]ID{a}, []string{"spanish"}); err != nil {
		t.Fatal(err)
	}
	answerUntilReview(t, c, clk, firstCard(t, c, a).ID)
	if _, err := c.Answer(firstCard(t, c, gone).ID, scheduler.Again, time.Second); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteNotes([]ID{gone}); err != nil {
		t.Fatal(err)
	}
	before, rev := collectionState(c), c.Revision()
	undo, _ := c.UndoStatus()
	r := checkDB(t, c)
	if len(r.Problems) != 0 {
		t.Fatalf("problems = %s", problems(r))
	}
	if r.OrphanedReviews != 1 {
		t.Fatalf("orphaned reviews = %d, want 1", r.OrphanedReviews)
	}
	if got := collectionState(c); !reflect.DeepEqual(got, before) {
		t.Fatalf("collection changed:\n%s", jsonDiff(got, before))
	}
	if c.Revision() != rev {
		t.Fatal("revision changed")
	}
	if u, _ := c.UndoStatus(); u != undo {
		t.Fatalf("undo = %q, want %q", u, undo)
	}
	reopen(t, c, store, clk)
}

// Orphaned card (spec): a card whose note no longer exists is deleted.
func TestCheckDatabaseCardWithoutNote(t *testing.T) {
	c, store, clk := testCollection(t)
	id := addNote(t, c, "Basic", "hola", "hello")
	card := firstCard(t, c, id)
	corrupt(t, c, &op{DelNotes: []ID{id}})
	r := checkDB(t, c)
	if problems(r) != "cardsWithoutNote=1" {
		t.Fatalf("problems = %s", problems(r))
	}
	if got := message(t, r, ProblemCardsWithoutNote); got != "1 card without a note was removed." {
		t.Fatalf("message = %q", got)
	}
	if _, err := c.Card(card.ID); err == nil {
		t.Fatal("card still exists")
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseNoteWithoutNoteType(t *testing.T) {
	c, store, clk := testCollection(t)
	nt := noteType(t, c, "Basic (and reversed card)")
	id := addNote(t, c, nt.Name, "gato", "cat")
	kept := addNote(t, c, "Basic", "hola", "hello")
	corrupt(t, c, &op{DelNoteTypes: []ID{nt.ID}})
	r := checkDB(t, c)
	if problems(r) != "notesWithoutType=1" {
		t.Fatalf("problems = %s", problems(r))
	}
	if got := message(t, r, ProblemNotesWithoutType); !strings.Contains(got, "1 note whose note type no longer exists was removed, with 2 cards") {
		t.Fatalf("message = %q", got)
	}
	if _, err := c.Note(id); err == nil || len(c.CardsOfNote(id)) != 0 {
		t.Fatal("note or its cards still exist")
	}
	if len(c.CardsOfNote(kept)) != 1 {
		t.Fatal("other note's card removed")
	}
	if got := fmt.Sprint(deletionLog(t, c)); got != "[:gato|cat:]" {
		t.Fatalf("deletion log = %s", got)
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseFieldCount(t *testing.T) {
	c, store, clk := testCollection(t)
	short := addNote(t, c, "Basic", "hola", "hello")
	long := addNote(t, c, "Basic", "gato", "cat")
	empty := addNote(t, c, "Basic", "perro", "dog")
	n1, n2, n3 := rawNote(c, short), rawNote(c, long), rawNote(c, empty)
	n1.Fields = []string{"hola"}
	n2.Fields = []string{"gato", "cat", "a <b>cat</b>", ""}
	n3.Fields = []string{"perro", "dog", ""}
	corrupt(t, c, &op{Notes: []*Note{n1, n2, n3}})
	clk.advance(time.Hour)
	r := checkDB(t, c)
	if problems(r) != "fieldCount=3" {
		t.Fatalf("problems = %s", problems(r))
	}
	if got := message(t, r, ProblemFieldCount); got != "3 notes with the wrong number of fields were fixed. The removed field content is in the deletion log." {
		t.Fatalf("message = %q", got)
	}
	for id, want := range map[ID]string{short: "hola|", long: "gato|cat", empty: "perro|dog"} {
		n, err := c.Note(id)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(n.Fields, "|"); got != want || !n.Modified.Equal(clk.now()) {
			t.Fatalf("note %d fields = %q (mod %v), want %q", id, got, n.Modified, want)
		}
	}
	// Only the note whose trimmed fields had content is logged, as it was.
	if got := fmt.Sprint(deletionLog(t, c)); got != "[Basic:gato|cat|a <b>cat</b>|:]" {
		t.Fatalf("deletion log = %s", got)
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseInvalidCards(t *testing.T) {
	c, store, clk := testCollection(t)
	basic := addNote(t, c, "Basic", "hola", "hello")
	cloze := addNote(t, c, "Cloze", "{{c1::Paris}}", "")
	gone := addNote(t, c, "Cloze", "{{c1::Rome}} and {{c2::Italy}}", "")
	b, z := rawCard(c, firstCard(t, c, basic).ID), rawCard(c, firstCard(t, c, cloze).ID)
	b.ID, b.Ord = c.ids.next(), 3
	z.ID, z.Ord = c.ids.next(), -1
	corrupt(t, c, &op{Cards: []*Card{b, z}})
	// A cloze number gone from the text is an empty card, not a problem.
	n := rawNote(c, gone)
	n.Fields[0] = "{{c1::Rome}} and Italy"
	corrupt(t, c, &op{Notes: []*Note{n}})
	r := checkDB(t, c)
	if problems(r) != "invalidCards=2" {
		t.Fatalf("problems = %s", problems(r))
	}
	if len(c.CardsOfNote(basic)) != 1 || len(c.CardsOfNote(cloze)) != 1 || len(c.CardsOfNote(gone)) != 2 {
		t.Fatalf("cards = %d %d %d", len(c.CardsOfNote(basic)), len(c.CardsOfNote(cloze)), len(c.CardsOfNote(gone)))
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseMissingCards(t *testing.T) {
	c, store, clk := testCollection(t)
	deck, err := c.CreateDeck("Spanish")
	if err != nil {
		t.Fatal(err)
	}
	id, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic (and reversed card)").ID, DeckID: deck, Fields: []string{"gato", "cat"}})
	if err != nil {
		t.Fatal(err)
	}
	lone := addNote(t, c, "Basic", "hola", "hello")
	cards := c.CardsOfNote(id)
	corrupt(t, c, &op{DelCards: []ID{cards[1].ID, firstCard(t, c, lone).ID}})
	r := checkDB(t, c)
	if problems(r) != "missingCards=2" {
		t.Fatalf("problems = %s", problems(r))
	}
	got := c.CardsOfNote(id)
	if len(got) != 2 || got[1].Ord != 1 || got[1].DeckID != deck || got[1].State != scheduler.New || got[1].NewPos == 0 {
		t.Fatalf("cards = %+v", got)
	}
	if got := c.CardsOfNote(lone); len(got) != 1 || got[0].DeckID != defaultDeck(c) {
		t.Fatalf("lone note cards = %+v", got)
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseMissingDeck(t *testing.T) {
	c, store, clk := testCollection(t)
	id := addNote(t, c, "Basic", "hola", "hello")
	card := rawCard(c, firstCard(t, c, id).ID)
	card.DeckID = 424242
	corrupt(t, c, &op{Cards: []*Card{card}})
	r := checkDB(t, c)
	if got := message(t, r, ProblemMissingDeck); got != "1 card in a missing deck was moved to Default." || problems(r) != "missingDeck=1" {
		t.Fatalf("problems = %s, message = %q", problems(r), got)
	}
	if got := c.mustCard(t, card.ID); got.DeckID != defaultDeck(c) {
		t.Fatalf("deck = %d", got.DeckID)
	}

	// Without a Default deck, one is created.
	if err := c.RenameDeck(defaultDeck(c), "Spanish"); err != nil {
		t.Fatal(err)
	}
	card = rawCard(c, card.ID)
	card.DeckID = 424242
	corrupt(t, c, &op{Cards: []*Card{card}})
	checkDB(t, c)
	if deckNames(c) != "Default, Spanish" {
		t.Fatalf("decks = %s", deckNames(c))
	}
	if d, _ := c.Deck(c.mustCard(t, card.ID).DeckID); d.Name != "Default" {
		t.Fatalf("card deck = %q", d.Name)
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseScheduling(t *testing.T) {
	c, store, clk := testCollection(t)
	var ids []ID
	for i := range 6 {
		ids = append(ids, firstCard(t, c, addNote(t, c, "Basic", fmt.Sprint("q", i), "a")).ID)
	}
	answerUntilReview(t, c, clk, ids[0])
	answerUntilReview(t, c, clk, ids[1])
	healthy := c.mustCard(t, ids[0])
	now := clk.now()

	// ids[0]: review card without a memory state, with history.
	noMem := rawCard(c, ids[0])
	noMem.Memory = nil
	// ids[1]: review card with an invalid memory state and a negative interval.
	badMem := rawCard(c, ids[1])
	badMem.Memory = &scheduler.Memory{Stability: 4, Difficulty: 0}
	badMem.ScheduledDays = -3
	// ids[2]: review card without memory or history, due date out of range.
	noHist := rawCard(c, ids[2])
	noHist.CardState = scheduler.CardState{State: scheduler.Review, Due: now.AddDate(500, 0, 0), ScheduledDays: 20, Reps: 3}
	// ids[3]: learning card without a due date and with negative counts.
	learning := rawCard(c, ids[3])
	learning.CardState = scheduler.CardState{State: scheduler.Learning, Step: -1, Reps: -2, Lapses: -1}
	// ids[4]: unknown state.
	unknown := rawCard(c, ids[4])
	unknown.CardState = scheduler.CardState{State: 9, Due: now}
	corrupt(t, c, &op{Cards: []*Card{noMem, badMem, noHist, learning, unknown}})

	r := checkDB(t, c)
	if problems(r) != "scheduling=5" {
		t.Fatalf("problems = %s", problems(r))
	}
	if got := message(t, r, ProblemScheduling); got != "5 cards with invalid scheduling were repaired. Memory states of 2 cards were recomputed from the review history." {
		t.Fatalf("message = %q", got)
	}
	// Recomputed from the same history, the memory state is the one the
	// answers gave.
	if got := c.mustCard(t, ids[0]); got.Memory == nil || math.Abs(got.Memory.Stability-healthy.Memory.Stability) > 1e-9 ||
		math.Abs(got.Memory.Difficulty-healthy.Memory.Difficulty) > 1e-9 {
		t.Fatalf("memory = %+v, want %+v", got.Memory, healthy.Memory)
	}
	if got := c.mustCard(t, ids[1]); got.Memory == nil || !validMemory(*got.Memory) || got.ScheduledDays != 1 {
		t.Fatalf("bad memory card = %+v", got.CardState)
	}
	if got := c.mustCard(t, ids[2]); got.Memory == nil || got.Memory.Stability != 20 || got.Memory.Difficulty != 5 || !got.Due.Equal(now) {
		t.Fatalf("no history card = %+v", got.CardState)
	}
	if got := c.mustCard(t, ids[3]); got.Step != 0 || got.Reps != 0 || got.Lapses != 0 || !got.Due.Equal(now) || got.Memory != nil {
		t.Fatalf("learning card = %+v", got.CardState)
	}
	if got := c.mustCard(t, ids[4]); got.State != scheduler.New || !got.Due.IsZero() || got.NewPos == 0 {
		t.Fatalf("unknown state card = %+v", got)
	}
	if got := c.mustCard(t, ids[5]); got.State != scheduler.New {
		t.Fatalf("untouched card = %+v", got.CardState)
	}
	reopen(t, c, store, clk)
}

func TestCheckDatabaseTags(t *testing.T) {
	c, store, clk := testCollection(t)
	id := addNote(t, c, "Basic", "hola", "hello")
	ok := addNote(t, c, "Basic", "gato", "cat")
	if err := c.AddTags([]ID{ok}, []string{"animal"}); err != nil {
		t.Fatal(err)
	}
	n := rawNote(c, id)
	n.Tags = []string{"verb", "Verb", "two words", ""}
	corrupt(t, c, &op{Notes: []*Note{n}})
	r := checkDB(t, c)
	if problems(r) != "tags=1" {
		t.Fatalf("problems = %s", problems(r))
	}
	if got, _ := c.Note(id); strings.Join(got.Tags, ",") != "verb,two,words" {
		t.Fatalf("tags = %q", got.Tags)
	}
	reopen(t, c, store, clk)
}

// The repair is one change: undo restores the corrupted collection, and the
// report lists every kind in a fixed order.
func TestCheckDatabaseOneChange(t *testing.T) {
	c, store, clk := testCollection(t)
	a := addNote(t, c, "Basic", "hola", "hello")
	b := addNote(t, c, "Basic", "gato", "cat")
	card := rawCard(c, firstCard(t, c, b).ID)
	card.DeckID = 424242
	n := rawNote(c, b)
	n.Tags = []string{"x", "X"}
	corrupt(t, c, &op{DelNotes: []ID{a}, Cards: []*Card{card}, Notes: []*Note{n}})
	corrupted := collectionState(c)
	r := checkDB(t, c)
	if problems(r) != "cardsWithoutNote=1 missingDeck=1 tags=1" {
		t.Fatalf("problems = %s", problems(r))
	}
	repaired := collectionState(c)
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Check database" {
		t.Fatalf("undo = %q", undo)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := collectionState(c); !reflect.DeepEqual(got, corrupted) {
		t.Fatalf("undo did not restore:\n%s", jsonDiff(got, corrupted))
	}
	if _, err := c.Redo(); err != nil {
		t.Fatal(err)
	}
	if got := collectionState(c); !reflect.DeepEqual(got, repaired) {
		t.Fatalf("redo did not repair:\n%s", jsonDiff(got, repaired))
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if got := collectionState(c); !reflect.DeepEqual(got, corrupted) {
		t.Fatalf("second undo did not restore:\n%s", jsonDiff(got, corrupted))
	}
	reopen(t, c, store, clk)
}

// If the repair cannot be saved, the collection stays unchanged.
func TestCheckDatabaseFailedSave(t *testing.T) {
	store := storage.NewMemoryStore()
	clk := &clock{t: time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)}
	fb := &failingBackend{Backend: store.Open()}
	c, _, err := Open(fb, Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	id := addNote(t, c, "Basic", "hola", "hello")
	corrupt(t, c, &op{DelNotes: []ID{id}})
	before := collectionState(c)
	fb.fail = true
	if _, err := c.CheckDatabase(func(string, int, int) {}); err == nil || IsUserError(err) {
		t.Fatalf("err = %v", err)
	}
	if got := collectionState(c); !reflect.DeepEqual(got, before) {
		t.Fatalf("collection changed:\n%s", jsonDiff(got, before))
	}
	fb.fail = false
	if r := checkDB(t, c); problems(r) != "cardsWithoutNote=1" {
		t.Fatalf("problems = %s", problems(r))
	}
}

// The check reports progress and compacts the storage.
func TestCheckDatabaseProgressAndCompaction(t *testing.T) {
	c, _, _ := testCollection(t)
	addNote(t, c, "Basic", "hola", "hello")
	if c.store.JournalEntries() == 0 {
		t.Fatal("no journal entries to compact")
	}
	stages := map[string]bool{}
	if _, err := c.CheckDatabase(func(stage string, done, total int) { stages[stage] = true }); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"notes", "cards", "history", "saving"} {
		if !stages[s] {
			t.Fatalf("no %q progress in %v", s, stages)
		}
	}
	if n := c.store.JournalEntries(); n != 0 {
		t.Fatalf("journal entries after check = %d", n)
	}
}

// Checking 50,000 cards stays fast enough to run with a progress bar.
func TestCheckDatabase50k(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	start := time.Now()
	r, err := c.CheckDatabase(func(string, int, int) {})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("checked 50,000 cards in %v: %s", time.Since(start), problems(r))
	if len(r.Problems) != 0 {
		t.Fatalf("problems = %s", problems(r))
	}
}
