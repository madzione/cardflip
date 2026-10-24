# Proposal

## Why

Audio is central to language decks: pronunciations as `[sound:word.mp3]`, listening exercises, and many decks with video clips. Cardflip strips all audio and video and skips those files on `.apkg` import, so such decks lose their main content, and a later re-import is needed to get the files back.

## What Changes

- `[sound:file]` tags, `<audio>` and `<video>` elements referring to media files play in the study screen. Audio shows a replay button where the tag is; video shows an inline player.
- Sounds of a side play automatically in order when it is shown; showing the answer or leaving the card stops the question's sounds.
- Keys: R or F5 replays the current side, 5 pauses or resumes, 6 and 7 seek back and forward 5 seconds.
- Deck options: "Don't play audio automatically" and "Skip question when replaying answer".
- Settings: "Show play buttons on cards with audio" and "Interrupt current audio when answering".
- Record own voice: during study, Shift+V (or a menu item) records from the microphone until stopped, and V plays the recording back. Recordings are temporary and never stored in the collection.
- `.apkg` import imports audio and video media (MP3, OGG, Opus, M4A/AAC, WAV, FLAC, WEBM, MP4); the "audio skipped" warning goes away, and only unsupported formats are reported.
- Browser card details show the same replay buttons (no autoplay).

## Capabilities

### New Capabilities

- `media-playback`: playing audio and video during study, the replay controls, the audio options and recording one's own voice.

### Modified Capabilities

- `card-rendering`: "Unsupported media degrades gracefully" is replaced by "Unplayable media degrades gracefully", which covers only media the device cannot play; new requirement that sound references become controls.
- `data-import`: "Anki package import" includes audio and video media; "Import progress and summary" warns only about unsupported formats.
- `collection`: new requirement for the two audio deck options.

## Impact

- Dependencies: `media_kit` (audio and video on all six platforms, including Linux and web) with its platform libs; `record` for the microphone.
- Platform setup: microphone permission strings (Android manifest, iOS/macOS `Info.plist` and entitlements), web requires HTTPS or localhost for the microphone.
- Go: `golang/render/sanitize.go` stops stripping sound tags, `<audio>` and `<video>`; the renderer emits `<cf-sound idx=N>` and `<cf-video idx=N>` and a per-side media list; `{{FrontSide}}` keeps excluding the question's audio; `importer/apkg/convert.go` stops filtering audio/video; `core.DeckOptions` gains the two audio options; `core.Settings` gains the two preferences.
- Flutter: new `lib/widgets/media_player.dart` (playback queue, seek, pause), `card_view.dart` (buttons, inline video), `study_screen.dart` (autoplay, keys, record own voice), `deck_options_dialog.dart`, `settings_screen.dart`.
- Storage: media files can be large; on the web they live in IndexedDB, so imports of audio-heavy decks use more browser storage (the persistence notice already covers eviction).
- Tests: `golang/render/*_test.go`, `golang/importer/apkg/apkg_test.go`, `golang/svc_test.go`, `test/study_test.dart`, `test/card_view_test.dart`, `test/import_test.dart`.
- Overlap: `add-text-to-speech` adds TTS utterances with the same autoplay and R/F5 replay; if both exist, they share one queue per side in document order.
