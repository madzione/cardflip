# Spec Delta

## ADDED Requirements

### Requirement: Custom study limits
Each deck's menu SHALL offer Custom study, with choices to increase today's new card limit and today's review card limit by a number, showing how many more cards are available for each.

#### Scenario: Ten more new cards
- **WHEN** the user opens Custom study for `Spanish`, chooses to increase today's new card limit by 10 and confirms
- **THEN** the deck list shows up to 10 more new cards for `Spanish` today, and its options are unchanged
