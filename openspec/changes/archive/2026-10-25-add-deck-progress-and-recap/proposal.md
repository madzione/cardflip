# Proposal

## Why

Streaks and achievements (`add-streaks-and-daily-goal`, `add-achievements`) say how consistently you study, but not how far along a deck is or how the last week went. The statistics answer both, but nobody opens them every week. This is phase 3 of `docs/proposals/gamification.md`: it brings deck progress and a weekly look back to where people already are, and adds a small reward for reaching the goal.

## What Changes

- **Deck progress on the overview**: the deck overview shows a bar of the deck's new, learning, young and mature cards (suspended left out) with the mature share ("62 % mature").
- **Deck progress on the deck list (optional)**: a setting *Show deck progress in the deck list* (off by default) adds a thin bar under each deck.
- **Weekly recap**: on the first open of a new week, a dismissible card at the top of the deck list shows last week's study days, answers, study time and true retention, each compared to the week before. Tapping it opens the statistics. Dismissing it hides it until the next week, on this device.
- **Celebrations**: a short, subtle animation on the study summary when the daily goal was reached or an achievement was earned in the session. Off with *Reduce motion*, and never during studying.
- All of it is hidden when *Show streaks and goals* is off, except deck progress on the overview, which is plain deck information.
- Deck achievements ("Every card of *Spanish* learned") are left for later, depending on how phase 2 is received.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `progress`: new requirements for the weekly recap and celebrations.
- `app-shell`: "Deck overview" shows deck progress; new requirement for the optional deck list progress bar and its setting.

Requires `add-streaks-and-daily-goal` and `add-achievements` to be archived first.

## Impact

- **Go**: per-deck card state counts (new, learning, young, mature, suspended, ignoring burying) in the deck tree (`golang/svc_decks.go`, `golang/core/decks.go`), a `ProgressService.WeeklyRecap()` reusing `stats.TodaySummary` over a week and `stats.TrueRetention`, new local settings `DeckProgressInList` and `RecapDismissedWeek`.
- **Flutter**: progress bar widget on the deck overview (`lib/features/decks/deck_overview_screen.dart`) and deck list rows, recap card on the deck list, celebration animation on the study summary, setting, strings.
- **Tests**: Go tests for counts and the recap, widget tests for bar, recap and celebration (including reduce motion).
