# Spec Delta

## ADDED Requirements

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

## MODIFIED Requirements

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
