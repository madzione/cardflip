# Tasks

## 1. Go API

- [x] 1.1 Add `Collection.NextCountChange()` in `golang/core/queue.go`: the earliest of the next unsuspended, unburied Learning/Relearning due time after now and `scheduler.NextDayStart(now, rollover)`. Add queue tests with an injected clock covering: a learning card due in 1 minute wins over rollover; suspended and buried learning cards are ignored; with no learning cards the result is the next day start. Verify that `go test ./core/...` passes.
- [x] 1.2 Add `DecksService.NextChange() (time.Time, error)` in `golang/svc_decks.go`, and add a `svc_test.go` case that answers a new card Again and asserts `NextChange` equals its due time. Verify that `go test ./...` passes.
- [x] 1.3 In `StudyService.state` (`golang/svc_study.go`), set `LearningDue` whenever `Queue.LearningLater` is non-zero, not only when no card is picked, and update its doc comment. Add a `svc_test.go` case: with one learning card pending and a review card on screen, `LearningDue` is set. Verify that `go test ./...` passes.
- [x] 1.4 Add `AutoContinueLearning bool` (`json:"autoContinueLearning,omitempty"`) to `core.Settings`, the field to the `Settings` DTO in `golang/svc_app.go`, and `AppService.SetAutoContinueLearning(bool)`. Add a `svc_test.go` case that sets it, reopens the collection and reads it back as true, and one that a collection without the field reads false. Verify that `go test ./...` passes.
- [x] 1.5 Run `fluffy generate` to regenerate `lib/golang/api.g.dart`, and verify that `fluffy generate --check` is clean and `fvm flutter analyze` reports no new issues.

## 2. Deck list refresh

- [x] 2.1 Add `decks.NextChange` (default: one hour ahead) and `app.SetAutoContinueLearning` handlers to `test/support/fake_go.dart`, plus the new settings field, and verify that the existing `fvm flutter test` suite still passes.
- [x] 2.2 In `DeckListModel`, call `decks.nextChange()` after `tree()` and keep one timer at `nextChange + 1 s`, capped at 15 minutes, replaced on every load and cancelled in `dispose`. Add a `test/deck_list_test.dart` test where the fake returns `nextChange = now + 60 s` and a tree with learning 0, then learning 1 on the next call; after `tester.pump(61 s)` the deck row shows 1. Verify that it passes.
- [x] 2.3 Reload `DeckListModel` on `AppLifecycleState.resumed` through an `AppLifecycleListener`. Add a test that sends a resumed lifecycle state and asserts `decks.Tree` was called again. Verify that it passes.
- [x] 2.4 Detect newly due learning cards: remember the previous `nextChange` and, on any reload after that moment, report top-level decks whose learning count increased through a broadcast `learningDue` stream. Add model-level tests: a time-driven increase emits one event with the deck; a reload before the due moment (simulating an answer) emits nothing; two top-level decks increasing emit one event with both. Verify that they pass.

## 3. Reminder

- [x] 3.1 Add an open-study-session `ValueNotifier<int>` to `AppState` in `lib/data/scope.dart`; increment it in `StudyScreen.initState` and decrement it in `dispose`. Add a `test/study_test.dart` assertion that the count is 1 while the screen is open and 0 after it pops.
- [x] 3.2 In `_AdaptiveScaffoldState` (`lib/app/scaffold.dart`), listen to `learningDue` and show a snackbar via `messengerKey` unless a session is open: one deck → "Learning cards are due in <deck>" with **Study** (select the Decks tab, pop its navigator to root, push `StudyScreen`); several → "Learning cards are due in N decks" with **Show decks**. Add `test/app_shell_test.dart` tests: reminder appears on the Browse tab and Study opens the study screen for that deck on the Decks tab; several-decks wording and action; no reminder while a study screen is open. Verify that they pass.

## 4. Study screen

- [x] 4.1 Add `autoContinueLearning` to `SettingsModel` with a setter, and a `SwitchListTile` "Continue automatically when learning cards are due" under "Next day starts at" in `settings_screen.dart`. Add a `test/settings_test.dart` test that toggling the switch calls `app.SetAutoContinueLearning(true)` and the switch shows on. Verify that it passes.
- [x] 4.2 In `_StudyScreenState`, keep one timer at `learningDue + 1 s` after every applied state, cancelled in `dispose`. With a card on screen it calls `_run(golang.study.current, keepRevealed: true)`, retrying after 1 s while busy. Add a `test/study_test.dart` test: the answer is revealed, the timer fires, the learning count updates and the answer stays revealed. Verify that it passes.
- [x] 4.3 With no card and auto-continue on, the timer calls `_run(golang.study.current)`. Add a test: the summary shows "More learning cards will be due…", after the timer the learning card's question is shown with no tap. Verify that it passes.
- [x] 4.4 With no card and auto-continue off, the timer fetches `study.current` into `_pending`; if it has a card, `_Summary` and `_NothingDue` show "Learning cards are due now" with a **Continue** button that applies `_pending` and starts the answer stopwatch; if not, apply the state and reschedule. Add tests: prompt and Continue from the summary; prompt from "Nothing due right now" when the session started before the card was due; no change when `learningDue` is null. Verify that they pass.

## 5. Integration check

- [x] 5.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 5.2 Manually on desktop: study a new card, rate Again (`<1m`), go back to Decks and confirm the learning count turns 1 and the reminder appears within a few seconds of the due time; then repeat staying on "Session complete" with the setting off (prompt appears) and on (card appears).
