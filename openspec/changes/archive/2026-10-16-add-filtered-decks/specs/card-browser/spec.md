# Spec Delta

## ADDED Requirements

### Requirement: Filtered deck from search
The browser SHALL offer "Create filtered deck" prefilled with the current search.

#### Scenario: From search
- **WHEN** the search is `tag:exam is:due` and the user chooses Create filtered deck
- **THEN** the filtered deck options open with the search `tag:exam is:due`

### Requirement: Filtered deck search
`deck:filtered` SHALL match cards that are currently in a filtered deck.

#### Scenario: Cards away from home
- **WHEN** 50 cards are in filtered decks and the user searches `deck:filtered`
- **THEN** those 50 cards are listed
