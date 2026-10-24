# Spec Delta

## ADDED Requirements

### Requirement: Find and replace in notes
The system SHALL replace text in a set of notes, in all fields, in one named field, or in tags. Matching SHALL be literal unless regular expressions are requested, in which case the replacement MAY refer to capture groups. Matching SHALL ignore case unless the user turns that off. All changes SHALL be applied as one operation, and cards SHALL be regenerated as for an edit.

#### Scenario: Fix a spelling
- **WHEN** the user replaces `recieve` with `receive` in all fields of 200 notes, 15 of which contain it
- **THEN** those 15 notes contain `receive` and the report says 15 notes changed

#### Scenario: Regex with group
- **WHEN** the user replaces `(\d+) kg` with `$1 kilograms` using regular expressions in the field Back
- **THEN** `5 kg` in Back becomes `5 kilograms`, and other fields are unchanged

#### Scenario: Rename a tag pattern
- **WHEN** the user replaces `ch0` with `chapter` in tags
- **THEN** the tag `ch01` becomes `chapter1`

#### Scenario: Invalid expression
- **WHEN** the user enters `(` with regular expressions on
- **THEN** the replacement is refused with an error and nothing changes
