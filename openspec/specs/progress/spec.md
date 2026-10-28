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
When streaks and goals are shown, the session-complete and the nothing-due screens SHALL show the current streak, saying it was extended when this session made today a study day; when a goal is set, either that it was reached or how much is left; and the level, saying it was reached when this session raised it, otherwise how many XP are left to the next level. On the session-complete screen, a goal reached or a level raised in the session SHALL be said in the reward card instead of a line, and the XP the session earned SHALL be shown when they are more than 0.

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

#### Scenario: XP earned
- **WHEN** the user had 3,712 XP when the session started and has 3,760 XP when it ends
- **THEN** the session-complete screen says "+48 XP this session"

#### Scenario: All answers undone
- **WHEN** the user answered one card and undid it, and the session ends
- **THEN** the session-complete screen shows no XP earned line

### Requirement: Suggest a goal once
When no daily goal is set, the study summary SHALL suggest setting one, with a way to open the setting, until the user opens or dismisses the suggestion once.

#### Scenario: Dismissed
- **WHEN** no goal is set and the user dismisses the suggestion
- **THEN** later summaries do not show it

### Requirement: Goals and streaks settings
Settings SHALL have a Goals and streaks section with *Show streaks and goals* (on by default), *Show rewards while studying* (on by default), *Celebration sounds and vibration* (on by default), *Show level-ups while studying* (on by default), *Daily goal* (none, cards or minutes, with a number) and *Rest days per week* (0, 1 or 2; 0 by default). *Show streaks and goals*, *Show rewards while studying*, *Celebration sounds and vibration* and *Show level-ups while studying* SHALL belong to the device; the daily goal and rest days SHALL belong to the collection. *Show rewards while studying*, *Celebration sounds and vibration* and *Show level-ups while studying* SHALL be hidden while *Show streaks and goals* is off.

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
- **THEN** no banner is shown, and the session still ends with the celebration sequence and the reward card

#### Scenario: Rewards setting stays on the device
- **WHEN** the user turns off *Show rewards while studying* on the laptop and syncs with the phone
- **THEN** the phone still shows reward banners

#### Scenario: Sounds setting stays on the device
- **WHEN** the user turns off *Celebration sounds and vibration* on the laptop and syncs with the phone
- **THEN** the phone still plays celebration sounds

#### Scenario: Level-ups setting stays on the device
- **WHEN** the user turns off *Show level-ups while studying* on the laptop and syncs with the phone
- **THEN** the phone still shows the level page during a session

### Requirement: Turn streaks and goals off
When *Show streaks and goals* is off, the deck list header, the summary lines and the goal suggestion SHALL NOT be shown. The Streaks section of the statistics SHALL still be shown.

#### Scenario: Hidden
- **WHEN** the user turns off *Show streaks and goals*
- **THEN** the deck list shows no streak row and the study summary shows only cards studied and time spent

### Requirement: Quiet presentation
Streaks, goals, levels and achievements SHALL NOT interrupt studying with dialogs, SHALL NOT show loss or shame messages for a broken streak, and SHALL follow minimalist mode and reduce motion. While studying, only the reward banner SHALL be shown, and the level page while *Show level-ups while studying* is on. Other full-screen celebrations SHALL be shown only when the session ends or is left with Back, as the celebration sequence.

#### Scenario: Broken streak
- **WHEN** the user had a 40-day streak and did not study for three days
- **THEN** the header shows the new current streak without a message about the lost streak, and the statistics still show 40 days as the longest streak

#### Scenario: Minimalist
- **WHEN** minimalist mode is on
- **THEN** the header row, the reward card and the celebration pages are shown without shadows, glows, rays or decorative fills

#### Scenario: No dialog while studying
- **WHEN** an answer earns an achievement tier
- **THEN** no dialog or celebration page opens and the next card is shown at once

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
Each celebration page SHALL animate its centrepiece once, within three seconds, with confetti on the level and achievement pages and a smaller burst on the streak and goal pages, and SHALL show its final state afterwards. The animation SHALL NOT block Continue once the input guard has passed. Under reduce motion, the pages SHALL be shown in their final state without counting, filling, flipping or confetti.

#### Scenario: Goal reached
- **WHEN** the session reached the daily goal
- **THEN** the goal page fills its ring once, turns it into a check with a small burst, and Continue stays usable after the input guard

#### Scenario: Level reached
- **WHEN** the session raised the level from 12 to 13
- **THEN** the level page fills the ring to the end of level 12, then shows 13 with confetti, once

#### Scenario: Continue during the animation
- **WHEN** the achievement page's confetti is still falling one second after it appeared and the user taps Continue
- **THEN** the next page is shown at once

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and the session reached the daily goal
- **THEN** the goal page shows the full ring with its check and "Daily goal reached" without an animation

#### Scenario: Goal already reached earlier
- **WHEN** the goal was reached in the morning and the user finishes another session in the evening
- **THEN** no goal page and no reward card entry are shown for the goal

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
When an answer during a session reaches the daily goal, raises the level while *Show level-ups while studying* is off, or earns an achievement tier not yet shown on this device, the study screen SHALL show a banner for about 2.5 seconds naming what was earned, its badge or icon popping in with a small confetti burst over the card area. Events from one answer SHALL share one banner, and a later banner SHALL replace a visible one. Each goal, level or tier SHALL get a banner at most once per session. A Level tier SHALL be left out when the same banner or the level page names the level reached.

#### Scenario: Goal reached
- **WHEN** the goal is 30 cards, 29 cards were answered today and the user answers a card
- **THEN** a banner says "Daily goal reached" with a confetti burst and disappears after about 2.5 seconds

#### Scenario: Level and tier from one answer
- **WHEN** *Show level-ups while studying* is off and an answer raises the level to 10, earns the Level 10 tier and earns Answers 1,000
- **THEN** one banner says "Level 10" and "New: 1,000 answers" and shows the Answers badge

#### Scenario: Undone and answered again
- **WHEN** an answer reached the goal and showed the banner, the user undoes it and answers the card again
- **THEN** no second banner is shown

#### Scenario: Nothing new
- **WHEN** the goal was already reached before the session and an answer earns nothing else
- **THEN** no banner is shown

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and an answer reaches the goal
- **THEN** the banner is shown without a pop-in or confetti

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
When the session reached the daily goal, raised the level or earned achievement tiers, the session-complete screen SHALL show a reward card above the progress lines, without an animation of its own. The card SHALL show the reached goal as a full ring, the new level in its ring, and the earned tiers as badges drawn as on the achievements screen, with more than three becoming "+N". Tapping the card SHALL open the achievements screen.

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

#### Scenario: After the sequence
- **WHEN** the celebration sequence ends and the summary appears
- **THEN** the reward card is shown in its final state without filling rings or a particle burst

### Requirement: Rewards after leaving a session
When a session is closed before its summary without the celebration pages, for example by switching profiles, the deck list SHALL show one short message naming what the session earned and did not show: the goal reached, the level reached and new tiers, with an action to open the achievements screen. Nothing shown by a reward banner or a celebration page SHALL be named again. Leaving with Back SHALL show the celebration pages instead.

#### Scenario: Banner off, left with Back
- **WHEN** *Show rewards while studying* and *Show level-ups while studying* are off and the user leaves a session with Back after it raised the level to 13 and reached the goal
- **THEN** the goal and level pages are shown, then the deck list without a message

#### Scenario: Banner shown, left with Back
- **WHEN** a banner showed "Level 13" and the user leaves with Back without earning anything else
- **THEN** the level page is shown, then the deck list without a message

#### Scenario: Closed without Back
- **WHEN** a session raised the level to 13 and reached the goal, no banner or page showed them, and the profile is switched before the summary
- **THEN** the deck list shows one message saying "Reached level 13 · Daily goal reached"


### Requirement: Celebration sequence
When a session ends on the session-complete screen and it extended the streak, reached the daily goal, raised the level or earned achievement tiers, the study screen SHALL first show one full-screen celebration page per event, in the order streak, daily goal, level, achievements, and then the summary. Nothing SHALL be celebrated twice in one session. No sequence SHALL be shown when *Show streaks and goals* is off.

#### Scenario: Several events
- **WHEN** the first session of the day extends the streak to 13 days, reaches the goal and raises the level to 13
- **THEN** a streak page, a goal page and a level page are shown in that order, followed by the summary

#### Scenario: Nothing earned
- **WHEN** the session studied cards but did not extend the streak, reach the goal, raise the level or earn a tier
- **THEN** the summary is shown at once, without celebration pages

#### Scenario: Banner already shown
- **WHEN** a reward banner said "Level 13" during the session
- **THEN** the session still ends with the level page

#### Scenario: Continued after the summary
- **WHEN** the sequence and summary were shown, a learning card became due, the user continued and the session ends again having earned nothing new
- **THEN** the summary is shown without celebration pages

#### Scenario: Streaks and goals off
- **WHEN** *Show streaks and goals* is off and the session raised the level
- **THEN** the summary is shown without celebration pages

### Requirement: Celebrations when leaving with Back
When the user leaves a session with Back before it ends, the study screen SHALL first show the celebration pages for what the session earned and no page celebrated yet, in the same order as at the end of a session, and then the deck list. Continue on the last page, Escape and Back SHALL go to the deck list. A session that earned nothing SHALL be left at once.

#### Scenario: Goal reached, left with Back
- **WHEN** a session reached the goal and the user leaves it with Back
- **THEN** the goal page is shown, and Continue shows the deck list without a message about the goal

#### Scenario: Skip after Back
- **WHEN** the user left a session with Back, the goal page is the first of two pages and the user presses Escape
- **THEN** the deck list is shown

#### Scenario: Nothing earned
- **WHEN** a session earned nothing and the user leaves it with Back
- **THEN** the deck list is shown at once

### Requirement: Streak celebration
The streak page SHALL show a large flame with the current streak counting up from the day before, "day streak", and the days of the current week, starting on the first day of the week of the interface language: studied days ticked, rest days the streak skipped marked as rest, other past days empty, today's tick appearing last, and later days blank. When the current streak equals the longest streak and is above 1, it SHALL say "New personal best".

#### Scenario: Extended
- **WHEN** the user studied on each of the last 12 days and today's session is the first of the day
- **THEN** the streak page counts from 12 to 13 and says "13 day streak"

#### Scenario: Week with a rest day
- **WHEN** the week starts on Monday, it is Thursday, rest days per week is 1 and the user studied on Monday and Wednesday but not on Tuesday
- **THEN** the strip shows Monday and Wednesday ticked, Tuesday as a rest day, Thursday ticked last, and Friday to Sunday blank

#### Scenario: Personal best
- **WHEN** the longest streak was 40 days and the session made the current streak 41
- **THEN** the streak page says "New personal best"

#### Scenario: First study day
- **WHEN** the user studies for the first time
- **THEN** the streak page says "1 day streak" without "New personal best"

### Requirement: Daily goal celebration
The daily goal page SHALL show the goal ring filling up and turning into a check, "Daily goal reached", today's progress toward the goal, and the XP the goal earned.

#### Scenario: Card goal
- **WHEN** the session took today's answers from 25 to 34 with a goal of 30 cards
- **THEN** the goal page says "Daily goal reached", "34 / 30 cards" and "+20 XP"

### Requirement: Level celebration
The level page SHALL show a large level ring filling from where the session started to the end of that level, then the new level number with a burst, "Level N", and how many XP are left to the next level. When the session raised the level by more than one, the ring SHALL fill once and then show the highest level reached.

#### Scenario: One level
- **WHEN** the session raised the level from 12 to 13 and the user has 3,910 XP
- **THEN** the level page fills the ring, shows 13, says "Level 13" and "640 XP to level 14"

#### Scenario: Two levels
- **WHEN** the session raised the level from 3 to 5
- **THEN** the level page fills the ring once and shows 5

### Requirement: Achievement celebration
Each achievement page SHALL show the tier's badge large, as drawn on the achievements screen, with its tier stars appearing one by one, "New achievement", the tier's name and rank, and the XP it earned. The first three tiers the session earned SHALL each have a page; further tiers SHALL share one last page that shows their badges and names them. A Level tier SHALL be left out when a level page is shown.

#### Scenario: One tier
- **WHEN** the session earned the silver Answers tier for 1,000 answers
- **THEN** an achievement page shows the Answers badge with its silver tier, "New achievement", "1,000 answers · Silver" and "+50 XP"

#### Scenario: Five tiers
- **WHEN** the session earned five tiers
- **THEN** three achievement pages show one tier each and a fourth page shows the other two badges and names them

#### Scenario: Level tier with level page
- **WHEN** the session raised the level to 10 and earned the Level 10 tier
- **THEN** a level page is shown and no achievement page shows the Level tier

### Requirement: Moving through celebrations
Each celebration page SHALL have a Continue button that shows the next page or the summary. Enter and Space SHALL do the same, and Escape or the system Back SHALL skip the remaining pages to the summary. For the first half second of each page, taps on Continue and these keys SHALL be ignored. Rating keys SHALL do nothing. Screen readers SHALL read each page's headline and text when it appears. Every celebration in the sequence SHALL count as shown.

#### Scenario: Continue
- **WHEN** the goal page is shown and the user taps Continue after two seconds
- **THEN** the next page is shown

#### Scenario: Space right after the last card
- **WHEN** the user presses Space 200 ms after the streak page appeared
- **THEN** the streak page stays

#### Scenario: Skip
- **WHEN** the streak page is the first of three pages and the user presses Escape
- **THEN** the summary is shown

#### Scenario: Rating key
- **WHEN** a celebration page is shown and the user presses 3
- **THEN** nothing happens

### Requirement: Celebration sounds and vibration
When *Celebration sounds and vibration* is on, each celebration page SHALL play a short sound of its kind (streak, goal, level, achievement) when it appears, and each reward banner a short sound as described in "Reward banner sound". On Android and iOS, a page SHALL also give a medium vibration and a banner a light one. The sounds SHALL NOT stop or pause card audio, and SHALL play under reduce motion too. When the setting is off, nothing SHALL be played or vibrated.

#### Scenario: Level page
- **WHEN** the setting is on and the level page appears on Android
- **THEN** the level sound plays and the phone vibrates once

#### Scenario: Banner
- **WHEN** the setting is on and an answer reaches the daily goal during study
- **THEN** the banner appears with the chime


#### Scenario: Off
- **WHEN** the setting is off and the session ends with three celebration pages
- **THEN** the pages are shown without any sound or vibration

#### Scenario: Desktop
- **WHEN** the setting is on and the level page appears on Linux
- **THEN** the level sound plays and nothing vibrates

### Requirement: Reward banner sound
When *Celebration sounds and vibration* is on, a reward banner SHALL play the sound of its biggest event: the achievement sound when it names a tier, else the level sound when it names a level, else a soft chime.

#### Scenario: Goal only
- **WHEN** an answer reaches the daily goal during study
- **THEN** the banner appears with the chime

#### Scenario: Goal and tier
- **WHEN** an answer reaches the goal and earns Answers 1,000
- **THEN** the banner appears with the achievement sound

#### Scenario: Level in a banner
- **WHEN** *Show level-ups while studying* is off and an answer raises the level
- **THEN** the banner appears with the level sound

### Requirement: Level-up while studying
While *Show level-ups while studying* is on, an answer during a session that raises the level SHALL show the level page at once, before the next card. Continue, Enter, Space, Escape and the system Back SHALL return to studying, after the same half-second guard as the other pages. A level shown this way SHALL NOT get a page again in the session. While the setting is off, a level-up SHALL get a banner and its page SHALL wait for the end of the session or for Back.

#### Scenario: Level-up mid-session
- **WHEN** the setting is on and an answer raises the level to 13 with cards still to study
- **THEN** the level page says "Level 13", and after Continue the next card is shown

#### Scenario: Escape returns to studying
- **WHEN** the level page is shown during a session and the user presses Escape after a second
- **THEN** the next card is shown and the session goes on

#### Scenario: Not again at the end
- **WHEN** the level page was shown during the session and the session ends without another level-up
- **THEN** the celebration sequence has no level page

#### Scenario: Setting off
- **WHEN** the setting is off and an answer raises the level to 13
- **THEN** a banner says "Level 13", and the level page is shown when the session ends

### Requirement: Studying waits for the level page
While the level page is shown during a session, the next card's answer timer, auto advance and audio SHALL wait until the page closes, and undo and the rating keys SHALL do nothing. Goal and tier events of the answer that raised the level SHALL get their banner when the page closes.

#### Scenario: Timer starts after the page
- **WHEN** the level page is shown for 5 seconds during a session and the user taps Continue
- **THEN** the next card's timer starts at 0

#### Scenario: Undo on the page
- **WHEN** the level page is shown during a session and the user presses Ctrl+Z
- **THEN** nothing happens

#### Scenario: Level and tier from one answer
- **WHEN** an answer raises the level to 10 and earns Answers 1,000 while *Show level-ups while studying* is on
- **THEN** the level page is shown, and after Continue a banner says "New: 1,000 answers"
