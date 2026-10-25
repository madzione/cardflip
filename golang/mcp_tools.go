//go:build !js

package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
	"cardflip/golang/internal/usermsg"
	"cardflip/golang/render"
)

// mcpInstructions tell agents what the server is for.
const mcpInstructions = "Cardflip is a flash card app. These tools work on the collection open in the app: " +
	"its decks, notes, cards and note types. Search queries use the app's search syntax, the same as Anki's. " +
	"Fields are given and returned by name. Every change is one step the user can undo in the app."

// mcpTools are the agent tools: thin adapters over the services the app
// uses, so that validation, messages and undo steps are the app's
// (design decision 5).
type mcpTools struct {
	s        *McpService
	a        *app
	notes    *NotesService
	browse   *BrowseService
	decks    *DecksService
	stats    *StatsService
	progress *ProgressService
	media    *MediaService
}

func newMcpTools(s *McpService) *mcpTools {
	a := s.a
	return &mcpTools{s: s, a: a, notes: &NotesService{a: a}, browse: &BrowseService{a: a},
		decks: &DecksService{a: a}, stats: &StatsService{a: a}, progress: &ProgressService{a: a}, media: &MediaService{a: a}}
}

// addMcpTools adds the agent tools to server.
func addMcpTools(server *mcp.Server, s *McpService) {
	t := newMcpTools(s)
	addReadTools(server, t)
	addNoteTools(server, t)
	addCardTools(server, t)
	addNoteTypeTools(server, t)
}

// errNoCollection is the tool error while the app has no collection open.
const errNoCollection = "No collection is open in Cardflip. Open a profile in the app first."

// col returns the open profile's collection, or an error saying none is
// open: the profile list is shown, or the collection could not be read.
func (t *mcpTools) col() (*core.Collection, error) {
	c, err := t.a.collection()
	if errors.Is(err, errNotOpened) {
		return nil, errors.New(errNoCollection)
	}
	if err != nil {
		return nil, fmt.Errorf("No collection is open in Cardflip: %s", usermsg.Plain(err.Error()))
	}
	return c, nil
}

// toolError is err as agents read it: in English, without tags. A batch add
// names the note that failed.
func toolError(err error) error {
	var ae *core.AddNotesError
	if errors.As(err, &ae) {
		return fmt.Errorf("Note %d: %s", ae.Index+1, usermsg.Plain(ae.Err.Error()))
	}
	return errors.New(usermsg.Plain(err.Error()))
}

var no = false

// readTool adds a tool that only reads the collection.
func readTool[In, Out any](server *mcp.Server, t *mcpTools, name, description string, h func(In) (Out, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &no}},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
			if _, err := t.col(); err != nil {
				var zero Out
				return nil, zero, err
			}
			out, err := h(in)
			if err != nil {
				return nil, out, toolError(err)
			}
			return nil, out, nil
		})
}

// changeTool adds a tool that changes the collection. destructive marks
// tools that remove or overwrite data, so clients can ask the user first.
// A call that made an undo step is listed in the activity.
func changeTool[In, Out any](server *mcp.Server, t *mcpTools, name, description string, destructive bool, h func(In) (Out, error)) {
	d := destructive
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description,
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &d, OpenWorldHint: &no}},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
			c, err := t.col()
			if err != nil {
				var zero Out
				return nil, zero, err
			}
			before, _ := c.LastStep()
			out, err := h(in)
			if err != nil {
				return nil, out, toolError(err)
			}
			if after, step := c.LastStep(); after != before && step != "" {
				t.s.recordActivity(step)
			}
			return nil, out, nil
		})
}

// fieldsByName returns the fields of a note of type typeName with fields
// names: base (nil for empty fields) with the fields of named replaced. An
// unknown name fails, listing the note type's fields.
func fieldsByName(typeName string, names []string, base []string, named map[string]string) ([]string, error) {
	out := make([]string, len(names))
	copy(out, base)
	var unknown []string
	for name, v := range named {
		i := slices.Index(names, name)
		if i < 0 {
			unknown = append(unknown, fmt.Sprintf("%q", name))
			continue
		}
		out[i] = v
	}
	if len(unknown) > 0 {
		slices.Sort(unknown)
		word := "field"
		if len(unknown) > 1 {
			word = "fields"
		}
		return nil, fmt.Errorf("Note type %q has no %s %s. Its fields are: %s.", typeName, word,
			strings.Join(unknown, ", "), strings.Join(names, ", "))
	}
	return out, nil
}

// namedFields returns fields keyed by their names.
func namedFields(names []string, fields []string) map[string]string {
	out := make(map[string]string, len(names))
	for i, name := range names {
		if i < len(fields) {
			out[name] = fields[i]
		} else {
			out[name] = ""
		}
	}
	return out
}

// cardText returns a card's question and answer as plain text, rendered as
// the app shows them, the answer without the question that {{FrontSide}}
// repeats.
func (t *mcpTools) cardText(c *core.Collection, id core.ID) (question, answer string, _ error) {
	r, err := c.RenderCard(id, t.a.renderEnv())
	if err != nil {
		return "", "", err
	}
	return render.PlainText(r.Question, r.QuestionMedia, r.QuestionTTS),
		render.PlainText(render.AnswerPart(r.Answer), r.AnswerMedia, r.AnswerTTS), nil
}
