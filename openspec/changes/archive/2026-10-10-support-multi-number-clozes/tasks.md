# Tasks

## 1. Parser and rendering

- [x] 1.1 Change `clozeOpen` in `golang/render/cloze.go` to parse `c` followed by one or more comma-separated positive numbers (optional spaces after commas) and `::`, returning the set of numbers; reject empty entries such as `c1,,2`. Add `cloze_test.go` cases for `c1,2`, `c1, 2, 3`, `c1,,2` and `c0,1`. Verify `go test ./render/...` passes.
- [x] 1.2 Store the numbers on `clozePiece` and update `writeCloze` and `plainCloze` so a deletion is active when the card's number is in its set, including nested deletions and hints. Add `cloze_test.go` cases for the spec scenarios (cards 1–3 of `{{c1,2::Paris}} is in {{c3::France}}`, the hint, and a multi-number deletion nested in a single-number one). Verify they pass.
- [x] 1.3 Make `ClozeNumbers` (and `ClozeCards`) return every listed number. Add a case asserting `{{c1,2::x}} {{c4::y}}` yields 1, 2, 4. Verify `go test ./render/...` passes.

## 2. Card generation and import

- [x] 2.1 Add a `core/notes_test.go` case: adding `{{c1,2::Paris}}` creates two cards, and editing it to `{{c1::Paris}}` keeps card 2 as an empty card and keeps card 1's scheduling. Verify `go test ./core/...` passes.
- [x] 2.2 Add a note with a multi-number deletion to the `.apkg` test fixture in `golang/importer/apkg/apkg_test.go` and assert the right number of cards and the rendered question. Verify `go test ./importer/...` passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
