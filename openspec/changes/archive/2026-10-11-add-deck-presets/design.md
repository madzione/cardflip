# Design

## Context

- `core.Deck` embeds `Options DeckOptions`; `Collection.params(card)` and `limited` in `core/queue.go` read `c.decks[id].Options`.
- The snapshot is versioned (`snapshotVersion = 1`); backups embed the snapshot and record a format version (`backup.Format`).

## Goals / Non-Goals

**Goals:** one place for options; no behaviour change for users who never touch presets; a lossless migration.

**Non-Goals:** preset-level FSRS parameter sharing beyond what is in `DeckOptions` (the optimizer change stores its parameters in `DeckOptions`, so they move with presets).

## Decisions

### 1. `DeckOptions` becomes the preset payload

`Preset{ID, Name, Options DeckOptions}`; `Deck{PresetID, Limits *ThisDeckLimits}`. Reading a deck's effective options = preset options with this-deck limits applied. All existing call sites switch to `c.optionsOf(deckID)`.

### 2. Migration

On loading a version-1 snapshot (or replaying a v1 journal before compaction), group decks by identical `Options`; the group equal to `DefaultDeckOptions()` maps to "Default", others to presets named after their first deck (by name order), deduplicated. Then write a version-2 snapshot at the next compaction. Restoring a v1 backup runs the same migration.

### 3. Import mapping

Modern packages store presets in `deck_config` (protobuf), legacy in `col.dconf` (JSON). Map the fields Cardflip has (limits, retention, steps, and others when those options exist); ignore the rest. Match an existing preset when name and mapped options are equal; otherwise create one, suffixing the name on collision. Anki's default preset (ID 1) maps to Cardflip's Default, as Anki does on import: its defaults differ from Cardflip's (sibling burying, gather and sort order), so matching by options would add a "Default (2)" on nearly every import.

## Risks / Trade-offs

- [Migration bug changes users' options] → Property test: for random collections, each deck's effective options after migration equal its options before.
- [Editing a shared preset surprises users] → The screen states "Used by N decks" next to the preset name.
