# Tasks

## 1. Go

- [x] 1.1 Add the four options to `core.DeckOptions` and the DTO with validation, and expose them in `StudyState`. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 1.2 Map the options in `.apkg`/`.colpkg` import (protobuf fields 41, 42, 36 and 43, legacy `dconf` keys) and export, with the fixture preset `Fixture Spanish` setting them (`tool/make_apkg_fixtures.py`) and `tool/check_apkg_export.py` comparing them. Verify `go test ./...`, `make_apkg_fixtures.py --check` and `check_apkg_export.py` pass.

## 2. Flutter

- [x] 2.1 Add the options to `deck_options_dialog.dart` with a widget test. Verify it passes.
- [x] 2.2 Add the session toggle (menu, Shift+A, indicator) and the countdown to `study_screen.dart`, running the configured actions and restarting on input. Add `test/study_test.dart` tests for both spec scenarios and for "show reminder". Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
