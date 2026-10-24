# Proposal

## Why

Resetting a card in Cardflip always puts it at the end of the new queue and keeps its review and lapse counts. Anki's reset asks whether to restore the card's original position among new cards (so a reset card is learned in its original order) and whether to reset the review and lapse counts (so leech detection and statistics start fresh). Anki also offers reset from the study screen, which is where a user notices a card needs relearning from scratch.

## What Changes

- Reset (in the browser and, new, in the study screen's menu) shows two options: **Restore original position where possible** (on by default) and **Reset repetition and lapse counts** (off by default).
- A card remembers its original new-card position when it first leaves the New state, so the position can be restored later.
- Resetting adds a manual entry to the review log (rating "manual"), so the card's history shows the reset. The existing review entries are kept, as today.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for reset options and original position.
- `study-session`: new requirement for resetting the current card.
- `card-browser`: new requirement for the reset options dialog.

## Impact

- Go: `core.Card` gains `OriginalPos int64` (`omitempty`), set when a New card is first answered; `Collection.ResetCards(ids, opts)`; `ReviewEntry` supports a manual kind (rating 0) that undo and statistics skip as answers.
- Flutter: a shared reset dialog in `lib/features/browse/`, used by `browse_screen.dart`, `card_details.dart` and `study_screen.dart`.
- Tests: `golang/core/cards_test.go`, `golang/core/reviews_test.go`, `golang/svc_test.go`, `test/browse_test.dart`, `test/study_test.dart`.
- `.apkg` import: Anki stores the original position for reset; when available (`original_position` in card data), it is imported.
