package core

import (
	"fmt"
	"sort"

	"cardflip/golang/render"
)

// TemplatePreview is a card type rendered from unsaved template text.
type TemplatePreview struct {
	render.Card
	// ClozeNumbers are the cloze numbers the previewed note has cards for,
	// for cloze note types; empty otherwise.
	ClozeNumbers []int
}

// PreviewTemplate renders the card type t of note type typeID with css,
// using unsaved text, for note noteID or, when noteID is 0, for placeholder
// fields "(Name)". ord is the card type index for standard note types and
// the cloze number for cloze note types, where 0 or a number the note has
// no card for picks its first. The collection is not changed.
func (c *Collection) PreviewTemplate(typeID ID, t CardTemplate, css string, noteID ID, ord int, env render.Env) (TemplatePreview, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nt, ok := c.noteTypes[typeID]
	if !ok {
		return TemplatePreview{}, userErr("noteTypeGone", "That note type no longer exists.")
	}
	rn := render.Note{NoteType: nt.Name, LatexSVG: nt.LatexSVG, HasMedia: c.store.HasMedia}
	if noteID == 0 {
		rn.Fields = make(map[string]string, len(nt.Fields))
		for _, name := range nt.Fields {
			rn.Fields[name] = "(" + name + ")"
		}
	} else {
		n, ok := c.notes[noteID]
		if !ok || n.TypeID != typeID {
			return TemplatePreview{}, userErr("noteGone", "That note no longer exists.")
		}
		rn.Fields = fieldMap(nt, n.Fields)
		rn.Tags = n.Tags
	}
	var out TemplatePreview
	tmpl := render.Template{QFmt: t.QFmt, AFmt: t.AFmt}
	clozeOrd := 0
	if nt.Kind == Cloze {
		if noteID != 0 {
			out.ClozeNumbers = render.ClozeCards(t.QFmt, rn.Fields)
		}
		if len(out.ClozeNumbers) == 0 {
			out.ClozeNumbers = []int{1}
		}
		clozeOrd = out.ClozeNumbers[0]
		for _, n := range out.ClozeNumbers {
			if n == ord {
				clozeOrd = ord
			}
		}
		rn.CardOrd = clozeOrd
		rn.CardName = fmt.Sprintf("Cloze %d", clozeOrd)
	} else {
		rn.CardOrd = ord + 1
		rn.CardName = t.Name
	}
	if noteID != 0 {
		rn.Deck = c.decks[c.previewDeck(noteID, rn.CardOrd-1)].Name
	}
	out.Card = render.RenderCard(tmpl, css, rn, clozeOrd, env)
	return out, nil
}

// previewDeck is the deck of the note's card with ord, else the deck its
// new cards go to.
func (c *Collection) previewDeck(nid ID, ord int) ID {
	for _, id := range c.cardsByNote[nid] {
		if c.cards[id].Ord == ord {
			return c.cards[id].HomeDeck()
		}
	}
	return c.deckOfNote(nid)
}

// NoteLabel names a note in a list: its first field as plain text.
type NoteLabel struct {
	ID    ID
	Label string
}

// RecentNotes returns up to limit notes of note type typeID, most recently
// changed first.
func (c *Collection) RecentNotes(typeID ID, limit int) []NoteLabel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var notes []*Note
	for _, n := range c.notes {
		if n.TypeID == typeID {
			notes = append(notes, n)
		}
	}
	sort.Slice(notes, func(i, j int) bool {
		if !notes[i].Modified.Equal(notes[j].Modified) {
			return notes[i].Modified.After(notes[j].Modified)
		}
		return notes[i].ID > notes[j].ID
	})
	if len(notes) > limit {
		notes = notes[:limit]
	}
	out := make([]NoteLabel, len(notes))
	for i, n := range notes {
		label := ""
		if len(n.Fields) > 0 {
			label = collapseSpace(render.StripHTML(n.Fields[0]))
		}
		out[i] = NoteLabel{ID: n.ID, Label: label}
	}
	return out
}
