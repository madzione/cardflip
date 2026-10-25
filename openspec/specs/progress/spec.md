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

### Requirement: Weekly recap
On the first time the deck list is shown in a new week, it SHALL show a card at the top with the previous week's study days, answers, study time and true retention, each compared to the week before. Weeks SHALL start on the first day of the week of the interface language and follow the study day boundary.

#### Scenario: New week
- **WHEN** it is Monday, last week the user studied on 6 days with 420 answers in 95 minutes and 91 % retention, and the week before on 4 days with 300 answers
- **THEN** the deck list shows a recap with 6 days, 420 answers, 95 minutes and 91 %, each with its change from the week before

#### Scenario: Open statistics
- **WHEN** the user taps the recap
- **THEN** the collection statistics open

#### Scenario: Nothing studied last week
- **WHEN** the user did not study in the previous week
- **THEN** no recap is shown

### Requirement: Dismiss the recap
The user SHALL be able to dismiss the recap. A dismissed recap SHALL stay hidden on this device until the next week. When *Show streaks and goals* is off, no recap SHALL be shown.

#### Scenario: Dismissed
- **WHEN** the user dismisses this week's recap and restarts the app
- **THEN** the deck list shows no recap until next week

### Requirement: Celebrations
When the daily goal was reached or an achievement tier was earned during a session, the session-complete screen SHALL play a short animation of at most two seconds that does not block any action. There SHALL be no animation during studying, when reduce motion is in effect, or when *Show streaks and goals* is off.

#### Scenario: Goal reached
- **WHEN** the session reached the daily goal
- **THEN** the summary plays a short celebration once while its buttons stay usable

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and the session reached the daily goal
- **THEN** the summary shows "Daily goal reached" without an animation

#### Scenario: Goal already reached earlier
- **WHEN** the goal was reached in the morning and the user finishes another session in the evening
- **THEN** no celebration is played
