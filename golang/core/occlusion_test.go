package core

import (
	"strings"
	"testing"

	"cardflip/golang/render"
)

// ioField builds an Occlusion field: one rectangle per number in ords, in
// "Hide all, guess one" mode, and a text label without a card.
func ioField(ords ...int) string {
	var parts []string
	for i, ord := range ords {
		parts = append(parts, "{{c"+string(rune('0'+ord))+"::image-occlusion:rect:left=."+string(rune('1'+i))+
			":top=.1:width=.1:height=.1:oi=1}}")
	}
	parts = append(parts, `{{c0::image-occlusion:text:left=.5:top=.5:text=Aorta:scale=1:fs=.05:oi=1}}`)
	return strings.Join(parts, "<br>")
}

func addOcclusion(t *testing.T, c *Collection, ords ...int) ID {
	t.Helper()
	return addNote(t, c, StockImageOcclusion, ioField(ords...), `<img src="heart.png">`, "Heart", "", "")
}

// Three labels (collection spec): one card per shape.
func TestImageOcclusionCards(t *testing.T) {
	c, _, _ := testCollection(t)
	if nt := noteType(t, c, StockImageOcclusion); !nt.IsImageOcclusion() ||
		strings.Join(nt.Fields, ",") != "Occlusion,Image,Header,Back Extra,Comments" {
		t.Fatalf("Image Occlusion = %+v", nt)
	}
	nid := addOcclusion(t, c, 1, 2, 3)
	cards := c.CardsOfNote(nid)
	if len(cards) != 3 {
		t.Fatalf("%d cards, want 3", len(cards))
	}
	r, err := c.RenderCard(cards[1].ID, render.Env{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<cf-occlusion image="heart.png"`, `active="2" side="q"`, "--active-shape-color:#ff8e8e", "Heart"} {
		if !strings.Contains(r.Question, want) {
			t.Errorf("question lacks %q: %s", want, r.Question)
		}
	}
	if strings.Contains(r.Question, "<script") || strings.Contains(r.Question, "<img") {
		t.Errorf("question = %s", r.Question)
	}
}

// Grouped shapes (collection spec): shapes sharing a number are one card.
func TestImageOcclusionGroups(t *testing.T) {
	c, _, _ := testCollection(t)
	nid := addOcclusion(t, c, 1, 2, 2)
	cards := c.CardsOfNote(nid)
	if len(cards) != 2 {
		t.Fatalf("%d cards, want 2", len(cards))
	}
	r, err := c.RenderCard(cards[1].ID, render.Env{})
	if err != nil {
		t.Fatal(err)
	}
	// The card of the group lists both of its shapes as number 2.
	if n := strings.Count(r.Question, "&#34;ord&#34;:2"); n != 2 {
		t.Errorf("%d shapes of group 2 in %s", n, r.Question)
	}
}

func TestImageOcclusionNoteTypeOnDemand(t *testing.T) {
	c, _, _ := testCollection(t)
	builtin := noteType(t, c, StockImageOcclusion)
	if id, err := c.ImageOcclusionNoteType(); err != nil || id != builtin.ID {
		t.Fatalf("built-in: %v, %v; want %v", id, err, builtin.ID)
	}

	// A collection from before the note type was built in.
	if err := c.DeleteNoteType(builtin.ID); err != nil {
		t.Fatal(err)
	}
	created, err := c.ImageOcclusionNoteType()
	if err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(created); nt.Name != StockImageOcclusion || !nt.IsImageOcclusion() {
		t.Fatalf("created %+v", nt)
	}
	if again, _ := c.ImageOcclusionNoteType(); again != created {
		t.Fatalf("second use created another note type")
	}

	// The name taken by a different note type: a numbered name.
	if err := c.DeleteNoteType(created); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddNoteType(StockBasic, "image occlusion"); err != nil {
		t.Fatal(err)
	}
	numbered, err := c.ImageOcclusionNoteType()
	if err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(numbered); nt.Name != "Image Occlusion 2" || !nt.IsImageOcclusion() {
		t.Fatalf("numbered %+v", nt)
	}
	if again, _ := c.ImageOcclusionNoteType(); again != numbered {
		t.Fatalf("second use created another note type")
	}
}

func TestIsImageOcclusion(t *testing.T) {
	c, _, _ := testCollection(t)
	io := noteType(t, c, StockImageOcclusion)
	clone, err := c.CloneNoteType(io.ID, "Anatomy masks")
	if err != nil {
		t.Fatal(err)
	}
	if nt, _ := c.NoteType(clone); !nt.IsImageOcclusion() {
		t.Errorf("clone is not Image Occlusion")
	}
	for _, name := range []string{StockBasic, StockCloze} {
		if nt := noteType(t, c, name); nt.IsImageOcclusion() {
			t.Errorf("%s is Image Occlusion", name)
		}
	}
	// A cloze note type with the fields, but not passing Occlusion through
	// cloze:, is not.
	other := io
	other.Templates = []CardTemplate{{Name: "x", QFmt: "{{cloze:Header}}{{Occlusion}}{{Image}}", AFmt: ""}}
	if other.IsImageOcclusion() {
		t.Errorf("template without cloze:Occlusion is Image Occlusion")
	}
}
