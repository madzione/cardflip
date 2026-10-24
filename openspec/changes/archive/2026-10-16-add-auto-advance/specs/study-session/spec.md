# Spec Delta

## ADDED Requirements

### Requirement: Auto advance
The user SHALL be able to turn auto advance on and off for the current session from the study menu and with Shift+A; it SHALL be off when a session starts. While on, after the deck's question time the question action SHALL run, and after the answer time the answer action SHALL run; a time of 0 SHALL disable that side. User input SHALL restart the countdown for the new state. The study screen SHALL show that auto advance is on.

#### Scenario: Hands-free review
- **WHEN** auto advance is on with 5 seconds for the question, 3 for the answer and the answer action "answer Good"
- **THEN** each card's answer appears after 5 seconds, and 3 seconds later the card is rated Good and the next card appears

#### Scenario: Off by default
- **WHEN** a session starts for a deck with auto advance times set
- **THEN** nothing happens until the user turns auto advance on
