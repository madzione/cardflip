# Tasks

## 1. Data

- [x] 1.1 Add sticky flags per field to the note type's per-field options (kept aligned on field add, remove and rename) with `NotesService.SetFieldSticky(typeID, index, on)`. Add Go tests, including a backup round trip. Verify `go test ./...` passes.
- [x] 1.2 Add the deck mode and `LastDeckByNoteType` to `core.Settings`; record the deck on every add; make `NotesService.Defaults` honour the mode. Add `svc_test.go` cases for both modes. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. Editor

- [x] 2.1 Add pin toggles to each field in `note_editor_screen.dart` (add screen only) and keep pinned fields after a successful add. Add `test/editor_test.dart` tests for the pinned-field scenario and that pins persist when reopening the add screen. Verify they pass.
- [x] 2.2 Add Ctrl/Cmd+Enter to add or save. Add tests for both screens. Verify they pass.
- [x] 2.3 Add the "Deck for new cards" setting to `settings_screen.dart` and switch the deck on note type change in that mode. Add tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
