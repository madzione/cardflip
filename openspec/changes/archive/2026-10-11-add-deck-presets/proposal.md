# Proposal

## Why

With many decks, setting options deck by deck is tedious and error-prone: a language collection with twenty subdecks needs the same retention and limits in each. Anki shares options through named presets, applies one to all subdecks in one step, and carries presets inside shared `.apkg` files. Cardflip stores options per deck, and `.apkg` import drops them, so every imported deck gets the defaults.

## What Changes

- Options live in named **presets**. Each deck uses one preset; a new collection has the preset **Default**, used by every deck.
- The deck options screen shows the deck's preset with a picker and actions: add, clone, rename, delete (decks using a deleted preset switch to Default; Default cannot be deleted), and **Save to all subdecks** (assigns the preset to the deck's descendants).
- The screen shows how many decks use the preset, since editing it changes all of them.
- **This deck** limits: new and review limits can be set for one deck, overriding its preset's limits.
- Search: `preset:"Name"` finds cards in decks using that preset.
- `.apkg` import brings the package's presets and assigns them to the decks it creates; existing decks keep their preset. A package preset whose name exists with different options is imported under a numbered name.
- Existing collections are migrated: decks with the default options use Default; each distinct set of other options becomes a preset named after the first deck that used it.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: "Deck options" now describes options as presets shared by decks; new requirements for managing presets and this-deck limits.
- `card-browser`: new requirement for the `preset:` search.
- `data-import`: new requirement for importing presets from `.apkg`.

## Impact

- Go: new `core.Preset{ID, Name, Options DeckOptions}` in the snapshot; `Deck.Options` is replaced by `Deck.PresetID` and `Deck.Limits` (this-deck overrides); every reader of deck options (`Collection.params`, `core/queue.go`) goes through the deck's preset; a snapshot migration (version 2) builds presets from per-deck options; `DecksService` gains preset calls; `search` gains `preset:`; `importer/apkg` reads `deck_config` (modern) and `dconf` (legacy).
- Backup format: the snapshot version increases; backups from older versions are migrated on restore; older app versions refuse new backups as designed.
- Flutter: `deck_options_dialog.dart` becomes a screen with the preset bar and actions.
- Tests: `golang/core/decks_test.go`, `golang/core/queue_test.go`, `golang/core` migration tests, `golang/search/search_test.go`, `golang/importer/apkg/apkg_test.go`, `golang/backup/backup_test.go`, `test/deck_list_test.dart`.
- Overlap: every change that adds deck options (`configure-learning-steps`, `add-leech-handling`, `add-display-order-options`, …) adds fields to `DeckOptions`; with presets, those fields live in the preset automatically. `extend-daily-limits`' today-only limits stay on the deck.
