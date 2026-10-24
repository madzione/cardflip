# Design

## Context

`StudyScreen.dispose` calls `MediaPlayer.dispose` (`lib/widgets/media_player.dart`), which runs:

1. `stop()`: because something is playing, it calls `_backend?.stop()` **without awaiting it**. `MediaKitBackend.stop` completes the pending play and calls `Player.stop()` (asynchronous, native).
2. `_backend?.dispose()`: `MediaKitBackend.dispose` cancels its subscriptions and calls `Player.dispose()`, again **without awaiting it**.

So `Player.stop()` and `Player.dispose()` run at the same time on libmpv. With a video, a `VideoController` is attached to the player and its native video output (texture and render context) is released by `Player.dispose()` while the stop is still in flight and the last frame may still be rendering.

When the user stops the video first, step 1 does nothing (`_current` is null), the player is already idle, and `dispose` alone does not crash. This matches the report and points to the overlap of stop and dispose. Task 1.1 confirmed it on Linux: a SIGSEGV on Flutter's raster thread inside the engine while it paints the video texture. The ordered shutdown (decision 1) on its own does **not** stop the crash: the texture is still in the frame being painted when the player is freed, however the stop and dispose are ordered. Removing the `Video` widget before the player is freed (decision 2) is what fixes it; decision 1 still removes the stop/dispose overlap.

`CardDetailsPane` (`lib/features/browse/card_details.dart`) owns a `MediaPlayer` the same way and has the same shutdown path. On Linux, pressing back from `CardDetailsScreen` (the browser's card preview as a route on narrow layouts) while its video plays crashes the same way, even with decision 1. Removing the wide-layout details pane in place (a search that drops its card) while its video plays does not crash: the `Video` widget leaves the tree without an exit transition.

## Goals / Non-Goals

**Goals:**
- Shutting a player down while it plays is safe on every platform, for audio and video.

**Non-Goals:**
- Changing how or when media plays.
- Upgrading media_kit. If the crash turns out to be a media_kit bug that ordering cannot avoid, that becomes a separate decision.

## Decisions

1. **The backend owns its shutdown order: `PlaybackBackend.dispose()` becomes `Future<void> close()`.** `MediaKitBackend.close` awaits `Player.stop()` and then `Player.dispose()`. `MediaPlayer.dispose` no longer calls the backend's `stop` itself: it resets its own state and starts `close()` without awaiting it, because `State.dispose` cannot be asynchronous. The backend is not used again after `close`, so a close that finishes after the screen is gone is fine. With the fake backend, tests can check that a disposing player calls only `close`, never `stop` and `close` side by side.
   - Alternative: only skip `stop()` in `MediaPlayer.dispose` and call `Player.dispose()` directly. Simpler, but it relies on media_kit handling dispose during playback, which is the suspected crash.
2. **The study screen stops playback when the pop starts, not when the route is disposed.** A `PopScope` with `onPopInvokedWithResult` calls `_player.stop()`, so the inline video is replaced by its play button during the pop transition and the `Video` widget is gone well before the player is disposed. This also ends the sound at the moment the user presses back, rather than after the transition.
   - Alternative: rely on decision 1 alone. Task 1.1 showed that it does not stop the crash: the `Video` widget is still painted during the exit transition, until the route and its player are disposed.
   - The browser's card preview does the same: `CardDetailsPane` stops its player in its own `PopScope`. In `CardDetailsScreen` that runs when the preview route pops; in the wide layout the pane is never popped and the scope does nothing.
3. **If task 1.1 shows a different cause**, update this design and the tasks before fixing; the spec stays the same.

## Risks / Trade-offs

- [The fake backend cannot reproduce a native crash, so widget tests only check the order of calls] → task 1.1 and the integration check reproduce it on a real device before and after the fix.
- [A close that never finishes (libmpv hangs on stop) leaks one player] → the close times out after a few seconds and disposes anyway.
