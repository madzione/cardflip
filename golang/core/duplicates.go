package core

import (
	"cmp"
	"slices"
	"strings"
)

// dupKey identifies notes that duplicate each other: the same note type and
// the same normalised first field.
type dupKey struct {
	typeID ID
	first  string
}

// duplicateText is what the duplicate warning compares: the first field,
// HTML-stripped, lower-cased and trimmed. "" means it never counts.
func duplicateText(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	return duplicateKey(fields[0])
}

// duplicateKey is a field's content as duplicates compare it: HTML-stripped,
// lower-cased and trimmed.
func duplicateKey(field string) string {
	return strings.ToLower(strings.TrimSpace(stripHTML(nfc(field))))
}

// indexFirstField updates the first-field index for note n replacing old
// (either may be nil). Notes with an empty first field are not indexed.
// Callers hold the write lock (or are in Open).
func (c *Collection) indexFirstField(old, n *Note) {
	var oldKey, newKey dupKey
	if old != nil {
		oldKey = dupKey{old.TypeID, duplicateText(old.Fields)}
	}
	if n != nil {
		newKey = dupKey{n.TypeID, duplicateText(n.Fields)}
	}
	if old != nil && n != nil && oldKey == newKey {
		return
	}
	if old != nil && oldKey.first != "" {
		ids := c.firstFields[oldKey]
		for i, id := range ids {
			if id == old.ID {
				ids = append(ids[:i:i], ids[i+1:]...)
				break
			}
		}
		if len(ids) == 0 {
			delete(c.firstFields, oldKey)
		} else {
			c.firstFields[oldKey] = ids
		}
	}
	if n != nil && newKey.first != "" {
		c.firstFields[newKey] = append(c.firstFields[newKey], n.ID)
	}
}

// Duplicates returns the notes of note type typeID whose first field equals
// firstField, ignoring HTML, case and surrounding whitespace, oldest first.
// Note exclude (the one being edited) is left out; an empty first field has
// no duplicates.
func (c *Collection) Duplicates(typeID ID, firstField string, exclude ID) []ID {
	key := dupKey{typeID, duplicateText([]string{firstField})}
	if key.first == "" {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []ID
	for _, id := range c.firstFields[key] {
		if id != exclude {
			out = append(out, id)
		}
	}
	sortIDs(out)
	return out
}

// DuplicateSet is a set of notes whose field has the same content.
type DuplicateSet struct {
	// Content is the field of the oldest note, HTML-stripped and trimmed.
	Content string
	// Notes are the notes, oldest first.
	Notes []ID
}

// FindDuplicates returns the sets of notes whose field fieldName has the
// same content, ignoring HTML, case and surrounding whitespace, among the
// notes of note types with a field of that name and, unless query is
// empty, a card matching query. Empty fields are never duplicates. Sets
// with more notes come first, then by content. An invalid query returns a
// *search.Error naming the bad term.
func (c *Collection) FindDuplicates(fieldName, query string) ([]DuplicateSet, error) {
	now := c.now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	var notes []ID
	if strings.TrimSpace(query) == "" {
		for id := range c.notes {
			notes = append(notes, id)
		}
	} else {
		cards, err := c.match(query, now)
		if err != nil {
			return nil, err
		}
		seen := make(map[ID]bool, len(cards))
		for _, id := range cards {
			if n := c.cards[id].NoteID; !seen[n] {
				seen[n] = true
				notes = append(notes, n)
			}
		}
	}
	sortIDs(notes)
	ord := make(map[ID]int, len(c.noteTypes))
	for id, nt := range c.noteTypes {
		ord[id] = slices.Index(nt.Fields, fieldName)
	}
	groups := map[string]*DuplicateSet{}
	for i, id := range notes {
		if i%200 == 199 {
			c.yield()
		}
		n := c.notes[id]
		j := ord[n.TypeID]
		if j < 0 || j >= len(n.Fields) {
			continue
		}
		key := duplicateKey(n.Fields[j])
		if key == "" {
			continue
		}
		g := groups[key]
		if g == nil {
			g = &DuplicateSet{Content: strings.TrimSpace(stripHTML(n.Fields[j]))}
			groups[key] = g
		}
		g.Notes = append(g.Notes, id)
	}
	out := []DuplicateSet{}
	for _, g := range groups {
		if len(g.Notes) > 1 {
			out = append(out, *g)
		}
	}
	slices.SortFunc(out, func(a, b DuplicateSet) int {
		return cmp.Or(cmp.Compare(len(b.Notes), len(a.Notes)),
			strings.Compare(strings.ToLower(a.Content), strings.ToLower(b.Content)), cmp.Compare(a.Notes[0], b.Notes[0]))
	})
	return out, nil
}
