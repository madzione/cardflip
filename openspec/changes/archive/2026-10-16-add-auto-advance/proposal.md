# Proposal

## Why

Some users review hands-free or at a fixed pace (listening practice, commuting with audio decks, speed drills). Anki's auto advance shows the answer and moves to the next card after configured times. Cardflip always waits for input.

## What Changes

- Deck options: **Seconds to show question for** and **Seconds to show answer for** (0 = disabled, defaults 0), **Question action** (show answer, or show reminder) and **Answer action** (bury card, answer Again, answer Good, answer Hard, or show reminder; default bury card).
- Auto advance is started per session from the study menu or with Shift+A, and stops when the session ends or the user toggles it off. It never starts on its own, so leaving the phone on the table does not answer cards.
- When a side has sounds playing (if audio playback exists), the countdown starts after they finish.
- Any user input on the card (reveal, rating, edit) resets the countdown for the new state.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the auto advance deck options.
- `study-session`: new requirement for auto advance during study.

## Impact

- Go: `core.DeckOptions` gains the four options; `StudyState` exposes them for the current card's deck.
- Flutter: `study_screen.dart` (toggle, countdown timer, actions, indicator showing auto advance is on).
- Go: the `.apkg` importer and exporter map the four options to and from Anki's preset (auto advance times are rounded to whole seconds on import).
- Tests: `golang/svc_test.go`, `test/study_test.dart`, the `.apkg` fixtures and `tool/check_apkg_export.py`.
- Overlap: `add-media-playback` defines the end of sound playback; without it, the countdown starts when the side is shown.
