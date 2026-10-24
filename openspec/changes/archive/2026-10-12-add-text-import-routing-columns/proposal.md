# Proposal

## Why

A text file exported by Anki with "include deck name", "include note type" and "include unique identifier" carries, per row, the note type, the deck, the tags and the note's GUID. Cardflip imports every row into one chosen note type and deck, cannot read those columns, and matches existing notes only by first field, so re-importing an edited export creates duplicates or skips changed notes instead of updating them.

## What Changes

- Headers `#notetype column:N`, `#deck column:N`, `#tags column:N` and `#guid column:N` mark columns that hold each row's note type name, deck name, tags and GUID. Headers `#notetype:Name` and `#deck:Name` set a default for rows without a column value.
- In the import screen, any column can be mapped to Note type, Deck, Tags or GUID as well as to a field.
- Rows are imported into their own note type (by name; an unknown name is reported and the row skipped) and deck (created if missing). Field columns map by position within each row's note type.
- GUID matching: a row whose GUID matches an existing note updates that note (when its note type matches), regardless of the duplicate setting; rows without a GUID use the duplicate setting as today.
- Duplicate scope option: match the first field within the note type (today) or within the note type and deck.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-import`: "CSV duplicate handling" gains GUID matching and the scope option; new requirement for per-row note type, deck, tags and GUID columns.

## Impact

- Go: `golang/importer/csv` (headers, per-row routing in `Build` into `core.ImportNote` with `TypeKey`, `DeckName`, `GUID`); `core.ImportBatch` already supports per-note decks and GUIDs; duplicate scope in `core/import.go`; `ImportOptions` gains column roles and the scope.
- Flutter: `import_screen.dart` (mapping choices per column, scope option, warnings for unknown note types).
- Tests: `golang/importer/csv/csv_test.go`, `golang/core` import tests, `golang/svc_test.go`, `test/import_test.dart`.
- Overlap: `extend-text-import-format` changes the same header parser.
