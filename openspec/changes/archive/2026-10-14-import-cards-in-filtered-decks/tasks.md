# Tasks

## 1. Deck of home-less cards

- [x] 1.1 Add a test to `golang/importer/apkg/apkg_test.go` that builds a `Package` in code with the regular deck `Kana (Hiragana)` (1 card), the filtered decks `Kana (Hiragana)::Basic` (2 cards, `ODeckID` 0) and `Kana (Hiragana)::Advanced` (2 cards, `ODeckID` 0), and a filtered deck `Cram` holding one card whose home deck is `Kana (Hiragana)`. Check, with and without scheduling, that `Convert` gives the card deck names `Kana (Hiragana)` (2), `Kana (Hiragana)::Basic` (2) and `Kana (Hiragana)::Advanced` (2), and that no card is in Default or `Cram`. Verify the test fails before the fix.
- [x] 1.2 In `golang/importer/apkg/convert.go`, give a card with `ODeckID == 0` in a filtered deck that deck's name (design decision 1). Verify `go test ./importer/...` in `golang/` passes, including the existing filtered deck tests of the collection packages.
- [x] 1.3 Add an `Import` test that imports the package from 1.1 into an empty collection and checks the deck tree (`Kana (Hiragana)` with `Basic` and `Advanced`), the card counts per deck, and that the new decks use the Default preset. Verify `go test ./importer/...` passes.

## 2. Integration check

- [x] 2.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import `~/Downloads/Japanese_Hiragana.apkg` in the app and confirm the deck list shows `Kana (Hiragana)` with `Basic` (46) and `Advanced` (53), 5 cards in the parent, and nothing in Default.
