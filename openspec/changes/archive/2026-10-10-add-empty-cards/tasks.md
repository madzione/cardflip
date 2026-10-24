# Tasks

## 1. Card generation

- [x] 1.1 Stop deleting cards in `syncCards` when their question becomes empty or their cloze number disappears; keep creating missing cards. Update the existing "Editing removes a cloze" test and add the template typo scenario. Verify `go test ./core/...` passes.
- [x] 1.2 Add `Collection.EmptyCards()` (grouped by note, marking all-empty notes) and `DeleteEmptyCards(ids)` keeping one card per note, in one op. Add tests for both Empty Cards scenarios. Verify `go test ./core/...` passes.

## 2. Services and UI

- [x] 2.1 Add `Empty` to `StudyCard` and `AppService.EmptyCards`/`DeleteEmptyCards`. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 2.2 Add the Empty Cards dialog to `settings_screen.dart` and the empty-card notice to `study_screen.dart`. Add `test/settings_test.dart` and `test/study_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
