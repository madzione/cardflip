# Spec Delta

## ADDED Requirements

### Requirement: Timer deck options
Each deck SHALL have the options maximum answer seconds (default 60, allowed 1 to 9999), show answer timer (off by default) and stop timer on answer (on by default).

#### Scenario: Out of range
- **WHEN** the user sets maximum answer seconds to 0
- **THEN** the value is rejected and the previous value is kept
