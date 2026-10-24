# Spec Delta

## ADDED Requirements

### Requirement: Selection commands
The browser SHALL offer Invert selection, which selects exactly the listed cards that were not selected, and Select notes, which adds every listed card of the notes of the selected cards to the selection.

#### Scenario: Invert
- **WHEN** 3 of 10 listed cards are selected and the user inverts the selection
- **THEN** the other 7 cards are selected

#### Scenario: Select notes
- **WHEN** the forward card of a reversed note is selected and both cards are listed, and the user chooses Select notes
- **THEN** both cards are selected

### Requirement: Resizable split
On wide screens, the user SHALL be able to drag the divider between the card list and the details, and the position SHALL be remembered.

#### Scenario: Remembered width
- **WHEN** the user drags the divider to make the list wider and reopens the browser
- **THEN** the list has the same width

### Requirement: Keyboard navigation
On platforms with a keyboard, Up and Down SHALL move the current row and show its details, Shift with Up or Down SHALL extend the selection, Home and End SHALL go to the first and last row, Ctrl+F (Cmd+F on macOS) SHALL focus the search field, Enter SHALL open the editor of the current card, and Delete SHALL delete the selected cards' notes after confirmation.

#### Scenario: Browse with arrows
- **WHEN** the list has focus on the first row and the user presses Down twice
- **THEN** the third row is current and its details are shown
