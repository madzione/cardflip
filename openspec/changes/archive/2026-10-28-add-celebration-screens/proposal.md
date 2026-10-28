# Proposal

## Why

Streaks, the daily goal, XP, levels, achievements and rewards all work now, but reaching one of them still does not feel like much. A level-up is a 48 px ring that refills inside a card on the summary. A new achievement is a badge the same size as the icons around it. The only motion is a 12-dot burst, and there is no sound and no haptic feedback. Duolingo shows how much these moments can carry. When a lesson ends, every milestone gets its own full-screen moment: a big animation, a number counting up, confetti, a sound, a vibration, and a Continue button that lets the user enjoy it before moving on. This change brings that to Cardflip. A reward only keeps the user going when it comes soon, so the pages also show when a session is left with Back, and a level-up gets its page right away, during the session, unless the user turns that off.

## What Changes

- **Celebration sequence after a session** (new): when a session ends on the session-complete screen, and it extended the streak, reached the daily goal, raised the level or earned achievement tiers, the study screen first shows one full-screen celebration page per event, then the summary. The order is: streak, daily goal, level, achievements. Each page has one large animated centrepiece, a headline, a supporting line, the XP it earned where that applies, and a Continue button. Enter or Space continue, Escape or Back skip to the summary, and input is ignored for the first half second so a fast rating key or Space press from the last card cannot skip a page by accident.
  - **Streak page**: a large flame with the day count rolling up from yesterday's number, "day streak", and a strip of this week's days. Studied days are ticked, rest days are marked and today's tick stamps in. "New personal best" appears when the current streak equals the longest.
  - **Daily goal page**: the goal ring fills and bursts into a check, with today's progress ("34 / 30 cards") and "+20 XP".
  - **Level page**: a large level ring fills to the end of the old level. The number then flips to the new level with a flash, rays and confetti, followed by "Level 13" and the XP to the next level.
  - **Achievement page**: the badge drops in large, about 160 px, with a shine sweep, its tier stars appearing one by one, and confetti. The headline is "New achievement", followed by the tier ("1,000 answers · Silver") and "+50 XP". Up to three tiers get their own page. Further tiers share one last page with their badges in a row.
- **Celebrations when leaving with Back** (new): leaving a session with Back first shows the pages for what it earned and no page celebrated yet, then the deck list. Escape or Back on a page skips the rest and leaves. A session that earned nothing is left at once.
- **Level-up while studying** (new): the answer that raises the level shows the level page right away, before the next card. The next card's timer, auto advance and audio wait until Continue, and undo and rating keys do nothing on the page. Continue, Escape and Back return to studying. Goal and tier events of the same answer get their banner after the page. A new device setting, *Settings → Goals and streaks → Show level-ups while studying*, is on by default. Off, a level-up is a banner and its page waits for the end of the session or for Back.
- **Sounds and vibration** (new): each celebration page plays a short sound of its own (streak, goal, level and achievement), and the in-session reward banner plays the sound of its biggest event: the achievement sound for a tier, the level sound for a level, and a soft chime for the goal. On Android and iOS, the device also vibrates: a medium impact for a page and a light one for the banner. A new device setting, *Settings → Goals and streaks → Celebration sounds and vibration*, is on by default and hidden when *Show streaks and goals* is off. The sounds are the C2-arcade set, made in SoundCraft from instrument tracks only, so they need no third-party licence.
- **Livelier reward banner**: the banner stays non-blocking. It ignores pointers, takes no focus and never covers the buttons. The badge or icon pops in with an overshoot and a small confetti burst goes off over the card area, together with its sound and the light vibration. It still lasts about 2.5 seconds. Big moments that a banner announced during the session are celebrated again in full at the end, like Duolingo does.
- **Summary after the sequence**: the reward card stays as a static recap with no animation of its own and opens the achievements screen as before. A new line, "+48 XP this session", is added to the progress lines.
- **Reduce motion and minimalist mode**: under reduce motion, the pages are still shown in their final state, with no rolling numbers, confetti or flips. Sounds and vibration follow their own setting. In minimalist mode the pages have a plain background with no glow, rays or decorative fills.
- **Not changed**: what XP, levels, streaks, goals and achievements count; the deck list message for a session closed without Back, for example by switching profiles; the nothing-due screen; and no dialogs while studying apart from the level page.

## Capabilities

### New Capabilities

None. The celebrations belong to `progress`, which owns the rewards, the banner, the reward card and celebrations.

### Modified Capabilities

- `progress`:
  - New requirements: "Celebration sequence", "Celebrations when leaving with Back", "Streak celebration", "Daily goal celebration", "Level celebration", "Achievement celebration", "Moving through celebrations", "Celebration sounds and vibration", "Reward banner sound", "Level-up while studying" and "Studying waits for the level page".
  - "Celebrations" now describes the pages instead of the reward card animation.
  - "Quiet presentation" allows the sequence when the session ends or is left with Back, and the level page during study.
  - "Reward banner during study" gains the pop-in, the confetti, the sound and the vibration, and names a level only while the level page during study is off.
  - "Reward card on the summary" becomes a static recap.
  - "Summary progress lines" gains the XP earned in the session.
  - "Goals and streaks settings" gains *Celebration sounds and vibration* and *Show level-ups while studying*.
  - "Rewards after leaving a session" applies only to a session closed without Back.
- `study-session`:
  - "Session completion" shows the celebration sequence before the summary when the session earned something.
  - "Keyboard control": Escape on a celebration page does what that page says instead of leaving.

## Impact

- **Go**: `Progress` gains `Week`, the state of each of the last 7 study days ("study", "rest" or "none"), for the streak page's strip, computed from the study days already held by `ProgressService`. `stats.StreakStats` gains the first day of the current streak. Two new local settings, `MuteCelebrations` and `DeferLevelUps`, are stored inverted so the zero value means on, kept local in `settings_sync.go`, with getters and setters on the app service. The fluffy API is regenerated (`fluffy generate`, without `dart format` on `lib/golang/`).
- **Flutter**:
  - new `lib/features/progress/celebration_sequence.dart` (the pages and their order)
  - new `lib/features/progress/confetti.dart`
  - new `lib/features/progress/celebration_feedback.dart` (sounds and haptics, replaceable in tests)
  - `celebration.dart`, `reward_banner.dart` and `reward_card.dart` updated
  - `session_rewards.dart` records the streak and XP at the session start
  - `study_screen.dart` shows the sequence before `_Summary` and on Back, and the level page during the session
  - the settings screen gets two switches
  - English and German strings are added
- **Assets**: `assets/sounds/` with five short mono WAV files, under 100 KB each, rendered from the C2-arcade SoundCraft session, which is kept as `tool/sounds/celebration-sounds-C2.scs`. They are played through `media_kit`, which is already a dependency, from `asset:///` URIs on a player separate from card audio.
- **Tests**:
  - widget tests for the sequence: order, skipping, the input guard, keys, Back, reduce motion, minimalist mode, the setting off, and Back from the session
  - widget tests for the level page during study: the held timer, auto advance and audio, Escape, undo, the banner after it, and the setting off
  - goldens for each page in light and dark
  - a fake feedback service that checks sounds and haptics
  - Go tests for `Week` and the local setting
  - an updated integration flow
- No new package dependencies.
