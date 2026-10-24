# Tasks

## 1. Go

- [x] 1.1 Add the three timer options to `core.DeckOptions` and the DTO (absent = defaults, validated range) and `TimeboxMinutes` to `core.Settings` with a setter. Add `decks_test.go` and `svc_test.go` cases. Verify `go test ./...` passes.
- [x] 1.2 Cap the duration in `Collection.Answer` with the card's deck option. Add a `reviews_test.go` case for the long-pause scenario. Verify `go test ./core/...` passes.
- [x] 1.3 Expose the deck's timer options in `StudyState`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. Flutter

- [x] 2.1 Add the options to `deck_options_dialog.dart` and the timebox setting to `settings_screen.dart`, with widget tests. Verify they pass.
- [x] 2.2 Add the timer widget to `study_screen.dart` (colour at maximum, freeze on answer). Add `test/study_test.dart` tests using `tester.pump(duration)`. Verify they pass.
- [x] 2.3 Add the timebox message with Continue and Finish. Add a test for the ten-minute scenario. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
