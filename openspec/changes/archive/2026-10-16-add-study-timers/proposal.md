# Proposal

## Why

Cardflip records how long each answer took without limit, so leaving a card on screen during a break records minutes of "answer time" and distorts any time statistics and future optimisation. Anki caps the recorded time (60 s by default), can show an on-screen timer for users who want to answer faster, and has a timebox that reports progress every few minutes so long sessions can be broken up.

## What Changes

- Deck option **Maximum answer seconds** (default 60, 1–9999): the time recorded for an answer is capped at this value.
- Deck options **Show answer timer** (off by default) shows a running timer on the study screen, turning a different colour once the maximum is reached; **Stop timer on answer** (on by default) freezes it when the answer is shown. These affect only the display.
- Setting **Timebox time limit** in minutes (0 = off, default 0): every time the limit passes during a session, a message shows how many cards were studied in that time, with Continue and Finish.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the timer deck options.
- `scheduling`: new requirement that recorded answer time is capped.
- `study-session`: new requirements for the on-screen timer and the timebox.
- `app-shell`: new requirement for the timebox setting.

## Impact

- Go: `core.DeckOptions` gains `MaxAnswerSeconds`, `ShowTimer`, `StopTimerOnAnswer` (absent = defaults); `Collection.Answer` caps the duration with the card's deck option; `core.Settings` gains `TimeboxMinutes`; `StudyState` carries the deck's timer options.
- Flutter: `study_screen.dart` (timer widget, timebox message), `deck_options_dialog.dart`, `settings_screen.dart`.
- Tests: `golang/core/reviews_test.go`, `golang/svc_test.go`, `test/study_test.dart`, `test/settings_test.dart`.
