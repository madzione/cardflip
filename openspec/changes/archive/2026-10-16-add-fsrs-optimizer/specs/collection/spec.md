# Spec Delta

## ADDED Requirements

### Requirement: FSRS parameter options
Each deck SHALL have FSRS parameters (empty meaning default parameters), a search that selects the reviews used for optimizing (default: the deck and its subdecks), an optional date before which reviewed cards are ignored, and a historical retention (default 0.90, allowed 0.50 to 0.99).

#### Scenario: Default parameters
- **WHEN** a deck has never been optimized
- **THEN** its parameters are empty and the default parameters are used

### Requirement: Reschedule on change
Each deck SHALL have the option "Reschedule cards on change", off by default. When it is on and the deck's desired retention or parameters change, the due dates of its review cards SHALL be recomputed from their review history with the new values, after a confirmation stating how many cards will be rescheduled.

#### Scenario: Lower retention reschedules
- **WHEN** the option is on and the user lowers desired retention from 0.90 to 0.85 and confirms
- **THEN** the deck's review cards get later due dates

#### Scenario: Option off
- **WHEN** the option is off and the user changes desired retention
- **THEN** no due date changes until cards are answered
