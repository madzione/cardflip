# Spec Delta

## MODIFIED Requirements

### Requirement: Order of new cards
New cards SHALL be selected for the day according to the studied deck's new card gather order, and shown according to its new card sort order; with the defaults, they are introduced in the order they were added to the collection (import order for imported cards). Random orders SHALL stay the same within a study day. Once a card of a note has been studied today, the note's other New cards SHALL NOT be shown that day when their deck buries new siblings, and its other Review cards SHALL NOT be shown that day when their deck buries review siblings (sibling burying).

#### Scenario: Sibling buried
- **WHEN** a deck buries new siblings and the forward card of a new reversed note is studied today
- **THEN** the reverse card is not shown again until the next study day

#### Scenario: Siblings by default
- **WHEN** a deck uses the default options and the first card of a new cloze note with three deletions is studied
- **THEN** the note's other two cards can be shown on the same day

#### Scenario: Review siblings allowed
- **WHEN** a deck does not bury review siblings and both cards of a reversed note are due for review today
- **THEN** both cards are shown today

#### Scenario: Random notes
- **WHEN** the gather order is "random notes" and the new limit is 10 of 500 new cards
- **THEN** today's 10 new cards come from randomly chosen notes, and the selection does not change after answering one
