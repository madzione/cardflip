# Tasks

## 1. Collection

- [x] 1.1 Add `Description` to `core.Deck` and `Collection.SetDeckDescription`; read `desc` in `importer/apkg` for new decks. Add `decks_test.go` and `apkg_test.go` cases. Verify `go test ./core/... ./importer/...` passes.
- [x] 1.2 Count buried cards (manual and sibling) per deck subtree in the queue view (design decision 1). Add `queue_test.go` cases. Verify `go test ./core/...` passes.
- [x] 1.3 Add `Collection.Unbury(deckID)` with the per-day marker (design decision 2). Add `queue_test.go` cases for both unbury scenarios and that the marker expires the next day. Verify `go test ./core/...` passes.

## 2. Service

- [x] 2.1 Add `DecksService.Overview(deckID)` (counts, buried, description, next due), `SetDescription` and `Unbury`; add `svc_test.go` cases; run `fluffy generate` and update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 3. UI

- [x] 3.1 Add `lib/features/decks/deck_overview_screen.dart` with counts, buried count, description (sanitised HTML, external links), Study now, Options, Unbury and a description editor. Add `test/deck_overview_test.dart` tests for the overview scenarios and Unbury. Verify they pass.
- [x] 3.2 Open the overview when a deck is tapped in `deck_list_screen.dart`, keeping the menu's direct Study action and the reminder's direct study. Update `test/deck_list_test.dart` and verify the suite passes.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Update `integration_test/flow.dart` if it taps a deck to study, and verify the integration test passes on Linux.
