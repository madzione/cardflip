# Tasks

## 1. Model

- [x] 1.1 Add `FilterConfig` to `core.Deck` and `HomeDeckID` (plus preview state) to `core.Card`; route options lookups through the home deck (design decision 1). Add the audit test from the design. Verify `go test ./core/...` passes.
- [x] 1.2 Add `BuildFiltered`, `EmptyFiltered` and rebuild with exclusions, limits, orders and the second filter; make deck deletion follow the collection requirement. Add `core/filtered_test.go` cases for the build, empty and delete scenarios. Verify `go test ./core/...` passes.

## 2. Studying

- [x] 2.1 Queue filtered decks without daily limits, return cards home on graduation and review, and schedule early reviews from elapsed time. Add tests for the review-ahead scenario. Verify `go test ./core/... ./scheduler/...` passes.
- [x] 2.2 Implement preview mode (design decision 2). Add tests for the preview scenario and the delays. Verify they pass.

## 3. Custom study and search

- [x] 3.1 Add the custom study generators (design decision 4) and "Custom Study Session" replacement. Add tests for both scenarios and each choice. Verify `go test ./core/...` passes.
- [x] 3.2 Add `deck:filtered` to `golang/search`. Add a test. Verify it passes.
- [x] 3.3 Add `DecksService` calls (create, update, build, rebuild, empty, custom study). Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 4. UI

- [x] 4.1 Add the filtered deck options screen and filtered-deck styling with Rebuild and Empty in `deck_list_screen.dart`; hide filtered decks from the add screen's deck picker. Add widget tests. Verify they pass.
- [x] 4.2 Add the custom study choices to the deck's Custom study dialog (creating the dialog if it does not exist yet) and "Create filtered deck" in `browse_screen.dart`. Add widget tests. Verify they pass.

## 5. Integration check

- [x] 5.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Manually build, study, preview and empty a filtered deck, and restore a backup taken while cards were in a filtered deck.
