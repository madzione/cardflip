# Proposal

## Why

When entering vocabulary, it is easy to add a word that is already in the collection, which creates competing cards and wastes reviews. Anki warns as soon as the first field matches an existing note of the same note type. Cardflip adds the duplicate without any hint.

## What Changes

- While typing in the add screen, when the first field (HTML stripped, case and surrounding whitespace ignored) equals the first field of an existing note of the same note type, the field is marked and a message says the note is a duplicate.
- The message offers **Show duplicates**, which opens the browser with those notes.
- The warning does not block adding; the user can still add the duplicate on purpose.
- In the edit screen, the warning ignores the note being edited.
- An empty first field never counts as a duplicate.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirement for the duplicate warning.

## Impact

- Go: `NotesService.Duplicates(typeID, firstField, excludeNoteID) ([]int64, error)` using a first-field index per note type kept in `core` (maintained on note changes), so the check is fast with 50,000 notes.
- Flutter: `note_editor_screen.dart` (debounced check, field marker, message with action), navigation to Browse with a search for the note IDs.
- Tests: `golang/core/notes_test.go`, `golang/svc_test.go`, `test/editor_test.dart`.
- The browser search used by "Show duplicates" needs to select specific notes; until an ID search exists, the action passes the note IDs to the browser directly instead of as search text.
