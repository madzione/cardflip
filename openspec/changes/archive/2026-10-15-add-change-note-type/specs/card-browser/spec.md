# Spec Delta

## ADDED Requirements

### Requirement: Change note type action
The browser SHALL offer "Change note type" for a selection whose notes all share one note type, with a dialog to choose the target note type and map fields and card types, defaulting to equal names and then positions. For a selection with several note types, the action SHALL explain that the notes must share one note type.

#### Scenario: Mixed selection
- **WHEN** the selection contains Basic and Cloze notes and the user chooses "Change note type"
- **THEN** the app explains that all selected cards must use the same note type, and nothing changes
