# Spec Delta

## ADDED Requirements

### Requirement: Multi-number clozes generate cards
For cloze note types, every number listed in a multi-number deletion SHALL count as a cloze number of the note when cards are generated.

#### Scenario: Two cards from one deletion
- **WHEN** the user adds a Cloze note with Text `{{c1,2::Paris}} is the capital of France`
- **THEN** two cards are created, for cloze 1 and cloze 2

#### Scenario: Removing a number
- **WHEN** that note is edited to `{{c1::Paris}} is the capital of France`
- **THEN** the card for cloze 2 is kept as an empty card, and the card for cloze 1 keeps its scheduling state
