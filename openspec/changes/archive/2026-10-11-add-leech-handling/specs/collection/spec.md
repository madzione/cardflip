# Spec Delta

## ADDED Requirements

### Requirement: Leech options
Each deck SHALL have a leech threshold in lapses (default 8, allowed 1 to 99) and a leech action: tag only (default), or suspend card and tag.

#### Scenario: Options stored
- **WHEN** the user sets the threshold to 5 and the action to suspend
- **THEN** the values persist for the deck and are included in backups
