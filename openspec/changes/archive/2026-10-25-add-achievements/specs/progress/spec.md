# Spec Delta

## ADDED Requirements

### Requirement: Achievement catalogue
The app SHALL offer these achievements with these tiers: Answers (100, 1,000, 10,000, 100,000 answers), Streak (longest streak of 7, 30, 100, 365 days), Study days (30, 100, 365, 1,000), Daily goal (goal met on 7, 30, 100 days), Long-term memory (100, 1,000, 5,000 cards whose interval reached 21 days) and Comeback (once). No achievement SHALL depend on ratings, answer speed or cards added.

#### Scenario: Answers tier
- **WHEN** the user's history has 1,020 counted answers
- **THEN** the Answers achievement has its first two tiers earned and shows 1,020 / 10,000 toward the next

#### Scenario: Long-term memory counts each card once
- **WHEN** a card reached an interval of 25 days, lapsed, and later reached 30 days
- **THEN** it counts once toward Long-term memory

#### Scenario: Comeback
- **WHEN** the user studies today after 15 days without study
- **THEN** the Comeback achievement is earned today

### Requirement: Achievements derived from the review history
Achievements SHALL be computed from the review history, using the study day and counted answer rules of this capability. Each earned tier SHALL have the study day on which its threshold was crossed, including for imported or synced history.

#### Scenario: Imported history
- **WHEN** the user imports an Anki collection whose 1,000th answer was on 2024-03-12
- **THEN** the Answers 1,000 tier shows it was earned on 2024-03-12

#### Scenario: Undo
- **WHEN** the user's 100th answer earned the Answers 100 tier and the user undoes it
- **THEN** the tier is locked again with 99 / 100

### Requirement: Daily goal achievement uses the current goal
The Daily goal achievement SHALL count days that met the current goal, and SHALL make no progress while no goal is set.

#### Scenario: No goal
- **WHEN** no daily goal is set
- **THEN** the Daily goal achievement shows that a goal is needed and no progress

### Requirement: Achievements screen
The deck list menu and the statistics SHALL open an achievements screen showing every achievement as a badge with its highest earned tier and the date it was earned, and, for the next tier, the progress toward it. Achievements with no earned tier SHALL be shown locked with their progress.

#### Scenario: Locked with progress
- **WHEN** the user's longest streak is 212 days
- **THEN** the Streak badge shows the 100-day tier as earned with its date and 212 / 365 toward the next

#### Scenario: Nothing earned
- **WHEN** the user has never studied
- **THEN** every badge is shown locked with zero progress

### Requirement: New achievement notice
When tiers become earned that this device has not shown yet, the app SHALL tell the user once on this device: on the study summary when they were earned in the session, otherwise with a short message on the deck list. When more than three are new at once, one message SHALL name how many were earned from the history.

#### Scenario: Earned on another device
- **WHEN** the user earned Study days 100 on the phone and the laptop syncs it
- **THEN** the laptop's deck list shows a short message for it once

#### Scenario: Import earns many
- **WHEN** an import earns 14 tiers at once
- **THEN** one message says that 14 achievements were earned from the history

#### Scenario: Shown once
- **WHEN** a new achievement was shown on the summary and the user returns to the deck list
- **THEN** the deck list does not show it again

### Requirement: Achievements hidden with streaks
When *Show streaks and goals* is off, the achievements screen entries and the new achievement notices SHALL NOT be shown.

#### Scenario: Hidden
- **WHEN** *Show streaks and goals* is off and the user earns Answers 1,000
- **THEN** no notice is shown and the deck list menu has no achievements entry

## MODIFIED Requirements

### Requirement: Current and longest streak
The current streak SHALL be the number of consecutive study days ending today, or ending yesterday when today is not yet a study day; in that case the streak SHALL be shown as at risk, not broken. The longest streak SHALL be the longest run of consecutive study days in the history. With rest days per week set to N, a run SHALL continue over days without study as long as no 7-day window inside it has more than N such days; rest days SHALL NOT add to the streak's length.

#### Scenario: Today not studied yet
- **WHEN** the user studied on each of the last 12 days but not yet today
- **THEN** the current streak is 12 days and is at risk

#### Scenario: Gap of one day
- **WHEN** the user studied today and on 5 days in a row before yesterday, but not yesterday
- **THEN** the current streak is 1 day and the longest streak is at least 5 days

#### Scenario: Daylight saving change
- **WHEN** the user studies every day across a daylight saving time change
- **THEN** the streak continues without a gap or a doubled day

#### Scenario: One rest day
- **WHEN** rest days per week is 1 and the user studied on the last 10 days except one day 4 days ago
- **THEN** the current streak continues and is 9 days

#### Scenario: Too many rest days
- **WHEN** rest days per week is 1 and the user skipped two days within the same week
- **THEN** the streak restarts after the second skipped day

#### Scenario: Applied to history
- **WHEN** the user changes rest days per week from 0 to 1
- **THEN** the longest streak is recomputed over the whole history with one rest day allowed

### Requirement: Goals and streaks settings
Settings SHALL have a Goals and streaks section with *Show streaks and goals* (on by default), *Daily goal* (none, cards or minutes, with a number) and *Rest days per week* (0, 1 or 2; 0 by default). *Show streaks and goals* SHALL belong to the device; the daily goal and rest days SHALL belong to the collection.

#### Scenario: Set a goal
- **WHEN** the user sets the daily goal to 20 minutes
- **THEN** the deck list header and summary show progress toward 20 minutes

#### Scenario: Persisted in backups
- **WHEN** the user sets a goal of 30 cards, makes a backup and restores it on another installation
- **THEN** the restored collection has a daily goal of 30 cards

#### Scenario: Rest days persisted
- **WHEN** the user sets 2 rest days per week, makes a backup and restores it
- **THEN** the restored collection allows 2 rest days per week
