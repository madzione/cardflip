# Design

## Context

`AdaptiveScaffold` (`lib/app/scaffold.dart`) shows the four destinations in an `IndexedStack`, each with its own `Navigator`. Hidden destinations stay mounted, so a `StudyScreen` or `CardDetailsPane` on a hidden destination is neither disposed nor told that it is hidden. `IndexedStack` wraps its children in `Visibility.maintain`, which also keeps tickers running, so `TickerMode` cannot serve as the signal either.

Each of those screens owns a `MediaPlayer` (`lib/widgets/media_player.dart`). The study screen already stops its player when it pushes another route (edit, card info), but nothing stops it when the destination changes. The index changes in three places: `_select` (the bar or rail), and `_showDecks` and `_study` (the learning due reminder).

## Goals / Non-Goals

**Goals:**
- Every destination change stops all playback, wherever it was started.

**Non-Goals:**
- Pausing and resuming where it left off. Stopping matches what the study screen does when it opens the editor or card info.
- Stopping a voice recording in progress. Recording is a separate path (`VoiceRecorder`) and is not playback; if it needs the same treatment, that is a separate change.
- Stopping playback when the app goes to the background (app lifecycle).

## Decisions

1. **`MediaPlayer` keeps a static set of live players, and `MediaPlayer.stopAll()` stops each one.** A player adds itself in its constructor and removes itself in `dispose`. Only one destination is visible at a time, so "stop everything" is exactly "stop what the user can no longer see", and it automatically covers any future owner of a `MediaPlayer`.
   - Alternative: an inherited "this destination is active" notifier above each `Navigator`, which every screen with a player listens to. It is more precise but needs code in every owner, and a new owner that forgets to listen brings the bug back.
2. **The scaffold changes the index in one helper that calls `MediaPlayer.stopAll()` when the index really changes.** `_select`, `_showDecks` and `_study` all go through it. Re-selecting the current destination keeps its pop-to-first behaviour; the popped study screen stops its playback in `dispose`.

## Risks / Trade-offs

- [A static set outlives tests and leaks players that are never disposed] → players remove themselves in `dispose`; `media_player_test.dart` checks that a disposed player is no longer stopped by `stopAll`.
- [Stopping also clears the queue, so a side's remaining sounds do not play after returning] → intended; the replay controls play the whole side again.
