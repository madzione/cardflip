# Design

## Context

- The statistics (`golang/stats`) load a scope's reviews into `stats.Data`, each with its study day (`scheduler.DayNumber`). The loader leaves out manual entries (rating 0) and preview answers. `StatsService` (`golang/svc_stats.go`) exposes one call per section through the `section` helper. The loader decodes the whole review log and is invalidated by every answer.
- The collection keeps two kinds of review data in memory: the review log months around today (`loadRevlog` in `golang/core/reviews.go`), and, once a history search needs it, the review index (`golang/core/reviewindex.go`): the time and rating of every answer of the whole log, previews left out, updated by `applyReviews` on every op. It holds times, not days, so a rollover change needs no rebuild.
- `core.Settings` fields are split into `syncedSettings` and `localSettings` (`golang/core/settings_sync.go`); a test checks every field is in exactly one list.
- On the Flutter side `DeckListModel` reloads on change notifications (`lib/data/changes.dart`). The study summary is `_Summary` in `lib/features/study/study_screen.dart`. Statistics sections live in `lib/features/stats/stats_sections.dart` and the PDF in `stats_pdf.dart`. All UI text comes from `app_en.arb` / `app_de.arb`, enforced by `no_literal_text_test`.
- The phased plan and the principles (derived not stored, never reward ratings or speed, quiet and optional) are in `docs/proposals/gamification.md`.

## Goals / Non-Goals

**Goals:**
- One streak algorithm, shared by the statistics section and the deck list, so the two never disagree for the collection scope.
- Today's progress after each answer in a few milliseconds with 200,000 reviews, natively and in WebAssembly.
- A shape that phase 2 (achievements, rest days) can extend without changing the streak function's callers.

**Non-Goals:**
- Per-deck streaks on the deck list or deck overview (statistics show them per scope).
- A stored history of goals.
- Notifications or reminders.

## Decisions

### 1. Streaks are a pure function of study days

`golang/stats/streak.go`: `Streaks(days []int, today int) StreakStats`, with `days` a sorted, de-duplicated set of study day numbers. It returns the current streak, whether it is at risk (today not in the set), the longest streak with its first and last day, and whether today is studied. Both the statistics section and `ProgressService` call it.

*Why*: a pure function is trivially testable for the edge cases (rollover, DST, gaps, empty history), and phase 2 can add a rest-day allowance as a parameter. *Alternative*: computing streaks inside each caller from its own data; rejected because the deck list and the statistics would drift.

DST is handled by working in study day numbers from `scheduler.DayNumber`, never by subtracting 24 h.

### 2. Statistics section uses `stats.Data`

`stats.Streaks(d *Data, goal Goal, period string) StreakSection` collects the study days of the scope from `d`, calls `Streaks`, counts study days in the period and, with a goal, the days whose cards or minutes met it. `StatsService.Streaks(scope, period)` uses the existing `section` helper. Because `stats.Data` already filters previews and manual entries, the study day rule comes for free.

### 3. A separate, cheap `ProgressService` for the deck list and summary

`golang/svc_progress.go`, `ProgressService.Today() (Progress, error)`:

- Study days: walk the review index, skip entries with an invalid rating (manual entries), map each time to `DayNumber` with the current rollover hour, collect into a set.
- Today's cards and minutes: from the loaded review log months (today is always loaded), the same filter as the statistics.
- Goal: from settings.
- `Progress` holds current streak, at risk, today studied, longest streak, today's cards, today's minutes, goal kind, goal value and reached. "Streak extended to …" is decided in Flutter: the study screen keeps the `Progress` it saw when the session started and compares today studied before and after.

The study day set is cached by (collection revision of the review data, rollover hour); the result by that plus today's day number. An answer adds at most today to the set, so after the first build the update is incremental: if today is already in the set nothing changes; otherwise insert it. Undo, import, restore and sync invalidate the cache.

*Why not reuse the statistics loader*: it decodes every review with its card state and is thrown away on every answer. *Alternative*: a stored per-day counter updated on answer, undo, import and sync apply. Kept as the fallback if the benchmark misses its target, because it is new stored data that sync must merge.

The review index is loaded lazily today. `ProgressService` triggers its load the first time it is asked, after the deck list has shown; until then the header shows nothing rather than a wrong streak.

### 4. Settings

New `core.Settings` fields:

| Field | List | Why |
|---|---|---|
| `DailyGoalKind` (`""`, `"cards"`, `"minutes"`) | synced | the goal is the user's intent for the collection |
| `DailyGoal` (int) | synced | |
| `HideProgress` (bool, default false so streaks show) | local | how this device looks, like `Minimalist` |
| `GoalSuggestionShown` (bool) | local | one-time notice, like `PersistenceNoticeShown` |

The visibility setting is stored inverted (`HideProgress`) so the zero value means "shown" for existing collections and backups.

### 5. Goal days use the current goal

The statistics judge past days against the current goal (see proposal and the `progress` spec). No goal history is stored. The streak ignores the goal, so changing it never looks like lost progress.

### 6. Flutter

- `lib/data/progress_model.dart`: a `ChangeNotifier` holding the last `Progress`, reloading on the same change notifications as `DeckListModel` and on the day change timer already used for due counts.
- Deck list: a `ProgressHeader` row as the first item of the deck list (not the app bar, which is full on phones). Flame icon (`Icons.local_fire_department`) and a small `CircularProgressIndicator`-based ring, colours from the theme. Hidden when `HideProgress` is set.
- Summary: `_Summary` and the nothing-due screen read `ProgressModel` and add up to two lines plus the one-time goal suggestion.
- Statistics: a Streaks section in `stats_sections.dart` and the PDF. The deck list header opens the statistics scrolled to it.
- Settings: a "Goals and streaks" group in the Study section.

### 7. MCP

`get_stats` adds `streak` (current, at risk, today studied) and `goal` (kind, value, progress, reached, or null) from `ProgressService`, for the collection regardless of the deck argument, since the streak is collection-wide.

## Risks / Trade-offs

- [Review index load on large collections, especially on the web] → load after first frame; benchmark `ProgressService.Today` with 200,000 reviews natively and under WebAssembly next to `golang/stats/bench_test.go`; fallback to a stored per-day count (decision 3).
- [Statistics and header disagree because they use different sources] → both call `Streaks`; a Go test builds the same history and compares the collection-scope section with `ProgressService.Today`.
- [Changing the goal changes "days goal met" retroactively] → accepted and specified; the streak is unaffected.
- [Flying east can skip a local day and break a streak] → accepted for now, like the rest of the app's day handling (see open questions).
- [Reviews with clock skew between devices land on an unexpected day] → sync already refuses devices whose clocks disagree (`device-sync` "Clocks must agree").

## Migration Plan

No migration: new settings default to their zero values (no goal, streaks shown). Existing review history immediately yields streaks. Rollback is removing the code; the settings fields are ignored by older versions.

## Open Questions

- Allow one grace day after a time-zone change? Deferred; adding it later only changes `Streaks`.
- A "finish all due cards" goal kind? Deferred to after phase 1; it cannot be judged for past days.
- Per-deck streak on the deck overview? Deferred; the statistics already show it per deck.
