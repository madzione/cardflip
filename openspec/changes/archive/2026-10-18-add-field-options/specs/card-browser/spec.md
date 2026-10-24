# Spec Delta

## ADDED Requirements

### Requirement: Sort by sort field
The browser SHALL offer a "Sort field" sort order, which sorts cards by their note's sort field with HTML stripped, case-insensitively.

#### Scenario: Sort by second field
- **WHEN** a note type's sort field is Meaning and the user sorts by "Sort field"
- **THEN** its cards are listed in order of their Meaning field

### Requirement: Fields excluded from unqualified search
Free-text search terms SHALL NOT match text in fields marked "exclude from unqualified searches".

#### Scenario: Notes field excluded
- **WHEN** the Notes field is excluded and contains `irregular`, and the user searches `irregular`
- **THEN** that card is not listed unless another field contains `irregular`

### Requirement: Per-field editor presentation
The editor SHALL show each field in its editing font and size, lay out right-to-left fields right to left, and start fields marked "collapse by default" collapsed, expanding them when tapped.

#### Scenario: Right-to-left field
- **WHEN** the Arabic field is marked right-to-left and the user types in it
- **THEN** the text is aligned right and the cursor moves right to left
