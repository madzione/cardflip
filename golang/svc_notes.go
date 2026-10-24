package main

import (
	"cardflip/golang/core"
	"cardflip/golang/render"
)

// NotesService manages note types, notes and tags.
type NotesService struct{ a *app }

// CardTemplate is a card type of a note type.
type CardTemplate struct {
	Name string `json:"name"`
	QFmt string `json:"qfmt"`
	AFmt string `json:"afmt"`
	// DeckOverride is the deck new cards of this card type are added to
	// instead of the note's deck; 0 for none.
	DeckOverride int64 `json:"deckOverride,omitempty"`
	// BrowserQFmt and BrowserAFmt are the simplified templates for the
	// browser; empty means the normal templates.
	BrowserQFmt string `json:"browserQfmt,omitempty"`
	BrowserAFmt string `json:"browserAfmt,omitempty"`
}

func cardTemplateOf(t core.CardTemplate) CardTemplate {
	return CardTemplate{Name: t.Name, QFmt: t.QFmt, AFmt: t.AFmt, DeckOverride: int64(t.DeckOverride),
		BrowserQFmt: t.BrowserQFmt, BrowserAFmt: t.BrowserAFmt}
}

func (t CardTemplate) core() core.CardTemplate {
	return core.CardTemplate{Name: t.Name, QFmt: t.QFmt, AFmt: t.AFmt, DeckOverride: core.ID(t.DeckOverride),
		BrowserQFmt: t.BrowserQFmt, BrowserAFmt: t.BrowserAFmt}
}

// NoteType describes a note type.
type NoteType struct {
	ID int64 `json:"id"`
	// Name of the note type, e.g. "Basic".
	Name string `json:"name"`
	// Cloze is set for cloze note types (one card per cloze number).
	Cloze bool `json:"cloze"`
	// ImageOcclusion is set for note types made like Anki's Image
	// Occlusion, whose notes are edited with the mask editor.
	ImageOcclusion bool `json:"imageOcclusion,omitempty"`
	// OcclusionField and ImageField are the indexes of the fields with an
	// Image Occlusion note's shapes and image, which the mask editor
	// edits; 0 when ImageOcclusion is not set.
	OcclusionField int      `json:"occlusionField,omitempty"`
	ImageField     int      `json:"imageField,omitempty"`
	Fields         []string `json:"fields"`
	// FieldOptions has one entry per field.
	FieldOptions []core.FieldOptions `json:"fieldOptions"`
	// SortField is the index of the field the browser's Sort field column
	// shows and sorts by.
	SortField int            `json:"sortField"`
	Templates []CardTemplate `json:"templates"`
	CSS       string         `json:"css"`
	NoteCount int            `json:"noteCount"`
}

func noteTypeOf(nt core.NoteType, count int) NoteType {
	out := NoteType{ID: int64(nt.ID), Name: nt.Name, Cloze: nt.Kind == core.Cloze,
		Fields: nt.Fields, SortField: nt.SortFieldIndex(), CSS: nt.CSS, NoteCount: count}
	out.OcclusionField, out.ImageField, out.ImageOcclusion = nt.OcclusionFields()
	for i := range nt.Fields {
		out.FieldOptions = append(out.FieldOptions, nt.FieldOption(i))
	}
	for _, t := range nt.Templates {
		out.Templates = append(out.Templates, cardTemplateOf(t))
	}
	return out
}

// Types returns all note types.
func (s *NotesService) Types() ([]NoteType, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	var out []NoteType
	for _, nt := range c.NoteTypes() {
		out = append(out, noteTypeOf(nt, c.NoteCount(nt.ID)))
	}
	return out, nil
}

// Type returns one note type.
func (s *NotesService) Type(id int64) (NoteType, error) {
	c, err := s.a.collection()
	if err != nil {
		return NoteType{}, err
	}
	nt, err := c.NoteType(core.ID(id))
	if err != nil {
		return NoteType{}, err
	}
	return noteTypeOf(nt, c.NoteCount(nt.ID)), nil
}

// EditorDefaults are the note type and deck the editor starts with: the
// ones used for the last added note.
type EditorDefaults struct {
	TypeID int64 `json:"typeId"`
	DeckID int64 `json:"deckId"`
}

// Defaults returns the editor pickers for note type typeID, or for the
// remembered note type when typeID is 0, falling back to Basic. The deck is
// the one last used with that note type when the setting DeckByNoteType is
// on (the add screen then asks again for each note type the user picks),
// else the last used deck; it falls back to the first deck.
func (s *NotesService) Defaults(typeID int64) (EditorDefaults, error) {
	c, err := s.a.collection()
	if err != nil {
		return EditorDefaults{}, err
	}
	st := c.Settings()
	out := EditorDefaults{TypeID: typeID, DeckID: int64(st.LastDeckID)}
	if typeID == 0 {
		out.TypeID = int64(st.LastNoteTypeID)
	}
	if _, err := c.NoteType(core.ID(out.TypeID)); err != nil {
		out.TypeID = 0
		if id, ok := c.BasicNoteType(); ok {
			out.TypeID = int64(id)
		} else if types := c.NoteTypes(); len(types) > 0 {
			out.TypeID = int64(types[0].ID)
		}
	}
	if deck, ok := st.LastDeckByNoteType[core.ID(out.TypeID)]; ok && st.DeckByNoteType {
		out.DeckID = int64(deck)
	}
	if d, err := c.Deck(core.ID(out.DeckID)); err != nil || d.Filter != nil {
		for _, d := range c.Decks() {
			if d.Filter == nil {
				out.DeckID = int64(d.ID)
				break
			}
		}
	}
	return out, nil
}

// NewNote is a note to add.
type NewNote struct {
	TypeID int64    `json:"typeId"`
	DeckID int64    `json:"deckId"`
	Fields []string `json:"fields"`
	// Tags may contain several space-separated tags per entry.
	Tags []string `json:"tags"`
}

// Add adds a note and its cards, and remembers its note type and deck.
func (s *NotesService) Add(note NewNote) (int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	id, err := c.AddNote(core.NoteInput{TypeID: core.ID(note.TypeID), DeckID: core.ID(note.DeckID), Fields: note.Fields, Tags: note.Tags})
	return int64(id), err
}

// Note is a note with its fields named.
type Note struct {
	ID     int64    `json:"id"`
	TypeID int64    `json:"typeId"`
	Fields []string `json:"fields"`
	Tags   []string `json:"tags"`
	// CardIDs are the note's cards, by card type.
	CardIDs []int64 `json:"cardIds"`
	// CardOrds are their card type indexes, parallel to CardIDs; for cloze
	// notes, their cloze numbers minus 1.
	CardOrds []int `json:"cardOrds"`
}

// Get returns a note.
func (s *NotesService) Get(id int64) (Note, error) {
	c, err := s.a.collection()
	if err != nil {
		return Note{}, err
	}
	n, err := c.Note(core.ID(id))
	if err != nil {
		return Note{}, err
	}
	out := Note{ID: int64(n.ID), TypeID: int64(n.TypeID), Fields: n.Fields, Tags: n.Tags}
	for _, card := range c.CardsOfNote(n.ID) {
		out.CardIDs = append(out.CardIDs, int64(card.ID))
		out.CardOrds = append(out.CardOrds, card.Ord)
	}
	return out, nil
}

// Update saves a note's fields and tags; its cards follow.
func (s *NotesService) Update(id int64, fields []string, tags []string) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.UpdateNote(core.ID(id), fields, tags)
}

// Duplicates returns the notes of note type typeID whose first field equals
// firstField, ignoring HTML, case and surrounding whitespace, for the
// editor's duplicate warning. excludeNoteID is the note being edited (0 when
// adding). An empty first field has no duplicates.
func (s *NotesService) Duplicates(typeID int64, firstField string, excludeNoteID int64) ([]int64, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	ids := c.Duplicates(core.ID(typeID), firstField, core.ID(excludeNoteID))
	out := make([]int64, len(ids))
	for i, id := range ids {
		out[i] = int64(id)
	}
	return out, nil
}

// Delete deletes notes and their cards.
func (s *NotesService) Delete(ids []int64) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return c.DeleteNotes(toIDs(ids))
}

// Tags lists the tags in use.
func (s *NotesService) Tags() ([]string, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	return c.AllTags(), nil
}

// RenameTag renames a tag and its child tags on all notes, merging into tags
// that already exist, and returns how many notes changed. Moving a tag under
// another is a rename to parent::tag.
func (s *NotesService) RenameTag(old, name string) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.RenameTag(old, name)
}

// MoveTags moves tags, with their child tags, under parent ("" for the top
// level), merging into tags that already exist, and returns how many notes
// changed.
func (s *NotesService) MoveTags(tags []string, parent string) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.ReparentTags(tags, parent)
}

// DeleteTags removes tags and their child tags from all notes and returns
// how many notes changed. The notes are kept.
func (s *NotesService) DeleteTags(tags []string) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.DeleteTags(tags)
}

// TagUsage counts the notes that have any of the tags or their child tags,
// for the delete confirmation.
func (s *NotesService) TagUsage(tags []string) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.TagUsage(tags), nil
}

// ClearUnusedTags removes tags no note uses from the tag list and returns how
// many it removed.
func (s *NotesService) ClearUnusedTags() (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.ClearUnusedTags(), nil
}

// StockTypes lists the stock note types the user can add by their English
// names, which identify them; the app shows them translated.
func (s *NotesService) StockTypes() []string {
	return core.StockNoteTypes()
}

// AddType creates a note type named name from the stock note type stock.
func (s *NotesService) AddType(stock string, name string) (int64, error) {
	var id core.ID
	err := s.edit(func(c *core.Collection) (err error) {
		id, err = c.AddNoteType(stock, name)
		return err
	})
	return int64(id), err
}

// ImageOcclusionType returns the note type Image Occlusion notes are added
// with, creating it in collections that have none.
func (s *NotesService) ImageOcclusionType() (int64, error) {
	var id core.ID
	err := s.edit(func(c *core.Collection) (err error) {
		id, err = c.ImageOcclusionNoteType()
		return err
	})
	return int64(id), err
}

// CloneType creates a note type named name as a copy of note type id,
// without its notes.
func (s *NotesService) CloneType(id int64, name string) (int64, error) {
	var clone core.ID
	err := s.edit(func(c *core.Collection) (err error) {
		clone, err = c.CloneNoteType(core.ID(id), name)
		return err
	})
	return int64(clone), err
}

// NoteTypeDeletionInfo counts what deleting a note type removes, for the
// confirmation.
type NoteTypeDeletionInfo struct {
	Notes int `json:"notes"`
	Cards int `json:"cards"`
}

// TypeDeletionInfo returns what DeleteType(id) would remove.
func (s *NotesService) TypeDeletionInfo(id int64) (NoteTypeDeletionInfo, error) {
	c, err := s.a.collection()
	if err != nil {
		return NoteTypeDeletionInfo{}, err
	}
	d, err := c.NoteTypeDeletionInfo(core.ID(id))
	return NoteTypeDeletionInfo(d), err
}

// DeleteType deletes a note type with all its notes and their cards. The
// last note type cannot be deleted.
func (s *NotesService) DeleteType(id int64) error {
	return s.edit(func(c *core.Collection) error { return c.DeleteNoteType(core.ID(id)) })
}

// RenameType renames a note type.
func (s *NotesService) RenameType(id int64, name string) error {
	return s.edit(func(c *core.Collection) error { return c.RenameNoteType(core.ID(id), name) })
}

// RenameField renames a field; templates are updated.
func (s *NotesService) RenameField(typeID int64, index int, name string) error {
	return s.edit(func(c *core.Collection) error { return c.RenameField(core.ID(typeID), index, name) })
}

// AddField adds a field at the end.
func (s *NotesService) AddField(typeID int64, name string) error {
	return s.edit(func(c *core.Collection) error { return c.AddField(core.ID(typeID), name) })
}

// RemoveField removes a field and its content from all notes of the type.
func (s *NotesService) RemoveField(typeID int64, index int) error {
	return s.edit(func(c *core.Collection) error { return c.RemoveField(core.ID(typeID), index) })
}

// MoveField moves field from to position to. Notes keep their values in
// the same named fields.
func (s *NotesService) MoveField(typeID int64, from, to int) error {
	return s.edit(func(c *core.Collection) error { return c.MoveField(core.ID(typeID), from, to) })
}

// SetFieldOptions changes the options of a field: editing font and size,
// right to left, collapsed, excluded from unqualified searches, and the
// HTML editor and pin settings.
func (s *NotesService) SetFieldOptions(typeID int64, index int, options core.FieldOptions) error {
	return s.edit(func(c *core.Collection) error { return c.SetFieldOptions(core.ID(typeID), index, options) })
}

// SetSortField makes a field the note type's sort field.
func (s *NotesService) SetSortField(typeID int64, index int) error {
	return s.edit(func(c *core.Collection) error { return c.SetSortField(core.ID(typeID), index) })
}

// SetFieldSticky sets whether a field keeps its content in the add screen
// after a note is added (the field's pin).
func (s *NotesService) SetFieldSticky(typeID int64, index int, on bool) error {
	return s.edit(func(c *core.Collection) error {
		nt, err := c.NoteType(core.ID(typeID))
		if err != nil {
			return err
		}
		opts := nt.FieldOption(index)
		opts.Sticky = on
		return c.SetFieldOptions(core.ID(typeID), index, opts)
	})
}

// AddTemplate adds a card type.
func (s *NotesService) AddTemplate(typeID int64, template CardTemplate) error {
	return s.edit(func(c *core.Collection) error { return c.AddTemplate(core.ID(typeID), template.core()) })
}

// UpdateTemplate changes a card type.
func (s *NotesService) UpdateTemplate(typeID int64, index int, template CardTemplate) error {
	return s.edit(func(c *core.Collection) error {
		return c.UpdateTemplate(core.ID(typeID), index, template.core())
	})
}

// CheckTemplate returns the problems that would refuse saving template as
// card type index of a note type (-1 for a new one), for showing them while
// the user edits. It returns no messages for a template that can be saved.
func (s *NotesService) CheckTemplate(typeID int64, index int, template CardTemplate) ([]string, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	problems, err := c.CheckTemplate(core.ID(typeID), index, template.core())
	if err != nil {
		return nil, err
	}
	out := make([]string, len(problems))
	for i, p := range problems {
		out[i] = p.Message
	}
	return out, nil
}

// MoveTemplate moves card type from of a standard note type to position
// to; cards keep their card type and scheduling.
func (s *NotesService) MoveTemplate(typeID int64, from, to int) error {
	return s.edit(func(c *core.Collection) error { return c.MoveTemplate(core.ID(typeID), from, to) })
}

// RemoveTemplate removes a card type and its cards.
func (s *NotesService) RemoveTemplate(typeID int64, index int) error {
	return s.edit(func(c *core.Collection) error { return c.RemoveTemplate(core.ID(typeID), index) })
}

// TemplateCardCount counts the cards of a card type, for the confirmation.
func (s *NotesService) TemplateCardCount(typeID int64, index int) (int, error) {
	c, err := s.a.collection()
	if err != nil {
		return 0, err
	}
	return c.TemplateCardCount(core.ID(typeID), index), nil
}

// SetCSS changes a note type's styling.
func (s *NotesService) SetCSS(typeID int64, css string) error {
	return s.edit(func(c *core.Collection) error { return c.SetNoteTypeCSS(core.ID(typeID), css) })
}

func (s *NotesService) edit(fn func(c *core.Collection) error) error {
	c, err := s.a.collection()
	if err != nil {
		return err
	}
	return fn(c)
}

func toIDs(ids []int64) []core.ID {
	out := make([]core.ID, len(ids))
	for i, id := range ids {
		out[i] = core.ID(id)
	}
	return out
}

func fromIDs(ids []core.ID) []int64 {
	out := make([]int64, len(ids))
	for i, id := range ids {
		out[i] = int64(id)
	}
	return out
}

// TemplatePreview is a card type rendered from unsaved template text, for
// the preview in the template and CSS editors.
type TemplatePreview struct {
	Question      string            `json:"question"`
	Answer        string            `json:"answer"`
	Background    string            `json:"background,omitempty"`
	Color         string            `json:"color,omitempty"`
	QuestionMedia []render.MediaRef `json:"questionMedia"`
	AnswerMedia   []render.MediaRef `json:"answerMedia"`
	Fonts         []render.FontFace `json:"fonts"`
	// ClozeNumbers are the cloze cards the note has, for cloze note types;
	// a note type without notes previews cloze card 1.
	ClozeNumbers []int `json:"clozeNumbers"`
}

// PreviewTemplate renders template and css of note type typeID without
// saving them: for note noteID, or for placeholder fields "(Name)" when
// noteID is 0. ord is the card type index, or the cloze number for cloze
// note types (0 for the note's first). night renders the dark theme
// variant. The collection is not changed.
func (s *NotesService) PreviewTemplate(typeID int64, template CardTemplate, css string, noteID int64, ord int, night bool) (TemplatePreview, error) {
	c, err := s.a.collection()
	if err != nil {
		return TemplatePreview{}, err
	}
	env := s.a.renderEnv()
	env.Night = night
	p, err := c.PreviewTemplate(core.ID(typeID), template.core(), css, core.ID(noteID), ord, env)
	if err != nil {
		return TemplatePreview{}, err
	}
	return TemplatePreview{
		Question: p.Question, Answer: p.Answer, Background: p.Background, Color: p.Color,
		QuestionMedia: p.QuestionMedia, AnswerMedia: p.AnswerMedia, Fonts: p.Fonts, ClozeNumbers: p.ClozeNumbers,
	}, nil
}

// NoteChoice is a note to pick from, named by its first field.
type NoteChoice struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
}

// PreviewNotes lists up to 50 notes of note type typeID to preview its
// templates with, most recently changed first.
func (s *NotesService) PreviewNotes(typeID int64) ([]NoteChoice, error) {
	c, err := s.a.collection()
	if err != nil {
		return nil, err
	}
	var out []NoteChoice
	for _, n := range c.RecentNotes(core.ID(typeID), 50) {
		out = append(out, NoteChoice{ID: int64(n.ID), Label: n.Label})
	}
	return out, nil
}
