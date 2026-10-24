# Spec Delta

## ADDED Requirements

### Requirement: Typing the answer
When the question contains a `type:` filter, the study screen SHALL show a text input in its place and focus it. Enter in the input SHALL reveal the answer, as SHALL "Show answer". While the input has focus, typed characters SHALL go into the input and SHALL NOT trigger study shortcuts. After the answer is revealed, the rating shortcuts SHALL work again.

#### Scenario: Type and reveal
- **WHEN** the user types `pero` into the input and presses Enter
- **THEN** the answer is shown with the comparison of `pero` against the expected text, together with the rating buttons

#### Scenario: Digits are typed
- **WHEN** the input has focus and the user presses 3
- **THEN** `3` is typed into the input and the card is not rated
