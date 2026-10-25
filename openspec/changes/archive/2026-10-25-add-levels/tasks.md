# Tasks

## 1. XP and the level curve

- [x] 1.1 Add `golang/stats/level.go` with the XP constants, `LevelOf(xp)` and `LevelStatus`; verify a table test for 0, 49, 50, 2,249, 2,250 and a large XP value (level, start, next)
- [x] 1.2 Split `EvalAchievements` into the two passes of design decision 1 (day numbers kept until the end, `History.XP` built between the passes, `FromXP` on the table); verify the existing `achievements_test.go` still passes unchanged and a new test for the XP of one day (answers, study day, goal met, mature cards, a tier bonus, Comeback as gold)

## 2. New achievements

- [x] 2.1 Add `study-time` to the table in screen order; verify table tests for progress in whole hours, each tier's earned day, and undo below a threshold
- [x] 2.2 Add `level` (second pass); verify tests that the Level 10 tier is earned on the day XP reach 2,250, that Level tiers add no XP, and that raising the goal can lower the level and un-earn a tier
- [x] 2.3 Add `SeenAchievementIDs` to `core.Settings` and `localSettings` and the evaluated-ids rule of design decision 3 in `ProgressService.refresh()`; verify service tests: a new profile (nil rule), an upgrade with history earning `study-time` and `level` tiers silently, and a later new tier announced as usual; verify `settings_sync_test.go` keeps it local

## 3. Service and MCP

- [x] 3.1 Add `Level`, `XP`, `LevelStart` and `NextLevelXP` to `Progress` and `ProgressService.Level()` with the breakdown; verify service tests after an answer, after undo and after a goal change, and extend `svc_progress_bench_test.go` to cover the XP pass with 200,000 reviews
- [x] 3.2 Add the `level` object to the MCP statistics tool; verify `mcp_tools_test.go` for a deck scope returning the collection's level
- [x] 3.3 Add a sync end-to-end test in `golang/sync_e2e_test.go` where both devices report the same level after syncing; verify it passes
- [x] 3.4 Regenerate the fluffy API with `fluffy generate` (no `dart format` on `lib/golang/`); verify the Flutter build compiles

## 4. Flutter

- [x] 4.1 Add the `study-time` and `level` looks, names, tier captions and notice labels to `lib/features/progress/achievements.dart`; extend `test/support/fake_go.dart`; verify `achievement_badge_test.dart` goldens with both new badges in light and dark themes
- [x] 4.2 Add the level ring to `progress_header.dart` with its own tap target and semantics label; verify `deck_list_test.dart` for level 1 when never studied, 460 / 600 at 3,760 XP, tapping the level opens the achievements screen, tapping the rest opens the statistics, and the minimalist look
- [x] 4.3 Add `level_card.dart` above the grid on the achievements screen; verify `achievements_test.dart` for the breakdown and the no-goal hint
- [x] 4.4 Add the level line, `_levelBefore`, the Level tier filter on the new-achievement line and the level-up celebration to `study_screen.dart`; verify `study_progress_test.dart` for "Reached level 13", "Level 12 · 140 XP to level 13", the filtered line, the celebration and no animation under reduce motion
- [x] 4.5 Add English and German strings for the level, the card, the summary lines and the two achievements; verify `no_literal_text_test` and l10n generation pass

## 5. Assets, docs and integration

- [x] 5.1 Add the Study time and Level badges to `assets/badges/achievement-badges.vectorcraft` and re-export the SVG and PNG; verify the PNG shows all eight badges and matches the widget goldens in style
- [x] 5.2 Update `README.md` with XP, levels and the two badges; verify the text matches the `progress` spec delta
- [x] 5.3 Extend `integration_test/flow.dart` to check the header level and the level card after the study flow; verify with `LANG=en_US` on Linux
- [x] 5.4 Run `openspec validate add-levels --strict`; verify it passes

## Workflow follow-up

- Archive after review.
