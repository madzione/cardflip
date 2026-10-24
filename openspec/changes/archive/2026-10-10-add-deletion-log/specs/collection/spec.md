# Spec Delta

## ADDED Requirements

### Requirement: Deletion log
Whenever notes are deleted, for any reason, each note's deletion time, ID, note type name, fields and tags SHALL be added to a deletion log stored with the collection and included in backups.

#### Scenario: Deleted with a deck
- **WHEN** the user deletes a deck containing 5 notes
- **THEN** the deletion log has 5 new entries with those notes' fields and tags

### Requirement: Export and clear the deletion log
The user SHALL be able to export the deletion log as a tab-separated text file that the text import can read, and to clear the log after a confirmation.

#### Scenario: Recover a note
- **WHEN** the user exports the deletion log and imports the file with the deleted note's note type
- **THEN** the deleted note's fields and tags are back in a new note
