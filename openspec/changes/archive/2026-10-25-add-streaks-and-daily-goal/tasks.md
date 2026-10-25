# Tasks

## 1. Streak algorithm

- [x] 1.1 Add `golang/stats/streak.go` with `Streaks(days []int, today int) StreakStats` (current, at risk, today studied, longest with first and last day); verify with table tests in `streak_test.go`: empty history, only today, today not studied yet, one-day gap, single long run, longest run in the past
- [x] 1.2 Add a Go test that builds reviews across a DST change and a rollover hour of 4 and checks the study days via `scheduler.DayNumber`; verify no gap and no doubled day

## 2. Settings

- [x] 2.1 Add `DailyGoalKind`, `DailyGoal` to `core.Settings` and `syncedSettings`; `HideProgress`, `GoalSuggestionShown` to `localSettings`; verify `settings_sync_test.go` passes and a backup round trip keeps the goal
- [x] 2.2 Validate the goal when saving settings (kind in `""`, `cards`, `minutes`; value 1 to 9,999 when a kind is set); verify with a settings unit test

## 3. Statistics section

- [x] 3.1 Add `stats.Streaks(d, goal, period)` returning current streak, longest streak with dates, study days in period, days in period and goal days; verify with tests on the fixture in `golang/stats/fixture_test.go`, including previews and manual entries ignored and a deck scope
- [x] 3.2 Add `StatsService.Streaks(scope, period)` in `golang/svc_stats.go` and regenerate the fluffy API (`fluffy generate`, never `dart format` on `lib/golang/`); verify the Flutter build compiles

## 4. Progress service

- [x] 4.1 Add `ProgressService.Today()` in `golang/svc_progress.go`: study days from the review index (manual entries skipped), today's cards and minutes from the loaded review log months, goal from settings; verify with service tests: no reviews, goal in progress, goal reached, at risk, undo of the only answer of the day, imported Anki history
- [x] 4.2 Cache the study day set and result by review revision, rollover hour and day; invalidate on undo, import, restore and sync apply; verify with a test that changes the rollover hour and one that undoes an answer
- [x] 4.3 Add a test comparing `ProgressService.Today()` with `StatsService.Streaks` for the collection scope on the same history; verify current and longest streak agree
- [x] 4.4 Add a benchmark next to `golang/stats/bench_test.go` with 200,000 reviews for the first call and for a call after one answer; verify a few milliseconds natively and record the WebAssembly result; if it misses, switch to the per-day count fallback from design decision 3 before continuing
- [x] 4.5 Extend `golang/sync_e2e_test.go`: two devices study on different days and the same day, sync, and both report the same study days and streak

## 5. Agents

- [x] 5.1 Add `streak` and `goal` to `get_stats` in `golang/mcp_tools_read.go` and update its description; verify with a case in `mcp_tools_test.go` for goal set and no goal

## 6. Flutter: model, header and summary

- [x] 6.1 Add `lib/data/progress_model.dart` reloading on change notifications and day change; verify with a unit test using `test/support/fake_go.dart`
- [x] 6.2 Add the deck list header row (streak, goal ring, at-risk text, tap opens statistics at Streaks); verify widget tests for no goal, goal in progress, goal reached, at risk, no streak, hidden and minimalist
- [x] 6.3 Add the streak and goal lines and the one-time goal suggestion to `_Summary` and the nothing-due screen in `lib/features/study/study_screen.dart`; verify widget tests for streak extended, goal left, goal reached, suggestion dismissed, hidden
- [x] 6.4 Add English and German strings to `app_en.arb` and `app_de.arb` with plural forms for days, cards and minutes; verify `no_literal_text_test` and the l10n generation pass

## 7. Flutter: statistics and settings

- [x] 7.1 Add the Streaks section to `stats_sections.dart` and `stats_pdf.dart`, with scroll-to support for the header link; verify a widget test for the section and that the PDF test includes it
- [x] 7.2 Add Settings → Study → Goals and streaks (*Show streaks and goals*, *Daily goal* with kind and number); verify widget tests that the values save and that hiding removes the header

## 8. Specs, docs and integration

- [x] 8.1 Extend `integration_test/flow.dart` with a study flow that sets a goal and ends with the streak and goal lines on the summary; verify with `LANG=en_US` on Linux
- [x] 8.2 Update `README.md` features list with streaks and daily goal; verify the text matches the `progress` spec
- [x] 8.3 Run `openspec validate add-streaks-and-daily-goal --strict`; verify it passes

## Workflow follow-up

- Archive the change after review, so `add-achievements` can build on the `progress` spec.
