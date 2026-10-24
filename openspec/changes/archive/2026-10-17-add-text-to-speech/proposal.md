# Proposal

## Why

Language decks use Anki's text-to-speech tags to read words aloud without shipping audio files: `{{tts en_US:Front}}` in templates and `[anki:tts lang=…]…[/anki:tts]` blocks. Cardflip renders `tts` as nothing, so these cards are silent and the listening practice is lost.

## What Changes

- `{{tts <lang>:Field}}` speaks the field's text (HTML stripped) with the device's speech engine in the given language. Options `voices=A,B` (first available voice wins) and `speed=1.2` are honoured where the engine supports them.
- `{{tts <lang>:cloze:Field}}` speaks the cloze field with the active deletion read as "blank" on the question, and `{{tts <lang>:cloze-only:Field}}` speaks only the deleted text.
- `[anki:tts lang=xx_YY voices=… speed=…]text[/anki:tts]` blocks in templates or fields are spoken the same way.
- TTS is spoken automatically when the question or answer is shown, in order of appearance, and a replay button appears where each tag is. R or F5 replays the current side.
- `{{tts-voices:}}` lists the voices available on the device, so users can pick names for `voices=`.
- When no voice exists for the language, the replay button is shown disabled with an explanation, and studying continues.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirements for TTS tags and the voice list.
- `study-session`: new requirement for speaking and replaying TTS during study.

## Impact

- Dependency: `flutter_tts` (Android, iOS, macOS, Windows, web). Linux is not covered by that plugin; Linux uses speech-dispatcher (`spd-say`) through a small platform adapter.
- Go: `golang/render` parses TTS tags into placeholders (`<cf-tts idx=N>`) plus a list of utterances per side (`render.Card.QuestionTTS`, `AnswerTTS`); `StudyCard` carries them.
- Flutter: a `lib/widgets/speech.dart` adapter, `lib/widgets/card_view.dart` (replay buttons), `lib/features/study/study_screen.dart` (autoplay, R/F5).
- Tests: `golang/render/tts_test.go`, `test/study_test.dart` with a fake speech adapter.
- Overlap: `add-media-playback` adds sound tags, autoplay and R/F5 replay for audio. If both are implemented, sounds and TTS share one playback queue per side, in document order; whichever change lands second wires its items into the other's queue.
