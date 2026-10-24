# Proposal

## Why

When the user rates a card Again (`<1m`) and goes back to the deck list, the learning count stays at 0 after the minute is up: the list only reloads on collection changes, and time passing is not one. The study screen has the same gap: "Session complete" says "More learning cards will be due at 14:05", but nothing happens at 14:05 until the user taps "Check again". Nothing tells the user that a card is waiting, so learning steps get missed and the short intervals stop doing their job.

## What Changes

- The deck list refreshes its due counts on its own when they change with time: when a learning card becomes due, and at the day rollover. It also refreshes when the app comes back to the foreground.
- When learning cards become due while the user is outside the study screen, the app shows an in-app reminder (a snackbar) naming the deck, with a **Study** action that opens that deck's study session.
- In the study screen, when the session is waiting on a learning card ("Session complete" or "Nothing due right now"), the screen reacts when the card becomes due. By default it shows a prompt that learning cards are due, with a **Continue** button. A new setting, **Continue automatically when learning cards are due** (off by default), makes the session load the card without asking.
- While a card is on screen, the study screen's learning count also updates when another learning card becomes due.
- No system (OS-level) notifications: reminders appear only while the app is open. No change to scheduling, the review log or the storage format, apart from one new boolean setting.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-shell`: new requirements that the deck list's due counts stay current as time passes, and that the app reminds the user when learning cards become due; the App settings requirement gains the auto-continue setting.
- `study-session`: Session completion changes so a waiting session reacts when learning cards become due (prompt by default, auto-continue by setting), and session counts stay current while a card is shown.

## Impact

- Go API: `DecksService` gains a call that returns when due counts will next change on their own (the earliest learning due time after now, or the next day start). `core.Settings` and `AppService` gain the auto-continue setting. `fluffy` regenerates `lib/golang/api.g.dart`.
- Flutter: `lib/data/deck_list_model.dart` (refresh timer, foreground refresh, detecting newly due learning cards), `lib/data/settings_model.dart`, `lib/features/settings/settings_screen.dart`, `lib/features/study/study_screen.dart`, `lib/features/decks/deck_list_screen.dart` and `lib/app/scaffold.dart` (opening a deck's study session from the reminder).
- Tests: `golang/core` queue tests, `golang/svc_test.go`, `test/deck_list_test.dart`, `test/study_test.dart`, `test/settings_test.dart`, and `test/support/fake_go.dart` for the new calls.
- Overlap: the in-flight `card-actions-and-wording` change modifies the app-shell "Deck list home screen" requirement; this change adds new requirements instead of modifying that one, so the two can be archived in either order.
