# Design

## Context

- `ProgressService.refresh()` (`golang/svc_progress.go`) keeps the per-day aggregates of the whole history (`days`, `answers`, `ms`, `mature`), extends them after each answer and passes them as `stats.History` to `stats.EvalAchievements` (`golang/stats/achievements.go`). Every achievement is an entry in the `Achievements` table with an `Eval` that returns its progress and the study day each tier was reached.
- Every source of XP lands on a study day: answers and goal days by definition, first-mature days because maturity comes from an answer, and every tier's earned day is a study day (each `Eval` walks `Days` or `Mature`). So XP can be one integer per entry of `History.Days`.
- The Flutter side maps achievement ids to looks, names and captions in `lib/features/progress/achievements.dart`; the summary lines and the celebration are in `_ProgressLines` in `lib/features/study/study_screen.dart`, which records `_studiedBefore` and `_goalReachedBefore` at session start.
- Motivation and the XP table are in `proposal.md`; the rules are in the `progress` spec delta.

## Goals / Non-Goals

**Goals:**
- XP and the level from the data `ProgressService` already holds, at no extra history pass and no new stored state.
- The two new badges as plain entries of the existing table, with no change to the badge widget beyond the two looks.

**Non-Goals:**
- Level titles or ranks ("Scholar"), configurable XP amounts or curves, XP multipliers or boosts.
- A level per deck or in the statistics screen.
- A separate notice for level-ups through sync or import: the header shows the new level, and the Level tiers use the existing new-achievement notice.

## Decisions

### 1. Two-pass evaluation

Tier bonuses make XP depend on the other achievements, and the Level achievement depends on XP. `EvalAchievements` therefore:

1. Evaluates every achievement except Level, keeping each tier's earned day as a day number (not yet a date string).
2. Builds `History.XP`, one value per entry of `Days`: `answers[i] + 10`, `+ 20` when `Goal.Met(answers[i], ms[i])`, `+ 5` per entry of `Mature` on that day, and the bonus of each tier earned that day (by tier name: 25, 50, 100, 200; Comeback's single tier counts as gold).
3. Evaluates Level, which reads `History.XP`.

Achievements gain a `FromXP bool` field to mark the second pass; Level is the only one. The constants live in `golang/stats/level.go` next to the curve. Matching mature and tier days to indexes in `Days` uses binary search; both lists are short compared to the answers.

*Alternative*: giving XP for Level tiers with a fixed point iteration. Rejected: it gains nothing for the user and makes the curve harder to reason about.

### 2. The curve

`LevelOf(xp) (level, start, next int)` with `start = 25·L·(L−1)` and `next = 25·(L+1)·L`. L is found from the quadratic (`floor((1 + sqrt(1 + 4·xp/25)) / 2)`) and corrected by one step either way to avoid float rounding at exact thresholds; a table test covers 0, 49, 50, 2,249, 2,250 and a large value.

With ~150 XP a day (100 answers, a study day, the goal and a few mature cards) level 10 takes about two weeks, 25 about three months and 50 a bit over a year. Heavy users get there sooner, which is fine because nothing is capped or ranked.

### 3. The new achievements

- `study-time`: tiers `10, 100, 500, 1000` hours; `Eval` adds up `Ms` and compares against `tier · 3,600,000`; progress is whole hours. Go's `int` is 64-bit on all targets, WebAssembly included, so the sum does not overflow.
- `level`: tiers `5, 10, 25, 50`; `Eval` turns the tiers into XP thresholds (`25·L·(L−1)`), reuses `cumulative` over `History.XP`, and reports the current level as progress.
- Table order (screen order): answers, streak, study-days, study-time, daily-goal, long-term, level, comeback.

Seen state needs no change: new keys (`study-time:10`, `level:5`) are simply unseen. Users upgrading with history would be greeted by every tier of the two new achievements at once. To avoid that, the first evaluation after the upgrade marks the tiers of achievements unknown to the seen list as seen, the same way the first evaluation does today when `SeenAchievements` is nil. This uses `SeenAchievementIDs` in `core.Settings`, listed in `localSettings`: the achievement ids this device has evaluated. An id missing from it has its earned tiers marked seen silently, then the id is added. Without that list, each new achievement in a later version would need its own migration.

### 4. Service surface

- `Progress` gains `Level`, `XP`, `LevelStart` and `NextLevelXP` (all from the same refresh, no extra cost).
- `ProgressService.Level()` returns `stats.LevelStatus` with those fields plus the breakdown (`FromAnswers`, `FromStudyDays`, `FromGoalDays`, `FromLongTerm`, `FromAchievements`), kept from the last refresh like `achievements`.
- The MCP statistics tool gets a `level` object (`level`, `xp`, `nextLevelXp`) next to `streak`.

### 5. Flutter

- **Header** (`progress_header.dart`): a compact level indicator at the end of the row: a 24 px `CircularProgressIndicator`-style ring (theme primary on the surface variant track, no fill in minimalist mode) around the level number, with its own `InkWell` that opens the achievements screen. Semantics: "Level 12, 460 of 600 XP to level 13". The rest of the row keeps opening the statistics.
- **Level card** (new `lib/features/progress/level_card.dart`), above the grid on the achievements screen: the level, a linear bar, "460 / 600 XP to level 13", the five sources as a small table, and one line explaining what earns XP. It loads `progress.level()` together with `progress.achievements()`.
- **Summary**: `_StudyScreenState` also records `_levelBefore`. `_ProgressLines` adds the level line after the goal line: `l.levelReached(level)` when `levelBefore < level`, else `l.levelXpLeft(level, next − xp, level + 1)`. The new-achievement line leaves out `level:*` tiers when the level line says the level was reached (they are still marked seen). The celebration triggers also on a level-up, with `Icons.trending_up` when no goal or tier celebration takes precedence; the existing order stays goal, then achievements.
- **Looks** in `achievements.dart`: `study-time` with `Icons.hourglass_bottom` on an indigo disc (`0xFF6F7FD6` → `0xFF3949AB`), `level` with `Icons.trending_up` on a pink disc (`0xFFF06A9B` → `0xFFC2185B`), distinct from the six existing disc colours. Names, tier captions ("100 hours", "Level 25") and notice labels ("100 hours of study", "Level 25") in English and German.
- **Draft**: the two badges are added to `assets/badges/achievement-badges.vectorcraft` and re-exported as SVG and PNG, so the golden test keeps a reference.

## Risks / Trade-offs

- [Changing the goal or rest days moves XP and can lower the level] → accepted and specified ("XP derived from the review history"), consistent with achievements; the header simply shows the new value with no loss message.
- [Upgrading users meet two new achievements with history behind them] → silenced by the evaluated-ids rule in decision 3.
- [Level-ups are frequent early on and could feel noisy] → only a summary line and the existing short celebration; no deck list notice for level-ups.
- [Tier bonuses tie XP to the achievement catalogue: adding an achievement later raises everyone's XP] → accepted; the level is a personal measure, not a ranking, and the change log will say so.
- [Refresh cost] → the XP pass is linear in the study days and runs only when achievements are evaluated; the existing progress benchmark is extended to cover it.

## Migration Plan

No data migration. On first start the new `SeenAchievementIDs` local setting is empty: for existing users the six known ids are added straight away (their tiers are already in the seen list) and the earned tiers of `study-time` and `level` are marked seen. New users get the existing nil rule. Rollback leaves one unused local setting and two unknown keys in the seen list, which the older version ignores.
