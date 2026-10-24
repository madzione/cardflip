# Design

## Context

- `DeckListModel` (`lib/data/deck_list_model.dart`) reloads `decks.tree()` only when `ChangeHub` emits a collection change. Counts come from `Collection.DeckCounts()` in `golang/core/queue.go`, which counts learning cards with `Due <= now`. A learning card that becomes due causes no write, so no change event, so no reload.
- `StudyService.state` (`golang/svc_study.go`) sets `StudyState.LearningDue` (from `Queue.LearningLater`) only when no card is picked. The study screen shows it in `_Summary` / `_NothingDue` and offers "Check again", which calls `study.current`. Nothing reloads on its own.
- The app has no notification plugin. In-app messages go through the global `messengerKey` (`lib/app/errors.dart`). Each tab has its own `Navigator` inside `AdaptiveScaffold` (`lib/app/scaffold.dart`).
- Settings live in `core.Settings` (stored in the collection, so backups include them) and are mirrored by `SettingsModel`.
- Widget tests drive Go through `FakeGo` and advance Dart timers with `tester.pump(duration)`; `DateTime.now()` is not faked.

## Goals / Non-Goals

**Goals:**
- One source of truth for "when do counts change on their own", computed in Go next to the queue rules.
- No polling: one timer per screen, set to the next relevant moment.
- Reminder and auto-refresh work the same on native and web.

**Non-Goals:**
- System notifications, or anything that runs while the app is closed or in the background.
- Reacting to the day rollover inside the study screen (the deck list covers it; a session left open overnight still needs "Check again" or leaving and re-entering).
- Learn-ahead (showing learning cards before they are due). The queue has no learn-ahead today and this change does not add one.

## Decisions

### 1. Go reports the next time counts change; Dart sets a timer

Add `Collection.NextCountChange() time.Time` in `golang/core/queue.go`: the earliest of
- the due time of any unsuspended, unburied Learning or Relearning card due after now, and
- `scheduler.NextDayStart(now, rollover)` (new day: reviews, new-card limits and buries change).

Expose it as `DecksService.NextChange() (time.Time, error)`. `DeckListModel.load()` calls it after `tree()` and sets a single `Timer` for that moment plus 1 s slack (Go counts a card as due at `Due <= now`; the slack absorbs rounding and timer jitter). Every reload cancels and replaces the timer.

Alternatives considered:
- *Go emits a change event when a card becomes due.* Rejected: the changes stream means "the collection was written"; every listener (Browse, settings) would reload for nothing, and a Go timer on the FFI/WASM side is harder to drive in Dart widget tests.
- *Poll every N seconds.* Rejected: up to N seconds late, wasted work, and still no clean answer for rollover.
- *Compute it in Dart from the tree.* Rejected: the tree has no due times, and suspend/bury/rollover rules would be duplicated.
- *Return it inside `Tree()`.* Rejected to keep `Tree()`'s signature and its callers unchanged; a second call is cheap.

### 2. Cap the timer and refresh on resume

The timer is capped at 15 minutes; when the cap fires the model just reloads and sets the next timer. This covers clock changes and desktop sleep, where a long Dart timer can fire late. `DeckListModel` also reloads on `AppLifecycleState.resumed` through an `AppLifecycleListener` (on mobile, timers pause in the background; on web, hidden tabs throttle timers and Flutter maps `visibilitychange` to the lifecycle).

### 3. Detect newly due learning cards from time-driven reloads

The model remembers the previous `nextChange`. On any reload where now has passed that time, it compares each top-level deck's (depth 0) learning count with the previous load. Decks whose count went up are reported as newly due. Using "the due moment has passed" instead of "the timer fired" means a change-triggered reload that happens to land just after the due time (for example, saving a note in Browse at that moment) still produces the reminder. Reloads before the due moment never do, so answering cards or importing does not trigger reminders.

The model exposes these as a broadcast stream of events (`LearningDue { decks }`), separate from `notifyListeners()`, so the deck list rebuilds as before and only one listener shows reminders.

### 4. The scaffold shows the reminder and opens the session

`_AdaptiveScaffoldState` subscribes to the event stream and shows a `SnackBar` through `messengerKey` (duration about 10 s, replacing any current snackbar):
- one deck: "Learning cards are due in <deck>" with **Study**, which selects the Decks tab, pops its navigator to the root and pushes `StudyScreen` for that deck;
- several decks: "Learning cards are due in N decks" with **Show decks**, which selects the Decks tab and pops it to the root.

It skips the reminder while a study session is open. `StudyScreen` marks itself in an app-scope `ValueNotifier<int>` (open session count) in `initState` / `dispose`; `AppState` (`lib/data/scope.dart`) owns it. `StudyScreen` already hides the current snackbar on open, which also clears a reminder the user ignored.

The scaffold owns this because it owns the per-tab navigator keys and is mounted for the whole app lifetime.

### 5. Study screen: one learning timer driven by `StudyState.LearningDue`

Go: `StudyService.state` sets `LearningDue` whenever `Queue.LearningLater` is non-zero, not only when no card is picked. The existing UI reads it only when `card == null`, so this is safe.

Dart: after every applied state, `_StudyScreenState` replaces a single timer set to `learningDue + 1 s` (none when null), cancelled in `dispose`. When it fires:
- **Card on screen:** `_run(golang.study.current, keepRevealed: true)`. `Current` keeps the same card, so only the counts change, and the answer stays revealed. If a call is already in flight (`_busy`), retry after 1 s.
- **No card, auto-continue on:** `_run(golang.study.current)`. The due card appears like after "Check again".
- **No card, auto-continue off:** call `study.current` and keep the result as `_pending` without applying it. If it has a card, `_Summary` / `_NothingDue` switch to a "Learning cards are due now" message with a primary **Continue** button that applies `_pending` (and starts the answer stopwatch then, not when the card was fetched). If it has no card (for example, the card was suspended in Browse meanwhile), apply it as normal and set the next timer from it.

Fetching in prompt mode instead of just flipping a flag confirms a card really is due before telling the user so. Go's session already holds the picked card as `current`, so applying `_pending` later is consistent.

### 6. Setting: `AutoContinueLearning`, off by default

`core.Settings` gains `AutoContinueLearning bool` (`json:"autoContinueLearning,omitempty"`); the Go `Settings` DTO gains the field and `AppService` gains `SetAutoContinueLearning(bool)`. Old collections and backups lack the field and read as false, so no migration is needed. `SettingsModel` mirrors it; the settings screen gets a `SwitchListTile` "Continue automatically when learning cards are due" under "Next day starts at". The study screen reads it from `AppScope.of(context).settings` when its timer fires, so changing it mid-session takes effect at the next due moment.

## Risks / Trade-offs

- [Device clock and Go clock disagree, or clock changes] → Both read the device clock in production; the 15-minute cap and resume reload bound the error.
- [Background on mobile: timers do not fire, so no reminder] → Accepted; system notifications are a non-goal. The resume reload updates counts, and the newly-due check produces the reminder on return.
- [Snackbar covering study buttons] → The reminder is never shown while a session is open, and the study screen's own reaction is in-page, not a snackbar.
- [Timer fires during a rating tap] → `_run` ignores calls while busy; the timer retries after 1 s.
- [Several cards due seconds apart cause repeated reminders] → Each time-driven reload replaces the current snackbar instead of stacking; the user sees at most one reminder at a time.
- [In-flight `card-actions-and-wording` also edits `deck_list_screen.dart`, `study_screen.dart` and the app-shell spec] → This change touches different requirements (adds new ones, modifies only "App settings") and different code regions; rebase whichever lands second.
