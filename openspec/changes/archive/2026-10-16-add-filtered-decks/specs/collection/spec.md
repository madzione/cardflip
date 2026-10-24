# Spec Delta

## ADDED Requirements

### Requirement: Deleting filtered decks
Deleting a filtered deck SHALL return its cards to their home decks and SHALL NOT delete any card. Deleting a normal deck SHALL also delete its cards that are currently in filtered decks, and the confirmation SHALL count them.

#### Scenario: Delete filtered deck
- **WHEN** the user deletes a filtered deck holding 20 cards
- **THEN** no card is deleted and the 20 cards are in their home decks
