# Tasks

## 1. Storage and references

- [x] 1.1 Add a media trash to `storage.Backend` (move to trash with time, list, restore, purge older than a time) for files and IndexedDB backends, with contract tests in `golang/storage/storagetest`. Verify `go test ./storage/...` passes and `fluffy build web` succeeds.
- [x] 1.2 Extend reference collection in `golang/core` to fields, templates and CSS (`img`, `[sound:]`, `audio`, `video`, `url()`), treat `_` files as used, and make compaction move orphans to the trash and purge after 30 days. Add tests for both check scenarios and compaction. Verify `go test ./core/...` passes.

## 2. Service and UI

- [x] 2.1 Add `AppService.CheckMedia`, `DeleteUnusedMedia`, `TagMissingMedia`, `EmptyMediaTrash` and `RestoreMediaTrash`. Add `svc_test.go` cases for each requirement. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 2.2 Add Check media and the report (lists, sizes, actions with confirmations) to `settings_screen.dart`. Add `test/settings_test.dart` tests. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
