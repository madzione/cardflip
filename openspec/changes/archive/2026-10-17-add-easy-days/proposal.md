# Proposal

## Why

Many learners have days with less study time, typically weekends. Anki's Easy Days let users mark weekdays as reduced or minimum, and the scheduler then moves reviews away from those days within the range it is allowed to choose from. Cardflip schedules without regard to weekdays.

## What Changes

- Deck option **Easy Days**: for each weekday, Normal (default), Reduced or Minimum.
- When a review interval is chosen, the due day is picked from the fuzz range, and days marked Reduced receive fewer reviews and days marked Minimum as few as possible; Normal days are preferred, then Reduced, then Minimum, weighted by how many reviews are already due on each day.
- The choice stays within the fuzz range, so intervals change by at most a few days and memory states are unaffected.
- Previews show the day actually chosen.
- Only the scheduling of new answers is affected; existing due dates are not moved.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the Easy Days option.
- `scheduling`: new requirement for choosing due days with Easy Days.

## Impact

- Go: `core.DeckOptions` gains `EasyDays [7]int` (0 normal, 1 reduced, 2 minimum); `scheduler.Params` gains the weekday weights and a due-load lookup (`func(day int) int`), provided by `Collection` from a per-day count of review due dates kept up to date on writes; `fuzz` becomes a weighted choice within the same range, still seeded per card so previews equal answers.
- Flutter: `deck_options_dialog.dart`.
- Tests: `golang/scheduler/fsrs_test.go`, `golang/core/queue_test.go`, `golang/svc_test.go`.
