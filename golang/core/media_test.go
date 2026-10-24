package core

import (
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"cardflip/golang/render"
	"cardflip/golang/storage"
)

// Compaction keeps audio and video files, not only images, and files that
// CSS refers to with url().
func TestMediaNamesIncludeSoundsAndVideos(t *testing.T) {
	got := MediaNames(`<img src="a.png">[sound: hola.mp3 ]`, `<audio src='b.ogg'></audio><video controls src=c.mp4>`,
		`<video><source src="d.webm"></video>`, `<img src="https://x/y.png">[sound:]`,
		`@font-face { src: url("_f.ttf") } .card { background: URL( 'bg.jpg' ) } i { background: url(i.png) }`,
		`<div style="background:url(&quot;s.png&quot;)">`, `b { background: url(data:image/png;base64,xx) } u { background: url(//cdn/x.png) }`,
		`<img src="my%20cat.jpg">`)
	want := []string{"a.png", "hola.mp3", "b.ogg", "c.mp4", "d.webm", "_f.ttf", "bg.jpg", "i.png", "&quot;s.png&quot;", `"s.png"`,
		"my%20cat.jpg", "my cat.jpg"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MediaNames = %q, want %q", got, want)
	}
}

// A file added in the editor survives a compaction before its note is
// saved, but not one in a later session if no note uses it.
func TestAddedMediaKeptForTheSession(t *testing.T) {
	c, store, clk := testCollection(t)
	name, err := c.AddMedia("pending.png", []byte("x"))
	if err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	err = c.compact()
	c.mu.Unlock()
	if err != nil || !c.HasMedia(name) {
		t.Fatalf("compaction in the same session removed %s: %v", name, err)
	}
	c2, _, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	c2.mu.Lock()
	err = c2.compact()
	c2.mu.Unlock()
	if err != nil || c2.HasMedia(name) {
		t.Fatalf("unreferenced %s kept in a later session: %v", name, err)
	}
}

func mediaCheck(t *testing.T, c *Collection) *MediaCheck {
	t.Helper()
	r, err := c.CheckMedia()
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func putMedia(t *testing.T, c *Collection, files ...string) {
	t.Helper()
	for _, name := range files {
		if err := c.store.PutMedia(name, []byte("data of "+name)); err != nil {
			t.Fatal(err)
		}
	}
}

func unusedNames(r *MediaCheck) []string {
	out := []string{}
	for _, f := range r.Unused {
		out = append(out, f.Name)
	}
	return out
}

// Check media (spec): unused and missing.
func TestCheckMediaUnusedAndMissing(t *testing.T) {
	c, _, _ := testCollection(t)
	putMedia(t, c, "old.jpg", "cat.jpg")
	addNoteIn(t, c, defaultDeck(c), "Basic", `<img src="cat.jpg">`, "[sound:gone.mp3]")
	r := mediaCheck(t, c)
	if got := unusedNames(r); !reflect.DeepEqual(got, []string{"old.jpg"}) {
		t.Fatalf("unused = %q, want old.jpg", got)
	}
	if r.UnusedBytes != int64(len("data of old.jpg")) || r.Unused[0].Size != r.UnusedBytes {
		t.Fatalf("unused size = %d", r.UnusedBytes)
	}
	if !reflect.DeepEqual(r.Missing, []string{"gone.mp3"}) || r.MissingNotes != 1 {
		t.Fatalf("missing = %q in %d notes, want gone.mp3 in 1", r.Missing, r.MissingNotes)
	}
}

// Check media (spec): a static template file is not unused. Files that
// templates and CSS refer to are in use too.
func TestCheckMediaStaticAndTemplateFiles(t *testing.T) {
	c, _, _ := testCollection(t)
	putMedia(t, c, "_logo.png", "tmpl.png", "font.woff", "a&b.mp3", "my cat.jpg")
	basic := noteType(t, c, "Basic")
	tmpl := basic.Templates[0]
	tmpl.QFmt += `<img src="tmpl.png">`
	if err := c.UpdateTemplate(basic.ID, 0, tmpl); err != nil {
		t.Fatal(err)
	}
	if err := c.SetNoteTypeCSS(basic.ID, basic.CSS+`@font-face { font-family: f; src: url("font.woff"); }`); err != nil {
		t.Fatal(err)
	}
	addNoteIn(t, c, defaultDeck(c), "Basic", "[sound:a&amp;b.mp3]", `<img src="my%20cat.jpg">`)
	r := mediaCheck(t, c)
	if len(r.Unused) != 0 || len(r.Missing) != 0 {
		t.Fatalf("unused = %q, missing = %q; want none", unusedNames(r), r.Missing)
	}
}

// Media trash (spec): delete unused files, then restore them.
func TestDeleteUnusedAndRestore(t *testing.T) {
	c, _, _ := testCollection(t)
	putMedia(t, c, "old.jpg", "older.png", "_static.css", "cat.jpg")
	addNoteIn(t, c, defaultDeck(c), "Basic", `<img src="cat.jpg">`, "")
	n, err := c.DeleteUnusedMedia()
	if err != nil || n != 2 {
		t.Fatalf("DeleteUnusedMedia = %d, %v; want 2", n, err)
	}
	r := mediaCheck(t, c)
	if len(r.Unused) != 0 || len(r.Trash) != 2 || r.Trash[0].Name != "old.jpg" || r.Trash[1].Name != "older.png" {
		t.Fatalf("after delete: unused %q, trash %v", unusedNames(r), r.Trash)
	}
	if r.TrashBytes != int64(len("data of old.jpg")+len("data of older.png")) {
		t.Fatalf("trash size = %d", r.TrashBytes)
	}
	if c.HasMedia("old.jpg") || !c.HasMedia("_static.css") || !c.HasMedia("cat.jpg") {
		t.Fatal("wrong files moved")
	}
	if n, err := c.RestoreMediaTrash(); err != nil || n != 2 {
		t.Fatalf("RestoreMediaTrash = %d, %v; want 2", n, err)
	}
	if data, err := c.Media("old.jpg"); err != nil || string(data) != "data of old.jpg" {
		t.Fatalf("restored old.jpg = %q, %v", data, err)
	}
	if r := mediaCheck(t, c); len(r.Trash) != 0 || len(r.Unused) != 2 {
		t.Fatalf("after restore: unused %q, trash %v", unusedNames(r), r.Trash)
	}
	c.DeleteUnusedMedia()
	if err := c.EmptyMediaTrash(); err != nil {
		t.Fatal(err)
	}
	if r := mediaCheck(t, c); len(r.Trash) != 0 || r.TrashBytes != 0 || c.HasMedia("old.jpg") {
		t.Fatalf("after empty: trash %v", r.Trash)
	}
}

// Tag notes with missing media (spec): find notes to fix.
func TestTagMissingMedia(t *testing.T) {
	c, _, _ := testCollection(t)
	putMedia(t, c, "here.png")
	deck := defaultDeck(c)
	var broken []ID
	for _, f := range []string{"[sound:gone.mp3]", `<img src="gone.png">`, `<img src="here.png"><img src="lost.png">`} {
		broken = append(broken, addNoteIn(t, c, deck, "Basic", f, ""))
	}
	fine := addNoteIn(t, c, deck, "Basic", `<img src="here.png">`, "no media")
	n, err := c.TagMissingMedia()
	if err != nil || n != 3 {
		t.Fatalf("TagMissingMedia = %d, %v; want 3", n, err)
	}
	for _, id := range broken {
		if n, _ := c.Note(id); !reflect.DeepEqual(n.Tags, []string{MissingMediaTag}) {
			t.Fatalf("note %d tags = %q", id, n.Tags)
		}
	}
	if n, _ := c.Note(fine); len(n.Tags) != 0 {
		t.Fatalf("note without missing media tagged: %q", n.Tags)
	}
	if _, err := c.Undo(); err != nil {
		t.Fatal(err)
	}
	if n, _ := c.Note(broken[0]); len(n.Tags) != 0 {
		t.Fatal("tagging is not undone in one step")
	}
}

// Compaction moves unreferenced media to the trash, keeps static files and
// files the templates and CSS use, and deletes files trashed more than 30
// days ago (spec: media trash).
func TestCompactionTrashesUnusedMedia(t *testing.T) {
	c, store, clk := testCollection(t)
	putMedia(t, c, "orphan.jpg", "_static.js", "bg.png")
	basic := noteType(t, c, "Basic")
	if err := c.SetNoteTypeCSS(basic.ID, `.card { background: url(bg.png); }`); err != nil {
		t.Fatal(err)
	}
	compact := func(c *Collection) {
		t.Helper()
		c.mu.Lock()
		defer c.mu.Unlock()
		if err := c.compact(); err != nil {
			t.Fatal(err)
		}
	}
	compact(c)
	r := mediaCheck(t, c)
	if len(r.Trash) != 1 || r.Trash[0].Name != "orphan.jpg" || !c.HasMedia("_static.js") || !c.HasMedia("bg.png") {
		t.Fatalf("trash after compaction = %v", r.Trash)
	}
	c2, _, err := Open(store.Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	clk.advance(30 * 24 * time.Hour)
	compact(c2)
	if r := mediaCheck(t, c2); len(r.Trash) != 1 {
		t.Fatalf("trash after 30 days = %v, want orphan.jpg still there", r.Trash)
	}
	clk.advance(time.Hour)
	compact(c2)
	if r := mediaCheck(t, c2); len(r.Trash) != 0 {
		t.Fatalf("trash after more than 30 days = %v, want empty", r.Trash)
	}
}

// LaTeX blocks show the image Anki generated for them when the media has
// it, PNG or SVG by the note type's setting, and their source otherwise.
func TestRenderLatexImages(t *testing.T) {
	c, _, _ := testCollection(t)
	const png = "latex-76b2878564ab19b80abb21ba964abe90fe120846.png"
	id := addNoteIn(t, c, defaultDeck(c), "Basic", "[$]x^2[/$]", "b")
	card := c.CardsOfNote(id)[0].ID
	question := func() string {
		t.Helper()
		r, err := c.RenderCard(card, render.Env{})
		if err != nil {
			t.Fatal(err)
		}
		return r.Question
	}
	if q := question(); !strings.Contains(q, `<pre class="latex">$x^2$</pre>`) || strings.Contains(q, "<img") {
		t.Fatalf("without the image: %s", q)
	}
	if name, err := c.AddMedia(png, []byte("png")); err != nil || name != png {
		t.Fatalf("AddMedia = %q, %v", name, err)
	}
	if q := question(); !strings.Contains(q, `src="`+png+`"`) {
		t.Fatalf("with the image: %s", q)
	}
	// Set to SVG, the note type looks for the SVG image, which is missing.
	c.noteTypes[noteType(t, c, "Basic").ID].LatexSVG = true
	if q := question(); !strings.Contains(q, "<pre") {
		t.Fatalf("SVG without the image: %s", q)
	}
	svg := strings.TrimSuffix(png, ".png") + ".svg"
	if name, err := c.AddMedia(svg, []byte("<svg/>")); err != nil || name != svg {
		t.Fatalf("AddMedia = %q, %v", name, err)
	}
	if q := question(); !strings.Contains(q, `src="`+svg+`"`) {
		t.Fatalf("SVG with the image: %s", q)
	}
}

func manifestOf(c *Collection) map[string]mediaEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.media == nil {
		return nil
	}
	return maps.Clone(c.media)
}

// The media manifest built on an existing store lists exactly the stored
// files with their content's hash, and survives a restart.
func TestMediaManifestBuild(t *testing.T) {
	c, store, clk := testCollection(t)
	files := map[string]string{"a.png": "png", "b.mp3": "mp3", "_static.css": "css"}
	for name, data := range files {
		if err := c.store.PutMedia(name, []byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	if c.HasMediaManifest() {
		t.Fatal("manifest before building it")
	}
	// Without a manifest, adding media journals nothing.
	before := c.store.JournalEntries()
	if _, err := c.AddMedia("c.jpg", []byte("jpg")); err != nil {
		t.Fatal(err)
	}
	if c.store.JournalEntries() != before {
		t.Fatal("media added without a manifest was journaled")
	}
	files["c.jpg"] = "jpg"
	noErr(t, c.BuildMediaManifest())
	got := manifestOf(c)
	if len(got) != len(files) {
		t.Fatalf("manifest = %v, want %v", got, files)
	}
	var stamp Stamp
	seqs := map[int64]bool{}
	for name, data := range files {
		e, ok := got[name]
		if !ok || e.SHA1 != sha1Hex([]byte(data)) || e.Trashed || e.Stamp == 0 {
			t.Fatalf("%s: %+v", name, e)
		}
		if stamp != 0 && e.Stamp != stamp || seqs[e.Seq] {
			t.Fatalf("%s: stamp %d (others %d), seq %d", name, e.Stamp, stamp, e.Seq)
		}
		stamp, seqs[e.Seq] = e.Stamp, true
	}
	// Building again changes nothing.
	noErr(t, c.BuildMediaManifest())
	if !maps.Equal(manifestOf(c), got) {
		t.Fatal("second build changed the manifest")
	}
	replayed, _, err := Open(store.Clone().Open(), Options{Now: clk.now})
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(manifestOf(replayed), got) {
		t.Fatalf("replayed manifest = %v", manifestOf(replayed))
	}
	c2 := reopen(t, c, store, clk)
	if !maps.Equal(manifestOf(c2), got) {
		t.Fatalf("reopened manifest = %v", manifestOf(c2))
	}
}

// A file added while the manifest is being built is in it.
func TestMediaManifestBuildWhileAdding(t *testing.T) {
	store := storage.NewMemoryStore()
	var c *Collection
	added := false
	yield := func() {
		if c != nil && c.mediaBuild != nil && !added {
			added = true
			if _, err := c.AddMedia("late.png", []byte("late")); err != nil {
				t.Error(err)
			}
		}
	}
	c, _, err := Open(store.Open(), Options{Yield: yield})
	if err != nil {
		t.Fatal(err)
	}
	putMedia(t, c, "early.png")
	noErr(t, c.BuildMediaManifest())
	if !added {
		t.Fatal("nothing added during the build")
	}
	got := manifestOf(c)
	if got["late.png"].SHA1 != sha1Hex([]byte("late")) || got["early.png"].SHA1 == "" {
		t.Fatalf("manifest = %v", got)
	}
}

// Adding, trashing and restoring files and importing them update the
// manifest, each with a newer stamp; none of it is an undo step.
func TestMediaManifestChanges(t *testing.T) {
	c, store, clk := testCollection(t)
	putMedia(t, c, "old.png")
	noErr(t, c.BuildMediaManifest())
	steps, _ := c.LastStep()
	name, err := c.AddMedia("cat.png", []byte("meow"))
	if err != nil {
		t.Fatal(err)
	}
	added := manifestOf(c)[name]
	if added.SHA1 != sha1Hex([]byte("meow")) || added.Trashed || added.Seq == 0 {
		t.Fatalf("added: %+v", added)
	}
	// The same content again changes nothing.
	if again, err := c.AddMedia("other.png", []byte("meow")); err != nil || again != name {
		t.Fatalf("AddMedia = %q, %v", again, err)
	}
	if manifestOf(c)[name] != added || len(manifestOf(c)) != 2 {
		t.Fatalf("manifest after adding the same content: %v", manifestOf(c))
	}
	// Check media moves old.png, which nothing refers to, to the trash
	// (cat.png was added in this session, so it counts as used).
	old := manifestOf(c)["old.png"]
	if n, err := c.DeleteUnusedMedia(); err != nil || n != 1 {
		t.Fatalf("DeleteUnusedMedia = %d, %v", n, err)
	}
	trashed := manifestOf(c)["old.png"]
	if !trashed.Trashed || trashed.SHA1 != old.SHA1 || trashed.Stamp <= added.Stamp || trashed.Seq <= added.Seq {
		t.Fatalf("trashed: %+v after %+v", trashed, old)
	}
	if n, err := c.RestoreMediaTrash(); err != nil || n != 1 {
		t.Fatalf("RestoreMediaTrash = %d, %v", n, err)
	}
	restored := manifestOf(c)["old.png"]
	if restored.Trashed || restored.SHA1 != old.SHA1 || restored.Stamp <= trashed.Stamp {
		t.Fatalf("restored: %+v after %+v", restored, trashed)
	}
	// Imported files are part of the import's op.
	b := &ImportBatch{Media: map[string][]byte{"dog.png": []byte("woof")}}
	if _, err := c.Import(b, nil); err != nil {
		t.Fatal(err)
	}
	if e := manifestOf(c)["dog.png"]; e.SHA1 != sha1Hex([]byte("woof")) || e.Stamp <= restored.Stamp {
		t.Fatalf("imported: %+v", e)
	}
	if after, _ := c.LastStep(); after != steps+1 { // only the import
		t.Fatalf("%d undo steps, want 1", after-steps)
	}
	reopen(t, c, store, clk)
}

// A restore replaces the media, so the manifest is built again later.
func TestMediaManifestDroppedByRestore(t *testing.T) {
	c, _, _ := testCollection(t)
	noErr(t, c.BuildMediaManifest())
	snap, _ := c.Snapshot()
	if err := c.Restore(snap, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if c.HasMediaManifest() {
		t.Fatal("manifest kept by restore")
	}
}
