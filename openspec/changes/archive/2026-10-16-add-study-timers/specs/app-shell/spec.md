# Spec Delta

## ADDED Requirements

### Requirement: Timebox setting
Settings SHALL include the timebox time limit in minutes, where 0 (the default) turns it off. The value SHALL persist and be included in backups.

#### Scenario: Enable timebox
- **WHEN** the user sets the timebox to 15 minutes and restarts the app
- **THEN** the setting is still 15 minutes
