# Spec Delta

## ADDED Requirements

### Requirement: Event searches
The search SHALL support `added:N`, `edited:N`, `rated:N`, `rated:N:R`, `introduced:N` and `resched:N`, matching cards added, whose note was edited, answered (with rating R from 1 to 4), first answered, or manually rescheduled within the last N study days, where 1 means today. `rated:` SHALL accept N up to 365. `prop:rated` and `prop:resched` SHALL compare the study day of the latest answer or manual reschedule with today.

#### Scenario: Added today
- **WHEN** the user added 12 cards today and searches `added:1`
- **THEN** those 12 cards are listed

#### Scenario: Failed this week
- **WHEN** the user searches `rated:7:1`
- **THEN** cards rated Again in the last 7 study days are listed

#### Scenario: Before rollover
- **WHEN** the rollover hour is 04:00, a card was added yesterday at 02:00, and the user searches `added:1` today at 10:00
- **THEN** that card is not listed, because it belongs to the previous study day

#### Scenario: Out of range
- **WHEN** the user searches `rated:400`
- **THEN** an error says the number of days must be at most 365

### Requirement: Recently added history
The add screen SHALL offer a History action that opens the browser with the search `added:1`.

#### Scenario: Review what was added
- **WHEN** the user adds 5 cards and chooses History
- **THEN** the browser opens with `added:1` and lists those cards
