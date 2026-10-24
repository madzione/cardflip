# Proposal

## Why

Anki's card type options cover three things Cardflip lacks: reordering card types (so "Recognition" comes before "Production"), a deck override (send production cards to a separate deck, a common setup for language learners), and simplified browser templates (so the browser's question column shows the word, not the whole styled front with hints). Without them, users have to delete and recreate card types, which loses scheduling, or move cards by hand after every add.

## What Changes

- **Reposition** a card type within a standard note type. Existing cards keep their scheduling; only the order changes, which affects the order cards of a note are introduced and listed.
- **Deck override** per card type: new cards of that card type are added to the chosen deck instead of the deck picked in the editor. Existing cards are not moved. Removing the override, or deleting that deck, falls back to the editor's deck.
- **Browser appearance** per card type: optional simplified question and answer templates used for the browser's question text (and any answer text the browser shows). Empty means the normal templates.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for card type order and deck override.
- `card-browser`: new requirement for browser appearance templates.

## Impact

- Go: `core.CardTemplate` gains `DeckOverride ID`, `BrowserQFmt` and `BrowserAFmt` (`omitempty`); `Collection.MoveTemplate(typeID, from, to)` remaps card ordinals of every note of that type in one op; `AddNote` and card generation use the override; the browser question text in `core/search.go` uses `BrowserQFmt` when set.
- Flutter: `lib/features/editor/note_type_manager.dart` (reposition handle and an options sheet per card type).
- Tests: `golang/core/notes_test.go`, `golang/core/search_test.go`, `golang/svc_test.go`, `test/editor_test.dart`.
- `.apkg` import maps Anki's `did` override and `bqfmt`/`bafmt` to these fields.
