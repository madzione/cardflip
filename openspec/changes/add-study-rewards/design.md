# Design

## Context

- `ProgressModel` (`lib/data/progress_model.dart`) reloads `golang.progress.today()` whenever the deck list reloads, and the deck list reloads after every change to the collection, including each answer and each undo. The study screen already gets a fresh `Progress` shortly after every answer, with no polling. The Go side updates incrementally: `BenchmarkProgressAfterAnswer` measures about 78 µs per refresh with 200,000 reviews (Ryzen 5 5600X).
- `_StudyScreenState` records `_studiedBefore`, `_goalReachedBefore` and `_levelBefore` at session start. `_ProgressLines` builds the summary lines, keeps the session's new tiers in its own `_new` list, marks them seen, and picks one icon for the `Celebration` overlay (`lib/features/progress/celebration.dart`).
- Tiers this device has not shown come from `Progress.newAchievements`, minus the ones `ProgressModel.markSeen` marked. `NewAchievementNotice` shows them on the deck list once no study screen is open (`openSessions == 0`).
- The study body is a column: counts row, notices, an `Expanded` card area (a scroll view with the flag and star in its top corners), a divider, then the reveal or rating buttons. Keyboard shortcuts are `Shortcuts`/`Actions` on the whole screen.
- Motivation is in `proposal.md`. The rules are in the `progress` and `study-session` spec deltas.

## Goals / Non-Goals

**Goals:**
- One detector of session rewards. The banner, the summary card and the message after Back all read from it, so they agree about what was earned and what was already shown.
- No new Go computation and no change to the `Progress` shape.

**Non-Goals:**
- Sounds or haptics.
- Rewards for anything new: per-card events, combos, speed or ratings. The proposal for streaks and goals rules these out.
- A reward history or a "rewards earned today" list. The achievements screen covers that.
- Making Back show the summary first. Back stays a plain exit.

## Decisions

### 1. Detecting rewards in the study screen

`_StudyScreenState` gets a `SessionRewards` object, a small plain Dart class in `lib/features/progress/session_rewards.dart`. It holds:

- the baseline at session start: whether the goal was reached, and the level, as today
- `earned`: goal reached, the highest level reached and the tiers, kept in order
- `shown`: the event keys a banner already showed (`goal`, `level:13`, `answers:1000`)

`_rate` sets `_awaitingProgress = true` before calling Go. A listener on `app.progress` runs `rewards.update(progress)` only while `_awaitingProgress` is set, then clears it. `update` returns the events that are new since the last update and not in `shown`:

- the goal is now reached and was not at the baseline
- the level is above the baseline and above every level already reported
- tiers in `newAchievements` that are not in `earned`

It adds the new tiers to `earned` and calls `markSeen` for them right away. The tiers then belong to the session, not to the deck list notice.

Undo reloads progress but does not set the flag, so it produces no event. Answering again after an undo finds the same keys in `shown` or `earned`, so it produces no second banner (spec scenario "Undone and answered again"). A sync landing during the session, which is rare, can deliver tiers in the same reload as an answer and be reported with it. That is accepted.

*Alternative*: return the events from `study.answer` in Go. Rejected because it would duplicate the progress refresh in the answer path and add fields to `StudyState` for something Flutter can already tell from data it receives.

`_ProgressLines` loses `_new` and its call to `markSeen`. It takes the tier list from `SessionRewards.earned`. The summary then names tiers the banner already showed, and nothing is marked seen twice. When the summary appears, the screen awaits one final progress load (as `_ProgressLines.initState` does today) and runs `update` without showing a banner, so the last answer's rewards still reach the card.

### 2. The banner

`lib/features/progress/reward_banner.dart`: a `RewardBanner` widget that the study screen places as a `Positioned` at the bottom center of the card area's `Stack`, 12 px above the divider. It is a pill: an icon or a 28 px `AchievementBadge` for the first tier, then the text (`l.goalReached`, `l.levelBannerText(n)`, `newAchievementsLine`) joined with " · ". It uses the surface-container-high colour with an outline, and no shadow in minimalist mode.

- The widget is wrapped in `IgnorePointer` and `ExcludeFocus`, and has no `Focus` or `Shortcuts` of its own, so taps fall through and keys keep reaching the screen's `Shortcuts`.
- `SemanticsService.sendAnnouncement` with `Assertiveness.polite` replaces a live region, because a live region would sit in the traversal order. The visual banner is `ExcludeSemantics`.
- Motion: it slides up 16 px with a fade over 200 ms, holds, then fades out over 300 ms, about 2.5 s in total. Under `reduceMotion(context)` it appears and disappears at once and stays for the same time.
- It is held in `_StudyScreenState` as `_banner` (the events and a key). A new event set replaces it, and the new key restarts the timer. When the banner hides, its keys go into `rewards.shown`. When it starts, a key goes into `shown` as well, so the banner counts as shown even if the user leaves halfway through it.
- It is hidden when *Show streaks and goals* is off or *Show rewards while studying* is off. Detection still runs, so the summary and the Back message still work.

*Alternative*: a top banner over the counts row. Rejected because it would hide the counts the user reads before answering. The bottom of the card area is usually empty because cards are aligned to the top. *Alternative*: a `SnackBar`. Rejected because snackbars sit over the rating buttons on phones, take part in focus traversal on desktop, and the study screen already hides snackbars when it opens.

### 3. The summary reward card

`lib/features/progress/reward_card.dart`: a `RewardCard` that takes `goalReached`, `levelFrom` (level and XP fraction at session start), `levelTo` and the tiers. It replaces the `Stack` with `Celebration` in `_ProgressLines` when `SessionRewards` has something earned. The streak line and the remaining goal or XP lines stay below it as text.

- Layout: a row of up to three items (goal ring with `Icons.flag_outlined`, level ring with the number, up to three badges, then "+N"), and below them one or two lines of text with the same strings as today.
- Animation, about 2.6 s, driven by a single `AnimationController`:
  - 0–1.2 s: the level ring fills from its start fraction to 1, then the number changes and the ring restarts at the new level's fraction.
  - 0.4–1.0 s: the goal ring fills from 0.
  - 0.8 s onwards: the badges scale in one after another.
  - From 1.2 s: `Celebration` plays its particle burst behind the card.
  - Everything sits under `IgnorePointer` except the card's own `InkWell`.
- Under reduce motion, the controller is set to 1 and no `Celebration` is added.
- Tapping the card calls `showAchievements(context)`. Semantics give it a button role whose label is the card's text.
- The 48 px icon chosen by priority (goal, then achievements, then level) goes away. The card shows all three at once.

### 4. Rewards after leaving a session

In `dispose`, when the summary was never shown, `_StudyScreenState` calls `app.progress.leftSession(rewards)` with the baseline and `shown`. `ProgressModel` keeps this as `pendingSession`. Dispose can come before the last answer's progress load finishes, so the comparison runs later: `NewAchievementNotice._check`, which already waits for `openSessions == 0` and a loaded progress, compares `pendingSession` with the current `Progress`:

- goal reached now, not at the baseline and not in `shown`
- level above the baseline and `level:<n>` not in `shown`
- plus the unseen tiers it already handles

When something is found, it builds one message with the same strings joined by " · " (`newAchievementsMessage` handles more than three tiers as today), marks the tiers seen and clears `pendingSession`. When the summary was shown, nothing is handed over: the card acknowledged everything up to then. "Check again" followed by more answers then Back is an edge case. The session keeps one `SessionRewards`, and the summary marks everything earned so far as shown, so only later events are handed over.

*Alternative*: show the summary when Back is pressed after the first answer. Rejected because it changes navigation that users and tests rely on, and adds a step for users who just want out.

### 5. The setting

`HideStudyRewards bool` in `core.Settings`, listed in `localSettings` and therefore kept out of sync like `HideProgress`. It is stored inverted so the zero value means "shown". The settings snapshot in `golang/svc_app.go` gains `ShowStudyRewards` (`!st.HideStudyRewards`), `AppService` gains `SetShowStudyRewards(on bool)`, and `SettingsModel` gains `showStudyRewards`, all following `ShowProgress`. The switch sits under *Show streaks and goals* and is hidden while that is off. After the Go change, run `fluffy generate` and leave `lib/golang/` unformatted.

## Risks / Trade-offs

- [Level-ups happen every few sessions early on (`add-levels` design: "could feel noisy"), and a banner each time may feel like too much] → It is one line for 2.5 s with no sound, a level shows at most once a session, and the setting turns it off. If testing shows it is still too much, levels 1 to 4 can show without a banner, which changes only the spec scenario thresholds.
- [The banner covers the last lines of a long card's answer for 2.5 s] → It ignores pointers, so scrolling under it still works. It appears right after an answer, when the next question is shown at the top of the scroll view.
- [The progress reload after an answer is asynchronous, so the banner can come a few frames after the next card] → That is acceptable for a non-blocking banner. Widget tests pump until the reload completes.
- [A sync during a session can make an earlier device's tiers look like this session's] → They are still new to this device, and the banner names them correctly. Only "earned in this session" is slightly off.
- [This reverses an archived decision ("no pop-ups during study")] → The decision's reasons, focus and input safety, become testable requirements ("Reward banner does not interrupt"). The README section on gamification needs updating so the docs do not contradict the app.

## Migration Plan

The setting's zero value means "on", so existing installs get banners without migration. Seen state is unchanged. Tiers earned before the upgrade stay seen.
