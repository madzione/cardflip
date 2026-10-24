package core

import (
	"bytes"
	"maps"
	"slices"

	"golang.org/x/text/unicode/norm"
)

// nfc returns s in Unicode NFC, the form stored text is in (collection
// spec, "Unicode normalisation"). Text already in NFC is returned as is.
func nfc(s string) string { return norm.NFC.String(s) }

// nfcAll returns ss with every string in NFC: ss itself when all are, else
// a new slice, so a slice shared with a stored object is never changed.
func nfcAll(ss []string) []string {
	i := slices.IndexFunc(ss, func(s string) bool { return !norm.NFC.IsNormalString(s) })
	if i < 0 {
		return ss
	}
	out := slices.Clone(ss)
	for ; i < len(out); i++ {
		out[i] = nfc(out[i])
	}
	return out
}

// nfcNote returns n with its fields and tags in NFC: n itself when they
// are, else a changed copy.
func nfcNote(n *Note) *Note {
	fields, tags := nfcAll(n.Fields), nfcAll(n.Tags)
	if sameSlice(fields, n.Fields) && sameSlice(tags, n.Tags) {
		return n
	}
	nn := *n
	nn.Fields = fields
	if !sameSlice(tags, n.Tags) {
		nn.Tags = NormalizeTags(tags)
	}
	return &nn
}

// nfcDeck returns d with its name and description in NFC: d itself when
// they are, else a changed copy.
func nfcDeck(d *Deck) *Deck {
	if norm.NFC.IsNormalString(d.Name) && norm.NFC.IsNormalString(d.Description) {
		return d
	}
	nd := *d
	nd.Name, nd.Description = nfc(d.Name), nfc(d.Description)
	return &nd
}

// nfcNoteType returns nt with its name, field names, templates and CSS in
// NFC: nt itself when they are, else a changed copy. Templates refer to
// fields and media by name, so they are normalised with them.
func nfcNoteType(nt *NoteType) *NoteType {
	text := []string{nt.Name, nt.CSS}
	for _, t := range nt.Templates {
		text = append(text, t.Name, t.QFmt, t.AFmt, t.BrowserQFmt, t.BrowserAFmt)
	}
	if sameSlice(nfcAll(text), text) && sameSlice(nfcAll(nt.Fields), nt.Fields) {
		return nt
	}
	cp := cloneNoteType(nt)
	cp.Name, cp.CSS = nfc(nt.Name), nfc(nt.CSS)
	cp.Fields = nfcAll(cp.Fields)
	for i := range cp.Templates {
		cp.Templates[i] = nfcTemplate(cp.Templates[i])
	}
	return cp
}

// nfcTemplate returns t with its name and formats in NFC.
func nfcTemplate(t CardTemplate) CardTemplate {
	t.Name, t.QFmt, t.AFmt = nfc(t.Name), nfc(t.QFmt), nfc(t.AFmt)
	t.BrowserQFmt, t.BrowserAFmt = nfc(t.BrowserQFmt), nfc(t.BrowserAFmt)
	return t
}

// sameSlice reports whether a and b are the same slice (not merely equal).
func sameSlice(a, b []string) bool {
	return len(a) == len(b) && (len(a) == 0 || &a[0] == &b[0])
}

// nfcOp puts the text of the objects o adds or changes in NFC, replacing
// those that are not with normalised copies. Entry points normalise their
// input already, so that names are compared in NFC; this makes sure that
// nothing reaches storage in another form.
func nfcOp(o *op) {
	for i, n := range o.Notes {
		o.Notes[i] = nfcNote(n)
	}
	for i, d := range o.Decks {
		o.Decks[i] = nfcDeck(d)
	}
	for i, nt := range o.NoteTypes {
		o.NoteTypes[i] = nfcNoteType(nt)
	}
}

// nfcMediaRenames returns the new name of each file in names whose name is
// not in NFC: its NFC form, or, when a different file has that name, the
// NFC form with a hash of the content (see hashedMediaName). Of those,
// clashes are the ones whose new name is not the NFC form: references to
// them must be rewritten before the text is normalised, since afterwards
// they would name the other file. A file whose NFC name holds the same
// content is renamed onto it.
func nfcMediaRenames(names []string, content func(string) ([]byte, error)) (renames, clashes map[string]string, err error) {
	names = slices.Sorted(slices.Values(names))
	// owner is the file that holds a name once the files are renamed.
	owner := map[string]string{}
	for _, name := range names {
		if norm.NFC.IsNormalString(name) {
			owner[name] = name
		}
	}
	renames, clashes = map[string]string{}, map[string]string{}
	for _, name := range names {
		if norm.NFC.IsNormalString(name) {
			continue
		}
		to := nfc(name)
		if other, ok := owner[to]; ok {
			data, err := content(name)
			if err != nil {
				return nil, nil, err
			}
			otherData, err := content(other)
			if err != nil {
				return nil, nil, err
			}
			if !bytes.Equal(data, otherData) {
				to = hashedMediaName(to, data)
				clashes[name] = to
			}
		}
		if _, ok := owner[to]; !ok {
			owner[to] = name
		}
		renames[name] = to
	}
	return renames, clashes, nil
}

// nfcBatch returns b with its text and media names in NFC (collection
// spec, "Unicode normalisation"), sharing what needs no change; b itself
// is not changed. A file whose NFC name the batch uses for a different
// file gets a hashed name, and the references to it are rewritten.
func nfcBatch(b *ImportBatch) *ImportBatch {
	nb := *b
	var clashes map[string]string
	nb.Media, clashes = nfcMedia(b.Media)
	// text normalises HTML that may refer to media.
	text := func(s string) string { return nfc(rewriteMedia(s, clashes)) }

	nb.NoteTypes = slices.Clone(b.NoteTypes)
	for i := range nb.NoteTypes {
		it := &nb.NoteTypes[i]
		nt := cloneNoteType(&it.Type)
		nt.CSS = text(nt.CSS)
		for j := range nt.Templates {
			t := &nt.Templates[j]
			t.QFmt, t.AFmt = text(t.QFmt), text(t.AFmt)
			t.BrowserQFmt, t.BrowserAFmt = text(t.BrowserQFmt), text(t.BrowserAFmt)
		}
		it.Type = *nfcNoteType(nt)
		it.TemplateDecks = nfcAll(it.TemplateDecks)
	}
	nb.Notes = slices.Clone(b.Notes)
	for i := range nb.Notes {
		in := &nb.Notes[i]
		fields := in.Fields
		if len(clashes) > 0 {
			fields = slices.Clone(fields)
			for j, f := range fields {
				fields[j] = rewriteMedia(f, clashes)
			}
		}
		in.Fields, in.Tags, in.DeckName = nfcAll(fields), nfcAll(in.Tags), nfc(in.DeckName)
		if slices.ContainsFunc(in.Cards, func(c ImportCard) bool { return !norm.NFC.IsNormalString(c.DeckName) }) {
			in.Cards = slices.Clone(in.Cards)
			for j := range in.Cards {
				in.Cards[j].DeckName = nfc(in.Cards[j].DeckName)
			}
		}
	}
	nb.Decks = nfcAll(b.Decks)
	if b.DeckPresets != nil {
		nb.DeckPresets = make(map[string]string, len(b.DeckPresets))
		for name, key := range b.DeckPresets {
			nb.DeckPresets[nfc(name)] = key
		}
	}
	if b.DeckDescriptions != nil {
		nb.DeckDescriptions = make(map[string]string, len(b.DeckDescriptions))
		for name, desc := range b.DeckDescriptions {
			nb.DeckDescriptions[nfc(name)] = text(desc)
		}
	}
	return &nb
}

// nfcMedia returns media with its names in NFC (see nfcMediaRenames):
// media itself when they are, else a new map. The clashes are the files
// that got a hashed name.
func nfcMedia(media map[string][]byte) (_ map[string][]byte, clashes map[string]string) {
	renames, clashes, _ := nfcMediaRenames(slices.Collect(maps.Keys(media)), func(name string) ([]byte, error) {
		return media[name], nil
	})
	if len(renames) == 0 {
		return media, nil
	}
	out := make(map[string][]byte, len(media))
	for name, data := range media {
		if _, ok := renames[name]; !ok {
			out[name] = data
		}
	}
	for from, to := range renames {
		out[to] = media[from]
	}
	return out, clashes
}

// nfcState puts the text of s in NFC (see nfcOp), after rewriting the
// references to the media files in clashes (see nfcMediaRenames), and
// marks it normalised. Changed objects are replaced by copies.
func nfcState(s *state, clashes map[string]string) {
	o := op{Notes: s.Notes, Decks: s.Decks, NoteTypes: s.NoteTypes}
	if len(clashes) > 0 {
		rewriteOpMedia(&o, clashes)
	}
	nfcOp(&o)
	s.Normalized = true
}

// rewriteOpMedia rewrites the references to renamed media files in the
// objects of o (see rewriteMedia), replacing changed ones by copies.
func rewriteOpMedia(o *op, renames map[string]string) {
	rewrite := func(ss []string) ([]string, bool) {
		out := make([]string, len(ss))
		changed := false
		for i, s := range ss {
			out[i] = rewriteMedia(s, renames)
			changed = changed || out[i] != s
		}
		return out, changed
	}
	for i, n := range o.Notes {
		if fields, ok := rewrite(n.Fields); ok {
			nn := cloneNote(n)
			nn.Fields = fields
			o.Notes[i] = &nn
		}
	}
	for i, d := range o.Decks {
		if desc := rewriteMedia(d.Description, renames); desc != d.Description {
			nd := *d
			nd.Description = desc
			o.Decks[i] = &nd
		}
	}
	for i, nt := range o.NoteTypes {
		text := []string{nt.CSS}
		for _, t := range nt.Templates {
			text = append(text, t.QFmt, t.AFmt, t.BrowserQFmt, t.BrowserAFmt)
		}
		text, ok := rewrite(text)
		if !ok {
			continue
		}
		cp := cloneNoteType(nt)
		cp.CSS = text[0]
		for j := range cp.Templates {
			t := &cp.Templates[j]
			t.QFmt, t.AFmt, t.BrowserQFmt, t.BrowserAFmt = text[1+4*j], text[2+4*j], text[3+4*j], text[4+4*j]
		}
		o.NoteTypes[i] = cp
	}
}

// migrateNFC normalises a collection saved by an earlier version (collection
// spec, "Unicode normalisation"): media files whose names are not in NFC
// are stored again under their new names, and the text of notes, decks and
// note types is put in NFC in one op. Like the other migrations it is
// applied in memory, not journaled, and saved by the compaction that
// follows in Open; until then the files are also kept under their old
// names, which it returns for the caller to trash. Note modification times
// are kept: the content the user sees does not change.
func (c *Collection) migrateNFC() ([]string, error) {
	var names []string
	data := map[string][]byte{}
	err := c.store.EachMedia(func(name string, d []byte) error {
		names = append(names, name)
		if !norm.NFC.IsNormalString(name) {
			data[name] = d
		}
		c.yield()
		return nil
	})
	if err != nil {
		return nil, err
	}
	renames, clashes, err := nfcMediaRenames(names, func(name string) ([]byte, error) {
		if d, ok := data[name]; ok {
			return d, nil
		}
		return c.store.GetMedia(name)
	})
	if err != nil {
		return nil, err
	}
	old := slices.Sorted(maps.Keys(renames))
	for _, from := range old {
		if err := c.store.PutMedia(renames[from], data[from]); err != nil {
			return nil, err
		}
		c.yield()
	}

	s := c.state()
	nfcState(s, clashes)
	o := &op{Kind: "normalize"}
	for _, n := range s.Notes {
		if n != c.notes[n.ID] {
			o.Notes = append(o.Notes, n)
		}
	}
	for _, d := range s.Decks {
		if d != c.decks[d.ID] {
			o.Decks = append(o.Decks, d)
		}
	}
	for _, nt := range s.NoteTypes {
		if nt != c.noteTypes[nt.ID] {
			o.NoteTypes = append(o.NoteTypes, nt)
		}
	}
	c.apply(o)
	c.mediaIndex = nil
	return old, nil
}
