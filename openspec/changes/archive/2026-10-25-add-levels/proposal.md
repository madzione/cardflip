# Proposal

## Why

Streaks reward today and achievements reward rare milestones, so for weeks at a time nothing visibly moves. A level that grows with every study session shows that progress continues between milestones, and it ties streaks, the daily goal, long-term memory and achievements into one number. Two new badges round this out: one for reaching levels, and one for the time spent studying, which the catalogue does not reward yet. This change builds on `add-streaks-and-daily-goal`, `add-achievements` and `add-deck-progress-and-recap`.

## What Changes

- **Experience points (XP)**, derived from the review history like streaks and achievements, so imports, sync and undo count:

  | Source | XP |
  |---|---|
  | Counted answer | 1 |
  | Study day | 10 |
  | Day the daily goal was met (current goal) | 20 |
  | Card reaching long-term memory (once per card) | 5 |
  | Achievement tier earned | bronze 25 · silver 50 · gold 100 · platinum 200 |

  Level achievement tiers earn no XP, so levels do not feed themselves. Ratings, speed and adding cards earn nothing, as with achievements.
- **Levels**: start at level 1. Going from level *n* to *n* + 1 takes 50 × *n* XP, so level 10 needs 2,250 XP and level 50 needs 61,250 XP. There is no top level.
- **Where the level shows**:
  - Deck list header: the level with a small progress ring beside the streak. Tapping it opens the achievements screen.
  - Achievements screen: a level card at the top with the XP toward the next level and where the XP came from.
  - Study summary: "Reached level 13" when the session raised the level, otherwise "Level 12 · 140 XP to level 13". A level-up plays the existing celebration.
  - MCP statistics tool: level and XP beside the streak.
- **New badges** in the achievement catalogue:

  | Achievement | Tiers | Counts |
  |---|---|---|
  | Level | 5 · 10 · 25 · 50 | level reached |
  | Study time | 10 · 100 · 500 · 1,000 hours | time spent answering |

  Both follow the existing badge style. Their earned dates come from history and are announced through the existing new-achievement notices.
- Levels and XP are hidden with streaks and achievements when *Show streaks and goals* is off. No new user settings. The only new stored data is a per-device list of evaluated achievements, so that upgrading users are not told at once about every tier of the two new badges their history already earned.
- Assumptions (see design): the XP amounts and the level curve are fixed in code. Changing the goal or rest days can move XP and the level retroactively, as it already does for achievements.

## Capabilities

### New Capabilities

None. XP and levels extend `progress`, which owns streaks, the goal and achievements.

### Modified Capabilities

- `progress`: new requirements for XP, XP derived from history, levels and the level card on the achievements screen; "Achievement catalogue" gains Level and Study time; "Deck list header", "Summary progress lines" and "Celebrations" include the level.
- `study-session`: "New achievements on the summary" leaves out a Level tier that the level line already names.
- `agent-access`: "Collection statistics tool" returns the level and XP.

## Impact

- **Go**: `golang/stats/achievements.go` (Study time and Level, two-pass evaluation), new `golang/stats/level.go` (XP per study day, level curve, breakdown), `ProgressService.Level()` and level fields on `Progress` (`golang/svc_progress.go`), local setting `SeenAchievementIDs` (`golang/core/model.go`, `settings_sync.go`), MCP statistics fields (`golang/mcp_tools_read.go`).
- **Flutter**: level ring on the deck list header (`lib/features/decks/progress_header.dart`), level card on the achievements screen, summary line and celebration trigger (`lib/features/study/study_screen.dart`), looks and strings for the two badges (`lib/features/progress/achievements.dart`), English and German strings.
- **Assets**: the two new badges added to the draft in `assets/badges/` as the visual reference.
- **Tests**: Go unit tests for XP, the curve and both achievements, service and MCP tests, widget and golden tests, integration flow.
