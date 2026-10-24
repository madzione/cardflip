# Tasks

## 1. Go

- [x] 1.1 Parse `@font-face` blocks in `golang/render/css.go` (family, first local `.ttf`/`.otf` `url()`, `font-weight`, `font-style`) and add `Fonts []FontFace` to `render.Card`; keep `font-family` declarations inlined as today. Add `css_test.go` cases: TTF accepted, WOFF2 and `https://` ignored, several faces for one family. Verify `go test ./render/...` passes.
- [x] 1.2 Add the font list to `StudyCard` and `CardDetails`, run `fluffy generate`, and add the field to `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean and `fvm flutter test` passes.

## 2. Flutter

- [x] 2.1 Add a font cache that loads a media file once through `FontLoader` under a unique family name (for example `cf-<hash of file name>`), and in `card_view.dart` wait for a card's fonts (timeout 1 s), then map inlined `font-family` names to the loaded names. Add `test/card_view_test.dart` tests: the font loader is called once for two cards using the same file, and a missing file renders with the default font. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a deck with a bundled `_font.ttf` and confirm the font on desktop, web and Android.
