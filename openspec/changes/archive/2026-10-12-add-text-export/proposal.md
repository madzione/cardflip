# Proposal

## Why

Plain text is the universal exchange format: users export vocabulary to spreadsheets, print lists, feed notes to other tools, or edit many notes in a text editor and re-import them. Anki exports notes and cards as text; Cardflip cannot export text at all.

## What Changes

- **Notes in plain text**: a tab-separated file with one note per line and the fields as stored (HTML kept), with options to add the tags, deck, note type and GUID as columns. The file starts with `#separator:tab`, `#html:true` and `#… column:N` headers, so the file can be imported again into Cardflip or Anki to update the notes.
- **Cards in plain text**: one card per line with the rendered question and answer as plain text, tab-separated, for printing or reading elsewhere.
- Scope: a deck and its subdecks (from the deck's menu), the selected notes or cards (from the browser), or the whole collection (from Settings).
- On the web the file is downloaded; on native platforms the user chooses where to save.

## Capabilities

### New Capabilities

- `data-export`: exporting notes and cards as plain text files.

### Modified Capabilities

- `card-browser`: new requirement for exporting the selection as text.

## Impact

- Go: a new `golang/exporter/text` package (TSV quoting compatible with the CSV importer and Anki, headers); `TransferService.ExportNotesText` and `ExportCardsText` with scopes.
- Flutter: an export dialog with a format choice, wired into the deck menu, the browser and Settings.
- Tests: round-trip with `golang/importer/csv` (export, re-import with update, no duplicates), `golang/svc_test.go`, widget tests.
- Overlap: `add-anki-package-export` also creates the `data-export` capability and the same entry points; if both exist they share one export dialog with a format choice. Re-importing matches notes by first field today; with `add-text-import-routing-columns`, the GUID, note type and deck columns are used too, so edits to the first field also update instead of duplicating.
