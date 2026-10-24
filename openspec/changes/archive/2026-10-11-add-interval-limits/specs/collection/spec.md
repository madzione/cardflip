# Spec Delta

## ADDED Requirements

### Requirement: Interval limit options
Each deck SHALL have the options maximum interval (default 36,500 days) and minimum interval after relearning (default 1 day), each allowed from 1 to 36,500 days, with the minimum not above the maximum.

#### Scenario: Minimum above maximum
- **WHEN** the user sets the maximum interval to 30 and the minimum interval after relearning to 60
- **THEN** the change is rejected with an explanation
