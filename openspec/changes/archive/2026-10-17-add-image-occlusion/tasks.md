# Tasks

## 1. Rendering in Go

- [x] 1.1 Create an Image Occlusion fixture deck in Anki (rect, ellipse, polygon, text, a group, both modes) and add it to `golang/importer/apkg/testdata`.
- [x] 1.2 Add `golang/render/occlusion.go` parsing occlusion deletions into shapes and emitting `<cf-occlusion>` for the active card and side, removing the template's container markup and script (design decision 1). Add `occlusion_test.go` with each shape kind, `oi=0` and `oi=1`, groups and unknown properties. Verify `go test ./render/...` passes.

## 2. Note type and import

- [x] 2.1 Add the Image Occlusion definition, include it in new collections, and create it on demand for existing ones (design decision 4). Add `core` tests for card generation from shapes and grouping. Verify `go test ./core/...` passes.
- [x] 2.2 Import the fixture and assert note type, media and card counts in `apkg_test.go`. Verify `go test ./importer/...` passes.

## 3. Display

- [x] 3.1 Add `lib/widgets/occlusion_view.dart` and map `cf-occlusion` to it in `card_view.dart`, using CSS variable colours when present. Add `test/card_view_test.dart` tests for both modes on question and answer (golden or paint-call checks). Verify they pass.

## 4. Mask editor

- [x] 4.1 Add `lib/features/editor/occlusion_editor.dart` with shape tools, selection, move/resize handles, delete, group/ungroup, mode switch and zoom, serialising to the Occlusion field (design decision 3). Add widget tests for drawing a rectangle, grouping two shapes and round-tripping an existing field. Verify they pass.
- [x] 4.2 Integrate the editor into `note_editor_screen.dart` for the Image Occlusion note type (image picking through the platform picker, stored as media). Add `test/editor_test.dart` tests for creating a note with two shapes and editing one away. Verify they pass.

## 5. Integration check

- [x] 5.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Manually create an occlusion note on desktop and on a phone, and study the imported fixture on web.
