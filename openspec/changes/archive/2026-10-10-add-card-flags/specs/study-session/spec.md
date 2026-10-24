# Spec Delta

## ADDED Requirements

### Requirement: Flag the current card
While studying, the user SHALL be able to set or clear a flag on the current card from the study menu, and with Ctrl+1 to Ctrl+7 (Cmd on macOS). Choosing the flag the card already has SHALL clear it. The current card's flag SHALL be visible on the study screen as a marker in the flag's colour.

#### Scenario: Toggle with keyboard
- **WHEN** the user presses Ctrl+1 on a card without a flag, and then Ctrl+1 again
- **THEN** the card first shows a red flag marker, and then no marker

#### Scenario: Change flag
- **WHEN** the current card has flag 1 and the user presses Ctrl+3
- **THEN** the card has flag 3
