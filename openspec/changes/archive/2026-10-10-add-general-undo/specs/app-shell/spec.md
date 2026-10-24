# Spec Delta

## ADDED Requirements

### Requirement: Undo controls
The app SHALL offer Undo and Redo that name the step they act on, through menu items and, outside text fields, Ctrl+Z, Ctrl+Shift+Z and Ctrl+Y (Cmd on macOS). After deleting, moving in bulk or importing, the app SHALL show a message with an Undo action.

#### Scenario: Undo from the snackbar
- **WHEN** the user moves 30 cards and taps Undo in the message that appears
- **THEN** the cards are back in their previous decks

#### Scenario: Named step
- **WHEN** the last change was deleting 3 notes
- **THEN** the Undo menu item reads "Undo Delete 3 notes"
