# Tasks

## 1. Collection

- [x] 1.1 Add `LeechThreshold` and `LeechAction` to `core.DeckOptions` and the DTO with validation. Add `decks_test.go` cases. Verify `go test ./core/...` passes.
- [x] 1.2 Apply the leech rule in `Collection.Answer` in the same op (tag, optional suspend) and record in the review entry what it changed; make `UndoReview` revert exactly that. Add `reviews_test.go` cases for both scheduling scenarios and the undo scenario, including when the note already had the tag. Verify `go test ./core/...` passes.
- [x] 1.3 Report `Leech` (and `Suspended`) for the last answer in `StudyState`. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 1.4 Strip `leech` on import as New in `importer/apkg`. Add an `apkg_test.go` case. Verify it passes.

## 2. UI

- [x] 2.1 Add the options to `deck_options_dialog.dart` and the leech message to `study_screen.dart`. Add widget tests for both. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
