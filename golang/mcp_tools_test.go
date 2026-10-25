//go:build !js

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
	"cardflip/golang/internal/usermsg"
)

// mcpClient connects an MCP client to the app's tools in process.
func mcpClient(t *testing.T, s *services) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := newMcpServer(s.mcp).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cs.Close()
		ss.Wait()
	})
	return cs
}

// callTool calls a tool and decodes its result into Out. A tool error is
// returned as its text.
func callTool[Out any](t *testing.T, cs *mcp.ClientSession, name string, args any) (Out, string) {
	t.Helper()
	var out Out
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		var texts []string
		for _, c := range res.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				texts = append(texts, tc.Text)
			}
		}
		return out, strings.Join(texts, "\n")
	}
	data, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("%s: %v in %s", name, err, data)
	}
	return out, ""
}

// call is callTool failing the test on a tool error.
func call[Out any](t *testing.T, cs *mcp.ClientSession, name string, args any) Out {
	t.Helper()
	out, msg := callTool[Out](t, cs, name, args)
	if msg != "" {
		t.Fatalf("%s: %s", name, msg)
	}
	return out
}

// typeNamed returns the note type called name.
func typeNamed(t *testing.T, s *services, name string) NoteType {
	t.Helper()
	for _, nt := range must(s.notes.Types()) {
		if nt.Name == name {
			return nt
		}
	}
	t.Fatalf("no note type %q", name)
	return NoteType{}
}

func TestMcpFieldsByName(t *testing.T) {
	names := []string{"Front", "Back"}
	got, err := fieldsByName("Basic", names, nil, map[string]string{"Back": "hello"})
	if err != nil || len(got) != 2 || got[0] != "" || got[1] != "hello" {
		t.Fatalf("fields = %q, %v", got, err)
	}
	got, err = fieldsByName("Basic", names, []string{"hola", "hi"}, map[string]string{"Back": "hello"})
	if err != nil || got[0] != "hola" || got[1] != "hello" {
		t.Fatalf("merged = %q, %v", got, err)
	}
	// Unknown field (spec): the call names it and the note type's fields.
	_, err = fieldsByName("Basic", names, nil, map[string]string{"Answer": "x", "Front": "y"})
	if err == nil || err.Error() != `Note type "Basic" has no field "Answer". Its fields are: Front, Back.` {
		t.Fatalf("err = %v", err)
	}
	if m := namedFields(names, []string{"hola"}); m["Front"] != "hola" || m["Back"] != "" || len(m) != 2 {
		t.Fatalf("named = %v", m)
	}
}

func TestMcpCardText(t *testing.T) {
	s, _ := opened(t)
	tools := newMcpTools(s.mcp)
	c := must(tools.col())
	deck := must(s.notes.Defaults(0)).DeckID
	// Basic card (spec): question hola, answer hello without the question.
	basic := typeNamed(t, s, "Basic")
	id := must(s.notes.Add(NewNote{TypeID: basic.ID, DeckID: deck, Fields: []string{"hola", "<b>hello</b>"}}))
	card := must(s.notes.Get(id)).CardIDs[0]
	q, a, err := tools.cardText(c, core.ID(card))
	if err != nil || q != "hola" || a != "hello" {
		t.Fatalf("basic = %q, %q, %v", q, a, err)
	}
	// Cloze card 2 (spec).
	cloze := typeNamed(t, s, "Cloze")
	id = must(s.notes.Add(NewNote{TypeID: cloze.ID, DeckID: deck, Fields: []string{"{{c1::Paris}} is the capital of {{c2::France}}"}}))
	n := must(s.notes.Get(id))
	for i, ord := range n.CardOrds {
		if ord != 1 {
			continue
		}
		q, a, err := tools.cardText(c, core.ID(n.CardIDs[i]))
		if err != nil || q != "Paris is the capital of [...]" || a != "Paris is the capital of France" {
			t.Fatalf("cloze = %q, %q, %v", q, a, err)
		}
		return
	}
	t.Fatal("no card 2")
}

func TestMcpNoCollection(t *testing.T) {
	s := newServices(t, nil)
	must(s.app.LoadProfiles("/data"))
	if _, err := newMcpTools(s.mcp).col(); err == nil || err.Error() != errNoCollection {
		t.Fatalf("err = %v", err)
	}
}

func TestMcpReadTools(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	basicNote(t, s, "other", "y", "verb") // in Default
	spanish := must(s.decks.Create("Spanish"))
	must(s.decks.Create("Spanish::Verbs"))
	for i := range 7 {
		basicNoteIn(t, s, spanish, fmt.Sprintf("word %d", i), "x")
	}

	decks := call[mcpDeckList](t, cs, "list_decks", map[string]any{})
	var found bool
	for _, d := range decks.Decks {
		if d.Name == "Spanish" {
			found = d.Cards == 7 && d.New == 7
		}
	}
	if !found {
		t.Fatalf("decks = %+v", decks)
	}
	ov := call[mcpDeckOverview](t, cs, "get_deck_overview", map[string]any{"deck": "spanish"})
	if ov.Deck != "Spanish" || ov.New != 7 {
		t.Fatalf("overview = %+v", ov)
	}
	if _, msg := callTool[mcpDeckOverview](t, cs, "get_deck_overview", map[string]any{"deck": "French"}); !strings.Contains(msg, `no deck named "French"`) {
		t.Fatalf("unknown deck: %q", msg)
	}
	types := call[mcpNoteTypeList](t, cs, "list_note_types", map[string]any{})
	if len(types.NoteTypes) == 0 || types.NoteTypes[0].Fields == nil || len(types.NoteTypes[0].Templates) == 0 {
		t.Fatalf("note types = %+v", types)
	}
	if tags := call[mcpTagList](t, cs, "list_tags", map[string]any{}); fmt.Sprint(tags.Tags) != "[verb]" {
		t.Fatalf("tags = %+v", tags)
	}

	// Search (spec): at most limit cards, the total, and the next offset.
	pg := call[mcpCardPage](t, cs, "search_cards", map[string]any{"query": "deck:Spanish is:new", "limit": 5})
	if pg.Total != 7 || len(pg.Cards) != 5 || pg.NextOffset == nil || *pg.NextOffset != 5 || pg.Cards[0].Question != "word 0" {
		t.Fatalf("page 1 = %+v", pg)
	}
	pg = call[mcpCardPage](t, cs, "search_cards", map[string]any{"query": "deck:Spanish is:new", "limit": 5, "offset": 5})
	if pg.Total != 7 || len(pg.Cards) != 2 || pg.NextOffset != nil || pg.Cards[1].Question != "word 6" {
		t.Fatalf("page 2 = %+v", pg)
	}
	if pg := call[mcpCardPage](t, cs, "search_cards", map[string]any{"query": ""}); pg.Total != 8 || len(pg.Cards) != 8 {
		t.Fatalf("all = %+v", pg)
	}
	// Invalid search (spec): the browser's message.
	_, want := s.browse.Search("(deck:Spanish", "created", false)
	if _, msg := callTool[mcpCardPage](t, cs, "search_cards", map[string]any{"query": "(deck:Spanish"}); want == nil || msg != toolError(want).Error() {
		t.Fatalf("invalid search = %q, want %v", msg, want)
	}

	notes := call[mcpNotePage](t, cs, "search_notes", map[string]any{"query": "tag:verb"})
	if notes.Total != 1 || notes.Notes[0].Fields["Front"] != "other" || notes.Notes[0].NoteType != "Basic" || len(notes.Notes[0].Cards) != 1 {
		t.Fatalf("notes = %+v", notes)
	}
	got := call[mcpNoteList](t, cs, "get_notes", map[string]any{"noteIds": []int64{notes.Notes[0].NoteID}})
	if len(got.Notes) != 1 || got.Notes[0].Fields["Back"] != "y" || fmt.Sprint(got.Notes[0].Tags) != "[verb]" {
		t.Fatalf("get_notes = %+v", got)
	}
	if _, msg := callTool[mcpNoteList](t, cs, "get_notes", map[string]any{"noteIds": []int64{1}}); msg != "Note 1: That note no longer exists." {
		t.Fatalf("missing note = %q", msg)
	}
	cards := call[mcpCardList](t, cs, "get_cards", map[string]any{"cardIds": notes.Notes[0].Cards})
	if c := cards.Cards[0]; c.Question != "other" || c.Answer != "y" || c.State != "new" || c.Fields["Front"] != "other" || c.Deck == "" || c.CardType != "Card 1" {
		t.Fatalf("get_cards = %+v", cards)
	}
}

func TestMcpStatsTool(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	spanish := must(s.decks.Create("Spanish"))
	verbs := must(s.decks.Create("Spanish::Verbs"))
	basicNote(t, s, "c", "y") // in Default
	basicNoteIn(t, s, spanish, "a", "x")
	basicNoteIn(t, s, verbs, "b", "x")
	all := call[mcpStats](t, cs, "get_stats", map[string]any{})
	if all.Scope != "collection" || all.New != 3 || all.Total != 3 || len(all.DueNext30Days) != 30 {
		t.Fatalf("collection = %+v", all)
	}
	// Deck forecast (spec): the deck and its subdecks only.
	deck := call[mcpStats](t, cs, "get_stats", map[string]any{"deck": "Spanish"})
	if deck.Scope != "Spanish" || deck.New != 2 || deck.Total != 2 {
		t.Fatalf("deck = %+v", deck)
	}
	// No goal: no streak yet, level 1 and a null goal.
	if all.Streak != (mcpStreak{}) || all.Level != (mcpLevel{Level: 1, NextLevelXP: 50}) || all.Goal != nil {
		t.Fatalf("no goal: streak %+v, level %+v, goal %+v", all.Streak, all.Level, all.Goal)
	}
	// Goal progress (spec): 20 of 30 cards today, a 12-day streak; the
	// streak is the collection's, also for a deck. Level (spec): so is the
	// level, from 31 answers, 12 study days and bronze Streak: 176 XP.
	importHistory(t, s, "history", 10, 1000, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1)
	studyNew(t, s, 20, 1000)
	check(t, s.app.SetDailyGoal("cards", 30))
	for _, in := range []map[string]any{{}, {"deck": "Spanish"}} {
		got := call[mcpStats](t, cs, "get_stats", in)
		if got.Streak != (mcpStreak{Current: 12, TodayStudied: true}) || got.Goal == nil ||
			*got.Goal != (mcpGoal{Kind: "cards", Value: 30, Progress: 20}) {
			t.Fatalf("%v: streak %+v, goal %+v", in, got.Streak, got.Goal)
		}
		if got.Level != (mcpLevel{Level: 3, XP: 176, NextLevelXP: 300}) {
			t.Fatalf("%v: level %+v", in, got.Level)
		}
	}
}

func TestMcpToolsWithoutCollection(t *testing.T) {
	s := newServices(t, nil)
	must(s.app.LoadProfiles("/data"))
	cs := mcpClient(t, s)
	// Profile list at startup (spec).
	if _, msg := callTool[mcpDeckList](t, cs, "list_decks", map[string]any{}); msg != errNoCollection {
		t.Fatalf("list_decks = %q", msg)
	}
}

func TestMcpNoteTools(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	basicNote(t, s, "hola", "hello") // the user's last add: Basic in Default
	before := must(s.notes.Defaults(0))
	must(s.decks.Create("Spanish"))
	activity := len(s.mcp.Activity())

	// Batch add (spec), with a duplicate skipped.
	res := call[mcpAddNotesResult](t, cs, "add_notes", map[string]any{"notes": []map[string]any{
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "uno", "Back": "one"}, "tags": []string{"num"}},
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "dos", "Back": "two"}},
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "tres"}},
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "HOLA", "Back": "hi"}},
		{"noteType": "Cloze", "deck": "Spanish", "fields": map[string]string{"Text": "{{c1::cuatro}} is four"}},
	}})
	if len(res.Added) != 4 || len(res.Duplicates) != 1 || res.Duplicates[0].Index != 3 || len(res.Added[0].Cards) != 1 {
		t.Fatalf("add = %+v", res)
	}
	// Add screen unchanged (spec).
	if after := must(s.notes.Defaults(0)); after != before {
		t.Fatalf("defaults = %+v, were %+v", after, before)
	}
	// One activity entry for the change, named as the undo step.
	act := s.mcp.Activity()
	if len(act) != activity+1 || usermsg.Plain(act[0].Step) != "Add 4 notes" {
		t.Fatalf("activity = %+v", act)
	}
	if undo := must(s.app.UndoInfo()); usermsg.Plain(undo.Undo) != "Add 4 notes" {
		t.Fatalf("undo = %q", undo.Undo)
	}
	// A read call adds no activity.
	call[mcpDeckList](t, cs, "list_decks", map[string]any{})
	if len(s.mcp.Activity()) != activity+1 {
		t.Fatal("read call listed")
	}

	// Unknown field (spec): no note of the call is added.
	notes := len(must(s.browse.SearchNotes("", "created", false)))
	_, msg := callTool[mcpAddNotesResult](t, cs, "add_notes", map[string]any{"notes": []map[string]any{
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "cinco"}},
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "seis", "Answer": "six"}},
	}})
	if msg != `Note 2: Note type "Basic" has no field "Answer". Its fields are: Front, Back.` {
		t.Fatalf("unknown field = %q", msg)
	}
	if n := len(must(s.browse.SearchNotes("", "created", false))); n != notes {
		t.Fatalf("%d notes, want %d", n, notes)
	}
	// A note without cards fails the whole call with the app's message.
	_, msg = callTool[mcpAddNotesResult](t, cs, "add_notes", map[string]any{"notes": []map[string]any{
		{"noteType": "Basic", "deck": "Spanish", "fields": map[string]string{"Front": "siete"}},
		{"noteType": "Cloze", "deck": "Spanish", "fields": map[string]string{"Text": "no cloze"}},
	}})
	if !strings.HasPrefix(msg, "Note 2: This would produce no cards.") || len(s.mcp.Activity()) != activity+1 {
		t.Fatalf("no cards = %q", msg)
	}

	// Update a field (spec): the other fields and scheduling stay.
	uno := res.Added[0]
	upd := call[mcpNoteResult](t, cs, "update_note", map[string]any{"noteId": uno.NoteID, "fields": map[string]string{"Back": "<b>one</b>"}})
	if upd.Note.Fields["Front"] != "uno" || upd.Note.Fields["Back"] != "<b>one</b>" || fmt.Sprint(upd.Note.Tags) != "[num]" ||
		fmt.Sprint(upd.Note.Cards) != fmt.Sprint(uno.Cards) {
		t.Fatalf("update = %+v", upd)
	}
	upd = call[mcpNoteResult](t, cs, "update_note", map[string]any{"noteId": uno.NoteID, "tags": []string{"a", "b"}})
	if fmt.Sprint(upd.Note.Tags) != "[a b]" || upd.Note.Fields["Back"] != "<b>one</b>" {
		t.Fatalf("update tags = %+v", upd)
	}

	// Tags.
	ids := []int64{res.Added[1].NoteID, res.Added[2].NoteID}
	if r := call[mcpCountResult](t, cs, "add_tags", map[string]any{"noteIds": ids, "tags": []string{"verb", "spanish::basics"}}); r.Notes != 2 {
		t.Fatalf("add_tags = %+v", r)
	}
	call[mcpCountResult](t, cs, "remove_tags", map[string]any{"noteIds": ids, "tags": []string{"VERB"}})
	if r := call[mcpCountResult](t, cs, "rename_tag", map[string]any{"tag": "spanish", "newName": "es"}); r.Notes != 2 {
		t.Fatalf("rename_tag = %+v", r)
	}
	if tags := must(s.notes.Tags()); !strings.Contains(fmt.Sprint(tags), "es::basics") || strings.Contains(fmt.Sprint(tags), "verb") {
		t.Fatalf("tags = %v", tags)
	}

	// Delete counts; the activity lists it above the add.
	del := call[mcpDeleteResult](t, cs, "delete_notes", map[string]any{"noteIds": []int64{res.Added[3].NoteID, 12345}})
	if del.DeletedNotes != 1 || del.DeletedCards != 1 {
		t.Fatalf("delete = %+v", del)
	}
	if act := s.mcp.Activity(); usermsg.Plain(act[0].Step) != "Delete 1 note" {
		t.Fatalf("activity = %+v", act)
	}
	// A change that changes nothing makes no step and is not listed.
	n := len(s.mcp.Activity())
	call[mcpCountResult](t, cs, "add_tags", map[string]any{"noteIds": ids, "tags": []string{"es::basics"}})
	if len(s.mcp.Activity()) != n {
		t.Fatal("no-op change listed")
	}
}

func TestMcpStoreMedia(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	png := base64.StdEncoding.EncodeToString(pngBytes(t, 2, 2))
	r := call[mcpMediaResult](t, cs, "store_media", map[string]any{"name": "map.png", "data": png})
	if r.Name != "map.png" {
		t.Fatalf("name = %q", r.Name)
	}
	// Same content: same name. Another file with the name: renamed.
	if r := call[mcpMediaResult](t, cs, "store_media", map[string]any{"name": "map.png", "data": png}); r.Name != "map.png" {
		t.Fatalf("same file = %q", r.Name)
	}
	other := base64.StdEncoding.EncodeToString(pngBytes(t, 3, 3))
	if r := call[mcpMediaResult](t, cs, "store_media", map[string]any{"name": "map.png", "data": other}); r.Name != "map-1.png" {
		t.Fatalf("name clash = %q", r.Name)
	}
	if got := must(s.media.Get("map-1.png")); base64.StdEncoding.EncodeToString(got) != other {
		t.Fatal("stored content differs")
	}
	if _, msg := callTool[mcpMediaResult](t, cs, "store_media", map[string]any{"name": "run.sh", "data": "aGk="}); !strings.Contains(msg, "Only image, audio and video") {
		t.Fatalf("script = %q", msg)
	}
	// Over HTTP, a file of 20 MB is taken and a larger one refused.
	hc := mcpHTTPClient(t, s)
	big := make([]byte, mcpMaxMedia)
	if r := call[mcpMediaResult](t, hc, "store_media", map[string]any{"name": "big.mp4", "data": base64.StdEncoding.EncodeToString(big)}); r.Name != "big.mp4" {
		t.Fatalf("20 MB = %q", r.Name)
	}
	big = append(big, 0)
	if _, msg := callTool[mcpMediaResult](t, hc, "store_media", map[string]any{"name": "big2.mp4", "data": base64.StdEncoding.EncodeToString(big)}); msg != "The file is larger than 20 MB." {
		t.Fatalf("big = %q", msg)
	}
}

// bearer adds the token to every request.
type bearer struct {
	token string
	next  http.RoundTripper
}

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return b.next.RoundTrip(r)
}

// mcpHTTPClient turns the server on and connects an MCP client to it over
// HTTP, as an agent does.
func mcpHTTPClient(t *testing.T, s *services) *mcp.ClientSession {
	t.Helper()
	st := mcpOn(t, s)
	transport := &mcp.StreamableClientTransport{Endpoint: st.URL,
		HTTPClient: &http.Client{Transport: bearer{st.Token, http.DefaultTransport}}}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(context.Background(), transport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

// pngBytes returns a w×h PNG image.
func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	check(t, png.Encode(&buf, image.NewGray(image.Rect(0, 0, w, h))))
	return buf.Bytes()
}

func TestMcpCardAndDeckTools(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	var cards []int64
	for i := range 6 {
		n := basicNote(t, s, fmt.Sprintf("w%d", i), "x")
		cards = append(cards, must(s.notes.Get(n)).CardIDs...)
	}
	// Suspend (spec): the cards are suspended and not studied.
	if r := call[mcpCardsResult](t, cs, "suspend_cards", map[string]any{"cardIds": cards[:5]}); r.Cards != 5 {
		t.Fatalf("suspend = %+v", r)
	}
	if n := len(must(s.browse.Search("is:suspended", "created", false))); n != 5 {
		t.Fatalf("%d suspended", n)
	}
	if n := len(must(s.browse.Search("is:new -is:suspended", "created", false))); n != 1 {
		t.Fatalf("%d to study", n)
	}
	call[mcpCardsResult](t, cs, "unsuspend_cards", map[string]any{"cardIds": cards[:1]})
	if _, msg := callTool[mcpCardsResult](t, cs, "suspend_cards", map[string]any{"cardIds": []int64{1}}); msg != "Card 1: That card no longer exists." {
		t.Fatalf("missing card = %q", msg)
	}

	// Setting a flag twice keeps it.
	for range 2 {
		call[mcpCardsResult](t, cs, "set_flag", map[string]any{"cardIds": cards[:2], "flag": 3})
	}
	if n := len(must(s.browse.Search("flag:3", "created", false))); n != 2 {
		t.Fatalf("%d flagged", n)
	}
	call[mcpCardsResult](t, cs, "set_flag", map[string]any{"cardIds": cards[:2], "flag": 0})
	if n := len(must(s.browse.Search("-flag:0", "created", false))); n != 0 {
		t.Fatalf("%d flagged", n)
	}

	// Nested deck (spec).
	d := call[mcpDeckResult](t, cs, "create_deck", map[string]any{"name": "Spanish::Verbs"})
	if d.Name != "Spanish::Verbs" {
		t.Fatalf("deck = %+v", d)
	}
	var names []string
	for _, n := range must(s.decks.Tree()) {
		names = append(names, n.FullName)
	}
	if !strings.Contains(fmt.Sprint(names), "Spanish Spanish::Verbs") {
		t.Fatalf("decks = %v", names)
	}
	// Duplicate deck name (spec): the app's message, nothing created.
	_, want := s.decks.Create("SPANISH")
	decks := len(must(s.decks.Tree()))
	if _, msg := callTool[mcpDeckResult](t, cs, "create_deck", map[string]any{"name": "spanish"}); want == nil || msg != usermsg.Plain(want.Error()) {
		t.Fatalf("taken name = %q, want %v", msg, want)
	}
	if n := len(must(s.decks.Tree())); n != decks {
		t.Fatalf("%d decks, want %d", n, decks)
	}
	r := call[mcpDeckResult](t, cs, "rename_deck", map[string]any{"deck": "spanish::verbs", "newName": "Spanish::Verbos"})
	if r.ID != d.ID || r.Name != "Spanish::Verbos" {
		t.Fatalf("rename = %+v", r)
	}

	call[mcpCardsResult](t, cs, "move_cards", map[string]any{"cardIds": cards[:3], "deck": "Spanish::Verbos"})
	if n := len(must(s.browse.Search(`"deck:Spanish::Verbos"`, "created", false))); n != 3 {
		t.Fatalf("%d moved", n)
	}
	call[mcpCardsResult](t, cs, "set_due_date", map[string]any{"cardIds": cards[:2], "days": "3"})
	if n := len(must(s.browse.Search("prop:due=3", "created", false))); n != 2 {
		t.Fatalf("%d due in 3 days", n)
	}
	if _, msg := callTool[mcpCardsResult](t, cs, "set_due_date", map[string]any{"cardIds": cards[:2], "days": "soon"}); msg == "" {
		t.Fatal("invalid due date accepted")
	}
	call[mcpCardsResult](t, cs, "reset_cards", map[string]any{"cardIds": cards[:2]})
	if n := len(must(s.browse.Search("is:review", "created", false))); n != 0 {
		t.Fatalf("%d review cards after reset", n)
	}
}

func TestMcpNoteTypeTools(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	vocab := call[mcpNoteTypeResult](t, cs, "create_note_type", map[string]any{"name": "Vocab", "stock": "basic"}).NoteType
	if vocab.Name != "Vocab" || fmt.Sprint(vocab.Fields) != "[Front Back]" {
		t.Fatalf("created = %+v", vocab)
	}
	vocabID := vocab.ID
	deck := must(s.notes.Defaults(0)).DeckID
	for _, w := range []string{"uno", "dos"} {
		must(s.notes.Add(NewNote{TypeID: vocabID, DeckID: deck, Fields: []string{w, "x"}}))
	}
	// Add a field (spec): every note of Vocab has it empty.
	vocab = call[mcpNoteTypeResult](t, cs, "add_field", map[string]any{"noteType": "Vocab", "field": "Example"}).NoteType
	if fmt.Sprint(vocab.Fields) != "[Front Back Example]" {
		t.Fatalf("fields = %v", vocab.Fields)
	}
	for _, n := range call[mcpNotePage](t, cs, "search_notes", map[string]any{"query": "note:Vocab"}).Notes {
		if v, ok := n.Fields["Example"]; !ok || v != "" {
			t.Fatalf("note = %+v", n)
		}
	}
	vocab = call[mcpNoteTypeResult](t, cs, "rename_field", map[string]any{"noteType": "Vocab", "field": "Example", "newName": "Sentence"}).NoteType
	vocab = call[mcpNoteTypeResult](t, cs, "move_field", map[string]any{"noteType": "Vocab", "name": "Sentence", "position": 0}).NoteType
	if fmt.Sprint(vocab.Fields) != "[Sentence Front Back]" {
		t.Fatalf("fields = %v", vocab.Fields)
	}
	rf := call[mcpRemoveFieldResult](t, cs, "remove_field", map[string]any{"noteType": "Vocab", "field": "Sentence"})
	if rf.NotesAffected != 2 || fmt.Sprint(rf.NoteType.Fields) != "[Front Back]" {
		t.Fatalf("remove field = %+v", rf)
	}

	// Template error (spec): the template check's message, nothing changed.
	want := s.notes.UpdateTemplate(vocabID, 0, CardTemplate{Name: "Card 1", QFmt: "{{Nope}}", AFmt: "{{Back}}"})
	if want == nil {
		t.Fatal("template with an unknown field saved")
	}
	_, msg := callTool[mcpNoteTypeResult](t, cs, "update_template", map[string]any{"noteType": "Vocab", "template": "Card 1", "front": "{{Nope}}"})
	if msg != usermsg.Plain(want.Error()) {
		t.Fatalf("template error = %q, want %q", msg, usermsg.Plain(want.Error()))
	}
	if nt := must(s.notes.Type(vocabID)); nt.Templates[0].QFmt != vocab.Templates[0].Front {
		t.Fatalf("template changed: %q", nt.Templates[0].QFmt)
	}

	vocab = call[mcpNoteTypeResult](t, cs, "add_template", map[string]any{"noteType": "Vocab", "name": "Reverse", "front": "{{Back}}", "back": "{{Front}}"}).NoteType
	if len(vocab.Templates) != 2 || len(must(s.browse.Search("note:Vocab", "created", false))) != 4 {
		t.Fatalf("add template = %+v", vocab)
	}
	vocab = call[mcpNoteTypeResult](t, cs, "update_template", map[string]any{"noteType": "Vocab", "template": "Reverse", "newName": "Back to front"}).NoteType
	vocab = call[mcpNoteTypeResult](t, cs, "move_template", map[string]any{"noteType": "Vocab", "name": "Back to front", "position": 0}).NoteType
	if vocab.Templates[0].Name != "Back to front" || vocab.Templates[0].Front != "{{Back}}" {
		t.Fatalf("templates = %+v", vocab.Templates)
	}
	rt := call[mcpRemoveTemplateResult](t, cs, "remove_template", map[string]any{"noteType": "Vocab", "template": "Back to front"})
	if rt.CardsDeleted != 2 || len(rt.NoteType.Templates) != 1 {
		t.Fatalf("remove template = %+v", rt)
	}
	vocab = call[mcpNoteTypeResult](t, cs, "set_note_type_css", map[string]any{"noteType": "Vocab", "css": ".card { color: red; }"}).NoteType
	if vocab.CSS != ".card { color: red; }" {
		t.Fatalf("css = %q", vocab.CSS)
	}
	vocab = call[mcpNoteTypeResult](t, cs, "rename_note_type", map[string]any{"noteType": "Vocab", "newName": "Words"}).NoteType
	copied := call[mcpNoteTypeResult](t, cs, "create_note_type", map[string]any{"name": "Words 2", "copyOf": "Words"}).NoteType
	if copied.CSS != vocab.CSS || copied.Notes != 0 {
		t.Fatalf("copy = %+v", copied)
	}
	del := call[mcpDeleteNoteTypeResult](t, cs, "delete_note_type", map[string]any{"noteType": "Words"})
	if del.NotesDeleted != 2 || del.CardsDeleted != 2 {
		t.Fatalf("delete = %+v", del)
	}
}

func TestMcpToolAnnotations(t *testing.T) {
	s, _ := opened(t)
	cs := mcpClient(t, s)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		a := tool.Annotations
		if a == nil || (!a.ReadOnlyHint && a.DestructiveHint == nil) || tool.Description == "" {
			t.Errorf("tool %s: annotations %+v", tool.Name, a)
		}
		if a != nil && a.ReadOnlyHint != (strings.HasPrefix(tool.Name, "list_") || strings.HasPrefix(tool.Name, "get_") || strings.HasPrefix(tool.Name, "search_")) {
			t.Errorf("tool %s: readOnlyHint %v", tool.Name, a.ReadOnlyHint)
		}
	}
}

// The tools of design decision 5.
var mcpToolNames = []string{
	"list_decks", "get_deck_overview", "list_note_types", "list_tags", "search_cards", "search_notes", "get_notes", "get_cards", "get_stats",
	"add_notes", "update_note", "delete_notes", "add_tags", "remove_tags", "rename_tag", "store_media",
	"suspend_cards", "unsuspend_cards", "set_flag", "move_cards", "reset_cards", "set_due_date",
	"create_deck", "rename_deck",
	"create_note_type", "rename_note_type", "add_field", "rename_field", "remove_field", "move_field",
	"add_template", "update_template", "remove_template", "move_template", "set_note_type_css", "delete_note_type",
}

// An agent over HTTP: initialize, list the tools, add a note, and the app
// sees the change (spec: changes show in the app).
func TestMcpOverHTTP(t *testing.T) {
	s, _ := opened(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := s.app.Changes(ctx)
	cs := mcpHTTPClient(t, s) // initializes the session
	if init := cs.InitializeResult(); init == nil || init.ServerInfo.Name != "cardflip" || init.Instructions == "" {
		t.Fatalf("initialize = %+v", init)
	}
	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	want := slices.Sorted(slices.Values(mcpToolNames))
	if !slices.Equal(names, want) {
		t.Fatalf("tools = %v\nwant %v", names, want)
	}
	added := call[mcpAddNotesResult](t, cs, "add_notes", map[string]any{"notes": []map[string]any{
		{"noteType": "Basic", "deck": "Default", "fields": map[string]string{"Front": "hola", "Back": "hello"}},
	}})
	if len(added.Added) != 1 {
		t.Fatalf("add = %+v", added)
	}
	timeout := time.After(5 * time.Second)
	for {
		select {
		case ch := <-changes:
			if ch.Kind == "addNotes" && ch.Notes && ch.Cards {
				return
			}
		case <-timeout:
			t.Fatal("no change for the added note")
		}
	}
}
