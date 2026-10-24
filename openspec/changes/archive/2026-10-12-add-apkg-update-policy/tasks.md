# Tasks

## 1. Import logic

- [x] 1.1 Add `Modified` to `core.NoteType` (updated on every change) and read note and note type modification times in `importer/apkg`. Add tests. Verify `go test ./core/... ./importer/...` passes.
- [x] 1.2 Apply the note update policy in `core/import.go` and count skipped-as-older notes. Add tests for the three re-import scenarios with fixtures of an older and a newer package. Verify `go test ./core/...` passes.
- [x] 1.3 Apply the note type policy and merge (match by name, add missing, never remove) or import-as-new. Add tests for both note type scenarios. Verify `go test ./core/...` passes.
- [x] 1.4 Add the options and counters to `ImportOptions` and `ImportSummary`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the three options to the package step of `import_screen.dart` and the new counters to the summary. Add `test/import_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
