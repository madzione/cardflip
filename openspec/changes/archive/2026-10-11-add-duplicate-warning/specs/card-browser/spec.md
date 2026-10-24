# Spec Delta

## ADDED Requirements

### Requirement: Duplicate warning
While the user enters a note, the editor SHALL warn when its first field, with HTML stripped and case and surrounding whitespace ignored, equals the first field of another note of the same note type. The warning SHALL NOT prevent adding, SHALL ignore the note being edited and empty first fields, and SHALL offer to show the matching notes in the browser.

#### Scenario: Duplicate word
- **WHEN** a Basic note with Front `perro` exists and the user types `Perro ` into the Front field of a new Basic note
- **THEN** the field is marked and a message says this is a duplicate, with an action to show it

#### Scenario: Different note type
- **WHEN** only a Cloze note contains `perro` as its first field and the user types `perro` into a Basic note
- **THEN** no warning is shown

#### Scenario: Add anyway
- **WHEN** the warning is shown and the user presses Add
- **THEN** the note is added
