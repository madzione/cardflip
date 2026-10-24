# Spec Delta

## MODIFIED Requirements

### Requirement: Card queue order
Within a session, cards due from learning SHALL take priority once due. Otherwise review cards and new cards SHALL be interleaved, shown before, or shown after each other according to the deck's new/review order, and review cards SHALL follow its review sort order. The session SHALL show remaining counts for new, learning and review cards, and update them after every answer.

#### Scenario: Counts decrease
- **WHEN** the session shows 5 review cards remaining and the user answers a review card Good
- **THEN** the review count shows 4

#### Scenario: New cards after reviews
- **WHEN** the new/review order is "after reviews" and a session has 10 review and 5 new cards
- **THEN** all 10 review cards are shown before the first new card

#### Scenario: Most at risk first
- **WHEN** the review sort order is "ascending retrievability"
- **THEN** due review cards are shown starting with the one least likely to be recalled
