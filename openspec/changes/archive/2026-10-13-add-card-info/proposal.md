# Proposal

## Why

When a card feels wrongly scheduled, users open Anki's Card Info to see its full review history and memory state. Cardflip's browser shows only a short scheduling summary, has no review history at all, and nothing can be inspected from the study screen, which is where the question comes up. Cardflip already stores every review, so the data exists.

## What Changes

- A **Card info** view with: added date, first and latest review, due date (or new-card position), interval, stability, difficulty, current retrievability, review count, lapse count, average and total answer time, card type, note type, deck, and card and note IDs.
- Below it, the card's review history, newest first: date and time, kind (learning, review, relearning, manual), rating, interval after the review, and answer time. Manual entries include resets and due-date changes when those features exist.
- Opened from the study screen with `I` (current card) and Ctrl+Alt+I / Cmd+Option+I (the previously answered card), and from a menu item; and from the browser's card details.
- While open from study, the view follows the current card when the session advances, on wide screens where it is shown side by side.

## Capabilities

### New Capabilities

- `card-info`: what the card info view shows and its review history.

### Modified Capabilities

- `study-session`: new requirement for opening card info during study.
- `card-browser`: new requirement for opening card info from the card details.

## Impact

- Go: `BrowseService.CardInfo(cardID)` built from `Collection.Card`, `ReviewsOfCard` and the scheduler's retrievability function (new `scheduler.Retrievability(mem, elapsed)`); `StudyService` remembers the previously answered card ID.
- Flutter: new `lib/features/browse/card_info_screen.dart`, wired into `study_screen.dart` and `card_details.dart`.
- Tests: `golang/scheduler/fsrs_test.go`, `golang/svc_test.go`, new `test/card_info_test.dart`, `test/study_test.dart`.
