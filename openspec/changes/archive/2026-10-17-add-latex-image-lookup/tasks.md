# Tasks

## 1. Naming and rendering

- [x] 1.1 Create a fixture deck in Anki with `[latex]`, `[$]` and `[$$]` blocks (including `<br>`, `&amp;` and a tag inside) and export it with media to `golang/importer/apkg/testdata`. Record the generated file names.
- [x] 1.2 Add `golang/render/latex.go` porting Anki's extraction, normalisation and `latex-<sha1>` naming, and add `latex_test.go` asserting the names from 1.1. Verify `go test ./render/...` passes.
- [x] 1.3 Replace blocks with `<img class=latex>` when the media exists and `<pre class=latex>` otherwise; the render context needs a media existence check passed from `core`. Add tests for both cases. Verify `go test ./render/... ./core/...` passes.

## 2. Note type setting and import

- [x] 2.1 Add `LatexSVG bool` (`json:"latexSvg,omitempty"`) to `core.NoteType`, read `latexsvg` in `importer/apkg/reader.go` for legacy and modern formats, and make sure `.svg` media is imported. Extend `apkg_test.go` with the fixture from 1.1. Verify `go test ./...` passes.

## 3. Display

- [x] 3.1 Add `flutter_svg` and display `.svg` media in `lib/widgets/card_view.dart`. Add a `test/card_view_test.dart` test with a small SVG from the fake media service. Verify it passes and `fluffy build web` succeeds.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import the fixture deck and confirm the images on desktop and web.
