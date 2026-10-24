# Proposal

## Why

Sibling burying keeps related cards (for example the two directions of a vocabulary note) off the same day, so one does not give away the other. Cardflip applies it always, to new and review siblings alike. Anki lets users choose separately for new and review siblings, because with large backlogs or deliberately related card types users often want siblings on the same day.

## What Changes

- Deck options **Bury new siblings** and **Bury review siblings**, both on by default (today's behaviour).
- With an option off, the corresponding siblings of a note studied today stay in today's queue.
- The options of the deck a card is in decide for that card.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the burying options.
- `scheduling`: "Order of new cards" (its sibling-burying sentence) now depends on the options.

## Impact

- Go: `core.DeckOptions` gains `BuryNew` and `BuryReview` (stored inverted as `noBuryNew`/`noBuryReview` with `omitempty`, so absent means on); `siblingBuried` in `core/queue.go` checks the card's deck options.
- Flutter: `deck_options_dialog.dart`.
- Tests: `golang/core/queue_test.go`, `golang/svc_test.go`.
- Overlap: `add-display-order-options` also modifies "Order of new cards"; `configure-learning-steps` adds an interday learning sibling option.
