# Tasks

## 1. Core

- [x] 1.1 Capture before-images in `commit` and build named inverse ops (design decision 1); keep the bounded history (decision 2). Add a generic test that runs every exported mutation on a fixture collection, undoes it, and compares the snapshot with the original. Verify `go test ./core/...` passes.
- [x] 1.2 Add `Collection.Undo()` and `Redo()` committing ops durably, clearing redo on new changes and clearing the history on restore and start fresh. Add tests for the spec scenarios and a crash-replay test after undo. Verify `go test ./core/...` passes.
- [x] 1.3 Make compaction keep media referenced by the history (decision 3). Add a test that undoing a note delete after compaction keeps its image. Verify it passes.
- [x] 1.4 Re-express study undo as a view of the history (decision 4) and keep the study-session tests green. Verify `go test ./...` passes.

## 2. Service and UI

- [x] 2.1 Add `AppService.UndoInfo`, `Undo` and `Redo`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 2.2 Add global shortcuts and menu items in `lib/app/scaffold.dart`, and Undo snackbars after delete, bulk move and import. Add `test/app_shell_test.dart` and `test/browse_test.dart` tests for both scenarios. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Check memory use after 30 large steps with the 50,000-card fixture.
