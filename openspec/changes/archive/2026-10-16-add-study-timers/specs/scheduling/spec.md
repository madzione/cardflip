# Spec Delta

## ADDED Requirements

### Requirement: Capped answer time
The answer time recorded in the review history SHALL be capped at the maximum answer seconds of the card's deck.

#### Scenario: Long pause
- **WHEN** the deck's maximum is 60 seconds and the user answers a card after 5 minutes
- **THEN** the review entry records 60 seconds
