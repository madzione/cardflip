# Proposal

## Why

Bugs, interrupted imports from older versions, or restored backups from a buggy build can leave the collection inconsistent: cards whose note is gone, notes with the wrong number of fields, cards in decks that no longer exist, impossible scheduling values. Anki's Check Database finds and repairs such problems. Cardflip has an internal storage self-test for integration tests (`AppService.CheckStorage`), but nothing users can run.

## What Changes

- **Check database** in Settings checks the collection and repairs what it can, then reports what it found and fixed:
  - cards whose note or note type no longer exists are deleted;
  - notes whose field count differs from their note type are padded or trimmed (trimmed content is kept in the deletion log if that exists, otherwise reported);
  - cards for templates or cloze numbers that no longer exist are deleted; missing cards are generated;
  - cards in a missing deck are moved to `Default` (created if needed);
  - invalid scheduling values (negative intervals, due dates out of range, a review card without a memory state) are repaired, with memory states recomputed from history where possible;
  - duplicate or invalid tags are normalised;
  - review log entries for missing cards are kept (for statistics) and counted;
  - storage is compacted afterwards.
- The check is one operation: it is saved completely or not at all, and it runs with progress without freezing the app.
- The report lists each kind of problem with a count, or says that no problems were found.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `persistence`: new requirement for checking and repairing the collection.
- `app-shell`: new requirement for the Check database entry.

## Impact

- Go: `Collection.CheckDatabase(progress)` in a new `golang/core/check.go`, building one op of repairs and a report; `AppService.CheckDatabase` streaming progress; compaction afterwards.
- Flutter: `settings_screen.dart` (entry, progress, report dialog).
- Tests: `golang/core/check_test.go` with a corrupted fixture for each problem kind, `golang/svc_test.go`, `test/settings_test.dart`.
- Overlap: with `add-automatic-backups`, a backup is made before the check.
