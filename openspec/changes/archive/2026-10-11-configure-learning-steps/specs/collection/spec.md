# Spec Delta

## ADDED Requirements

### Requirement: Step deck options
Each deck SHALL have the options learning steps and relearning steps, written as space-separated durations with units `s`, `m`, `h` or `d` (defaults `1m 10m` and `10m`), interday learning/review order (mix with reviews, before reviews, after reviews; default mix), and bury interday learning siblings (off by default). Steps that are not positive, longer than 365 days, or more than 20 SHALL be rejected with an explanation.

#### Scenario: Valid steps
- **WHEN** the user sets learning steps to `30s 5m 1h 2d`
- **THEN** the steps are saved and shown the same way when the options are reopened

#### Scenario: Invalid step
- **WHEN** the user enters `5x`
- **THEN** the value is rejected with an explanation and the previous steps are kept
