# Spec Delta

## ADDED Requirements

### Requirement: Set due date during study
While studying, a menu item and Ctrl+Shift+D (Cmd+Shift+D on macOS) SHALL set the current card's due date. When the card is no longer due today, the session SHALL advance to the next card.

#### Scenario: Push a card to next week
- **WHEN** the user sets the current card's due date to `7`
- **THEN** the next card is shown and the review count decreases by one
