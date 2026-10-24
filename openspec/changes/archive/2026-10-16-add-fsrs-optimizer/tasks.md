# Tasks

## 1. Optimizer

- [x] 1.1 Add history extraction for training (per-card sequences, same-day merge, ignore-before date, historical retention for missing history) in `golang/scheduler/optimizer.go`. Add tests on a small fixture log. Verify `go test ./scheduler/...` passes.
- [x] 1.2 Implement pre-training and Adam training with analytic gradients and clamping (design decision 1). Add tests: recover known parameters from synthetic histories within tolerance; match an fsrs-rs reference result on a fixture within tolerance. Verify `go test ./scheduler/...` passes.
- [x] 1.3 Add log-loss evaluation, "only if better" and the health check. Add tests. Verify they pass.

## 2. Collection

- [x] 2.1 Add the FSRS options to `core.DeckOptions` and the DTO; use `Weights` in `Collection.params` (ignoring a wrong length per design decision 2). Add tests that intervals change with weights. Verify `go test ./core/...` passes.
- [x] 2.2 Implement rescheduling on change (design decision 4) with a count for the confirmation. Add tests for both spec scenarios. Verify `go test ./core/...` passes.
- [x] 2.3 Add `DecksService.Optimize(ctx, deckID)` and `OptimizeAll(ctx)` streaming progress and the result, plus `RescheduleCount`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 3. UI

- [x] 3.1 Add an FSRS section to `deck_options_dialog.dart` (parameters, search, ignore-before date, historical retention, Optimize, Optimize all, health check, reschedule switch with confirmation, progress and cancel). Add widget tests for success, not enough reviews, and cancel. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Time optimizing 100,000 synthetic reviews on Linux and in Chrome.
