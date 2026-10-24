# Spec Delta

## MODIFIED Requirements

### Requirement: Order of new cards
New cards SHALL be selected for the day according to the studied deck's new card gather order, and shown according to its new card sort order; with the defaults, they are introduced in the order they were added to the collection (import order for imported cards). Random orders SHALL stay the same within a study day. Cards of the same note SHALL NOT be shown on the same day once one of them has been studied (sibling burying).

#### Scenario: Sibling buried
- **WHEN** the forward card of a reversed note is studied today
- **THEN** the reverse card is not shown again until the next study day

#### Scenario: Random notes
- **WHEN** the gather order is "random notes" and the new limit is 10 of 500 new cards
- **THEN** today's 10 new cards come from randomly chosen notes, and the selection does not change after answering one
