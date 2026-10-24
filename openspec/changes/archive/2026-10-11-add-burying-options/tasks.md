# Tasks

## 1. Queue

- [x] 1.1 Add the two options to `core.DeckOptions` (absent = on) and the DTO, and use them in the sibling check in `core/queue.go`. Add `queue_test.go` cases for both options off, and defaults unchanged. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add a Burying section to `deck_options_dialog.dart` with a widget test. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
