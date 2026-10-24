# Tasks

## 1. Reading and mapping

- [x] 1.1 Create a fixture deck in Anki with review history (learning, review, relearning, a manual reschedule and a filtered-deck review), export it with scheduling in modern and legacy formats, and add it to `golang/importer/apkg/testdata`.
- [x] 1.2 Read `revlog` in `importer/apkg/reader.go` and map entries (ease 0 or type 4/5 → manual; type 3 → review; negative intervals as seconds). Add `apkg_test.go` cases for both formats. Verify `go test ./importer/...` passes.

## 2. Writing and memory states

- [x] 2.1 Carry entries in `core.ImportBatch` for newly added cards with "keep scheduling", write them through the revlog segments in the import commit, de-duplicate by Anki revlog ID, mark them not undoable, and count them in the summary. Add `core` tests for the three history scenarios. Verify `go test ./core/...` passes.
- [x] 2.2 Replay FSRS over the history for cards without an Anki memory state in `scheduler/anki.go`. Add tests for the SM-2 scenario. Verify `go test ./scheduler/... ./core/...` passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a real collection export with 100,000+ reviews and check progress and time.
