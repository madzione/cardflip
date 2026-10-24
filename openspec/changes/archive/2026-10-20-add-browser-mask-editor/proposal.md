# Proposal

## Why

The browser's card details pane edits an Image Occlusion note as five plain fields: the Occlusion field shows raw `{{c1::image-occlusion:rect:…}}` text and the Image field an `<img>` tag, with no image and no shape tools. Changing masks means leaving the browser and finding the note again from study, and hand-editing the raw field risks breaking the shapes or reusing card numbers. The mask editor already exists in the Add/Edit screen; the browser's editor is the one place it is missing.

## What Changes

- The browser's details editor shows the mask editor (image, shape tools, grouping, mode, zoom, Change image) in place of the Occlusion and Image fields for Image Occlusion notes, as the note editor screen does. The other fields (Header, Back Extra, Comments) and the tags stay ordinary inputs.
- Saving from the browser keeps the shapes' card numbers, gives new shapes numbers no card of the note has, and refuses to save a note without an image or without shapes, with the same messages as the note editor.
- The pane's Edit action focuses the first ordinary field of an Image Occlusion note instead of the hidden Occlusion field.
- The image-and-mask section, image loading and picking, and the save check move out of the note editor screen into shared editor code, so both editors behave the same. No behavior change in the note editor screen.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: "Mask editor" now covers the browser's details editor as well as the Add/Edit screen; "Card details and editing" says the editor of an Image Occlusion note is the mask editor plus its other fields.

## Impact

- `lib/features/browse/card_details.dart` (`NoteFieldsEditor`): occlusion-aware fields, save and focus.
- `lib/features/editor/note_editor_screen.dart`: occlusion section, image load/pick and save check extracted to a shared widget/helper (new file under `lib/features/editor/`).
- Tests: `test/browse_test.dart` (or `test/editor_edit_test.dart`) gains browser mask editor tests; existing `test/editor_test.dart` Image Occlusion tests must keep passing.
- No Go, API, storage or l10n changes expected; existing strings (`chooseImage`, `changeImage`, `drawShapes`, `chooseImageToDrawOn`, …) are reused.
