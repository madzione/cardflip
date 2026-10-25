package core

import (
	"cardflip/golang/internal/usermsg"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
)

// undoFixture is a small collection with nested decks, every built-in note
// type, tags, an image, an empty card and reviewed cards.
type undoFixture struct {
	c        *Collection
	store    *storage.MemoryStore
	clk      *clock
	spanish  ID
	exam     ID // preset of Lang::Spanish, random insertion order
	basic    ID
	cloze    ID
	hola     ID // Basic note in Lang::Spanish, tagged, reviewed
	uno      ID // Basic (and reversed card) note
	paris    ID // Cloze note with an empty card
	cat      ID // note showing cat.png
	reviewed ID // card of hola, answered Good
	entry    ID // its review log entry
	cram     ID // filtered deck holding the unburied cards tagged verb
	buried   ID // card of uno tagged verb, buried for a day
}

func newUndoFixture(t *testing.T) *undoFixture {
	t.Helper()
	c, store, clk := testCollection(t)
	f := &undoFixture{c: c, store: store, clk: clk}
	var err error
	if f.spanish, err = c.CreateDeck("Lang::Spanish"); err != nil {
		t.Fatal(err)
	}
	if f.exam, err = c.ClonePreset(f.spanish, "Exam"); err != nil {
		t.Fatal(err)
	}
	exam := DefaultDeckOptions()
	exam.InsertionOrder = InsertRandom
	if err := c.SetPresetOptions(f.exam, exam); err != nil {
		t.Fatal(err)
	}
	f.basic = noteType(t, c, "Basic").ID
	f.cloze = noteType(t, c, "Cloze").ID
	// As in collections from before Image Occlusion was built in, which
	// get the note type on first use.
	if err := c.DeleteNoteType(noteType(t, c, StockImageOcclusion).ID); err != nil {
		t.Fatal(err)
	}
	f.hola = addNoteIn(t, c, f.spanish, "Basic", "hola", "hello")
	f.uno = addNote(t, c, "Basic (and reversed card)", "uno", "one")
	f.paris = addNote(t, c, "Cloze", "{{c1::Paris}} is in {{c2::France}}", "")
	if err := c.UpdateNote(f.paris, []string{"{{c1::Paris}} is in France", ""}, nil); err != nil {
		t.Fatal(err)
	}
	if err := c.store.PutMedia("cat.png", []byte("meow")); err != nil {
		t.Fatal(err)
	}
	f.cat = addNote(t, c, "Basic", `<img src="cat.png">`, "cat")
	f.reviewed = c.CardsOfNote(f.hola)[0].ID
	e, err := c.Answer(f.reviewed, scheduler.Good, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	f.entry = e.ID
	clk.advance(time.Hour)
	if _, err := c.Answer(c.CardsOfNote(f.uno)[0].ID, scheduler.Again, time.Second); err != nil {
		t.Fatal(err)
	}
	if err := c.AddTags([]ID{f.hola, f.uno}, []string{"verb"}); err != nil {
		t.Fatal(err)
	}
	f.buried = c.CardsOfNote(f.uno)[1].ID
	if err := c.BuryCards([]ID{f.buried}); err != nil {
		t.Fatal(err)
	}
	if f.cram, _, err = c.CreateFilteredDeck("Cram", DefaultFilterConfig("tag:verb")); err != nil {
		t.Fatal(err)
	}
	if err := c.SetDeckCollapsed(c.deckByName("Lang").ID, true); err != nil {
		t.Fatal(err)
	}
	return f
}

// collectionImage is everything undo must restore: the collection (except
// the ID counter, which never goes back), the review log in memory and the
// stored review log.
type collectionImage struct {
	State  *state
	Revlog []ID
	Stored map[string][]string
}

func imageOf(t *testing.T, c *Collection) collectionImage {
	t.Helper()
	c.mu.RLock()
	st := withoutSyncMeta(normalizeState(c.state()))
	st.NextID = 0
	var revlog []ID
	for _, e := range c.revlog {
		revlog = append(revlog, e.ID)
	}
	c.mu.RUnlock()
	stored := map[string][]string{}
	months, err := c.RevlogMonths()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range months {
		entries, err := c.ReadRevlog(m)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			stored[m] = append(stored[m], string(e))
		}
	}
	return collectionImage{State: st, Revlog: revlog, Stored: stored}
}

func sameImage(t *testing.T, what string, got, want collectionImage) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: collection differs:\n%s", what, jsonDiff(got, want))
	}
}

// jsonDiff shows the lines in which the JSON of got and want differ.
func jsonDiff(got, want any) string {
	g, _ := json.MarshalIndent(got, "", " ")
	w, _ := json.MarshalIndent(want, "", " ")
	gl, wl := strings.Split(string(g), "\n"), strings.Split(string(w), "\n")
	var out []string
	for i := 0; i < max(len(gl), len(wl)) && len(out) < 20; i++ {
		var a, b string
		if i < len(gl) {
			a = gl[i]
		}
		if i < len(wl) {
			b = wl[i]
		}
		if a != b {
			out = append(out, fmt.Sprintf("line %d\n got %s\nwant %s", i+1, a, b))
		}
	}
	return strings.Join(out, "\n")
}

// undoMutations runs each exported mutation once on the fixture. Every
// mutation must be listed here (TestUndoCoversEveryMutation).
var undoMutations = map[string]func(t *testing.T, f *undoFixture) error{
	"AddNote": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.AddNote(NoteInput{TypeID: f.cloze, DeckID: f.spanish, Fields: []string{"{{c1::a}} {{c2::b}}"}})
		return err
	},
	"AddNotes": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.AddNotes([]NoteInput{
			{TypeID: f.cloze, DeckID: f.spanish, Fields: []string{"{{c1::a}} {{c2::b}}"}},
			{TypeID: f.cloze, DeckID: f.spanish, Fields: []string{"{{c1::c}}"}},
		}, AddNotesOptions{})
		return err
	},
	"UpdateNote": func(t *testing.T, f *undoFixture) error {
		// Adds a second cloze card and changes tags.
		return f.c.UpdateNote(f.paris, []string{"{{c1::Paris}} is in {{c2::France}}", "x"}, []string{"geo"})
	},
	"DeleteNotes": func(t *testing.T, f *undoFixture) error {
		return f.c.DeleteNotes([]ID{f.hola, f.uno, f.cat})
	},
	"AddTags": func(t *testing.T, f *undoFixture) error {
		return f.c.AddTags([]ID{f.hola, f.paris}, []string{"new tags"})
	},
	"RemoveTags": func(t *testing.T, f *undoFixture) error {
		return f.c.RemoveTags([]ID{f.hola, f.uno}, []string{"VERB"})
	},
	"ToggleMark": func(t *testing.T, f *undoFixture) error {
		return f.c.ToggleMark([]ID{f.hola, f.paris})
	},
	"RenameTag": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.RenameTag("Verb", "grammar::verb")
		return err
	},
	"ReparentTags": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.ReparentTags([]string{"verb"}, "grammar")
		return err
	},
	"DeleteTags": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.DeleteTags([]string{"verb"})
		return err
	},
	"FindReplace": func(t *testing.T, f *undoFixture) error {
		// Adds a cloze card to Paris; hola and uno have no match.
		_, err := f.c.FindReplace([]ID{f.hola, f.paris, f.uno}, FindReplace{Find: "(france)", Replace: "{{c3::$1}}", Regex: true, IgnoreCase: true})
		return err
	},
	"ChangeNoteType": func(t *testing.T, f *undoFixture) error {
		// uno's reverse card moves to Card 1, its forward card is deleted.
		_, err := f.c.ChangeNoteType([]ID{f.uno}, f.basic, []int{1, 0}, []int{1})
		return err
	},
	"TagMissingMedia": func(t *testing.T, f *undoFixture) error {
		if err := f.c.store.TrashMedia([]string{"cat.png"}, f.clk.now()); err != nil {
			return err
		}
		_, err := f.c.TagMissingMedia()
		return err
	},
	"MoveCards": func(t *testing.T, f *undoFixture) error {
		return f.c.MoveCards(allCards(f.c), f.spanish)
	},
	"SetSuspended": func(t *testing.T, f *undoFixture) error {
		return f.c.SetSuspended(allCards(f.c), true)
	},
	"SetFlag": func(t *testing.T, f *undoFixture) error {
		return f.c.SetFlag(allCards(f.c), 2)
	},
	"BuryCards": func(t *testing.T, f *undoFixture) error {
		return f.c.BuryCards(allCards(f.c))
	},
	"ResetCards": func(t *testing.T, f *undoFixture) error {
		return f.c.ResetCards([]ID{f.reviewed}, ResetOptions{RestorePosition: true})
	},
	"SetDueDate": func(t *testing.T, f *undoFixture) error {
		return f.c.SetDueDate(allCards(f.c), "3-9!")
	},
	"Reposition": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.Reposition(allCards(f.c), 0, 1, false, true)
		return err
	},
	"DeleteEmptyCards": func(t *testing.T, f *undoFixture) error {
		return f.c.DeleteEmptyCards(allCards(f.c))
	},
	"CreateDeck": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.CreateDeck("Science::Physics::Optics")
		return err
	},
	"RenameDeck": func(t *testing.T, f *undoFixture) error {
		return f.c.RenameDeck(f.c.deckByName("Lang").ID, "Languages::Romance")
	},
	"SetDeckOptions": func(t *testing.T, f *undoFixture) error {
		return f.c.SetDeckOptions(f.spanish, deckOpts(5, 50, 0.85))
	},
	"AddPreset": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.AddPreset(f.spanish, "Fresh")
		return err
	},
	"ClonePreset": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.ClonePreset(f.spanish, "Exam copy")
		return err
	},
	"RenamePreset": func(t *testing.T, f *undoFixture) error {
		return f.c.RenamePreset(f.exam, "Finals")
	},
	"DeletePreset": func(t *testing.T, f *undoFixture) error {
		return f.c.DeletePreset(f.exam)
	},
	"SetDeckPreset": func(t *testing.T, f *undoFixture) error {
		return f.c.SetDeckPreset(f.c.deckByName("Lang").ID, f.exam)
	},
	"SavePresetToSubdecks": func(t *testing.T, f *undoFixture) error {
		return f.c.SavePresetToSubdecks(f.c.deckByName("Lang").ID)
	},
	"SetPresetOptions": func(t *testing.T, f *undoFixture) error {
		return f.c.SetPresetOptions(f.exam, deckOpts(5, 50, 0.85))
	},
	"SaveDeckOptions": func(t *testing.T, f *undoFixture) error {
		n := 3
		o := DefaultDeckOptions()
		o.InsertionOrder = InsertRandom
		o.NewPerDay = 7
		return f.c.SaveDeckOptions(f.c.deckByName("Lang").ID, f.exam, o, DeckLimits{Review: &n})
	},
	"SetDeckLimits": func(t *testing.T, f *undoFixture) error {
		n := 5
		return f.c.SetDeckLimits(f.spanish, DeckLimits{New: &n})
	},
	"IncreaseTodayLimits": func(t *testing.T, f *undoFixture) error {
		return f.c.IncreaseTodayLimits(f.spanish, 10, 5)
	},
	"Unbury": func(t *testing.T, f *undoFixture) error {
		return f.c.Unbury(f.c.deckByName("Lang").ID)
	},
	"SetDeckDescription": func(t *testing.T, f *undoFixture) error {
		return f.c.SetDeckDescription(f.spanish, "Words from <b>Book 1</b>")
	},
	"DeleteDeck": func(t *testing.T, f *undoFixture) error {
		return f.c.DeleteDeck(f.c.deckByName("Lang").ID)
	},
	"CreateFilteredDeck": func(t *testing.T, f *undoFixture) error {
		_, _, err := f.c.CreateFilteredDeck("Lang::Cram", DefaultFilterConfig("note:Cloze"))
		return err
	},
	"UpdateFilteredDeck": func(t *testing.T, f *undoFixture) error {
		cfg := DefaultFilterConfig("note:Cloze")
		cfg.Reschedule = false
		_, err := f.c.UpdateFilteredDeck(f.cram, "Old::Cram", cfg)
		return err
	},
	"RebuildFilteredDeck": func(t *testing.T, f *undoFixture) error {
		// The buried card is no longer buried, so the deck gains it.
		f.clk.advance(24 * time.Hour)
		_, err := f.c.RebuildFilteredDeck(f.cram)
		return err
	},
	"EmptyFilteredDeck": func(t *testing.T, f *undoFixture) error {
		return f.c.EmptyFilteredDeck(f.cram)
	},
	"CustomStudy": func(t *testing.T, f *undoFixture) error {
		_, _, err := f.c.CustomStudy(f.c.deckByName(DefaultDeckName).ID, CustomStudy{Kind: CustomByState, State: CustomNew, Cards: 10})
		return err
	},
	"AddNoteType": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.AddNoteType(StockOptionalReversed, "Optional")
		return err
	},
	"ImageOcclusionNoteType": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.ImageOcclusionNoteType()
		return err
	},
	"CloneNoteType": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.CloneNoteType(f.basic, "Basic copy")
		return err
	},
	"DeleteNoteType": func(t *testing.T, f *undoFixture) error {
		return f.c.DeleteNoteType(f.basic)
	},
	"RenameNoteType": func(t *testing.T, f *undoFixture) error {
		return f.c.RenameNoteType(f.basic, "Simple")
	},
	"RenameField": func(t *testing.T, f *undoFixture) error {
		return f.c.RenameField(f.basic, 0, "Question")
	},
	"AddField": func(t *testing.T, f *undoFixture) error {
		return f.c.AddField(f.basic, "Extra")
	},
	"RemoveField": func(t *testing.T, f *undoFixture) error {
		return f.c.RemoveField(f.basic, 1)
	},
	"AddTemplate": func(t *testing.T, f *undoFixture) error {
		return f.c.AddTemplate(f.basic, CardTemplate{Name: "Reverse", QFmt: "{{Back}}", AFmt: "{{Front}}"})
	},
	"UpdateTemplate": func(t *testing.T, f *undoFixture) error {
		return f.c.UpdateTemplate(f.basic, 0, CardTemplate{Name: "Card 1", QFmt: "{{Front}}?", AFmt: "{{Back}}"})
	},
	"RemoveTemplate": func(t *testing.T, f *undoFixture) error {
		rev := noteType(t, f.c, "Basic (and reversed card)").ID
		return f.c.RemoveTemplate(rev, 1)
	},
	"MoveTemplate": func(t *testing.T, f *undoFixture) error {
		rev := noteType(t, f.c, "Basic (and reversed card)").ID
		return f.c.MoveTemplate(rev, 1, 0)
	},
	"SetNoteTypeCSS": func(t *testing.T, f *undoFixture) error {
		return f.c.SetNoteTypeCSS(f.basic, ".card { color: red }")
	},
	"SetFieldOptions": func(t *testing.T, f *undoFixture) error {
		return f.c.SetFieldOptions(f.basic, 1, FieldOptions{HTMLByDefault: true})
	},
	"MoveField": func(t *testing.T, f *undoFixture) error {
		return f.c.MoveField(f.basic, 1, 0)
	},
	"SetSortField": func(t *testing.T, f *undoFixture) error {
		return f.c.SetSortField(f.basic, 1)
	},
	"Import": func(t *testing.T, f *undoFixture) error {
		hola, _ := f.c.Note(f.hola)
		_, err := f.c.Import(&ImportBatch{
			NoteTypes: []ImportNoteType{
				{Key: "basic", ExistingID: f.basic},
				{Key: "vocab", Type: NoteType{Name: "Vocab", Kind: Standard, Fields: []string{"Word", "Meaning"},
					Templates: []CardTemplate{{Name: "C", QFmt: "{{Word}}", AFmt: "{{Meaning}}"}}}},
			},
			Notes: []ImportNote{
				{TypeKey: "vocab", DeckName: "Imported::Words", Fields: []string{`perro <img src="dog.png">`, "dog"}},
				// A card with review history: one entry today (in memory)
				// and one a year ago (stored only).
				{TypeKey: "vocab", DeckName: "Imported::Words", Fields: []string{"gato", "cat"},
					Cards: []ImportCard{{Ord: 0, State: scheduler.CardState{State: scheduler.Learning, Due: f.clk.now()},
						History: []ImportReview{
							{Time: f.clk.now().AddDate(-1, 0, 0), Rating: scheduler.Good, Source: "test:1"},
							{Time: f.clk.now().Add(-time.Minute), Rating: scheduler.Again, StateBefore: scheduler.Review, Source: "test:2"},
						}}}},
				{TypeKey: "basic", GUID: hola.GUID, DeckName: "Imported", Fields: []string{"hola!", "hello"}},
			},
			Media:      map[string][]byte{"dog.png": []byte("woof")},
			Duplicates: DupDuplicate,
		}, nil)
		return err
	},
	"Answer": func(t *testing.T, f *undoFixture) error {
		_, err := f.c.Answer(f.reviewed, scheduler.Easy, 3*time.Second)
		return err
	},
	"UndoReview": func(t *testing.T, f *undoFixture) error {
		// Not the latest step: reverting the answer is a step of its own.
		_, err := f.c.UndoReview(f.entry)
		return err
	},
}

// Exported methods of Collection that are not undo steps.
var notUndoSteps = []string{
	// Reads.
	"AllCards", "AllNoteIDs", "AllTags", "BasicNoteType", "Card", "CardCount", "CardHistory", "CheckSearch", "CheckTemplate", "CardsOfNote", "CardsOfNotes", "CardStatesPerDeck", "Deck", "DeckCounts", "DeckDeletionInfo", "DeckLabel", "DeckOptions", "Decks", "DeletionLog", "Duplicates",
	"EffectiveLimits", "Export", "ExportReviews", "ExtraCards", "PresetUsage", "Presets",
	"Optimize", "OptimizeSearch", "RescheduleCount", "RescheduleCountAll", "Simulate", "SuggestRetention",
	"EachMedia", "EmptyCards", "FindDuplicates", "FindReplacePreview", "ChangeNoteTypePreview", "HasMedia", "IsEmptyCard", "Media", "NextCountChange", "NextDue", "IsMarked", "Note",
	"NoteCount", "NoteType", "NoteTypeDeletionInfo", "NoteTypes", "Now", "Overview", "PreviewTemplate", "Previews", "QuestionText", "Queue", "ReadDeletionLog", "ReadRevlog", "RenderCard", "RenderCardTyped",
	"RecentAnswers", "RecentDecks", "RecentNotes", "ReviewsExist", "StudyDays", "ReviewsOfCard", "Revision", "RevlogMonths", "NoteRows", "Rows", "Search", "SearchNotes", "Settings", "Snapshot",
	"Subscribe", "TagUsage", "TemplateCardCount", "UndoStatus", "LastStep",
	// The history itself, view state, settings and the built-in names
	// (app state, not collection data), and replacing the
	// collection, which clears the history.
	"Undo", "Redo", "SetBuiltinNames", "SetDeckCollapsed", "UpdateSettings", "SetCurrentDeck", "Restore", "RestoreKeepingMedia", "ReplaceWithImport", "Close",
	// The deletion log is not part of the collection the history reverts;
	// clearing it is confirmed instead.
	"ClearDeletionLog",
	// Adding a media file changes no note; undoing the note edit that uses
	// it leaves the file to compaction.
	"AddMedia",
	// Media files are not part of the collection the history reverts:
	// deleting them moves them to the media trash, which restores them.
	"CheckMedia", "DeleteUnusedMedia", "RestoreMediaTrash", "EmptyMediaTrash",
	// The tag list is derived from the notes, so there is nothing to clear.
	"ClearUnusedTags",
	// Sync state: the device, its slot and the media manifest are not
	// collection content the history reverts (sync design decisions 4 and
	// 13).
	"SetDevice", "SetSlots", "Slot", "Slots", "BuildMediaManifest", "HasMediaManifest",
	"Delta", "ApplyBatch", "MediaBlob", "MissingMedia", "PutSyncedMedia", "Identities", "Renumber", "PruneTombstones",
	"Epoch", "Seq", "BeginReplace",
}

// undoTestedElsewhere are undo steps that need a setup undoMutations cannot
// give: Check database only changes a corrupted collection
// (TestCheckDatabaseOneChange), and Optimize all needs hundreds of reviews
// (TestOptimizeAll).
var undoTestedElsewhere = []string{"CheckDatabase", "OptimizeAll"}

// replayed opens a copy of the fixture's store, as the app would after a
// crash at this moment, and returns what it holds.
func replayed(t *testing.T, f *undoFixture) collectionImage {
	t.Helper()
	c, fresh, err := Open(f.store.Clone().Open(), Options{Now: f.clk.now})
	if err != nil || fresh {
		t.Fatalf("replay: %v fresh=%v", err, fresh)
	}
	return imageOf(t, c)
}

func allCards(c *Collection) []ID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var ids []ID
	for id := range c.cards {
		ids = append(ids, id)
	}
	sortIDs(ids)
	return ids
}

// A new exported mutation must be added to undoMutations, so that the test
// below checks its undo.
func TestUndoCoversEveryMutation(t *testing.T) {
	known := map[string]bool{}
	for _, name := range append(notUndoSteps, undoTestedElsewhere...) {
		known[name] = true
	}
	for name := range undoMutations {
		known[name] = true
	}
	typ := reflect.TypeOf(&Collection{})
	var missing []string
	for i := range typ.NumMethod() {
		if name := typ.Method(i).Name; !known[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("add these methods to undoMutations or notUndoSteps: %s", strings.Join(missing, ", "))
	}
}

// Undo changes (spec): every mutation is undone exactly, redone exactly, and
// both are durable.
func TestUndoEveryMutation(t *testing.T) {
	names := make([]string, 0, len(undoMutations))
	for name := range undoMutations {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			f := newUndoFixture(t)
			c := f.c
			orig := imageOf(t, c)
			if err := undoMutations[name](t, f); err != nil {
				t.Fatal(err)
			}
			after := imageOf(t, c)
			if reflect.DeepEqual(after, orig) {
				t.Fatal("the mutation changed nothing")
			}
			undo, _ := c.UndoStatus()
			if undo == "" || undo == "Change" {
				t.Fatalf("step name = %q", undo)
			}
			if got, err := c.Undo(); err != nil || got != undo {
				t.Fatalf("undo = %q, %v", got, err)
			}
			sameImage(t, "after undo", imageOf(t, c), orig)
			sameImage(t, "replayed after undo", replayed(t, f), orig)
			if _, redo := c.UndoStatus(); redo != undo {
				t.Fatalf("redo = %q, want %q", redo, undo)
			}
			if _, err := c.Redo(); err != nil {
				t.Fatal(err)
			}
			sameImage(t, "after redo", imageOf(t, c), after)
			sameImage(t, "replayed after redo", replayed(t, f), after)
			// And undo again, from the refreshed before-images.
			if _, err := c.Undo(); err != nil {
				t.Fatal(err)
			}
			sameImage(t, "after second undo", imageOf(t, c), orig)
			sameImage(t, "replayed after second undo", replayed(t, f), orig)
		})
	}
}

// The history keeps the last MaxUndoSteps steps, and fewer when their
// before-images pass the size bound; the latest step is always kept.
func TestUndoHistoryBounds(t *testing.T) {
	c, _, _ := testCollection(t)
	deck := defaultDeck(c)
	for i := range MaxUndoSteps + 5 {
		if err := c.SetDeckOptions(deck, deckOpts(i+1, 100, 0.9)); err != nil {
			t.Fatal(err)
		}
	}
	undone := 0
	for {
		if _, err := c.Undo(); err != nil {
			mustUserErr(t, err, "nothing to undo")
			break
		}
		undone++
	}
	if o := deckOptions(c, deck); undone != MaxUndoSteps || o.NewPerDay != 5 {
		t.Fatalf("undid %d steps, new per day %d", undone, o.NewPerDay)
	}

	// A step larger than the size bound drops all older steps.
	defer func(old int) { maxUndoBytes = old }(maxUndoBytes)
	maxUndoBytes = 4 << 10
	c, _, _ = testCollection(t)
	addNote(t, c, "Basic", "small", "")
	big := addNote(t, c, "Basic", strings.Repeat("x", 8<<10), "")
	if err := c.DeleteNotes([]ID{big}); err != nil {
		t.Fatal(err)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Delete 1 note" {
		t.Fatalf("undo = %q", undo)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Note(big); err != nil {
		t.Fatal("the large step was not undone")
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "" {
		t.Fatalf("older steps kept: %q", undo)
	}
}

// Undo a delete (spec): the notes and their cards come back with their
// scheduling and review history.
func TestUndoDeleteNotes(t *testing.T) {
	f := newUndoFixture(t)
	c := f.c
	orig := imageOf(t, c)
	if err := c.DeleteNotes([]ID{f.hola, f.uno, f.paris}); err != nil {
		t.Fatal(err)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Delete 3 notes" {
		t.Fatalf("undo = %q", undo)
	}
	if name, err := c.Undo(); err != nil || usermsg.Plain(name) != "Delete 3 notes" {
		t.Fatalf("undo = %q, %v", name, err)
	}
	sameImage(t, "after undo", imageOf(t, c), orig)
	card, _ := c.Card(f.reviewed)
	if card.State == scheduler.New || len(c.ReviewsOfCard(f.reviewed)) != 1 {
		t.Fatalf("scheduling or history lost: %+v", card)
	}
}

// Undo an import (spec): the notes, their cards and the decks the import
// created are gone.
func TestUndoImport(t *testing.T) {
	c, _, _ := testCollection(t)
	orig := imageOf(t, c)
	basic := noteType(t, c, "Basic").ID
	b := &ImportBatch{NoteTypes: []ImportNoteType{{Key: "b", ExistingID: basic}}, Duplicates: DupDuplicate}
	for i := range 200 {
		b.Notes = append(b.Notes, ImportNote{TypeKey: "b", DeckName: "Package::Chapter " + string(rune('A'+i%3)),
			Fields: []string{fmt.Sprintf("q%d", i), "a"}})
	}
	sum, err := c.Import(b, nil)
	if err != nil || sum.NotesAdded != 200 {
		t.Fatalf("import: %+v, %v", sum, err)
	}
	if name, err := c.Undo(); err != nil || usermsg.Plain(name) != "Import" {
		t.Fatalf("undo = %q, %v", name, err)
	}
	sameImage(t, "after undo", imageOf(t, c), orig)
	if names := deckNames(c); names != DefaultDeckName {
		t.Fatalf("decks = %s", names)
	}
}

// Redo a move (spec), and redo is only possible until another change.
func TestRedoMove(t *testing.T) {
	c, _, _ := testCollection(t)
	id := addNote(t, c, "Basic (and reversed card)", "q", "a")
	archive, _ := c.CreateDeck("Archive")
	cards := []ID{c.CardsOfNote(id)[0].ID, c.CardsOfNote(id)[1].ID}
	if err := c.MoveCards(cards, archive); err != nil {
		t.Fatal(err)
	}
	inArchive := func() bool {
		for _, card := range c.CardsOfNote(id) {
			if card.DeckID != archive {
				return false
			}
		}
		return true
	}
	if _, err := c.Undo(); err != nil || inArchive() {
		t.Fatalf("undo: %v", err)
	}
	if name, err := c.Redo(); err != nil || usermsg.Plain(name) != "Move 2 cards" || !inArchive() {
		t.Fatalf("redo = %q, %v", name, err)
	}
	if _, redo := c.UndoStatus(); usermsg.Plain(redo) != "" {
		t.Fatalf("redo after redo = %q", redo)
	}

	// A new change clears the redo stack.
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	addNote(t, c, "Basic", "other", "")
	if _, redo := c.UndoStatus(); usermsg.Plain(redo) != "" {
		t.Fatalf("redo after a new change = %q", redo)
	}
	_, err := c.Redo()
	mustUserErr(t, err, "nothing to redo")
}

// History cleared by restore (spec).
func TestRestoreClearsHistory(t *testing.T) {
	c, _, _ := testCollection(t)
	snap, _ := c.Snapshot()
	addNote(t, c, "Basic", "q", "")
	addNote(t, c, "Basic", "q2", "")
	c.Undo()
	if err := c.Restore(snap, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if undo, redo := c.UndoStatus(); usermsg.Plain(undo) != "" || usermsg.Plain(redo) != "" {
		t.Fatalf("history after restore: %q, %q", undo, redo)
	}
}

// Undo is durable: after a crash right after an undo the collection opens
// as undone, with an empty history.
func TestCrashAfterUndo(t *testing.T) {
	f := newUndoFixture(t)
	orig := imageOf(t, f.c)
	if err := f.c.DeleteDeck(f.c.deckByName("Lang").ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.c.Undo(); err != nil {
		t.Fatal(err)
	}
	c2, _, err := Open(f.store.Clone().Open(), Options{Now: f.clk.now})
	if err != nil {
		t.Fatal(err)
	}
	sameImage(t, "after crash", imageOf(t, c2), orig)
	if undo, redo := c2.UndoStatus(); usermsg.Plain(undo) != "" || usermsg.Plain(redo) != "" {
		t.Fatalf("history after reopening: %q, %q", undo, redo)
	}
}

// Collapsing decks and changing settings are not steps: they neither
// appear in the history nor clear redo, and undo leaves them alone.
func TestUntrackedChanges(t *testing.T) {
	c, _, _ := testCollection(t)
	lang, _ := c.CreateDeck("Lang::Spanish")
	parent := c.deckByName("Lang").ID
	if err := c.RenameDeck(parent, "Languages"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetDeckCollapsed(parent, true); err != nil {
		t.Fatal(err)
	}
	s := c.Settings()
	s.Theme = ThemeDark
	if err := c.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Rename deck" {
		t.Fatalf("undo = %q", undo)
	}
	c.Undo()
	if d, _ := c.Deck(parent); d.Name != "Lang" || !d.Collapsed {
		t.Fatalf("after undo: %+v", d)
	}
	c.SetDeckCollapsed(parent, false)
	if _, redo := c.UndoStatus(); usermsg.Plain(redo) != "Rename deck" {
		t.Fatalf("redo = %q", redo)
	}
	c.Redo()
	if d, _ := c.Deck(parent); d.Name != "Languages" || d.Collapsed {
		t.Fatalf("after redo: %+v", d)
	}

	// Undoing an add reverts the remembered deck, but not the theme changed
	// since.
	if _, err := c.AddNote(NoteInput{TypeID: noteType(t, c, "Basic").ID, DeckID: lang, Fields: []string{"q"}}); err != nil {
		t.Fatal(err)
	}
	s = c.Settings()
	s.Theme = ThemeLight
	c.UpdateSettings(s)
	c.Undo()
	if s := c.Settings(); s.LastDeckID == lang || s.Theme != ThemeLight {
		t.Fatalf("settings after undo: %+v", s)
	}
}

// Undoing the latest answer from study goes through the history, so it can
// be redone; undoing an older one is a step of its own.
func TestUndoReviewUsesHistory(t *testing.T) {
	f := newUndoFixture(t)
	c := f.c
	e, err := c.Answer(f.reviewed, scheduler.Easy, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if undo, _ := c.UndoStatus(); usermsg.Plain(undo) != "Answer card" {
		t.Fatalf("undo = %q", undo)
	}
	if _, err := c.UndoReview(e.ID); err != nil {
		t.Fatal(err)
	}
	if undo, redo := c.UndoStatus(); usermsg.Plain(undo) != "Build filtered deck" || usermsg.Plain(redo) != "Answer card" {
		t.Fatalf("history = %q, %q", undo, redo)
	}
	if _, err := c.Redo(); err != nil || len(c.ReviewsOfCard(f.reviewed)) != 2 {
		t.Fatalf("redo: %v", err)
	}
	if !c.ReviewsExist([]ID{e.ID, f.entry})[e.ID] {
		t.Fatal("redone entry missing")
	}
}

// Undo a delete (spec): the note's image is still there after a compaction
// in between, since the history references it. Without the history it is
// removed.
func TestUndoKeepsMediaThroughCompaction(t *testing.T) {
	f := newUndoFixture(t)
	c := f.c
	if err := c.DeleteNotes([]ID{f.cat}); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	err := c.compact()
	c.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if data, err := c.Media("cat.png"); err != nil || string(data) != "meow" {
		t.Fatalf("image after undo: %q, %v", data, err)
	}

	// Redo the delete: the redo stack holds the note, so the image stays.
	c.Redo()
	c.mu.Lock()
	c.compact()
	c.mu.Unlock()
	if !c.HasMedia("cat.png") {
		t.Fatal("image removed while the note can be brought back")
	}
	snap, _ := c.Snapshot()
	c.RestoreKeepingMedia(snap, nil, nil)
	c.mu.Lock()
	c.compact()
	c.mu.Unlock()
	if c.HasMedia("cat.png") {
		t.Fatal("unreferenced image kept without history")
	}
}

func heapAfterGC() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// After 30 steps that each change all 50,000 cards or notes, the history
// holds about its 50 MB bound (design decision 2), not 30 copies.
func TestUndoMemory50k(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := largeCollection(t, 50000)
	home := defaultDeck(c)
	archive, _ := c.CreateDeck("Archive")
	cards := allCards(c)
	var notes []ID
	for id := range c.notes {
		notes = append(notes, id)
	}
	base := heapAfterGC()
	for i := range 30 {
		var err error
		switch i % 3 {
		case 0:
			err = c.MoveCards(cards, archive)
		case 1:
			err = c.MoveCards(cards, home)
		case 2:
			err = c.AddTags(notes, []string{fmt.Sprintf("round%d", i)})
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	with := heapAfterGC()
	c.mu.RLock()
	steps, estimate := len(c.hist.undo), c.hist.bytes
	c.mu.RUnlock()
	c.mu.Lock()
	c.hist.clear()
	c.mu.Unlock()
	without := heapAfterGC()
	runtime.KeepAlive(c)
	history := int64(with) - int64(without)
	t.Logf("collection %d MB; after 30 steps: %d steps kept, estimated %d MB, measured %d MB of history",
		base>>20, steps, estimate>>20, history>>20)
	if steps == 30 || steps == 0 {
		t.Fatalf("%d steps kept", steps)
	}
	if history > int64(maxUndoBytes)*3/2 {
		t.Fatalf("history uses %d MB", history>>20)
	}
}
