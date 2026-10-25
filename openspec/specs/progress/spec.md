# progress Specification

## Purpose

Motivation to study a little every day: study days, streaks and a daily goal, derived from the review history, shown quietly on the deck list and the study summary, and possible to turn off.

## Requirements

### Requirement: Study days
A study day SHALL be a day, by the rollover hour, with at least one answer in the review history. Answers in filtered decks that reschedule cards SHALL count. Previews in filtered decks and manual entries (set due date, reset, forget) SHALL NOT count. A day studied on several devices SHALL be one study day.

#### Scenario: Rollover hour
- **WHEN** the rollover hour is 4 and the user's only answer of the night is at 02:30
- **THEN** that answer belongs to the previous study day

#### Scenario: Preview does not count
- **WHEN** the user's only activity today is previewing cards in a filtered deck that does not reschedule
- **THEN** today is not a study day

#### Scenario: Set due date does not count
- **WHEN** the user's only activity today is setting the due date of 10 cards
- **THEN** today is not a study day

### Requirement: Streaks derived from the review history
Streaks, goal progress and study days SHALL be computed from the review history and not stored. They SHALL therefore include history from imports and backups, follow undo, reflect synced reviews, and be separate for each profile.

#### Scenario: Imported Anki history
- **WHEN** the user imports an Anki collection with reviews on each of the last 50 days
- **THEN** the current streak is 50 days

#### Scenario: Undo the only answer of the day
- **WHEN** the user answers one card today, which extended the streak to 13, and then undoes it
- **THEN** today is no longer a study day and the streak shows 12 days, at risk

#### Scenario: Two devices
- **WHEN** the user studied on the phone yesterday and on the laptop today, and the devices sync
- **THEN** both devices show a streak that includes yesterday and today

### Requirement: Current and longest streak
The current streak SHALL be the number of consecutive study days ending today, or ending yesterday when today is not yet a study day; in that case the streak SHALL be shown as at risk, not broken. The longest streak SHALL be the longest run of consecutive study days in the history.

#### Scenario: Today not studied yet
- **WHEN** the user studied on each of the last 12 days but not yet today
- **THEN** the current streak is 12 days and is at risk

#### Scenario: Gap of one day
- **WHEN** the user studied today and on 5 days in a row before yesterday, but not yesterday
- **THEN** the current streak is 1 day and the longest streak is at least 5 days

#### Scenario: Daylight saving change
- **WHEN** the user studies every day across a daylight saving time change
- **THEN** the streak continues without a gap or a doubled day

### Requirement: Daily goal
The daily goal SHALL be none (the default), a number of cards, or a number of minutes. Progress SHALL be today's number of counted answers or today's study time. The goal SHALL be reached when progress meets it. The streak SHALL NOT depend on the goal.

#### Scenario: Card goal
- **WHEN** the goal is 30 cards and the user has answered 18 cards today
- **THEN** goal progress is 18 of 30 and the goal is not reached

#### Scenario: Minute goal
- **WHEN** the goal is 15 minutes and the user has studied for 16 minutes today
- **THEN** the goal is reached

#### Scenario: Streak without the goal
- **WHEN** the goal is 30 cards and the user answers 5 cards today
- **THEN** today is a study day and the streak includes it

### Requirement: Days the goal was met
Whether a past day met the goal SHALL be judged against the current goal.

#### Scenario: Goal raised
- **WHEN** the user answered 30 cards on each of the last 10 days and changes the goal from 20 to 50 cards
- **THEN** the number of days the goal was met in that period is 0

### Requirement: Deck list header
When streaks and goals are shown, the deck list SHALL show a row above the decks with the current streak and, when a goal is set, the goal progress. When the streak is at risk, the row SHALL say that studying today keeps it, without warning colours. Tapping the row SHALL open the collection statistics at the Streaks section.

#### Scenario: Goal in progress
- **WHEN** the current streak is 12 days, the goal is 30 cards and 18 cards were answered today
- **THEN** the row shows "12 days" and "18 / 30 cards"

#### Scenario: No goal
- **WHEN** no daily goal is set and the current streak is 12 days
- **THEN** the row shows only the streak

#### Scenario: At risk
- **WHEN** the streak is 12 days and today has no answers yet
- **THEN** the row says "Study today to keep your 12-day streak"

#### Scenario: No streak
- **WHEN** the user has never studied
- **THEN** the row shows a streak of 0 days without any message about losing it

#### Scenario: Open statistics
- **WHEN** the user taps the row
- **THEN** the collection statistics open at the Streaks section

### Requirement: Header stays current
The deck list header SHALL update after answers, undo, imports, restores, sync and a change of study day, without restarting the app.

#### Scenario: After a session
- **WHEN** the user answers 12 cards and returns to the deck list
- **THEN** the header shows the new goal progress and, if today was the first study of the day, the extended streak

#### Scenario: Day change
- **WHEN** the app stays open across the rollover hour
- **THEN** the header shows today's progress as 0 and the streak as at risk

### Requirement: Summary progress lines
When streaks and goals are shown, the session-complete and the nothing-due screens SHALL show the current streak, saying it was extended when this session made today a study day, and, when a goal is set, either that it was reached or how much is left.

#### Scenario: Streak extended
- **WHEN** the user finishes the first session of the day and the streak becomes 13 days
- **THEN** the summary says "Streak extended to 13 days"

#### Scenario: Goal left
- **WHEN** the goal is 30 cards and 18 cards were answered today
- **THEN** the summary says "12 more cards to reach today's goal"

#### Scenario: Goal reached
- **WHEN** the goal is 30 cards and 34 cards were answered today
- **THEN** the summary says "Daily goal reached"

### Requirement: Suggest a goal once
When no daily goal is set, the study summary SHALL suggest setting one, with a way to open the setting, until the user opens or dismisses the suggestion once.

#### Scenario: Dismissed
- **WHEN** no goal is set and the user dismisses the suggestion
- **THEN** later summaries do not show it

### Requirement: Goals and streaks settings
Settings SHALL have a Goals and streaks section with *Show streaks and goals* (on by default) and *Daily goal* (none, cards or minutes, with a number). *Show streaks and goals* SHALL belong to the device; the daily goal SHALL belong to the collection.

#### Scenario: Set a goal
- **WHEN** the user sets the daily goal to 20 minutes
- **THEN** the deck list header and summary show progress toward 20 minutes

#### Scenario: Persisted in backups
- **WHEN** the user sets a goal of 30 cards, makes a backup and restores it on another installation
- **THEN** the restored collection has a daily goal of 30 cards

### Requirement: Turn streaks and goals off
When *Show streaks and goals* is off, the deck list header, the summary lines and the goal suggestion SHALL NOT be shown. The Streaks section of the statistics SHALL still be shown.

#### Scenario: Hidden
- **WHEN** the user turns off *Show streaks and goals*
- **THEN** the deck list shows no streak row and the study summary shows only cards studied and time spent

### Requirement: Quiet presentation
Streaks and goals SHALL NOT interrupt studying with dialogs or pop-ups, SHALL NOT show loss or shame messages for a broken streak, and SHALL follow minimalist mode and reduce motion.

#### Scenario: Broken streak
- **WHEN** the user had a 40-day streak and did not study for three days
- **THEN** the header shows the new current streak without a message about the lost streak, and the statistics still show 40 days as the longest streak

#### Scenario: Minimalist
- **WHEN** minimalist mode is on
- **THEN** the header row is shown without shadows or decorative fills
