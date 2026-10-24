# Spec Delta

## ADDED Requirements

### Requirement: New card order options
Each deck SHALL have the options new card gather order (deck, deck then random notes, ascending position, descending position, random notes, random cards; default ascending position) and new card sort order (order gathered, card type then order gathered, card type then random, random note then card type, random; default order gathered).

#### Scenario: New card defaults
- **WHEN** a new deck is created
- **THEN** its new card gather order is ascending position and its sort order is order gathered

### Requirement: Review order options
Each deck SHALL have the options new/review order (mix, before reviews, after reviews; default mix) and review sort order (due date then random, due date then deck, deck then due date, ascending intervals, descending intervals, ascending difficulty, descending difficulty, ascending retrievability, relative overdueness; default due date then random).

#### Scenario: Review defaults
- **WHEN** a new deck is created
- **THEN** its new/review order is mix and its review sort order is due date then random
