# Tasks

## 1. Recent decks

- [x] 1.1 Add `RecentDecks` to `core.Settings`, updated by `StudyService.Start`, pruned of deleted decks, exposed in the settings DTO. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. Picker

- [x] 2.1 Add `lib/features/decks/deck_picker.dart` with word-order-independent filtering, recent-first ordering, keyboard navigation and study on choose. Add widget tests for the spec scenarios. Verify they pass.
- [x] 2.2 Register `/` in `lib/app/scaffold.dart` (ignored while a text field has focus) and add the toolbar entry in `deck_list_screen.dart`. Add `test/app_shell_test.dart` tests including the text field scenario. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
