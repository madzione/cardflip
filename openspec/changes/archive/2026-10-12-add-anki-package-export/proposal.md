# Proposal

## Why

Cardflip can only export its own `.cardflip` backup. Users cannot share a deck with friends or classmates, cannot publish a deck, and cannot go back to Anki without losing their work. Anki deck packages (`.apkg`) and collection packages (`.colpkg`) are the common exchange format for spaced-repetition decks.

## What Changes

- **Export deck** (`.apkg`) from a deck's menu on the deck list: the deck and its subdecks, with options:
  - **Include scheduling information** (off by default; when off, cards are exported as new and without history);
  - **Include deck presets** (the deck options);
  - **Include media** (on by default);
  - **Support older Anki versions** (writes the older collection format, larger and slower to import).
- **Export collection** (`.colpkg`) from Settings: the whole collection with scheduling and history, with or without media, modern or legacy format.
- **Export selected notes** from the browser as `.apkg` (their cards in their decks).
- Packages open in current Anki and, with "support older Anki versions", in Anki 2.1.x. Re-importing an exported package into Cardflip updates notes instead of duplicating them (GUIDs are kept).
- Cardflip's FSRS memory states are exported in the card data Anki reads, so scheduling carries over.
- On the web, the file is downloaded; on native platforms, the user chooses where to save.

## Capabilities

### New Capabilities

- `data-export`: exporting decks, selected notes and the whole collection in formats other apps can read.

### Modified Capabilities

- `app-shell`: new requirements for the export entries on the deck list and in Settings.
- `card-browser`: new requirement for exporting selected notes.

## Impact

- Go: a new `golang/exporter/apkg` package with a minimal pure-Go SQLite writer (tables and pages only, no SQL engine), Anki schema 11 (`col` JSON for models, decks and deck configs) for both modern and legacy packages, media map, zip; mapping from Cardflip cards to Anki `type`/`queue`/`due`/`ivl`/`factor`/`data` (FSRS state) and the review log to `revlog`; `TransferService.ExportDeck`, `ExportNotes`, `ExportCollection` streaming progress.
- Flutter: export dialogs in `deck_list_screen.dart`, `browse_screen.dart`, `settings_screen.dart`; saving via `file_picker` (native) or download (web), as backups do today.
- Tests: round-trip tests (export with the new writer, import with the existing reader), validation of the SQLite files with a reference reader in tests, and a manual import into Anki.
