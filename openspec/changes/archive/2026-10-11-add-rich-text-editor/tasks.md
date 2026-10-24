# Tasks

## 1. Data and settings

- [x] 1.1 Add `FieldOptions []FieldOptions{HTMLByDefault bool}` to `core.NoteType`, kept aligned in `AddField`, `RemoveField` and `RenameField`, and exposed in the `NoteType` DTO with a setter; add `StripPasteFormatting` to `core.Settings` with a setter. Add `core` and `svc_test.go` tests, run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. Converters

- [x] 2.1 Add `flutter_quill` and write HTML → document and document → HTML converters for the subset in design decision 1, with Anki-style output (decision 3). Add `test/rich_field_test.dart` round-trip tests for each element, nesting, and real fields from two shared decks. Verify they pass.
- [x] 2.2 Implement the lossless guard (decision 2) with tests for a table, a `div` with a class, and a field that round-trips cleanly. Verify they pass.

## 3. Editor UI

- [x] 3.1 Add `lib/features/editor/rich_field.dart` with the toolbar, shortcuts, HTML toggle (Ctrl/Cmd+Shift+X), the HTML-by-default option and the guard note, and use it in `note_editor_screen.dart`. Add `test/editor_test.dart` tests for Ctrl+B saving `<b>`, the toggle, and the imported-table scenario. Verify they pass.
- [x] 3.2 Add paste handling (decision 4) and the plain-paste shortcut, honouring the setting; add the setting to `settings_screen.dart`. Add tests with a fake clipboard. Verify they pass.
- [x] 3.3 Show images inline with a resize handle that writes `width`. Add a test that dragging the handle changes the saved width. Verify it passes.
- [x] 3.4 Add the "Use HTML editor by default" switch to the field list in `note_type_manager.dart`. Add a test. Verify it passes.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Compare web bundle size before and after, and test IME input (Japanese) on Android and macOS.
