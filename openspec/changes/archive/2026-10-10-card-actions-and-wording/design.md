# Design

## Context

- `BrowseScreen` treats "selection mode" as `_selected.isNotEmpty`. The only way in is `onLongPress` on a row, and every bulk action (suspend, unsuspend, delete, …) lives in the selection app bar.
- `CardDetailsPane` is shared by the wide layout (right-hand pane) and the narrow `CardDetailsScreen`. It has no actions. Its `NoteFieldsEditor` comes last, after the scheduling table.
- The Go `BrowseService` already has `SetSuspended(ids, bool)` and `DeleteNotes(ids)`, which both accept a single ID. `NoteCount(ids)` exists only for the delete confirmation. `core.Collection.CardsOfNote` already exists.
- Both screens already reload on `changes.stream`, and `_search()` already clears `_details` when the shown card is no longer in the results.

## Goals / Non-Goals

**Goals:**
- Reuse the existing bulk Go calls for the single-card actions. The only Go change is the count the new confirmation needs.
- One implementation of the card actions, used by both the wide pane and the narrow screen.

**Non-Goals:**
- Renaming code identifiers (`NoteEditorScreen`, `golang.notes`, `deleteNotes`, …). Only user-facing strings change.
- Rewording import summaries and backup-damage diagnostics, which talk about notes in a technical context.
- Ctrl/Shift-click range selection, keyboard shortcuts for card actions, or undo for delete.
- Changing the "Add" nav destination label, which `app-shell` requires as is.

## Decisions

**1. A deletion-impact call replaces `NoteCount`.**
`BrowseService.DeleteImpact(cardIDs) → {notes, cards}` returns the number of distinct notes and the total number of cards those notes have, counted with `CardsOfNote`. The confirmation shows `cards`, and when `cards > len(selected)` it adds "including N other cards from the same notes". `NoteCount` is removed because nothing else calls it.
*Alternative:* count siblings in Dart from `Row.noteId`. That was rejected because rows load lazily per page, so siblings outside the loaded pages or outside the search would be missed.

**2. `CardDetails` gains `noteCards int`.**
The editor's "changes also apply to N other cards" hint and the single-card delete confirmation both read it, so the details view needs no extra call. Single-card delete still goes through the same confirmation helper as bulk delete, for one wording.

**3. Single-card actions use the existing bulk calls with `[cardId]`.**
`setSuspended([id], !suspended)` and `deleteNotes([id])`. No new per-card Go methods are added.

**4. Where the actions go.**
- Narrow (`CardDetailsScreen`): in the `AppBar` as icon buttons: Edit (`Icons.edit_outlined`), Suspend/Unsuspend (`Icons.pause_circle_outline` / `Icons.play_circle_outline`) and Delete (`Icons.delete_outline`), each with a tooltip.
- Wide (`CardDetailsPane` inside `BrowseScreen`): the same three buttons in a small header row at the top of the pane, since the pane has no app bar.
- The pane gets a `showActions` flag. `CardDetailsScreen` sets it to false and puts the actions in its own app bar instead. Both paths call one `CardActions` helper (an action list built from `CardDetails`), so the labels and behavior cannot drift apart.

**5. Edit focuses the inline editor and opens no new screen.**
The editor moves directly under the answer. Edit calls `Scrollable.ensureVisible` on the editor and requests focus on the first field through a `FocusNode` that `NoteFieldsEditor` exposes.
*Alternative:* push `NoteEditorScreen(noteId:)`, like the study screen does. That was rejected because it gives two editors for the same content in the browser, and the spec already requires an inline editor.

**6. Selection mode becomes explicit state.**
A `_selecting` bool is added. A Select toolbar button (`Icons.checklist`, tooltip "Select") and a long-press both set it. Clearing the selection with the close button resets it. Deselecting the last card keeps the mode active, so the user can keep picking without long-pressing again. With nothing selected, the bulk actions are disabled and the title reads "Select cards".

**7. After a delete.**
- Wide: no extra code. The change event triggers `_search()`, which drops `_details`.
- Narrow: `CardDetailsScreen` pops itself after a successful delete, before the reload tries to fetch a card that no longer exists.

**8. Strings.**
The Go editor error in `core/notes.go` changes from "This note would produce no cards…" to "This would produce no cards…". The Dart labels change as the proposal lists. Full list:

| Where | Old | New |
|---|---|---|
| Browse toolbar tooltip | Add note | Add card |
| Deck row menu | Add note | Add card |
| Empty state | Add your first note… / Add a note | Add your first card… / Add a card |
| Editor title (add / edit) | Add / Edit note | Add card / Edit card |
| Editor button (add) | Add note | Add card |
| Study menu | Edit note | Edit card |
| Details section / button / toast | Note / Save note / Note saved | Edit card / Save / Card saved |
| Bulk menu | Delete notes | Delete cards |
| Delete dialog | Delete N notes? | Delete N cards? |

## Risks / Trade-offs

- [A user deletes one direction of a reversed card and loses both] → The confirmation names the extra cards explicitly. Deleting a single card while keeping its note is out of scope. In Anki, that is done by editing the note type.
- [Tests and the integration flow find widgets by the old strings] → They are updated in the same change. `fvm flutter test` and the integration flow must pass.
- [The `NoteCount` removal breaks the generated API for any other caller] → `grep` shows only `browse_screen.dart` and `svc_test.go` use it, and both are updated.
