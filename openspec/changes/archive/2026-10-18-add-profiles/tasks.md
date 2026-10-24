# Tasks

## 1. Storage and app

- [x] 1.1 Add a profile registry (names, last used, startup setting) stored outside collections, and per-profile data locations for the files and IndexedDB backends with the existing location as `Default`. Add storage and `svc_test.go` tests for the upgrade scenario. Verify `go test ./...` passes and `fluffy build web` succeeds.
- [x] 1.2 Add `AppService` profile calls (list, add, rename, delete with card count, switch) that close and reopen collections and reset study and transfer state. Add `svc_test.go` cases for separation, management and backups per profile. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. UI

- [x] 2.1 Add the profile list screen (open, add, rename, delete with confirmation) and the Settings entry and startup setting. Add `test/settings_test.dart` tests. Verify they pass.
- [x] 2.2 Show the profile list at startup when set, and the profile name on the deck list when there are several, in `lib/app/bootstrap.dart` and `deck_list_screen.dart`. Add `test/app_shell_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Add an integration test that creates two profiles, adds a card in each and switches, and run it on Linux and web.
