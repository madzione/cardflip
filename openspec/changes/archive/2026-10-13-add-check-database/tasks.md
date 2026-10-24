# Tasks

## 1. Check

- [x] 1.1 Add `golang/core/check.go` with one detector and repair per problem kind, collecting repairs into one op and a report. Add `check_test.go` with a corrupted fixture per kind (built by applying raw ops that bypass validation) and the healthy scenario. Verify `go test ./core/...` passes.
- [x] 1.2 Add `AppService.CheckDatabase(ctx)` with progress, committing the op and compacting afterwards, and the pre-operation backup hook when available. Add `svc_test.go` cases, including a failed save leaving the collection unchanged. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add Check database to `settings_screen.dart` with progress and the report dialog. Add `test/settings_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Run the check on the 50,000-card fixture and record the time. Recorded: about 400 ms for 50,000 cards, including compaction (`TestCheckDatabase50k`, 2026-10-13).
