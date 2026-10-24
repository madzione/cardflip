# Spec Delta

## MODIFIED Requirements

### Requirement: Order of new cards
New cards SHALL be introduced in the order they were added to the collection (import order for imported cards). Once a card of a note has been studied today, the note's other New cards SHALL NOT be shown that day when their deck buries new siblings, and its other Review cards SHALL NOT be shown that day when their deck buries review siblings (sibling burying).

#### Scenario: Sibling buried
- **WHEN** the forward card of a reversed note is studied today
- **THEN** the reverse card is not shown again until the next study day

#### Scenario: Review siblings allowed
- **WHEN** a deck does not bury review siblings and both cards of a reversed note are due for review today
- **THEN** both cards are shown today
