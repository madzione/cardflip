# Tasks

## 1. Queue

- [x] 1.1 Add `LearnAheadMinutes *int` to `core.Settings` (nil = 20) with `AppService.SetLearnAheadMinutes` (0–1440) and the DTO field. Add `svc_test.go` cases including the nil default. Verify `go test ./...` passes.
- [x] 1.2 Add `LearningAhead []ID` (learning cards due after now and within the limit, earliest first) to `core.Queue`, use it in `PickNext` only when the other queues are empty, and make `LearningLater` the first learning card beyond the limit. Add `queue_test.go` cases for both learn-ahead scenarios. Verify `go test ./core/...` passes.
- [x] 1.3 Make `StudyState.LearningDue` report the moment the next learning card comes within the limit (its due time minus the limit). Add `svc_test.go` cases for "Reacting at the limit". Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 2. Flutter

- [x] 2.1 Add the setting to `settings_screen.dart` and `SettingsModel`. Add a `test/settings_test.dart` test. Verify it passes.
- [x] 2.2 Update `test/study_test.dart` cases that rely on waiting for learning cards to set the limit to 0 in the fake, and add tests for "Finish without waiting" and "Reacting at the limit". Verify the suite passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
