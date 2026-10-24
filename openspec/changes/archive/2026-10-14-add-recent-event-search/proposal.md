# Proposal

## Why

"What did I add today?" and "which cards did I fail this week?" are among the most common searches in Anki (`added:1`, `rated:7:1`). Cardflip cannot search by when something happened, and the add screen has no way to get back to recently added notes.

## What Changes

- Searches by study day, counted back from today with the rollover hour (`1` = today):
  - `added:N` cards added in the last N days;
  - `edited:N` cards whose note was edited in the last N days;
  - `rated:N` cards answered in the last N days, and `rated:N:R` answered with rating R (1–4);
  - `introduced:N` cards answered for the first time in the last N days;
  - `resched:N` cards whose due date was changed manually in the last N days.
- `prop:rated` and `prop:resched` compare the day of the latest answer or manual reschedule with today (`prop:rated=-1` = yesterday).
- N is limited to 365 for `rated:` as in Anki, to keep it fast.
- A **History** button in the add screen opens the browser with `added:1`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for event searches and the add screen's History button.

## Impact

- Go: card creation time comes from the card ID, which is epoch milliseconds (`core/ids.go`), so imported cards count as added on the day of the import; note `Modified` already exists; review-log lookups per card for `rated`, `introduced` and manual entries, indexed by day; `search` term kinds.
- Flutter: `note_editor_screen.dart` (History button, opening Browse with a search).
- Tests: `golang/search/search_test.go`, `golang/core/search_test.go`, `test/editor_test.dart`.
- `resched:` and `prop:resched` match nothing until manual rescheduling entries exist (they are added by the set-due-date and reset-options changes).
