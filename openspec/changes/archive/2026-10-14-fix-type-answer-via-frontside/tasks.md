# Tasks

## 1. Renderer

- [x] 1.1 In `golang/render/template.go`, have `typeAnswer` on the question side record the first active filter's expected value (after `cloze:`) and `nc` flag on the context, and expose them from rendering the question. Verify the existing `TestTypeAnswer` cases still pass with `go test ./render/...` in `golang/`.
- [x] 1.2 In `golang/render/render.go`, build `{{FrontSide}}` with a private marker in place of the input placeholder. After rendering the answer template, remove every marker when the answer context set `typeUsed`. Otherwise replace the first marker with the recorded filter's comparison (`CompareTyped`, or the expected text when `typed` is nil), remove the rest, and move the first `<hr id=answer>` line (the `answerLine` pattern) in front of the comparison. Verify with `go test ./render/...`.
- [x] 1.3 Add `TestTypeAnswer` cases for the new spec scenarios: question-only filter with typed text, with `nc:`, with `cloze:`, with nothing typed, without an `<hr id=answer>`, and with `{{FrontSide}}` used twice. Add a `RenderCard` case checking that `AnswerPart` of a question-only preview contains the expected text and that no answer contains the marker. Verify they pass.
- [x] 1.4 Confirm the "type filter on both sides" scenario is covered by the existing `TestRenderCardTyped` and `TestTypeAnswer` cases (comparison shown once, no `cf-typeans` in the answer). Verify they pass unchanged.

## 2. Integration check

- [x] 2.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 2.2 In the app, study a card of a note type with `{{Front}} {{type:Back}}` / `{{FrontSide}}<hr id=answer>{{Back}}`, type a wrong answer, reveal, and verify the comparison shows below the answer line. Open the card in the browser details and verify the expected text shows there too.
