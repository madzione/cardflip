# Spec Delta

## ADDED Requirements

### Requirement: Card info during study
While studying, `I` and a menu item SHALL open card info for the current card, and Ctrl+Alt+I (Cmd+Option+I on macOS) SHALL open it for the previously answered card. Opening card info SHALL NOT change the session, and closing it SHALL return to the same card in the same revealed state.

#### Scenario: Previous card
- **WHEN** the user rates a card Again and presses Ctrl+Alt+I
- **THEN** card info for that card opens, and its history includes the Again rating

#### Scenario: Return unchanged
- **WHEN** the user opens card info on a revealed answer and closes it
- **THEN** the same answer is still shown with the rating buttons
