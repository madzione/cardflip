# Proposal

## Why

Shared decks that use `[latex]…[/latex]` ship the generated images in their media (`latex-<hash>.png`), so they display everywhere without a TeX installation. Anki finds those images by hashing the LaTeX source. Cardflip does not do this lookup, so the raw `[latex]` source shows as text even though the image was imported.

## What Changes

- The renderer replaces `[latex]…[/latex]`, `[$]…[/$]` and `[$$]…[/$$]` with an image whose file name is derived from the source the way Anki derives it (`latex-<sha1>.png`, or `.svg` when the note type is set to SVG).
- When that image is not in the collection's media, the LaTeX source is shown in a monospace box instead of a broken image.
- `.apkg` import keeps the note type's SVG setting, and keeps importing the pre-generated images as media (already the case for PNG; SVG images are added).
- SVG images display on every platform.
- Generating new LaTeX images is out of scope: it needs a TeX installation, which Cardflip does not ship.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirement for pre-generated LaTeX images.
- `data-import`: new requirement that the note type's LaTeX image format is kept.

## Impact

- Go: new `golang/render/latex.go`, a `LatexSVG bool` on `core.NoteType` (`omitempty`), `importer/apkg/reader.go` reads `latexsvg` from the note type config.
- Flutter: `flutter_svg` dependency for SVG media in `lib/widgets/card_view.dart`.
- Tests: `golang/render/latex_test.go` (hash fixtures produced by Anki), `golang/importer/apkg/apkg_test.go`, `test/card_view_test.dart`.
