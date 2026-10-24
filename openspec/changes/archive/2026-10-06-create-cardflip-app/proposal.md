# Proposal

## Why

Cardflip is a new spaced-repetition flash card trainer, similar to Anki, that runs on Linux, Windows, macOS, Android, iOS and the web from one codebase. Building it as a Flutter UI on a Go core (wired together with fluffy) lets the scheduling, storage and import logic live in plain, testable Go that behaves the same on every platform. Supporting Anki `.apkg` import from day one means existing Anki users can bring their decks with them.

## What Changes

- Create the Flutter app `cardflip` with fluffy at the repository root. It targets all six fluffy platforms, and the Flutter SDK is pinned through FVM.
- Add a Go core (`golang/`) that owns the whole collection: decks, note types, notes, cards, review history and media.
- Schedule reviews with FSRS, using default parameters and a configurable desired retention per deck. Daily new and review limits and a configurable day rollover hour apply.
- Add a study session: show the question, reveal the answer, rate it Again/Hard/Good/Easy with a preview of the next interval, and undo the last answer. Keyboard shortcuts work on desktop and web.
- Render cards from note-type templates, an Anki-compatible subset with field substitution, `{{FrontSide}}`, conditionals and cloze deletions. Cards show as HTML with embedded images.
- Add deck management (hierarchical decks, per-deck options), a note editor (Basic, Basic + Reversed and Cloze built in) and a card browser with search, edit, suspend and delete.
- Import from CSV/TSV files and from Anki `.apkg` packages: legacy `collection.anki2`/`anki21` and the current `collection.anki21b` format, including images. The user chooses whether imported cards keep their Anki scheduling or start as new.
- Back up and restore the full collection as a `.cardflip` archive (JSON plus media in a zip).
- Store everything locally and durably on every platform: files on native and IndexedDB on the web. There is no sync between devices.

## Capabilities

### New Capabilities
- `collection`: decks (hierarchy, options), note types, notes, cards, tags, suspend/delete, and how they relate.
- `scheduling`: FSRS memory states, rating outcomes, daily limits, day boundary, and which cards are due.
- `study-session`: the review flow the user sees: queue order, reveal, rating, interval previews, undo, keyboard control, session completion.
- `card-rendering`: turning a note and template into question/answer HTML (fields, FrontSide, conditionals, cloze) and displaying it with media.
- `card-browser`: searching, filtering, viewing and bulk-editing cards and notes.
- `data-import`: importing CSV/TSV files and Anki `.apkg` packages into the collection.
- `backup`: exporting the full collection to a `.cardflip` archive and restoring from one.
- `persistence`: durable local storage of the collection and media on native platforms and the web, including crash safety.
- `app-shell`: supported platforms, navigation, responsive layout, theming and first-run experience.

### Modified Capabilities
<!-- None: the project has no existing specs. -->

## Impact

- **New code**: the whole Flutter project (`lib/`, platform folders, `pubspec.yaml`) and the Go module `golang/`, plus the fluffy-generated bindings (`lib/golang/api.g.dart`, `golang/fluffy_api.go`) and the message layer that fluffy owns.
- **Go dependencies**: fluffy runtime, an FSRS implementation (`github.com/open-spaced-repetition/go-fsrs`), a pure-Go read-only SQLite reader for `.apkg`, zstd (`github.com/klauspost/compress`) and protobuf wire decoding (`google.golang.org/protobuf`) for `anki21b`. Every dependency must compile for `js/wasm`.
- **Dart dependencies**: `ffi` (added by fluffy), `path_provider`, `file_picker`, and an HTML renderer (`flutter_widget_from_html_core`).
- **Tooling**: Go 1.27, Flutter 3.47.0 via FVM, fluffy 0.1.0 (`~/go/bin` must be on `PATH` for builds), and per-platform toolchains (C compiler, Android NDK, Xcode).
- **Out of scope**: device sync and accounts, FSRS parameter optimization, audio/video playback, CSV and `.apkg` export, add-ons, rich statistics beyond today's counts, and an image-occlusion editor.
