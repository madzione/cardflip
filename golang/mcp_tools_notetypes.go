//go:build !js

package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
)

type mcpCreateNoteTypeInput struct {
	Name   string `json:"name" jsonschema:"the new note type's name"`
	Stock  string `json:"stock,omitempty" jsonschema:"a stock note type to start from: Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer) or Cloze"`
	CopyOf string `json:"copyOf,omitempty" jsonschema:"instead of stock, the name of a note type to copy, without its notes"`
}

type mcpNoteTypeInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
}

type mcpRenameNoteTypeInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	NewName  string `json:"newName"`
}

type mcpFieldInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	Field    string `json:"field" jsonschema:"the field's name"`
}

type mcpRenameFieldInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	Field    string `json:"field" jsonschema:"the field's name"`
	NewName  string `json:"newName" jsonschema:"the new name; templates are updated"`
}

type mcpMoveInputByName struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	Name     string `json:"name" jsonschema:"the field's or card type's name"`
	Position int    `json:"position" jsonschema:"the new position, from 0"`
}

type mcpAddTemplateInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	Name     string `json:"name" jsonschema:"the card type's name"`
	Front    string `json:"front" jsonschema:"the front (question) template, e.g. {{Front}}"`
	Back     string `json:"back" jsonschema:"the back (answer) template, e.g. {{FrontSide}}<hr id=answer>{{Back}}"`
}

type mcpUpdateTemplateInput struct {
	NoteType string  `json:"noteType" jsonschema:"the note type's name"`
	Template string  `json:"template" jsonschema:"the card type's name"`
	NewName  *string `json:"newName,omitempty" jsonschema:"a new name; unchanged when not given"`
	Front    *string `json:"front,omitempty" jsonschema:"a new front template; unchanged when not given"`
	Back     *string `json:"back,omitempty" jsonschema:"a new back template; unchanged when not given"`
}

type mcpTemplateInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	Template string `json:"template" jsonschema:"the card type's name"`
}

type mcpCSSInput struct {
	NoteType string `json:"noteType" jsonschema:"the note type's name"`
	CSS      string `json:"css" jsonschema:"the styling of its cards, replacing the current one"`
}

type mcpNoteTypeResult struct {
	NoteType mcpNoteType `json:"noteType"`
}

type mcpRemoveFieldResult struct {
	NoteType      mcpNoteType `json:"noteType"`
	NotesAffected int         `json:"notesAffected" jsonschema:"notes that lost the field's content"`
}

type mcpRemoveTemplateResult struct {
	NoteType     mcpNoteType `json:"noteType"`
	CardsDeleted int         `json:"cardsDeleted"`
}

type mcpDeleteNoteTypeResult struct {
	NotesDeleted int `json:"notesDeleted"`
	CardsDeleted int `json:"cardsDeleted"`
}

func addNoteTypeTools(server *mcp.Server, t *mcpTools) {
	changeTool(server, t, "create_note_type", "Create a note type from a stock note type, or as a copy of another note type without its notes.", false,
		func(in mcpCreateNoteTypeInput) (mcpNoteTypeResult, error) {
			var id int64
			var err error
			switch {
			case (in.Stock == "") == (in.CopyOf == ""):
				return mcpNoteTypeResult{}, fmt.Errorf("Give either stock or copyOf.")
			case in.Stock != "":
				stock := slices.IndexFunc(core.StockNoteTypes(), func(s string) bool { return strings.EqualFold(s, in.Stock) })
				if stock < 0 {
					return mcpNoteTypeResult{}, fmt.Errorf("There is no stock note type %q. The stock note types are: %s.",
						in.Stock, strings.Join(core.StockNoteTypes(), ", "))
				}
				id, err = t.notes.AddType(core.StockNoteTypes()[stock], in.Name)
			default:
				from, ferr := t.noteTypeNamed(in.CopyOf)
				if ferr != nil {
					return mcpNoteTypeResult{}, ferr
				}
				id, err = t.notes.CloneType(from.ID, in.Name)
			}
			if err != nil {
				return mcpNoteTypeResult{}, err
			}
			return t.noteTypeResult(id)
		})
	changeTool(server, t, "rename_note_type", "Rename a note type.", false,
		func(in mcpRenameNoteTypeInput) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error { return t.notes.RenameType(nt.ID, in.NewName) })
		})
	changeTool(server, t, "add_field", "Add a field at the end of a note type's fields; its notes get it empty.", false,
		func(in mcpFieldInput) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error { return t.notes.AddField(nt.ID, in.Field) })
		})
	changeTool(server, t, "rename_field", "Rename a field of a note type; its templates are updated.", false,
		func(in mcpRenameFieldInput) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error {
				i, err := fieldIndex(nt, in.Field)
				if err != nil {
					return err
				}
				return t.notes.RenameField(nt.ID, i, in.NewName)
			})
		})
	changeTool(server, t, "remove_field", "Remove a field and its content from every note of the note type.", true,
		func(in mcpFieldInput) (mcpRemoveFieldResult, error) {
			var affected int
			res, err := t.editNoteType(in.NoteType, func(nt NoteType) error {
				i, err := fieldIndex(nt, in.Field)
				if err != nil {
					return err
				}
				affected = nt.NoteCount
				return t.notes.RemoveField(nt.ID, i)
			})
			return mcpRemoveFieldResult{NoteType: res.NoteType, NotesAffected: affected}, err
		})
	changeTool(server, t, "move_field", "Move a field to another position; notes keep their content in it.", false,
		func(in mcpMoveInputByName) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error {
				i, err := fieldIndex(nt, in.Name)
				if err != nil {
					return err
				}
				return t.notes.MoveField(nt.ID, i, in.Position)
			})
		})
	changeTool(server, t, "add_template", "Add a card type to a note type; its notes get cards of it where its front has content.", false,
		func(in mcpAddTemplateInput) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error {
				return t.notes.AddTemplate(nt.ID, CardTemplate{Name: in.Name, QFmt: in.Front, AFmt: in.Back})
			})
		})
	changeTool(server, t, "update_template", "Change a card type's name, front or back template.", true,
		func(in mcpUpdateTemplateInput) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error {
				i, err := templateIndex(nt, in.Template)
				if err != nil {
					return err
				}
				tm := nt.Templates[i]
				if in.NewName != nil {
					tm.Name = *in.NewName
				}
				if in.Front != nil {
					tm.QFmt = *in.Front
				}
				if in.Back != nil {
					tm.AFmt = *in.Back
				}
				return t.notes.UpdateTemplate(nt.ID, i, tm)
			})
		})
	changeTool(server, t, "remove_template", "Remove a card type and all its cards.", true,
		func(in mcpTemplateInput) (mcpRemoveTemplateResult, error) {
			var deleted int
			res, err := t.editNoteType(in.NoteType, func(nt NoteType) error {
				i, err := templateIndex(nt, in.Template)
				if err != nil {
					return err
				}
				if deleted, err = t.notes.TemplateCardCount(nt.ID, i); err != nil {
					return err
				}
				return t.notes.RemoveTemplate(nt.ID, i)
			})
			return mcpRemoveTemplateResult{NoteType: res.NoteType, CardsDeleted: deleted}, err
		})
	changeTool(server, t, "move_template", "Move a card type to another position; cards keep their card type and scheduling.", false,
		func(in mcpMoveInputByName) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error {
				i, err := templateIndex(nt, in.Name)
				if err != nil {
					return err
				}
				return t.notes.MoveTemplate(nt.ID, i, in.Position)
			})
		})
	changeTool(server, t, "set_note_type_css", "Replace the styling (CSS) of a note type's cards.", true,
		func(in mcpCSSInput) (mcpNoteTypeResult, error) {
			return t.editNoteType(in.NoteType, func(nt NoteType) error { return t.notes.SetCSS(nt.ID, in.CSS) })
		})
	changeTool(server, t, "delete_note_type", "Delete a note type with all its notes and their cards.", true,
		func(in mcpNoteTypeInput) (mcpDeleteNoteTypeResult, error) {
			nt, err := t.noteTypeNamed(in.NoteType)
			if err != nil {
				return mcpDeleteNoteTypeResult{}, err
			}
			info, err := t.notes.TypeDeletionInfo(nt.ID)
			if err != nil {
				return mcpDeleteNoteTypeResult{}, err
			}
			if err := t.notes.DeleteType(nt.ID); err != nil {
				return mcpDeleteNoteTypeResult{}, err
			}
			return mcpDeleteNoteTypeResult{NotesDeleted: info.Notes, CardsDeleted: info.Cards}, nil
		})
}

// editNoteType applies edit to the note type called name and returns it
// changed.
func (t *mcpTools) editNoteType(name string, edit func(NoteType) error) (mcpNoteTypeResult, error) {
	nt, err := t.noteTypeNamed(name)
	if err != nil {
		return mcpNoteTypeResult{}, err
	}
	if err := edit(nt); err != nil {
		return mcpNoteTypeResult{}, err
	}
	return t.noteTypeResult(nt.ID)
}

func (t *mcpTools) noteTypeResult(id int64) (mcpNoteTypeResult, error) {
	nt, err := t.notes.Type(id)
	if err != nil {
		return mcpNoteTypeResult{}, err
	}
	return mcpNoteTypeResult{NoteType: mcpNoteTypeOf(nt)}, nil
}

func fieldIndex(nt NoteType, name string) (int, error) {
	if i := slices.Index(nt.Fields, name); i >= 0 {
		return i, nil
	}
	return 0, fmt.Errorf("Note type %q has no field %q. Its fields are: %s.", nt.Name, name, strings.Join(nt.Fields, ", "))
}

func templateIndex(nt NoteType, name string) (int, error) {
	var names []string
	for i, tm := range nt.Templates {
		if tm.Name == name {
			return i, nil
		}
		names = append(names, tm.Name)
	}
	return 0, fmt.Errorf("Note type %q has no card type %q. Its card types are: %s.", nt.Name, name, strings.Join(names, ", "))
}
