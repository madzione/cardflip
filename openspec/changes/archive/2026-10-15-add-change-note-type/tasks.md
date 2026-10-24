# Tasks

## 1. Collection

- [x] 1.1 Add `Collection.ChangeNoteTypePreview(noteIDs, target, fieldMap, templateMap)` returning notes, cards deleted, cards created and the source fields that lose content, refusing notes of different types. Add `notes_test.go` cases for the counts in the spec scenarios. Verify `go test ./core/...` passes.
- [x] 1.2 Add `Collection.ChangeNoteType(…)` applying the same rules in one op: rewrite note fields and type, remap or delete cards (keeping card IDs, so review history stays attached), generate new cards. Add tests for each spec scenario and a reopen test. Verify `go test ./core/...` passes.

## 2. Service and UI

- [x] 2.1 Add `BrowseService.ChangeNoteTypePreview` and `ChangeNoteType` taking card IDs; run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 2.2 Add `lib/features/browse/change_note_type.dart` (target picker, field and card type mapping with defaults, confirmation with counts) and the bulk action in `browse_screen.dart`. Add `test/browse_test.dart` tests: default mapping by name, mixed selection message, confirm applies. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
