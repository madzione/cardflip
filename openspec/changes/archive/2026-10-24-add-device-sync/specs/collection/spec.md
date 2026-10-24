# Spec Delta

## MODIFIED Requirements

### Requirement: Undo changes
Every change to the collection made through the app SHALL be undoable, as a named step, in reverse order, for at least the last 30 steps of the current app session. Undoing SHALL restore exactly the previous state of everything the step changed. The undo history SHALL be cleared when the app closes, when a backup is restored, and when a new collection is started. Changes received by device sync SHALL NOT be undo steps; when they change or delete anything that a step in the undo or redo history changed, the history SHALL be cleared. Otherwise it SHALL be kept. Undoing a step that was already synced SHALL be synced as a new change.

#### Scenario: Undo a delete
- **WHEN** the user deletes 3 notes in the browser and chooses Undo
- **THEN** the 3 notes and their cards are back with their scheduling, review history and images

#### Scenario: Undo an import
- **WHEN** the user imports a package that adds 200 notes and chooses Undo
- **THEN** the 200 notes, their cards and the decks the import created are gone

#### Scenario: History cleared by restore
- **WHEN** the user restores a backup
- **THEN** there is nothing to undo

#### Scenario: Unrelated synced change
- **WHEN** the user edits note A on the laptop, and a sync then brings an edit of note B from the phone
- **THEN** Undo on the laptop still reverts the edit of note A, and note B keeps the phone's edit

#### Scenario: Synced change to the same note
- **WHEN** the user edits note A on the laptop, and a sync then brings a later edit of note A from the phone
- **THEN** there is nothing to undo on the laptop

#### Scenario: Undo after sync
- **WHEN** the user deletes a note on the laptop, the deletion syncs to the phone, and the user then chooses Undo on the laptop
- **THEN** the note is back on the laptop, and after the next sync it is back on the phone too
