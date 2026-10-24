# Spec Delta

## ADDED Requirements

### Requirement: Reposition fields
The user SHALL be able to move a field of a note type to another position. Every note of that type SHALL keep each value in the same named field, and cards SHALL render as before.

#### Scenario: Move Back first
- **WHEN** a Basic note type's field Back is moved before Front
- **THEN** the editor lists Back first, every note's Back value is unchanged, and cards show the same content

### Requirement: Field options
Each field SHALL have the options editing font, editing size, right-to-left, collapse by default, and exclude from unqualified searches. Each note type SHALL have exactly one sort field, the first field by default. Options SHALL follow the field when it is renamed or moved, and SHALL be included in backups and kept on `.apkg` import.

#### Scenario: Sort field follows move
- **WHEN** the sort field is Front and Front is moved to the second position
- **THEN** Front is still the sort field

#### Scenario: Imported options
- **WHEN** an `.apkg` whose note type has a right-to-left field and sorts by its second field is imported
- **THEN** the imported note type has the same right-to-left field and sort field
