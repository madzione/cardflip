# Proposal

## Why

Choosing a desired retention is a trade-off: higher retention means more reviews per day. Users have no way to see that cost in Cardflip before changing the value. Anki's FSRS simulator shows the expected daily workload and knowledge over time for different settings, and "Help me decide" suggests the retention with the best knowledge-per-time ratio.

## What Changes

- A **Simulate** action in the deck options' FSRS section opens a simulator for the deck (and its subdecks):
  - inputs: days to simulate (default 365), new cards per day, maximum reviews per day, desired retention, and optionally a second retention to compare;
  - output: a chart of expected reviews per day and of total memorised cards over the simulated period, and the total review time, starting from the deck's current cards and their memory states.
- **Help me decide** runs the simulation over a range of retentions (0.70–0.99) and suggests the one that minimises review time per memorised card, with the curve shown.
- The simulator does not change anything; applying a suggestion sets the deck's desired retention, which the user can still cancel.
- Simulations run in the background with progress.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for simulating workload and suggesting a retention.

## Impact

- Go: new `golang/scheduler/simulate.go` (day-by-day Monte Carlo simulation with the FSRS model, the deck's parameters or defaults, average answer times and rating distribution from the deck's history, seeded for reproducible results); `DecksService.Simulate` and `SuggestRetention` streaming progress.
- Flutter: a simulator screen with charts (`fl_chart`), opened from `deck_options_dialog.dart`.
- Tests: `golang/scheduler/simulate_test.go`, `golang/svc_test.go`, a widget test for the screen.
