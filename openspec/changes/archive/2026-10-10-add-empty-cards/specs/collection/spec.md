# Spec Delta

## MODIFIED Requirements

### Requirement: Notes generate cards
A note SHALL hold one value per field of its note type, plus tags. For standard note types, the system SHALL create a card for each template whose question renders non-empty and that has no card yet. For cloze note types, it SHALL create a card for each cloze number in the note's fields that has no card yet. Cards whose question becomes empty, or whose cloze number disappears, after an edit SHALL be kept as empty cards with their scheduling until the user removes them with Empty Cards.

#### Scenario: Reversed card
- **WHEN** the user adds a `Basic (and reversed card)` note with Front `hola` and Back `hello`
- **THEN** two cards are created, one asking `hola` and one asking `hello`

#### Scenario: Cloze cards
- **WHEN** the user adds a Cloze note with Text `{{c1::Paris}} is the capital of {{c2::France}}`
- **THEN** two cards are created, for cloze 1 and cloze 2

#### Scenario: Editing removes a cloze
- **WHEN** the user edits that note and removes `{{c2::France}}`
- **THEN** the card for cloze 2 is kept as an empty card, and the card for cloze 1 keeps its scheduling state

#### Scenario: Template typo
- **WHEN** the user changes a card type's front template from `{{Front}}` to `{{Frnt}}` in a note type with 1,000 notes
- **THEN** no card is deleted, and the 1,000 cards are empty cards

## ADDED Requirements

### Requirement: Empty Cards tool
The user SHALL be able to list all empty cards, grouped by note with their card type or cloze number and a total count, and delete them after confirmation in one operation. For a note whose cards are all empty, one card SHALL be kept so that no note is deleted, and the list SHALL say so.

#### Scenario: Remove empty cards
- **WHEN** 3 notes each have one empty card among other cards and the user confirms Empty Cards
- **THEN** those 3 cards are deleted and every note still has its other cards

#### Scenario: All cards empty
- **WHEN** a note's only card is empty
- **THEN** the list marks the note, and deleting empty cards keeps that card
