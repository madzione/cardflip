# Spec Delta

## ADDED Requirements

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

## MODIFIED Requirements

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
