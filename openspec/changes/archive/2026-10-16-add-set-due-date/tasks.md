# Tasks

## 1. Scheduling

- [x] 1.1 Add a parser for `N`, `N-M` and `!` with validation errors, and `Collection.SetDueDate(ids, spec)` applying the rules (state, interval, memory state for New cards, card-seeded randomness for ranges, manual review entries) in one op. Add `cards_test.go` cases for each spec scenario and for learning cards. Verify `go test ./core/...` passes.

## 2. Services and UI

- [x] 2.1 Add `StudyService.SetDueDate(cardID, spec)` and `BrowseService.SetDueDate(cardIDs, spec)`, plus the last value in settings. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 2.2 Add the dialog (syntax help, remembered value, inline error) and wire it into the browser bulk actions, card details and study menu with Ctrl/Cmd+Shift+D. Add `test/browse_test.dart` and `test/study_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
