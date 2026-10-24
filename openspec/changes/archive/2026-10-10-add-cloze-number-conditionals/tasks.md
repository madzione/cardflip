# Tasks

## 1. Rendering

- [x] 1.1 In `golang/render/template.go`, when rendering a `sectionNode` whose name matches `c<digits>` and is not a field of the note, and the context has a cloze ordinal, evaluate the section against `clozeOrd` instead of a field lookup (no unknown-field marker). Add `template_test.go` cases for the three spec scenarios and for a standard note type, where `{{#c1}}` without a field still shows the unknown-field marker. Verify `go test ./render/...` passes.
- [x] 1.2 Make sure `HasContent` and `usedNonEmpty` are not affected by the condition (a cloze-number section does not count as a used field). Add a case where a question consisting only of `{{#c1}}x{{/c1}}{{cloze:Text}}` still generates cards only for the note's cloze numbers. Verify `go test ./render/... ./core/...` passes.

## 2. Integration check

- [x] 2.1 Run `go test ./...` in `golang/`, and verify it passes.
