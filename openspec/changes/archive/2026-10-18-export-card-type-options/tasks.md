# Tasks

## 1. Export

- [x] 1.1 In `golang/exporter/apkg/apkg.go`, write each template's `bqfmt`/`bafmt` from `BrowserQFmt`/`BrowserAFmt` (empty when unset) and `did` as `exporter.deckIDs[DeckOverride]` when that deck is exported, else null. Verify `go test ./exporter/...` passes, `TestColumnJSONMatchesAnki` included.
- [x] 1.2 Add `golang/exporter/apkg/apkg_test.go` tests: a deck export of `Spanish` whose `Production` card type overrides to `Spanish::Production` and has a browser question template round-trips through `importer/apkg` into a new collection with both options (current and older format); an override to a deck outside the export reads back from the package with `DeckID` 0 and the browser templates kept; an override to the Default deck maps to Anki's deck 1. Verify they pass.

## 2. Integration check

- [x] 2.1 Run `go vet ./...` and `go test ./...` in `golang/`, and verify all pass.
- [x] 2.2 Export a deck with an override and browser templates, import the `.apkg` into Anki, and check that the card type's browser appearance shows the templates and the deck override names the imported deck. Record the result (including whether Anki keeps the override) in this task.
  - Result (2026-10-18, Anki 26.09.3 through its `anki` Python package, which uses the same backend as the desktop app; current and older format, imported into a fresh collection): the browser question template `{{Back}}` and the empty answer template arrive for every package. The override does not survive Anki's import. Anki gives the imported decks new IDs and keeps the template's `did` unchanged, so it names no deck: the package's `Spanish::Production` (1791892800006) becomes deck 1792308933985 while the template keeps 1791892800006. An override to the Default deck (`did` 1) works, because Anki's deck 1 keeps its ID. An override to a deck outside the export arrives as null. Anki's own export of the same setup (deck `Spanish`, `Production` overridden to `Spanish::Production`) behaves the same when imported into Anki, so the package is correct and Anki drops the override itself. The cards themselves land in `Spanish::Production`. Cardflip's import restores the override (`TestCardTypeOptionsRoundTrip`).

## Workflow follow-up

- Apply after `add-card-type-options` is archived. When done, archive with `/opsx:archive export-card-type-options` and tick its box in `ROADMAP.md`.
