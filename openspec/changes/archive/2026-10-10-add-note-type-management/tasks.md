# Tasks

## 1. Collection

- [x] 1.1 Add stock definitions for the five stock types next to `builtinNoteTypes()` in `golang/core/notetypes.go`, with the optional reversed template `{{#Add Reverse}}{{Back}}{{/Add Reverse}}` and the type-in front `{{Front}}\n\n{{type:Back}}`. Add a test that a fresh collection still has exactly the three current built-ins. Verify `go test ./core/...` passes.
- [x] 1.2 Add `AddNoteType(stock, name)` and `CloneNoteType(id, name)` with case-insensitive unique names. Add tests for the spec scenarios. Verify `go test ./core/...` passes.
- [x] 1.3 Add `NoteTypeDeletionInfo(id)` (notes, cards) and `DeleteNoteType(id)` deleting notes and cards in one op and refusing the last note type; clear `Settings.LastNoteTypeID` when it points to the deleted type. Add tests for the spec scenarios, including that the deletion survives reopening. Verify `go test ./core/...` passes.
- [x] 1.4 Add a test that a Basic (optional reversed card) note generates the reverse card only with Add Reverse filled, and that filling it later adds the card. Verify it passes.

## 2. Service and UI

- [x] 2.1 Add `NotesService.StockTypes()`, `AddType`, `CloneType`, `TypeDeletionInfo` and `DeleteType`; add `svc_test.go` cases; run `fluffy generate` and update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 2.2 Add Add (stock picker plus name), Clone (name prefilled `<name> copy`) and Delete (confirmation with counts) to `NoteTypeListScreen` in `note_type_manager.dart`. Add `test/editor_test.dart` tests for each dialog and for the disabled delete on the last note type. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
