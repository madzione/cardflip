# Tasks

## 1. SQLite writer

- [x] 1.1 Add `golang/exporter/sqlite` writing a database file with tables and indexes (design decision 2). Add tests that write tables of 0, 1 and 100,000 rows and read them back with `importer/apkg/sqlite`, and, in native test builds only, with `modernc.org/sqlite` running `PRAGMA integrity_check`. Verify `go test ./exporter/...` passes and `fluffy build web` still succeeds.

## 2. Package writer

- [x] 2.1 Add `golang/exporter/apkg` writing schema 11 (`col` JSON with models, decks, deck configs), notes and cards with the mapping from design decisions 3 and 4, `revlog` when scheduling is included, the media map and the zip (modern default with dummy `collection.anki2`, legacy option). Add round-trip tests through the existing importer, with and without scheduling. Verify `go test ./...` passes.
- [x] 2.2 Add deck, selected-notes and collection exports to `TransferService` with progress streaming and file naming. Add `svc_test.go` cases, including the re-import scenario. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 3. UI

- [x] 3.1 Add the export dialog with options and progress, and wire it into the deck menu (`deck_list_screen.dart`), the browser selection (`browse_screen.dart`) and Settings (collection). Save via file picker on native and download on web, as backups do. Add widget tests for each entry point. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 4.2 Manually import exported packages (deck with and without scheduling, legacy option, collection) into current Anki and into Anki 2.1.66, and confirm notes, media, scheduling and history.
