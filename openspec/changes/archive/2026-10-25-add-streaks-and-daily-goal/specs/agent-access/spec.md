# Spec Delta

## MODIFIED Requirements

### Requirement: Collection statistics tool
The server SHALL offer a tool returning, for the collection or one deck, today's studied cards and time, the counts of new, learning, review and suspended cards, and the number of cards due on each of the next 30 days. It SHALL also return the collection's current streak, whether today is a study day, and the daily goal with today's progress, or that no goal is set.

#### Scenario: Deck forecast
- **WHEN** an agent asks for the statistics of deck `Spanish`
- **THEN** it gets those counts for `Spanish` and its subdecks only

#### Scenario: Goal progress
- **WHEN** the goal is 30 cards, 20 cards were answered today and the streak is 12 days, and an agent asks for statistics
- **THEN** it gets a current streak of 12 days, today studied, and goal progress of 20 of 30 cards
