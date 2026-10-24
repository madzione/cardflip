# Design

## Context

The exporter writes Anki's legacy collection schema for both package formats (`collection.anki21` and, for older Anki, `collection.anki2`). Note types go into the `col.models` JSON through `model()` in `golang/exporter/apkg/apkg.go`, which hard-codes `"bqfmt": ""`, `"bafmt": ""` and `"did": nil` for every template. The exporter already maps every exported deck to its Anki ID in `exporter.deckIDs` (Cardflip IDs are kept, the top-level Default deck becomes Anki's deck 1). `ExportSet.Decks` holds exactly the exported decks: the deck tree for a deck export, the cards' decks for a note or card export, every regular deck for the whole collection.

The import side (`importer/apkg`, from `add-card-type-options`) already reads `bqfmt`/`bafmt` and `did` from legacy JSON and keeps an override when the package contains that deck and the deck exists after the import.

## Goals / Non-Goals

**Goals:**
- Write the three options so that Cardflip's own importer restores them (see specs).

**Non-Goals:**
- Writing the modern protobuf note type tables: the exporter does not write that schema at all.
- Browser font name and size (`bfont`/`bsize`): Cardflip has no such options; they stay `""`/`0`.
- Making the import create an override deck that no card of the package is in. A deck export whose override subdeck is empty keeps `did` in the package, but importing it into a collection without that deck drops the override, as the import already does today.

## Decisions

1. **Map the override through `exporter.deckIDs`.** `model()` gets the deck ID mapping (or the exporter becomes its receiver) and writes `did` as `deckIDs[t.DeckOverride]` when the deck is in the map, else `nil`. Alternative: always write the Cardflip deck ID. Rejected: the Default deck changes ID on export, and an ID with no deck in the package means nothing to the importing app.
2. **No override for a deck outside the export.** Writing a `did` that the package does not contain would leave Anki and Cardflip pointing at an unknown deck, which Cardflip's import ignores anyway. Null is what Anki itself writes for no override.
3. **Empty strings for missing browser templates**, as now and as Anki writes them, so the JSON stays as Anki's own exports and `TestColumnJSONMatchesAnki` keeps its key set.

## Risks / Trade-offs

- [Anki's handling of `did` on import is outside our control: it may keep or drop an override to a deck it creates] → The spec only promises what the package contains and the round trip into Cardflip; a manual check in Anki is part of the tasks.
- [Card type order of a moved card type] → Nothing to do: templates are written in their current order with `ord` = index, and cards carry their current `ord`.
