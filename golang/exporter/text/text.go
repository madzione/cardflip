// Package text exports notes and cards as tab-separated plain text
// (data-export spec). Notes keep their fields as stored, with Anki's header
// lines, so the file imports again into Cardflip or Anki; cards hold their
// rendered question and answer without HTML, for printing or reading
// elsewhere.
package text

import (
	"strconv"
	"strings"

	"cardflip/golang/core"
	"cardflip/golang/render"
)

// NoteOptions choose the columns added to the fields of each note.
type NoteOptions struct {
	Tags     bool
	Deck     bool
	NoteType bool
	GUID     bool
}

// WriteNotes writes the set's notes, one per line. Like Anki, the columns
// are the GUID, note type and deck (when chosen), the fields, padded to the
// largest field count so that the tags (when chosen) are always the last
// column. A note's deck is the deck of its first card.
func WriteNotes(set *core.ExportSet, o NoteOptions, yield func()) []byte {
	if yield == nil {
		yield = func() {}
	}
	types := map[core.ID]*core.NoteType{}
	fieldColumns := 0
	for i := range set.NoteTypes {
		nt := &set.NoteTypes[i]
		types[nt.ID] = nt
		fieldColumns = max(fieldColumns, len(nt.Fields))
	}
	for _, n := range set.Notes {
		fieldColumns = max(fieldColumns, len(n.Fields))
	}
	deckNames := map[core.ID]string{}
	for _, d := range set.Decks {
		deckNames[d.ID] = d.Name
	}
	firstCard := map[core.ID]core.Card{}
	for _, c := range set.Cards {
		if f, ok := firstCard[c.NoteID]; !ok || c.Ord < f.Ord {
			firstCard[c.NoteID] = c
		}
	}

	var sb strings.Builder
	sb.WriteString("#separator:tab\n#html:true\n")
	col := 0
	header := func(name string) {
		col++
		sb.WriteString("#" + name + " column:" + strconv.Itoa(col) + "\n")
	}
	if o.GUID {
		header("guid")
	}
	if o.NoteType {
		header("notetype")
	}
	if o.Deck {
		header("deck")
	}
	// prefix counts the columns before the fields.
	prefix := col
	if o.Tags {
		col += fieldColumns
		header("tags")
	}

	row := make([]string, 0, fieldColumns+4)
	for i, n := range set.Notes {
		if i%500 == 0 {
			yield()
		}
		row = row[:0]
		if o.GUID {
			row = append(row, n.GUID)
		}
		if o.NoteType {
			name := ""
			if nt, ok := types[n.TypeID]; ok {
				name = nt.Name
			}
			row = append(row, name)
		}
		if o.Deck {
			row = append(row, deckNames[firstCard[n.ID].DeckID])
		}
		row = append(row, n.Fields...)
		if o.Tags {
			for len(row) < prefix+fieldColumns {
				row = append(row, "")
			}
			row = append(row, strings.Join(n.Tags, " "))
		}
		writeRow(&sb, row)
	}
	return []byte(sb.String())
}

// WriteCards writes one line per card: its question and answer, rendered
// for display, as plain text. The answer leaves out the question that
// {{FrontSide}} repeats above the <hr id=answer> line.
func WriteCards(cards []render.Card, yield func()) []byte {
	if yield == nil {
		yield = func() {}
	}
	var sb strings.Builder
	sb.WriteString("#separator:tab\n#html:false\n")
	for i, c := range cards {
		if i%500 == 0 {
			yield()
		}
		writeRow(&sb, []string{PlainText(c.Question), PlainText(render.AnswerPart(c.Answer))})
	}
	return []byte(sb.String())
}

// PlainText turns rendered card HTML into one line of text: tags removed,
// entities decoded, and whitespace and line breaks collapsed into single
// spaces.
func PlainText(html string) string {
	return strings.Join(strings.Fields(render.StripHTML(html)), " ")
}

// writeRow writes a tab-separated line. A value is quoted, with its quotes
// doubled, when it holds a tab, quote or line break, or, in the first
// column, starts with "#", which would make the line a header or comment.
func writeRow(sb *strings.Builder, values []string) {
	for i, v := range values {
		if i > 0 {
			sb.WriteByte('\t')
		}
		if strings.ContainsAny(v, "\t\"\r\n") || i == 0 && strings.HasPrefix(v, "#") {
			sb.WriteByte('"')
			sb.WriteString(strings.ReplaceAll(v, `"`, `""`))
			sb.WriteByte('"')
		} else {
			sb.WriteString(v)
		}
	}
	sb.WriteByte('\n')
}
