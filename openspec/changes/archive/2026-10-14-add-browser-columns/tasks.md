# Tasks

## 1. Go

- [x] 1.1 Define the column set and compute requested columns in `Collection.Rows` (answer and sort field text cached with invalidation like `questionCache`). Add `search_test.go` cases for each column's value. Verify `go test ./core/...` passes.
- [x] 1.2 Add a sort key for every column in `Collection.Search` and `sortCards`. Add tests for sorting by lapses, interval, answer and tags. Add a benchmark sorting 50,000 cards by answer, and verify it stays under 500 ms on a typical desktop.
- [x] 1.3 Change `BrowseService.Rows(ids, columns)` and `Search(query, column, descending)`, add `BrowserColumns` to settings. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Build the table layout for wide screens (header click to sort, header menu to show, hide and reorder, remembered) and the title/subtitle layout for narrow ones in `browse_screen.dart`. Add `test/browse_test.dart` tests for the three spec scenarios on both layouts. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Run the performance integration test in `integration_test/perf` and verify the 1-second requirement.
