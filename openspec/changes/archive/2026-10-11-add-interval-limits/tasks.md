# Tasks

## 1. Scheduler and options

- [x] 1.1 Add `MinimumRelearnInterval` to `scheduler.Params` and apply it when a Relearning card graduates (and on Again with empty relearning steps), keeping Hard ≤ Good ≤ Easy ordering. Add `fsrs_test.go` cases for both spec scenarios, including with fuzz. Verify `go test ./scheduler/...` passes.
- [x] 1.2 Add `MaxInterval` and `MinRelearnInterval` to `core.DeckOptions` and the DTO with validation, and pass them in `Collection.params`. Add `decks_test.go` and `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add both fields to `deck_options_dialog.dart` with validation. Add widget tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
