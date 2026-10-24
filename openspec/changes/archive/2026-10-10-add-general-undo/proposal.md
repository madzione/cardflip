# Proposal

## Why

In Cardflip only study answers can be undone. Deleting the wrong notes, a bulk move to the wrong deck, a tag removed from 500 notes or an unwanted import cannot be reverted except by restoring a backup, which loses everything since. Anki can undo (and redo) almost every operation, which makes bulk actions safe to use.

## What Changes

- Every change made through the app becomes an undoable step with a name: note add, edit and delete; card suspend, bury, move, reset, flag and due date changes; tag changes; deck create, rename, delete and option changes; note type changes; imports; and study answers.
- **Undo** (Ctrl+Z / Cmd+Z outside text fields, and an Undo item in the app's menus) reverts the most recent step and says what was undone ("Undo Delete 3 notes"). **Redo** (Ctrl+Shift+Z / Cmd+Shift+Z, Ctrl+Y) re-applies it.
- Destructive actions (delete, bulk move, import) show a snackbar with an Undo action.
- The undo history holds the last 30 steps of the current app session; it is cleared when the app closes, when a backup is restored, and when storage is reloaded.
- The study screen's undo of answers keeps working as specified; it is the same history, limited to the session's answers.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for undoing and redoing changes.
- `app-shell`: new requirement for the undo controls.

## Impact

- Go: `core` records, for each committed op, an inverse op built from the before-images of everything it changes (decks, note types, notes, cards, settings, review entries), with a name; `Collection.Undo()` and `Redo()` commit the inverse or the original as new ops (so undo is itself durable and crash-safe); size limits keep the history bounded. `AppService.UndoInfo`, `Undo`, `Redo`.
- Media are not deleted by undo; compaction already keeps referenced media, and media of an undone delete come back because notes still reference them (media are only removed at compaction, which also waits while the history references them).
- Flutter: `lib/app/scaffold.dart` (global shortcuts and menu item), `lib/app/errors.dart` (snackbar with Undo), call sites of destructive actions.
- Tests: `golang/core` undo tests for each op kind, `golang/svc_test.go`, `test/app_shell_test.dart`, `test/browse_test.dart`.
