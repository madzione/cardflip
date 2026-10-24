# Proposal

## Why

When a video is playing during study and the user leaves the session with the back button, the app crashes. It does not crash when the user stops the video first. A crash can lose the user's place and looks like data loss, so it has to be fixed before more playback features (text to speech, auto advance) build on the same player.

## What Changes

- Leaving the study screen while a video or sound is playing stops playback and closes the screen normally, the same as when playback was stopped first.
- The same applies wherever a card's media player is closed while it plays: the browser's card preview when it is closed or shows another card.
- No visible behaviour changes otherwise.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `media-playback`: new requirement that closing a screen while its audio or video plays stops playback without crashing.

## Impact

- Flutter: `lib/widgets/media_player.dart` (the order in which `MediaPlayer` and `MediaKitBackend` shut down), possibly `lib/features/study/study_screen.dart` (stop playback as soon as the route starts to pop).
- Tests: `test/media_player_test.dart`, `test/study_test.dart`, `test/support/fake_media.dart` (the fake backend implements and logs `close`).
- Not affected: Go, the collection.
- Overlap: `stop-media-on-tab-switch` adds a registry of live players to the same class; whichever lands second keeps the registry's removal in `dispose`.
