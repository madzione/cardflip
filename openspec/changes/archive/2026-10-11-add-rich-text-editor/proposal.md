# Proposal

## Why

Note fields in Cardflip are plain text boxes: to make a word bold or a list, the user has to type HTML by hand, and imported fields show their raw markup while editing. Every Anki user expects a formatting toolbar, and this is the most visible gap in the add and edit screens.

## What Changes

- Each field becomes a visual (WYSIWYG) editor with a toolbar: bold, italic, underline, subscript, superscript, text colour, highlight colour, clear formatting, bulleted and numbered lists, alignment, and indent/outdent. Common shortcuts work (Ctrl/Cmd+B, I, U, =, Shift+=, R to remove formatting).
- Each field has an HTML source toggle (`</>`, Ctrl/Cmd+Shift+X) that switches between the visual editor and the raw HTML. A per-field option "Use HTML editor by default" opens that field in HTML mode.
- Fields whose HTML contains markup the visual editor cannot keep without loss open in HTML mode with a short note, so editing never silently drops formatting from imported notes.
- Pasting keeps basic formatting (the toolbar's set); Ctrl/Cmd+Shift+V pastes plain text. A setting "Strip formatting when pasting" inverts this.
- Images in a field are shown inline in the visual editor and can be resized by dragging a handle, which sets the image's width.
- The saved field HTML is clean: only the elements and attributes the toolbar produces, plus anything the user wrote in HTML mode.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for the visual field editor, the HTML toggle, paste behaviour and image resizing.
- `app-shell`: new requirement for the paste setting.

## Impact

- Dependency: a pure-Dart rich text editor (`flutter_quill`), plus HTML ↔ editor document converters for the supported subset.
- Go: `core.NoteType` gains a per-field option for "HTML editor by default" (a `FieldOptions` slice parallel to `Fields`, `omitempty`); `core.Settings` gains the paste setting.
- Flutter: a new `lib/features/editor/rich_field.dart` used by `note_editor_screen.dart`, the browser editor in `card_details.dart` and the study edit screen.
- Tests: converter round-trip tests in `test/rich_field_test.dart`, `test/editor_test.dart`, `test/editor_edit_test.dart`, `test/settings_test.dart`.
