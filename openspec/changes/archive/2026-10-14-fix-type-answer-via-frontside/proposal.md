# Proposal

## Why

A card whose front template has `{{type:Field}}` shows an answer input, but after revealing, the comparison of the typed text with the expected answer only appears when the back template has its own `{{type:Field}}`. In Anki, the back shows the comparison when `{{type:Field}}` is only on the front and the back includes it through `{{FrontSide}}`. Many shared decks and user-made note types are built this way (`{{Front}} {{type:Back}}` / `{{FrontSide}}<hr id=answer>{{Back}}`). With these, Cardflip asks for an answer, then throws away what was typed and never shows the comparison.

## What Changes

- When the back template has no `{{type:…}}` of its own, the comparison (or the expected text when nothing was typed) appears where the front's `{{type:…}}` sits inside `{{FrontSide}}`, using that filter's field and modifiers (`nc:`, `cloze:`).
- As in Anki, when the comparison comes from `{{FrontSide}}` and the back has an `<hr id=answer>` line, the line moves to just before the comparison. The comparison then counts as part of the answer, for example in the browser's answer column and in text exports.
- Back templates with their own `{{type:…}}` work as before: `{{FrontSide}}` leaves out the front's input, and the comparison is shown once, where the back puts it.
- Previews without typed text (browser details, template editor) follow the same rules and show the expected text.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: the "Type-in-the-answer comparison" requirement covers a `{{type:…}}` that reaches the back only through `{{FrontSide}}`.

## Impact

- `golang/render/render.go` (`answer`, which builds `{{FrontSide}}` without the input placeholder) and `golang/render/template.go` (`typeAnswer`, which needs to remember the front's filter field and modifiers).
- Render tests in `golang/render/template_test.go` and `golang/render/sanitize_test.go`.
- No API change: `StudyCard.TypeAnswer`, `StudyService.Reveal` and the study screen stay as they are. `Reveal` already re-renders the answer with the typed text.
- No data or migration changes.
