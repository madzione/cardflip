# Proposal

## Why

In dark mode, parts of the typed-answer comparison are drawn in black on the dark card and can't be read, especially from a distance. Note types imported from Anki often style the comparison marks with a fixed text colour that only works on a light card or on their own highlight. For example, the imported "Hiragana Note" has `.typeGood { color: black; background-color: light green; }`. `light green` is not a valid CSS colour, so the highlight is dropped. The correctly typed characters then show black on the dark surface, while `typeBad` (red) and `typeMissed` (grey) keep their highlights and stay readable. The comparison ignores the app theme, even though the rest of the card follows it.

## What Changes

- On the answer side, each part of the type-answer comparison (the `<code id="typeans">` block, its `typeGood`/`typeBad`/`typeMissed` marks, the arrow, and unmarked expected text) is checked for contrast. The check compares the part's text colour with the background it is actually drawn on: its own valid background if it has one, otherwise the nearest ancestor's, otherwise the card surface.
- When that contrast is too low to read, the text uses the card's text colour instead (the note type's `.card` colour or the theme's). If that colour is also too faint against the background, black or white is used, whichever contrasts more.
- Readable colours from the note type are kept as they are. Black on a valid light-green, red or grey highlight still shows black, and nothing changes in light mode for CSS that was readable there.
- The check works the same in light and dark mode. Light text without a background on a light card is corrected too.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirement "Readable type-answer comparison". The comparison stays readable in both themes, even when the note type's CSS gives its marks a colour that would vanish against the background. "Type-in-the-answer comparison" and "Theme-aware card display" are unchanged.

## Impact

- `lib/widgets/card_view.dart`: before `HtmlWidget` builds the answer, rewrite the inline `color` of comparison elements that would be unreadable. `customStylesBuilder` can't do this, because `flutter_widget_from_html_core` applies an element's inline styles after the custom styles, so the inlined `color: black` would still win.
- `lib/widgets/css_color.dart`: a contrast-ratio helper next to `parseCssColor`.
- Tests in `test/card_view_test.dart`.
- No Go, API, data or migration changes. The renderer still inlines the note type's CSS unchanged, and other card content is not touched.
- Overlap: `broaden-css-support` (phase 11) adds id selectors such as `#typeans`, which can give the comparison block colours too. The contrast check reads inlined styles, so it covers those rules without changes.
