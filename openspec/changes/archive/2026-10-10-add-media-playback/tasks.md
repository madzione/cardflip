# Tasks

## 1. Rendering and import in Go

- [x] 1.1 Stop stripping sound references in `golang/render/sanitize.go`; emit `<cf-sound idx=N>` / `<cf-video idx=N>` and per-side media lists on `render.Card`; keep `{{FrontSide}}` free of the question's sounds. Add tests for `[sound:]`, `<audio>`, `<video>`, `[sound:clip.mp4]` and FrontSide. Verify `go test ./render/...` passes.
- [x] 1.2 Import audio and video in `importer/apkg/convert.go` (keep a warning only for extensions outside the supported list) and update the batch comment in `core/import.go`. Update `apkg_test.go` expectations (audio imported, summary has no audio warning). Verify `go test ./importer/... ./core/...` passes.
- [x] 1.3 Add the two audio options to `core.DeckOptions` (`omitempty`) and the `DeckOptions` DTO, and the two preferences to `core.Settings` and the settings DTO with setters. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`, and verify `fluffy generate --check` is clean and `fvm flutter test` passes.

## 2. Player

- [x] 2.1 Add `media_kit` and its platform packages; add `lib/widgets/media_player.dart` with a per-side queue (play list, stop, pause/resume, seek ±5 s) and media materialisation per design decision 3. Unit-test the queue ordering and the replay rules from design decision 4 with a fake backend. Verify tests pass and `fluffy build linux`, `fluffy build web` and `fluffy build apk` succeed.
- [x] 2.2 Map `cf-sound` to a replay button and `cf-video` to an inline player in `card_view.dart`, with disabled states for missing or unplayable files, and hide buttons when the preference is off. Add `test/card_view_test.dart` tests. Verify they pass.

## 3. Study and options UI

- [x] 3.1 In `study_screen.dart`, autoplay per side, stop on reveal and leave per the interrupt preference, and add R/F5, 5, 6, 7. Add `test/study_test.dart` tests with the fake backend for autoplay, autoplay off, replay order and pause. Verify they pass.
- [x] 3.2 Add the two deck options to `deck_options_dialog.dart` and the two preferences to `settings_screen.dart`, with widget tests that they call the setters. Verify they pass.

## 4. Record own voice

- [x] 4.1 Add `record`, microphone permission entries for Android, iOS and macOS, and Shift+V / V plus a menu item in the study screen per design decision 5. Add a `test/study_test.dart` test with a fake recorder that the recording is dropped on the next card, and one for a refused permission message. Verify they pass.

## 5. Integration check

- [x] 5.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 5.2 Manually import an audio deck and a deck with an MP4 clip; study on Linux, Windows or macOS, Android and Chrome; check autoplay, replay, pause, seek and recording.
