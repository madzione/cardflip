# Spec Delta

## ADDED Requirements

### Requirement: Mark action and marked rows
The browser SHALL offer "Toggle mark" for the selected cards' notes, and SHALL tint rows of marked notes purple.

#### Scenario: Bulk mark
- **WHEN** the user selects 3 cards of unmarked notes and chooses "Toggle mark"
- **THEN** the 3 notes have the tag `marked` and their rows are tinted purple
