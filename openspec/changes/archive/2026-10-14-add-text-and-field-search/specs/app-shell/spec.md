# Spec Delta

## ADDED Requirements

### Requirement: Ignore accents setting
Settings SHALL include "Ignore accents in search", off by default. The value SHALL persist and be included in backups.

#### Scenario: Setting on
- **WHEN** the setting is on and the user searches `cafe`
- **THEN** notes containing `café` are listed
