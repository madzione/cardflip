# Spec Delta

## ADDED Requirements

### Requirement: Editing buttons in the field toolbar
The buttons that change a field's content, Cloze deletion (in cloze note types), Math, Attach and Record audio, SHALL be together, in that order, at the end of the field toolbar, set apart from the formatting buttons, in the add and edit screens and the browser's edit pane, and nowhere else in those editors. They SHALL act on the focused field, or the last focused one, and SHALL stay enabled while it shows its HTML source.

#### Scenario: Basic note type
- **WHEN** the user opens the Add card screen with the Basic note type
- **THEN** the field toolbar ends with Math, Attach and Record audio, and the app bar has the Note types button and none of these

#### Scenario: Cloze note type
- **WHEN** the user switches the Add card screen to the Cloze note type
- **THEN** the field toolbar ends with Cloze deletion, Math, Attach and Record audio

#### Scenario: Same group when editing in the browser
- **WHEN** the user selects a card of a cloze note in the browser
- **THEN** the edit pane's field toolbar ends with Cloze deletion, Math, Attach and Record audio, and the row with Save has only Save

#### Scenario: Cloze in the browser
- **WHEN** the Text field in the browser's edit pane contains `{{c1::Paris}} is in France`, the user selects `France` and chooses Cloze deletion
- **THEN** the field contains `{{c1::Paris}} is in {{c2::France}}`

#### Scenario: HTML source
- **WHEN** the Back field shows its HTML source, the formatting buttons are disabled, and the user chooses inline math and then attaches `cat.jpg`
- **THEN** Math and Attach are enabled and the Back field's source contains `\(\)` and `<img src="cat.jpg">`
