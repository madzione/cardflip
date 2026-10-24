# Spec Delta

## ADDED Requirements

### Requirement: Readable type-answer comparison
Every part of the type-in-the-answer comparison SHALL be readable in light and dark mode. When the colour a part gets from the note type's CSS has too little contrast with the background it is drawn on, the part SHALL use the card's text colour instead. If that colour is also too faint there, it SHALL use black or white, whichever contrasts more. Colours with enough contrast SHALL be kept.

#### Scenario: Black mark without a valid highlight in dark mode
- **WHEN** the app is in dark mode, the note type CSS has `.typeGood { color: black; background-color: light green; }`, the expected text is `ka` and the user typed `ka`
- **THEN** `ka` shows in the card's text colour, readable on the dark card

#### Scenario: Black mark on its own highlight is kept
- **WHEN** the app is in dark mode, the note type CSS has `.typeBad { color: black; background-color: red; }` and the user typed a wrong character
- **THEN** the wrong character shows black on red

#### Scenario: Unchanged in light mode
- **WHEN** the app is in light mode, the note type CSS has `.typeGood { color: black; }` and the user typed the expected text
- **THEN** the comparison shows black on the light card

#### Scenario: Card colour also too faint
- **WHEN** the note type CSS has `.typeGood { color: #ddd; background-color: #eee; }` and `.card { color: white; }`, and the user typed the expected text
- **THEN** the comparison shows black on the `#eee` highlight

#### Scenario: Rest of the card untouched
- **WHEN** the app is in dark mode and the note type CSS gives a `.hint` element outside the comparison `color: black`
- **THEN** that element still shows black
