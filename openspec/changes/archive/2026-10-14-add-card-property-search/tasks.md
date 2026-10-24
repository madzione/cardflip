# Tasks

## 1. Search data

- [x] 1.1 Extend `search.Card` and its filling in `core/search.go` with the properties listed in the proposal, computing retrievability at search time and buried kinds from the day view. Verify `go test ./core/...` passes.
- [x] 1.2 Add `CurrentDeckID` to `core.Settings`, set by `StudyService.Start`. Add a `svc_test.go` case. Verify `go test ./...` passes.

## 2. Terms

- [x] 2.1 Add `prop:` parsing and comparisons, including the `prop:ease` message and unknown-property errors. Add `search_test.go` cases for each property and the scenarios. Verify `go test ./search/...` passes.
- [x] 2.2 Add `is:buried*` and the relearning semantics of `is:learn` and `is:review`. Add cases, including the lapsed-card scenario. Verify they pass.
- [x] 2.3 Add `card:`, `tag:none`, `deck:current`, `deck:<name>::*` and `nid:`/`cid:`. Add cases for each scenario. Verify `go test ./search/... ./core/...` passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
