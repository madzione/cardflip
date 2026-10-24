# Tasks

## 1. Go

- [x] 1.1 Add `SavedSearches` and `DefaultSearch` to `core.Settings`, and `BrowseService` calls to list, save (parse to validate, unique names), rename and delete, plus a setter for the default search. Add `svc_test.go` cases. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the saved-search menu (run, save current, rename, delete) to `browse_screen.dart` and start with the default search. Add `test/browse_test.dart` tests for both scenarios of each requirement. Verify they pass.
- [x] 2.2 Add the default search field to `settings_screen.dart` with validation. Add a `test/settings_test.dart` test. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
