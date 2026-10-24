# Spec Delta

## ADDED Requirements

### Requirement: Leech message and undo
When an answer makes a card a leech, the study screen SHALL show a message saying so, and that the card was suspended when it was. Undoing that answer SHALL also undo the suspension and remove the `leech` tag if that answer added it.

#### Scenario: Message
- **WHEN** an answer makes the current card a leech with the suspend action
- **THEN** a message says the card is a leech and was suspended

#### Scenario: Undo leech
- **WHEN** the user undoes that answer
- **THEN** the card is unsuspended, its note no longer has the tag `leech`, and the card is shown again
