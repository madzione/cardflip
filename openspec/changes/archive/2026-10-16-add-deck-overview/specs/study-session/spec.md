# Spec Delta

## MODIFIED Requirements

### Requirement: Start a session
The user SHALL be able to start studying any deck that has due cards, from the deck's overview or directly from the deck list's study action. The session SHALL include the due cards of the deck and its children, according to the scheduling rules. Starting a session directly for a deck with nothing due SHALL show a "congratulations, nothing due" screen that says when the next card is due.

#### Scenario: Nothing due
- **WHEN** the user starts studying a deck with no due cards from the deck list's study action and the next card is due tomorrow
- **THEN** the app shows that studying is complete for now and that the next card is due tomorrow

#### Scenario: Study from the overview
- **WHEN** the user chooses Study now on a deck's overview
- **THEN** the session starts with that deck's due cards
