# Tasks

## 1. Service

- [x] 1.1 Add `StudyService.BuryNote`, `SuspendNote` and `DeleteNote(cardID)` in `golang/svc_study.go` acting on all cards of the note and returning the next state. Add `svc_test.go` cases for the spec scenarios, including that queue counts drop accordingly. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. Study screen

- [x] 2.1 Add the note actions with confirmation for delete to the study menu and keys. Add `test/study_test.dart` tests for suspend note, delete note with confirmation, and each key. Verify they pass.
- [x] 2.2 Let `NoteEditorScreen` take initial values (type, deck, fields, tags) and add Create copy to the study menu, returning to the same card. Add tests in `test/editor_test.dart` and `test/study_test.dart`. Verify they pass.
- [x] 2.3 Open `deck_options_dialog.dart` from study with `O` and a menu item, and reload the session state after it closes. Add a test for the lowered limit scenario. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
