# Tasks

## 1. Scheduler

- [x] 1.1 Replace the empty-steps branch in `scheduler/fsrs.go` with go-fsrs short-term scheduling until a delay of a day or more (design decision 2). Add `fsrs_test.go` cases: ordered delays, eventual graduation, and preview equal to answer. Verify `go test ./scheduler/...` passes.
- [x] 1.2 Add step fuzz (design decision 4) and update `fsrs_test.go` expectations for intraday delays, including the bounded scenario and preview equality. Verify `go test ./scheduler/...` passes.
- [x] 1.3 Convert steps ending on a later study day into interday due days (`CardState.Interday`, design decision 3). Add tests for the one-day step scenario and a `1h` step at 23:30 with rollover 04:00. Verify they pass.

## 2. Options and queue

- [x] 2.1 Add the four options to `core.DeckOptions` and the DTO (absent = defaults, validated) and build `scheduler.Params` from them in `Collection.params`. Add `decks_test.go` and `svc_test.go` cases. Verify `go test ./...` passes.
- [x] 2.2 Treat interday learning cards in `core/queue.go` per design decision 3 (due by day, learning count, review limit, order option, optional sibling burying). Add `queue_test.go` cases for each, and that intraday learning still comes first. Verify `go test ./core/...` passes.
- [x] 2.3 Run `fluffy generate`, update `test/support/fake_go.dart`, and verify `fluffy generate --check` is clean and `fvm flutter test` passes.

## 3. UI

- [x] 3.1 Add step fields (parse/format, inline validation, help text), the order picker and the burying switch to `deck_options_dialog.dart`. Add widget tests for valid and invalid steps. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Manually study with steps `1m 1d` across a rollover.
