# Proposal

## Why

Users need to move cards in time by hand: postpone a backlog before an exam, spread a freshly imported batch over the next weeks, or bring a card forward. Anki's Set Due Date does this from the study screen and the browser. Cardflip has no way to change a card's due date except by answering it.

## What Changes

- **Set due date** in the browser (bulk) and in the study screen's menu (current card), shortcut Ctrl+Shift+D (Cmd+Shift+D on macOS).
- Input syntax as in Anki: `0` = today, `1` = tomorrow, `5` = in 5 days, `60-90` = a random day in that range per card. A trailing `!` (`30!`) also sets the card's interval to the new delay.
- New cards become Review cards due on that day. Without `!`, a review card keeps its interval; with `!`, its interval becomes the number of days until the new due date.
- The memory state (stability, difficulty) is kept; new cards get an initial memory state as if first rated Good.
- Each change adds a manual entry to the card's review history.
- The dialog remembers the last input.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `scheduling`: new requirement for setting due dates.
- `study-session`: new requirement for setting the current card's due date.
- `card-browser`: new requirement for the bulk action.

## Impact

- Go: `Collection.SetDueDate(ids, spec)` with a parser for `N`, `N-M`, `!`; per-card randomness seeded by card ID for reproducible tests; manual review entries (shared with reset if both exist); `StudyService.SetDueDate`, `BrowseService.SetDueDate`.
- Flutter: a dialog in `lib/features/browse/` used by `browse_screen.dart`, `card_details.dart` and `study_screen.dart`.
- Tests: `golang/core/cards_test.go`, `golang/svc_test.go`, `test/browse_test.dart`, `test/study_test.dart`.
