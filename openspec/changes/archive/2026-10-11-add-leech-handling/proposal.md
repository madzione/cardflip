# Proposal

## Why

Leeches are cards a user keeps forgetting; they cost a large share of review time. Anki detects them by lapse count (8 by default), tags them `leech`, can suspend them, and warns the user so the card can be rewritten. Cardflip has no leech detection, so problem cards stay in rotation indefinitely and unnoticed.

## What Changes

- Deck options **Leech threshold** (lapses, default 8, 1–99) and **Leech action** (tag only, or suspend card and tag; default tag only).
- When a review card lapses and its lapse count reaches the threshold, and every half threshold after that (rounded up: 8, 12, 16, …), the note gets the tag `leech`, the card is suspended if so configured, and the study screen shows a message that the card is a leech.
- Undoing that answer reverses the suspension and the tag (if the tag was added by it).
- Importing an `.apkg` as New (without scheduling) removes the tag `leech`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the leech deck options.
- `scheduling`: new requirement for leech detection.
- `study-session`: new requirement for the leech message and undo.
- `data-import`: new requirement for removing `leech` on import as New.

## Impact

- Go: `core.DeckOptions` gains `LeechThreshold`, `LeechAction`; `Collection.Answer` applies the leech rule in the same op as the answer (note tags, card suspension) and records what it changed in the review entry so undo can revert it; `StudyState` reports that the last answer made a leech; `importer/apkg` strips the tag when scheduling is not kept.
- Flutter: `deck_options_dialog.dart`, `study_screen.dart` (snackbar message).
- Tests: `golang/core/reviews_test.go`, `golang/svc_test.go`, `golang/importer/apkg/apkg_test.go`, `test/study_test.dart`.
