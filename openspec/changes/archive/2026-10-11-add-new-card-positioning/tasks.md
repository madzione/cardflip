# Tasks

## 1. Collection

- [x] 1.1 Add `Collection.Reposition(ids, start, step, randomize, shift)` returning how many cards were skipped. Add `cards_test.go` cases for both reposition scenarios and that queue order follows. Verify `go test ./core/...` passes.
- [x] 1.2 Add `InsertionOrder` to `core.DeckOptions` and the DTO; assign positions per option in `AddNote` and `Import`; reposition existing new cards when the option changes. Add tests for the insertion order scenarios. Verify `go test ./core/...` passes.
- [x] 1.3 Add the position to `Row` and `CardDetails` and `BrowseService.Reposition`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the reposition dialog to the browser bulk actions, the position text in rows and details, and the option in `deck_options_dialog.dart`. Add `test/browse_test.dart` tests for the dialog and the `New #42` text. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
