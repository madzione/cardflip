# Tasks

## 1. Matching data

- [x] 1.1 Extend `search.Card` with field names and per-field stripped (entity-decoded) and raw text, filled and cached in `core/search.go` with the same invalidation as the question cache. Add a benchmark that filling 50,000 cards stays under 200 ms. Verify `go test ./core/...` passes.

## 2. Terms

- [x] 2.1 Add wildcard text and `w:` terms with escapes. Add `search_test.go` cases. Verify `go test ./search/...` passes.
- [x] 2.2 Add field terms (exact, wildcards, field-name wildcards, empty, non-empty, exists). Add cases for the spec scenarios. Verify they pass.
- [x] 2.3 Add `nc:` (NFD and combining-mark removal), `sc:` (cloze markup removal) and the global ignore-accents flag. Add cases. Verify they pass.
- [x] 2.4 Add `re:`, `<field>:re:` and `tag:re:` with `(?-i)` handling and error reporting. Add cases. Verify they pass.

## 3. Setting

- [x] 3.1 Add `IgnoreAccents` to `core.Settings` with a setter, pass it into query matching, and add the switch to `settings_screen.dart`. Add `svc_test.go` and `test/settings_test.dart` tests. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Time `front:*x*` and `re:` searches on the 50,000-card fixture and verify results appear within 1 second.
