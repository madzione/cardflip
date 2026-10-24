# Proposal

## Why

Learning and relearning steps are the deck option Anki users change most: `1m 10m 1h 1d` for slow learners, `10m` only for quick decks, or no steps at all to let FSRS decide. Cardflip hard-codes 1m/10m and 10m in `scheduler.DefaultParams`, so users cannot adapt how new and forgotten cards are drilled. The original design already noted that exposing the steps would be an additive change.

## What Changes

- Deck options **Learning steps** and **Relearning steps**: space-separated durations with units `s`, `m`, `h`, `d` (for example `1m 10m 1d`), defaults `1m 10m` and `10m`.
- Empty learning or relearning steps let FSRS decide the short-term delays: cards follow FSRS's short-term scheduling until they reach a delay of a day or more. (Today, empty steps graduate immediately; with configurable steps that becomes a user-visible choice, so it follows Anki.)
- Steps of a day or more, or steps that end after the study day ends, make the card an **interday learning** card: due on a day rather than at a time, shown from the start of that study day, counted in the learning count, and limited by the review limit.
- Deck option **Interday learning/review order**: mix with reviews (default), show before reviews, or show after reviews.
- Deck option **Bury interday learning siblings** (off by default), as Anki has.
- Learning step fuzz: intraday step delays get a small random extra delay (up to 25 % of the step, at most 5 minutes), seeded like interval fuzz so the preview equals the result. This keeps cards learned together from always returning in the same order.
- Changing steps applies to each card's next answer; cards in the middle of a step keep their current due time.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `scheduling`: "Learning steps" becomes configurable with fuzz; "Lapses" depends on relearning steps; new requirement for interday learning.
- `collection`: new requirement for the step deck options.
- `study-session`: new requirement for interday learning order.

## Impact

- Go: `core.DeckOptions` gains `LearningSteps`, `RelearningSteps` (as durations, absent = defaults), `InterdayLearningOrder`, `BuryInterdayLearning`; `Collection.params` builds `scheduler.Params` from them; `scheduler/fsrs.go` handles empty steps with go-fsrs short-term scheduling, interday steps (due at day start) and step fuzz; `core/queue.go` handles interday learning cards (due day, review limit, order).
- Flutter: `deck_options_dialog.dart` (step fields with validation and help).
- Tests: `golang/scheduler/fsrs_test.go`, `golang/core/queue_test.go`, `golang/svc_test.go`, `test/study_test.dart` (where previews change), `test/deck_list_test.dart`.
- Overlap: `add-burying-options` adds switches for new and review siblings; `add-display-order-options` modifies the queue order; both touch `core/queue.go` near the interday changes.
