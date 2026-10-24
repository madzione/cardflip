# Proposal

## Why

When the only cards left are learning cards due in a few minutes, Cardflip makes the user wait (the waiting screen, or the auto-continue setting). Anki instead shows learning cards up to 20 minutes early once nothing else is due, the "learn ahead limit", so a session can be finished in one go. Users coming from Anki find the waiting screen surprising and often leave with learning cards unfinished.

## What Changes

- A setting **Learn ahead limit** in minutes (default 20, 0 turns it off).
- During a session, when no learning, review or new card is due, a learning card due within the limit is shown right away.
- The waiting screen, the "Session complete" summary's "next learning card at …" and the auto-continue behaviour now apply only to learning cards due later than the limit.
- Deck list counts are unchanged: they still count learning cards that are due now.
- Showing a card early does not change how it is scheduled: the scheduler uses the actual time of the answer.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `study-session`: new requirement for learning ahead; "Session completion" and "Learning cards becoming due while waiting" now refer to learning cards due beyond the limit, and their examples change accordingly.
- `app-shell`: new requirement for the setting.

## Impact

- Go: `core.Settings` gains `LearnAheadMinutes` (absent means 20, so existing collections get Anki's default); `Queue` gains the learning cards due within the limit; `PickNext` uses them when nothing else is due; `StudyState.LearningDue` and the waiting logic use the first learning card beyond the limit.
- Flutter: `settings_screen.dart`; no study screen change beyond what the state already drives.
- Tests: `golang/core/queue_test.go`, `golang/svc_test.go`, `test/study_test.dart`, `test/settings_test.dart`.
- Behaviour change: with the default of 20 minutes, sessions no longer stop for learning cards due within 20 minutes. Users who prefer the old behaviour set the limit to 0.
