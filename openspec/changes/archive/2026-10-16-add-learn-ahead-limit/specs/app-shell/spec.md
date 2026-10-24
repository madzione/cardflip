# Spec Delta

## ADDED Requirements

### Requirement: Learn ahead limit setting
Settings SHALL include the learn ahead limit in minutes, from 0 to 1440, defaulting to 20. The value SHALL persist and be included in backups.

#### Scenario: Turn off learning ahead
- **WHEN** the user sets the learn ahead limit to 0 and restarts the app
- **THEN** the limit is still 0
