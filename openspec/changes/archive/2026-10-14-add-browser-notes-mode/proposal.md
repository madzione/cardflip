# Proposal

## Why

The browser lists cards, so a reversed note appears twice and a cloze note with five deletions five times. When the task is about content (fixing typos, tagging, deleting notes), users want one row per note. Anki's browser has a cards/notes toggle for this.

## What Changes

- A toggle in the browser switches between **Cards** and **Notes** mode; the choice is remembered.
- In Notes mode, each row is a note: its sort field (the first field), note type, number of cards, deck (or "multiple"), tags, and the earliest due date of its cards.
- The search syntax is the same; a note is listed when any of its cards matches.
- Selecting a note shows its first card's details; bulk actions apply to all cards of the selected notes (suspend, move, tags, reset, delete, and the other bulk actions).
- Sorting in Notes mode: created, sort field, earliest due.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for Notes mode.

## Impact

- Go: `BrowseService.SearchNotes(query, sort, descending)` and `NoteRows(ids)` in `golang/svc_browse.go`, built on `Collection.Search` grouping by note; `core.Settings` gains `BrowserMode`.
- Flutter: `browse_screen.dart` (toggle, note rows, selection mapping to card IDs for bulk actions), `card_details.dart` (opened for the note's first card).
- Tests: `golang/core/search_test.go`, `golang/svc_test.go`, `test/browse_test.dart`.
