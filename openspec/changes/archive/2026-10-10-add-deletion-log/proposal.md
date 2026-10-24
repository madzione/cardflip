# Proposal

## Why

Deleted notes in Cardflip are gone for good once the action is confirmed. Anki writes every deleted note to a deletion log (`deleted.txt`) in a format that can be imported again, which has saved many users from accidental deletions discovered days later.

## What Changes

- Every deleted note (from the browser, from study, by deleting a deck or a note type, or as part of other operations) is appended to a deletion log: deletion time, note ID, note type name, fields and tags.
- Settings has **Export deletion log**, which saves or downloads the log as a tab-separated text file with headers that the text importer understands (fields first, then tags), and **Clear deletion log**.
- The log is part of the collection storage and is included in backups.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for logging deleted notes and exporting or clearing the log.

## Impact

- Go: the deletion op in `golang/core` appends entries to a deletion log stored through `storage` (append-only, like review log segments); `AppService.ExportDeletionLog` and `ClearDeletionLog`; backup and restore include it.
- Flutter: `settings_screen.dart`.
- Tests: `golang/core` deletion tests, `golang/backup/backup_test.go`, `golang/svc_test.go`, `test/settings_test.dart`.
