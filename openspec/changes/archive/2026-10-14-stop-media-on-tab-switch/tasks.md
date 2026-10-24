# Tasks

## 1. Stop every player

- [x] 1.1 In `lib/widgets/media_player.dart`, register each `MediaPlayer` in a static set on construction, remove it in `dispose`, and add `MediaPlayer.stopAll()`. Add `test/media_player_test.dart` cases: two players playing are both stopped by `stopAll` (the fake backend logs `stop` for each), and a disposed player is not touched. Verify `fvm flutter test test/media_player_test.dart` passes.

## 2. Stop on destination change

- [x] 2.1 In `lib/app/scaffold.dart`, route every index change (`_select`, `_showDecks`, `_study`) through one helper that calls `MediaPlayer.stopAll()` when the index changes. Add `test/app_shell_test.dart` widget tests with `useFakeMedia`: a sound playing in study stops when Browse is selected; a sound playing in the browser's card preview stops when Decks is selected; returning to Decks shows the same card with nothing playing, and R plays it again. Verify `fvm flutter test test/app_shell_test.dart` passes.

## 3. Integration check

- [x] 3.1 Run `fvm flutter analyze` and `fvm flutter test`, and verify both pass. On Linux, start a card with audio, switch to Browse and back, and confirm the audio stopped and R replays it.
