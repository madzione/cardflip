# Tasks

## 1. History data

- [ ] 1.1 Extend the review index build in `golang/core/reviewindex.go` with `progressHistory` (counted answers with time and duration, first time each card's `IvlAfter` reached 21) and keep it current in `applyReviews`; verify with core tests for answer, undo, import and sync apply
- [ ] 1.2 Fold `progressHistory` into per-day answers and minutes for the current rollover hour in `ProgressService`, cached like the study day set; verify with a test that changes the rollover hour
- [ ] 1.3 Extend the phase 1 benchmark to include the history aggregates with 200,000 reviews; verify time and memory natively and record the WebAssembly result

## 2. Rest days

- [ ] 2.1 Add the rest day allowance to `Streaks` and pass `RestDaysPerWeek` from all callers; verify table tests: 0, 1 and 2 rest days, two skips in one window, skips at the run's edges, at risk today, longest streak with rest days
- [ ] 2.2 Add `RestDaysPerWeek` to `core.Settings` and `syncedSettings`, validated to 0 to 2; verify `settings_sync_test.go` and a backup round trip

## 3. Achievements

- [ ] 3.1 Add `golang/stats/achievements.go` with the definitions table and `Eval` for Answers, Streak, Study days, Daily goal, Long-term memory and Comeback; verify a table test per achievement for progress, each tier's earned day, no goal, undo below a threshold, and a card counted once for Long-term memory
- [ ] 3.2 Add `ProgressService.Achievements()` and `NewAchievements` on `Progress`, plus `MarkSeen`; add `SeenAchievements` to `localSettings` with the first-evaluation rule from design decision 4; verify service tests for new, seen, first evaluation after upgrade, and more than three at once
- [ ] 3.3 Add an import test with dated Anki history and a sync end-to-end test in `golang/sync_e2e_test.go` where one device earns a tier; verify historical earned dates and that each device reports it new once
- [ ] 3.4 Regenerate the fluffy API with `fluffy generate` (no `dart format` on `lib/golang/`); verify the Flutter build compiles

## 4. Flutter

- [ ] 4.1 Add the `AchievementBadge` widget (rim per tier, disc colour and Material icon per achievement, stars, locked state, minimalist variant) following `assets/badges/achievement-badges.png`; verify golden tests for each tier and the locked state in light and dark themes
- [ ] 4.2 Add `lib/features/progress/achievements_screen.dart` with earned dates and progress, opened from the deck list menu and the Streaks statistics section; verify widget tests for nothing earned, partly earned, and hidden when *Show streaks and goals* is off
- [ ] 4.3 Add the "New: …" line to the study summary and the deck list snackbar (single and combined); verify widget tests that each notice shows once and marks the tiers seen
- [ ] 4.4 Add *Rest days per week* to Settings → Goals and streaks; verify a widget test that the value saves and the header streak updates
- [ ] 4.5 Add English and German strings for achievement names, tier names, progress, notices and the setting; verify `no_literal_text_test` and l10n generation pass

## 5. Specs, docs and integration

- [ ] 5.1 Extend `integration_test/flow.dart` to open the achievements screen after the study flow; verify with `LANG=en_US` on Linux
- [ ] 5.2 Update `README.md` with achievements; verify the text matches the `progress` spec
- [ ] 5.3 Run `openspec validate add-achievements --strict` after `add-streaks-and-daily-goal` is archived; verify it passes

## Workflow follow-up

- Archive after review; `add-deck-progress-and-recap` builds on it for celebrations of earned achievements.
