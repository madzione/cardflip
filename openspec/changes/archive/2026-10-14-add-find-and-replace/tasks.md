# Tasks

## 1. Collection

- [x] 1.1 Add `Collection.FindReplacePreview` and `FindReplace` in `golang/core/notes.go` (literal or RE2, case option, all fields, one field by name, or tags), applied in one op with card regeneration. Add `notes_test.go` cases for every scenario and for a replacement that empties a field. Verify `go test ./core/...` passes.
- [x] 1.2 Add `BrowseService.FindReplacePreview` and `FindReplace` taking card IDs (or all). Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add `lib/features/browse/find_replace.dart` (inputs, field list from selected note types, live count, confirm, result) and the toolbar action. Add `test/browse_test.dart` tests for the preview count and the invalid expression message. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
