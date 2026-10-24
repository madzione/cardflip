# Tasks

## 1. Go

- [x] 1.1 Add `Collection.FindDuplicates(fieldName, query)` returning sets (content, note IDs) and reuse `AddTags` for tagging. Add `notes_test.go` cases for the scenarios, empty fields, and a field name present in two note types. Add `BrowseService.FindDuplicates`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add `lib/features/browse/find_duplicates.dart` (field picker from all note types, optional search, result list, select set to show notes, Tag duplicates) and a toolbar action. Add `test/browse_test.dart` tests for both requirements. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
