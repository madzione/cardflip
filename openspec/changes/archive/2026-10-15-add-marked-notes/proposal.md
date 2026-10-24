# Proposal

## Why

Anki users mark notes they want to revisit (fix a typo later, add an example) by toggling the `marked` tag with one key while studying. Anki shows a star on marked cards and colours their browser rows. In Cardflip the tag can be typed by hand, but nothing treats it specially, so the quick workflow is missing.

## What Changes

- While studying, `*` (and a menu item) toggles the `marked` tag on the current card's note; a star shows on marked cards.
- The browser gets a bulk **Toggle mark** action and tints rows of marked notes purple (a flag colour, where one exists, takes precedence).
- Importing an `.apkg` as New (without scheduling) removes the `marked` tag from imported notes, as Anki does.
- The tag stays an ordinary tag: it can be searched with `tag:marked` and edited like any other.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the `marked` tag toggle.
- `study-session`: new requirement for marking during study.
- `card-browser`: new requirement for the mark action and row tint.
- `data-import`: new requirement for removing `marked` on import as New.

## Impact

- Go: `Collection.ToggleMark(noteIDs)` (adds `marked` to all when any lacks it, otherwise removes it); `Row` and `StudyCard` gain `Marked bool`; `importer/apkg` strips the tag when scheduling is not kept.
- Flutter: `study_screen.dart` (key, menu item, star), `browse_screen.dart` (bulk action, row tint).
- Tests: `golang/core/notes_test.go`, `golang/svc_test.go`, `golang/importer/apkg/apkg_test.go`, `test/study_test.dart`, `test/browse_test.dart`.
- Overlap: `add-card-flags` adds row tints for flags and suspended cards; if both exist, the order is flag, then marked, then suspended.
