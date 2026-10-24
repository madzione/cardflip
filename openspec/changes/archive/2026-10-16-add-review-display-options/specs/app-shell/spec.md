# Spec Delta

## ADDED Requirements

### Requirement: Review settings
Settings SHALL include a Review section with: show remaining card count (on), show next review time above answer buttons (on), Space/Enter also answers Good (on), hide bars during review (off), and question fade in milliseconds (100, from 0 to 1000). The values SHALL persist and be included in backups.

#### Scenario: Setting persists
- **WHEN** the user turns off "Show remaining card count" and restarts the app
- **THEN** the setting is still off
