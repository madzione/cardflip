# Tasks

## 1. Go

- [x] 1.1 Add `scheduler.Retrievability(mem Memory, elapsed time.Duration) float64` using the FSRS forgetting curve of the library parameters. Add `fsrs_test.go` cases (1.0 at zero elapsed; equal to the desired retention at the scheduled interval within rounding). Verify `go test ./scheduler/...` passes.
- [x] 1.2 Add `BrowseService.CardInfo(cardID)` returning the summary fields and history (kind from `StateBefore`; manual entries when present). Add `svc_test.go` cases for the spec scenarios. Verify `go test ./...` passes.
- [x] 1.3 Track the previously answered card in the study session and expose `StudyService.PreviousCardID()`. Add a `svc_test.go` case, including after undo. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add `lib/features/browse/card_info_screen.dart` (summary and history table, adaptive layout). Add `test/card_info_test.dart` tests for a review card and a new card. Verify they pass.
- [x] 2.2 Open it from `study_screen.dart` (`I`, Ctrl/Cmd+Alt/Option+I, menu) and `card_details.dart`. Add `test/study_test.dart` tests for the two study scenarios. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
