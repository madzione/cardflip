# Tasks

## 1. Go

- [x] 1.1 Export a small collection from Anki as `.colpkg` in modern and legacy formats and add both to `golang/importer/apkg/testdata`. Recognise `.colpkg` in `kindOfFile` and `Inspect`. Add tests. Verify `go test ./...` passes.
- [x] 1.2 Build a fresh collection from the package (empty decks included, filtered decks folded into home decks) and swap it in through the restore path, keeping `Settings`, all or nothing. Add `svc_test.go` cases for the move, cancel (no call) and corrupt scenarios. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Allow `.colpkg` in `pick.dart` and show the replace confirmation with a suggestion to export a backup first in `import_screen.dart`. Add `test/import_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a real Anki collection on Linux and in Chrome.
