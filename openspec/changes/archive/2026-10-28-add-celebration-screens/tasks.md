# Tasks

## 1. Go data and setting

- [x] 1.1 Add `CurrentFirst` to `stats.StreakStats` and `Week` (7 days ending today: "study", "rest", "none") and `GoalXP` to `Progress`, and `XP` to `NewAchievement` (design decision 3); verify `streak_test.go` for `CurrentFirst` with and without rest days, and `svc_progress_test.go` for `Week` with a rest day inside the streak, a missed day before the streak, a day without study before any study, and `XP` 50 for a silver tier and 0 for a Level tier
- [x] 1.2 Add `MuteCelebrations` to `core.Settings` and `localSettings`, `CelebrationSounds` to the settings snapshot and `AppService.SetCelebrationSounds` (design decision 8); verify `settings_sync_test.go` keeps it local and a service test round-trips it
- [x] 1.3 Regenerate the fluffy API with `fluffy generate` (no `dart format` on `lib/golang/`) and extend `test/support/fake_go.dart` with `week`, `goalXP`, the tier `xp` and the setting; verify `go test ./...` in `golang/` and `fvm flutter analyze` pass

## 2. Setting in the app

- [x] 2.1 Add `celebrationSounds` to `SettingsModel`, the *Celebration sounds and vibration* switch under *Show rewards while studying* (hidden while *Show streaks and goals* is off), and English and German strings; verify `settings_test.dart` for the switch calling the setter and being hidden, and that `no_literal_text_test` passes

## 3. Sounds and haptics

- [x] 3.1 Add the C2-arcade sounds from `sound-options/C2-arcade/` as `assets/sounds/{chime,streak,goal,level,achievement}.wav` and its session as `tool/sounds/celebration-sounds-C2.scs` (design decision 6), and register them in `pubspec.yaml`; verify a test that each file is a mono 22,050 Hz 16-bit WAV under 100 KB, and listen on a laptop and a phone speaker
- [x] 3.2 Add `lib/features/progress/celebration_feedback.dart` (`CelebrationFeedback` interface, `instance`, a `media_kit` implementation with its own player at 70 % volume, haptics on Android/iOS only, no-op when the setting is off) and a `FakeCelebrationFeedback` in `test/support/`; verify unit tests that the fake records nothing when the setting is off, and that `vibrate` is not forwarded on Linux or the web (`debugDefaultTargetPlatformOverride`)

## 4. Session rewards for celebrations

- [x] 4.1 Extend `SessionRewards` with `streakBefore`, `xpBefore`, `celebrated` and `celebrations(p, studiedBefore:)` returning the sealed `CelebrationEvent` list (design decision 2); verify `session_rewards_test.dart` for the order streak/goal/level/achievements, the streak only when the session made today a study day, a Level tier dropped beside a level entry, five tiers giving three single pages and one combined page, events shown by a banner still celebrated, and a second call after new answers returning only new events

## 5. Confetti and pages

- [x] 5.1 Add `lib/features/progress/confetti.dart` (fixed-seed, `IgnorePointer`, about 1.8 s) (design decision 5); verify a widget test that it ignores taps and finishes, and a golden at mid-burst
- [x] 5.2 Add `lib/features/progress/celebration_sequence.dart` with `CelebrationPage` for the streak (flame, rolling number, week strip aligned to `firstDayOfWeekIndex`, "New personal best"), goal, level and achievement pages and the combined page, the 500 ms input guard, Enter/Space/Escape, semantics, minimalist mode and reduce motion (design decision 4), playing the page's sound and medium haptic once on show; verify widget tests for each page's texts from the spec scenarios ("13 day streak", "34 / 30 cards", "+20 XP", "Level 13", "640 XP to level 14", "1,000 answers · Silver", "+50 XP"), the week strip with a rest day and Monday/Sunday week starts, the guard ignoring Space at 200 ms and accepting it at 600 ms, the final state at once under reduce motion, and the fake feedback receiving exactly one sound and haptic per page
- [x] 5.3 Add goldens for the streak, goal, level, achievement and combined pages at their final state in light and dark themes, plus the level page in minimalist mode; verify the goldens are reviewed and committed

## 6. Study screen integration

- [x] 6.1 Show the sequence before `_Summary` in `_StudyScreenState` (design decision 1): build the list once the final progress has arrived, show an empty surface until then with a 400 ms fallback, Continue/Escape/Back (`PopScope`) handling, and no sequence when *Show streaks and goals* is off; verify `study_progress_test.dart` for the streak, goal and level pages in order then the summary, no pages when nothing was earned, Escape skipping to the summary, system Back on a page showing the summary instead of leaving, rating key 3 doing nothing on a page, the level page after a banner already said "Level 13", and no new pages on a second summary after "learning card due now"
- [x] 6.2 Make `RewardCard` static (remove `animate`, controller and burst) and add the "+N XP this session" line to `_ProgressLines`; verify `study_progress_test.dart` for "+48 XP this session", no XP line after answer-then-undo, the card without a `celebration` key, and updated reward card goldens
- [x] 6.3 Make `RewardBanner` livelier (40 px leading badge or icon with an `easeOutBack` pop-in, a `Celebration` burst behind it) and play the chime and light haptic from `_showBanner`; verify `study_progress_test.dart` that the banner still ignores taps and keys, has no `celebration` key under reduce motion, and the fake feedback records `chime` once per banner and nothing with the setting off, plus updated banner goldens

## 7. Docs and integration

- [x] 7.1 Update `README.md`'s gamification section for the celebration pages, sounds and vibration, and the new setting; verify it matches the spec deltas
- [x] 7.2 Extend `integration_test/flow.dart` to finish a session that reaches a small goal and step through the goal and level pages with Continue to the summary; verify with `LANG=en_US` on Linux and in Chrome, with `--enable-chrome-logs` showing no audio errors
- [x] 7.3 Run `openspec validate add-celebration-screens --strict`, `fvm flutter test` and `go test ./...` in `golang/`; verify all pass

## 8. Back and the level page during study

- [x] 8.1 Add `DeferLevelUps` to `core.Settings` and `localSettings`, `ShowLevelUpsWhileStudying` to the settings snapshot and `AppService.SetShowLevelUpsWhileStudying` (design decision 8), regenerate the fluffy API and extend `test/support/fake_go.dart`; verify `settings_sync_test.go` keeps it local, a service test round-trips it, and `go test ./...` and `fvm flutter analyze` pass
- [x] 8.2 Add `showLevelUpsWhileStudying` to `SettingsModel` and the *Show level-ups while studying* switch under *Celebration sounds and vibration* (hidden while *Show streaks and goals* is off), with English and German strings; verify `settings_test.dart` for the switch calling the setter and being hidden, and `no_literal_text_test`
- [x] 8.3 Add `SessionRewards.levelCelebration(p)` (design decision 2); verify `session_rewards_test.dart` that it returns the level once, records it as celebrated and shown, that a later `celebrations` call leaves it out, and that a further level-up fills from where the page ended
- [x] 8.4 Show the level page during study (design decision 9): hold the next card's timer, audio and auto advance until the page closes, make undo, rating and card action keys do nothing on it, show the same answer's goal and tier banner after it, and defer a timebox message; verify `study_progress_test.dart` for "Level 13" before the next card, the timer at 0 after Continue, Escape returning to studying, Ctrl+Z and 3 doing nothing, "New: 1,000 answers" in a banner after Continue without the Level tier, no level page at the session end, the banner instead with the setting off, and the level sound and one medium vibration on the page
- [x] 8.5 Show the pages when leaving with Back (design decision 1): wait for the final progress, pop after the last page, on Escape or Back, and at once without pages; verify `study_progress_test.dart` for the goal page then the deck list, Escape on the first of two pages leaving, nothing earned leaving at once, a banner-shown level still getting its page, and `deck_list_test.dart` for no message after Back and the message after a session closed without Back
- [x] 8.6 Play the banner's biggest event's sound (design decision 7); verify `study_progress_test.dart` for the chime with the goal alone, the achievement sound with a tier, and the level sound for a level banner with *Show level-ups while studying* off
- [x] 8.7 Update `README.md` for Back and the level page during study, run the integration flow on Linux and in Chrome (`LANG=en_US`, `--enable-chrome-logs`), and run `openspec validate add-celebration-screens --strict`, `fvm flutter test` and `go test ./...`; verify all pass

## Workflow follow-up

- Try the sequence on a real Android phone and in the web build for timing, sound volume and haptic strength before archiving.
- Archive after review.
