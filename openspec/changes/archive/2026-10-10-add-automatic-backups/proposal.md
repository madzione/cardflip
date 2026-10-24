# Proposal

## Why

Cardflip only has manual backups, and each one asks where to save. Users who never export a backup lose everything when storage is cleared (a real risk on the web) or after a mistaken bulk change. Anki backs up automatically every 30 minutes and before risky operations, keeps a thinned set of daily, weekly and monthly backups, and restores one from a list.

## What Changes

- **Automatic backups** of the collection (without media) every 30 minutes while the app is open and the collection has changed since the last backup, and before restoring a backup, importing a collection package, and checking the database.
- **Thinning**: keep the backups of today, plus one per day for the last 8 days, one per week for the last 8 weeks and one per month for the last 8 months (all configurable); older ones are deleted.
- **Restore from automatic backup**: Settings lists automatic backups with date and time; restoring one works like restoring a backup file (confirmation, all or nothing). Media stay as they are.
- **Create backup now** (Settings, and Ctrl+Alt+B / Cmd+Option+B) makes an automatic-style backup immediately.
- Settings: backup interval in minutes (0 turns periodic backups off; backups before risky operations always happen), and the daily, weekly and monthly counts.
- Storage: on native platforms, a `backups` folder in the app data directory; on the web, browser storage next to the collection. Settings shows the space used.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `backup`: new requirements for automatic backups, thinning, restoring from the list and creating a backup now.
- `app-shell`: new requirement for the backup settings.

## Impact

- Go: `golang/backup` gains a backup store (list, write, delete, thin) on top of `storage.Backend` (files backend: a directory; IndexedDB backend: a separate object store); the app runs the periodic timer and the pre-operation hook; `AppService` gains backup calls; `core.Settings` gains the backup settings.
- Flutter: `settings_screen.dart` (list, restore, create now, settings, space used).
- Tests: `golang/backup/backup_test.go`, `golang/storage` contract tests for the backup store, `golang/svc_test.go`, `test/settings_test.dart`.
- Overlap: `add-colpkg-import` and `add-check-database` add risky operations; each calls the pre-operation backup when this change exists.
