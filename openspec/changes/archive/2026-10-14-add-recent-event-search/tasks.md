# Tasks

## 1. Data

- [x] 1.1 Derive card creation time from the card ID (IDs are strictly increasing epoch milliseconds, `core/ids.go`) in the search data. Add a test that a card created at a fixed clock time matches `added:1` on that study day only. Verify `go test ./core/...` passes.
- [x] 1.2 Build per-card day indexes from the review log (latest answer day, first answer day, ratings by day, manual entry days), maintained on answer, undo and import. Add `core` tests. Verify they pass.

## 2. Terms

- [x] 2.1 Add the event terms and `prop:rated`/`prop:resched` to `golang/search`, with day boundaries from the rollover hour and the 365 limit. Add `search_test.go` and `core/search_test.go` cases for every spec scenario. Verify `go test ./search/... ./core/...` passes.

## 3. Add screen

- [x] 3.1 Add the History action to `note_editor_screen.dart`, opening Browse with `added:1`. Add a `test/editor_test.dart` test. Verify it passes.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
