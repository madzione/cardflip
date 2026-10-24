package apkg

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/render"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"

	"google.golang.org/protobuf/encoding/protowire"
)

// The clock in these tests is 2026-10-13 12:00 UTC: 285 days after the
// fixtures' crt (2026-01-01 04:00 UTC), so Anki's due day 300 is 15 days
// away.
var testNow = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC).In(time.Local)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func newCollection(t *testing.T) *core.Collection {
	t.Helper()
	c, _, err := core.Open(storage.NewMemory(), core.Options{Now: func() time.Time { return testNow }})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func importFixture(t *testing.T, c *core.Collection, name string, keep bool) core.ImportSummary {
	t.Helper()
	sum, err := Import(c, fixture(t, name), Options{KeepScheduling: keep}, nil, nil)
	if err != nil {
		t.Fatalf("import %s: %v", name, err)
	}
	return sum
}

// noteByGUID finds an imported note and its cards.
func noteByGUID(t *testing.T, c *core.Collection, guid string) (core.Note, []core.Card) {
	t.Helper()
	ids, _ := c.Search("", core.ColCreated, false)
	for _, row := range c.Rows(ids) {
		n, _ := c.Note(row.NoteID)
		if n.GUID == guid {
			return n, c.CardsOfNote(n.ID)
		}
	}
	t.Fatalf("no note %s", guid)
	return core.Note{}, nil
}

type noteView struct {
	Type   string
	Fields []string
	Tags   []string
	Cards  []string
}

// collectionView summarises a collection for comparisons.
func collectionView(t *testing.T, c *core.Collection) (decks []string, notes map[string]noteView) {
	t.Helper()
	for _, d := range c.Decks() {
		decks = append(decks, d.Name)
	}
	notes = map[string]noteView{}
	ids, _ := c.Search("", core.ColCreated, false)
	for _, row := range c.Rows(ids) {
		n, _ := c.Note(row.NoteID)
		nt, _ := c.NoteType(n.TypeID)
		v := notes[n.GUID]
		v.Type, v.Fields, v.Tags = nt.Name, n.Fields, n.Tags
		card, _ := c.Card(row.CardID)
		mem := ""
		if card.Memory != nil {
			mem = strconv.FormatFloat(card.Memory.Stability, 'f', -1, 64) + "/" + strconv.FormatFloat(card.Memory.Difficulty, 'f', -1, 64)
		}
		v.Cards = append(v.Cards, strings.Join([]string{row.Deck, row.CardName, card.State.String(), card.Due.UTC().Format(time.RFC3339), mem,
			map[bool]string{true: "suspended"}[card.Suspended]}, " "))
		sort.Strings(v.Cards)
		notes[n.GUID] = v
	}
	return decks, notes
}

func TestModernAnkiExport(t *testing.T) {
	c := newCollection(t)
	// Modern Anki export (spec): decks, notes and cards appear, and cards
	// render with the original templates.
	sum := importFixture(t, c, "modern.apkg", true)
	decks, notes := collectionView(t, c)
	if strings.Join(decks, ", ") != "Default, Fixture, Fixture::Geo, Fixture::Spanish" {
		t.Fatalf("decks = %v", decks)
	}
	if len(notes) != 8 || sum.NotesAdded != 8 || sum.CardsAdded != 12 {
		t.Fatalf("notes %d, summary %+v", len(notes), sum)
	}
	n, cards := noteByGUID(t, c, "cf-cloze-paris")
	if len(cards) != 2 {
		t.Fatalf("cloze cards = %+v", cards)
	}
	r, _ := c.RenderCard(cards[0].ID, render.Env{})
	if !strings.Contains(r.Question, "[city]") || !strings.Contains(r.Question, "is the capital of France") {
		t.Fatalf("cloze question = %q", r.Question)
	}
	if r.Background != "#fdf6e3" || r.Color != "#222222" {
		t.Fatalf("note type CSS not used: %+v", r)
	}
	if !strings.Contains(r.Question, `class="cloze" style="font-weight:bold;color:blue"`) {
		t.Fatalf("cloze style = %q", r.Question)
	}
	if nt, _ := c.NoteType(n.TypeID); nt.Kind != core.Cloze {
		t.Fatalf("cloze note type kind = %v", nt.Kind)
	}
	// Multi-number deletion: {{c1,2::Paris}} is hidden on cards 1 and 2,
	// and card 3 asks for France.
	_, multi := noteByGUID(t, c, "cf-cloze-multi")
	if len(multi) != 3 {
		t.Fatalf("multi-number cloze cards = %+v", multi)
	}
	for i, want := range []string{"[...] is in France", "[...] is in France", "Paris is in [...]"} {
		if q := c.QuestionText(multi[i].ID); q != want {
			t.Fatalf("multi-number cloze card %d question = %q, want %q", i+1, q, want)
		}
	}
	_, rev := noteByGUID(t, c, "cf-rev-comer")
	if len(rev) != 2 {
		t.Fatalf("reversed cards = %d", len(rev))
	}
	if q := c.QuestionText(rev[1].ID); q != "to eat" {
		t.Fatalf("reverse question = %q", q)
	}
	// Imported image shown (card-rendering spec): the media is stored.
	for _, name := range []string{"cat.jpg", "dog.png"} {
		if data, err := c.Media(name); err != nil || len(data) == 0 {
			t.Fatalf("media %s: %v", name, err)
		}
	}
	// Audio imported (data-import spec).
	for _, name := range []string{"hola.mp3", "adios.mp3"} {
		if data, err := c.Media(name); err != nil || len(data) == 0 {
			t.Fatalf("media %s: %v", name, err)
		}
	}
	tags, _ := noteByGUID(t, c, "cf-basic-hola")
	if strings.Join(tags.Tags, " ") != "spanish verb" {
		t.Fatalf("tags = %v", tags.Tags)
	}
}

func TestLegacyExportSameContent(t *testing.T) {
	// Legacy export (spec): the same content as from the modern format.
	modern, legacy := newCollection(t), newCollection(t)
	importFixture(t, modern, "modern.apkg", true)
	importFixture(t, legacy, "legacy.apkg", true)
	md, mn := collectionView(t, modern)
	ld, ln := collectionView(t, legacy)
	if !reflect.DeepEqual(md, ld) {
		t.Fatalf("decks differ: %v vs %v", md, ld)
	}
	if !reflect.DeepEqual(mn, ln) {
		for guid := range mn {
			if !reflect.DeepEqual(mn[guid], ln[guid]) {
				t.Errorf("%s:\nmodern %+v\nlegacy %+v", guid, mn[guid], ln[guid])
			}
		}
		t.FailNow()
	}
}

func TestKeepScheduling(t *testing.T) {
	c := newCollection(t)
	importFixture(t, c, "modern.apkg", true)

	// Keep scheduling (spec): a review card due in 15 days stays due in 15.
	_, cards := noteByGUID(t, c, "cf-basic-hola")
	hola := cards[0]
	if hola.State != scheduler.Review {
		t.Fatalf("state = %v", hola.State)
	}
	if days := scheduler.DayNumber(hola.Due, 4) - scheduler.DayNumber(testNow, 4); days != 15 {
		t.Fatalf("due in %d days, want 15 (due %v)", days, hola.Due)
	}
	// Memory state taken from Anki.
	if hola.Memory == nil || hola.Memory.Stability != 14.25 || hola.Memory.Difficulty != 5.5 || hola.Lapses != 1 || hola.Reps != 4 {
		t.Fatalf("hola = %+v", hola)
	}
	// Suspended stays suspended. Card scheduled with SM-2 (spec): its
	// memory state is computed from its 8 reviews, not estimated from ivl
	// and ease, and its due date is kept.
	_, cards = noteByGUID(t, c, "cf-basic-cat")
	cat := cards[0]
	p, _ := Read(fixture(t, "modern.apkg"), nil)
	var history []scheduler.AnkiReview
	for _, r := range p.Revlog {
		if cardKeys(p)[r.CardID] == "cf-basic-cat/0" {
			history = append(history, scheduler.AnkiReview{Time: time.UnixMilli(r.ID), Ease: r.Ease, Ivl: r.Ivl, Factor: r.Factor, Type: r.Type})
		}
	}
	replayed := scheduler.FromAnki(scheduler.AnkiCard{Type: 2, Queue: -1, Due: 310, Ivl: 20, Factor: 2300, History: history}, time.Unix(p.Crt, 0))
	if len(history) != 8 || !cat.Suspended || cat.State != scheduler.Review || cat.Memory == nil ||
		*cat.Memory != *replayed.Memory || cat.Memory.Stability == 20 {
		t.Fatalf("cat = %+v mem %+v", cat, cat.Memory)
	}
	if days := scheduler.DayNumber(cat.Due, 4) - scheduler.DayNumber(testNow, 4); days != 25 {
		t.Fatalf("cat due in %d days, want 25 (due %v)", days, cat.Due)
	}
	// Learning card keeps its due time; new cards stay New.
	_, cards = noteByGUID(t, c, "cf-cloze-paris")
	if cards[0].State != scheduler.Learning || cards[0].Due.Unix() != 1790000000 || cards[1].State != scheduler.New {
		t.Fatalf("cloze cards = %+v", cards)
	}
	// New cards keep Anki's order.
	q, _ := c.Queue(c.Decks()[1].ID) // Fixture
	var order []string
	for _, id := range q.New {
		card, _ := c.Card(id)
		n, _ := c.Note(card.NoteID)
		order = append(order, n.GUID)
	}
	if strings.Join(order, ",") != "cf-basic-dog,cf-rev-comer,cf-rev-comer,cf-cloze-paris,cf-audio-hola,cf-audio-adios,cf-cloze-multi,cf-cloze-multi,cf-cloze-multi" {
		t.Fatalf("new order = %v", order)
	}
}

func TestResetScheduling(t *testing.T) {
	c := newCollection(t)
	// Reset scheduling (spec): every imported card is New.
	importFixture(t, c, "modern.apkg", false)
	ids, _ := c.Search("", core.ColCreated, false)
	if len(ids) != 12 {
		t.Fatalf("cards = %d", len(ids))
	}
	for _, id := range ids {
		card, _ := c.Card(id)
		if card.State != scheduler.New || card.Suspended || card.Memory != nil {
			t.Fatalf("card = %+v", card)
		}
	}
}

func TestImportAsNewRemovesLeechAndMarkedTags(t *testing.T) {
	for _, keep := range []bool{false, true} {
		p, err := Read(fixture(t, "modern.apkg"), nil)
		if err != nil {
			t.Fatal(err)
		}
		for i := range p.Notes {
			p.Notes[i].Tags = append(p.Notes[i].Tags, "Leech", "marked", "hard")
		}
		b, err := Convert(p, keep)
		if err != nil {
			t.Fatal(err)
		}
		c := newCollection(t)
		if _, err := c.Import(b, nil); err != nil {
			t.Fatal(err)
		}
		// Shared deck with leeches, shared deck with marks (spec): imported
		// as New, no note has the tags; with scheduling kept, every note
		// keeps them.
		all, _ := c.Search("", core.ColCreated, false)
		for _, tag := range []string{"leech", "marked"} {
			tagged, _ := c.Search("tag:"+tag, core.ColCreated, false)
			if want := map[bool]int{false: 0, true: len(all)}[keep]; len(tagged) != want {
				t.Fatalf("keep %v: %d of %d cards tagged %s", keep, len(tagged), len(all), tag)
			}
		}
		if hard, _ := c.Search("tag:hard", core.ColCreated, false); len(hard) != len(all) {
			t.Fatalf("keep %v: other tag lost on %d cards", keep, len(all)-len(hard))
		}
	}
}

func TestImportFlags(t *testing.T) {
	for _, name := range []string{"modern.apkg", "legacy.apkg"} {
		// Flag kept (spec): the green forward card keeps flag 3, its reverse
		// card has none.
		c := newCollection(t)
		importFixture(t, c, name, true)
		_, cards := noteByGUID(t, c, "cf-rev-comer")
		if cards[0].Flag != 3 || cards[1].Flag != 0 {
			t.Fatalf("%s: flags = %d, %d", name, cards[0].Flag, cards[1].Flag)
		}
		if ids, _ := c.Search("-flag:0", core.ColCreated, false); len(ids) != 1 || ids[0] != cards[0].ID {
			t.Fatalf("%s: flagged cards = %v", name, ids)
		}
		// Existing cards keep their current flag on re-import.
		if err := c.SetFlag([]core.ID{cards[0].ID}, 5); err != nil {
			t.Fatal(err)
		}
		importFixture(t, c, name, true)
		if _, cards = noteByGUID(t, c, "cf-rev-comer"); cards[0].Flag != 5 {
			t.Fatalf("%s: re-import changed flag to %d", name, cards[0].Flag)
		}

		// Import as new (spec): no card has a flag.
		c = newCollection(t)
		importFixture(t, c, name, false)
		if ids, _ := c.Search("-flag:0", core.ColCreated, false); len(ids) != 0 {
			t.Fatalf("%s: flagged cards after import as new = %v", name, ids)
		}
	}
}

func TestImportTwice(t *testing.T) {
	c := newCollection(t)
	importFixture(t, c, "modern.apkg", true)
	_, before := noteByGUID(t, c, "cf-basic-hola")
	types := len(c.NoteTypes())
	// Import twice (spec): each note once; scheduling unchanged, even when
	// the second import resets scheduling.
	sum := importFixture(t, c, "legacy.apkg", false)
	if sum.NotesAdded != 0 || sum.NotesSkipped != 8 || sum.CardsAdded != 0 {
		t.Fatalf("second import = %+v", sum)
	}
	ids, _ := c.Search("", core.ColCreated, false)
	if len(ids) != 12 {
		t.Fatalf("cards = %d", len(ids))
	}
	_, after := noteByGUID(t, c, "cf-basic-hola")
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("scheduling changed:\n%+v\n%+v", before, after)
	}
	if n := len(c.Decks()); n != 4 {
		t.Fatalf("decks = %d", n)
	}
	if n := len(c.NoteTypes()); n != types {
		t.Fatalf("note types went from %d to %d", types, n)
	}
	// Content changed later updates the note.
	p, _ := Read(fixture(t, "modern.apkg"), nil)
	for i := range p.Notes {
		if p.Notes[i].GUID == "cf-basic-hola" {
			p.Notes[i].Fields[1] = "hi"
			p.Notes[i].Mod++
		}
	}
	b, _ := Convert(p, true)
	sum, err := c.Import(b, nil)
	if err != nil || sum.NotesUpdated != 1 {
		t.Fatalf("update import = %+v %v", sum, err)
	}
	n, after := noteByGUID(t, c, "cf-basic-hola")
	if n.Fields[1] != "hi" || !reflect.DeepEqual(before, after) {
		t.Fatalf("note %q, cards changed: %v", n.Fields, !reflect.DeepEqual(before, after))
	}
}

// cardKeys maps a package's card IDs to "<guid>/<ord>".
func cardKeys(p *Package) map[int64]string {
	guids := map[int64]string{}
	for _, n := range p.Notes {
		guids[n.ID] = n.GUID
	}
	out := map[int64]string{}
	for _, c := range p.Cards {
		out[c.ID] = guids[c.NoteID] + "/" + strconv.Itoa(c.Ord)
	}
	return out
}

func TestReadReviewHistory(t *testing.T) {
	m := loadManifest(t)
	for _, name := range []string{"modern.apkg", "legacy.apkg"} {
		p, err := Read(fixture(t, name), nil)
		if err != nil {
			t.Fatal(err)
		}
		// The review log is read in full, in ID order, in both formats.
		keys := cardKeys(p)
		got := map[string][][7]int64{}
		for i, r := range p.Revlog {
			if i > 0 && r.ID <= p.Revlog[i-1].ID {
				t.Fatalf("%s: revlog not in ID order at %d", name, i)
			}
			got[keys[r.CardID]] = append(got[keys[r.CardID]],
				[7]int64{r.ID, int64(r.Ease), r.Ivl, r.LastIvl, r.Factor, r.Time, int64(r.Type)})
		}
		want := m.Packages[name].Revlog
		if len(want) != 3 || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: revlog\n got %v\nwant %v", name, got, want)
		}
	}
}

// historyOf returns the converted history of a package card.
func historyOf(t *testing.T, b *core.ImportBatch, guid string, ord int) []core.ImportReview {
	t.Helper()
	for _, n := range b.Notes {
		if n.GUID != guid {
			continue
		}
		for _, c := range n.Cards {
			if c.Ord == ord {
				return c.History
			}
		}
	}
	t.Fatalf("no card %s/%d", guid, ord)
	return nil
}

func TestConvertReviewHistory(t *testing.T) {
	crt := time.Unix(1767240000, 0)
	day := func(d int, minutes int) time.Time {
		return crt.Add(time.Duration(d)*24*time.Hour + 2*time.Hour + time.Duration(minutes)*time.Minute)
	}
	type entry struct {
		Time        time.Time
		Rating      scheduler.Rating
		StateBefore scheduler.State
		IvlBefore   int
		IvlAfter    int
		DurationMs  int64
	}
	for _, name := range []string{"modern.apkg", "legacy.apkg"} {
		p, err := Read(fixture(t, name), nil)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Convert(p, true)
		if err != nil {
			t.Fatal(err)
		}
		view := func(h []core.ImportReview) []entry {
			var out []entry
			for _, r := range h {
				out = append(out, entry{r.Time, r.Rating, r.StateBefore, r.IvlBefore, r.IvlAfter, r.DurationMs})
			}
			return out
		}
		// Learning, reviews, a lapse with relearning, a filtered review (a
		// review) and a set due date (manual, rating 0). Negative Anki
		// intervals are seconds, less than a day.
		hola := historyOf(t, b, "cf-basic-hola", 0)
		wantHola := []entry{
			{day(4, 0), scheduler.Good, scheduler.New, 0, 0, 8000},
			{day(4, 10), scheduler.Good, scheduler.Learning, 0, 1, 5000},
			{day(5, 0), scheduler.Good, scheduler.Review, 1, 3, 4000},
			{day(8, 0), scheduler.Good, scheduler.Review, 3, 8, 3500},
			{day(16, 0), scheduler.Again, scheduler.Review, 8, 0, 12000},
			{day(16, 10), scheduler.Good, scheduler.Relearning, 0, 1, 6000},
			{day(17, 0), scheduler.Good, scheduler.Review, 1, 3, 4000},
			{day(20, 0), scheduler.Good, scheduler.Review, 3, 7, 3000},
			{day(27, 0), scheduler.Easy, scheduler.Review, 7, 20, 2000},
			{day(47, 0), scheduler.Hard, scheduler.Review, 20, 25, 9000},
			{day(60, 0), 0, scheduler.Review, 25, 40, 0},
			{day(200, 0), scheduler.Good, scheduler.Review, 40, 15, 4500},
		}
		if got := view(hola); !reflect.DeepEqual(got, wantHola) {
			t.Fatalf("%s: hola history\n got %+v\nwant %+v", name, got, wantHola)
		}
		if hola[0].Source != "anki:"+strconv.FormatInt(day(4, 0).UnixMilli(), 10) {
			t.Fatalf("%s: source = %q", name, hola[0].Source)
		}
		// A cram review is a review; a reset (manual, interval 0) makes the
		// next learning entry a New card's again.
		paris := historyOf(t, b, "cf-cloze-paris", 0)
		wantParis := []entry{
			{day(250, 0), scheduler.Good, scheduler.New, 0, 0, 6000},
			{day(250, 5), scheduler.Good, scheduler.Review, 0, 0, 3000},
			{day(251, 0), 0, scheduler.Learning, 0, 0, 0},
			{day(263, 0), scheduler.Again, scheduler.New, 0, 0, 10000},
		}
		if got := view(paris); !reflect.DeepEqual(got, wantParis) {
			t.Fatalf("%s: paris history\n got %+v\nwant %+v", name, got, wantParis)
		}
		if n := len(historyOf(t, b, "cf-basic-cat", 0)); n != 8 {
			t.Fatalf("%s: cat history = %d entries", name, n)
		}
		if h := historyOf(t, b, "cf-cloze-paris", 1); h != nil {
			t.Fatalf("%s: new card history = %+v", name, h)
		}

		// Import as new: no history.
		b, err = Convert(p, false)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range b.Notes {
			for _, c := range n.Cards {
				if c.History != nil {
					t.Fatalf("%s: history imported as new: %+v", name, c.History)
				}
			}
		}
	}
}

// storedReviews returns the stored review log entries of a card.
func storedReviews(t *testing.T, c *core.Collection, card core.ID) []core.ReviewEntry {
	t.Helper()
	months, err := c.RevlogMonths()
	if err != nil {
		t.Fatal(err)
	}
	var out []core.ReviewEntry
	for _, m := range months {
		entries, err := c.ReadRevlog(m)
		if err != nil {
			t.Fatal(err)
		}
		for _, raw := range entries {
			var e core.ReviewEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				t.Fatal(err)
			}
			if card == 0 || e.CardID == card {
				out = append(out, e)
			}
		}
	}
	return out
}

func TestImportReviewHistory(t *testing.T) {
	for _, name := range []string{"modern.apkg", "legacy.apkg"} {
		// History kept (spec): the card reviewed 12 times has the 12
		// entries with their dates and ratings.
		c := newCollection(t)
		sum := importFixture(t, c, name, true)
		if sum.Reviews != 24 {
			t.Fatalf("%s: summary = %+v", name, sum)
		}
		_, cards := noteByGUID(t, c, "cf-basic-hola")
		hola := storedReviews(t, c, cards[0].ID)
		var ratings []int
		for _, e := range hola {
			ratings = append(ratings, int(e.Rating))
		}
		if len(hola) != 12 || !hola[0].Time.Equal(time.UnixMilli(1767592800000)) ||
			!reflect.DeepEqual(ratings, []int{3, 3, 3, 3, 1, 3, 3, 3, 4, 2, 0, 3}) {
			t.Fatalf("%s: hola history = %+v", name, hola)
		}

		// Re-import (spec): still 12 entries.
		if sum := importFixture(t, c, name, true); sum.Reviews != 0 {
			t.Fatalf("%s: re-import summary = %+v", name, sum)
		}
		if n := len(storedReviews(t, c, cards[0].ID)); n != 12 {
			t.Fatalf("%s: after re-import: %d entries", name, n)
		}
		if n := len(storedReviews(t, c, 0)); n != 24 {
			t.Fatalf("%s: after re-import: %d entries in all", name, n)
		}

		// Import as new (spec): no history.
		c = newCollection(t)
		if sum := importFixture(t, c, name, false); sum.Reviews != 0 {
			t.Fatalf("%s: import as new summary = %+v", name, sum)
		}
		if n := len(storedReviews(t, c, 0)); n != 0 {
			t.Fatalf("%s: import as new: %d entries", name, n)
		}
	}
}

func TestSummaryWarnings(t *testing.T) {
	c := newCollection(t)
	sum := importFixture(t, c, "modern.apkg", true)
	// Summary shown (spec): notes added, audio counted with the media and
	// no warning about it.
	if sum.NotesAdded != 8 || sum.Media != 4 || len(sum.Warnings) != 0 {
		t.Fatalf("summary = %+v", sum)
	}
}

func TestUnsupportedMediaWarning(t *testing.T) {
	p, err := Read(fixture(t, "modern.apkg"), nil)
	if err != nil {
		t.Fatal(err)
	}
	p.Media["clip.mkv"] = []byte("x")
	p.Media["voice.amr"] = []byte("x")
	p.Media["clip.mp4"] = []byte("x")
	b, err := Convert(p, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := b.Media["clip.mp4"]; !ok || b.Media["clip.mkv"] != nil || b.Media["voice.amr"] != nil {
		t.Fatalf("media = %d files", len(b.Media))
	}
	if len(b.Warnings) != 1 || !strings.Contains(b.Warnings[0], "2 audio/video files were skipped") ||
		!strings.Contains(b.Warnings[0], "clip.mkv, voice.amr") {
		t.Fatalf("warnings = %v", b.Warnings)
	}
}

func TestCorruptPackage(t *testing.T) {
	c := newCollection(t)
	// Corrupt package (spec): an error explains the file could not be read,
	// and nothing is added.
	_, err := Import(c, fixture(t, "truncated.apkg"), Options{KeepScheduling: true}, nil, nil)
	if !errors.Is(err, ErrUnreadable) || !strings.Contains(err.Error(), "damaged") {
		t.Fatalf("err = %v", err)
	}
	decks, notes := collectionView(t, c)
	if len(decks) != 1 || len(notes) != 0 || c.HasMedia("cat.jpg") {
		t.Fatalf("collection changed: decks %v notes %d", decks, len(notes))
	}
	for _, bad := range [][]byte{nil, []byte("PK\x03\x04garbage"), []byte("not a zip")} {
		if _, err := Read(bad, nil); !errors.Is(err, ErrUnreadable) {
			t.Fatalf("Read(%q) err = %v", bad, err)
		}
	}
}

func TestNewerPackageVersionRejected(t *testing.T) {
	_, err := Read(rezip(t, fixture(t, "modern.apkg"), "meta", []byte{0x08, 0x09}), nil)
	if !errors.Is(err, ErrUnreadable) || !strings.Contains(err.Error(), "newer Anki version") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadPresets(t *testing.T) {
	m := loadManifest(t)
	round := func(xs []float64) []float64 {
		out := []float64{}
		for _, x := range xs {
			out = append(out, math.Round(x*1e4)/1e4)
		}
		return out
	}
	for _, name := range []string{"modern.apkg", "legacy.apkg"} {
		p, err := Read(fixture(t, name), nil)
		if err != nil {
			t.Fatal(err)
		}
		// Both formats give every deck the preset fixtures.json records.
		want := m.Packages[name].DeckPresets
		if len(want) != 3 || want["Fixture::Spanish"].Name != "Fixture Spanish" {
			t.Fatalf("%s: fixtures.json presets = %v", name, want)
		}
		for _, d := range p.Decks {
			w, ok := want[d.Name]
			if !ok {
				continue
			}
			pr, ok := p.Presets[d.PresetID]
			if !ok {
				t.Fatalf("%s: %s uses missing preset %d", name, d.Name, d.PresetID)
			}
			got := fixturePreset{Name: pr.Name, NewPerDay: pr.NewPerDay, ReviewsPerDay: pr.ReviewsPerDay,
				LearnSteps: round(pr.LearnSteps), RelearnSteps: round(pr.RelearnSteps),
				DesiredRetention: round([]float64{pr.DesiredRetention})[0], MaxIvl: pr.MaxInterval, MinInt: pr.MinLapseInterval,
				LeechThreshold: pr.LeechThreshold, LeechAction: pr.LeechAction, InsertOrder: pr.InsertOrder,
				GatherPriority: pr.GatherPriority, SortOrder: pr.SortOrder, NewMix: pr.NewMix, InterdayMix: pr.InterdayMix,
				ReviewOrder: pr.ReviewOrder, BuryNew: pr.BuryNew, BuryReviews: pr.BuryReviews,
				BuryInterdayLearning: pr.BuryInterdayLearning, DisableAutoplay: pr.DisableAutoplay,
				SkipQuestionWhenReplaying: pr.SkipQuestionWhenReplaying,
				SecondsToShowQuestion:     pr.SecondsToShowQuestion, SecondsToShowAnswer: pr.SecondsToShowAnswer,
				QuestionAction: pr.QuestionAction, AnswerAction: pr.AnswerAction}
			if !reflect.DeepEqual(got, w) {
				t.Errorf("%s: %s preset\n got %+v\nwant %+v", name, d.Name, got, w)
			}
			if (d.PresetID == AnkiDefaultPreset) != (d.Name != "Fixture::Spanish") {
				t.Errorf("%s: %s uses preset %d", name, d.Name, d.PresetID)
			}
		}
	}
}

// spanishOptions are the options of the fixture's preset Fixture Spanish
// in Cardflip's terms.
func spanishOptions() core.DeckOptions {
	o := core.DefaultDeckOptions()
	o.NewPerDay, o.ReviewsPerDay, o.DesiredRetention = 10, 150, 0.85
	o.LearningSteps, o.RelearningSteps = []int{120, 900}, []int{300}
	o.MaxInterval, o.MinRelearnInterval = 3650, 2
	o.LeechThreshold, o.LeechAction = 6, core.LeechSuspend
	o.BuryNew, o.BuryReview, o.BuryInterdayLearning = true, false, true
	o.DisableAutoplay, o.SkipQuestionOnReplay = true, true
	o.NewReviewOrder, o.InterdayLearningOrder = core.NewBeforeReviews, core.NewAfterReviews
	o.ReviewSortOrder = core.ReviewIntervalsAscending
	o.NewSortOrder, o.NewGatherOrder = core.SortRandomNoteThenType, core.GatherDeckRandomNotes
	o.SecondsToShowQuestion, o.SecondsToShowAnswer = 5, 3
	o.QuestionAction, o.AnswerAction = core.QuestionShowReminder, core.AnswerGood
	return o
}

// deckPresets returns deck name -> preset name, and the presets by name.
func deckPresets(c *core.Collection) (map[string]string, map[string]core.Preset) {
	byID, byName := map[core.ID]core.Preset{}, map[string]core.Preset{}
	for _, p := range c.Presets() {
		byID[p.ID], byName[p.Name] = p, p
	}
	decks := map[string]string{}
	for _, d := range c.Decks() {
		decks[d.Name] = byID[d.PresetID].Name
	}
	return decks, byName
}

func TestImportPresets(t *testing.T) {
	for _, name := range []string{"modern.apkg", "legacy.apkg"} {
		c := newCollection(t)
		importFixture(t, c, name, true)
		// Shared deck settings (spec): the new deck uses an imported preset
		// with the package's options, 10 new cards per day among them.
		decks, presets := deckPresets(c)
		if decks["Fixture::Spanish"] != "Fixture Spanish" {
			t.Fatalf("%s: deck presets = %v", name, decks)
		}
		if got := presets["Fixture Spanish"].Options; !got.Equal(spanishOptions()) {
			t.Errorf("%s: options\n got %+v\nwant %+v", name, got, spanishOptions())
		}
		// Anki's default preset (spec): the other decks use Default, and no
		// preset is added for it.
		if decks["Fixture"] != "Default" || decks["Fixture::Geo"] != "Default" || len(presets) != 2 {
			t.Errorf("%s: deck presets = %v, presets %v", name, decks, presets)
		}
		if !presets["Default"].Options.Equal(core.DefaultDeckOptions()) {
			t.Errorf("%s: Default changed: %+v", name, presets["Default"].Options)
		}
		// Undo removes the imported preset with the decks.
		if _, err := c.Undo(); err != nil {
			t.Fatal(err)
		}
		if _, presets := deckPresets(c); len(presets) != 1 {
			t.Errorf("%s: presets after undo = %v", name, presets)
		}
	}
}

func TestImportPresetsReusedAndRenamed(t *testing.T) {
	c := newCollection(t)
	importFixture(t, c, "modern.apkg", true)
	_, presets := deckPresets(c)
	spanish := presets["Fixture Spanish"].ID
	// Reused: a preset with the same name and options serves the decks a
	// later import creates.
	for _, d := range c.Decks() {
		if d.Name == "Fixture" {
			if err := c.DeleteDeck(d.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	importFixture(t, c, "legacy.apkg", true)
	decks, presets := deckPresets(c)
	if decks["Fixture::Spanish"] != "Fixture Spanish" || len(presets) != 2 || presets["Fixture Spanish"].ID != spanish {
		t.Fatalf("reimport: decks %v, presets %v", decks, presets)
	}
	// Existing decks keep their preset.
	var deck core.ID
	for _, d := range c.Decks() {
		if d.Name == "Fixture::Spanish" {
			deck = d.ID
		}
	}
	if err := c.SetDeckPreset(deck, core.DefaultPresetID); err != nil {
		t.Fatal(err)
	}
	importFixture(t, c, "modern.apkg", true)
	if decks, _ := deckPresets(c); decks["Fixture::Spanish"] != "Default" {
		t.Fatalf("existing deck: presets %v", decks)
	}

	// Renamed: a preset with the name but other options stays as it is,
	// and the package's preset is imported under a numbered name.
	c = newCollection(t)
	d, _ := c.CreateDeck("Mine")
	mine, err := c.AddPreset(d, "fixture spanish")
	if err != nil {
		t.Fatal(err)
	}
	importFixture(t, c, "modern.apkg", true)
	decks, presets = deckPresets(c)
	if decks["Fixture::Spanish"] != "Fixture Spanish (2)" || decks["Mine"] != "fixture spanish" {
		t.Fatalf("renamed: decks %v", decks)
	}
	if !presets["Fixture Spanish (2)"].Options.Equal(spanishOptions()) || !presets["fixture spanish"].Options.Equal(core.DefaultDeckOptions()) ||
		presets["fixture spanish"].ID != mine {
		t.Fatalf("renamed: presets %v", presets)
	}
	// Importing it once more reuses the numbered preset.
	for _, d := range c.Decks() {
		if d.Name == "Fixture" {
			c.DeleteDeck(d.ID)
		}
	}
	importFixture(t, c, "legacy.apkg", true)
	if decks, presets := deckPresets(c); decks["Fixture::Spanish"] != "Fixture Spanish (2)" || len(presets) != 3 {
		t.Fatalf("numbered reimport: decks %v, presets %d", decks, len(presets))
	}
}

func TestPresetOptionsClamped(t *testing.T) {
	// Values outside Cardflip's ranges are clamped; unmapped enums keep
	// the defaults.
	o := presetOptions(Preset{NewPerDay: 20000, ReviewsPerDay: -1, DesiredRetention: 0.5, MaxInterval: 99999,
		MinLapseInterval: 0, LeechThreshold: 500, LeechAction: 1, LearnSteps: []float64{0.5, 0, 1e9},
		ReviewOrder: 8, GatherPriority: 99})
	if o.NewPerDay != 9999 || o.ReviewsPerDay != 0 || o.DesiredRetention != core.MinRetention ||
		o.MaxInterval != core.MaxIntervalDays || o.MinRelearnInterval != 1 || o.LeechThreshold != core.MaxLeechThreshold ||
		o.LeechAction != core.LeechTag || !reflect.DeepEqual(o.LearningSteps, []int{30}) || len(o.RelearningSteps) != 0 ||
		o.ReviewSortOrder != core.ReviewDueRandom || o.NewGatherOrder != core.GatherAscending {
		t.Fatalf("options = %+v", o)
	}
	if err := core.ValidateDeckOptions(o); err != nil {
		t.Fatal(err)
	}
	// Auto advance times are whole seconds: a fraction rounds, but a short
	// time stays on; unknown actions keep the defaults.
	for _, c := range []struct {
		in   float64
		want int
	}{{0, 0}, {-2, 0}, {0.2, 1}, {2.5, 3}, {2.4, 2}, {12345, core.MaxAutoAdvanceSeconds}} {
		o := presetOptions(Preset{SecondsToShowQuestion: c.in, SecondsToShowAnswer: c.in, QuestionAction: 7, AnswerAction: 9})
		if o.SecondsToShowQuestion != c.want || o.SecondsToShowAnswer != c.want ||
			o.QuestionAction != core.QuestionShowAnswer || o.AnswerAction != core.AnswerBuryCard {
			t.Errorf("%v seconds: options = %+v", c.in, o)
		}
		if err := core.ValidateDeckOptions(o); err != nil {
			t.Error(err)
		}
	}
}

// colpkgFixtureDecks are the decks of the collection package fixtures, in
// Cardflip: every deck of the Anki collection apart from the filtered one.
const colpkgFixtureDecks = "Default, Fixture, Fixture::Geo, Fixture::Spanish, Inbox, Inbox::Later"

func TestReadCollectionPackage(t *testing.T) {
	m := loadManifest(t)
	for _, name := range []string{"modern.colpkg", "legacy.colpkg"} {
		want := m.Packages[name]
		p, err := Read(fixture(t, name), nil)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if p.Format != want.Format {
			t.Errorf("%s: format %s, want %s", name, p.Format, want.Format)
		}
		var decks, filtered []string
		for _, d := range p.Decks {
			decks = append(decks, d.Name)
			if d.Filtered {
				filtered = append(filtered, d.Name)
			}
		}
		sort.Strings(decks)
		if !reflect.DeepEqual(decks, want.Decks) || !reflect.DeepEqual(filtered, want.FilteredDecks) || len(filtered) != 1 {
			t.Errorf("%s: decks %v (filtered %v), want %v (filtered %v)", name, decks, filtered, want.Decks, want.FilteredDecks)
		}
		var types []string
		for _, nt := range p.NoteTypes {
			types = append(types, nt.Name)
		}
		sort.Strings(types)
		if len(types) != len(want.Notetypes) || len(types) <= 3 {
			t.Errorf("%s: note types %v, want %d", name, types, len(want.Notetypes))
		}
		if len(p.Notes) != 8 || len(p.Cards) != 12 || len(p.Media) != 4 {
			t.Errorf("%s: %d notes, %d cards, %d media", name, len(p.Notes), len(p.Cards), len(p.Media))
		}
		// The card in the filtered deck knows its home deck.
		keys := cardKeys(p)
		for _, c := range p.Cards {
			inFiltered := p.Decks[c.DeckID].Filtered
			if inFiltered != (keys[c.ID] == "cf-basic-dog/0") {
				t.Errorf("%s: %s in deck %q", name, keys[c.ID], p.Decks[c.DeckID].Name)
			}
			if inFiltered && p.Decks[c.ODeckID].Name != "Fixture::Spanish" {
				t.Errorf("%s: home deck %q", name, p.Decks[c.ODeckID].Name)
			}
		}
	}
}

func TestConvertCollection(t *testing.T) {
	p, err := Read(fixture(t, "modern.colpkg"), nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ConvertCollection(p)
	if err != nil {
		t.Fatal(err)
	}
	// Every normal deck, the empty ones included; not the filtered deck.
	if got := strings.Join(b.Decks, ", "); got != colpkgFixtureDecks {
		t.Fatalf("decks = %s", got)
	}
	// Every note type, used or not.
	if len(b.NoteTypes) != len(p.NoteTypes) {
		t.Fatalf("%d note types, want %d", len(b.NoteTypes), len(p.NoteTypes))
	}
	// Cards of filtered decks go to their home decks, with scheduling.
	for _, n := range b.Notes {
		if n.GUID == "cf-basic-dog" && n.Cards[0].DeckName != "Fixture::Spanish" {
			t.Fatalf("filtered card in %q", n.Cards[0].DeckName)
		}
		if n.GUID == "cf-basic-hola" && n.Cards[0].State.State == scheduler.New {
			t.Fatal("scheduling not kept")
		}
	}
}

// Move from Anki (spec): the collection package replaces the collection;
// the settings are kept.
func TestReplaceWithCollectionPackage(t *testing.T) {
	views := map[string][]any{}
	for _, name := range []string{"modern.colpkg", "legacy.colpkg"} {
		c := newCollection(t)
		st := c.Settings()
		st.Theme, st.RolloverHour = "dark", 6
		if err := c.UpdateSettings(st); err != nil {
			t.Fatal(err)
		}
		types := c.NoteTypes()
		if _, err := c.AddNote(core.NoteInput{TypeID: types[0].ID, DeckID: c.Decks()[0].ID, Fields: []string{"replaced", "note"}}); err != nil {
			t.Fatal(err)
		}
		b, err := ReadCollection(fixture(t, name), nil)
		if err != nil {
			t.Fatal(err)
		}
		sum, err := c.ReplaceWithImport(b, nil)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if sum.NotesAdded != 8 || sum.CardsAdded != 12 || sum.Reviews != 24 || sum.Media != 4 {
			t.Errorf("%s: summary %+v", name, sum)
		}
		decks, notes := collectionView(t, c)
		if got := strings.Join(decks, ", "); got != colpkgFixtureDecks {
			t.Errorf("%s: decks = %s", name, got)
		}
		if len(notes) != 8 {
			t.Errorf("%s: %d notes, the old one should be gone", name, len(notes))
		}
		var typeNames []string
		for _, nt := range c.NoteTypes() {
			typeNames = append(typeNames, nt.Name)
		}
		// Anki's note types only: no copies of Cardflip's built-in ones.
		if len(typeNames) != len(loadManifest(t).Packages[name].Notetypes) || slices.Contains(typeNames, "Basic (2)") {
			t.Errorf("%s: note types %v", name, typeNames)
		}
		if got := c.Settings(); got.Theme != "dark" || got.RolloverHour != 6 || got.LastDeckID != 0 {
			t.Errorf("%s: settings %+v", name, got)
		}
		if !c.HasMedia("dog.png") {
			t.Errorf("%s: media missing", name)
		}
		presets, _ := deckPresets(c)
		if presets["Fixture::Spanish"] != "Fixture Spanish" || presets["Inbox::Later"] != core.DefaultPresetName {
			t.Errorf("%s: presets %v", name, presets)
		}
		views[name] = []any{decks, notes}
	}
	if !reflect.DeepEqual(views["modern.colpkg"], views["legacy.colpkg"]) {
		t.Fatal("legacy collection package differs from the modern one")
	}
}

func TestCorruptCollectionPackage(t *testing.T) {
	if _, err := ReadCollection(fixture(t, "truncated.colpkg"), nil); !errors.Is(err, ErrUnreadable) {
		t.Fatalf("err = %v", err)
	}
}

// Modification times of notes and note types are read from both formats,
// and Convert passes them on.
func TestReadModificationTimes(t *testing.T) {
	m := loadManifest(t)
	if len(m.Updates) != 4 {
		t.Fatalf("%d update packages in fixtures.json", len(m.Updates))
	}
	for name, want := range m.Updates {
		p, err := Read(fixture(t, name), nil)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(p.NoteTypes) != 1 || len(want.Notetypes) != 1 {
			t.Fatalf("%s: note types %+v", name, p.NoteTypes)
		}
		for _, nt := range p.NoteTypes {
			if nt.Mod != want.Notetypes[0].Mod || !slices.Equal(nt.Fields, want.Notetypes[0].Fields) {
				t.Errorf("%s: note type mod %d fields %q, want %d %q", name, nt.Mod, nt.Fields,
					want.Notetypes[0].Mod, want.Notetypes[0].Fields)
			}
		}
		if len(p.Notes) != len(want.Notes) {
			t.Fatalf("%s: %d notes", name, len(p.Notes))
		}
		for _, n := range p.Notes {
			if w := want.Notes[n.GUID]; n.Mod != w.Mod || !slices.Equal(n.Fields, w.Fields) {
				t.Errorf("%s: note %s mod %d fields %q, want %d %q", name, n.GUID, n.Mod, n.Fields, w.Mod, w.Fields)
			}
		}
		b, err := Convert(p, false)
		if err != nil {
			t.Fatal(err)
		}
		if got := b.NoteTypes[0].Type.Modified.Unix(); got != want.Notetypes[0].Mod {
			t.Errorf("%s: converted note type modified %d", name, got)
		}
		for _, n := range b.Notes {
			if got := n.Modified.Unix(); got != want.Notes[n.GUID].Mod {
				t.Errorf("%s: converted note %s modified %d", name, n.GUID, got)
			}
		}
	}
	if !unixTime(0).IsZero() {
		t.Fatal("unknown time is not zero")
	}
}

func importUpdate(t *testing.T, c *core.Collection, name string, opts Options) core.ImportSummary {
	t.Helper()
	sum, err := Import(c, fixture(t, name), opts, nil, nil)
	if err != nil {
		t.Fatalf("import %s: %v", name, err)
	}
	return sum
}

func noteTypeNamed(t *testing.T, c *core.Collection, name string) core.NoteType {
	t.Helper()
	for _, nt := range c.NoteTypes() {
		if nt.Name == name {
			return nt
		}
	}
	t.Fatalf("no note type %q", name)
	return core.NoteType{}
}

// Re-importing a package (data-import spec), with the update fixtures:
// import twice, local edit kept, updated deck.
func TestReimportUpdatePolicy(t *testing.T) {
	m := loadManifest(t)
	for _, first := range []string{"update-v1.apkg", "update-v1-legacy.apkg"} {
		// Import twice: each note once, nothing changes.
		c := newCollection(t)
		builtins := len(c.NoteTypes())
		importUpdate(t, c, first, Options{})
		for _, again := range []string{"update-v1.apkg", "update-v1-legacy.apkg"} {
			sum := importUpdate(t, c, again, Options{})
			if sum.NotesAdded != 0 || sum.NotesSkipped != 2 || sum.NotesUpdated != 0 || sum.NoteTypesAdded+sum.NoteTypesUpdated != 0 {
				t.Fatalf("%s then %s: %+v", first, again, sum)
			}
		}
		if ids, _ := c.Search("", core.ColCreated, false); len(ids) != 2 || len(c.NoteTypes()) != builtins+1 {
			t.Fatalf("%s twice: %d cards, %d note types", first, len(ids), len(c.NoteTypes()))
		}

		// Local edit kept: the user's edit (now, after both versions)
		// survives version 1 and version 2.
		hola, cards := noteByGUID(t, c, "cf-up-hola")
		if want := m.Updates["update-v1.apkg"].Notes["cf-up-hola"].Mod; hola.Modified.Unix() != want {
			t.Fatalf("imported note modified %v, want %d", hola.Modified, want)
		}
		if err := c.UpdateNote(hola.ID, []string{"hola", "hello (mine)"}, nil); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"update-v1.apkg", "update-v2.apkg"} {
			sum := importUpdate(t, c, name, Options{})
			if sum.NotesSkippedOlder != 1 || sum.NotesUpdated != 0 {
				t.Fatalf("%s over a local edit: %+v", name, sum)
			}
			if n, _ := noteByGUID(t, c, "cf-up-hola"); n.Fields[1] != "hello (mine)" {
				t.Fatalf("%s: local edit lost: %q", name, n.Fields)
			}
		}

		// Updated deck: in a collection without the edit, version 2 brings
		// the fix and the new styling; scheduling stays.
		c = newCollection(t)
		importUpdate(t, c, first, Options{KeepScheduling: true})
		_, cards = noteByGUID(t, c, "cf-up-hola")
		sum := importUpdate(t, c, "update-v2.apkg", Options{})
		if sum.NotesUpdated != 1 || sum.NotesSkipped != 1 || sum.NoteTypesUpdated != 1 || sum.NoteTypesAdded != 0 {
			t.Fatalf("%s then v2: %+v", first, sum)
		}
		n, after := noteByGUID(t, c, "cf-up-hola")
		if n.Fields[1] != "hello" || !reflect.DeepEqual(cards, after) {
			t.Fatalf("%s then v2: note %q, cards changed %v", first, n.Fields, !reflect.DeepEqual(cards, after))
		}
		nt := noteTypeNamed(t, c, "Update Vocab")
		if !strings.Contains(nt.CSS, "28px") || nt.Modified.Unix() != m.Updates["update-v2.apkg"].Notetypes[0].Mod {
			t.Fatalf("note type after v2: css %q modified %v", nt.CSS, nt.Modified)
		}
		// Version 1 again changes nothing back.
		if sum := importUpdate(t, c, first, Options{}); sum.NotesSkippedOlder != 1 || sum.NoteTypesUpdated != 0 {
			t.Fatalf("v1 after v2: %+v", sum)
		}
		if nt := noteTypeNamed(t, c, "Update Vocab"); !strings.Contains(nt.CSS, "28px") {
			t.Fatalf("v1 after v2 restyled the note type: %q", nt.CSS)
		}
		// "Always" takes version 1 back; "never" keeps whatever is there.
		sum = importUpdate(t, c, first, Options{UpdateNotes: core.UpdateAlways, UpdateNoteTypes: core.UpdateAlways})
		if n, _ := noteByGUID(t, c, "cf-up-hola"); sum.NotesUpdated != 1 || sum.NoteTypesUpdated != 1 || n.Fields[1] != "helo" {
			t.Fatalf("always v1: %+v, note %q", sum, n.Fields)
		}
		sum = importUpdate(t, c, "update-v2.apkg", Options{UpdateNotes: core.UpdateNever, UpdateNoteTypes: core.UpdateNever})
		if n, _ := noteByGUID(t, c, "cf-up-hola"); sum.NotesUpdated != 0 || sum.NoteTypesUpdated != 0 || n.Fields[1] != "helo" {
			t.Fatalf("never v2: %+v, note %q", sum, n.Fields)
		}
	}
}

// Note type updates on import (data-import spec), with the update fixtures:
// version 3 adds the field Audio.
func TestReimportNoteTypeChange(t *testing.T) {
	// Merge adds a field.
	c := newCollection(t)
	importUpdate(t, c, "update-v1.apkg", Options{})
	before := noteTypeNamed(t, c, "Update Vocab")
	sum := importUpdate(t, c, "update-v3.apkg", Options{MergeNoteTypes: true})
	// cf-up-gato only gains the empty field: unchanged.
	if sum.NoteTypesMerged != 1 || sum.NoteTypesAdded != 0 || sum.NotesUpdated != 1 || sum.NotesSkipped != 1 || sum.NotesAdded != 1 {
		t.Fatalf("merge: %+v", sum)
	}
	nt := noteTypeNamed(t, c, "Update Vocab")
	if nt.ID != before.ID || strings.Join(nt.Fields, ",") != "Front,Back,Audio" || !strings.Contains(nt.Templates[0].AFmt, "{{Audio}}") {
		t.Fatalf("merged note type: %+v", nt)
	}
	for guid, want := range map[string]string{"cf-up-hola": "hola|hello|hola-audio", "cf-up-gato": "gato|cat|", "cf-up-perro": "perro|dog|perro-audio"} {
		if n, _ := noteByGUID(t, c, guid); n.TypeID != nt.ID || strings.Join(n.Fields, "|") != want {
			t.Fatalf("merge: %s = %q (type %d)", guid, n.Fields, n.TypeID)
		}
	}

	// No merge: the package's notes get a new note type, the existing one
	// and its notes stay as they were.
	c = newCollection(t)
	importUpdate(t, c, "update-v1.apkg", Options{})
	before = noteTypeNamed(t, c, "Update Vocab")
	sum = importUpdate(t, c, "update-v3.apkg", Options{})
	if sum.NoteTypesAdded != 1 || sum.NoteTypesMerged != 0 || sum.NotesUpdated != 0 || sum.NotesAdded != 1 {
		t.Fatalf("no merge: %+v", sum)
	}
	if nt := noteTypeNamed(t, c, "Update Vocab"); !reflect.DeepEqual(nt, before) {
		t.Fatalf("existing note type changed: %+v", nt)
	}
	added := noteTypeNamed(t, c, "Update Vocab (2)")
	if strings.Join(added.Fields, ",") != "Front,Back,Audio" {
		t.Fatalf("new note type: %+v", added)
	}
	if n, _ := noteByGUID(t, c, "cf-up-perro"); n.TypeID != added.ID {
		t.Fatalf("cf-up-perro uses note type %d, want %d", n.TypeID, added.ID)
	}
	if n, _ := noteByGUID(t, c, "cf-up-hola"); n.TypeID != before.ID || strings.Join(n.Fields, "|") != "hola|helo" {
		t.Fatalf("cf-up-hola = %q", n.Fields)
	}
}

// hiraganaPackage is shaped like the shared deck Japanese_Hiragana.apkg:
// its subdecks are filtered decks whose cards have no home deck. Cram is
// an ordinary filtered deck whose card has a home deck.
func hiraganaPackage() *Package {
	p := &Package{
		Format: "anki2", Crt: 1767240000,
		Decks: map[int64]Deck{
			1:  {Name: "Default", PresetID: AnkiDefaultPreset},
			10: {Name: "Kana (Hiragana)", PresetID: AnkiDefaultPreset},
			11: {Name: "Kana (Hiragana)::Basic", Filtered: true, PresetID: 5},
			12: {Name: "Kana (Hiragana)::Advanced", Filtered: true, PresetID: 5},
			13: {Name: "Cram", Filtered: true, PresetID: 5},
		},
		NoteTypes: map[int64]NoteType{
			100: {Name: "Kana", Fields: []string{"Front", "Back"},
				Templates: []Template{{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Back}}"}}},
		},
	}
	for i, home := range []struct{ deck, odeck int64 }{{10, 0}, {11, 0}, {11, 0}, {12, 0}, {12, 0}, {13, 10}} {
		id := int64(i + 1)
		p.Notes = append(p.Notes, Note{ID: id, GUID: "kana-" + strconv.FormatInt(id, 10), TypeID: 100,
			Fields: []string{"front " + strconv.FormatInt(id, 10), "back"}})
		c := Card{ID: 1000 + id, NoteID: id, DeckID: home.deck, Due: id}
		if home.odeck != 0 {
			c.ODeckID, c.ODue = home.odeck, id
		}
		p.Cards = append(p.Cards, c)
	}
	return p
}

var hiraganaDecks = map[string]int{"Kana (Hiragana)": 2, "Kana (Hiragana)::Basic": 2, "Kana (Hiragana)::Advanced": 2}

// Cards in filtered decks without a home deck (data-import spec).
func TestConvertFilteredDeckWithoutHome(t *testing.T) {
	for _, keep := range []bool{true, false} {
		b, err := Convert(hiraganaPackage(), keep)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]int{}
		for _, n := range b.Notes {
			for _, c := range n.Cards {
				got[c.DeckName]++
			}
		}
		if !reflect.DeepEqual(got, hiraganaDecks) {
			t.Errorf("keep %v: cards per deck %v, want %v", keep, got, hiraganaDecks)
		}
	}
}

// Shared deck with filtered subdecks (data-import spec): imported into an
// empty collection, as Import does after reading the package.
func TestImportFilteredDeckWithoutHome(t *testing.T) {
	for _, keep := range []bool{true, false} {
		c := newCollection(t)
		b, err := Convert(hiraganaPackage(), keep)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.Import(b, nil); err != nil {
			t.Fatal(err)
		}
		var decks []string
		for _, d := range c.Decks() {
			decks = append(decks, d.Name)
		}
		sort.Strings(decks)
		if want := []string{"Default", "Kana (Hiragana)", "Kana (Hiragana)::Advanced", "Kana (Hiragana)::Basic"}; !reflect.DeepEqual(decks, want) {
			t.Errorf("keep %v: decks %v, want %v", keep, decks, want)
		}
		got := map[string]int{}
		ids, _ := c.Search("", core.ColCreated, false)
		for _, row := range c.Rows(ids) {
			got[row.Deck]++
		}
		if !reflect.DeepEqual(got, hiraganaDecks) {
			t.Errorf("keep %v: cards per deck %v, want %v", keep, got, hiraganaDecks)
		}
		// A filtered deck's preset is not mapped: the new decks use Default.
		presets, _ := deckPresets(c)
		for name := range hiraganaDecks {
			if presets[name] != "Default" {
				t.Errorf("keep %v: %s uses preset %q", keep, name, presets[name])
			}
		}
	}
}

// Deck descriptions are read from both formats: Deck.Normal.description in
// anki21b, "desc" in the legacy deck JSON.
func TestReadDeckDescriptions(t *testing.T) {
	var normal []byte
	normal = protowire.AppendTag(normal, 1, protowire.VarintType)
	normal = protowire.AppendVarint(normal, 7)
	normal = protowire.AppendTag(normal, 4, protowire.BytesType)
	normal = protowire.AppendString(normal, "Words from <b>Book 1</b>")
	var kind []byte
	kind = protowire.AppendTag(kind, 1, protowire.BytesType)
	kind = protowire.AppendBytes(kind, normal)
	if d := modernDeck("Shared\x1fSpanish", kind); d != (Deck{Name: "Shared::Spanish", PresetID: 7, Description: "Words from <b>Book 1</b>"}) {
		t.Errorf("modern deck = %+v", d)
	}
	p := &Package{Decks: map[int64]Deck{}, NoteTypes: map[int64]NoteType{}}
	if err := readLegacyTypes("{}", `{"5": {"name": "Shared", "dyn": 0, "conf": 1, "desc": "Read <i>me</i>"}}`, p); err != nil {
		t.Fatal(err)
	}
	if d := p.Decks[5]; d != (Deck{Name: "Shared", PresetID: 1, Description: "Read <i>me</i>"}) {
		t.Errorf("legacy deck = %+v", d)
	}
}

// Shared deck description (data-import spec): a deck the import creates
// gets the package's description; a filtered deck's is ignored.
func TestImportDeckDescription(t *testing.T) {
	p := hiraganaPackage()
	d := p.Decks[10]
	d.Description = "All 46 <b>hiragana</b>"
	p.Decks[10] = d
	d = p.Decks[11]
	d.Description = "filtered"
	p.Decks[11] = d
	b, err := Convert(p, true)
	if err != nil {
		t.Fatal(err)
	}
	c := newCollection(t)
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	for _, d := range c.Decks() {
		want := ""
		if d.Name == "Kana (Hiragana)" {
			want = "All 46 <b>hiragana</b>"
		}
		if d.Description != want {
			t.Errorf("%s: description %q, want %q", d.Name, d.Description, want)
		}
	}
}

// Anki's original position ("pos" in card data) is imported for studied
// cards, among the imported new cards, so that a reset can restore it.
func TestImportOriginalPosition(t *testing.T) {
	p := &Package{
		Format: "anki2", Crt: 1767240000,
		Decks: map[int64]Deck{1: {Name: "Default", PresetID: AnkiDefaultPreset}},
		NoteTypes: map[int64]NoteType{
			100: {Name: "Kana", Fields: []string{"Front", "Back"},
				Templates: []Template{{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Back}}"}}},
		},
	}
	cards := []Card{
		{Type: 0, Queue: 0, Due: 1},
		{Type: 2, Queue: 2, Due: 300, Ivl: 10, Factor: 2500, Reps: 3, Data: `{"pos":2,"s":10,"d":5}`},
		{Type: 0, Queue: 0, Due: 3},
		{Type: 2, Queue: 2, Due: 300, Ivl: 10, Factor: 2500, Reps: 3},
	}
	for i, card := range cards {
		id := int64(i + 1)
		p.Notes = append(p.Notes, Note{ID: id, GUID: "n" + strconv.FormatInt(id, 10), TypeID: 100,
			Fields: []string{"front " + strconv.FormatInt(id, 10), "back"}})
		card.ID, card.NoteID, card.DeckID = 1000+id, id, 1
		p.Cards = append(p.Cards, card)
	}
	c := newCollection(t)
	b, err := Convert(p, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	var got []core.Card
	for i := range cards {
		_, cs := noteByGUID(t, c, "n"+strconv.Itoa(i+1))
		got = append(got, cs[0])
	}
	studied := got[1]
	if studied.OriginalPos == nil || *studied.OriginalPos <= got[0].NewPos || *studied.OriginalPos >= got[2].NewPos {
		t.Fatalf("original position %v, new cards at %d and %d", studied.OriginalPos, got[0].NewPos, got[2].NewPos)
	}
	if got[3].OriginalPos != nil {
		t.Fatalf("card without pos: original position %d", *got[3].OriginalPos)
	}
	if err := c.ResetCards([]core.ID{studied.ID}, core.ResetOptions{RestorePosition: true}); err != nil {
		t.Fatal(err)
	}
	q, err := c.Queue(c.Decks()[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := []core.ID{got[0].ID, studied.ID, got[2].ID}; !slices.Equal(q.New, want) {
		t.Fatalf("new queue %v, want %v", q.New, want)
	}
}

// LaTeX image format on Anki import (data-import spec), with the LaTeX
// fixtures in both formats: each note type keeps PNG or SVG, the images,
// SVG ones included, are imported, and the cards show them. The note whose
// images the deck lacks shows its LaTeX source.
func TestImportLatexImages(t *testing.T) {
	m := loadManifest(t)
	if len(m.Latex.Packages) != 2 || len(m.Latex.Names) != 3 {
		t.Fatalf("fixtures.json latex = %+v", m.Latex)
	}
	for name, want := range m.Latex.Packages {
		t.Run(name, func(t *testing.T) {
			p, err := Read(fixture(t, name), nil)
			if err != nil {
				t.Fatal(err)
			}
			svgTypes := map[string]bool{}
			for _, nt := range p.NoteTypes {
				svgTypes[nt.Name] = nt.LatexSVG
			}
			for _, nt := range want.Notetypes {
				if got, ok := svgTypes[nt.Name]; !ok || got != nt.LatexSVG {
					t.Errorf("read note type %q: LatexSVG = %v, want %v", nt.Name, got, nt.LatexSVG)
				}
			}

			c := newCollection(t)
			importFixture(t, c, name, false)
			for _, nt := range want.Notetypes {
				if got := noteTypeNamed(t, c, nt.Name).LatexSVG; got != nt.LatexSVG {
					t.Errorf("imported note type %q: LatexSVG = %v, want %v", nt.Name, got, nt.LatexSVG)
				}
			}
			svgs := 0
			for file := range want.Media {
				if !c.HasMedia(file) {
					t.Errorf("media %s not imported", file)
				}
				if strings.HasSuffix(file, ".svg") {
					svgs++
				}
			}
			if len(want.Media) != 5 || svgs != 2 {
				t.Fatalf("fixture media = %v", want.Media)
			}

			for guid, fields := range m.Latex.Names {
				_, cards := noteByGUID(t, c, guid)
				r, err := c.RenderCard(cards[0].ID, render.Env{})
				if err != nil {
					t.Fatal(err)
				}
				for _, files := range fields {
					for _, file := range files {
						shown := strings.Contains(r.Answer, `src="`+file+`"`)
						if guid == m.Latex.Missing {
							if shown || c.HasMedia(file) {
								t.Errorf("%s: %s shown, but the deck has no such image", guid, file)
							}
						} else if !shown {
							t.Errorf("%s: answer does not show %s: %s", guid, file, r.Answer)
						}
					}
				}
				if guid == m.Latex.Missing && !strings.Contains(r.Question, `<pre class="latex">&#92;LaTeX</pre>`) {
					t.Errorf("%s: question = %s, want the LaTeX source", guid, r.Question)
				}
				if strings.Contains(r.Answer, "[latex]") || strings.Contains(r.Answer, "[$]") || strings.Contains(r.Answer, "[$$]") {
					t.Errorf("%s: answer keeps a LaTeX block: %s", guid, r.Answer)
				}
			}
		})
	}
}

// Imported occlusion deck (data-import spec): Anki's Image Occlusion notes
// keep their note type and image, and every card shows its shapes.
func TestImportImageOcclusion(t *testing.T) {
	m := loadManifest(t)
	if len(m.Occlusion) != 2 {
		t.Fatalf("fixtures.json occlusion = %+v", m.Occlusion)
	}
	for name, want := range m.Occlusion {
		t.Run(name, func(t *testing.T) {
			if len(want.Notetypes) != 1 || len(want.Notes) != 2 || len(want.Cards) != 5 || len(want.Media) != 1 {
				t.Fatalf("fixture = %+v", want)
			}
			c := newCollection(t)
			sum := importFixture(t, c, name, false)
			if sum.NotesAdded != 2 || sum.CardsAdded != 5 || sum.NoteTypesAdded != 1 {
				t.Errorf("summary = %+v", sum)
			}
			for file := range want.Media {
				if !c.HasMedia(file) {
					t.Errorf("media %s not imported", file)
				}
			}
			cardsOf := map[string]int{}
			for _, card := range want.Cards {
				cardsOf[card.Note]++
			}
			for guid, wn := range want.Notes {
				n, cards := noteByGUID(t, c, guid)
				nt, err := c.NoteType(n.TypeID)
				if err != nil {
					t.Fatal(err)
				}
				if !nt.IsImageOcclusion() || !slices.Equal(nt.Fields, want.Notetypes[0].Fields) || !slices.Equal(n.Fields, wn.Fields) {
					t.Errorf("%s: note type %q %v, fields %q", guid, nt.Name, nt.Fields, n.Fields)
				}
				if len(cards) != cardsOf[guid] {
					t.Fatalf("%s: %d cards, want %d", guid, len(cards), cardsOf[guid])
				}
				for _, card := range cards {
					r, err := c.RenderCard(card.ID, render.Env{})
					if err != nil {
						t.Fatal(err)
					}
					active := `active="` + strconv.Itoa(card.Ord+1) + `"`
					for side, html := range map[string]string{"q": r.Question, "a": r.Answer} {
						if !strings.Contains(html, `<cf-occlusion image="diagram.png"`) || !strings.Contains(html, active+` side="`+side+`"`) ||
							strings.Contains(html, "<script") || strings.Contains(html, "<img") {
							t.Errorf("%s card %d %s = %s", guid, card.Ord+1, side, html)
						}
					}
				}
			}
		})
	}
}

// Card type options are read from both package formats.
func TestReadCardTypeOptions(t *testing.T) {
	var cfg []byte
	for _, f := range []struct {
		num protowire.Number
		s   string
	}{{1, "{{Front}}"}, {2, "{{Back}}"}, {3, "{{Word}}"}, {4, "{{Meaning}}"}} {
		cfg = protowire.AppendTag(cfg, f.num, protowire.BytesType)
		cfg = protowire.AppendString(cfg, f.s)
	}
	cfg = protowire.AppendTag(cfg, 5, protowire.VarintType)
	cfg = protowire.AppendVarint(cfg, 1700000000123)
	want := Template{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Back}}", BrowserQFmt: "{{Word}}", BrowserAFmt: "{{Meaning}}", DeckID: 1700000000123}
	if got := modernTemplate("Card 1", cfg); got != want {
		t.Errorf("modern template = %+v", got)
	}

	p := &Package{Decks: map[int64]Deck{}, NoteTypes: map[int64]NoteType{}}
	models := `{"7": {"name": "Vocab", "type": 0, "flds": [{"name": "Front", "ord": 0}], "tmpls": [
		{"name": "Card 1", "ord": 0, "qfmt": "{{Front}}", "afmt": "{{Back}}", "bqfmt": "{{Word}}", "bafmt": "{{Meaning}}", "did": 1700000000123},
		{"name": "Card 2", "ord": 1, "qfmt": "{{Back}}", "afmt": "{{Front}}", "bqfmt": "", "bafmt": "", "did": null}]}}`
	if err := readLegacyTypes(models, "{}", p); err != nil {
		t.Fatal(err)
	}
	if got := p.NoteTypes[7].Templates; len(got) != 2 || got[0] != want ||
		got[1] != (Template{Name: "Card 2", QFmt: "{{Back}}", AFmt: "{{Front}}"}) {
		t.Errorf("legacy templates = %+v", got)
	}
}

// Imported override (spec): a card type keeps its deck override when the
// deck is imported or exists, and its browser appearance templates.
func TestImportCardTypeOptions(t *testing.T) {
	p := &Package{
		Format: "anki2", Crt: 1767240000,
		Decks: map[int64]Deck{
			1:  {Name: "Default", PresetID: AnkiDefaultPreset},
			10: {Name: "Spanish", PresetID: AnkiDefaultPreset},
			11: {Name: "Spanish::Production", PresetID: AnkiDefaultPreset},
			12: {Name: "Cram", Filtered: true, PresetID: 5},
			13: {Name: "Elsewhere", PresetID: AnkiDefaultPreset},
		},
		NoteTypes: map[int64]NoteType{
			100: {Name: "Vocab", Fields: []string{"Front", "Back"}, Templates: []Template{
				{Name: "Recognition", QFmt: "{{Front}}", AFmt: "{{Back}}", BrowserQFmt: "{{Front}}!"},
				{Name: "Production", QFmt: "{{Back}}", AFmt: "{{Front}}", BrowserAFmt: "{{Front}}?", DeckID: 11},
			}},
			// Overrides to a filtered deck, a deck no card is in, and a deck
			// that is not in the package are dropped.
			200: {Name: "Other", Fields: []string{"Front", "Back"}, Templates: []Template{
				{Name: "A", QFmt: "{{Front}}", AFmt: "{{Back}}", DeckID: 12},
				{Name: "B", QFmt: "{{Back}}", AFmt: "{{Front}}", DeckID: 13},
				{Name: "C", QFmt: "{{Front}}{{Back}}", AFmt: "x", DeckID: 99},
			}},
		},
		Notes: []Note{
			{ID: 1, GUID: "v1", TypeID: 100, Fields: []string{"hola", "hello"}},
			{ID: 2, GUID: "o1", TypeID: 200, Fields: []string{"a", "b"}},
		},
		Cards: []Card{
			{ID: 1001, NoteID: 1, DeckID: 10, Ord: 0},
			{ID: 1002, NoteID: 1, DeckID: 11, Ord: 1},
			{ID: 1003, NoteID: 2, DeckID: 10, Ord: 0},
			{ID: 1004, NoteID: 2, DeckID: 10, Ord: 1},
			{ID: 1005, NoteID: 2, DeckID: 10, Ord: 2},
		},
	}
	b, err := Convert(p, false)
	if err != nil {
		t.Fatal(err)
	}
	c := newCollection(t)
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	deckID := map[string]core.ID{}
	for _, d := range c.Decks() {
		deckID[d.Name] = d.ID
	}
	var vocab, other core.NoteType
	for _, nt := range c.NoteTypes() {
		switch nt.Name {
		case "Vocab":
			vocab = nt
		case "Other":
			other = nt
		}
	}
	want := []core.CardTemplate{
		{Name: "Recognition", QFmt: "{{Front}}", AFmt: "{{Back}}", BrowserQFmt: "{{Front}}!"},
		{Name: "Production", QFmt: "{{Back}}", AFmt: "{{Front}}", BrowserAFmt: "{{Front}}?", DeckOverride: deckID["Spanish::Production"]},
	}
	if deckID["Spanish::Production"] == 0 || !reflect.DeepEqual(vocab.Templates, want) {
		t.Fatalf("Vocab templates = %+v, want %+v", vocab.Templates, want)
	}
	for _, tm := range other.Templates {
		if tm.DeckOverride != 0 {
			t.Errorf("Other card type %s overrides to %d", tm.Name, tm.DeckOverride)
		}
	}

	// A later package naming a deck that exists by then keeps the override.
	elsewhere, err := c.CreateDeck("Elsewhere")
	if err != nil {
		t.Fatal(err)
	}
	p.NoteTypes[300] = NoteType{Name: "Third", Fields: []string{"Front"}, Templates: []Template{
		{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Front}}", DeckID: 13}}}
	p.Notes = []Note{{ID: 3, GUID: "t1", TypeID: 300, Fields: []string{"x"}}}
	p.Cards = []Card{{ID: 1006, NoteID: 3, DeckID: 10}}
	if b, err = Convert(p, false); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, nt := range c.NoteTypes() {
		if nt.Name == "Third" {
			found = true
			if nt.Templates[0].DeckOverride != elsewhere {
				t.Fatalf("Third = %+v, want override %d", nt.Templates, elsewhere)
			}
		}
	}
	if !found {
		t.Fatal("note type Third not imported")
	}
}

// Field options and the sort field are read from both package formats.
func TestReadFieldOptions(t *testing.T) {
	var cfg []byte
	for _, num := range []protowire.Number{2, 7, 8} {
		cfg = protowire.AppendTag(cfg, num, protowire.VarintType)
		cfg = protowire.AppendVarint(cfg, 1)
	}
	cfg = protowire.AppendTag(cfg, 3, protowire.BytesType)
	cfg = protowire.AppendString(cfg, "Amiri")
	cfg = protowire.AppendTag(cfg, 4, protowire.VarintType)
	cfg = protowire.AppendVarint(cfg, 28)
	cfg = protowire.AppendTag(cfg, 5, protowire.BytesType)
	cfg = protowire.AppendString(cfg, "a description")
	want := FieldConfig{RTL: true, Font: "Amiri", Size: 28, Collapsed: true, ExcludeFromSearch: true}
	if got := modernFieldConfig(cfg); got != want {
		t.Errorf("modern field config = %+v", got)
	}

	p := &Package{Decks: map[int64]Deck{}, NoteTypes: map[int64]NoteType{}}
	models := `{"7": {"name": "Arabic", "type": 0, "sortf": 1, "flds": [
		{"name": "English", "ord": 0, "sticky": true, "rtl": false, "font": "Arial", "size": 20, "plainText": true},
		{"name": "Arabic", "ord": 1, "rtl": true, "font": "Amiri", "size": 28, "collapsed": true, "excludeFromSearch": true}],
		"tmpls": [{"name": "Card 1", "ord": 0, "qfmt": "{{English}}", "afmt": "{{Arabic}}"}]}}`
	if err := readLegacyTypes(models, "{}", p); err != nil {
		t.Fatal(err)
	}
	nt := p.NoteTypes[7]
	if nt.SortField != 1 || !reflect.DeepEqual(nt.FieldConfigs, []FieldConfig{
		{Sticky: true, Font: "Arial", Size: 20, PlainText: true}, want}) {
		t.Errorf("legacy note type: sort field %d, fields %+v", nt.SortField, nt.FieldConfigs)
	}
}

// Imported options (spec): a note type with a right-to-left field that
// sorts by its second field keeps both. Anki's default font and size are
// the editor's defaults.
func TestImportFieldOptions(t *testing.T) {
	p := &Package{
		Format: "anki2", Crt: 1767240000,
		Decks: map[int64]Deck{1: {Name: "Default", PresetID: AnkiDefaultPreset}},
		NoteTypes: map[int64]NoteType{
			100: {Name: "Arabic", Fields: []string{"English", "Arabic", "Notes"}, SortField: 1,
				FieldConfigs: []FieldConfig{
					{Font: "Arial", Size: 20, Sticky: true, PlainText: true},
					{Font: "Amiri", Size: 28, RTL: true, Collapsed: true},
					{Font: "Arial", Size: 20, ExcludeFromSearch: true},
				},
				Templates: []Template{{Name: "Card 1", QFmt: "{{English}}", AFmt: "{{Arabic}}"}}},
			// A sort field that names no field sorts by the first, and
			// fields with Anki's defaults store nothing.
			200: {Name: "Plain", Fields: []string{"Front", "Back"}, SortField: 5,
				FieldConfigs: []FieldConfig{{Font: "Arial", Size: 20}, {Font: "Arial", Size: 20}},
				Templates:    []Template{{Name: "Card 1", QFmt: "{{Front}}", AFmt: "{{Back}}"}}},
		},
		Notes: []Note{
			{ID: 1, GUID: "a1", TypeID: 100, Fields: []string{"book", "كتاب", "irregular plural"}},
			{ID: 2, GUID: "p1", TypeID: 200, Fields: []string{"a", "b"}},
		},
		Cards: []Card{{ID: 1001, NoteID: 1, DeckID: 1}, {ID: 1002, NoteID: 2, DeckID: 1}},
	}
	b, err := Convert(p, false)
	if err != nil {
		t.Fatal(err)
	}
	c := newCollection(t)
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	types := map[string]core.NoteType{}
	for _, nt := range c.NoteTypes() {
		types[nt.Name] = nt
	}
	arabic := types["Arabic"]
	want := []core.FieldOptions{
		{Sticky: true, HTMLByDefault: true},
		{Font: "Amiri", Size: 28, RTL: true, Collapsed: true},
		{ExcludeFromSearch: true},
	}
	if arabic.SortFieldIndex() != 1 || !reflect.DeepEqual(arabic.FieldOptions, want) {
		t.Fatalf("Arabic: sort field %d, options %+v", arabic.SortFieldIndex(), arabic.FieldOptions)
	}
	if plain := types["Plain"]; plain.SortFieldIndex() != 0 || plain.FieldOptions != nil {
		t.Fatalf("Plain: sort field %d, options %+v", plain.SortFieldIndex(), plain.FieldOptions)
	}
	if ids, _ := c.Search("irregular", core.ColCreated, false); len(ids) != 0 {
		t.Fatalf("excluded field matched: %v", ids)
	}
}
