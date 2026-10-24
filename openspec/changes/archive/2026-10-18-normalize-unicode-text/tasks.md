# Tasks

## 1. Normalisation

- [x] 1.1 Add `golang.org/x/text/unicode/norm` and normalise to NFC in note add/update, tag edits, deck and note type renames, field renames, templates, CSS, deck descriptions, `AddMedia` names and `Collection.Import` (media names included). Add `core` tests for each entry point with decomposed input. Verify `go test ./core/...` passes.
- [x] 1.2 Normalise queries in `search.Parse`. Add `search_test.go` cases for the search scenario. Verify `go test ./search/...` passes.
- [x] 1.3 Add a one-time migration on open that normalises existing text in one op, renames stored media with non-NFC names, and marks the snapshot as normalised. Add a test that opens an old snapshot with decomposed text, checks the normalised result, and checks that a second open does not write. Verify `go test ./core/...` passes.

## 2. Integration check

- [x] 2.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Record the WASM size difference from `fluffy build web`.
  - WASM size: `golang.wasm` from `fluffy build web` is 13,935,348 bytes, against 13,863,930 for the same tree with the NFC code stubbed out (same flags, `-trimpath -ldflags="-s -w"`): +71,418 bytes (+0.5%). The `norm` tables were linked already (`render` and `search` used them), so this is code only, not the ~100 KB of tables the proposal estimated.
