# Spec Delta

## ADDED Requirements

### Requirement: Audio deck options
Each deck SHALL have the options "Don't play audio automatically" and "Skip question when replaying answer", both off by default.

#### Scenario: Options persist
- **WHEN** the user turns on "Don't play audio automatically" for a deck and restarts the app
- **THEN** the option is still on for that deck and is included in backups
