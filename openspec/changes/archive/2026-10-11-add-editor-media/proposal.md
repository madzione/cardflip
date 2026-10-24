# Proposal

## Why

The only way to get an image onto a Cardflip card today is to import an `.apkg`. Users cannot attach a picture, paste a screenshot, or add a pronunciation recording while creating cards, which Anki's editor does with one button. This blocks visual and audio cards made by the user.

## What Changes

- An **Attach** button (paperclip) in the editor picks image, audio or video files and inserts them into the focused field: images as `<img src="…">`, audio and video as `[sound:…]`.
- Pasting an image from the clipboard inserts it into the field. A setting "Paste images as PNG" chooses PNG (lossless) over JPEG (smaller, the default).
- Dropping files onto a field on desktop and web inserts them like Attach.
- A **Record audio** button records from the microphone and inserts the recording as `[sound:…]`.
- Added files are stored in the collection's media under a name derived from their content (so the same file added twice is stored once), keeping the original extension; a file name that already exists with different content gets a numbered suffix.
- Media added to a note that is never saved is not kept after the editor closes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for attaching, pasting, dropping and recording media in the editor.
- `app-shell`: new requirement for the "Paste images as PNG" setting.

## Impact

- Go: `MediaService.Add(name, bytes) (storedName, error)` writing through the storage media API (content hash dedupe, name collision handling), and `MediaService.AddPastedImage(bytes, asPng)`, which converts a clipboard image to JPEG or PNG with the standard library; media referenced by no note stays subject to the existing compaction garbage collection, which covers unsaved notes.
- Flutter: `note_editor_screen.dart` (Attach, Record, paste and drop handling), `file_picker` (already a dependency), `quill_native_bridge` (already a dependency) for reading clipboard images, `desktop_drop` for drag-and-drop on desktop and web, `record` for audio; microphone permission entries for Android, iOS and macOS.
- Tests: `golang/svc_test.go`, `golang/storage` tests for media writes, `test/editor_test.dart`.
