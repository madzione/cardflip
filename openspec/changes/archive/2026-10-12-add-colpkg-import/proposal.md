# Proposal

## Why

The simplest way to move from Anki is to export the whole collection as a `.colpkg` and open it elsewhere. Cardflip only reads `.apkg` deck packages, so users have to export each top-level deck separately, and collection-wide items such as empty decks and all note types do not come along in one step.

## What Changes

- `.colpkg` files (Anki collection packages, modern and legacy formats) can be chosen in Import.
- Importing a `.colpkg` **replaces** the current collection after a confirmation warning that current data will be overwritten (like restoring a backup), and suggests exporting a backup first.
- Everything Cardflip supports is taken over: decks (including empty ones), note types, notes, tags, cards with scheduling, and media. Cardflip's own app settings (theme, rollover hour and similar) are kept.
- Anki data Cardflip has no use for (sync state, add-on config) is ignored; filtered decks are converted to their home decks.
- If the import fails, the current collection is left unchanged.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-import`: "Choose an import file" includes `.colpkg`; new requirement for replacing the collection from a collection package.

## Impact

- Go: `golang/importer/apkg` already reads the collection files and media maps; `TransferService.Inspect` recognises `.colpkg`; a new path builds a fresh collection from the package (via `core.Import` into an empty collection) and swaps it in through the same mechanism as backup restore (`Collection.Restore`), keeping `Settings`.
- Flutter: `lib/features/import/import_screen.dart` and `pick.dart` (file type, replace confirmation).
- Tests: `golang/importer/apkg/apkg_test.go` (a `.colpkg` fixture), `golang/svc_test.go` (replace, failure keeps data), `test/import_test.dart`.
- Overlap: `import-apkg-review-history` adds review history to package imports; the same reader path gives `.colpkg` imports history once both exist. `add-automatic-backups` makes a backup before replacing.
