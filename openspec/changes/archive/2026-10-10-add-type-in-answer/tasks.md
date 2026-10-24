# Tasks

## 1. Comparison in Go

- [x] 1.1 Add `golang/render/typeans.go` with the comparison from design decision 3 (NFC, trim, strip HTML, LCS on code points, Anki's `typeGood`/`typeBad`/`typeMissed` markup, `nc:` combining-mark removal, 2,000 code point cap). Add `typeans_test.go` with exact match, typo, extra characters, empty input, `nc:` and the cap. Verify `go test ./render/...` passes.
- [x] 1.2 In the renderer, emit `<cf-typeans>` for the first `type:` on the question side, render the comparison on the answer side when typed text is supplied (expected text otherwise), and resolve `type:cloze:` to the active cloze answers. Allow `cf-typeans` in `sanitize.go`. Add `template_test.go` cases for the spec scenarios, including a second `type:` rendering nothing. Verify `go test ./render/...` passes.

## 2. Study service

- [x] 2.1 Add `TypeAnswer bool` to `StudyCard` and `StudyService.Reveal(cardID int64, typed string) (StudyCard, error)` in `golang/svc_study.go`. Add `svc_test.go` cases for a Basic card with `{{type:Back}}`: `TypeAnswer` is set and `Reveal` returns the comparison. Run `fluffy generate` and verify `fluffy generate --check` is clean.
- [x] 2.2 Add `study.Reveal` to `test/support/fake_go.dart` and verify the existing `fvm flutter test` suite still passes.

## 3. Flutter

- [x] 3.1 Map `cf-typeans` to a `TextField` in `lib/widgets/card_view.dart` (controller and focus node supplied by the caller; no field in browser previews). Add a `test/card_view_test.dart` test that the field appears where the element is. Verify it passes.
- [x] 3.2 In `study_screen.dart`, focus the field when the card has `TypeAnswer`, let the field consume key events, reveal on Enter (not while composing) and on "Show answer" by calling `study.reveal` with the text, then unfocus. Add `test/study_test.dart` tests: typing `3` does not rate, Enter reveals and shows the comparison, rating keys work after reveal. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 4.2 Manually on desktop, web and Android: study a `{{type:Back}}` card, type a wrong answer, confirm the comparison and that the on-screen keyboard and Enter behave.
