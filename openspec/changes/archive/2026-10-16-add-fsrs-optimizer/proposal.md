# Proposal

## Why

FSRS's main benefit over older schedulers comes from parameters fitted to the user's own review history: with default parameters, intervals are tuned to an average learner. Anki's optimizer fits them in a few seconds. Cardflip uses default parameters only (the original design listed the optimizer as a non-goal), so its scheduling is less accurate than Anki's for the same user and the same reviews.

## What Changes

- Deck option **FSRS parameters**, empty by default (use the defaults), with **Optimize** and **Optimize all decks**.
- Optimizing fits parameters to the review history selected by a search (default: the deck and its subdecks), skipping cards reviewed before an optional **Ignore cards reviewed before** date, and using **Historical retention** (default 0.90) for cards whose earlier history is missing.
- At least 400 reviews are required; with fewer, optimizing explains that there is not enough data and keeps the current parameters.
- New parameters are only kept when they fit the history better than the current ones; the result says whether they changed.
- **Health check**: an optional check after optimizing that reports when the fitted parameters look unreliable (for example, too little or inconsistent data) instead of applying them silently.
- **Reschedule cards on change**: when on, changing desired retention or parameters recomputes the due dates of the affected review cards from their memory states; when off (default), only future answers use the new values.
- Optimizing runs in the background with progress and can be cancelled; the app stays usable.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for the FSRS parameter options and rescheduling on change.
- `scheduling`: "FSRS scheduling" uses the deck's parameters when set; new requirement for optimizing.

## Impact

- Go: new `golang/scheduler/optimizer.go` (FSRS training in pure Go, so it runs in WASM), memory-state recomputation from history with given parameters, `scheduler.Params.Weights`; `core.DeckOptions` gains `Weights []float64`, `IgnoreBefore`, `HistoricalRetention`, `SearchOverride`, `RescheduleOnChange`; `DecksService.Optimize(deckID)` streams progress like imports.
- Flutter: an FSRS section in `deck_options_dialog.dart` with progress and the result.
- Tests: `golang/scheduler/optimizer_test.go` (convergence on synthetic histories generated from known parameters), `golang/core` tests for rescheduling, `golang/svc_test.go`.
- Performance: training on 100,000 reviews must finish within about a minute on a typical desktop, longer on the web; progress keeps the UI informed.
