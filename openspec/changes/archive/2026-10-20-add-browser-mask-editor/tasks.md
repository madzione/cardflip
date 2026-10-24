# Tasks

## 1. Shared occlusion editing code

- [x] 1.1 Create `lib/features/editor/occlusion_fields.dart` with the `OcclusionFields` mixin and `occlusionSection` (controller, image HTML and bytes, init/dispose, load and pick image, `shownAsField`, save check, writing the Occlusion and Image fields back), moving the code out of `_NoteEditorScreenState` with the same widget keys and messages; verify `flutter analyze` is clean
- [x] 1.2 Switch `NoteEditorScreen` to the mixin and remove its private copies; verify `flutter test test/editor_test.dart test/occlusion_editor_test.dart` passes unchanged

## 2. Mask editor in the browser's details editor

- [x] 2.1 Use the mixin in `NoteFieldsEditor` (`lib/features/browse/card_details.dart`): for Image Occlusion note types, build the controller with `keepNumbers` and the note's `cardOrds + 1`, show `occlusionSection` in place of the Occlusion field, hide the Image field, and dispose it with the editor; verify by a widget test that selecting an Image Occlusion card shows `occlusion-canvas` and the shape tools, and no `details-field-0` / `details-field-1`
- [x] 2.2 Run the save check and write `occlusion.field` and the image HTML into the saved fields in `NoteFieldsEditor._save`, showing the problem in the pane's error text; verify by widget tests that drawing a rectangle on a note with shapes 1 and 2 saves a field with c1, c2 and c3 (via `notes.Update`), and that deleting all shapes shows the "draw shapes" message without calling `notes.Update`
- [x] 2.3 Verify Change image in the browser by a widget test: with a `FakePicker` image, Change image updates the canvas, and saving stores the file through `media.Add` and the new `<img>` in the Image field
- [x] 2.4 Attach `firstFieldFocusNode` and the initial `_lastField` to the first shown field in `NoteFieldsEditor`; verify by a widget test that the pane's Edit action focuses the Header field of an Image Occlusion note and still focuses field 0 of a Basic note

## 3. Integration checks

- [x] 3.1 Run `flutter analyze` and `flutter test` and verify both pass
- [x] 3.2 Run the app (Linux desktop), open Browse, select an Image Occlusion card, draw/move/delete a shape and save in the pane, then study the note and verify the cards match the masks; also check the narrow-layout `CardDetailsScreen` shows the canvas within the screen width
