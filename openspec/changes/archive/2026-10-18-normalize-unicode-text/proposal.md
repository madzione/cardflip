# Proposal

## Why

The same visible text can be stored as different Unicode sequences: `é` as one code point or as `e` plus a combining accent, depending on the keyboard, the OS (macOS often produces decomposed text) or the source of a paste. Searches, the duplicate check and type-in comparisons then fail for text that looks identical. Anki normalises text to NFC when saving.

## What Changes

- Field contents, tags, deck names and descriptions, note type names, field names, card templates and note type CSS are normalised to Unicode NFC when they are saved (add, edit, rename, import).
- Media file names are normalised to NFC when media is added or imported, so the references in normalised fields, templates and CSS keep matching them. As in Anki.
- Search queries are normalised the same way before matching.
- Existing collections are normalised once, the first time the new version opens them, in a single journaled operation. Stored media files with non-NFC names are renamed to NFC; when the NFC name is taken by a different file, the file gets a hashed name, as import does, and the references are rewritten.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for Unicode normalisation of stored text.

## Impact

- Go: `golang.org/x/text/unicode/norm` dependency; normalisation in `core` at the validation step of every mutation that stores text, in `core/import.go`, in media naming (`core/media.go`), and in `search.Parse`; a one-time migration op recorded by a snapshot field (`normalized: true`).
- Tests: `golang/core/notes_test.go`, `golang/core/decks_test.go`, `golang/search/search_test.go`, a migration test in `golang/core`.
- WASM size: the `norm` tables add roughly 100 KB to the web build.
