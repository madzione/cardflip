# Spec Delta

## ADDED Requirements

### Requirement: Interval limits
No review interval, including the Easy interval and fuzz, SHALL exceed the deck's maximum interval. A card graduating from relearning SHALL get at least the deck's minimum interval after relearning. The interval previews SHALL reflect both limits.

#### Scenario: Capped interval
- **WHEN** the deck's maximum interval is 180 days and a mature card's Easy interval would be 400 days
- **THEN** Easy shows and gives 180 days

#### Scenario: Relearned card
- **WHEN** the minimum interval after relearning is 3 days and a relearning card graduates with a computed interval of 1 day
- **THEN** it is due in 3 days
