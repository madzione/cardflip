# Tasks

## 1. Settings

- [x] 1.1 Add the five settings to `core.Settings` (absent = defaults) with setters and DTO fields. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 1.2 Add the Review section to `settings_screen.dart` and `SettingsModel`. Add `test/settings_test.dart` tests. Verify they pass.

## 2. Study screen

- [x] 2.1 Honour the count, interval label and Space/Enter settings in `study_screen.dart`. Add `test/study_test.dart` tests for each modified scenario. Verify they pass.
- [x] 2.2 Hide navigation and app bar while studying when the setting is on (coordinated with `lib/app/scaffold.dart`), keeping Escape and a top-edge leave control. Add a test. Verify it passes.
- [x] 2.3 Add the question fade and the scroll-to-answer behaviour (position of `id=answer` reported by `card_view.dart`). Add tests for fade 0 and for scrolling on a tall card. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
