# Spec Delta

## ADDED Requirements

### Requirement: Reposition new cards
The user SHALL be able to give selected New cards consecutive positions from a start position with a step, optionally in random order, and optionally shifting the positions of other new cards at or after the start so no positions are shared. Cards that are not New SHALL be left unchanged.

#### Scenario: Move a chapter to the front
- **WHEN** the user repositions 20 new cards with start 0, step 1 and shifting on
- **THEN** those 20 cards are introduced before all other new cards, in their previous relative order

#### Scenario: Mixed selection
- **WHEN** the selection contains 5 new and 3 review cards
- **THEN** only the 5 new cards are repositioned and the app says 3 cards were not new

### Requirement: Insertion order
Each deck SHALL have an insertion order option, sequential (default) or random. With random, cards added to the deck SHALL get random positions among its new cards. Changing the option SHALL reposition the deck's existing new cards: shuffled for random, in order added for sequential.

#### Scenario: Random insertion
- **WHEN** a deck's insertion order is random and 100 cards are imported into it
- **THEN** the new cards are not introduced in the import order

#### Scenario: Back to sequential
- **WHEN** the option is changed back to sequential
- **THEN** the deck's new cards are introduced in the order they were added
