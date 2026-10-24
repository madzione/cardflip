# Tasks

## 1. Go

- [x] 1.1 Parse `{{tts …:Field}}` (language, `voices=`, `speed=`, `cloze:` and `cloze-only:` modifiers) and `[anki:tts …]…[/anki:tts]` in `golang/render`, emitting `<cf-tts idx=N>` and a per-side utterance list on `render.Card`; `{{FrontSide}}` must not repeat the question's utterances on the answer. Add `tts_test.go` with the spec scenarios, options and a block inside a field. Verify `go test ./render/...` passes.
- [x] 1.2 Render `{{tts-voices:}}` as a placeholder the card view fills with the device's voices. Add the utterance lists to `StudyCard` and `CardDetails`, run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean and `fvm flutter test` passes.

## 2. Speech adapter

- [x] 2.1 Add `flutter_tts` and `lib/widgets/speech.dart` with `voices()`, `speak(utterance)` and `stop()`; on Linux use `spd-say` (language, rate) when present and report no voices otherwise. Add a unit test for the language/voice selection logic with a fake engine. Verify it passes and `fluffy build linux` and `fluffy build web` succeed.

## 3. Study screen

- [x] 3.1 Show replay buttons for `cf-tts` in `card_view.dart` (disabled with a tooltip when no voice matches) and fill `{{tts-voices:}}`. Add `test/card_view_test.dart` tests with a fake adapter. Verify they pass.
- [x] 3.2 In `study_screen.dart`, speak a side's utterances in order when it is shown, stop on reveal and on leaving, and replay with R/F5. Add `test/study_test.dart` tests for autoplay, stop on reveal and R. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Manually check a TTS deck on Linux (with speech-dispatcher), Android and Chrome.
