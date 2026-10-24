# Spec Delta

## MODIFIED Requirements

### Requirement: Due cards and daily limits
A deck's due cards SHALL be: Learning and Relearning cards due by now; Review cards due by the end of the current study day, up to the effective review limit minus reviews already done today; and New cards, up to the effective new limit minus new cards already introduced today. The effective limit SHALL be the deck's today-only limit when set for the current day, otherwise its daily limit. When the deck's "new cards ignore review limit" option is off, new cards SHALL also be limited to the review limit's remaining room after due reviews. Counts for a parent deck SHALL include its children, and each child's own limits SHALL also apply. When studying a subdeck whose "limits start from top" option is on, the limits of its parent decks SHALL also apply.

#### Scenario: New limit reached
- **WHEN** a deck has a new-per-day limit of 20, 50 new cards, and 20 new cards have been introduced today
- **THEN** the deck shows 0 new cards due until the next rollover

#### Scenario: Learning ignores limits
- **WHEN** the review limit for today is used up and a Relearning card becomes due
- **THEN** the Relearning card is still shown

#### Scenario: Today-only increase
- **WHEN** 20 of 20 new cards were introduced today and the user increases today's new card limit by 10
- **THEN** the deck shows 10 new cards due

#### Scenario: Review limit caps new cards
- **WHEN** "new cards ignore review limit" is off, the review limit is 100, 95 reviews are due and the new limit is 20
- **THEN** the deck shows 95 review and 5 new cards

#### Scenario: Limits from the top
- **WHEN** `Languages` has a new limit of 10, `Languages::Spanish` has 20 and "limits start from top" is on for `Spanish`, and the user studies `Spanish`
- **THEN** at most 10 new cards are shown
