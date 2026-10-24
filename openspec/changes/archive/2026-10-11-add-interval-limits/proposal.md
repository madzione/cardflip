# Proposal

## Why

Some users want to see every card at least once or twice a year (exam preparation, professional knowledge), so they cap the maximum interval. Others want forgotten cards to come back after more than one day once relearned. Anki exposes both as deck options; Cardflip hard-codes the maximum at 36,500 days and has no minimum interval after relearning.

## What Changes

- Deck option **Maximum interval** in days (1–36,500, default 36,500): no review interval exceeds it, including the Easy interval and fuzz.
- Deck option **Minimum interval** after relearning in days (1–36,500, default 1): a card that graduates from relearning gets at least this interval.
- Both apply to the next answer; existing due dates are not changed.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the two interval options.
- `scheduling`: new requirement for applying the limits.

## Impact

- Go: `core.DeckOptions` gains `MaxInterval` and `MinRelearnInterval` (absent = defaults); `scheduler.Params` gains `MinimumRelearnInterval`; `scheduler/fsrs.go` applies it when graduating from Relearning (and when relearning steps are empty).
- Flutter: `deck_options_dialog.dart`.
- Tests: `golang/scheduler/fsrs_test.go`, `golang/core/decks_test.go`, `golang/svc_test.go`.
