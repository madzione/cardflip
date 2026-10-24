# Proposal

## Why

When a deck is imported from Anki with its scheduling, Cardflip keeps due dates and memory states but not the review history (the importer never reads Anki's `revlog` table). Users moving from Anki lose years of history: card info has nothing to show, statistics start from zero, and FSRS parameters cannot be optimized from their past reviews.

## What Changes

- When an `.apkg` is imported with "keep scheduling", the review history of newly added cards is imported too: time, rating, kind (learning, review, relearning, manual), interval before and after, and answer time.
- Entries Anki marks as manual or rescheduled are imported as manual entries; Anki's filtered-deck ("cram") entries are imported as reviews.
- When a card has no FSRS memory state in Anki, its memory state is computed from the imported history with FSRS instead of being estimated from interval and ease.
- Re-importing the same package does not duplicate history entries.
- Imported entries cannot be undone as study answers.
- The import summary reports the number of review entries imported.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-import`: new requirements for importing review history and deriving memory states from it.

## Impact

- Go: `golang/importer/apkg/reader.go` reads `revlog` (legacy and modern schemas are the same table); `core.ImportBatch` gains entries per card; `core/import.go` writes them through the review log segments (`storage.RevlogWrite`) in the import's commit, de-duplicated by Anki revlog ID kept on the entry (`Source`); `scheduler/anki.go` gains memory state from history (replaying FSRS over the entries).
- Tests: `golang/importer/apkg/apkg_test.go` (fixture with history), `golang/scheduler` tests for the replay, `golang/core` import tests, `golang/svc_test.go`.
- Storage: large histories (hundreds of thousands of entries) are written in month segments like normal reviews; import progress covers this step.
