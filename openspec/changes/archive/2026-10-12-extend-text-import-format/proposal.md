# Proposal

## Why

Text files exported by Anki and many tools carry more header lines than Cardflip understands, and some files are not HTML at all. Cardflip reads only `#separator` and `#columns`, imports later `#` comment lines as notes, always treats field text as HTML (so `a < b` or `R&D` can render wrongly), detects only comma, tab and semicolon, and offers no way to override a wrong guess.

## What Changes

- Lines starting with `#` after the header block are comments and are skipped (reported in the summary as skipped comment lines).
- Header `#html:true|false` sets the "Allow HTML in fields" toggle; header `#tags:a b` adds those tags to every imported note.
- An **Allow HTML in fields** toggle in the import screen (default on, or from the header). When off, `<`, `>` and `&` are imported as literal text (escaped), and line breaks become `<br>`.
- Separators: `#separator:` accepts `Comma`, `Semicolon`, `Tab`, `Space`, `Pipe`, `Colon` and the literal characters; the import screen shows the detected separator and lets the user choose another one, re-parsing the preview.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-import`: new requirements for comment lines, HTML handling, extra headers and separator choice.

## Impact

- Go: `golang/importer/csv/csv.go` (`Parse` reads the extra headers and skips comments; `Build` escapes fields when HTML is off and applies header tags; a `Separator` override); `CsvPreview` and `ImportOptions` gain `AllowHTML`, `Separator` and header tags.
- Flutter: `lib/features/import/import_screen.dart` (separator dropdown, HTML toggle).
- Tests: `golang/importer/csv/csv_test.go`, `golang/svc_test.go`, `test/import_test.dart`.
- Overlap: `add-text-import-routing-columns` adds the note type, deck, tags and GUID column headers to the same parser.
