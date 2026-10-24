# Spec Delta

## ADDED Requirements

### Requirement: Marked tag on import as new
When an `.apkg` is imported with "import as new", the tag `marked` SHALL be removed from the imported notes.

#### Scenario: Shared deck with marks
- **WHEN** a package with notes tagged `marked` is imported as New
- **THEN** none of the imported notes has the tag `marked`
