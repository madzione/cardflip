# Proposal

## Why

Anki lets users tune what the review screen shows: hide the remaining counts (some find them stressful), hide the next-interval labels on the buttons (to avoid being influenced), turn off Space/Enter answering Good (to prevent accidental ratings), hide the top and bottom bars for a distraction-free screen, fade in the question, and scroll long answers to the answer divider. Cardflip always shows everything and has none of these behaviours.

## What Changes

- Settings in a new **Review** section:
  - **Show remaining card count** (on by default).
  - **Show next review time above answer buttons** (on by default).
  - **Space/Enter also answers Good** (on by default; when off, they only reveal the answer).
  - **Hide bars during review** (off by default): the app bar and navigation are hidden while studying; the counts and buttons stay; a tap at the top edge or Escape still leaves.
  - **Question fade** in milliseconds (default 100, 0 = off).
- On reveal, when the answer contains an element with `id="answer"` (as in the default `<hr id=answer>`), the card scrolls so that element is at the top of the view, if the card is taller than the screen.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `study-session`: "Keyboard control" and "Rate with interval previews" honour the new settings; new requirements for count visibility, hidden bars, question fade and scrolling to the answer.
- `app-shell`: new requirement for the Review settings.

## Impact

- Go: `core.Settings` gains the five settings (absent = defaults) with setters.
- Flutter: `study_screen.dart` (counts, button labels, key handling, hidden bars, fade, scroll), `lib/app/scaffold.dart` (hiding navigation while a session is open), `settings_screen.dart`; `card_view.dart` reports the position of `id=answer`.
- Tests: `golang/svc_test.go`, `test/study_test.dart`, `test/settings_test.dart`.
