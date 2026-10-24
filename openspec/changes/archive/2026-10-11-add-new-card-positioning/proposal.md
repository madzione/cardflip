# Proposal

## Why

New cards are introduced in the order they were added, and Cardflip offers no way to change that. Users need to move a chapter to the front before an exam, randomise a frequency-sorted deck they do not want to learn in order, or interleave two imported decks. Anki's Reposition (browser) and Insertion order (deck option) cover these.

## What Changes

- Browser action **Reposition** for selected new cards: start position, step, "Randomize order", and "Shift position of existing cards" (to make room instead of creating ties). Non-new cards in the selection are ignored, with a note saying how many.
- Deck option **Insertion order**: sequential (default) or random. With random, newly added cards of that deck get random positions among the deck's new cards. Changing the option repositions the deck's existing new cards accordingly (random shuffles them; sequential restores the order added).
- The browser shows a new card's position (for example `New #42`) in its due column and details.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for repositioning and insertion order.
- `card-browser`: new requirement for the Reposition action and showing positions.

## Impact

- Go: `Collection.Reposition(ids, start, step, randomize, shift)` updating `Card.NewPos`; `core.DeckOptions` gains `InsertionOrder`; `AddNote` and imports assign positions per that option; `Row` shows the position for new cards.
- Flutter: a reposition dialog in `lib/features/browse/`, `browse_screen.dart`, `deck_options_dialog.dart`.
- Tests: `golang/core/cards_test.go`, `golang/core/queue_test.go`, `golang/svc_test.go`, `test/browse_test.dart`.
- Overlap: `add-display-order-options` makes the use of positions configurable; this change only changes the positions.
