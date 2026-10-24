# Tasks

## 1. Import

- [x] 1.1 Parse the column and default headers in `golang/importer/csv/csv.go` and route rows in `Build` to `ImportNote` with note type key, deck name, tags and GUID; skip and report unknown note types. Add `csv_test.go` cases for an Anki notes export fixture and the unknown note type scenario. Verify `go test ./importer/csv/...` passes.
- [x] 1.2 Apply GUID matching before first-field duplicate handling and add the deck scope in `core/import.go`. Add `core` tests for the three duplicate scenarios. Verify `go test ./core/...` passes.
- [x] 1.3 Add column roles and the scope to `CsvPreview` and `ImportOptions`. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Offer Note type, Deck, Tags and GUID in each column's mapping in `import_screen.dart`, with the scope option and unknown note type warnings. Add `test/import_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Round-trip: export notes from Anki with all columns, import, edit a field in Anki, export and import again, and confirm the note is updated.
