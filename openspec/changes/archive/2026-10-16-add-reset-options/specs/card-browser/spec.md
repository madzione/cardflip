# Spec Delta

## ADDED Requirements

### Requirement: Reset dialog
Resetting cards in the browser SHALL show the reset options with their defaults before applying the reset to the selected cards.

#### Scenario: Options shown
- **WHEN** the user selects 5 cards and chooses Reset
- **THEN** a dialog shows "Restore original position where possible" on and "Reset repetition and lapse counts" off, and nothing changes until the user confirms
