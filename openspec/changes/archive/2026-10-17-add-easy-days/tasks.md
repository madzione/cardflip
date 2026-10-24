# Tasks

## 1. Scheduler

- [x] 1.1 Add weekday levels and a due-load lookup to `scheduler.Params`, and make `fuzz` a seeded weighted choice within the existing range (Normal > Reduced > Minimum, lower load preferred; unchanged when all Normal). Add `fsrs_test.go` cases for the three spec scenarios and preview equality. Verify `go test ./scheduler/...` passes.

## 2. Collection

- [x] 2.1 Keep a per-day count of review due dates in `core` (updated on every op that changes due dates) and pass the lookup and the deck's levels in `Collection.params`. Add tests that the counts stay correct after answers, undo and imports. Verify `go test ./core/...` passes.
- [x] 2.2 Add `EasyDays` to `core.DeckOptions` and the DTO. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 3. UI

- [x] 3.1 Add an Easy Days row (seven three-way selectors) to `deck_options_dialog.dart` with a widget test. Verify it passes.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Simulate 30 days of reviews in a test with Sunday as Minimum and check Sunday's load.
