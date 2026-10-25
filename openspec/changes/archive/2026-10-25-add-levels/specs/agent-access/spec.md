# Spec Delta

## MODIFIED Requirements

### Requirement: Collection statistics tool
The server SHALL offer a tool returning, for the collection or one deck, today's studied cards and time, the counts of new, learning, review and suspended cards, and the number of cards due on each of the next 30 days. It SHALL also return the collection's current streak, whether today is a study day, the daily goal with today's progress, or that no goal is set, and the collection's level with its XP and the XP needed for the next level.

#### Scenario: Deck forecast
- **WHEN** an agent asks for the statistics of deck `Spanish`
- **THEN** it gets those counts for `Spanish` and its subdecks only

#### Scenario: Goal progress
- **WHEN** the goal is 30 cards, 20 cards were answered today and the streak is 12 days, and an agent asks for statistics
- **THEN** it gets a current streak of 12 days, today studied, and goal progress of 20 of 30 cards

#### Scenario: Level
- **WHEN** the user has 3,760 XP and an agent asks for the statistics of deck `Spanish`
- **THEN** it gets the collection's level 12, 3,760 XP and 3,900 XP for level 13
