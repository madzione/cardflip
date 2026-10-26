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
When streaks and goals are shown, the deck list SHALL show a row above the decks with the current streak, the level with its progress toward the next level, and, when a goal is set, the goal progress. When the streak is at risk, the row SHALL say that studying today keeps it, without warning colours. Tapping the row SHALL open the collection statistics at the Streaks section; tapping the level SHALL open the achievements screen.

#### Scenario: Goal in progress
- **WHEN** the current streak is 12 days, the goal is 30 cards and 18 cards were answered today
- **THEN** the row shows "12 days" and "18 / 30 cards"

#### Scenario: No goal
- **WHEN** no daily goal is set and the current streak is 12 days
- **THEN** the row shows only the streak and the level

#### Scenario: At risk
- **WHEN** the streak is 12 days and today has no answers yet
- **THEN** the row says "Study today to keep your 12-day streak"

#### Scenario: No streak
- **WHEN** the user has never studied
- **THEN** the row shows a streak of 0 days without any message about losing it, and level 1

#### Scenario: Open statistics
- **WHEN** the user taps the row outside the level
- **THEN** the collection statistics open at the Streaks section

#### Scenario: Level
- **WHEN** the user has 3,760 XP
- **THEN** the row shows level 12 with its progress toward level 13 at 460 of 600 XP

#### Scenario: Open achievements
- **WHEN** the user taps the level
- **THEN** the achievements screen opens

### Requirement: Header stays current
The deck list header SHALL update after answers, undo, imports, restores, sync and a change of study day, without restarting the app.

#### Scenario: After a session
- **WHEN** the user answers 12 cards and returns to the deck list
- **THEN** the header shows the new goal progress and, if today was the first study of the day, the extended streak

#### Scenario: Day change
- **WHEN** the app stays open across the rollover hour
- **THEN** the header shows today's progress as 0 and the streak as at risk

### Requirement: Summary progress lines
When streaks and goals are shown, the session-complete and the nothing-due screens SHALL show the current streak, saying it was extended when this session made today a study day; when a goal is set, either that it was reached or how much is left; and the level, saying it was reached when this session raised it, otherwise how many XP are left to the next level. On the session-complete screen, a goal reached or a level raised in the session SHALL be said in the reward card instead of a line.

#### Scenario: Streak extended
- **WHEN** the user finishes the first session of the day and the streak becomes 13 days
- **THEN** the summary says "Streak extended to 13 days"

#### Scenario: Goal left
- **WHEN** the goal is 30 cards and 18 cards were answered today
- **THEN** the summary says "12 more cards to reach today's goal"

#### Scenario: Goal reached
- **WHEN** the goal is 30 cards and 34 cards were answered today, the goal having been reached in an earlier session
- **THEN** the summary says "Daily goal reached" as a line

#### Scenario: Goal reached in the session
- **WHEN** the session took today's answers from 25 to 34 with a goal of 30 cards
- **THEN** the reward card says "Daily goal reached" and no line repeats it

#### Scenario: Level reached
- **WHEN** the user was at level 12 when the session started and is at level 13 when it ends
- **THEN** the reward card on the session-complete screen says "Reached level 13"

#### Scenario: XP left
- **WHEN** the session did not raise the level and the user has 3,760 XP
- **THEN** the summary says "Level 12 · 140 XP to level 13"

### Requirement: Suggest a goal once
When no daily goal is set, the study summary SHALL suggest setting one, with a way to open the setting, until the user opens or dismisses the suggestion once.

#### Scenario: Dismissed
- **WHEN** no goal is set and the user dismisses the suggestion
- **THEN** later summaries do not show it

### Requirement: Goals and streaks settings
Settings SHALL have a Goals and streaks section with *Show streaks and goals* (on by default), *Show rewards while studying* (on by default), *Daily goal* (none, cards or minutes, with a number) and *Rest days per week* (0, 1 or 2; 0 by default). *Show streaks and goals* and *Show rewards while studying* SHALL belong to the device; the daily goal and rest days SHALL belong to the collection. *Show rewards while studying* SHALL be hidden while *Show streaks and goals* is off.

#### Scenario: Set a goal
- **WHEN** the user sets the daily goal to 20 minutes
- **THEN** the deck list header and summary show progress toward 20 minutes

#### Scenario: Persisted in backups
- **WHEN** the user sets a goal of 30 cards, makes a backup and restores it on another installation
- **THEN** the restored collection has a daily goal of 30 cards

#### Scenario: Rest days persisted
- **WHEN** the user sets 2 rest days per week, makes a backup and restores it
- **THEN** the restored collection allows 2 rest days per week

#### Scenario: Rewards while studying off
- **WHEN** the user turns off *Show rewards while studying* and an answer reaches the goal
- **THEN** no banner is shown, and the summary still shows the reward card

#### Scenario: Rewards setting stays on the device
- **WHEN** the user turns off *Show rewards while studying* on the laptop and syncs with the phone
- **THEN** the phone still shows reward banners

### Requirement: Turn streaks and goals off
When *Show streaks and goals* is off, the deck list header, the summary lines and the goal suggestion SHALL NOT be shown. The Streaks section of the statistics SHALL still be shown.

#### Scenario: Hidden
- **WHEN** the user turns off *Show streaks and goals*
- **THEN** the deck list shows no streak row and the study summary shows only cards studied and time spent

### Requirement: Quiet presentation
Streaks, goals, levels and achievements SHALL NOT interrupt studying with dialogs, SHALL NOT show loss or shame messages for a broken streak, and SHALL follow minimalist mode and reduce motion. The only thing shown while studying SHALL be the reward banner.

#### Scenario: Broken streak
- **WHEN** the user had a 40-day streak and did not study for three days
- **THEN** the header shows the new current streak without a message about the lost streak, and the statistics still show 40 days as the longest streak

#### Scenario: Minimalist
- **WHEN** minimalist mode is on
- **THEN** the header row and the reward card are shown without shadows or decorative fills

#### Scenario: No dialog while studying
- **WHEN** an answer earns an achievement tier
- **THEN** no dialog opens and the next card is shown at once

### Requirement: Achievement catalogue
The app SHALL offer these achievements with these tiers: Answers (100, 1,000, 10,000, 100,000 answers), Streak (longest streak of 7, 30, 100, 365 days), Study days (30, 100, 365, 1,000), Study time (10, 100, 500, 1,000 hours spent answering), Daily goal (goal met on 7, 30, 100 days), Long-term memory (100, 1,000, 5,000 cards whose interval reached 21 days), Level (level 5, 10, 25, 50) and Comeback (once). No achievement SHALL depend on ratings, answer speed or cards added.

#### Scenario: Answers tier
- **WHEN** the user's history has 1,020 counted answers
- **THEN** the Answers achievement has its first two tiers earned and shows 1,020 / 10,000 toward the next

#### Scenario: Long-term memory counts each card once
- **WHEN** a card reached an interval of 25 days, lapsed, and later reached 30 days
- **THEN** it counts once toward Long-term memory

#### Scenario: Comeback
- **WHEN** the user studies today after 15 days without study
- **THEN** the Comeback achievement is earned today

#### Scenario: Study time tier
- **WHEN** the user had spent 9 hours 50 minutes answering and studies for 15 minutes today
- **THEN** the Study time 10-hour tier is earned today and the badge shows 10 / 100 hours toward the next

#### Scenario: Level tier
- **WHEN** the user's XP reach 2,250 on 2025-06-03
- **THEN** the Level 10 tier is earned with the date 2025-06-03

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
When tiers become earned that this device has not shown yet, the app SHALL tell the user once on this device: with a reward banner when they were earned in a session and the banner is on; on the study summary when they were earned in the session; otherwise with a short message on the deck list. When more than three are new at once, one message SHALL name how many were earned from the history.

#### Scenario: Earned on another device
- **WHEN** the user earned Study days 100 on the phone and the laptop syncs it
- **THEN** the laptop's deck list shows a short message for it once

#### Scenario: Import earns many
- **WHEN** an import earns 14 tiers at once
- **THEN** one message says that 14 achievements were earned from the history

#### Scenario: Shown once
- **WHEN** a new achievement was shown on the summary and the user returns to the deck list
- **THEN** the deck list does not show it again

#### Scenario: Shown by the banner
- **WHEN** a banner showed "New: 1,000 answers" and the user leaves the session with Back
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
When the session-complete screen shows a reward card, it SHALL animate the card once, for at most three seconds, with its rings filling, its badges appearing and a short particle burst, without blocking any action. There SHALL be no animation when reduce motion is in effect or when *Show streaks and goals* is off; the card SHALL then show its final state.

#### Scenario: Goal reached
- **WHEN** the session reached the daily goal
- **THEN** the reward card fills the goal ring once while the summary's buttons stay usable

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and the session reached the daily goal
- **THEN** the reward card shows the full goal ring and "Daily goal reached" without an animation

#### Scenario: Goal already reached earlier
- **WHEN** the goal was reached in the morning and the user finishes another session in the evening
- **THEN** no reward card and no celebration are shown for the goal

#### Scenario: Level reached
- **WHEN** the session raised the level from 12 to 13
- **THEN** the reward card's level ring fills to the end of level 12 and then shows level 13, once

### Requirement: Experience points
Experience points (XP) SHALL be earned from the review history: 1 per counted answer, 10 per study day, 20 per day that met the current daily goal, 5 per card the first time its interval reached 21 days, and 25, 50, 100 or 200 per bronze, silver, gold or platinum achievement tier earned, except Level tiers. Ratings, answer speed and cards added SHALL NOT earn XP. Each XP SHALL belong to the study day it was earned on.

#### Scenario: One day
- **WHEN** the goal is 20 cards and the user's only study day has 30 answers, 2 of which first gave a card an interval of 21 days or more, and no achievement tier was earned
- **THEN** the user has 30 + 10 + 20 + 10 = 70 XP

#### Scenario: Achievement tier
- **WHEN** the user's 1,000th answer earns the silver Answers tier
- **THEN** 50 XP are added on that study day besides the XP for the answer

#### Scenario: Level tiers earn no XP
- **WHEN** the user reaches level 5 and earns the bronze Level tier
- **THEN** the user's XP do not change because of that tier

#### Scenario: Card counted once
- **WHEN** a card reached an interval of 25 days, lapsed, and later reached 30 days
- **THEN** it earned 5 XP once

### Requirement: XP derived from the review history
XP and the level SHALL be computed from the review history and the current settings, not stored. They SHALL include imported and synced history, follow undo, be the same on every synced device, be separate for each profile, and change retroactively when the daily goal or the rest days per week change.

#### Scenario: Undo the only answer of the day
- **WHEN** the user's only answer today earned 11 XP (1 for the answer, 10 for the study day) and the user undoes it
- **THEN** the user's XP are 11 lower than before the answer

#### Scenario: Imported history
- **WHEN** the user imports an Anki collection with history
- **THEN** XP and the level include that history

#### Scenario: Goal raised
- **WHEN** the user raises the daily goal so that fewer past days met it
- **THEN** the XP for goal days are recomputed with the new goal and the level may go down

### Requirement: Levels
The user SHALL start at level 1 with 0 XP. Going from level n to level n + 1 SHALL take 50 × n XP, so reaching level L takes 25 × L × (L − 1) XP in total. There SHALL be no top level.

#### Scenario: Exactly at a level
- **WHEN** the user has 2,250 XP
- **THEN** the user is at level 10 with 0 / 500 XP toward level 11

#### Scenario: Just below a level
- **WHEN** the user has 2,249 XP
- **THEN** the user is at level 9 with 449 / 450 XP toward level 10

#### Scenario: Never studied
- **WHEN** the user has never studied
- **THEN** the user is at level 1 with 0 / 50 XP toward level 2

### Requirement: Level card
When streaks and goals are shown, the achievements screen SHALL show, above the badges, a card with the level, the XP toward the next level, and how many XP came from answers, study days, goal days, long-term memory and achievements, with a short explanation of how XP are earned.

#### Scenario: Breakdown
- **WHEN** the user has 3,760 XP: 2,900 from answers, 300 from study days, 200 from goal days, 135 from long-term memory and 225 from achievements
- **THEN** the card shows level 12, 460 / 600 XP toward level 13 and each of those amounts

#### Scenario: No goal
- **WHEN** no daily goal is set
- **THEN** the card shows 0 XP from goal days and says that setting a goal earns more

### Requirement: Reward banner during study
When an answer during a session reaches the daily goal, raises the level or earns an achievement tier not yet shown on this device, the study screen SHALL show a banner for about 2.5 seconds naming what was earned. Events from one answer SHALL share one banner, and a later banner SHALL replace a visible one. Each goal, level or tier SHALL get a banner at most once per session. A Level tier SHALL be left out when the same banner names the level reached.

#### Scenario: Goal reached
- **WHEN** the goal is 30 cards, 29 cards were answered today and the user answers a card
- **THEN** a banner says "Daily goal reached" and disappears after about 2.5 seconds

#### Scenario: Level and tier from one answer
- **WHEN** an answer raises the level to 10, earns the Level 10 tier and earns Answers 1,000
- **THEN** one banner says "Level 10" and "New: 1,000 answers" and shows the Answers badge

#### Scenario: Undone and answered again
- **WHEN** an answer reached the goal and showed the banner, the user undoes it and answers the card again
- **THEN** no second banner is shown

#### Scenario: Nothing new
- **WHEN** the goal was already reached before the session and an answer earns nothing else
- **THEN** no banner is shown

### Requirement: Reward banner does not interrupt
The reward banner SHALL NOT take focus or react to keys, taps or clicks. It SHALL sit at the bottom edge of the card area, never covering the answer buttons, the reveal button or the remaining counts. Answering, revealing, undoing and every keyboard shortcut SHALL work while it is shown. Under reduce motion, it SHALL appear and disappear without sliding or scaling. Screen readers SHALL announce its text politely, without moving their focus.

#### Scenario: Rating key while shown
- **WHEN** a banner is shown and the user presses 3 after revealing the answer
- **THEN** the card is rated Good and the banner stays until its time is up

#### Scenario: Tap on the banner
- **WHEN** the user taps where the banner is shown
- **THEN** the tap reaches what is underneath the banner, and the banner stays

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and an answer reaches the goal
- **THEN** the banner appears and disappears without animation

### Requirement: Reward card on the summary
When the session reached the daily goal, raised the level or earned achievement tiers, the session-complete screen SHALL show a reward card above the progress lines. The card SHALL show the reached goal as a full ring, the new level in a ring filled from the start of the session's level, and the earned tiers as badges drawn as on the achievements screen, with more than three becoming "+N". Tapping the card SHALL open the achievements screen.

#### Scenario: Level and badge
- **WHEN** the session raised the level from 12 to 13 and earned Answers 1,000
- **THEN** the card shows level 13 in its ring, "Reached level 13", the Answers badge with its 1,000 tier and "New: 1,000 answers"

#### Scenario: Many tiers
- **WHEN** the session earned five tiers
- **THEN** the card shows three badges and "+2", and its text names all five

#### Scenario: Open achievements
- **WHEN** the user taps the reward card
- **THEN** the achievements screen opens

#### Scenario: Nothing earned
- **WHEN** the session did not reach the goal, raise the level or earn a tier
- **THEN** the summary shows no reward card

### Requirement: Rewards after leaving a session
When the user leaves a session before its summary, the deck list SHALL show one short message naming what the session earned and did not show: the goal reached, the level reached and new tiers, with an action to open the achievements screen. Nothing shown by a reward banner SHALL be named again.

#### Scenario: Banner off, left with Back
- **WHEN** *Show rewards while studying* is off and the user leaves a session with Back after it raised the level to 13 and reached the goal
- **THEN** the deck list shows one message saying "Reached level 13 · Daily goal reached"

#### Scenario: Banner shown, left with Back
- **WHEN** a banner showed "Level 13" and the user leaves with Back without earning anything else
- **THEN** the deck list shows no message
