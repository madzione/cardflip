# Tasks

## 1. Collection

- [x] 1.1 Add per-field options and `SortField` to `core.NoteType` and the DTO, kept aligned on add, remove, rename; add `Collection.MoveField` reordering fields and note values in one op (sort field index follows). Add `notes_test.go` cases for the spec scenarios and a backup round trip. Verify `go test ./core/...` passes.
- [x] 1.2 Exclude marked fields from the free-text search text and add the `sortField` sort order in `core/search.go`. Add `search_test.go` cases. Verify `go test ./core/... ./search/...` passes.
- [x] 1.3 Map Anki's field options and `sortf` in `importer/apkg`. Add `apkg_test.go` coverage. Verify `go test ./importer/...` passes.
- [x] 1.4 Add `NotesService.MoveField` and `SetFieldOptions`, run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add drag-to-reorder and an options sheet per field in `note_type_manager.dart`. Add `test/editor_test.dart` tests for moving a field and setting the sort field. Verify they pass.
- [x] 2.2 Apply font, size, direction and collapse in `note_editor_screen.dart`. Add tests for the RTL scenario and collapse. Verify they pass.
- [x] 2.3 Add "Sort field" to the browser's sort menu in `browse_screen.dart`. Add a `test/browse_test.dart` test. Verify it passes.
- [x] 2.4 Add a `test/card_view_test.dart` test for `dir="rtl"`; if the HTML widget ignores `dir`, map it to `Directionality` in the widget factory. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
