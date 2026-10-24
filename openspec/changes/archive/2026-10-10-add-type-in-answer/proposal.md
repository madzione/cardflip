# Proposal

## Why

`{{type:Field}}` is how Anki users practise active recall by typing, and many language decks depend on it. Cardflip renders `type:` only as the expected answer on the back: there is no input box and no comparison, so these cards turn into plain flip cards.

## What Changes

- On the question side, `{{type:Field}}` shows a text input where the filter appears. Enter (or "Show answer") reveals the answer.
- On the answer side, `{{type:Field}}` shows a comparison of what was typed with the expected text: matching characters, wrong characters and missing characters are marked differently. If nothing was typed, the expected answer is shown as today.
- `{{type:cloze:Text}}` compares against the answer(s) of the active cloze number. `{{type:nc:Field}}` ignores accents (combining marks) when comparing.
- The comparison ignores HTML in the expected field and surrounding whitespace.
- The browser's preview keeps showing the expected answer without an input box.
- While the input has focus, the study keyboard shortcuts (1–4, Space) type text instead of rating; Enter reveals the answer.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: the "Field filters" requirement changes `type:` from plain content to an input and comparison; new requirement for the comparison rules.
- `study-session`: new requirement for typing an answer during study.

## Impact

- Go: `golang/render` (a type-answer placeholder on the question side, comparison rendering on the answer side, `type:cloze:` and `type:nc:`), `golang/svc_study.go` (the answer side is rendered with the typed text), `StudyCard` gains `TypeAnswer bool`.
- Flutter: `lib/widgets/card_view.dart` (inline `TextField` for the placeholder element), `lib/features/study/study_screen.dart` (focus handling, shortcuts disabled while typing, passing the typed text when revealing).
- Tests: `golang/render/template_test.go`, a new `golang/render/typeans_test.go`, `golang/svc_test.go`, `test/study_test.dart`, `test/card_view_test.dart`.
- Overlap: the stock note type "Basic (type in the answer)" is part of the `add-note-type-management` change; it works with or without this change.
