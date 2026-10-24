# Tasks

## 1. Tag operations

- [x] 1.1 Add `Collection.RenameTag`, `DeleteTag`, `TagUsage` and `ClearUnusedTags` in `golang/core/notes.go` as single ops (case-insensitive matching, children, merge). Add `notes_test.go` cases for every collection scenario. Verify `go test ./core/...` passes.
- [x] 1.2 Add the `NotesService` calls. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. Sidebar

- [x] 2.1 Add `lib/features/browse/browse_sidebar.dart` (sections, trees, filter, rail on wide screens, drawer on narrow) and integrate it in `browse_screen.dart`. Add `test/browse_test.dart` tests for the tag tree and filter. Verify they pass.
- [x] 2.2 Implement click, Ctrl/Cmd-click and Alt/Option-click search composition (quoting names with spaces) and the long-press menu. Add tests for the search scenarios. Verify they pass.
- [x] 2.3 Add context actions and drag-and-drop for tags (rename, delete, move, multi-select) and decks (rename, delete, move), plus "Clear unused tags". Add tests for the drag-to-reparent and delete-tag scenarios. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
