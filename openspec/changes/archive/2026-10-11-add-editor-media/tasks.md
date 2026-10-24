# Tasks

## 1. Media storage

- [x] 1.1 Add `MediaService.Add(name string, data []byte) (string, error)` in `golang/svc_transfer.go` with content-hash reuse and collision renaming, writing through the storage layer and announcing a change. Add `svc_test.go` cases for the naming scenarios and that an unreferenced file is removed by compaction. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.
- [x] 1.2 Add `PasteImagesAsPng` to `core.Settings` with a setter and a settings switch in `settings_screen.dart`. Add Go and widget tests. Verify they pass.

## 2. Editor

- [x] 2.1 Add the Attach button to `note_editor_screen.dart` and the browser's edit pane (`card_details.dart`, sharing `field_media.dart`) using `file_picker` (images, audio, video), inserting the right markup at the cursor of the focused field. Add `test/editor_test.dart` tests for an image and an audio file with a fake picker. Verify they pass.
- [x] 2.2 Add image paste (via `quill_native_bridge`, already a dependency, instead of `super_clipboard`, which needs a Rust toolchain; converted to JPEG or PNG per setting by `MediaService.AddPastedImage` in Go) and file drop (`desktop_drop` on desktop and web, where it uses the browser's drag events). Add tests with fake clipboard and drop sources. Verify they pass.
- [x] 2.3 Add Record audio with `record`, permission entries for Android, iOS and macOS (present since add-media-playback; the iOS and macOS usage text now mentions editor recordings), and the refused-permission message. Add a test with a fake recorder. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Manually attach, paste and record on Linux, Android and Chrome.
