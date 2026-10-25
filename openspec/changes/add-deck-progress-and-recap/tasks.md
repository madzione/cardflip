# Tasks

## 1. Deck progress data

- [ ] 1.1 Replace `CardsPerDeck()` with `CardStatesPerDeck()` in `golang/core` (new, learning incl. relearning, young, mature, suspended; burying ignored); verify core tests for each state, a buried mature card, and a suspended card
- [ ] 1.2 Add `Progress` to `DeckNode` in `golang/svc_decks.go`, summed into ancestors like `Cards`; verify a service test with subdecks and the 62 % example from the spec
- [ ] 1.3 Add a deck tree benchmark with 100,000 cards; verify the reload time stays within a few percent of the current tree
- [ ] 1.4 Add `DeckProgressInList` and `RecapDismissedWeek` to `localSettings`; verify `settings_sync_test.go` passes

## 2. Weekly recap

- [ ] 2.1 Split the counting in `stats.TrueRetention` into a helper over a day range; verify existing true retention tests still pass and a new test for a 7-day range
- [ ] 2.2 Add `ProgressService.WeeklyRecap(firstWeekday)` with study days, answers, minutes and retention for the previous and the week before; verify tests for a Monday and a Sunday week start, the rollover hour, and no study last week returning nil
- [ ] 2.3 Regenerate the fluffy API with `fluffy generate` (no `dart format` on `lib/golang/`); verify the Flutter build compiles

## 3. Flutter

- [ ] 3.1 Add a deck progress bar widget and show it on the deck overview with "N % mature"; verify widget tests for an empty deck, the spec example, and dark theme
- [ ] 3.2 Show the thin bar on deck list rows when `DeckProgressInList` is on and progress is shown; add the setting to Goals and streaks; verify widget tests for on, off and hidden
- [ ] 3.3 Add the recap card at the top of the deck list (comparisons, tap opens statistics, dismiss stores the week); verify widget tests for new week, dismissed, no study last week, hidden
- [ ] 3.4 Add the celebration overlay on the study summary for goal reached or new achievements; verify widget tests that it plays once, keeps buttons tappable, and does not play with reduce motion, when the goal was reached earlier, or when hidden
- [ ] 3.5 Add English and German strings for the bar, recap and setting; verify `no_literal_text_test` and l10n generation pass

## 4. Specs, docs and integration

- [ ] 4.1 Extend `integration_test/flow.dart` to check the deck overview progress bar after studying; verify with `LANG=en_US` on Linux
- [ ] 4.2 Update `README.md` with deck progress and the weekly recap; verify the text matches the specs
- [ ] 4.3 Run `openspec validate add-deck-progress-and-recap --strict` after the earlier phases are archived; verify it passes

## Workflow follow-up

- Archive after `add-streaks-and-daily-goal` and `add-achievements`.
