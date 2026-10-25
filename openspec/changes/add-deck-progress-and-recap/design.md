# Design

## Context

- `DecksService.Tree()` (`golang/svc_decks.go`) builds the deck list rows from `c.Decks()`, `c.DeckCounts()` (due counts) and `c.CardsPerDeck()` (own card totals), then adds each deck's totals to its ancestors. It runs on every deck list reload.
- `stats.CardCounts` counts a scope's cards by state, but puts buried cards in their own bucket and needs a loaded `stats.Data`, which is too heavy for every deck row.
- `stats.TodaySummary(d)` and `stats.TrueRetention(d)` compute answers, time and retention from `stats.Data`; the recap needs them for two past weeks.
- After phases 1 and 2: `ProgressService` with cached study days and per-day answers and minutes, the "Goals and streaks" settings, `HideProgress`, achievements with `NewAchievements`, and the study summary lines.
- Reduce motion is resolved in Flutter from the setting and the platform (`app-shell` "Reduce motion").

## Goals / Non-Goals

**Goals:**
- Deck progress for every deck row without a noticeable cost on deck list reloads.
- Reuse the statistics' definitions (young below 21 days, true retention) so numbers match the statistics screen.

**Non-Goals:**
- Deck achievements.
- A recap history or a recap for arbitrary weeks (the statistics cover that).
- Animations anywhere but the study summary.

## Decisions

### 1. Per-state card counts in the deck tree

Replace `CardsPerDeck()` with `CardStatesPerDeck()` returning, per deck, new, learning (learning and relearning together), young, mature and suspended, ignoring burying, in the same pass over cards. `DeckNode` gains `Progress {New, Learning, Young, Mature}`; `Cards` stays as the sum including suspended. Subdeck totals are added to ancestors as today. The mature share is `Mature / (New + Learning + Young + Mature)`, rounded down to a whole percent, 0 for an empty deck.

*Alternative*: compute on demand per deck from `stats.Data`. Rejected: one statistics load per row.

The same data feeds the overview and the optional deck list bar, so they always agree.

### 2. Weekly recap from the progress history

`ProgressService.WeeklyRecap(firstWeekday int) (*Recap, error)`: `firstWeekday` comes from Flutter (`MaterialLocalizations.firstDayOfWeekIndex` of the interface language) because the Go side does not know the locale. Weeks are ranges of study day numbers. Study days, answers and minutes come from the per-day aggregates of phase 2. True retention needs pass or fail of young and mature reviews, so the recap loads `stats.Data` for the collection over the two weeks only, and calls `stats.TrueRetention`'s counting over a day range (a small helper split out of `TrueRetention`). Returns nil when the previous week has no study days.

It is computed once on the first deck list show of a week, not on every reload.

### 3. Recap dismissal

`RecapDismissedWeek` (int, the first study day number of the dismissed week) in `localSettings`. The deck list shows the recap when the current week's start differs from it. Dismissing stores the current week's start.

### 4. Celebrations

The summary already knows whether the goal was reached in this session (phase 1 compares `Progress` before and after) and which achievements are new (phase 2). If either is true, and reduce motion is not in effect, and progress is shown, it plays a one-shot animation built with Flutter's implicit animations: a brief scale and fade of the goal or badge icon with a few theme-coloured particles drawn by a `CustomPainter`, under 2 seconds, in an `IgnorePointer` overlay so buttons stay usable. No new dependency (no Lottie or confetti package).

### 5. Settings

`DeckProgressInList` (bool, default false) in `localSettings`, shown under Settings → Goals and streaks. Deck progress on the overview is not affected by `HideProgress`; the list bars and the recap are.

## Risks / Trade-offs

- [Deck tree reload gets slower with per-state counting] → the pass already visits every card for totals; a deck tree benchmark with 100,000 cards checks the difference.
- [Recap week start differs between devices with different languages] → acceptable; the recap is per device.
- [Retention load for two weeks on a big collection] → bounded to two weeks of reviews and done once per week.

## Migration Plan

New local settings default to off and 0. No data migration. Rollback leaves unused fields.
