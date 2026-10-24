# Spec Delta

## ADDED Requirements

### Requirement: Undo changes
Every change to the collection made through the app SHALL be undoable, as a named step, in reverse order, for at least the last 30 steps of the current app session. Undoing SHALL restore exactly the previous state of everything the step changed. The undo history SHALL be cleared when the app closes, when a backup is restored, and when a new collection is started.

#### Scenario: Undo a delete
- **WHEN** the user deletes 3 notes in the browser and chooses Undo
- **THEN** the 3 notes and their cards are back with their scheduling, review history and images

#### Scenario: Undo an import
- **WHEN** the user imports a package that adds 200 notes and chooses Undo
- **THEN** the 200 notes, their cards and the decks the import created are gone

#### Scenario: History cleared by restore
- **WHEN** the user restores a backup
- **THEN** there is nothing to undo

### Requirement: Redo
After an undo, and until another change is made, the user SHALL be able to redo the undone step.

#### Scenario: Redo a move
- **WHEN** the user moves cards to `Archive`, undoes it, and chooses Redo
- **THEN** the cards are in `Archive` again
