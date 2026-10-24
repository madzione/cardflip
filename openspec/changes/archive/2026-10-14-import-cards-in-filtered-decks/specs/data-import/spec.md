# Spec Delta

## ADDED Requirements

### Requirement: Cards in filtered decks without a home deck
When an Anki package (`.apkg` or `.colpkg`) has cards in a filtered deck but without a home deck, the import SHALL put those cards in a regular deck with the filtered deck's full name, creating it and its parent decks as needed, with the Default preset. Cards in a filtered deck that have a home deck SHALL still go to their home deck. Such cards SHALL never be placed in Default for lack of a deck.

#### Scenario: Shared deck with filtered subdecks
- **WHEN** the user imports a package with the regular deck `Kana (Hiragana)` holding 5 cards and the filtered decks `Kana (Hiragana)::Basic` and `Kana (Hiragana)::Advanced` holding 46 and 53 cards without a home deck
- **THEN** the deck list shows `Kana (Hiragana)` with the subdecks `Basic` and `Advanced`, holding 5, 46 and 53 cards, and no card of the package is in Default

#### Scenario: Card with a home deck
- **WHEN** a package has a card in a filtered deck whose home deck is `Fixture::Spanish`
- **THEN** the card is imported into `Fixture::Spanish`, and no deck is created for the filtered deck

#### Scenario: Imported as new
- **WHEN** the same package as in "Shared deck with filtered subdecks" is imported without scheduling
- **THEN** the cards are in the same decks, as new cards
