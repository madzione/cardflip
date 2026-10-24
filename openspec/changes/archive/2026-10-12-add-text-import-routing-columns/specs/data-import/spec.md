# Spec Delta

## MODIFIED Requirements

### Requirement: CSV duplicate handling
When importing delimited text, a row whose GUID matches an existing note of the same note type SHALL update that note. Otherwise, a row whose first mapped field equals an existing note's first field, for the same note type, or for the same note type and deck when the user chooses that scope, SHALL be handled according to the user's choice: skip it (default), update the existing note, or add it as a duplicate.

#### Scenario: Skip duplicates
- **WHEN** the file contains a row whose Front matches an existing Basic note and "skip" is chosen
- **THEN** no new note is created for that row, and the import summary counts it as skipped

#### Scenario: GUID update
- **WHEN** a row's GUID matches an existing note and its Back field differs
- **THEN** the existing note's Back is updated, whatever the duplicate choice

#### Scenario: Deck scope
- **WHEN** the scope is "note type and deck" and a row's Front matches a Basic note in another deck
- **THEN** the row is added as a new note

## ADDED Requirements

### Requirement: Per-row note type, deck, tags and GUID
Columns SHALL be assignable to note type, deck, tags or GUID, from `#notetype column:`, `#deck column:`, `#tags column:` and `#guid column:` headers or in the import screen. Each row SHALL be imported into its note type, by name, and its deck, created if missing, with field columns mapped by position within that note type. `#notetype:` and `#deck:` headers SHALL set defaults for empty column values. A row naming an unknown note type SHALL be skipped and reported.

#### Scenario: Anki notes export
- **WHEN** a file exported by Anki with deck, note type and GUID columns is imported
- **THEN** each note is created in its own deck with its own note type and GUID

#### Scenario: Unknown note type
- **WHEN** a row names a note type that does not exist
- **THEN** the row is skipped and the summary names the note type
