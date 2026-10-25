# Design

## Context

- After `add-streaks-and-daily-goal`: `stats.Streaks(days, today)` is a pure function over study day numbers, `ProgressService.Today()` (`golang/svc_progress.go`) builds the study day set from the review index and caches it, and settings have `DailyGoalKind`, `DailyGoal`, `HideProgress` and `GoalSuggestionShown`.
- The review index (`golang/core/reviewindex.go`) decodes the whole stored log but keeps only time and rating per entry (`search.LogEntry`). A `core.ReviewEntry` also has `IvlAfter` and `DurationMs`, which achievements need over the whole history.
- The draft badge designs are in `assets/badges/` (`achievement-badges.vectorcraft` source, `.svg`, `.png`): a scalloped rim in bronze, silver, gold, a coloured disc per achievement with a white Material-style icon, one to three stars for the tier, and a grey locked state with a padlock.
- Motivation and the catalogue are in `proposal.md`; the rules are in the `progress` spec delta.

## Goals / Non-Goals

**Goals:**
- Achievements and their earned dates from history, cheap enough to compute after a session.
- One place to add an achievement: a definitions table in Go, no Flutter change beyond strings and an icon.
- Rest days as a parameter of the existing streak function.

**Non-Goals:**
- Deck achievements (possible in phase 3).
- Bundled image assets per badge.
- Syncing which achievements were shown.

## Decisions

### 1. Per-day aggregates beside the review index

When the review index is built, the same decode also fills a `progressHistory`. It cannot store per-day totals, because days depend on the rollover hour, which may change. So it keeps per counted answer its time and `DurationMs`, and per card the time `IvlAfter` first reached 21. `ProgressService` folds this into per-day answers and milliseconds for the current rollover hour, cached like the study day set in phase 1. `applyReviews` keeps it current, as it does for the index.

*Alternative*: use `stats.Data` (the statistics loader). Rejected for the same reason as in phase 1: it decodes card states too and is dropped on every answer.

Memory: about 24 bytes per answer, ~5 MB for 200,000 answers, acceptable; the benchmark from phase 1 is extended to cover it.

### 2. Definitions table

`golang/stats/achievements.go`:

```go
type Achievement struct {
    ID     string   // "answers", "streak", "study-days", "daily-goal", "long-term", "comeback"
    Tiers  []int    // thresholds; one tier of 1 for Comeback
    Eval   func(h History) (progress int, earned []int) // earned day per reached tier
}
```

`History` holds the sorted study days, per-day answers and minutes, first-mature days, goal and rest days. Each `Eval` walks the days once and records the day each threshold is crossed. Streak uses the longest-streak-so-far as it walks; Daily goal counts days meeting the current goal (none: no progress); Comeback finds the latest study day preceded by a gap of 14 or more days.

### 3. Tier names

Thresholds map to bronze, silver, gold and platinum (fourth tier, for the four-tier achievements). Three-tier achievements stop at gold; Comeback has one tier shown with a gold rim, as in the draft. The badge shows the highest earned tier; stars under the disc show the tier number (one to three, four for platinum).

### 4. Seen state

`SeenAchievements` in `localSettings`: a list of `"<id>:<tier>"` keys this device has shown. `Progress` gains `NewAchievements`: earned tiers not in the list. The summary or the deck list snackbar shows them and calls `ProgressService.MarkSeen(keys)`. More than three new tiers become one "N achievements from your history" message. While `SeenAchievements` has never been set (nil, not empty), the first evaluation marks everything already earned as seen without a notice, so upgrading users are not greeted by a wall of old milestones; the achievements screen still shows them with their dates.

*Alternative*: synced seen state. Rejected because a device that was not in use when the milestone happened should still say so once.

### 5. Rest days

`Streaks(days, today, restPerWeek)`: walking back from today (or yesterday when at risk), a missing day is allowed if adding it keeps every 7-day window of the run at or below `restPerWeek` missing days; rest days do not count toward the length. The longest streak uses the same rule. `RestDaysPerWeek` (0 to 2) joins `syncedSettings`. Phase 1 callers pass the setting.

### 6. Flutter

- `lib/features/progress/achievements_screen.dart`: a responsive grid of `AchievementBadge` widgets.
- `AchievementBadge`: drawn with `CustomPaint` for the scalloped rim (tier colours as theme extension tokens matching the draft palette), a disc with the achievement's colour and a Material icon (`check_box`/`style` for Answers, `local_fire_department` for Streak, `event_available` for Study days, `track_changes` for Daily goal, `eco` for Long-term memory, `restore` for Comeback), and stars. Locked: greyscale with a `lock` icon. Minimalist removes the gloss and shadow.
- Entries: deck list menu and a link in the Streaks statistics section.
- Summary: one "New: …" line. Deck list: a `SnackBar` when `NewAchievements` is non-empty after a sync or import.

## Risks / Trade-offs

- [Imports or the first sync change many earned dates at once] → one combined notice; first-run marks existing achievements seen.
- [Changing the goal or rest days changes earned achievements retroactively, possibly un-earning one] → accepted, consistent with "derived, not stored"; the seen state keeps a re-earned tier from being announced twice.
- [Badge drawing differs from the draft] → the draft PNG in `assets/badges/` is the reference for a golden test of the badge widget.
- [Memory for the history aggregates on the web] → measured in the extended benchmark.

## Migration Plan

New settings default to 0 rest days and an empty seen list; the first-run step fills the seen list. No data migration. Rollback leaves two unused settings fields.
