# Tasks

## 1. Service

- [x] 1.1 Add `NotesService.PreviewTemplate(typeID, template CardTemplate, css string, noteID int64, ord int)` returning question and answer HTML; use placeholder fields `(Name)` when `noteID` is 0, and list a note's cloze numbers for cloze types. Add `svc_test.go` cases for the spec scenarios, asserting the collection is unchanged. Run `fluffy generate` and update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the preview to `TemplateEditorScreen` (side-by-side at ≥ 840 px, tabs below), debounced at 300 ms, with a note picker, a cloze number picker for cloze types, and a light/dark toggle. Add `test/editor_test.dart` tests: typing updates the preview after the debounce, placeholder text without notes. Verify they pass.
- [x] 2.2 Add the same preview to the CSS editor. Add a test that a CSS change updates the preview. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
