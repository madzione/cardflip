# Proposal

## Why

Spaced repetition only works if you study every day, and Cardflip gives no feedback on that. The study summary says "Session complete" and the statistics show a calendar heatmap, but nothing says that today is your 40th day in a row or how far you are from what you meant to do today. Anki users get this from add-ons; Cardflip has no add-on system, so it has to be built in. This is phase 1 of `docs/proposals/gamification.md`. It sets up the parts that phases 2 (`add-achievements`) and 3 (`add-deck-progress-and-recap`) build on.

## What Changes

- **Study days and streaks**: a study day is a day, by the rollover hour, with at least one answer in the review log. Previews and manual entries (set due date, reset) don't count. The current streak counts back from today, or from yesterday when today has no answers yet. The longest streak is kept. Both are computed from the review log, not stored, so they sync, survive backups and imports (including Anki history), follow undo, and are separate for each profile.
- **Daily goal**: none (default), a number of cards, or a number of minutes. Progress counts today's answers or study time. The streak does not depend on the goal: any study day keeps it.
- **Deck list header**: a slim row above the decks with the current streak and, when a goal is set, a progress ring. When today has no answers yet and the streak is still alive, it says "Study today to keep your 12-day streak", without warning colours. Tapping it opens the statistics on the Streaks section.
- **Study summary**: the session-complete and "nothing due" screens get up to two lines: the streak ("Streak extended to 13 days" the first time today) and the goal ("Daily goal reached" or "12 more cards to reach today's goal"). With no goal set, the summary suggests setting one once.
- **Statistics**: a new Streaks section with current streak, longest streak, study days in the period and days the goal was met, for every scope, also in the PDF.
- **Settings → Study → Goals and streaks**: *Show streaks and goals* (on by default, per device) and *Daily goal* (synced).
- **Agents**: `get_stats` also returns the current streak and today's goal progress.
- Nothing rewards ratings, speed or adding new cards, and there are no leaderboards, penalties or loss messages.

## Capabilities

### New Capabilities

- `progress`: study days, streaks, the daily goal, the deck list header, the study summary lines and the visibility switch. Phases 2 and 3 add achievements, rest days, the weekly recap and celebrations to it.

### Modified Capabilities

- `statistics`: new "Streaks" requirement; "Save statistics" includes the Streaks section in the PDF.
- `study-session`: "Session completion" and "Start a session" (nothing due screen) show the streak and goal lines.
- `device-sync`: "What is synced" includes the daily goal and not the visibility switch.
- `agent-access`: "Collection statistics tool" returns the current streak and today's goal progress.

## Impact

- **Go**: `golang/stats/streak.go` (pure streak function over study days), a Streaks section in `golang/stats` and `StatsService` (`golang/svc_stats.go`), a new `ProgressService` (`golang/svc_progress.go`) using the review index (`golang/core/reviewindex.go`) and the loaded review log months, new `core.Settings` fields `DailyGoalKind`, `DailyGoal` (synced) and `HideProgress` (local), `get_stats` in `golang/mcp_tools_read.go`. Regenerated fluffy API.
- **Flutter**: `lib/data/progress_model.dart`, deck list header (`lib/features/decks/deck_list_screen.dart`), summary lines (`_Summary` and the nothing-due screen in `lib/features/study/study_screen.dart`), Streaks section (`lib/features/stats/stats_sections.dart`, `stats_pdf.dart`), settings, English and German strings.
- **Data**: no new stored data apart from three settings; no snapshot version change.
- **Performance**: the deck list and summary need today's progress after every answer; a benchmark with 200,000 reviews (native and WebAssembly) guards it.
- **Tests**: Go unit, sync end-to-end, Flutter widget and one integration flow.
