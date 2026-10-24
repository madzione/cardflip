# Tasks

## 1. Go API

- [x] 1.1 Replace `BrowseService.NoteCount` with `DeleteImpact(cardIDs) (DeleteImpact{Notes, Cards int}, error)` in `golang/svc_browse.go`, counting cards with `core.Collection.CardsOfNote`. Update `golang/svc_test.go` to assert that deleting one card of a reversed note reports `{1, 2}`, and verify that `go test ./...` passes.
- [x] 1.2 Add `NoteCards int` (`json:"noteCards"`) to `CardDetails` and fill it in `BrowseService.Card`. Add a `svc_test.go` assertion that a reversed note's card reports 2, and verify that `go test ./...` passes.
- [x] 1.3 Reword the three `noCardsError` messages in `golang/core/notes.go` from "This note would produce no cards…" to "This would produce no cards…". Verify that `go test ./...` passes (the tests match "would produce no cards").
- [x] 1.4 Run `fluffy generate` to regenerate `lib/golang/api.g.dart`, and verify that `fluffy generate --check` is clean.

## 2. Card details actions

- [x] 2.1 Move `NoteFieldsEditor` in `CardDetailsPane` directly under the answer, retitle its section "Edit card", rename its button to "Save" and its toast to "Card saved", and show "Changes also apply to N other card(s) from this note" when `noteCards > 1`. Update the `test/browse_test.dart` edit test for the new labels, add a test for the shared-note hint, and verify that `fvm flutter test test/browse_test.dart` passes.
- [x] 2.2 Add a shared delete-confirmation helper, "Delete N cards?" with an "including M other cards from the same notes" line when the impact has more cards than were picked, built on `browse.deleteImpact`. Cover both wordings in `test/browse_test.dart`.
- [x] 2.3 Add the card actions (Edit, Suspend/Unsuspend by state, Delete) as a header row in `CardDetailsPane` on the wide layout, and in the `CardDetailsScreen` app bar on narrow layouts, through a `showActions` flag. Suspend/Unsuspend calls `browse.setSuspended([id], …)`, Delete calls the helper from 2.2 and then `browse.deleteNotes([id])`, and the narrow screen pops after a delete. Add widget tests for unsuspend from details (row and action update), delete from details on desktop and phone sizes (details close), and cancel (no `DeleteNotes` call). Verify that they pass.
- [x] 2.4 Make Edit scroll the inline editor into view and focus its first field, through a `FocusNode` exposed by `NoteFieldsEditor`. Add a widget test asserting that `details-field-0` has focus after tapping Edit.

## 3. Browse selection

- [x] 3.1 Add explicit `_selecting` state to `BrowseScreen` with a "Select" toolbar button (`Icons.checklist`). Long-press still enters the mode, the close button leaves it, and with nothing selected the title is "Select cards" and the bulk actions are disabled. Add a widget test that selects two rows without long-press and runs a bulk action, and verify that the existing long-press tests still pass.
- [x] 3.2 Rename the bulk menu item to "Delete cards" and route it through the helper from 2.2. Update the "deleting notes needs confirmation" test (`browse.DeleteImpact` fake, "Delete 2 cards?") and verify that it passes.
- [x] 3.3 Change the Browse toolbar "Add note" tooltip to "Add card", and verify it with a `find.byTooltip('Add card')` assertion.

## 4. Wording elsewhere

- [x] 4.1 In `note_editor_screen.dart`, set the title to "Add card" / "Edit card" and the add button to "Add card". Update `test/editor_test.dart`, `test/editor_edit_test.dart` and `integration_test/flow.dart`, and verify that `fvm flutter test` passes for those files.
- [x] 4.2 In `study_screen.dart`, change the "Edit note" menu item to "Edit card". Update `test/study_test.dart` and verify that it passes.
- [x] 4.3 In `deck_list_screen.dart`, change the deck menu to "Add card" and the empty state to "Add your first card, …" / "Add a card". Update `test/deck_list_test.dart` and verify that it passes.
- [x] 4.4 Grep `lib/` for the user-facing strings `'… note…'` and confirm that only "Note type" (and import/backup screens, which are out of scope) remain.

## 5. Integration check

- [x] 5.1 Run `tool/check.sh` (go vet native and wasm, go test, `fluffy generate --check`, flutter analyze, flutter test) and verify that it ends with "All checks passed."
- [x] 5.2 Run the app on Linux (`fluffy run -d linux`) and check by hand: suspend a card while studying, then in Browse open it, unsuspend it from the details, edit it with the Edit action, enter multi-select with the Select button, and delete a reversed card and see the sibling count in the confirmation.

## Workflow follow-up

- Archive the change with `/opsx:archive` once it is implemented and reviewed, then check that `openspec/specs/card-browser` and `openspec/specs/app-shell` contain the updated requirements.
