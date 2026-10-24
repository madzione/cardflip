# Tasks

## 1. Limits

- [x] 1.1 Add `TodayLimits` to `core.Deck` and the two options to `core.DeckOptions` (absent = defaults), with DTO fields. Add `decks_test.go` cases, including expiry at rollover. Verify `go test ./core/...` passes.
- [x] 1.2 Use effective limits in `core/queue.go`, cap new cards by the review room when the option is off, and apply parent limits for the studied deck when "limits start from top" is on. Add `queue_test.go` cases for every scheduling scenario. Verify `go test ./core/...` passes.
- [x] 1.3 Add `DecksService.CustomStudyInfo(deckID)` (available extra new and review cards) and `IncreaseToday(deckID, newBy, reviewBy)`. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add `lib/features/decks/custom_study_dialog.dart` and a Custom study entry in the deck menu in `deck_list_screen.dart`. Add a `test/deck_list_test.dart` test for the ten-more scenario. Verify it passes.
- [x] 2.2 Add the two options to `deck_options_dialog.dart` with widget tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
