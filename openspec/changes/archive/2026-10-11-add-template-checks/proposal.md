# Proposal

## Why

A template mistake in Cardflip only shows up later, as an inline error marker on rendered cards, or worse, as cards that silently disappear because the front became empty (cards with a blank front are deleted on save). Anki checks templates when they are saved and refuses changes that would break cards. Cardflip only checks that the question is not empty text.

## What Changes

- Saving a card type (front or back template) is refused, with an explanation, when:
  - a section is not closed or a closing tag has no opening tag;
  - a template refers to a field that does not exist (special fields like `FrontSide`, `Tags`, `Deck`, `Card`, `CardFlag`, `Subdeck`, `Type` are allowed);
  - the front template does not use any field, so it could never produce a card;
  - in a standard note type, the front template is identical to another card type's front;
  - in a cloze note type, the front template has no `{{cloze:…}}` filter.
- Adding or renaming a field to a reserved name (`Tags`, `Type`, `Deck`, `Card`, `FrontSide`, compared case-insensitively) is refused, because those names are taken by special fields.
- The template editor shows the problem before saving, next to the template, as the user types.
- Imported note types are not checked at import; the checks apply when the user edits them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for template checks and reserved field names.

## Impact

- Go: `golang/render` exposes a template check (`render.CheckTemplate`) built on `parseTemplate`; `validTemplate` and `uniqueField` in `golang/core/notetypes.go` call it; a `NotesService.CheckTemplate` call for live feedback.
- Flutter: `TemplateEditorScreen` in `lib/features/editor/note_type_manager.dart` shows the check result under the editor.
- Tests: `golang/render/template_test.go`, `golang/core/notes_test.go`, `test/editor_test.dart`.
