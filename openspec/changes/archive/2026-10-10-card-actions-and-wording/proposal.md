# Proposal

## Why

Users can suspend a card in one tap while studying, but undoing that or deleting a card means finding a hidden gesture: long-press a row in Browse, then open the ⋮ menu. Nothing in the UI hints at that gesture, and on desktop a long mouse press is not something people try. The card details view has no actions at all, and its inline editor sits below the scheduling table where it is easy to miss. So users conclude that cards cannot be unsuspended, edited or deleted. The UI also mixes Anki's "note" with "card": the nav tab says "Add", the editor title says "Add", and the buttons say "Add note", in an app called Cardflip.

## What Changes

- The card details view (the side pane on wide screens and the full screen on narrow ones) gets per-card actions: **Edit**, **Suspend** or **Unsuspend** (whichever fits the card's state), and **Delete**. Delete asks for confirmation and says how many cards will be deleted, including other cards made from the same note.
- **Edit** in the details view jumps to the existing inline editor and focuses its first field. The editor moves up, directly under the rendered question and answer, so it is visible without scrolling past the scheduling table.
- The Browse toolbar gets a visible **Select** button that enters multi-select mode. Long-press still works as a shortcut.
- The UI says "card" for what the user adds, edits and deletes. "Add note", "Add a note", "Edit note", "Save note", "Delete notes" and the "Note saved" message become "Add card", "Add a card", "Edit card", "Save", "Delete cards" and "Card saved". "Note type" stays, since it names the template that decides which cards are made. When an action reaches cards the user did not pick (for example the reverse card of a "Basic (and reversed card)" note), the UI says so.
- The bulk delete confirmation counts cards, not notes, for example "Delete 3 cards?", and mentions any extra cards from the same notes.
- No change to data, scheduling or the storage format. Internally the code still uses notes and cards as now.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: card details gain per-card actions (edit, suspend/unsuspend, delete) and the editor moves up; bulk actions gain a visible way into selection mode and a card-based delete confirmation; "Add notes" becomes adding cards.
- `app-shell`: a new terminology requirement (the UI says "card", and keeps "note type"), which also covers relabelling the study screen's "Edit note" action; the deck-list action and the first-run hint say "add a card".

## Impact

- Flutter UI: `lib/features/browse/browse_screen.dart`, `lib/features/browse/card_details.dart`, `lib/features/editor/note_editor_screen.dart`, `lib/features/decks/deck_list_screen.dart`, `lib/features/study/study_screen.dart`.
- Go API: `BrowseService` gains a call that returns how many notes and cards a deletion would remove, replacing `NoteCount`, and `CardDetails` gains the number of cards the note has. `fluffy` regenerates `lib/golang/api.g.dart`.
- Tests: `test/browse_test.dart`, `test/editor_test.dart`, `test/editor_edit_test.dart`, `test/study_test.dart`, `test/deck_list_test.dart`, `integration_test/flow.dart` and `golang/svc_test.go` need updating for the new labels and the new Go call.
