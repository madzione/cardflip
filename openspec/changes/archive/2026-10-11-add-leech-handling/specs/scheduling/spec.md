# Spec Delta

## ADDED Requirements

### Requirement: Leech detection
When rating a Review card Again brings its lapse count to the deck's leech threshold, or to the threshold plus a multiple of half the threshold rounded up, the card's note SHALL get the tag `leech`, and with the suspend action the card SHALL be suspended, in the same operation as the answer.

#### Scenario: Threshold reached
- **WHEN** the threshold is 8, the action is suspend, and a card with 7 lapses is rated Again
- **THEN** the card has 8 lapses, is suspended, and its note has the tag `leech`

#### Scenario: Repeat warning
- **WHEN** the threshold is 8, the action is tag only, and the card reaches 12 lapses
- **THEN** it is detected as a leech again
