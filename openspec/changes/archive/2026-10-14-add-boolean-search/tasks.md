# Tasks

## 1. Parser

- [x] 1.1 Add a tokenizer to `golang/search/search.go` handling quotes, escapes, parentheses, `-` before terms and groups, and the words `and`/`or`. Add `search_test.go` cases for every escape and quoting form. Verify `go test ./search/...` passes.
- [x] 1.2 Replace the flat term list with an expression tree (recursive descent, AND tighter than OR) and evaluate it in `Match`. Keep all existing tests green and add cases for the spec scenarios and nested groups. Verify `go test ./search/... ./core/...` passes.
- [x] 1.3 Report syntax errors with the problem and position through the existing `search.Error`. Add tests for unclosed and stray parentheses, dangling `or` and empty groups. Verify they pass.

## 2. Integration check

- [x] 2.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Add a `test/browse_test.dart` case showing the unclosed-group error.
