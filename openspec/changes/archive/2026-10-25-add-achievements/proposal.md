# Proposal

## Why

Streaks and the daily goal (`add-streaks-and-daily-goal`) reward today. Milestones give long-term users something to look back on: a thousandth answer, a year of study days, a thousand cards in long-term memory. A streak that breaks after one sick day also discourages more than it motivates, so streaks need a small allowance for rest days. This is phase 2 of `docs/proposals/gamification.md` and builds on phase 1.

## What Changes

- **Achievements**: a fixed catalogue, each with tiers, all derived from the review history:

  | Achievement | Tiers | Counts |
  |---|---|---|
  | Answers | 100 · 1,000 · 10,000 · 100,000 | answers |
  | Streak | 7 · 30 · 100 · 365 days | longest streak ever |
  | Study days | 30 · 100 · 365 · 1,000 | study days |
  | Daily goal | 7 · 30 · 100 days | days the goal was met |
  | Long-term memory | 100 · 1,000 · 5,000 cards | cards whose interval first reached 21 days |
  | Comeback | once | studying again after 14 or more days without study |

  Earned dates are historical, so importing an Anki collection earns achievements with their real dates. Nothing rewards ratings, speed, sessions without Again, or adding new cards.
- **Achievements screen**: opened from the deck list menu and from the statistics. A grid of badges; earned tiers show the date, locked ones show progress ("640 / 1,000"). Badges follow the draft designs in `assets/badges/` (Material icons and theme colours, bronze, silver, gold and platinum rims).
- **New achievement notice**: one line on the study summary ("New: 1,000 answers") and a snackbar on the deck list for achievements earned through sync or import. Many at once (after an import or first sync) become one notice ("14 achievements from your history"). No pop-ups during study.
- **Seen state**: each device shows a new achievement once. This is the only new stored data besides the setting below.
- **Rest days**: *Settings → Goals and streaks → Rest days per week* (0, 1 or 2; default 0). A streak survives up to that many days without study in any 7-day window. Synced, and applied to history too.
- Achievements are hidden together with streaks when *Show streaks and goals* is off.

## Capabilities

### New Capabilities

None. Achievements and rest days extend `progress`, which phase 1 introduces.

### Modified Capabilities

- `progress`: new requirements for achievements, tiers, the achievements screen, new-achievement notices and seen state; rest days change "Current and longest streak" and "Goals and streaks settings".
- `study-session`: new requirement for the new-achievement line on the summary.
- `device-sync`: "What is synced" includes the rest days setting and not the seen state.

Requires `add-streaks-and-daily-goal` to be archived first: the `progress` deltas modify its requirements.

## Impact

- **Go**: `golang/stats/achievements.go` (definitions table), rest days in `golang/stats/streak.go`, `ProgressService.Achievements()` and new-achievement fields on `Progress`, the review index extended with per-day aggregates and first-mature days, new settings `RestDaysPerWeek` (synced) and `SeenAchievements` (local).
- **Flutter**: achievements screen (`lib/features/progress/`), badge widget drawn from Material icons and theme colours, summary line, deck list snackbar, menu and statistics entries, rest days setting, strings.
- **Assets**: `assets/badges/` holds the draft designs (`.vectorcraft` source, SVG, PNG) as the visual reference; they are not bundled in the app.
- **Tests**: Go unit tests for each achievement and rest days, sync and import tests for earned dates and seen state, widget tests.
