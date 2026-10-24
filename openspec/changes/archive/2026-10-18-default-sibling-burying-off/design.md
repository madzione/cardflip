# Design

## Context

- `core.DefaultDeckOptions()` sets `BuryNew` and `BuryReview` to true. It is the starting point for the Default preset of a new collection (`defaultPreset`), for presets added with "start with the default options" (`AddPreset`), for imported Anki presets before their values are copied in (`importer/apkg/convert.go`), and for decks without a preset (`optionsOf`).
- `DeckOptions.UnmarshalJSON` decodes stored options on top of `DefaultDeckOptions()`, so a key missing from stored JSON takes the default. Presets (snapshot version 2) always store `buryNew` and `buryReview`. Decks stored before presets existed (version 1) may lack them; `migratePresets` turns those decks into presets and compares their options with `DefaultDeckOptions()` to decide which go into the Default preset.
- The collection snapshot has a version (`snapshotVersion = 2`); opening an older snapshot or restoring an older backup migrates it (`migrateDecks`, `migrateState`) and compacts the store.
- The exporter already writes Anki's defaults (burying off) when presets are not exported (`ankiDefaultOptions`).

## Goals / Non-Goals

**Goals:**
- Burying off for every Default preset Cardflip ships: new collections, existing collections and restored older backups.
- No other preset changes its burying options on its own.

**Non-Goals:**
- Changing how burying works when it is on, or the interday learning option (already off).
- Exempting particular note types (such as Image Occlusion) from burying.

## Decisions

### 1. New defaults

`DefaultDeckOptions()` sets `BuryNew` and `BuryReview` to false. Every place that starts from the defaults (new collection, added preset, decks without a preset) gets burying off without further changes. The options screen's initial values match.

### 2. Decoding keeps "missing means on"

`DeckOptions.UnmarshalJSON` sets `BuryNew` and `BuryReview` to true before decoding, so options stored without these keys read as they did before (on). Current data always stores both keys, so this only affects version 1 decks. `migratePresets` compares version 1 decks with the old defaults (the new defaults with burying on), so the same decks as before go into the Default preset.

Alternative: let missing keys take the new default. Rejected: user presets from version 1 collections would silently stop burying.

### 3. One-time migration of the Default preset

`snapshotVersion` becomes 3. When a snapshot older than 3 is opened, after the journal is replayed and decks are migrated to presets, the Default preset (`DefaultPresetID`) gets `BuryNew` and `BuryReview` set to false, and the collection is compacted so the migration runs once. Restoring a backup older than version 3 does the same to the restored state (`migrateState`). Other presets are not touched: those were added by the user or imported from Anki, whose values are deliberate.

The migration is not an undo step and not in the history: it happens while opening, like the preset migration.

Alternative: only change the defaults (existing Default presets keep burying). Rejected by the user: every Default preset Cardflip ships should have burying off.

## Risks / Trade-offs

- [A user who turned burying on in the Default preset on purpose loses it once] → Accepted as decided; turning it back on is one switch in the deck options.
- [Tests and the integration flow assume burying by default] → Updated in the tasks; they set burying on where they test burying.
- [Snapshot version 3 cannot be opened by older Cardflip versions] → As with version 2, older versions report that the collection was saved by a newer Cardflip.
