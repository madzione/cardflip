# Spec Delta

## ADDED Requirements

### Requirement: Empty cards during study
When the card being studied is an empty card, the study screen SHALL show a notice that the card is empty and can be removed with Empty Cards, together with the normal reveal and rating controls.

#### Scenario: Empty cloze card
- **WHEN** the card for a removed cloze number comes up for review
- **THEN** the screen says the card is empty and how to remove it, and the user can still rate it
