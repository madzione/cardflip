# Tasks

## 1. Writer

- [x] 1.1 Add `golang/exporter/text` writing notes (headers, quoting, optional columns) and cards (rendered and stripped). Add round-trip tests through `golang/importer/csv` for quoting, line breaks and tabs. Verify `go test ./exporter/... ./importer/...` passes.
- [x] 1.2 Add `TransferService.ExportNotesText` and `ExportCardsText` with deck, selection and collection scopes. Add `svc_test.go` cases, including the re-import scenario with the "update" duplicate option (no duplicates). Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the text export dialog (notes or cards, column options) to the deck menu, the browser selection and Settings, saving via file picker or download. Add widget tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import an exported notes file into Anki and confirm the columns are recognised.
