//go:build !js

package main

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cardflip/golang/core"
)

type mcpCardsResult struct {
	Cards int `json:"cards" jsonschema:"the cards the call applied to"`
}

type mcpFlagInput struct {
	CardIDs []int64 `json:"cardIds" jsonschema:"card IDs, at most 500"`
	Flag    int     `json:"flag" jsonschema:"0 clears the flag; 1 red, 2 orange, 3 green, 4 blue, 5 pink, 6 turquoise, 7 purple"`
}

type mcpMoveInput struct {
	CardIDs []int64 `json:"cardIds" jsonschema:"card IDs, at most 500"`
	Deck    string  `json:"deck" jsonschema:"the deck's full name, e.g. Spanish::Verbs"`
}

type mcpResetInput struct {
	CardIDs []int64 `json:"cardIds" jsonschema:"card IDs, at most 500"`
	// Pointers, so that missing values take the app's defaults.
	RestorePosition *bool `json:"restorePosition,omitempty" jsonschema:"put each card back at its original position among new cards (default true)"`
	ResetCounts     *bool `json:"resetCounts,omitempty" jsonschema:"set the review and lapse counts to zero (default false)"`
}

type mcpDueDateInput struct {
	CardIDs []int64 `json:"cardIds" jsonschema:"card IDs, at most 500"`
	Days    string  `json:"days" jsonschema:"as the app's Set due date: N days from today (0 is today), N-M a random day in that range, a trailing ! also sets the interval"`
}

type mcpCreateDeckInput struct {
	Name string `json:"name" jsonschema:"the deck's full name; Spanish::Verbs creates Verbs inside Spanish, and Spanish when missing"`
}

type mcpRenameDeckInput struct {
	Deck    string `json:"deck" jsonschema:"the deck's current full name"`
	NewName string `json:"newName" jsonschema:"the new full name; changing the part before :: moves the deck"`
}

type mcpDeckResult struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func addCardTools(server *mcp.Server, t *mcpTools) {
	changeTool(server, t, "suspend_cards", "Suspend cards, so they are not shown when studying.", false,
		func(in mcpCardIDsInput) (mcpCardsResult, error) {
			return t.cardAction(in.CardIDs, func(ids []int64) error { return t.browse.SetSuspended(ids, true) })
		})
	changeTool(server, t, "unsuspend_cards", "Unsuspend cards.", false,
		func(in mcpCardIDsInput) (mcpCardsResult, error) {
			return t.cardAction(in.CardIDs, func(ids []int64) error { return t.browse.SetSuspended(ids, false) })
		})
	changeTool(server, t, "set_flag", "Set a flag on cards, or clear their flags with flag 0.", false,
		func(in mcpFlagInput) (mcpCardsResult, error) {
			if in.Flag < 0 || in.Flag > 7 {
				return mcpCardsResult{}, fmt.Errorf("The flag must be 0 to 7.")
			}
			return t.cardAction(in.CardIDs, func(ids []int64) error {
				c, err := t.col()
				if err != nil {
					return err
				}
				// The browser's SetFlag clears a flag every card has
				// already; agents only set it.
				var change []int64
				for _, id := range ids {
					if card, err := c.Card(core.ID(id)); err == nil && (in.Flag == 0 || card.Flag != in.Flag) {
						change = append(change, id)
					}
				}
				if len(change) == 0 {
					return nil
				}
				return t.browse.SetFlag(change, in.Flag)
			})
		})
	changeTool(server, t, "move_cards", "Move cards to another deck; their scheduling is unchanged.", false,
		func(in mcpMoveInput) (mcpCardsResult, error) {
			c, err := t.col()
			if err != nil {
				return mcpCardsResult{}, err
			}
			deck, err := deckID(c, in.Deck)
			if err != nil {
				return mcpCardsResult{}, err
			}
			return t.cardAction(in.CardIDs, func(ids []int64) error { return t.browse.Move(ids, int64(deck)) })
		})
	changeTool(server, t, "reset_cards", "Make cards new again, as the app's Reset. Their review history is kept.", true,
		func(in mcpResetInput) (mcpCardsResult, error) {
			opts := ResetOptions{RestorePosition: true}
			if in.RestorePosition != nil {
				opts.RestorePosition = *in.RestorePosition
			}
			if in.ResetCounts != nil {
				opts.ResetCounts = *in.ResetCounts
			}
			return t.cardAction(in.CardIDs, func(ids []int64) error { return t.browse.ResetToNew(ids, opts) })
		})
	changeTool(server, t, "set_due_date", "Make cards review cards due on a given day, as the app's Set due date.", true,
		func(in mcpDueDateInput) (mcpCardsResult, error) {
			return t.cardAction(in.CardIDs, func(ids []int64) error { return t.browse.SetDueDate(ids, in.Days) })
		})
	changeTool(server, t, "create_deck", "Create a deck; Parent::Child creates a subdeck, and its parent when missing.", false,
		func(in mcpCreateDeckInput) (mcpDeckResult, error) {
			id, err := t.decks.Create(in.Name)
			if err != nil {
				return mcpDeckResult{}, err
			}
			return t.deckResult(id)
		})
	changeTool(server, t, "rename_deck", "Rename a deck, or move it by changing the part before ::. Its subdecks move with it.", false,
		func(in mcpRenameDeckInput) (mcpDeckResult, error) {
			c, err := t.col()
			if err != nil {
				return mcpDeckResult{}, err
			}
			id, err := deckID(c, in.Deck)
			if err != nil {
				return mcpDeckResult{}, err
			}
			if err := t.decks.Rename(int64(id), in.NewName); err != nil {
				return mcpDeckResult{}, err
			}
			return t.deckResult(int64(id))
		})
}

// cardAction checks that the cards exist and applies do to them.
func (t *mcpTools) cardAction(ids []int64, do func([]int64) error) (mcpCardsResult, error) {
	if err := checkIDCount(len(ids), "card"); err != nil {
		return mcpCardsResult{}, err
	}
	c, err := t.col()
	if err != nil {
		return mcpCardsResult{}, err
	}
	for _, id := range ids {
		if _, err := c.Card(core.ID(id)); err != nil {
			return mcpCardsResult{}, fmt.Errorf("Card %d: %w", id, err)
		}
	}
	return mcpCardsResult{Cards: len(ids)}, do(ids)
}

func (t *mcpTools) deckResult(id int64) (mcpDeckResult, error) {
	c, err := t.col()
	if err != nil {
		return mcpDeckResult{}, err
	}
	d, err := c.Deck(core.ID(id))
	return mcpDeckResult{ID: id, Name: d.Name}, err
}
