# Tasks

## 1. Queue

- [x] 1.1 Add the four enums to `core.DeckOptions` and the DTO with validation (absent = defaults). Add `decks_test.go` cases for defaults and invalid values. Verify `go test ./core/...` passes.
- [x] 1.2 Implement new card gathering and sorting in `core/queue.go` with day-stable seeded randomness, keeping sibling burying and per-deck limits. Add `queue_test.go` cases for each gather and sort order, and that the default order equals today's. Verify `go test ./core/...` passes.
- [x] 1.3 Implement review sort orders (retrievability via the scheduler's forgetting curve; relative overdueness as elapsed days divided by interval) and new/review order in `PickNext`. Add `queue_test.go` cases for the study-session scenarios. Verify `go test ./core/...` passes.
- [x] 1.4 Run `fluffy generate`, update `test/support/fake_go.dart`, and verify `fluffy generate --check` is clean and `go test ./...` passes.

## 2. UI

- [x] 2.1 Add a Display order section to `deck_options_dialog.dart`. Add a widget test that changing an order calls `decks.setOptions` with it. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Check queue building time with the 50,000-card fixture.
