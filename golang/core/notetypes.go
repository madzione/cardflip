package core

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"cardflip/golang/render"
)

const builtinCSS = `.card {
  font-family: arial, sans-serif;
  font-size: 20px;
  line-height: 1.5;
  text-align: center;
}
`

const builtinClozeCSS = builtinCSS + `.cloze {
  font-weight: bold;
  color: #3b82f6;
}
`

// Anki's Image Occlusion card (26.09), unchanged so notes render in both
// apps: Anki draws the shapes with the script, Cardflip natively (see
// render.OcclusionShape). The CSS variables are the shape colours. The
// template and the variables are Anki's, under the AGPL version 3 or later,
// see THIRD_PARTY_NOTICES.md.
const (
	occlusionFront = `{{#Header}}<div>{{Header}}</div>{{/Header}}
<div style="display: none">{{cloze:Occlusion}}</div>
<div id="err"></div>
<div id="image-occlusion-container">
    {{Image}}
    <canvas id="image-occlusion-canvas"></canvas>
</div>
<script>
try {
    anki.imageOcclusion.setup();
} catch (exc) {
    document.getElementById("err").innerHTML = ` + "`Error loading image occlusion. Is your Anki version up to date?<br><br>${exc}`" + `;
}
</script>
`
	occlusionBack = occlusionFront + `
<div><button id="toggle">Toggle Masks</button></div>
{{#Back Extra}}<div>{{Back Extra}}</div>{{/Back Extra}}
`
	occlusionCSS = `#image-occlusion-canvas {
    --inactive-shape-color: #ffeba2;
    --active-shape-color: #ff8e8e;
    --inactive-shape-border: 1px #212121;
    --active-shape-border: 1px #212121;
    --highlight-shape-color: #ff8e8e00;
    --highlight-shape-border: 1px #ff8e8e;
}

` + builtinCSS
)

const (
	basicAnswer   = "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}"
	reverseAnswer = "{{FrontSide}}\n\n<hr id=answer>\n\n{{Front}}"
)

// Stock note type names.
const (
	StockBasic            = "Basic"
	StockReversed         = "Basic (and reversed card)"
	StockOptionalReversed = "Basic (optional reversed card)"
	StockTypeIn           = "Basic (type in the answer)"
	StockCloze            = "Cloze"
	StockImageOcclusion   = "Image Occlusion"
)

// stockNoteTypes returns the note types the user can add, named after
// themselves and without IDs.
func stockNoteTypes() []*NoteType {
	return []*NoteType{
		{
			Name: StockBasic, Kind: Standard, Fields: []string{"Front", "Back"}, CSS: builtinCSS,
			Templates: []CardTemplate{{Name: "Card 1", QFmt: "{{Front}}", AFmt: basicAnswer}},
		},
		{
			Name: StockReversed, Kind: Standard, Fields: []string{"Front", "Back"}, CSS: builtinCSS,
			Templates: []CardTemplate{
				{Name: "Card 1", QFmt: "{{Front}}", AFmt: basicAnswer},
				{Name: "Card 2", QFmt: "{{Back}}", AFmt: reverseAnswer},
			},
		},
		{
			Name: StockOptionalReversed, Kind: Standard, Fields: []string{"Front", "Back", "Add Reverse"}, CSS: builtinCSS,
			Templates: []CardTemplate{
				{Name: "Card 1", QFmt: "{{Front}}", AFmt: basicAnswer},
				{Name: "Card 2", QFmt: "{{#Add Reverse}}{{Back}}{{/Add Reverse}}", AFmt: reverseAnswer},
			},
		},
		{
			Name: StockTypeIn, Kind: Standard, Fields: []string{"Front", "Back"}, CSS: builtinCSS,
			Templates: []CardTemplate{{Name: "Card 1", QFmt: "{{Front}}\n\n{{type:Back}}",
				AFmt: "{{Front}}\n\n<hr id=answer>\n\n{{type:Back}}"}},
		},
		{
			Name: StockCloze, Kind: Cloze, Fields: []string{"Text", "Back Extra"}, CSS: builtinClozeCSS,
			Templates: []CardTemplate{{Name: "Cloze", QFmt: "{{cloze:Text}}", AFmt: "{{cloze:Text}}<br>\n{{Back Extra}}"}},
		},
		{
			// As in Anki: one card per shape number in Occlusion.
			Name: StockImageOcclusion, Kind: Cloze, Fields: []string{"Occlusion", "Image", "Header", "Back Extra", "Comments"},
			CSS: occlusionCSS, Templates: []CardTemplate{{Name: "Image Occlusion", QFmt: occlusionFront, AFmt: occlusionBack}},
		},
	}
}

// StockNoteTypes lists the names of the stock note types, in menu order.
func StockNoteTypes() []string {
	var out []string
	for _, nt := range stockNoteTypes() {
		out = append(out, nt.Name)
	}
	return out
}

func stockNoteType(name string) *NoteType {
	for _, nt := range stockNoteTypes() {
		if nt.Name == name {
			return nt
		}
	}
	return nil
}

// builtinNoteTypes returns the note types of a new collection (without
// IDs), named in names.
func builtinNoteTypes(names BuiltinNames) []*NoteType {
	return []*NoteType{names.stockNoteType(StockBasic), names.stockNoteType(StockReversed),
		names.stockNoteType(StockCloze), names.stockNoteType(StockImageOcclusion)}
}

// IsImageOcclusion reports whether the note type is made like Anki's Image
// Occlusion (see OcclusionFields), so its notes are edited with the mask
// editor. Renamed and cloned ones count too.
func (nt *NoteType) IsImageOcclusion() bool {
	_, _, ok := nt.OcclusionFields()
	return ok
}

// occlusionContainer is where Anki's Image Occlusion template shows the
// image.
const occlusionContainer = `id="image-occlusion-container"`

// OcclusionFields returns the indexes of the fields that hold an Image
// Occlusion note's shapes and its image, and false when the note type is
// not made like Anki's Image Occlusion: a cloze note type whose card passes
// the shapes field through cloze: and shows the image field first inside
// image-occlusion-container, as Anki's template does in every language.
// Note types with the fields Occlusion and Image and cloze:Occlusion count
// too, whatever their template.
func (nt *NoteType) OcclusionFields() (shapes, image int, ok bool) {
	if nt.Kind != Cloze || len(nt.Templates) == 0 {
		return 0, 0, false
	}
	qfmt := nt.Templates[0].QFmt
	clozes := render.ClozeFields(qfmt)
	if slices.Contains(clozes, "Occlusion") {
		shapes, image = slices.Index(nt.Fields, "Occlusion"), slices.Index(nt.Fields, "Image")
		if shapes >= 0 && image >= 0 {
			return shapes, image, true
		}
	}
	at := strings.Index(qfmt, occlusionContainer)
	if len(clozes) == 0 || at < 0 {
		return 0, 0, false
	}
	shapes = slices.Index(nt.Fields, clozes[0])
	image = -1
	if after := render.FieldsUsed(qfmt[at:]); len(after) > 0 {
		image = slices.Index(nt.Fields, after[0])
	}
	if shapes < 0 || image < 0 || shapes == image {
		return 0, 0, false
	}
	return shapes, image, true
}

// ImageOcclusionNoteType returns the note type that Image Occlusion notes
// are added with: the one with the built-in name of Image Occlusion (or the
// English one), which collections from before it was built in get on first
// use. When that name is taken by a different note type, "Image Occlusion
// 2" (or the next number) is used in the same way.
func (c *Collection) ImageOcclusionNoteType() (ID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	base := c.builtin.noteTypeName(StockImageOcclusion)
	if base != StockImageOcclusion {
		for _, nt := range c.noteTypes {
			if strings.EqualFold(nt.Name, StockImageOcclusion) && nt.IsImageOcclusion() {
				return nt.ID, nil
			}
		}
	}
	name := base
	for i := 2; ; i++ {
		var found *NoteType
		for _, nt := range c.noteTypes {
			if strings.EqualFold(nt.Name, name) {
				found = nt
			}
		}
		if found != nil && found.IsImageOcclusion() {
			return found.ID, nil
		}
		if found == nil {
			break
		}
		name = base + " " + strconv.Itoa(i)
	}
	return c.addNoteType("addNoteType", c.builtin.stockNoteType(StockImageOcclusion), name)
}

func cloneNoteType(nt *NoteType) *NoteType {
	n := *nt
	n.Fields = append([]string(nil), nt.Fields...)
	n.FieldOptions = append([]FieldOptions(nil), nt.FieldOptions...)
	n.Templates = append([]CardTemplate(nil), nt.Templates...)
	return &n
}

// NoteTypes returns all note types sorted by name.
func (c *Collection) NoteTypes() []NoteType {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]NoteType, 0, len(c.noteTypes))
	for _, nt := range c.noteTypes {
		out = append(out, *cloneNoteType(nt))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// NoteType returns one note type.
func (c *Collection) NoteType(id ID) (NoteType, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nt, ok := c.noteTypes[id]
	if !ok {
		return NoteType{}, userErr("noteTypeGone", "That note type no longer exists.")
	}
	return *cloneNoteType(nt), nil
}

// noteTypeName checks a note type name: non-empty and, ignoring case, not
// used by another note type than except. Callers hold the lock.
func (c *Collection) noteTypeName(name string, except ID) (string, error) {
	name = strings.TrimSpace(nfc(name))
	if name == "" {
		return "", userErr("noteTypeNameEmpty", "Please enter a note type name.")
	}
	for _, other := range c.noteTypes {
		if other.ID != except && strings.EqualFold(other.Name, name) {
			return "", userErr("noteTypeNameTaken", "A note type named %q already exists.", other.Name)
		}
	}
	return name, nil
}

// addNoteType saves nt as a new note type named name and returns its ID.
func (c *Collection) addNoteType(kind string, nt *NoteType, name string) (ID, error) {
	name, err := c.noteTypeName(name, 0)
	if err != nil {
		return 0, err
	}
	nt.ID, nt.Name, nt.Modified = c.ids.next(), name, c.now()
	if err := c.commit(&op{Kind: kind, NoteTypes: []*NoteType{nt}}); err != nil {
		return 0, err
	}
	return nt.ID, nil
}

// AddNoteType creates a note type named name from stock note type stock
// (an English stock name), with its fields and card types named in the
// built-in names.
func (c *Collection) AddNoteType(stock, name string) (ID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	nt := c.builtin.stockNoteType(stock)
	if nt == nil {
		return 0, userErr("unknownStockType", "Unknown stock note type %q.", stock)
	}
	return c.addNoteType("addNoteType", nt, name)
}

// CloneNoteType creates a note type named name with the kind, fields, card
// types and styling of note type id, and no notes. The clone is not linked
// to where the original was imported from.
func (c *Collection) CloneNoteType(id ID, name string) (ID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.noteTypes[id]
	if !ok {
		return 0, userErr("noteTypeGone", "That note type no longer exists.")
	}
	nt := cloneNoteType(old)
	nt.Source = ""
	return c.addNoteType("cloneNoteType", nt, name)
}

// NoteTypeDeletion describes what deleting a note type removes, for the
// confirmation.
type NoteTypeDeletion struct {
	Notes int `json:"notes"`
	Cards int `json:"cards"`
}

// NoteTypeDeletionInfo counts the notes and cards that DeleteNoteType(id)
// removes.
func (c *Collection) NoteTypeDeletionInfo(id ID) (NoteTypeDeletion, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.noteTypes[id]; !ok {
		return NoteTypeDeletion{}, userErr("noteTypeGone", "That note type no longer exists.")
	}
	var out NoteTypeDeletion
	for nid, n := range c.notes {
		if n.TypeID == id {
			out.Notes++
			out.Cards += len(c.cardsByNote[nid])
		}
	}
	return out, nil
}

// DeleteNoteType deletes a note type with all its notes and their cards.
// The last remaining note type cannot be deleted.
func (c *Collection) DeleteNoteType(id ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.noteTypes[id]; !ok {
		return userErr("noteTypeGone", "That note type no longer exists.")
	}
	if len(c.noteTypes) == 1 {
		return userErr("lastNoteType", "The last note type cannot be deleted. A collection always has at least one note type.")
	}
	o := &op{Kind: "deleteNoteType", DelNoteTypes: []ID{id}}
	for nid, n := range c.notes {
		if n.TypeID == id {
			o.DelNotes = append(o.DelNotes, nid)
			o.DelCards = append(o.DelCards, c.cardsByNote[nid]...)
		}
	}
	sortIDs(o.DelNotes)
	sortIDs(o.DelCards)
	if _, ok := c.settings.LastDeckByNoteType[id]; ok || c.settings.LastNoteTypeID == id {
		s := c.settings
		if s.LastNoteTypeID == id {
			s.LastNoteTypeID = 0
		}
		s.LastDeckByNoteType = maps.Clone(s.LastDeckByNoteType)
		delete(s.LastDeckByNoteType, id)
		o.Settings = &s
	}
	return c.commit(o)
}

// validName checks a field or template name: non-empty, and free of
// characters that have a meaning in templates.
func validName(kind, name string) (string, error) {
	name = strings.TrimSpace(nfc(name))
	// fieldName… or cardTypeName…
	code := "fieldName"
	if kind == "card type" {
		code = "cardTypeName"
	}
	if name == "" {
		return "", userErr(code+"Empty", "Please enter a "+kind+" name.")
	}
	if strings.ContainsAny(name, ":{}\"") || strings.ContainsAny(name[:1], "#^/") {
		return "", userErr(code+"Invalid", "A "+kind+" name cannot contain : { } \" or start with # ^ /.")
	}
	return name, nil
}

// NoteCount returns how many notes use a note type, for confirmations.
func (c *Collection) NoteCount(typeID ID) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, note := range c.notes {
		if note.TypeID == typeID {
			n++
		}
	}
	return n
}

// TemplateCardCount returns how many cards a template of a note type has,
// for the confirmation before removing it.
func (c *Collection) TemplateCardCount(typeID ID, index int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, card := range c.cards {
		if card.Ord == index && c.notes[card.NoteID].TypeID == typeID {
			n++
		}
	}
	return n
}

// changeNoteType applies edit to a copy of the note type and to copies of
// all its notes, then regenerates every note's cards; cards whose question
// becomes empty are kept as empty cards. remapOrd maps old card ords to new
// ones (-1 deletes); nil keeps ords. A change that would leave a note
// without cards is rejected.
func (c *Collection) changeNoteType(kind string, typeID ID, edit func(nt *NoteType) error,
	editFields func(fields []string) []string, remapOrd func(int) int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.noteTypes[typeID]
	if !ok {
		return userErr("noteTypeGone", "That note type no longer exists.")
	}
	nt := cloneNoteType(old)
	if err := edit(nt); err != nil {
		return err
	}
	nt.Modified = c.now()
	o := &op{Kind: kind, NoteTypes: []*NoteType{nt}}
	var noteIDs []ID
	for id, n := range c.notes {
		if n.TypeID == typeID {
			noteIDs = append(noteIDs, id)
		}
	}
	sortIDs(noteIDs)
	empty := 0
	for i, id := range noteIDs {
		if i%500 == 0 {
			c.yield()
		}
		n := c.notes[id]
		note := n
		if editFields != nil {
			nn := cloneNote(n)
			nn.Fields = editFields(nn.Fields)
			nn.Modified = c.now()
			o.Notes = append(o.Notes, &nn)
			note = &nn
		}
		if c.syncCards(o, nt, note, cardOrds(nt, note.Fields), c.deckOfNote(id), remapOrd) == 0 {
			empty++
		}
	}
	if empty > 0 {
		return userErr("changeLeavesNotesEmpty", "This change would leave %d note(s) without cards, so it was not made.", empty)
	}
	return c.commit(o)
}

// RenameNoteType renames a note type. Names are unique, ignoring case.
func (c *Collection) RenameNoteType(id ID, name string) error {
	return c.changeNoteType("renameNoteType", id, func(nt *NoteType) error {
		name, err := c.noteTypeName(name, id)
		if err != nil {
			return err
		}
		nt.Name = name
		return nil
	}, nil, nil)
}

// FieldOption returns the options of field index; fields without an entry
// have the defaults.
func (nt *NoteType) FieldOption(index int) FieldOptions {
	if index >= 0 && index < len(nt.FieldOptions) {
		return nt.FieldOptions[index]
	}
	return FieldOptions{}
}

// SortFieldIndex returns the index of the note type's sort field: the
// stored one while it names a field, else the first field.
func (nt *NoteType) SortFieldIndex() int {
	if nt.SortField >= 0 && nt.SortField < len(nt.Fields) {
		return nt.SortField
	}
	return 0
}

// editFieldOptions applies edit to a FieldOptions slice padded to the
// fields, then trims trailing defaults, so the slice stays aligned with the
// fields and is empty while every field has the defaults.
func (nt *NoteType) editFieldOptions(edit func(opts []FieldOptions) []FieldOptions) {
	opts := make([]FieldOptions, len(nt.Fields))
	copy(opts, nt.FieldOptions)
	opts = edit(opts)
	for len(opts) > 0 && opts[len(opts)-1] == (FieldOptions{}) {
		opts = opts[:len(opts)-1]
	}
	if len(opts) == 0 {
		opts = nil
	}
	nt.FieldOptions = opts
}

func fieldIndex(nt *NoteType, index int) error {
	if index < 0 || index >= len(nt.Fields) {
		return userErr("fieldGone", "That field no longer exists.")
	}
	return nil
}

func uniqueField(nt *NoteType, name string, except int) error {
	for i, f := range nt.Fields {
		if i != except && strings.EqualFold(f, name) {
			return userErr("fieldNameTaken", "The note type already has a field named %q.", f)
		}
	}
	return nil
}

// reservedFieldNames are taken by special fields of the same name.
var reservedFieldNames = []string{"Tags", "Type", "Deck", "Card", "FrontSide"}

// reservedField refuses a field name that a special field already uses,
// ignoring case.
func reservedField(name string) error {
	for _, r := range reservedFieldNames {
		if strings.EqualFold(r, name) {
			return userErr("reservedFieldName", "%q is a reserved name: templates use {{%s}} for a special field. Please choose another name.", name, r)
		}
	}
	return nil
}

// RenameField renames a field and updates the templates that use it.
func (c *Collection) RenameField(typeID ID, index int, name string) error {
	name, err := validName("field", name)
	if err != nil {
		return err
	}
	return c.changeNoteType("renameField", typeID, func(nt *NoteType) error {
		if err := fieldIndex(nt, index); err != nil {
			return err
		}
		if err := uniqueField(nt, name, index); err != nil {
			return err
		}
		old := nt.Fields[index]
		// Imported note types may use reserved names; keeping one is fine.
		if !strings.EqualFold(old, name) {
			if err := reservedField(name); err != nil {
				return err
			}
		}
		nt.Fields[index] = name
		for i := range nt.Templates {
			nt.Templates[i].QFmt = render.RenameField(nt.Templates[i].QFmt, old, name)
			nt.Templates[i].AFmt = render.RenameField(nt.Templates[i].AFmt, old, name)
		}
		return nil
	}, nil, nil)
}

// AddField adds an empty field at the end of a note type's fields.
func (c *Collection) AddField(typeID ID, name string) error {
	name, err := validName("field", name)
	if err != nil {
		return err
	}
	if err := reservedField(name); err != nil {
		return err
	}
	return c.changeNoteType("addField", typeID, func(nt *NoteType) error {
		if err := uniqueField(nt, name, -1); err != nil {
			return err
		}
		nt.Fields = append(nt.Fields, name)
		return nil
	}, func(fields []string) []string { return append(fields, "") }, nil)
}

// RemoveField removes a field and its content from all notes of the type,
// and drops it from the templates.
func (c *Collection) RemoveField(typeID ID, index int) error {
	return c.changeNoteType("removeField", typeID, func(nt *NoteType) error {
		if err := fieldIndex(nt, index); err != nil {
			return err
		}
		if len(nt.Fields) == 1 {
			return userErr("needsField", "A note type needs at least one field.")
		}
		name := nt.Fields[index]
		nt.editFieldOptions(func(opts []FieldOptions) []FieldOptions {
			return append(opts[:index:index], opts[index+1:]...)
		})
		// The sort field stays on its field; without it, the first field
		// sorts.
		switch sf := nt.SortFieldIndex(); {
		case sf == index:
			nt.SortField = 0
		case sf > index:
			nt.SortField = sf - 1
		}
		nt.Fields = append(nt.Fields[:index:index], nt.Fields[index+1:]...)
		for i := range nt.Templates {
			nt.Templates[i].QFmt = render.RemoveField(nt.Templates[i].QFmt, name)
			nt.Templates[i].AFmt = render.RemoveField(nt.Templates[i].AFmt, name)
		}
		return nil
	}, func(fields []string) []string {
		if index < len(fields) {
			return append(fields[:index:index], fields[index+1:]...)
		}
		return fields
	}, nil)
}

// MoveField moves field from of a note type to position to. Notes keep
// each value in its field, and the field keeps its options; templates refer
// to fields by name, so cards are unchanged.
func (c *Collection) MoveField(typeID ID, from, to int) error {
	if from == to {
		return nil
	}
	return c.changeNoteType("moveField", typeID, func(nt *NoteType) error {
		if from < 0 || from >= len(nt.Fields) || to < 0 || to >= len(nt.Fields) {
			return userErr("fieldGone", "That field no longer exists.")
		}
		nt.editFieldOptions(func(opts []FieldOptions) []FieldOptions { return moveItem(opts, from, to) })
		nt.SortField = movedIndex(nt.SortFieldIndex(), from, to)
		nt.Fields = moveItem(nt.Fields, from, to)
		return nil
	}, func(fields []string) []string {
		// Notes with missing fields (see CheckDatabase) are padded first.
		for len(fields) <= max(from, to) {
			fields = append(fields, "")
		}
		return moveItem(fields, from, to)
	}, nil)
}

// moveItem returns a copy of s with item from moved to index to.
func moveItem[T any](s []T, from, to int) []T {
	out := slices.Clone(s)
	v := out[from]
	return slices.Insert(slices.Delete(out, from, from+1), to, v)
}

// movedIndex is where the item at index i is after moveItem(s, from, to).
func movedIndex(i, from, to int) int {
	switch {
	case i == from:
		return to
	case from < to && i > from && i <= to:
		return i - 1
	case to < from && i >= to && i < from:
		return i + 1
	}
	return i
}

// SetSortField makes field index the note type's sort field.
func (c *Collection) SetSortField(typeID ID, index int) error {
	return c.changeNoteType("sortField", typeID, func(nt *NoteType) error {
		if err := fieldIndex(nt, index); err != nil {
			return err
		}
		nt.SortField = index
		return nil
	}, nil, nil)
}

// SetFieldOptions changes the options of a field of a note type.
func (c *Collection) SetFieldOptions(typeID ID, index int, opts FieldOptions) error {
	opts.Font = strings.TrimSpace(opts.Font)
	opts.Size = max(opts.Size, 0)
	return c.changeNoteType("fieldOptions", typeID, func(nt *NoteType) error {
		if err := fieldIndex(nt, index); err != nil {
			return err
		}
		nt.editFieldOptions(func(all []FieldOptions) []FieldOptions {
			all[index] = opts
			return all
		})
		return nil
	}, nil, nil)
}

func validTemplate(nt *NoteType, t CardTemplate, except int) (CardTemplate, error) {
	t = nfcTemplate(t)
	name, err := validName("card type", t.Name)
	if err != nil {
		return t, err
	}
	t.Name = name
	for i, o := range nt.Templates {
		if i != except && strings.EqualFold(o.Name, name) {
			return t, userErr("cardTypeNameTaken", "The note type already has a card type named %q.", o.Name)
		}
	}
	if strings.TrimSpace(t.QFmt) == "" {
		return t, userErr("questionTemplateEmpty", "The question template cannot be empty.")
	}
	if problems := templateProblems(nt, t, except); len(problems) > 0 {
		msgs := make([]string, len(problems))
		for i, p := range problems {
			msgs[i] = p.Message
		}
		// Each problem is tagged, so the app translates them one by one.
		return t, &Error{Msg: strings.Join(msgs, "\n")}
	}
	return t, nil
}

// templateProblems reports what makes saving t as card type except (-1 for
// a new one) of nt unsafe: the render checks, and in standard note types a
// front identical to another card type's.
func templateProblems(nt *NoteType, t CardTemplate, except int) []render.Problem {
	t = nfcTemplate(t)
	problems := render.CheckTemplate(t.QFmt, t.AFmt, nt.Fields, nt.Kind == Cloze)
	if nt.Kind == Cloze {
		return problems
	}
	for i, o := range nt.Templates {
		if i != except && strings.TrimSpace(o.QFmt) == strings.TrimSpace(t.QFmt) {
			problems = append(problems, render.Problem{Message: fmt.Sprintf(
				"Card type %q has the same front template, so both card types would show the same question.", o.Name)})
			break
		}
	}
	return problems
}

// CheckTemplate reports the problems that would refuse saving t as card
// type index (-1 for a new one) of a note type, for feedback while editing.
func (c *Collection) CheckTemplate(typeID ID, index int, t CardTemplate) ([]render.Problem, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nt, ok := c.noteTypes[typeID]
	if !ok {
		return nil, userErr("noteTypeGone", "That note type no longer exists.")
	}
	return templateProblems(nt, t, index), nil
}

// AddTemplate adds a card template to a standard note type; existing notes
// get the new card where its question is not empty.
func (c *Collection) AddTemplate(typeID ID, t CardTemplate) error {
	return c.changeNoteType("addTemplate", typeID, func(nt *NoteType) error {
		if nt.Kind == Cloze {
			return userErr("clozeOneCardType", "A cloze note type has exactly one card type.")
		}
		t, err := validTemplate(nt, t, -1)
		if err != nil {
			return err
		}
		nt.Templates = append(nt.Templates, t)
		return nil
	}, nil, nil)
}

// UpdateTemplate changes a template's name, formats and options; cards are
// regenerated for the new question format.
func (c *Collection) UpdateTemplate(typeID ID, index int, t CardTemplate) error {
	return c.changeNoteType("updateTemplate", typeID, func(nt *NoteType) error {
		if index < 0 || index >= len(nt.Templates) {
			return userErr("cardTypeGone", "That card type no longer exists.")
		}
		t, err := validTemplate(nt, t, index)
		if err != nil {
			return err
		}
		if t.DeckOverride != 0 {
			if d, ok := c.decks[t.DeckOverride]; !ok {
				return userErr("deckGone", "That deck no longer exists.")
			} else if d.Filter != nil {
				return errFilteredTarget(d)
			}
		}
		nt.Templates[index] = t
		return nil
	}, nil, nil)
}

// MoveTemplate moves card type from of a standard note type to position to.
// Cards keep their card type, deck and scheduling; their ords follow the
// card types.
func (c *Collection) MoveTemplate(typeID ID, from, to int) error {
	if from == to {
		return nil
	}
	return c.changeNoteType("moveTemplate", typeID, func(nt *NoteType) error {
		if nt.Kind == Cloze {
			return userErr("clozeOneCardType", "A cloze note type has exactly one card type.")
		}
		if from < 0 || from >= len(nt.Templates) || to < 0 || to >= len(nt.Templates) {
			return userErr("cardTypeGone", "That card type no longer exists.")
		}
		nt.Templates = moveItem(nt.Templates, from, to)
		return nil
	}, nil, func(ord int) int { return movedIndex(ord, from, to) })
}

// RemoveTemplate removes a card template and deletes its cards. It is
// rejected when notes would be left without cards.
func (c *Collection) RemoveTemplate(typeID ID, index int) error {
	return c.changeNoteType("removeTemplate", typeID, func(nt *NoteType) error {
		if index < 0 || index >= len(nt.Templates) {
			return userErr("cardTypeGone", "That card type no longer exists.")
		}
		if len(nt.Templates) == 1 {
			return userErr("needsCardType", "A note type needs at least one card type.")
		}
		nt.Templates = append(nt.Templates[:index:index], nt.Templates[index+1:]...)
		return nil
	}, nil, func(ord int) int {
		switch {
		case ord == index:
			return -1
		case ord > index:
			return ord - 1
		}
		return ord
	})
}

// SetNoteTypeCSS changes a note type's styling.
func (c *Collection) SetNoteTypeCSS(typeID ID, css string) error {
	return c.changeNoteType("noteTypeCSS", typeID, func(nt *NoteType) error {
		nt.CSS = nfc(css)
		return nil
	}, nil, nil)
}
