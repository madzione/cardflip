# Design

## Context

- The proposal ("Why") explains the motivation. The `progress` and `study-session` deltas give the rules.
- `_StudyScreenState` (`lib/features/study/study_screen.dart`) owns a `SessionRewards` (`lib/features/progress/session_rewards.dart`). It records the goal and level at session start and the goal, level and tiers earned. It also records `shown`, what a banner or the summary already showed. `_progressChanged` runs `SessionRewards.update` after every answer and calls `showAll()` once the summary is on show. `_summaryShown` is `state.card == null && state.studied > 0`.
- `_Summary` → `_ProgressLines` builds the streak, goal and level lines and the `RewardCard`. The card animates 2.6 s (`AnimationController`, rings, staggered badge scale-in, then the 12-dot `Celebration` burst). `_studiedBefore` tells whether the session made today a study day.
- `Progress` (`golang/svc_progress.go`) holds `Current`, `Longest`, `TodayStudied`, today's cards/minutes, the goal, `Level`/`XP`/`LevelStart`/`NextLevelXP` and `NewAchievements` (`ID`, `Threshold`, `TierName`, `Key`). `ProgressService` keeps the sorted study days in memory. `stats.Streaks` tracks the first day of the running streak internally (`streak.first`) but does not return it for the current streak. The XP rules are constants in `golang/stats/level.go` (`XPPerGoalDay`, `XPPerTier`).
- Device-only settings follow the `HideStudyRewards` pattern. The field is stored inverted in `core.Settings`, listed in `localSettings` in `settings_sync.go`, and exposed through the app service and `SettingsModel`.
- Audio: `media_kit` is already a dependency, used through `MediaPlayer`/`MediaKitBackend` (`lib/widgets/media_player.dart`) for card audio. That player is a queue tied to card playback and is stopped by `MediaPlayer.stopAll()` on tab switches. No sound assets exist yet, and nothing in the app uses haptics apart from one `selectionClick` in the occlusion editor.
- `reduceMotion(context)` and `settings.minimalist` are the switches every progress widget already honours. Goldens live in `test/goldens/` and are driven by `test/reward_goldens_test.dart`.

## Goals / Non-Goals

**Goals:**
- One place that decides what the pages contain, at the session end, on Back and for a level-up during study: `SessionRewards`, the same object the banner, the card and the deck list message read, extended with "celebrated".
- Pages that are cheap to test: every animation is a pure function of one controller value, so a test or golden can pump to any moment.
- Sounds and haptics behind one small interface that tests replace, so no test touches `media_kit` or platform channels.

**Non-Goals:**
- A mascot, characters or Lottie/Rive animations. Everything is drawn with Flutter widgets and `CustomPainter`, with no new dependency.
- Leagues, gems, quests, combos or any reward for ratings or speed. These stay ruled out as before.
- Sounds for ordinary answers ("correct" dings). Only rewards make sound.
- Full pages during study for anything but a level-up. The goal and tiers keep the banner while studying and get their pages when the session ends or is left with Back.

## Decisions

### 1. The sequence is part of the study screen, not a new route

`_StudyScreenState` gets `List<CelebrationEvent>? _celebrations` and `int _page`. When the summary would appear (`_summaryShown`) and the final progress has arrived in `_progressChanged`, it asks `SessionRewards.celebrations(p, studiedBefore: _studiedBefore)` for the list. While the list is non-empty and `_page < length`, the body renders `CelebrationPage(_celebrations[_page])` instead of `_Summary`. Continue increments `_page`, and Escape/Back sets it to the end.

Until the final progress arrives (normally the same frame; the reload was already started by the last answer), the body shows an empty surface instead of the summary. This avoids flashing the summary before the first page. A 400 ms fallback shows the summary anyway if progress never arrives, for example when the profile is closing.

*Alternative*: push a separate full-screen route. Rejected because the study screen already owns the keyboard `Shortcuts`, `PopScope`, `_openSessions` and the "left before the summary" logic. A route would duplicate the Back handling, and `dispose` would have to decide whether the summary counts as shown while a route sits on top.

`PopScope` on the study screen intercepts Back while a page is shown and skips to the summary instead of leaving. The rating and reveal shortcuts are already inactive without a card. The page adds its own `Shortcuts` for Enter, Space and Escape inside a `Focus` that autofocuses.

**Leaving with Back.** While a card is shown, `PopScope` also intercepts Back (system Back, the app bar's back arrow, Escape through `_LeaveIntent`, and the leave bar of hidden bars, which all call `maybePop`). It builds the list with `celebrations(...)` from the progress after the last answer, waiting for it the same way as at the summary (empty surface, 400 ms fallback). With pages, the body shows them, and a flag `_leaving` makes Continue on the last page, Escape and Back pop the screen instead of showing the summary. Without pages, it pops at once. The pages add their keys to `shown`, so `dispose` hands `ProgressModel.leftSession` nothing to name. The message stays for a screen disposed without Back, for example when the profile is switched.

### 2. `SessionRewards` tracks what was celebrated

`SessionRewards` gains:

- `streakBefore` (Progress `current` at start) and `xpBefore` (Progress `xp` at start), both recorded in `start()`
- `celebrated`, a set of event keys like `shown`, using the existing `goalKey`, `levelKey(n)` and tier keys, plus a new `streakKey`

`celebrations(p, studiedBefore:)` returns the ordered list. Each entry is a sealed `CelebrationEvent` (named so it does not clash with the existing `Celebration` burst widget): `StreakCelebration(from, to, week, best)`, `GoalCelebration(progress, goal, kind, xp)`, `LevelCelebration(fromLevel, fromFraction, toLevel, xpLeft)`, `AchievementCelebration(tier)` and `MoreAchievementsCelebration(tiers)`. It includes everything earned that is not in `celebrated`, ignoring `shown`, because a banner does not replace the page. It applies the rules from the spec: Level tiers are dropped when a level entry exists, at most three single tier pages are created and the rest are collected into one, and the streak is included only when `studiedBefore == false && p.todayStudied`. The keys are added to `celebrated` when the list is built, so a second summary in the same session (after "learning card due now") only gets new events. The method is pure apart from that set, so it is unit tested in `session_rewards_test.dart`.

The XP-earned line on the summary is `p.xp - xpBefore`, shown when it is positive.

`levelCelebration(p)` returns the `LevelCelebration` for a level above the start and above any level celebrated so far, or null, and records it in `celebrated` and `shown`, as `celebrations` does. The level page during study (decision 9) uses it. A later `celebrations` call then leaves that level out, and fills a further level-up from where the page ended.

### 3. Go supplies the week strip and the XP amounts

- `stats.StreakStats` gains `CurrentFirst`, the first day of the current streak's run, which `streak.first` already holds.
- `Progress` gains `Week []string`: the 7 study days ending today, oldest first. A day is `"study"` when it is in `s.days`, `"rest"` when it is not, the current streak is above 0, and the day lies between `CurrentFirst` and today, and `"none"` otherwise. Flutter aligns the strip to `MaterialLocalizations.firstDayOfWeekIndex`, as the weekly recap does. Days before the week start are dropped, and days after today are drawn blank. When today is the first day of the week, only today is shown in the strip.
- `Progress` gains `GoalXP` (`stats.XPPerGoalDay` when a goal is set, else 0), and `NewAchievement` gains `XP` (`stats.XPPerTier[TierName]`, 0 for Level tiers). The pages then show XP amounts without duplicating the rules in Dart.

*Alternative*: compute rest days in Dart from a list of study days. Rejected because the rest-day rule (no 7-day window with more than N skips) already lives in Go's `streak` type, and duplicating it would let the two drift apart.

The cost is one binary search per day over the in-memory days, 7 per refresh. It is negligible next to the existing incremental refresh (`BenchmarkProgressAfterAnswer`).

### 4. Pages: one controller and pure painters

`lib/features/progress/celebration_sequence.dart` contains `CelebrationPage`. It is a `StatefulWidget` with one `AnimationController` of 2.4 s, started on mount and set straight to 1 under reduce motion. Each kind maps the controller value through `Interval`s:

| Page | 0–0.4 | 0.4–0.6 | 0.6–1.0 |
|------|-------|---------|---------|
| Streak | flame scales in (easeOutBack) | number rolls from `from` to `to` (vertical slide of digits) | today's tick stamps in (scale 1.4→1), small burst |
| Goal | ring fills to 1 | ring turns into a check (cross-fade + scale) | burst |
| Level | ring fills from `fromFraction` to 1 | number flips (`from` → `to`, `Transform` rotateX), flash | rays rotate and fade, confetti |
| Achievement | badge drops in from above with a slight overshoot | shine sweep (`ShaderMask` gradient) | stars appear one by one, confetti |

The layout is a centred column: the centrepiece (160 logical px, scaled by UI size), the headline (`headlineMedium`, bold), the supporting line, an XP chip ("+50 XP") and, at the bottom, a full-width `FilledButton` Continue (max width 400). The background is `colorScheme.surface`. Outside minimalist mode, a soft radial glow in `primaryContainer` sits behind the centrepiece, plus rays on the level page. In minimalist mode both are left out.

The input guard is a `_ready` flag set by a 500 ms timer that starts when the page appears (not from the animation, so it also works under reduce motion). Until it is set, Continue is visually enabled but ignores taps, and the Enter, Space and Escape intents return early. *Alternative*: disable the button. Rejected because a button that flickers from disabled to enabled reads as a glitch.

Semantics: the page is wrapped in `Semantics(liveRegion: true)` with the headline and lines as its label, and the decorative centrepiece is excluded. Continue gets focus.

### 5. Confetti

`lib/features/progress/confetti.dart` contains `Confetti`, an `IgnorePointer` `CustomPaint` with about 60 rectangles and circles in the theme's primary, tertiary, secondary and the tier's colour. They burst upward from an origin, then fall under gravity with drift and spin, over 1.8 s. Positions come from a fixed-seed `math.Random(seed)`, so goldens are stable and two bursts on one screen can differ by seed. The existing `Celebration` dot burst stays for the smaller streak/goal bursts and the reward banner.

*Alternative*: the `confetti` package. Rejected because it is a new dependency for about 80 lines of painter, and it uses physics with wall-clock time, which makes goldens flaky.

### 6. Sounds: the C2-arcade set on a dedicated player

- The sounds are the C2-arcade set, chosen from the four candidates in `sound-options/` (glass bells, wooden marimba, arcade and C2-arcade; `sound-options/spectrograms.png` compares them). Its SoundCraft session has instrument tracks only (bell, spark, riser, echo, bass, drums), so it needs no third-party samples and no licence. The five sounds are rendered from it as mono 22,050 Hz 16-bit WAVs into `assets/sounds/`:
  - `chime` (≈0.6 s) for the banner
  - `streak` (≈0.9 s)
  - `goal` (≈0.8 s)
  - `level` (≈1.3 s)
  - `achievement` (≈1.1 s)

  The WAVs are committed, and the session is kept as `tool/sounds/celebration-sounds-C2.scs` so they can be changed and rendered again. A test checks that each file is a mono 22,050 Hz 16-bit WAV under 100 KB.
- WAV was chosen because every target plays it: libmpv on desktop and mobile, and HTML audio on all four web browsers, including Safari, which is unreliable with Ogg. The total size is under 250 KB.
- `lib/features/progress/celebration_feedback.dart` defines `abstract class CelebrationFeedback { void play(CelebrationSound s); void vibrate(CelebrationHaptic h); }` with a `static CelebrationFeedback instance`, which tests replace, following the `MediaPlayer.createBackend` pattern. The real implementation lazily creates one `media_kit` `Player` of its own, separate from `MediaPlayer`, so it is never queued behind card audio, never stops it, and is not affected by `MediaPlayer.stopAll()`. It opens `Media('asset:///assets/sounds/<name>.wav')` at volume 70 %. Haptics call `HapticFeedback.mediumImpact()` or `lightImpact()` only when `!kIsWeb` and the platform is Android or iOS.
- The pages call `play` and `vibrate` once in `initState` when the setting is on. The study screen does the same for the banner where it calls `_showBanner`. The feedback service also checks the setting, so a caller cannot forget it.
- The player is created on the first session start that has the setting on, so the first sound has no decoder warm-up. It is disposed with the study screen.

*Alternative*: reuse `MediaPlayer`. Rejected because it is a sequential queue with per-card generation tracking. A reward chime would cancel or wait behind the card's audio.

### 7. Banner and reward card changes

- `RewardBanner`: the leading badge or icon grows from 28 to 40 px and scales in with `easeOutBack` over the first 300 ms. A `Celebration` burst (`IgnorePointer`) is placed behind the leading icon. The text and timing are unchanged. Under reduce motion there is no scale and no burst, as before.
- The banner's sound is that of its biggest event: `CelebrationSound.achievement` when `events.named` has a tier, else `level` when `events.level` is set, else `chime`. `_showBanner` picks it from the `RewardEvents`.
- `RewardCard`: `animate` is removed and the card always renders its final state (controller value 1, no burst). The `Celebration` class is still used by the banner and the pages.

### 8. Settings

`core.Settings.MuteCelebrations` (`json:"muteCelebrations,omitempty"`) is added to `localSettings`. `AppSettings` gets `CelebrationSounds: !st.MuteCelebrations` and `SetCelebrationSounds(on)`. `SettingsModel.celebrationSounds` is added with a setter that uses the same optimistic-update pattern as `showStudyRewards`. The switch sits under *Show rewards while studying* and is hidden with it.

*Show level-ups while studying* follows the same pattern: `core.Settings.DeferLevelUps` (`json:"deferLevelUps,omitempty"`) in `localSettings`, `AppSettings.ShowLevelUpsWhileStudying: !st.DeferLevelUps`, `AppService.SetShowLevelUpsWhileStudying(on)` and `SettingsModel.showLevelUpsWhileStudying`. Its switch sits under *Celebration sounds and vibration* and is hidden while *Show streaks and goals* is off. It does not depend on *Show rewards while studying*: with banners off, a level-up still gets its page during study.

### 9. The level page during study

When `_progressChanged` handles the progress after an answer, the session still has a card and *Show level-ups while studying* is on, it asks `SessionRewards.levelCelebration(p)`. A level page then covers the body like the session-end pages, and the level is taken out of the banner's events. A Level tier is left out of the banner too, as `RewardEvents.named` already does for a level. The rest of that answer's events (goal, tiers) wait in `_bannerAfterPage` and are shown when the page closes.

The progress normally arrives just after the next card. That card is already on screen under the page, so the page holds it rather than delaying the answer:

- **On show:** the card's audio and speech stop (`_player.stop()`), the answer timer pauses (`_timer.stop()`), and auto advance's countdown is cancelled.
- **On close:** the timer is reset and started, so the card's time starts at 0. The question's audio plays again through `_showQuestion`, and `_restartAutoAdvance()` runs.

Continue, Enter, Space, Escape and Back close the page and return to studying. `PopScope` intercepts Back while it is shown. Ctrl+Z and the rating, reveal and card action intents do nothing while a page is shown: `_StudyAction` checks a `_celebrating` flag. A timebox that falls due on the same answer waits until the page closes, because `_checkTimebox` runs after `_run` and is deferred while a page is shown.

When the level-up comes with the session's last answer, there is no next card, and the level is part of the session-end sequence in its usual place.

*Alternative*: hold every answer's next card until its progress has arrived, so the page comes first. Rejected because it delays every answer for the rare level-up.

## Risks / Trade-offs

- [The sounds may not carry on small speakers] → Listen on a phone speaker and laptop speakers. If they fall short, change the C2-arcade session and render it again; that changes only the asset files.
- [`media_kit` latency on the first `open` of an asset] → Create the player at session start and open `chime` muted once to warm it. If latency stays noticeable on Android, preload each `Media` into a small pool of players.
- [`asset:///` URIs on the web] → `media_kit` maps them to the web asset path. The integration test on Chrome plays a page sound and checks for no error in the console logs (see the memory note on `--enable-chrome-logs`).
- [Users who study in public get surprised by sound] → One switch in Settings, reachable from the Goals and streaks section. Sounds are short and at 70 % volume. The system media volume applies.
- [A daily streak page every first session can feel repetitive] → This is intended, as in Duolingo. Escape or Back skips it, and the input guard is only 500 ms. If feedback says otherwise, a later change can make it a setting.
- [Extra taps before the summary on sessions that earn a lot] → At most 6 pages: streak, goal, level, three achievements, plus one combined page. Escape or Back skips them all.
- [Back now takes an extra step when the session earned something] → The pages show only for what no page celebrated yet, and Escape or Back on the first page leaves at once.
- [The level page interrupts studying] → It is the only interruption, it comes at most once per level, and *Show level-ups while studying* turns it off. The input guard keeps a fast rating key from closing it by accident.
- [The next card flashes before the level page] → The progress usually arrives in the same or the next frame, so the card is visible for at most a frame or two. Its timer, audio and auto advance restart when the page closes, so nothing is lost.
- [Golden churn] → New goldens per page and theme. Reward card goldens are updated once, now static.

## Migration Plan

There is no data migration. `MuteCelebrations` and `DeferLevelUps` default to false (sounds on, level-ups shown while studying). A collection synced to an older device keeps working, because the field is local and never synced. Rollback means reverting the commit.
