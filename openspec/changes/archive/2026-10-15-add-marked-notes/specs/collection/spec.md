# Spec Delta

## ADDED Requirements

### Requirement: Marked notes
Toggling the mark on notes SHALL add the tag `marked` to all of them when any lacks it, and otherwise remove it from all of them. The tag SHALL behave as any other tag.

#### Scenario: Mark a note
- **WHEN** the user toggles the mark on a note without the `marked` tag
- **THEN** the note has the tag `marked`
