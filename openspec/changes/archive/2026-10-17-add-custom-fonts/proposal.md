# Proposal

## Why

Language decks often ship a font for their script (`_NotoSansJP.ttf`, `_arial_unicode.ttf`) and load it with `@font-face` in the note type CSS. Cardflip ignores `@font-face` and `url()`, so these cards fall back to the system font, which may lack the glyphs (showing boxes) or look wrong.

## What Changes

- `@font-face` rules in the note type CSS whose `src` refers to a TrueType or OpenType file in the collection's media are loaded, and `font-family` declarations using that family name apply to the card.
- Other font formats (WOFF, WOFF2) and remote URLs are ignored, and the text falls back to the next family in the list or the default font.
- Fonts are loaded once per media file and reused across cards; a card waits for its fonts before it is shown, up to a short timeout.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirement for media fonts.

## Impact

- Go: `golang/render/css.go` parses `@font-face` (family, media file, weight, style) and `render.Card` gains a font list; `StudyCard` and `CardDetails` carry it.
- Flutter: `lib/widgets/card_view.dart` loads fonts with `FontLoader` from media bytes (`MediaService.Get`) under a collection-unique family name, and rewrites inlined `font-family` values to that name; a small font cache next to `lib/widgets/media_cache.dart`.
- Tests: `golang/render/css_test.go`, `golang/svc_test.go`, `test/card_view_test.dart`.
