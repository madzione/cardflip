# Design

## Context

- `render.RenderCard` returns question and answer HTML in one call; `StudyService.state` sends both to Dart, which shows the answer after reveal. `type:` is currently a no-op filter that prints the field.
- `lib/widgets/card_view.dart` renders HTML with `flutter_widget_from_html_core`, which accepts a custom `WidgetFactory` to replace elements with widgets.
- Scripts are not executed, so an HTML `<input>` would be inert.

## Goals / Non-Goals

**Goals:**
- Comparison computed in Go, so web and native show the same result and it is unit-testable.
- The input is a native Flutter `TextField`, placed inline where the filter is.

**Non-Goals:**
- Multiple `type:` filters on one card (Anki supports only the first; later ones render as nothing here, as in Anki).
- Persisting the typed text.

## Decisions

### 1. Question side: a placeholder element

The renderer emits `<cf-typeans></cf-typeans>` for the first `type:` filter on the question side; sanitization allows this one custom tag. `card_view.dart` maps it to a `TextField` through the widget factory. `StudyCard.TypeAnswer` tells the study screen there is an input to focus on show.

### 2. Answer side: render on reveal

Add `StudyService.Reveal(cardID, typed string) (StudyCard, error)`, which renders the answer with the typed text. The answer HTML in `StudyState` stays as today for cards without `type:`, so the reveal call is only made when `TypeAnswer` is set. This keeps the comparison in Go without sending every keystroke.

Alternative: diff in Dart. Rejected: it would duplicate HTML stripping, cloze answer extraction and `nc:` normalisation that already live in Go.

### 3. Comparison algorithm

Character-level diff (longest common subsequence on Unicode code points) of the typed text against the expected text, both whitespace-trimmed and NFC-normalised; the expected text has HTML stripped and entities decoded. Output as Anki does: when everything matches, one block marked correct; otherwise the typed text with wrong characters marked, an arrow, and the expected text with missing characters marked. Classes `typeGood`, `typeBad`, `typeMissed` and the wrapper `<code id=typeans>` match Anki, so deck CSS keeps working.

`nc:` removes combining marks after NFD decomposition from both sides before comparing. `type:cloze:` uses the active cloze's answers; several deletions with the same number are joined with `, `.

### 4. Keyboard

While the `TextField` has focus, the study screen's `Shortcuts` must not consume digits and Space. Enter in the field triggers reveal. After reveal, focus leaves the field so 1–4 rate again.

## Risks / Trade-offs

- [IME composition on mobile and web (Japanese input) sends Enter to confirm a composition] → Only reveal on Enter when the field is not composing (`TextEditingValue.composing` is empty).
- [Long expected answers make the diff slow] → LCS on strings of a few hundred characters is trivial; cap at 2,000 code points and fall back to showing both texts without marks.
