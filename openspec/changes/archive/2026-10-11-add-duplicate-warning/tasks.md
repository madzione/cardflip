# Tasks

## 1. Go

- [x] 1.1 Add a first-field index per note type in `golang/core` (normalised: HTML stripped, lower-cased, trimmed), maintained on add, update, delete, import and note type changes. Add `notes_test.go` cases, including after an import and a field rename. Verify `go test ./core/...` passes.
- [x] 1.2 Add `NotesService.Duplicates(typeID, firstField, excludeNoteID)`; add `svc_test.go` cases for the spec scenarios. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. Editor

- [x] 2.1 In `note_editor_screen.dart`, check the first field on change (debounced 300 ms), mark it and show the message with "Show duplicates" (opening Browse with those notes). Add `test/editor_test.dart` tests for the three scenarios and for the edit screen ignoring the edited note. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Confirm the check stays responsive with the 50,000-card performance fixture.
