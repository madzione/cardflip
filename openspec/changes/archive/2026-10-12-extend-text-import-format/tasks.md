# Tasks

## 1. Parser

- [x] 1.1 In `golang/importer/csv/csv.go`, read `#html:` and `#tags:`, accept all separator names and characters, and skip `#` comment lines after the header block with a count. Add `csv_test.go` cases for each header and the comment scenario. Verify `go test ./importer/csv/...` passes.
- [x] 1.2 In `Build`, escape `<`, `>` and `&` and convert newlines to `<br>` when HTML is off, and add header tags to every note. Add cases. Verify they pass.
- [x] 1.3 Add `AllowHTML`, `Separator` and header tags to `CsvPreview` and `ImportOptions`, re-parsing with an override separator in `Inspect`. Add `svc_test.go` cases and the comment count in `ImportSummary`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the separator dropdown and HTML toggle to `import_screen.dart`, refreshing the preview on change. Add `test/import_test.dart` tests for the override scenario and the toggle. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a notes text file exported by current Anki with headers.
