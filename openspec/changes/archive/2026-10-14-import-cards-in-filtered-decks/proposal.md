# Proposal

## Why

Importing the shared deck `Japanese_Hiragana.apkg` puts the cards in the wrong decks. The package has the decks `Kana (Hiragana)`, `Kana (Hiragana)::Basic` (46 cards) and `Kana (Hiragana)::Advanced` (53 cards). After the import, 99 of its 104 cards are in **Default** and only 5 are in `Kana (Hiragana)`; the two subdecks do not exist.

The `::` in the names is not the cause: Cardflip creates `A::B` hierarchies correctly. The cause is that both subdecks are stored as **filtered decks** (`dyn: 1`) in the package, and their cards have no home deck (`odid = 0`). This happens with older shared decks whose author had built them as filtered decks before sharing. Cardflip's importer only looks for a card's deck among regular decks, and a card in a filtered deck is moved to its home deck. Without a home deck, the card falls back to Default.

Verified by running the package through `apkg.Read` and `apkg.Convert`: `"Default": 99, "Kana (Hiragana)": 5`.

## What Changes

- A card in an Anki filtered deck that has no home deck is imported into a regular deck with the filtered deck's name, keeping its place in the deck hierarchy. For the Hiragana package this gives `Kana (Hiragana)` (5 cards), `Kana (Hiragana)::Basic` (46) and `Kana (Hiragana)::Advanced` (53).
- Cards in filtered decks that do have a home deck keep going to their home deck, as now.
- A filtered deck that holds no home-less cards is still not created.
- A deck created this way uses the Default preset, since a filtered deck has no options preset of its own to map.
- The same rule applies to `.apkg` and `.colpkg` imports, with and without scheduling.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `data-import`: new requirement for cards that sit in an Anki filtered deck without a home deck.

## Impact

- Go: `golang/importer/apkg/convert.go` (`Convert`'s deck lookup; `ConvertCollection`'s deck list).
- Tests: `golang/importer/apkg/apkg_test.go` (a package built in the test with a filtered deck holding home-less cards, converted and imported).
- Not affected: Flutter, the reader (`reader.go` already reads `dyn` and `odid` for both formats).
- Overlap: `add-filtered-decks` brings real filtered decks to Cardflip and says the `.apkg` import keeps putting cards from filtered decks into their home decks. Whichever lands second makes sure that home-less cards still go to a regular deck, also when the target collection already has a filtered deck of that name.
