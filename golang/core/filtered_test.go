package core

import (
	"cardflip/golang/internal/usermsg"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"cardflip/golang/scheduler"
)

// taggedCards adds n Basic notes tagged exam to deck and returns their
// cards.
func taggedCards(t *testing.T, c *Collection, deck ID, n int) []ID {
	t.Helper()
	var notes, cards []ID
	for i := range n {
		notes = append(notes, addNoteIn(t, c, deck, "Basic", fmt.Sprintf("q%d", i), "a"))
	}
	if err := c.AddTags(notes, []string{"exam"}); err != nil {
		t.Fatal(err)
	}
	for _, nid := range notes {
		cards = append(cards, c.CardsOfNote(nid)[0].ID)
	}
	return cards
}

// buildFiltered creates a filtered deck and checks how many cards it got.
func buildFiltered(t *testing.T, c *Collection, name string, cfg FilterConfig, want int) ID {
	t.Helper()
	id, n, err := c.CreateFilteredDeck(name, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if n != want {
		t.Fatalf("%s holds %d cards, want %d", name, n, want)
	}
	return id
}

// cardsIn returns the cards in deck, sorted by ID.
func cardsIn(c *Collection, deck ID) []Card {
	var out []Card
	for _, card := range c.AllCards() {
		if card.DeckID == deck {
			out = append(out, card)
		}
	}
	slices.SortFunc(out, func(a, b Card) int { return int(a.ID - b.ID) })
	return out
}

func filterConfig(search string, limit int, order string) FilterConfig {
	cfg := DefaultFilterConfig(search)
	cfg.Terms[0].Limit, cfg.Terms[0].Order = limit, order
	return cfg
}

// Build from a tag (filtered-decks spec): up to the limit of matching
// cards move in and remember their home decks.
func TestBuildFilteredFromTag(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	taggedCards(t, c, spanish, 80)
	addNote(t, c, "Basic", "untagged", "")
	cram := buildFiltered(t, c, "Cram", filterConfig("tag:exam", 50, FilterRandom), 50)
	in := cardsIn(c, cram)
	if len(in) != 50 {
		t.Fatalf("%d cards in the deck", len(in))
	}
	var pos []int
	for _, card := range in {
		if card.HomeDeckID != spanish || card.State != scheduler.New {
			t.Fatalf("card %+v", card)
		}
		pos = append(pos, card.FilteredPos)
	}
	slices.Sort(pos)
	if pos[0] != 1 || pos[49] != 50 {
		t.Fatalf("positions %v", pos)
	}
	if d, _ := c.Deck(cram); d.Filter == nil || d.Filter.Terms[0].Search != "tag:exam" {
		t.Fatalf("deck %+v", d)
	}
	reopen(t, c, store, clk)
}

// Suspended excluded (filtered-decks spec), and buried cards and cards in
// other filtered decks; the second filter takes cards after the first.
func TestBuildFilteredExclusions(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 6)
	if err := c.SetSuspended(cards[:1], true); err != nil {
		t.Fatal(err)
	}
	if err := c.BuryCards(cards[1:2]); err != nil {
		t.Fatal(err)
	}
	other := buildFiltered(t, c, "Other", filterConfig(fmt.Sprintf("cid:%d", cards[2]), 10, FilterAdded), 1)
	cfg := filterConfig("tag:exam", 1, FilterAdded)
	cfg.Terms = append(cfg.Terms, FilterTerm{Search: "deck:Spanish", Limit: 5, Order: FilterLatestAdded})
	cram := buildFiltered(t, c, "Cram", cfg, 3)
	got := map[ID]int{}
	for _, card := range cardsIn(c, cram) {
		got[card.ID] = card.FilteredPos
	}
	// First filter: the earliest added of 3, 4, 5; second: the rest, latest
	// first.
	want := map[ID]int{cards[3]: 1, cards[5]: 2, cards[4]: 3}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("positions %v, want %v", got, want)
	}
	for _, id := range cards[:2] {
		if card := c.mustCard(t, id); card.DeckID != spanish || card.HomeDeckID != 0 {
			t.Fatalf("excluded card %+v", card)
		}
	}
	if card := c.mustCard(t, cards[2]); card.DeckID != other {
		t.Fatalf("card of the other filtered deck moved: %+v", card)
	}
}

// Building with no matching card is an error and changes nothing.
func TestBuildFilteredNoMatch(t *testing.T) {
	c, _, _ := testCollection(t)
	before := deckNames(c)
	_, _, err := c.CreateFilteredDeck("Cram", DefaultFilterConfig("tag:nothing"))
	mustUserErr(t, err, "No cards matched")
	if deckNames(c) != before {
		t.Fatalf("decks %s", deckNames(c))
	}
	_, _, err = c.CreateFilteredDeck("Cram", DefaultFilterConfig("deck:("))
	if err == nil {
		t.Fatal("invalid search accepted")
	}
	cfg := DefaultFilterConfig("")
	cfg.Terms[0].Order = "sideways"
	_, _, err = c.CreateFilteredDeck("Cram", cfg)
	mustUserErr(t, err, "Unknown filtered deck order")
}

// Orders pick and sort a search's cards.
func TestFilteredOrders(t *testing.T) {
	c, _, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 3)
	now := clk.now()
	// Review cards: intervals 5, 20, 10; lapses 2, 0, 7.
	ivls, lapses := []int{5, 20, 10}, []int{2, 0, 7}
	o := &op{}
	for i, id := range cards {
		card := c.mustCard(t, id)
		card.CardState = scheduler.CardState{State: scheduler.Review, ScheduledDays: ivls[i], Lapses: lapses[i],
			Due: now.AddDate(0, 0, ivls[i]-3), LastReview: now.AddDate(0, 0, -3), Memory: &scheduler.Memory{Stability: float64(ivls[i]), Difficulty: 5}}
		o.Cards = append(o.Cards, &card)
	}
	corrupt(t, c, o)
	for _, tc := range []struct {
		order string
		want  []int
	}{
		{FilterIntervalsAscending, []int{0, 2, 1}},
		{FilterIntervalsDescending, []int{1, 2, 0}},
		{FilterMostLapses, []int{2, 0, 1}},
		{FilterAdded, []int{0, 1, 2}},
		{FilterLatestAdded, []int{2, 1, 0}},
		{FilterDue, []int{0, 2, 1}},
		{FilterRetrievability, []int{0, 2, 1}},
	} {
		d := buildFiltered(t, c, "Cram", filterConfig("tag:exam", 10, tc.order), 3)
		q, _ := c.Queue(d)
		var got []int
		for _, id := range q.Review {
			got = append(got, slices.Index(cards, id))
		}
		if fmt.Sprint(got) != fmt.Sprint(tc.want) {
			t.Errorf("%s: order %v, want %v", tc.order, got, tc.want)
		}
		if err := c.DeleteDeck(d); err != nil {
			t.Fatal(err)
		}
	}
}

// Empty (filtered-decks spec): the cards are back home with their
// scheduling as it is, and the deck is empty.
func TestEmptyFilteredDeck(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 50)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 50)
	if _, err := c.Answer(cards[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	learning := c.mustCard(t, cards[0]).CardState
	if err := c.EmptyFilteredDeck(cram); err != nil {
		t.Fatal(err)
	}
	if n := len(cardsIn(c, cram)); n != 0 {
		t.Fatalf("%d cards left", n)
	}
	for _, id := range cards {
		card := c.mustCard(t, id)
		if card.DeckID != spanish || card.HomeDeckID != 0 || card.FilteredPos != 0 {
			t.Fatalf("card %+v", card)
		}
	}
	if got := c.mustCard(t, cards[0]).CardState; got.State != learning.State || !got.Due.Equal(learning.Due) {
		t.Fatalf("scheduling changed: %+v", got)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Empty filtered deck" {
		t.Fatalf("undo = %q", undo)
	}
	if _, err := c.Deck(cram); err != nil {
		t.Fatal("the emptied deck is gone")
	}
	reopen(t, c, store, clk)
}

// Rebuilding returns the current cards home first and builds again.
func TestRebuildFilteredDeck(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 4)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 4)
	gone := c.mustCard(t, cards[0]).NoteID
	if err := c.RemoveTags([]ID{gone}, []string{"exam"}); err != nil {
		t.Fatal(err)
	}
	if n, err := c.RebuildFilteredDeck(cram); err != nil || n != 3 {
		t.Fatalf("rebuild = %d, %v", n, err)
	}
	if card := c.mustCard(t, cards[0]); card.DeckID != spanish || card.HomeDeckID != 0 {
		t.Fatalf("untagged card %+v", card)
	}
	if n := len(cardsIn(c, cram)); n != 3 {
		t.Fatalf("%d cards in the deck", n)
	}
	// Changing the search rebuilds too; a search matching nothing keeps
	// the deck as it is.
	cfg := DefaultFilterConfig(fmt.Sprintf("cid:%d", cards[1]))
	if n, err := c.UpdateFilteredDeck(cram, "Exam::Cram", cfg); err != nil || n != 1 {
		t.Fatalf("update = %d, %v", n, err)
	}
	if d, _ := c.Deck(cram); d.Name != "Exam::Cram" || d.Filter.Terms[0].Search != cfg.Terms[0].Search {
		t.Fatalf("deck %+v", d)
	}
	_, err := c.UpdateFilteredDeck(cram, "Exam::Cram", DefaultFilterConfig("tag:nothing"))
	mustUserErr(t, err, "No cards matched")
	if n := len(cardsIn(c, cram)); n != 1 {
		t.Fatalf("%d cards in the deck", n)
	}
	if _, err := c.RebuildFilteredDeck(spanish); err == nil {
		t.Fatal("rebuilt a regular deck")
	}
}

// Delete filtered deck (collection spec): no card is deleted and the cards
// are back home.
func TestDeleteFilteredDeck(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 20)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 20)
	if info, _ := c.DeckDeletionInfo(cram); info != (DeckDeletion{Decks: 1, Returned: 20}) {
		t.Fatalf("info %+v", info)
	}
	if err := c.DeleteDeck(cram); err != nil {
		t.Fatal(err)
	}
	if n := len(c.AllCards()); n != 20 {
		t.Fatalf("%d cards left", n)
	}
	for _, id := range cards {
		if card := c.mustCard(t, id); card.DeckID != spanish || card.HomeDeckID != 0 {
			t.Fatalf("card %+v", card)
		}
	}
	reopen(t, c, store, clk)
}

// Deleting a regular deck also deletes its cards in filtered decks, and the
// confirmation counts them.
func TestDeleteHomeDeckOfFilteredCards(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	french, _ := c.CreateDeck("French")
	taggedCards(t, c, spanish, 3)
	keep := taggedCards(t, c, french, 2)
	cram := buildFiltered(t, c, "Spanish::Cram", DefaultFilterConfig("tag:exam"), 5)
	if info, _ := c.DeckDeletionInfo(spanish); info != (DeckDeletion{Decks: 2, Cards: 3, Returned: 2}) {
		t.Fatalf("info %+v", info)
	}
	if err := c.DeleteDeck(spanish); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Deck(cram); err == nil {
		t.Fatal("the filtered subdeck is left")
	}
	if n := len(c.AllCards()); n != 2 {
		t.Fatalf("%d cards left", n)
	}
	for _, id := range keep {
		if card := c.mustCard(t, id); card.DeckID != french || card.HomeDeckID != 0 {
			t.Fatalf("card %+v", card)
		}
	}

	// Cards away in a filtered deck elsewhere count as in filtered decks.
	cram = buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 2)
	if info, _ := c.DeckDeletionInfo(french); info != (DeckDeletion{Decks: 1, Cards: 2, InFiltered: 2}) {
		t.Fatalf("info %+v", info)
	}
	if err := c.DeleteDeck(french); err != nil {
		t.Fatal(err)
	}
	if n := len(cardsIn(c, cram)); n != 0 {
		t.Fatalf("%d cards left in the filtered deck", n)
	}
}

// Filtered deck limits (filtered-decks spec): no subdecks, no notes added,
// no cards moved in; the last regular deck stays.
func TestFilteredDeckLimits(t *testing.T) {
	c, _, _ := testCollection(t)
	def := defaultDeck(c)
	taggedCards(t, c, def, 1)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 1)
	_, err := c.CreateDeck("Cram::Sub")
	mustUserErr(t, err, "cannot have subdecks")
	other, _ := c.CreateDeck("Other")
	mustUserErr(t, c.RenameDeck(other, "Cram::Other"), "cannot have subdecks")
	_, _, err = c.CreateFilteredDeck("Cram::Again", DefaultFilterConfig(""))
	mustUserErr(t, err, "cannot have subdecks")
	_, err = c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic").ID, DeckID: cram, Fields: []string{"q"}})
	mustUserErr(t, err, "filtered deck")
	mustUserErr(t, c.MoveCards([]ID{cardsIn(c, cram)[0].ID}, cram), "filtered deck")
	_, err = c.Import(&ImportBatch{NoteTypes: []ImportNoteType{{Key: "b", ExistingID: noteType(t, c, "Basic").ID}},
		Notes: []ImportNote{{TypeKey: "b", DeckName: "Cram", Fields: []string{"x", "y"}}}}, nil)
	mustUserErr(t, err, "filtered deck")
	if err := c.DeleteDeck(other); err != nil {
		t.Fatal(err)
	}
	mustUserErr(t, c.DeleteDeck(def), "last deck")

	// Moving a card out of a filtered deck makes the target its home.
	other, _ = c.CreateDeck("Other")
	id := cardsIn(c, cram)[0].ID
	if err := c.MoveCards([]ID{id}, other); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, id); card.DeckID != other || card.HomeDeckID != 0 {
		t.Fatalf("moved card %+v", card)
	}
}

// Audit (design decision 1): cards in a filtered deck report the filtered
// deck for display and the home deck for options.
func TestFilteredDeckAudit(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	exam := ownPreset(t, c, spanish)
	opts := DefaultDeckOptions()
	opts.LearningSteps, opts.NewPerDay = []int{120, 7200}, 1
	if err := c.SetPresetOptions(exam, opts); err != nil {
		t.Fatal(err)
	}
	cards := taggedCards(t, c, spanish, 3)
	before, _ := c.Previews(cards[0])
	cram := buildFiltered(t, c, "Cram", filterConfig("tag:exam", 2, FilterAdded), 2)
	id := cards[0]

	// Options: scheduling, leech handling and rendering use the home deck.
	if after, _ := c.Previews(id); after != before {
		t.Fatalf("previews %v, want %v", after, before)
	}
	c.mu.RLock()
	card := c.cards[id]
	_, _, rn, _ := c.renderInput(card)
	o := c.deckOptions(card)
	c.mu.RUnlock()
	if rn.Deck != "Spanish" || o.NewPerDay != 1 {
		t.Fatalf("{{Deck}} = %q, options %+v", rn.Deck, o)
	}
	// Counts: the filtered deck counts its cards, the home deck the rest.
	counts := c.DeckCounts()
	if counts[cram].New != 2 || counts[spanish].New != 1 {
		t.Fatalf("counts %v", counts)
	}
	// Browser: the deck column shows the filtered deck with the home deck.
	if rows := c.Rows([]ID{id}, ColDeck); rows[0].Deck != "Cram (Spanish)" {
		t.Fatalf("deck column %q", rows[0].Deck)
	}
	if c.DeckLabel(cards[2]) != "Spanish" {
		t.Fatalf("label %q", c.DeckLabel(cards[2]))
	}
	// Search: deck: matches both decks, preset: the home deck's.
	for _, q := range []string{"deck:Cram", "deck:Spanish", "deck:filtered", "preset:\"Own Spanish\"", "-deck:Default"} {
		if !slices.Contains(searchIDs(t, c, q), id) {
			t.Errorf("%s does not match", q)
		}
	}
	if slices.Contains(searchIDs(t, c, "deck:filtered"), cards[2]) {
		t.Error("deck:filtered matches a card at home")
	}
	// Export: the cards are in their home deck; the filtered deck is left
	// out, also when it is exported.
	for _, scope := range []ExportScope{{DeckID: spanish}, {DeckID: cram}, {All: true}, {CardIDs: []ID{id}}} {
		set, err := c.Export(scope)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range set.Decks {
			if d.Filter != nil {
				t.Errorf("%+v: exported filtered deck", scope)
			}
		}
		for _, card := range set.Cards {
			if card.DeckID != spanish || card.HomeDeckID != 0 || card.FilteredPos != 0 {
				t.Errorf("%+v: exported card %+v", scope, card)
			}
		}
		if !slices.ContainsFunc(set.Cards, func(card Card) bool { return card.ID == id }) {
			t.Errorf("%+v: card not exported", scope)
		}
	}
	// New cards of an existing note go to its home deck.
	c.mu.RLock()
	deck := c.deckOfNote(card.NoteID)
	c.mu.RUnlock()
	if deck != spanish {
		t.Fatalf("new cards go to %d", deck)
	}
}

// Filtered decks show all their cards without daily limits.
func TestFilteredQueueWithoutLimits(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	if err := c.SetDeckOptions(spanish, deckOpts(0, 0, 0.9)); err != nil {
		t.Fatal(err)
	}
	taggedCards(t, c, spanish, 30)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 30)
	q, _ := c.Queue(cram)
	if len(q.New) != 30 || q.NewReviewOrder != NewMix {
		t.Fatalf("queue %+v", q)
	}
	// A filtered subdeck's cards are studied there, not in the parent.
	if err := c.EmptyFilteredDeck(cram); err != nil {
		t.Fatal(err)
	}
	sub := buildFiltered(t, c, "Spanish::Cram", DefaultFilterConfig("tag:exam"), 30)
	counts := c.DeckCounts()
	if counts[spanish].New != 0 || counts[sub].New != 30 {
		t.Fatalf("counts %v", counts)
	}
}

// Review ahead (filtered-decks spec): a card reviewed before it is due gets
// a shorter interval than on its due date, and is back home.
func TestFilteredReviewAhead(t *testing.T) {
	c, _, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	id := taggedCards(t, c, spanish, 1)[0]
	now := clk.now()
	card := c.mustCard(t, id)
	card.CardState = scheduler.CardState{State: scheduler.Review, ScheduledDays: 10, Reps: 3,
		Due: now.AddDate(0, 0, 5), LastReview: now.AddDate(0, 0, -5), Memory: &scheduler.Memory{Stability: 10, Difficulty: 5}}
	corrupt(t, c, &op{Cards: []*Card{&card}})
	c.mu.RLock()
	onDue := scheduler.Answer(card.CardState, scheduler.Good, card.Due, c.params(&card), int64(id))
	c.mu.RUnlock()

	cram := buildFiltered(t, c, "Ahead", filterConfig("prop:due<=5 prop:due>0", 10, FilterDue), 1)
	q, _ := c.Queue(cram)
	if len(q.Review) != 1 || q.Counts().Review != 1 {
		t.Fatalf("queue %+v", q)
	}
	if _, err := c.Answer(id, scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	got := c.mustCard(t, id)
	if got.ScheduledDays >= onDue.State.ScheduledDays {
		t.Fatalf("interval %d, on the due date %d", got.ScheduledDays, onDue.State.ScheduledDays)
	}
	if got.DeckID != spanish || got.HomeDeckID != 0 {
		t.Fatalf("card %+v", got)
	}
}

// A card in a filtered deck stays there while it is learning and returns
// home when it graduates; a review card answered Again stays to relearn.
func TestFilteredCardsReturnOnGraduation(t *testing.T) {
	c, _, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 2)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 2)
	if _, err := c.Answer(cards[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[0]); card.DeckID != cram || card.State != scheduler.Learning {
		t.Fatalf("card %+v", card)
	}
	// The 10 minute step is within the learn ahead limit.
	if q, _ := c.Queue(cram); !slices.Equal(q.LearningAhead, cards[:1]) || len(q.Learning) != 0 {
		t.Fatalf("queue %+v", q)
	}
	clk.advance(time.Hour)
	if q, _ := c.Queue(cram); !slices.Contains(q.Learning, cards[0]) {
		t.Fatalf("queue %+v", q)
	}
	if _, err := c.Answer(cards[0], scheduler.Good, time.Second); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[0]); card.DeckID != spanish || card.State != scheduler.Review {
		t.Fatalf("card %+v", card)
	}
	if _, err := c.Answer(cards[1], scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[1]); card.DeckID != spanish {
		t.Fatalf("card %+v", card)
	}

	// Again on a review card: relearning in the filtered deck.
	clk.advance(48 * time.Hour)
	buildFiltered(t, c, "Again", DefaultFilterConfig("tag:exam"), 2)
	if _, err := c.Answer(cards[0], scheduler.Again, time.Second); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[0]); card.HomeDeckID != spanish || card.State != scheduler.Relearning {
		t.Fatalf("card %+v", card)
	}
}

// Preview new cards (filtered-decks spec): Good with a delay of 0 leaves
// the card New, back home, and not shown again; Again and Hard show it
// again after their delays without changing its scheduling.
func TestFilteredPreview(t *testing.T) {
	c, store, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 4)
	cfg := filterConfig("tag:exam", 10, FilterAdded)
	cfg.Reschedule = false
	cram := buildFiltered(t, c, "Preview", cfg, 4)
	previews, _ := c.Previews(cards[0])
	var labels []string
	for _, p := range previews {
		labels = append(labels, p.Label)
	}
	if got := strings.Join(labels, " "); got != "<1m <10m End End" {
		t.Fatalf("labels %q", got)
	}

	e, err := c.Answer(cards[0], scheduler.Good, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	card := c.mustCard(t, cards[0])
	if card.State != scheduler.New || card.Reps != 0 || card.DeckID != spanish || card.HomeDeckID != 0 || !e.Preview {
		t.Fatalf("card %+v, entry %+v", card, e)
	}
	if q, _ := c.Queue(cram); slices.Contains(q.New, cards[0]) || len(q.New) != 3 {
		t.Fatalf("queue %+v", q)
	}

	now := clk.now()
	if _, err := c.Answer(cards[1], scheduler.Again, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Answer(cards[2], scheduler.Hard, time.Second); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[1]); card.State != scheduler.New || card.DeckID != cram || !card.PreviewDue.Equal(now.Add(time.Minute)) {
		t.Fatalf("card %+v", card)
	}
	q, _ := c.Queue(cram)
	if !slices.Equal(q.New, cards[3:]) || !slices.Equal(q.LearningAhead, cards[1:3]) {
		t.Fatalf("queue %+v", q)
	}
	clk.advance(2 * time.Minute)
	if q, _ := c.Queue(cram); !slices.Equal(q.Learning, cards[1:2]) {
		t.Fatalf("queue %+v", q)
	}
	if _, err := c.Answer(cards[1], scheduler.Easy, time.Second); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[1]); card.State != scheduler.New || card.DeckID != spanish {
		t.Fatalf("card %+v", card)
	}

	// Previews are no reviews: the home deck's new cards and rated: are
	// unaffected.
	if counts := c.DeckCounts(); counts[spanish].New != 2 {
		t.Fatalf("counts %v", counts)
	}
	if ids := searchIDs(t, c, "rated:1"); len(ids) != 0 {
		t.Fatalf("rated:1 = %v", ids)
	}
	if h := c.ReviewsOfCard(cards[0]); len(h) != 1 || !h[0].Preview {
		t.Fatalf("history %+v", h)
	}

	// Undoing an older preview answer puts the card back in the deck.
	if err := c.SetFlag([]ID{cards[3]}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UndoReview(e.ID); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[0]); card.DeckID != cram || card.HomeDeckID != spanish || card.FilteredPos != 1 {
		t.Fatalf("card after undo %+v", card)
	}
	reopen(t, c, store, clk)
}

// A backup taken while cards are in a filtered deck restores them there.
func TestRestoreFilteredDeck(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 3)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 3)
	snap, err := c.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.EmptyFilteredDeck(cram); err != nil {
		t.Fatal(err)
	}
	if err := c.Restore(snap, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range cards {
		if card := c.mustCard(t, id); card.DeckID != cram || card.HomeDeckID != spanish {
			t.Fatalf("card %+v", card)
		}
	}
	if err := c.EmptyFilteredDeck(cram); err != nil {
		t.Fatal(err)
	}
	if card := c.mustCard(t, cards[0]); card.DeckID != spanish {
		t.Fatalf("card %+v", card)
	}
}

// Check database returns cards whose filtered deck is gone.
func TestCheckDatabaseFilteredCards(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	cards := taggedCards(t, c, spanish, 2)
	cram := buildFiltered(t, c, "Cram", DefaultFilterConfig("tag:exam"), 2)
	corrupt(t, c, &op{DelDecks: []ID{cram}})
	r := checkDB(t, c)
	if got := problems(r); !strings.Contains(got, ProblemMissingDeck+"=2") {
		t.Fatalf("problems %s", got)
	}
	for _, id := range cards {
		if card := c.mustCard(t, id); card.DeckID != spanish || card.HomeDeckID != 0 {
			t.Fatalf("card %+v", card)
		}
	}
}

// Forgotten cards (filtered-decks spec): the session holds the deck's
// cards rated Again in the last N days.
func TestCustomStudyForgotten(t *testing.T) {
	c, _, clk := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	verbs, _ := c.CreateDeck("Spanish::Verbs")
	french, _ := c.CreateDeck("French")
	old := taggedCards(t, c, spanish, 1)[0]
	recent := taggedCards(t, c, verbs, 1)[0]
	good := taggedCards(t, c, spanish, 1)[0]
	other := taggedCards(t, c, french, 1)[0]
	if _, err := c.Answer(old, scheduler.Again, time.Second); err != nil {
		t.Fatal(err)
	}
	clk.advance(72 * time.Hour)
	for id, r := range map[ID]scheduler.Rating{recent: scheduler.Again, good: scheduler.Good, other: scheduler.Again} {
		if _, err := c.Answer(id, r, time.Second); err != nil {
			t.Fatal(err)
		}
	}
	clk.advance(24 * time.Hour)
	session, n, err := c.CustomStudy(spanish, CustomStudy{Kind: CustomForgotten, Days: 2})
	if err != nil || n != 1 {
		t.Fatalf("custom study = %d, %v", n, err)
	}
	d, _ := c.Deck(session)
	if d.Name != CustomStudyDeckName || d.Filter == nil || !d.Filter.Reschedule {
		t.Fatalf("deck %+v", d)
	}
	if in := cardsIn(c, session); len(in) != 1 || in[0].ID != recent || in[0].HomeDeckID != verbs {
		t.Fatalf("cards %+v", in)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Custom study" {
		t.Fatalf("undo = %q", undo)
	}
}

// Replace previous session (filtered-decks spec): its cards return home and
// the deck is built again for the new choice.
func TestCustomStudyReplacesSession(t *testing.T) {
	c, _, _ := testCollection(t)
	spanish, _ := c.CreateDeck("Spanish")
	french, _ := c.CreateDeck("French")
	es := taggedCards(t, c, spanish, 2)
	fr := taggedCards(t, c, french, 3)
	first, n, err := c.CustomStudy(spanish, CustomStudy{Kind: CustomByState, State: CustomNew, Cards: 10})
	if err != nil || n != 2 {
		t.Fatalf("first = %d, %v", n, err)
	}
	second, n, err := c.CustomStudy(french, CustomStudy{Kind: CustomByState, State: CustomNew, Cards: 10})
	if err != nil || n != 3 || second != first {
		t.Fatalf("second = %d (deck %d, was %d), %v", n, second, first, err)
	}
	for _, id := range es {
		if card := c.mustCard(t, id); card.DeckID != spanish || card.HomeDeckID != 0 {
			t.Fatalf("previous session's card %+v", card)
		}
	}
	for _, id := range fr {
		if card := c.mustCard(t, id); card.DeckID != first {
			t.Fatalf("card %+v", card)
		}
	}
	// No match keeps the previous session.
	_, _, err = c.CustomStudy(spanish, CustomStudy{Kind: CustomForgotten, Days: 1})
	mustUserErr(t, err, "No cards matched")
	if n := len(cardsIn(c, first)); n != 3 {
		t.Fatalf("%d cards in the session", n)
	}
	// A session cannot be made from a filtered deck, nor replace a regular
	// deck of that name.
	_, _, err = c.CustomStudy(first, CustomStudy{Kind: CustomByState, State: CustomAll, Cards: 1})
	mustUserErr(t, err, "regular deck")
	if err := c.DeleteDeck(first); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateDeck(CustomStudyDeckName); err != nil {
		t.Fatal(err)
	}
	_, _, err = c.CustomStudy(spanish, CustomStudy{Kind: CustomByState, State: CustomAll, Cards: 1})
	mustUserErr(t, err, "Rename it")
}

// Each custom study choice gathers its cards.
func TestCustomStudyChoices(t *testing.T) {
	c, _, clk := testCollection(t)
	spanish, _ := c.CreateDeck(`Spanish "B1"`)
	cards := taggedCards(t, c, spanish, 5)
	now := clk.now()
	// 0: due in 3 days, 1: due today, 2: due in 10 days; 3, 4 new, 4
	// tagged verb.
	o := &op{}
	for i, days := range []int{3, 0, 10} {
		card := c.mustCard(t, cards[i])
		card.CardState = scheduler.CardState{State: scheduler.Review, ScheduledDays: 20, Due: now.AddDate(0, 0, days),
			LastReview: now.AddDate(0, 0, days-20), Memory: &scheduler.Memory{Stability: 20, Difficulty: 5}}
		o.Cards = append(o.Cards, &card)
	}
	corrupt(t, c, o)
	if err := c.AddTags([]ID{c.mustCard(t, cards[4]).NoteID}, []string{"verb"}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		cs         CustomStudy
		want       []int
		reschedule bool
	}{
		{CustomStudy{Kind: CustomAhead, Days: 5}, []int{0}, true},
		{CustomStudy{Kind: CustomAhead, Days: 30}, []int{0, 2}, true},
		{CustomStudy{Kind: CustomPreviewNew, Days: 1}, []int{3, 4}, false},
		{CustomStudy{Kind: CustomByState, State: CustomNew, Cards: 1}, []int{3}, true},
		{CustomStudy{Kind: CustomByState, State: CustomDue, Cards: 10}, []int{1}, true},
		{CustomStudy{Kind: CustomByState, State: CustomAll, Cards: 10}, []int{0, 1, 2, 3, 4}, false},
		{CustomStudy{Kind: CustomByState, State: CustomAll, Tags: []string{"verb", "noun"}, Cards: 10}, []int{4}, false},
	} {
		session, _, err := c.CustomStudy(spanish, tc.cs)
		if err != nil {
			t.Errorf("%+v: %v", tc.cs, err)
			continue
		}
		var got []int
		for _, card := range cardsIn(c, session) {
			got = append(got, slices.Index(cards, card.ID))
		}
		d, _ := c.Deck(session)
		if fmt.Sprint(got) != fmt.Sprint(tc.want) || d.Filter.Reschedule != tc.reschedule {
			t.Errorf("%+v: cards %v reschedule %v, want %v %v", tc.cs, got, d.Filter.Reschedule, tc.want, tc.reschedule)
		}
	}
	for cs, msg := range map[*CustomStudy]string{
		{Kind: CustomForgotten}:                             "number of days",
		{Kind: CustomAhead, Days: MaxCustomDays + 1}:        "number of days",
		{Kind: CustomByState, State: CustomNew}:             "number of cards",
		{Kind: CustomByState, State: "suspended", Cards: 1}: "Unknown card state",
		{Kind: "later"}:                                     "Unknown custom study",
	} {
		_, _, err := c.CustomStudy(spanish, *cs)
		mustUserErr(t, err, msg)
	}
}
