# Design

## Context

- Sanitization in `golang/render/sanitize.go` removes `[sound:]`, `<audio>` and `<video>`; the importer drops audio/video files with a warning.
- Media bytes are read through `MediaService.Get(name)`; on the web they come from IndexedDB through Go/WASM, on native from files under the app data directory (`storage/backend_files.go`).
- The study screen owns keyboard shortcuts and per-card state.

## Goals / Non-Goals

**Goals:** identical behaviour on all six platforms; no network; one player queue per side.

**Non-Goals:** waveform display, playback speed control, storing recordings, audio in the browser list.

## Decisions

### 1. `media_kit` for audio and video

It covers Linux, Windows, macOS, Android, iOS and web with one API and plays audio and video. `just_audio` lacks Linux/Windows without extra federated packages and does not do video; `audioplayers` does not do video.

### 2. Go emits placeholders and an ordered media list

The renderer replaces each sound reference with `<cf-sound idx=N>` (or `<cf-video idx=N>` for video extensions and `<video>`), and returns `[]MediaRef{File, Kind}` for each side. The study screen plays the list for autoplay; the card view maps placeholders to buttons or players. Keeping order in Go matches the document order of the rendered template.

### 3. Feeding media to the player

Native: media is materialised once into a cache directory (path from `path_provider`) and played by file path, since the files backend already stores media as files where possible; the cache is cleared at startup. Web: bytes are wrapped in a `Blob` URL that is revoked when the card changes.

### 4. Autoplay rules

On showing a side: if the deck option "Don't play audio automatically" is off, play its list in order. On reveal: stop the question unless "Interrupt current audio when answering" is off, in which case the answer's list is queued after it. Replay (R/F5) on the answer side plays the question list first unless "Skip question when replaying answer" is on — matching Anki's behaviour.

### 5. Record own voice

`record` writes to a temporary file (native) or memory blob (web); one recording per card is kept until the card changes. Recording requires permission; if refused, an explanation is shown and nothing else changes.

### 6. Formats

Accepted extensions are those `media_kit` plays on every platform. On web, codec support depends on the browser (for example, OGG in Safari); a failed load shows a disabled button with an explanation instead of an error dialog.

## Risks / Trade-offs

- [App size grows with `media_kit` native libs] → Use the audio-only libs where video is not needed is not possible since video is in scope; document the size increase in the README.
- [Large media in IndexedDB on the web] → Imports already show progress; storage errors abort the import atomically as today.
- [Autoplay blocked by browsers before user interaction] → The study session starts from a click, which unlocks audio; if playback is still refused, the replay buttons remain.
