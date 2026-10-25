# Proposal

## Why

Reaching the daily goal, a level or an achievement is only acknowledged on the session-complete screen, as one line of text and a 1.6-second icon burst. That screen appears only when the queue runs out for today or the user taps Finish after a timebox. A user who leaves with Back, which is common with large decks or high new-card limits, never sees it: an achievement turns up later as a snackbar on the deck list, while a level-up or a reached goal only changes the numbers in the header. Even on the summary, the reward comes many cards after the moment it was earned. The gamification phases chose "no pop-ups during study" to protect focus and the rating keys. This change keeps that protection but rewards the user when the moment happens, and makes the summary reward hard to miss. It builds on `add-levels`.

## What Changes

- **Reward banner during study** (new): when an answer reaches the daily goal, raises the level or earns an achievement tier, a small banner appears for about 2.5 seconds, such as "Daily goal reached", "Level 13" or "New: 1,000 answers" with the badge. It sits at the bottom edge of the card area, just above the answer buttons, where cards are usually empty because they are aligned to the top. It is not a dialog. It takes no focus, does not react to keys, taps or clicks, and never covers the answer buttons or the counts. Each event shows once per session, and several events from one answer share one banner. Under reduce motion it appears and disappears without sliding. The banner revises the archived decisions "No pop-ups during study" (`add-achievements`) and "never during studying" (`add-deck-progress-and-recap`). The protection they were meant to give, no interruption and no stolen input, becomes a requirement of the banner.
- **Reward card on the summary** (replaces the text-only reward): when the session reached the goal, raised the level or earned tiers, the session-complete screen shows a card above the progress lines. The card shows the earned badges as drawn on the achievements screen, a level ring that fills up to the new level, and a goal ring that fills up. It animates for up to 3 seconds with the existing particle burst, never blocks the buttons, and opens the achievements screen when tapped. Under reduce motion it shows its final state. The streak line and the XP-left line remain as text. The goal, level and new-achievement lines move into the card.
- **Leaving before the summary**: when the user leaves a session with Back after it earned something that was not shown, because the banner is off, the deck list shows one short message for the whole session, such as "Reached level 13 · Daily goal reached · New: 1,000 answers", with *Show achievements*. This extends the existing new-achievement message, which today covers only achievements. Nothing earned in the session is announced twice.
- **Setting**: *Settings → Goals and streaks → Show rewards while studying* (on by default, per device). When it is off, nothing appears during study. The summary card and the Back message still appear. Everything stays hidden when *Show streaks and goals* is off.
- No new data in Go, and no change to how XP, levels, goals or achievements are calculated. Progress already reloads after every answer (about 80 µs with 200,000 reviews, `BenchmarkProgressAfterAnswer`).

## Capabilities

### New Capabilities

None. The rewards extend `progress`, which owns the goal, levels, achievements and celebrations.

### Modified Capabilities

- `progress`: new requirements "Reward banner during study", "Reward card on the summary" and "Rewards after leaving a session". "Quiet presentation" allows the non-blocking banner and keeps dialogs out. "Celebrations" moves to the reward card with a 3-second limit. "Summary progress lines" moves the goal-reached and level-reached lines into the card. "New achievement notice" covers tiers announced by the banner. "Goals and streaks settings" gains *Show rewards while studying*.
- `study-session`: "New achievements on the summary" names the session's tiers in the reward card, including those already shown by a banner, and no longer says that nothing appears while studying.

## Impact

- **Go**: one local setting, `HideStudyRewards`, stored inverted so the zero value means "shown", in `golang/core/model.go`, the `localSettings` list and `settings_sync.go` (kept local). The settings service gets a getter and a setter. The fluffy API is regenerated.
- **Flutter**:
  - new `lib/features/progress/reward_banner.dart` and `lib/features/progress/reward_card.dart`
  - in `lib/features/study/study_screen.dart`: event detection after each answer, the session's earned tiers kept in `_StudyScreenState` instead of `_ProgressLines`, the card on `_Summary`, and a report of unshown rewards when the session is left
  - `lib/data/progress_model.dart`: the session rewards waiting for the deck list
  - `lib/features/progress/new_achievement_notice.dart`: goal and level in the message
  - the settings screen switch
  - English and German strings
- **Tests**: widget tests for the banner (shown once, ignores keys and taps, reduce motion, setting off), the card (badges, rings, tap, reduce motion, minimalist) and the Back message. Updated tests for `study_progress_test.dart` and `deck_list_test.dart`. A settings sync test that the setting stays local. The integration flow checks the banner.
- Requires `add-levels` to be archived first.
