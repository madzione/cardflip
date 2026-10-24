# Proposal

## Why

Users regularly need to move notes to a different note type: from Basic to "Basic (and reversed card)" to start practising the other direction, or from an imported type to their own. Cardflip has no way to do this, so the only option is to re-create the notes, which loses all review history and scheduling.

## What Changes

- A **Change note type** bulk action in the browser for selected cards whose notes all share one note type.
- The user picks the target note type and maps fields: each target field takes the content of one source field, or stays empty. By default, fields map by equal name, then by position.
- For standard to standard, the user also maps card types: each target card type takes the cards of one source card type, or none. Mapped cards keep their scheduling and review history; source cards that are not mapped are deleted.
- When the source or target is a cloze note type, cards keep their scheduling when the new note generates a card with the same number (cloze number or card type position), and other cards are deleted or created as card generation dictates.
- Before applying, a confirmation states how many notes change, and how many cards will be deleted or created.
- Content of unmapped source fields is lost; the confirmation lists those fields.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for changing the note type of notes.
- `card-browser`: new requirement for the bulk action.

## Impact

- Go: `Collection.ChangeNoteType(noteIDs, targetType, fieldMap []int, templateMap []int)` in `golang/core/notes.go`, applied as one op that rewrites notes and cards; a preview call that returns the counts; `BrowseService` gains both calls.
- Flutter: a mapping dialog in `lib/features/browse/` (new `change_note_type.dart`), wired into `browse_screen.dart` bulk actions.
- Tests: `golang/core/notes_test.go`, `golang/svc_test.go`, `test/browse_test.dart`.
