# Proposal

## Why

Editing a card template in Cardflip is blind: the template editor shows the raw front, back and CSS, and the only way to see the result is to save, leave, find a card and open it. Anki's card template editor shows a live preview, which is what makes template editing practical.

## What Changes

- The card type editor shows a live preview of the front and back, updated as the user types (debounced), using the unsaved template text and the note type's CSS.
- The preview uses a real note of that type when one exists (the user can pick which), and otherwise placeholder text: each field shows `(FieldName)`.
- For cloze note types the preview shows cloze card 1, and the user can choose another cloze number present in the note.
- A light/dark toggle in the preview switches between the theme variants.
- The CSS editor gets the same preview.
- Template errors show in the preview the same way they show on cards.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for previewing note type templates while editing.

## Impact

- Go: `NotesService.PreviewTemplate(typeID, template, css, noteID, ord)` renders unsaved templates without touching the collection; uses `render.RenderCard` with a placeholder note when `noteID` is 0.
- Flutter: `TemplateEditorScreen` and the CSS editor in `lib/features/editor/note_type_manager.dart` (side-by-side on wide screens, tabs on narrow ones), reusing `lib/widgets/card_view.dart`.
- Tests: `golang/svc_test.go`, `test/editor_test.dart`, `test/support/fake_go.dart`.
