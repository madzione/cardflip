# Spec Delta

## ADDED Requirements

### Requirement: Find and Replace dialog
The browser SHALL offer Find and Replace with the inputs find, replace with, in (all fields, a field of the selected notes' note types, or tags), selected notes only (on when cards are selected), treat input as regular expression, and ignore case. Before applying, it SHALL show how many notes will change.

#### Scenario: Preview count
- **WHEN** the user enters a find text that occurs in 7 of the selected notes
- **THEN** the dialog says 7 notes will change before the user confirms
