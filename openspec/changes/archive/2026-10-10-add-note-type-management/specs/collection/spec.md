# Spec Delta

## ADDED Requirements

### Requirement: Create note types
The user SHALL be able to create a note type from a stock type, giving it a name. The stock types SHALL be Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer) and Cloze. Note type names SHALL be unique, compared case-insensitively.

#### Scenario: Add from stock
- **WHEN** the user adds a note type from stock type Basic named `Vocabulary`
- **THEN** a note type `Vocabulary` with fields Front and Back and one card type exists, and it has no notes

#### Scenario: Name taken
- **WHEN** the user adds a note type named `basic` while `Basic` exists
- **THEN** the creation is rejected with an error that the name is taken

### Requirement: Clone note types
The user SHALL be able to clone a note type, which SHALL create a new note type with a new name and the same kind, fields, card types and styling, without notes.

#### Scenario: Clone imported type
- **WHEN** the user clones an imported note type with 4 fields and 2 card types as `Experiment`
- **THEN** `Experiment` has the same 4 fields, 2 card types and CSS, and no notes

### Requirement: Delete note types
The user SHALL be able to delete a note type, which SHALL delete all its notes and their cards. The app SHALL ask for confirmation stating how many notes and cards will be deleted. The last remaining note type SHALL NOT be deletable.

#### Scenario: Delete with notes
- **WHEN** the user deletes a note type that has 10 notes with 20 cards and confirms
- **THEN** the note type, its 10 notes and 20 cards are gone

#### Scenario: Last note type
- **WHEN** only one note type exists and the user tries to delete it
- **THEN** the deletion is rejected and the note type remains

### Requirement: Optional reversed card
The stock type Basic (optional reversed card) SHALL have the fields Front, Back and Add Reverse, and SHALL generate the reverse card only when Add Reverse is not empty.

#### Scenario: Reverse on demand
- **WHEN** the user adds a note of that type with Front `hola`, Back `hello` and Add Reverse `y`
- **THEN** two cards are created, and a note with an empty Add Reverse gets only the forward card
