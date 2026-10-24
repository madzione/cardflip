# Proposal

## Why

Audio or video that is playing keeps playing when the user switches to another destination in the navigation bar or rail (Decks, Browse, Add, Settings). The study screen and the browser's card preview stay alive in the background, because each destination keeps its own screens, so their players never learn that they are no longer visible. The user hears a card they can no longer see and has to go back to stop it.

## What Changes

- Switching to another top-level destination stops all audio and video playback, in the study screen and in the browser's card preview. This also covers switches the app makes itself, such as the learning due reminder's Study and Show decks actions.
- Coming back to the destination does not resume or replay anything. The card stays where it was, and the replay controls (R, F5, the play buttons) play it again on request.
- Tapping the destination that is already selected keeps its current behaviour: it goes back to the destination's first screen, which closes a study session and stops its playback.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `media-playback`: new requirement that playback stops when the user switches to another destination.

## Impact

- Flutter: `lib/widgets/media_player.dart` (a way to stop every live player), `lib/app/scaffold.dart` (stop playback whenever the selected destination changes).
- Tests: `test/media_player_test.dart`, `test/app_shell_test.dart`.
- Not affected: Go, the collection, the study session's state.
- Overlap: `add-text-to-speech` and `add-auto-advance` add more kinds of playback to the same player; whichever lands second must make sure its playback stops too. `fix-video-crash-on-leaving-study` changes how the player shuts down; the two changes touch different methods.
