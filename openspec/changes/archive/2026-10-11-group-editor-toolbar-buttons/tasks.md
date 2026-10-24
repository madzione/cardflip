# Tasks

## 1. Toolbar

- [x] 1.1 Give `RichToolbar` (`lib/features/editor/rich_field.dart`) an optional list of trailing buttons, shown after a vertical divider at the end of its `Wrap` as one child (a `Row`, so the group never splits across lines) and left as given while the field is in HTML mode (only the formatting buttons are disabled then). Add a `test/rich_field_test.dart` case that a trailing button stays enabled in HTML mode while Bold is disabled, and verify it passes.

## 2. Editors

- [x] 2.1 Grow the `FieldMedia` mixin (`lib/features/editor/field_media.dart`, renamed to `field_tools.dart` / `FieldTools`) into the source of all editing buttons: move the cloze insertion from `note_editor_screen.dart` into it (next unused number across all fields of the note, wrapping the selection of the last focused field), add the Math wrap, and replace `mediaButtons()` with `toolButtons()` returning Cloze deletion (only when the mixin's note type is a cloze type), `MathMenuButton`, Attach and Record audio, in that order. Verify `fvm flutter analyze` passes and the existing cloze, math and media tests still pass.
- [x] 2.2 In `note_editor_screen.dart` and `card_details.dart`, pass `toolButtons()` to the toolbar; remove Cloze, Math, Attach and Record audio from the app bar (Note types stays) and Math, Attach and Record audio from the row with Save (Save stays). In `test/editor_test.dart` and `test/browse_test.dart` cover the spec scenarios: the order inside `rich-toolbar` for Basic and Cloze in the Add card screen and for a cloze note in the browser, the app bar keeping only Note types, the Save row keeping only Save, Cloze in the browser giving `{{c2::France}}`, and the HTML source scenario (inline math and `cat.jpg` inserted while the formatting buttons are disabled). Verify they pass.

## 3. Integration check

- [x] 3.1 Run `fvm flutter analyze` and `fvm flutter test` and verify both pass. On a phone-width window (Linux at 400 px and Android), check that the toolbar wraps to a second line with the editing group kept together, and that the app bar shows only the title and Note types.

## Workflow follow-up

- Implement only after `add-editor-media` is archived.
- Archive with `/opsx:archive group-editor-toolbar-buttons` and tick it in `ROADMAP.md`.
