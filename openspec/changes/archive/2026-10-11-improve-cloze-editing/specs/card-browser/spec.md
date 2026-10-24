# Spec Delta

## ADDED Requirements

### Requirement: Cloze insertion shortcuts
In cloze note types, Ctrl+Shift+C (Cmd+Shift+C on macOS) SHALL wrap the selection in a deletion with the next unused cloze number of the note, and Ctrl+Alt+Shift+C (Cmd+Option+Shift+C on macOS) SHALL use the highest number already in the note, or 1 when there is none. Numbers SHALL be taken from all fields of the note. The same two actions SHALL be available from the toolbar.

#### Scenario: Next number
- **WHEN** the Text field contains `{{c1::Paris}} is in France`, the user selects `France` and presses Ctrl+Shift+C
- **THEN** the field contains `{{c1::Paris}} is in {{c2::France}}`

#### Scenario: Same number
- **WHEN** the user instead presses Ctrl+Alt+Shift+C
- **THEN** the field contains `{{c1::Paris}} is in {{c1::France}}`

#### Scenario: Number from another field
- **WHEN** Back Extra contains `{{c3::x}}`, Text has no clozes, and the user presses Ctrl+Shift+C in Text
- **THEN** the new deletion in Text is `c4`
