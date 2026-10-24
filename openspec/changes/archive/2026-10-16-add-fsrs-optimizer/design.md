# Design

## Context

- `scheduler/fsrs.go` uses `fsrs.DefaultParam()` from go-fsrs v3 for memory states and intervals; go-fsrs has no optimizer.
- The review log keeps every answer with the state before it, which is the training data FSRS needs (per card: sequence of (elapsed days, rating)).
- Long operations already stream progress to Dart (imports) and yield in WASM (`internal/yield`).

## Goals / Non-Goals

**Goals:** parameters compatible with go-fsrs's model version; results close to Anki's optimizer on the same data; runs on all platforms including web.

**Non-Goals:** bit-identical results with Anki; GPU acceleration; optimizing across collections.

## Decisions

### 1. Implement training in Go

Port the FSRS training procedure from the reference implementation (fsrs-rs): build per-card review sequences (first review per day only, same-day reviews merged as Anki does), compute the forgetting-curve log loss, and minimise it with Adam over a few epochs using analytic gradients of the FSRS formulas, with parameter clamping as in the reference. Pre-training for the initial stabilities uses the closed-form fit from the reference.

Alternatives: call fsrs-rs through cgo (not possible in WASM); use go-fsrs (no optimizer); finite-difference gradients (too slow for 19+ parameters on large histories).

### 2. Model version pinned to go-fsrs

The parameter count and formulas follow the go-fsrs version in `go.mod`. If go-fsrs is upgraded to a new model version, stored parameters of the old length are treated as absent (defaults) until re-optimized.

### 3. "Only if better"

Evaluate log loss of current and new parameters on the same data; keep the new ones only if lower, as Anki does.

### 4. Rescheduling on change

Recompute each affected review card's memory state by replaying its history with the new parameters, then set its due date from the last review plus the new interval (no fuzz change beyond the existing seeded fuzz). Done in one op; large decks report progress.

## Risks / Trade-offs

- [Port diverges from reference] → Test against synthetic data from known parameters (must recover them within tolerance) and compare to fsrs-rs output on a fixture history.
- [WASM speed] → Run in batches with yields; show progress and allow cancel.
