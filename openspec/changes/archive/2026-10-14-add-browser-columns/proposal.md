# Proposal

## Why

Cardflip's browser row is fixed (question, deck, state or due, note type) and sorts by only three keys. Reviewing a deck often needs other information at a glance: the answer, interval, difficulty, lapses, review count, tags, or when a card was last modified, and sorting by them to find the hardest or most-lapsed cards. Anki's browser has configurable, sortable columns.

## What Changes

- Available columns: Question, Answer, Sort field, Card type, Note type, Deck, Due, Interval, Difficulty, Stability, Retrievability, Reviews, Lapses, Tags, Created, Card modified, Note modified.
- The user shows, hides and reorders columns (column header menu, or a "Columns" dialog on touch devices); the layout is remembered.
- Any column can be the sort column, ascending or descending, by clicking its header.
- On narrow screens, the row shows the first column as the title and the next two as a subtitle, so the same choice works on phones.
- Defaults reproduce today's row: Question, Deck, Due, Note type.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: "Card list" changes from a fixed row and three sort keys to configurable columns sortable by any column.

## Impact

- Go: `Row` becomes a column-value map computed only for the requested columns (`BrowseService.Rows(ids, columns)`); `Collection.Search` gains sort keys for every column (answer and sort field text cached like the question); `core.Settings` gains `BrowserColumns`.
- Flutter: `browse_screen.dart` becomes a table on wide screens (headers, resizing, reordering) and keeps list tiles on narrow screens.
- Performance: sorting by answer or computed values over 50,000 cards must stay within the existing responsiveness requirement; computed keys are cached and invalidated like the question cache.
- Tests: `golang/core/search_test.go`, `golang/svc_test.go`, `test/browse_test.dart`, the performance integration test.
- Overlap: `add-card-type-options` (browser appearance templates) affects the Question and Answer columns; `add-field-options` defines the sort field (the first field until then).
