# Tasks

## 1. Filters

- [x] 1.1 Add `golang/render/furigana.go` with `furigana`, `kana` and `kanji` transforms following Anki's reading pattern (` ?([^ >]+?)\[(.+?)\]`, skipping matches whose reading starts with `sound:`). Add `furigana_test.go` with the spec scenarios, a field with HTML tags around the base, and text without readings. Verify `go test ./render/...` passes.
- [x] 1.2 Register the three filters in `context.filter` in `template.go`, so they chain with other filters (for example `{{furigana:text:Field}}`). Add `template_test.go` cases. Verify `go test ./render/...` passes.
- [x] 1.3 Allow `ruby`, `rt` and `rp` through `sanitize.go` and add a `sanitize_test.go` case. Verify it passes.

## 2. Display

- [x] 2.1 Add a `test/card_view_test.dart` test that renders `<ruby>日本語<rt>にほんご</rt></ruby>` and asserts both texts are present with the reading above the base (smaller font, positioned above). If the HTML package does not lay out ruby, add a custom widget factory for `ruby` in `lib/widgets/card_view.dart`. Verify the test passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a Japanese `.apkg` that uses `furigana:` and confirm the readings display on desktop and web.
