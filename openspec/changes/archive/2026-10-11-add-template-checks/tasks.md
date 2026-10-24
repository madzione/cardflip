# Tasks

## 1. Checks

- [x] 1.1 Add `render.CheckTemplate(qfmt, afmt string, fields []string, cloze bool) []Problem` reporting syntax errors from `parseTemplate`, unknown fields (allowing special fields and the filters' field part), a front without fields, and a missing `cloze:` filter for cloze types. Add `template_test.go` cases for each problem and for valid built-in templates. Verify `go test ./render/...` passes.
- [x] 1.2 Call it from `validTemplate` in `golang/core/notetypes.go` and add the identical-front check against other card types. Make sure built-in and stock note types still pass. Add `notes_test.go` cases for the spec scenarios, asserting a refused save changes nothing. Verify `go test ./core/...` passes.
- [x] 1.3 Refuse reserved field names in `AddField` and `RenameField`. Add tests, including that an imported note type with a field named `Type` still loads and renders. Verify `go test ./core/...` passes.
- [x] 1.4 Report a `{{` without `}}`, a tag without a field name, and crossed sections (once) in `parseTemplate`, keeping broken tags as text so the card renders and field renames keep them. Add golden, `CheckTemplate` and `RenameField` cases in `template_test.go` and the spec scenarios in `notes_test.go`. Verify `go test ./render/... ./core/...` passes.

## 2. Live feedback

- [x] 2.1 Add `NotesService.CheckTemplate(typeID, template)` returning problem messages; run `fluffy generate` and update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 2.2 Show problems under the template fields in `TemplateEditorScreen`, debounced at 300 ms, and disable Save while there are problems. Add `test/editor_test.dart` tests for the live message and the disabled Save. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
