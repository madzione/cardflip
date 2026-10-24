# Proposal

## Why

Imported decks and years of manual entry accumulate duplicate notes, which waste reviews and confuse learning. Anki's Find Duplicates lists notes with the same content in a chosen field, optionally within a search, and tags them for cleanup. Cardflip has no way to find duplicates.

## What Changes

- A **Find duplicates** action in the browser: choose a field (from all note types that have a field of that name), optionally restrict with a search, and run.
- Results list each set of duplicates: the shared content (HTML stripped, case-insensitive, surrounding whitespace ignored) and the number of notes; selecting a set shows those notes in the browser.
- **Tag duplicates** adds the tag `duplicate` to every note in the results, so they can be reviewed with `tag:duplicate`.
- Empty fields are not reported as duplicates.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirement for finding and tagging duplicates.

## Impact

- Go: `Collection.FindDuplicates(fieldName, query)` grouping notes by normalised field content; `BrowseService` calls.
- Flutter: `lib/features/browse/find_duplicates.dart`, a toolbar action in `browse_screen.dart`.
- Tests: `golang/core/notes_test.go`, `golang/svc_test.go`, `test/browse_test.dart`.
