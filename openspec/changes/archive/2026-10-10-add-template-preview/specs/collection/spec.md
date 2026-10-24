# Spec Delta

## ADDED Requirements

### Requirement: Template preview
While editing a card type's front or back template or the note type's CSS, the app SHALL show the rendered front and back using the unsaved text, updating within a second of the last change. The preview SHALL use a note of that type chosen by the user, or placeholder field text `(FieldName)` when the note type has no notes. Previewing SHALL NOT change the collection.

#### Scenario: Live update
- **WHEN** the user changes the front template from `{{Front}}` to `<b>{{Front}}</b>` without saving
- **THEN** the preview shows the front field in bold, and cards in the collection are unchanged

#### Scenario: No notes yet
- **WHEN** the user edits a template of a note type with no notes
- **THEN** the preview shows `(Front)` where `{{Front}}` is used

#### Scenario: Cloze number
- **WHEN** the user previews a Cloze note with deletions c1 and c2 and chooses card 2
- **THEN** the preview shows the question of cloze card 2
