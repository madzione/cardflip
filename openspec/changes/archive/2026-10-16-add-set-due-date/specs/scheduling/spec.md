# Spec Delta

## ADDED Requirements

### Requirement: Set due date
The user SHALL be able to set cards to be due on a chosen study day: `N` days from today (`0` = today), or a random day per card in `N-M`. A trailing `!` SHALL also set the card's interval to the days until that due date; without it, a review card's interval SHALL stay unchanged. Invalid input SHALL be rejected with an explanation, and no card SHALL change.

#### Scenario: Postpone
- **WHEN** a review card with a 10-day interval, due today, is set to `7`
- **THEN** it is due in 7 days and its interval is still 10 days

#### Scenario: Range
- **WHEN** 100 new cards are set to `1-10`
- **THEN** they are Review cards, each due on a day from tomorrow to 10 days from now, spread across that range

#### Scenario: Interval too
- **WHEN** a review card is set to `30!`
- **THEN** it is due in 30 days with an interval of 30 days

#### Scenario: Invalid input
- **WHEN** the user enters `10-5`
- **THEN** the input is rejected and no card changes

### Requirement: Card state after setting a due date
Cards whose due date is set SHALL become Review cards, whatever their previous state. Memory states SHALL be kept, and a New card SHALL get the memory state of a first Good rating. Each change SHALL add a manual entry to the card's review history.

#### Scenario: New card scheduled
- **WHEN** a New card is set to `2`
- **THEN** it is a Review card due in 2 days, with a stability and difficulty, and its history has one manual entry
