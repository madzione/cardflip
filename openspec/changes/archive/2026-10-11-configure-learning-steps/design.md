# Design

## Context

- `scheduler.Params` already has `LearningSteps`, `RelearningSteps` and `MaximumInterval`; `answer` handles empty steps by graduating, and `hardDelay` follows Anki. `Collection.params(card)` builds params with `DefaultParams(retention)`.
- The queue (`core/queue.go`) treats every Learning/Relearning card as intraday: due when `Due <= now`, outside daily limits.
- go-fsrs's basic scheduler has a short-term mode that returns minute-level intervals for Again/Hard/Good in learning.

## Goals / Non-Goals

**Goals:** Anki-compatible step semantics, including interday learning; preview equals result.

**Non-Goals:** per-card step overrides; changing steps retroactively for cards already waiting.

## Decisions

### 1. Steps as durations in deck options

Stored as seconds (`[]int`) to avoid parsing in storage; parsed and formatted in Dart (`1m 10m 1d`) with validation (positive, at most 365 days, at most 20 steps). Absent field = defaults, so no migration.

### 2. Empty steps use FSRS short-term

When the relevant step list is empty, the outcome comes from go-fsrs's short-term scheduler (state, delay); once it schedules a delay of at least one day, the card graduates with the FSRS interval as today. This matches Anki's "leave blank" behaviour with FSRS.

### 3. Interday learning

A learning step outcome whose due time is at or after the next day start is converted to a day: `Due` = the start of that study day (rounded like reviews), with a flag `Interday` in `CardState`. The queue treats interday learning cards like reviews for limits and due-ness (due when their day has started), counts them in the learning count, and orders them per `InterdayLearningOrder` relative to reviews. Intraday learning stays first, as now.

### 4. Step fuzz

For intraday step delays: extra = `min(25 % of step, 5 min)` × a value in [0,1) from the same seeded hash as interval fuzz (card ID + reps), so `Preview` and `Answer` agree. Steps under 1 minute get no fuzz.

## Risks / Trade-offs

- [Interday learning changes counts and limits users see] → Only appears when users configure steps of a day or more; defaults keep today's behaviour.
- [go-fsrs short-term output differs from Anki's] → Anki's FSRS short-term is also derived from the FSRS formulas; test that delays are ordered (Again ≤ Hard ≤ Good) and graduate eventually.
