# Proposal

## Why

Japanese decks write readings as `日本語[にほんご]` and show them with the `furigana:`, `kana:` and `kanji:` filters. Cardflip ignores unknown filters, so these cards show the raw brackets instead of ruby text. For Japanese learners this affects most shared decks.

## What Changes

- Add the filters `furigana:`, `kana:` and `kanji:` to the template renderer, using Anki's reading syntax: `base[reading]`, where the base is the run of characters before `[` up to the previous space or tag, and a single space before the base is dropped.
- `furigana:` renders ruby text (`<ruby>base<rt>reading</rt></ruby>`), `kana:` renders only the readings, `kanji:` renders only the base text.
- `[sound:…]` tags are never treated as readings.
- The card display renders ruby text with the reading above the base on every platform.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirement for the furigana filters and ruby display.

## Impact

- Go: `golang/render/template.go` (`context.filter`), a new `golang/render/furigana.go`; `golang/render/sanitize.go` must allow `ruby`, `rt` and `rp`.
- Flutter: `lib/widgets/card_view.dart` must render `<ruby>`; `flutter_widget_from_html_core` supports ruby in its core package, which needs a widget test to confirm on our version.
- Tests: `golang/render/template_test.go`, a new `golang/render/furigana_test.go`, `test/card_view_test.dart`.
