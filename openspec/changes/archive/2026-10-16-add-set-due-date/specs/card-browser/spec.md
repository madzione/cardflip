# Spec Delta

## ADDED Requirements

### Requirement: Set due date action
The browser SHALL offer "Set due date" for the selected cards, with an input that accepts the set-due-date syntax, explains it, and remembers the last value.

#### Scenario: Bulk postpone
- **WHEN** the user selects 20 cards and sets the due date to `3`
- **THEN** all 20 cards are due in 3 days
