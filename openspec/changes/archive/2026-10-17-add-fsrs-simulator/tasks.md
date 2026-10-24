# Tasks

## 1. Simulation

- [x] 1.1 Add `golang/scheduler/simulate.go` (seeded day-by-day simulation using FSRS formulas, limits, answer times and rating probabilities). Add `simulate_test.go`: higher retention gives more reviews; same seed gives the same result; zero new cards keep the card count fixed. Verify `go test ./scheduler/...` passes.
- [x] 1.2 Add the retention search (0.70–0.99, cost per memorised card) with a test on a synthetic deck where the minimum is known to be interior. Verify it passes.
- [x] 1.3 Add `DecksService.Simulate` and `SuggestRetention` with progress, extracting answer times and rating distribution from the deck's history (defaults when the history is short). Add `svc_test.go` cases that the collection is unchanged. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add `fl_chart` and a simulator screen (inputs, two-curve chart, totals, Help me decide with apply confirmation) opened from the deck options. Add widget tests for the compare and suggestion flows. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Time a 365-day simulation of a 10,000-card deck on Linux and in Chrome.
