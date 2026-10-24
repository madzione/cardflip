# Proposal

## Why

Power users repeat the same browser searches every day ("leeches in Spanish", "added this week without audio"). Anki lets them save named searches and set a default search the browser starts with. In Cardflip, searches are typed again each time and the browser always opens unfiltered.

## What Changes

- **Save search**: the current search text can be saved under a name. Saved searches are listed in a menu next to the search field; choosing one runs it. They can be renamed and deleted.
- A setting **Default browser search**: the search the browser shows when it opens (empty by default, meaning all cards).
- Saved searches and the default search are stored in the collection, so backups include them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for saved searches and the default search.

## Impact

- Go: `core.Settings` gains `SavedSearches []SavedSearch{Name, Query}` and `DefaultSearch`; `BrowseService` gains save, rename, delete and list calls; saving validates the query by parsing it.
- Flutter: `browse_screen.dart` (menu next to the search field), `settings_screen.dart` (default search).
- Tests: `golang/svc_test.go`, `test/browse_test.dart`, `test/settings_test.dart`.
- Overlap: `add-browser-sidebar` can show saved searches as a sidebar section.
