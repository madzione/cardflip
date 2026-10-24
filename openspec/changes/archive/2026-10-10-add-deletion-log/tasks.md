# Tasks

## 1. Log

- [x] 1.1 Append deletion entries in the same commit as every note deletion in `golang/core` (browser, study, notes left without cards by a deck delete, note type delete), stored append-only through `storage`. Add tests for each path, including crash replay. Verify `go test ./core/...` passes.
- [x] 1.2 Include the log in backup export and restore. Add a `backup_test.go` round trip. Verify `go test ./backup/...` passes.
- [x] 1.3 Add `AppService.ExportDeletionLog` (TSV with headers readable by `importer/csv`) and `ClearDeletionLog`. Add `svc_test.go` cases including importing the exported file. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add Export deletion log and Clear deletion log (with confirmation) to `settings_screen.dart`. Add `test/settings_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
