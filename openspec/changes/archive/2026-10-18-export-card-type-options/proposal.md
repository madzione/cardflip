# Proposal

## Why

`add-card-type-options` added a deck override and browser appearance templates per card type, and the `.apkg` import reads them. The package export still writes empty `bqfmt`/`bafmt` and a null `did` for every card type, so both options are lost when a deck is shared or the collection is moved to Anki or exported and imported again.

## What Changes

- `.apkg` and `.colpkg` export write each card type's browser question and answer templates (`bqfmt`/`bafmt`), empty when the card type has none.
- Export writes each card type's deck override (`did`) as the Anki ID the export gives that deck, when the deck is part of the export. When the deck is not exported (a deck or note export that does not contain it), `did` stays null and the card type has no override in the package.
- Both the current and the older ("support older Anki versions") package formats carry the options, since both are written from the same note type JSON.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-export`: new requirement that packages keep card type options.

## Impact

- Go: `golang/exporter/apkg/apkg.go`: `model()` takes the card type options and the exporter's deck ID mapping (`exporter.deckIDs`).
- Tests: `golang/exporter/apkg/apkg_test.go` (round trip through `importer/apkg`, override to a deck outside the export).
- No change to the Dart API, the UI or the import side.
- Overlap: depends on `add-card-type-options` (the `CardTemplate` fields and their import). Apply it after that change.
