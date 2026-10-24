# Proposal

## Why

When a shared deck is updated (fixed typos, new fields), users re-import it. Cardflip always overwrites the content of every matching note, which also overwrites the user's own edits to older notes, and how it handles a note type that changed on both sides is not defined. Anki lets the user choose: update only if the package's version is newer (default), always, or never, separately for notes and note types, and can merge note types.

## What Changes

- Import options for `.apkg`: **Update notes** and **Update note types**, each *if newer* (default), *always* or *never*.
- *If newer* compares modification times: a note is updated only when the package's note was modified later than the collection's; likewise for note types.
- **Merge note types** (off by default): when an existing note type and the package's version differ in fields or templates, fields and templates are matched by name, new ones are added, and none are removed; notes are updated against the merged type. When off, a differing note type is imported as a new note type, and notes of that type in the package use it.
- Scheduling of existing cards stays unchanged in all cases, as today.
- The summary reports notes updated, notes skipped because the collection's version is newer, and note types updated, merged or added.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-import`: "Re-importing a package" follows the update policy; new requirement for note type handling.

## Impact

- Go: `core.NoteType` gains `Modified` (set on every note type change; absent treated as oldest); the apkg reader reads note and note type modification times; `core/import.go` applies the policies and the merge; `ImportOptions` gains the three options; the summary gains counters.
- Flutter: `import_screen.dart` (options in the package import step).
- Tests: `golang/core` import tests, `golang/importer/apkg/apkg_test.go` (fixtures with a newer and an older version of a deck), `golang/svc_test.go`, `test/import_test.dart`.
