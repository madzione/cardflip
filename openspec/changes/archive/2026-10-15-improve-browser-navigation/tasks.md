# Tasks

## 1. Selection

- [x] 1.1 Add `BrowseService.CardsOfNotes(cardIDs)` and `BrowserSplit` in settings. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 1.2 Add Invert selection and Select notes to the selection menu with shortcuts in `browse_screen.dart`. Add `test/browse_test.dart` tests for both scenarios. Verify they pass.

## 2. Layout and keys

- [x] 2.1 Add the draggable, remembered divider on wide screens. Add a test. Verify it passes.
- [x] 2.2 Add keyboard navigation with `Shortcuts`/`Actions`. Add tests for arrows, Shift-extend, Home/End, Ctrl+F, Enter and Delete. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
