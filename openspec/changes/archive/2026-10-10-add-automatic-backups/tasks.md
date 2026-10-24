# Tasks

## 1. Backup store

- [x] 1.1 Add a backup store to `storage.Backend` (list, write, read, delete) for the files backend (a `backups` directory) and the IndexedDB backend (an object store), with contract tests in `golang/storage/storagetest`. Verify `go test ./storage/...` passes and `fluffy build web` succeeds.
- [x] 1.2 Add writing collection-only backups and thinning to `golang/backup`. Add `backup_test.go` cases for the thinning scenario with an injected clock. Verify `go test ./backup/...` passes.

## 2. App behaviour

- [x] 2.1 Run the periodic backup (only when changed) and the pre-operation backup before restore (and expose a hook other risky operations call). Add backup settings to `core.Settings`. Add `svc_test.go` cases for both automatic backup scenarios. Verify `go test ./...` passes.
- [x] 2.2 Add `AppService` calls to list, restore (media untouched) and create backups and report space used. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 3. UI

- [x] 3.1 Add the backup section to `settings_screen.dart` (settings, list, restore with confirmation, create now, space used) and the Ctrl/Cmd+Alt/Option+B shortcut. Add `test/settings_test.dart` tests. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Check on the web that backups survive a reload and that restoring one works.
