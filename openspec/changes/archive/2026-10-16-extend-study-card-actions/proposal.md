# Proposal

## Why

While studying, Cardflip can edit, bury or suspend the current card only. Anki's **More** menu also acts on the whole note (bury or suspend all its cards, delete it), creates a copy of the note to make a similar card quickly, and opens the deck's options without leaving the session. Users switching from Anki look for these when cleaning up a deck during review.

## What Changes

- **Bury note** (`=`) and **Suspend note** (`!`): bury or suspend all cards of the current note, and move on.
- **Delete note** (Ctrl/Cmd+Delete): deletes the current note with all its cards after a confirmation that states the card count, and moves on.
- **Create copy**: opens the add screen with the current note's note type, deck, fields and tags filled in; the session continues when the add screen is closed.
- **Options** (`O`): opens the options of the deck being studied; changes apply to the rest of the session.
- The existing card actions get keys too: bury card `-`, suspend card `@`, edit `E`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `study-session`: new requirements for note-level actions, create copy and opening deck options during study.

## Impact

- Go: `Collection.BuryCards` and `SetSuspended` already take card IDs; `StudyService` gains `BuryNote`, `SuspendNote` and `DeleteNote` that resolve the note's cards and return the next state; the session's undo stack is unaffected (these are not answers).
- Flutter: `study_screen.dart` (menu items, keys), `note_editor_screen.dart` accepts initial values for a copy, `deck_options_dialog.dart` opened from study.
- Tests: `golang/svc_test.go`, `test/study_test.dart`, `test/editor_test.dart`.
