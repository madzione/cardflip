package core

import (
	"strings"

	"cardflip/golang/render"
)

// BuiltinNames are the names that built-in things get when they are
// created: the first deck and the default preset of a new collection, and
// the stock note types. The app supplies them in the interface language,
// since Go has no translations; a name that is missing or not allowed
// falls back to English.
type BuiltinNames struct {
	Deck   string `json:"deck"`
	Preset string `json:"preset"`
	// NoteTypes names the stock note types; missing ones stay English.
	NoteTypes []StockNames `json:"noteTypes"`
}

// StockNames names one stock note type, its fields and its card types, in
// the order of the English stock note type.
type StockNames struct {
	// Stock is the English name that identifies it, e.g. StockBasic.
	Stock     string   `json:"stock"`
	Name      string   `json:"name"`
	Fields    []string `json:"fields"`
	CardTypes []string `json:"cardTypes"`
}

// deckName returns the name of a new collection's deck.
func (n BuiltinNames) deckName() string {
	if name, err := normalizeDeckName(n.Deck); err == nil && !strings.Contains(name, DeckSeparator) {
		return name
	}
	return DefaultDeckName
}

// presetName returns the name of a new collection's preset.
func (n BuiltinNames) presetName() string {
	if name := strings.TrimSpace(nfc(n.Preset)); name != "" {
		return name
	}
	return DefaultPresetName
}

// stockNoteType returns stock note type stock (an English stock name)
// named in these names, or nil when there is no such stock note type.
func (n BuiltinNames) stockNoteType(stock string) *NoteType {
	nt := stockNoteType(stock)
	if nt == nil {
		return nil
	}
	for _, s := range n.NoteTypes {
		if s.Stock == stock {
			if named, ok := s.apply(nt); ok {
				return named
			}
		}
	}
	return nt
}

// noteTypeName returns the name stock note type stock gets.
func (n BuiltinNames) noteTypeName(stock string) string {
	if nt := n.stockNoteType(stock); nt != nil {
		return nt.Name
	}
	return stock
}

// apply returns English stock note type nt with these names, or false
// when they do not fit it: a name is missing, invalid or used twice.
func (s StockNames) apply(nt *NoteType) (*NoteType, bool) {
	if len(s.Fields) != len(nt.Fields) || len(s.CardTypes) != len(nt.Templates) {
		return nil, false
	}
	name := strings.TrimSpace(nfc(s.Name))
	if name == "" {
		return nil, false
	}
	fields, ok := validNames("field", s.Fields)
	if !ok {
		return nil, false
	}
	for _, f := range fields {
		if reservedField(f) != nil {
			return nil, false
		}
	}
	cardTypes, ok := validNames("card type", s.CardTypes)
	if !ok {
		return nil, false
	}
	renamed := map[string]string{}
	for i, f := range nt.Fields {
		renamed[f] = fields[i]
	}
	out := cloneNoteType(nt)
	out.Name, out.Fields = name, fields
	for i := range out.Templates {
		t := &out.Templates[i]
		t.Name = cardTypes[i]
		t.QFmt = render.RenameFields(t.QFmt, renamed)
		t.AFmt = render.RenameFields(t.AFmt, renamed)
	}
	return out, true
}

// validNames checks names with validName and that no two are the same,
// ignoring case.
func validNames(kind string, names []string) ([]string, bool) {
	out := make([]string, len(names))
	seen := map[string]bool{}
	for i, n := range names {
		v, err := validName(kind, n)
		if err != nil || seen[strings.ToLower(v)] {
			return nil, false
		}
		seen[strings.ToLower(v)] = true
		out[i] = v
	}
	return out, true
}

// builtinNames returns the names set with SetBuiltinNames.
func (c *Collection) builtinNames() BuiltinNames {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.builtin
}

// SetBuiltinNames sets the names that built-in things are created with
// from now on. Existing decks, presets and note types keep their names.
func (c *Collection) SetBuiltinNames(names BuiltinNames) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.builtin = names
}

// BasicNoteType returns the note type named like the stock note type
// Basic, in the built-in names or else in English, which pickers fall back
// to.
func (c *Collection) BasicNoteType() (ID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, name := range []string{c.builtin.noteTypeName(StockBasic), StockBasic} {
		for _, nt := range c.noteTypes {
			if nt.Name == name {
				return nt.ID, true
			}
		}
	}
	return 0, false
}
