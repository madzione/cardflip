# Spec Delta

## ADDED Requirements

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

## MODIFIED Requirements

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

### Requirement: Summary progress lines
When streaks and goals are shown, the session-complete and the nothing-due screens SHALL show the current streak, saying it was extended when this session made today a study day; when a goal is set, either that it was reached or how much is left; and the level, saying it was reached when this session raised it, otherwise how many XP are left to the next level.

#### Scenario: Streak extended
- **WHEN** the user finishes the first session of the day and the streak becomes 13 days
- **THEN** the summary says "Streak extended to 13 days"

#### Scenario: Goal left
- **WHEN** the goal is 30 cards and 18 cards were answered today
- **THEN** the summary says "12 more cards to reach today's goal"

#### Scenario: Goal reached
- **WHEN** the goal is 30 cards and 34 cards were answered today
- **THEN** the summary says "Daily goal reached"

#### Scenario: Level reached
- **WHEN** the user was at level 12 when the session started and is at level 13 when it ends
- **THEN** the session-complete screen says "Reached level 13"

#### Scenario: XP left
- **WHEN** the session did not raise the level and the user has 3,760 XP
- **THEN** the summary says "Level 12 · 140 XP to level 13"

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

### Requirement: Celebrations
When the daily goal was reached, an achievement tier was earned or the level was raised during a session, the session-complete screen SHALL play a short animation of at most two seconds that does not block any action. There SHALL be no animation during studying, when reduce motion is in effect, or when *Show streaks and goals* is off.

#### Scenario: Goal reached
- **WHEN** the session reached the daily goal
- **THEN** the summary plays a short celebration once while its buttons stay usable

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and the session reached the daily goal
- **THEN** the summary shows "Daily goal reached" without an animation

#### Scenario: Goal already reached earlier
- **WHEN** the goal was reached in the morning and the user finishes another session in the evening
- **THEN** no celebration is played

#### Scenario: Level reached
- **WHEN** the session raised the level from 12 to 13
- **THEN** the summary plays a short celebration once
