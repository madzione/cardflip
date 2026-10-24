# Tasks

## 1. Collection

- [x] 1.1 Add `OriginalPos` to `core.Card`, set from `NewPos` on the first answer of a New card in `core/reviews.go`, and import it from Anki card data when present. Add tests that it is set once and survives backups. Verify `go test ./...` passes.
- [x] 1.2 Change `Collection.ResetCards(ids, opts)` to restore positions and optionally reset counts, and add a manual review entry per card (rating 0, skipped by undo of answers and by today's answer counts). Add `cards_test.go` and `reviews_test.go` cases for the spec scenarios. Verify `go test ./core/...` passes.

## 2. Services and UI

- [x] 2.1 Extend `BrowseService.ResetToNew` with options and add `StudyService.Reset(cardID, opts)`; run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 2.2 Add the reset dialog and use it in `browse_screen.dart`, `card_details.dart` and the study menu. Add `test/browse_test.dart` and `test/study_test.dart` tests for the defaults and for confirming. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
