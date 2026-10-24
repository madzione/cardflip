# Tasks

## 1. Reproduce

- [x] 1.1 Run the app on Linux (`fvm flutter run -d linux`) with a deck that has a video card, play the video, press back, and record the crash output (native stack or Dart error). Confirm that it matches the stop/dispose overlap in design.md; if not, update design.md before going on. Verify by noting the crash output in this task.
  - Reproduced with a scratch Linux integration test (real Go core and media_kit, a 20 s H.264 clip, back pressed about 3 s into playback): the process dies with SIGSEGV (SEGV_MAPERR) on thread `io.flutter.rast`, top frames in `libflutter_linux_gtk.so` (+0x1e4e24a, then repeated layer-tree recursion +0x2152e8c/+0x215cc41), no Dart error. That is the raster thread painting the video's texture after its native video output was freed, which is design.md's overlap. With 2.1 alone it still crashes, with the same stack; with 2.1 and 3.1 three rounds of back-while-playing pass and the deck plays again. See design.md, Context.

## 2. Ordered shutdown

- [x] 2.1 In `lib/widgets/media_player.dart`, replace `PlaybackBackend.dispose()` with `Future<void> close()`; implement it in `MediaKitBackend` as await `Player.stop()`, then `Player.dispose()`, with a timeout; make `MediaPlayer.dispose` reset its state and call `close()` once, without calling `stop` on the backend. Update `FakeBackend` in `test/support/fake_media.dart` to log `close`. Add a `test/media_player_test.dart` case: disposing a player while it plays logs exactly `close` (no `stop`), and the pending play completes. Verify `fvm flutter test test/media_player_test.dart` passes.

## 3. Stop when the pop starts

- [x] 3.1 In `lib/features/study/study_screen.dart`, stop the player in a `PopScope.onPopInvokedWithResult`. Add a `test/study_test.dart` widget test with `useFakeMedia`: a video card is playing, the back button is pressed, the deck list is shown, the fake logs `stop` before `close`, and opening the deck again plays the video. Verify `fvm flutter test test/study_test.dart` passes.
- [x] 3.2 In `lib/features/browse/card_details.dart`, stop the pane's player in a `PopScope.onPopInvokedWithResult` in `CardDetailsPane`. Add a `test/browse_test.dart` widget test with `useFakeMedia`: on a phone layout the card preview plays a video, the back button is pressed, the fake logs `stop` when the pop starts and `close` after it, and the browser list is shown. Verify `fvm flutter test test/browse_test.dart` passes, and that pressing back from the preview while a video plays no longer crashes on Linux.

## 4. Integration check

- [x] 4.1 Run `fvm flutter analyze` and `fvm flutter test`, and verify both pass. Repeat the steps of 1.1 on Linux and on Android: press back while a video plays and while a sound plays, and confirm the app keeps running and the deck can be studied again.
