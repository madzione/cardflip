# Tasks

## 1. Setting

- [ ] 1.1 Add `HideStudyRewards` to `core.Settings` and `localSettings`, `ShowStudyRewards` to the settings snapshot and `AppService.SetShowStudyRewards` (design decision 5); verify `settings_sync_test.go` keeps it local and a service test round-trips it
- [ ] 1.2 Regenerate the fluffy API with `fluffy generate` (no `dart format` on `lib/golang/`), add `showStudyRewards` to `SettingsModel`, the *Show rewards while studying* switch under *Show streaks and goals* (hidden while that is off), English and German strings, and the fake in `test/support/fake_go.dart`; verify `settings_test.dart` for the switch calling the setter and being hidden, and that `no_literal_text_test` passes

## 2. Session reward detection

- [ ] 2.1 Add `lib/features/progress/session_rewards.dart` (baseline, `earned`, `shown`, `update`) per design decision 1; verify unit tests for goal reached, level raised twice in one session, tiers from one update, no event after an undo reload, the same event again after undo, and a Level tier left out when the level is reported
- [ ] 2.2 Wire it into `_StudyScreenState`: create it from the existing `_studiedBefore`/`_goalReachedBefore`/`_levelBefore` baseline, set `_awaitingProgress` in `_rate`, listen to `app.progress`, mark tiers seen on detection; move the tier list out of `_ProgressLines` so the summary reads `earned`; verify the existing `study_progress_test.dart` summary tests still pass unchanged

## 3. Reward banner

- [ ] 3.1 Add `lib/features/progress/reward_banner.dart` and place it at the bottom of the card area's `Stack` in `study_screen.dart` (design decision 2), with the polite announcement and the setting and *Show streaks and goals* checks; verify `study_progress_test.dart` tests for "Daily goal reached" after the goal-reaching answer, one combined banner for level and tier, no second banner after undo and re-answer, no banner with nothing new, none with the setting off, rating key 3 rating Good while it shows, a tap passing through, the semantics announcement, and no animation under reduce motion
- [ ] 3.2 Add a golden test of the banner with a badge in light and dark themes and in minimalist mode; verify the goldens are reviewed and committed

## 4. Summary reward card

- [ ] 4.1 Add `lib/features/progress/reward_card.dart` (rings, up to three badges and "+N", text, `InkWell` to the achievements screen) and use it in `_ProgressLines` instead of the goal-reached and level-reached lines and the `Celebration` stack (design decision 3); verify `study_progress_test.dart` for the level-and-badge card, five tiers showing "+2" with all named, tapping opening the achievements screen, no card when nothing was earned, "Daily goal reached" as a line when reached in an earlier session, and the nothing-due screen unchanged
- [ ] 4.2 Add the card animation with `Celebration` behind it; verify tests that it settles within 3 s with the buttons tappable during it, that reduce motion shows the final state with no `celebration` key, and a golden of the settled card in light, dark and minimalist
- [ ] 4.3 Update `study-session` behaviour for tiers shown by a banner; verify a `study_progress_test.dart` test that a tier announced by a banner is still named on the card

## 5. Rewards after leaving a session

- [ ] 5.1 Add `leftSession` and `pendingSession` to `ProgressModel`, call it from `_StudyScreenState.dispose` when no summary was shown, and extend `NewAchievementNotice` to add the goal and level to its message (design decision 4); verify `deck_list_test.dart` for "Reached level 13 · Daily goal reached" after Back with the banner off, no message after Back when the banner showed everything, no repeat of a tier the banner showed, and no message after a summary was shown

## 6. Docs and integration

- [ ] 6.1 Update `README.md`'s gamification section for the banner, the card and the setting; verify it matches the spec deltas and no longer says there are no pop-ups during study
- [ ] 6.2 Extend `integration_test/flow.dart` to answer until a goal of a few cards is reached and check the banner appears and the summary shows the card; verify with `LANG=en_US` on Linux
- [ ] 6.3 Run `openspec validate add-study-rewards --strict`, `fvm flutter test` and `go test ./...` in `golang/`; verify all pass

## Workflow follow-up

- Archive `add-levels` before this change: its deltas for "Summary progress lines", "Celebrations" and "New achievements on the summary" are the base these deltas modify.
- Archive after review.
