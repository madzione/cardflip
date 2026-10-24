# Tasks

## 1. Go

- [x] 1.1 Add `Collection.ToggleMark(noteIDs)` with the any-lacks rule, and `Marked` on `Row` and `StudyCard`; add `StudyService.ToggleMark(cardID)` and `BrowseService.ToggleMark(cardIDs)`. Add `core` and `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 1.2 Strip `marked` in `importer/apkg` when scheduling is not kept. Add an `apkg_test.go` case. Verify it passes.

## 2. Flutter

- [x] 2.1 Add the `*` shortcut, the menu item and the star to `study_screen.dart`. Add a `test/study_test.dart` test for the toggle scenario. Verify it passes.
- [x] 2.2 Add "Toggle mark" and the purple row tint to `browse_screen.dart`. Add a `test/browse_test.dart` test. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
