# Proposal

## Why

Cardflip only applies CSS rules for `.card`, an element, classes, or an element with classes. Shared decks routinely use descendant and child selectors (`.front b`, `#answer + div`), ids, pseudo-classes such as `:first-child`, background images, and image sizing overrides (`img { max-width: none }`). Those rules are dropped today, so many decks look different from Anki, sometimes badly (hidden elements show, layouts collapse, images are shrunk).

## What Changes

- CSS selectors may use descendant (` `), child (`>`), adjacent (`+`) and general sibling (`~`) combinators, id selectors, attribute selectors (`[attr]`, `[attr=value]`), the universal selector, and the pseudo-classes `:first-child`, `:last-child`, `:nth-child()`, `:only-child` and `:not()` with a simple selector. Selector lists (`a, b`) apply to each part.
- Specificity and source order decide between conflicting rules as in browsers; `!important` wins over normal declarations.
- `background-image: url(…)` naming a media file shows that image.
- Images honour `width`, `height`, `max-width` and `max-height` from CSS and from `width`/`height` attributes; without them, images are still scaled to fit the card width.
- Rules with unsupported selectors (for example `:hover`, `::before`) are ignored as today.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: the "Images" requirement allows CSS and attributes to size images; new requirement for the supported selector set and background images.

## Impact

- Go: `golang/render/css.go` becomes a small selector engine over a parsed element tree (built with the existing `tdewolff/parse` HTML tokenizer); `inline` computes matches with ancestors and siblings.
- Flutter: `lib/widgets/card_view.dart` image builder honours inline sizes; background images through a custom widget factory for elements with an inlined `background-image`.
- Tests: `golang/render/css_test.go` (new), `test/card_view_test.dart`.
- Overlap: `support-environment-css` also edits `css.go` matching; whichever lands second rebases onto the other.
- Performance: the browser question column does not use CSS, so the cost only applies to cards being displayed.
