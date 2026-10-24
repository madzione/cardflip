# Proposal

## Why

Image Occlusion (hiding labels on a diagram and recalling them one at a time) is one of the most used study techniques in medicine, biology and geography, and it is built into Anki. Cardflip has no way to create such cards, and Image Occlusion notes imported from Anki do not render: their masks are drawn by JavaScript, which Cardflip does not run.

## What Changes

- A stock note type **Image Occlusion** with fields Occlusion, Image, Header, Back Extra and Comments, matching Anki's, so notes move between the two apps unchanged.
- Rendering draws the masks natively, from the occlusion data stored in the Occlusion field (`{{cN::image-occlusion:rect:left=…:top=…:width=…:height=…:oi=1}}`, and `ellipse`, `polygon` and `text` shapes). It does not depend on the template's script.
- Two modes per note: **Hide all, guess one** (all shapes masked, the one being asked highlighted) and **Hide one, guess one** (only the asked shape masked). On the answer, the asked shape is revealed and outlined.
- Shapes with the same number form a group that is asked together as one card.
- A mask editor in the add and edit screens: pick an image, draw rectangles, ellipses, polygons and text, move, resize and delete shapes, group and ungroup, choose the mode, and fill Header and Back Extra.
- `.apkg` import of Image Occlusion notes keeps the note type and renders the masks.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the Image Occlusion stock note type and its card generation.
- `card-rendering`: new requirement for drawing occlusion masks.
- `card-browser`: new requirement for the mask editor.
- `data-import`: new requirement that imported Image Occlusion notes render.

## Impact

- Go: `golang/render/occlusion.go` parses occlusion shapes from cloze deletions and emits a `<cf-occlusion>` element with the image and shape list per card; card generation already counts cloze numbers. The note type is added as a built-in for new collections and created on demand for existing ones (when the user first adds an Image Occlusion note, or on import).
- Flutter: `lib/widgets/occlusion_view.dart` (image plus `CustomPaint` masks, scaled with the image), `lib/features/editor/occlusion_editor.dart` (the mask editor), integration in `note_editor_screen.dart` and `card_view.dart`.
- Tests: `golang/render/occlusion_test.go`, `golang/importer/apkg/apkg_test.go` (an Image Occlusion fixture exported from Anki), `test/card_view_test.dart`, `test/editor_test.dart`.
