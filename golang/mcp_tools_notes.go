//go:build !js

package main

import (
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
)

// mcpMaxMedia is the largest media file store_media takes.
const mcpMaxMedia = 20 << 20

// mcpMediaExts are the image, audio and video files store_media takes.
var mcpMediaExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true, ".bmp": true,
	".avif": true, ".tif": true, ".tiff": true, ".ico": true,
	".mp3": true, ".ogg": true, ".oga": true, ".opus": true, ".wav": true, ".m4a": true, ".aac": true,
	".flac": true, ".spx": true,
	".mp4": true, ".webm": true, ".m4v": true, ".mov": true, ".mkv": true, ".avi": true, ".mpg": true,
	".mpeg": true, ".ogv": true, ".3gp": true,
}

type mcpNewNote struct {
	NoteType string            `json:"noteType" jsonschema:"the note type's name, e.g. Basic or Cloze"`
	Deck     string            `json:"deck" jsonschema:"the deck's full name, e.g. Spanish::Verbs"`
	Fields   map[string]string `json:"fields" jsonschema:"fields by name, as HTML; missing fields are empty"`
	Tags     []string          `json:"tags,omitempty"`
}

type mcpAddNotesInput struct {
	Notes []mcpNewNote `json:"notes" jsonschema:"1 to 500 notes"`
	// SkipDuplicates is a pointer so that a missing value means true.
	SkipDuplicates *bool `json:"skipDuplicates,omitempty" jsonschema:"skip notes whose first field matches a note of the same note type (default true)"`
}

type mcpAddedNote struct {
	Index  int     `json:"index" jsonschema:"the note's position in the call, from 0"`
	NoteID int64   `json:"noteId"`
	Cards  []int64 `json:"cardIds"`
}

type mcpDuplicate struct {
	Index       int   `json:"index" jsonschema:"the note's position in the call, from 0"`
	DuplicateOf int64 `json:"duplicateOf" jsonschema:"the note it duplicates"`
}

type mcpAddNotesResult struct {
	Added      []mcpAddedNote `json:"added"`
	Duplicates []mcpDuplicate `json:"duplicates" jsonschema:"notes skipped as duplicates"`
}

type mcpUpdateNoteInput struct {
	NoteID int64             `json:"noteId"`
	Fields map[string]string `json:"fields,omitempty" jsonschema:"fields to change, by name, as HTML; other fields stay"`
	Tags   *[]string         `json:"tags,omitempty" jsonschema:"the note's new tags, replacing all of them; unchanged when not given"`
}

type mcpNoteResult struct {
	Note mcpNote `json:"note"`
}

type mcpDeleteResult struct {
	DeletedNotes int `json:"deletedNotes"`
	DeletedCards int `json:"deletedCards"`
}

type mcpTagsInput struct {
	NoteIDs []int64  `json:"noteIds" jsonschema:"note IDs, at most 500"`
	Tags    []string `json:"tags" jsonschema:"tags; a tag has no spaces, child tags follow their parent after ::"`
}

type mcpCountResult struct {
	Notes int `json:"notes" jsonschema:"the notes the call applied to"`
}

type mcpRenameTagInput struct {
	Tag     string `json:"tag" jsonschema:"the tag to rename; its child tags are renamed with it"`
	NewName string `json:"newName"`
}

type mcpMediaInput struct {
	Name string `json:"name" jsonschema:"the file name with its extension, e.g. map.png"`
	Data string `json:"data" jsonschema:"the file's content, base64, at most 20 MB"`
}

type mcpMediaResult struct {
	Name string `json:"name" jsonschema:"the name to use in fields, as <img src=\"name\"> or [sound:name]; it differs from the given one when that was taken by another file"`
}

func addNoteTools(server *mcp.Server, t *mcpTools) {
	changeTool(server, t, "add_notes", "Add 1 to 500 notes, with their cards, as one step the user can undo. "+
		"By default a note whose first field matches a note of the same note type is skipped and reported. "+
		"An invalid note fails the whole call and nothing is added.", false,
		func(in mcpAddNotesInput) (mcpAddNotesResult, error) {
			return t.addNotes(in)
		})
	changeTool(server, t, "update_note", "Change a note's fields (by name) and optionally its tags. Its cards keep their scheduling.", true,
		func(in mcpUpdateNoteInput) (mcpNoteResult, error) {
			n, err := t.notes.Get(in.NoteID)
			if err != nil {
				return mcpNoteResult{}, err
			}
			nt, err := t.notes.Type(n.TypeID)
			if err != nil {
				return mcpNoteResult{}, err
			}
			fields, err := fieldsByName(nt.Name, nt.Fields, n.Fields, in.Fields)
			if err != nil {
				return mcpNoteResult{}, err
			}
			tags := n.Tags
			if in.Tags != nil {
				tags = *in.Tags
			}
			if err := t.notes.Update(in.NoteID, fields, tags); err != nil {
				return mcpNoteResult{}, err
			}
			notes, err := t.getNotes([]int64{in.NoteID})
			if err != nil {
				return mcpNoteResult{}, err
			}
			return mcpNoteResult{Note: notes[0]}, nil
		})
	changeTool(server, t, "delete_notes", "Delete notes with all their cards and review history.", true,
		func(in mcpNoteIDsInput) (mcpDeleteResult, error) {
			if err := checkIDCount(len(in.NoteIDs), "note"); err != nil {
				return mcpDeleteResult{}, err
			}
			c, err := t.col()
			if err != nil {
				return mcpDeleteResult{}, err
			}
			var out mcpDeleteResult
			seen := map[int64]bool{}
			for _, id := range in.NoteIDs {
				if _, err := c.Note(core.ID(id)); err == nil && !seen[id] {
					seen[id] = true
					out.DeletedNotes++
					out.DeletedCards += len(c.CardsOfNote(core.ID(id)))
				}
			}
			return out, t.notes.Delete(in.NoteIDs)
		})
	changeTool(server, t, "add_tags", "Add tags to notes.", false,
		func(in mcpTagsInput) (mcpCountResult, error) {
			return t.editTags(in, func(c *core.Collection, ids []core.ID) error { return c.AddTags(ids, in.Tags) })
		})
	changeTool(server, t, "remove_tags", "Remove tags from notes, ignoring case.", true,
		func(in mcpTagsInput) (mcpCountResult, error) {
			return t.editTags(in, func(c *core.Collection, ids []core.ID) error { return c.RemoveTags(ids, in.Tags) })
		})
	changeTool(server, t, "rename_tag", "Rename a tag and its child tags on all notes, merging into a tag of that name when there is one.", true,
		func(in mcpRenameTagInput) (mcpCountResult, error) {
			n, err := t.notes.RenameTag(in.Tag, in.NewName)
			return mcpCountResult{Notes: n}, err
		})
	changeTool(server, t, "store_media", "Store an image, audio or video file (base64, at most 20 MB) in the collection's media "+
		"and return the name to use in fields, as <img src=\"name\"> or [sound:name].", false,
		func(in mcpMediaInput) (mcpMediaResult, error) {
			if !mcpMediaExts[strings.ToLower(path.Ext(in.Name))] {
				return mcpMediaResult{}, fmt.Errorf("Only image, audio and video files can be stored; %q has no such extension.", in.Name)
			}
			if base64.StdEncoding.DecodedLen(len(in.Data)) > mcpMaxMedia+3 {
				return mcpMediaResult{}, fmt.Errorf("The file is larger than 20 MB.")
			}
			data, err := decodeBase64(in.Data)
			if err != nil {
				return mcpMediaResult{}, fmt.Errorf("The data is not valid base64: %v", err)
			}
			if len(data) > mcpMaxMedia {
				return mcpMediaResult{}, fmt.Errorf("The file is larger than 20 MB.")
			}
			name, err := t.media.Add(in.Name, data)
			return mcpMediaResult{Name: name}, err
		})
}

// decodeBase64 decodes standard or URL-safe base64, with or without
// padding.
func decodeBase64(s string) ([]byte, error) {
	s = strings.TrimRight(strings.Join(strings.Fields(s), ""), "=")
	if strings.ContainsAny(s, "-_") {
		return base64.RawURLEncoding.DecodeString(s)
	}
	return base64.RawStdEncoding.DecodeString(s)
}

// addNotes adds a batch of notes as one undo step, without changing the
// Add screen's note type and deck (design decision 6).
func (t *mcpTools) addNotes(in mcpAddNotesInput) (mcpAddNotesResult, error) {
	if len(in.Notes) == 0 || len(in.Notes) > mcpMaxLimit {
		return mcpAddNotesResult{}, fmt.Errorf("Give 1 to %d notes.", mcpMaxLimit)
	}
	c, err := t.col()
	if err != nil {
		return mcpAddNotesResult{}, err
	}
	types := map[string]NoteType{}
	decks := map[string]core.ID{}
	inputs := make([]core.NoteInput, len(in.Notes))
	for i, n := range in.Notes {
		nt, ok := types[n.NoteType]
		if !ok {
			if nt, err = t.noteTypeNamed(n.NoteType); err != nil {
				return mcpAddNotesResult{}, fmt.Errorf("Note %d: %w", i+1, err)
			}
			types[n.NoteType] = nt
		}
		deck, ok := decks[n.Deck]
		if !ok {
			if deck, err = deckID(c, n.Deck); err != nil {
				return mcpAddNotesResult{}, fmt.Errorf("Note %d: %w", i+1, err)
			}
			decks[n.Deck] = deck
		}
		fields, err := fieldsByName(nt.Name, nt.Fields, nil, n.Fields)
		if err != nil {
			return mcpAddNotesResult{}, fmt.Errorf("Note %d: %w", i+1, err)
		}
		inputs[i] = core.NoteInput{TypeID: core.ID(nt.ID), DeckID: deck, Fields: fields, Tags: n.Tags}
	}
	skip := in.SkipDuplicates == nil || *in.SkipDuplicates
	results, err := c.AddNotes(inputs, core.AddNotesOptions{KeepDefaults: true, SkipDuplicates: skip})
	if err != nil {
		return mcpAddNotesResult{}, err
	}
	out := mcpAddNotesResult{Added: []mcpAddedNote{}, Duplicates: []mcpDuplicate{}}
	for i, r := range results {
		if r.NoteID == 0 {
			out.Duplicates = append(out.Duplicates, mcpDuplicate{Index: i, DuplicateOf: int64(r.DuplicateOf)})
			continue
		}
		out.Added = append(out.Added, mcpAddedNote{Index: i, NoteID: int64(r.NoteID), Cards: fromIDs(r.Cards)})
	}
	return out, nil
}

// editTags changes the tags of existing notes and counts them.
func (t *mcpTools) editTags(in mcpTagsInput, edit func(*core.Collection, []core.ID) error) (mcpCountResult, error) {
	if err := checkIDCount(len(in.NoteIDs), "note"); err != nil {
		return mcpCountResult{}, err
	}
	if len(core.NormalizeTags(in.Tags)) == 0 {
		return mcpCountResult{}, fmt.Errorf("Give at least one tag.")
	}
	c, err := t.col()
	if err != nil {
		return mcpCountResult{}, err
	}
	var ids []core.ID
	for _, id := range in.NoteIDs {
		if _, err := c.Note(core.ID(id)); err != nil {
			return mcpCountResult{}, fmt.Errorf("Note %d: %w", id, err)
		}
		ids = append(ids, core.ID(id))
	}
	return mcpCountResult{Notes: len(ids)}, edit(c, ids)
}
