# Spec Delta

## ADDED Requirements

### Requirement: Leech tag on import as new
When an `.apkg` is imported with "import as new", the tag `leech` SHALL be removed from the imported notes.

#### Scenario: Shared deck with leeches
- **WHEN** a package with notes tagged `leech` is imported as New
- **THEN** none of the imported notes has the tag `leech`
