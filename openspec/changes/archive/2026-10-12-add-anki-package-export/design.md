# Design

## Context

- The importer (`golang/importer/apkg`) reads both Anki formats with its own read-only SQLite reader (`importer/apkg/sqlite`), because a cgo SQLite does not build for WASM and an embedded SQL engine is too large.
- Current Anki imports packages with `collection.anki21b` (schema 18, zstd), `collection.anki21` or `collection.anki2` (schema 11).

## Goals / Non-Goals

**Goals:** packages that current Anki and Anki 2.1.x import without warnings; pure Go; works in WASM; reasonable size.

**Non-Goals:** writing schema 18 (`anki21b`), which needs protobuf configs for every table; exporting Cardflip-only features that Anki lacks.

## Decisions

### 1. Write schema 11 for every package

Current Anki still imports `collection.anki21` (schema 11) packages, and 2.1.x reads `collection.anki2`. Default packages contain `collection.anki21` plus the dummy `collection.anki2` Anki writes for old clients; "support older Anki versions" writes the real data as `collection.anki2`. This avoids reproducing schema 18's protobuf tables.

Alternative: schema 18. Rejected for now: much larger surface, and current Anki accepts schema 11.

### 2. Minimal SQLite writer

Write the database file directly: header, schema table, table b-trees (leaf and interior pages), and the indexes Anki expects to exist (Anki recreates missing indexes on open, so only tables are mandatory; we write the standard indexes for safety). Records use SQLite's record format; integer primary keys as rowids. Validated in tests by reading back with the existing reader and with a reference SQLite (`modernc.org/sqlite` in test-only code, native builds).

### 3. Card and scheduling mapping

New cards: `type=0, queue=0, due=position`. Learning: `type=1, queue=1, due=epoch seconds` (interday learning `queue=3, due=day`). Review: `type=2, queue=2, due=days since collection creation, ivl`. Suspended `queue=-1`, buried `queue=-3`. `factor` from difficulty with Anki's FSRS mapping; FSRS state in `data` as `{"s":…, "d":…, "dr":…}`. Review log entries map back to `revlog` with ease, type and intervals. Without scheduling, cards are new and no revlog is written.

### 4. Note types and decks

Note types become schema-11 models (fields, templates, CSS, cloze type, sort field when known). Deck options become deck configs when "include deck presets" is on; otherwise decks use Anki's default config.

## Risks / Trade-offs

- [Anki rejects a subtle format detail] → Keep a manual check in tasks (import into current Anki and 2.1.66) and test fixtures that compare our `col` JSON keys with an Anki-exported package.
- [Large collections in WASM memory] → Write pages incrementally into the zip stream; media streamed from storage.
