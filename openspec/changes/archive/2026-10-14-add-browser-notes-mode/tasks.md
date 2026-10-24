# Tasks

## 1. Go

- [x] 1.1 Add note grouping to `core` search (first-card representative, sort by created, first field, earliest due) and `BrowseService.SearchNotes` and `NoteRows`; add `BrowserMode` to settings. Add `search_test.go` and `svc_test.go` cases for the spec scenarios. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. Browser

- [x] 2.1 Add the mode toggle, note rows and note sorting to `browse_screen.dart`, mapping selections to all card IDs for bulk actions and opening details for the first card. Add `test/browse_test.dart` tests for the three scenarios and that the mode is remembered. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Check responsiveness with the 50,000-card fixture in Notes mode.
