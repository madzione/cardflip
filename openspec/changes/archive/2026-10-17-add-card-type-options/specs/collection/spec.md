# Spec Delta

## ADDED Requirements

### Requirement: Reposition card types
The user SHALL be able to move a card type of a standard note type to another position. Every existing card SHALL keep its scheduling and stay attached to the same card type.

#### Scenario: Swap two card types
- **WHEN** a note type has card types `Recognition` and `Production` and the user moves `Production` to the first position
- **THEN** `Production` is listed first, and every card keeps its card type, deck and scheduling

### Requirement: Card type deck override
A card type SHALL optionally name a deck. When a note is added or edited so that a new card of that card type is generated, that card SHALL be placed in the named deck instead of the note's deck. Existing cards SHALL NOT be moved. When the named deck no longer exists, the note's deck SHALL be used.

#### Scenario: Production cards to their own deck
- **WHEN** the card type `Production` overrides to deck `Spanish::Production` and the user adds a note in deck `Spanish`
- **THEN** the `Recognition` card is in `Spanish` and the `Production` card is in `Spanish::Production`

#### Scenario: Override deck deleted
- **WHEN** the override deck has been deleted and the user adds a note
- **THEN** all its cards are in the deck chosen in the editor

### Requirement: Import of card type options
Importing an `.apkg` SHALL keep each card type's deck override when that deck is imported or exists, and its browser appearance templates.

#### Scenario: Imported override
- **WHEN** a package's card type overrides to a deck that the package also contains
- **THEN** the imported card type overrides to that deck
