# card-info Specification

## Purpose

The card info view: a card's identity, current scheduling and memory state, and its complete review history, for understanding how a card has been scheduled.

## Requirements

### Requirement: Card summary
Card info SHALL show the card's added date, first and latest review dates, due date or new-card position, interval, stability, difficulty, current retrievability, review count, lapse count, average and total answer time, card type, note type, deck, and card and note IDs. Values that do not apply, such as stability for a New card, SHALL be shown as absent rather than zero.

#### Scenario: Review card
- **WHEN** the user opens card info for a Review card reviewed 5 times with 1 lapse
- **THEN** it shows 5 reviews, 1 lapse, its stability, difficulty and retrievability, and its due date

#### Scenario: New card
- **WHEN** the user opens card info for a New card
- **THEN** it shows its position among new cards and no stability, difficulty or retrievability

### Requirement: Review history
Card info SHALL list every review log entry of the card, newest first, with date and time, kind (learning, review, relearning or manual), rating, interval after the entry, and answer time.

#### Scenario: History after studying
- **WHEN** a new card was rated Good, then Good again 10 minutes later
- **THEN** card info lists two learning entries with rating Good, the later one first, and the second with an interval of at least one day
