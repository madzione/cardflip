# Spec Delta

## MODIFIED Requirements

### Requirement: Re-importing a package
Importing an `.apkg` whose notes already exist in the collection (matched by Anki's note GUID) SHALL NOT duplicate them. It SHALL update their content when the "Update notes" option is "always", or when it is "if newer" (the default) and the package's note was modified after the collection's; with "never", existing notes SHALL stay unchanged. Existing cards' scheduling SHALL stay unchanged.

#### Scenario: Import twice
- **WHEN** the same `.apkg` is imported twice
- **THEN** the collection holds each note once

#### Scenario: Local edit kept
- **WHEN** the user edited a note after importing a deck, and re-imports the same package with "if newer"
- **THEN** the user's edit is kept

#### Scenario: Updated deck
- **WHEN** the deck's author fixed a note and the user imports the new package with "if newer"
- **THEN** the note shows the author's fix

## ADDED Requirements

### Requirement: Note type updates on import
When a package's note type already exists, the "Update note types" option (if newer, always, never) SHALL decide whether its templates and styling are updated. When the existing and package versions differ in fields or templates and "Merge note types" is on, fields and templates SHALL be matched by name and missing ones added, without removing any; when it is off, the package's version SHALL be imported as a separate note type for the package's notes.

#### Scenario: Merge adds a field
- **WHEN** the package's note type has an extra field `Audio` and merging is on
- **THEN** the existing note type gains the field `Audio`, and the updated notes have its content

#### Scenario: No merge
- **WHEN** the same package is imported with merging off
- **THEN** a new note type is created for the package's notes, and the existing note type is unchanged
