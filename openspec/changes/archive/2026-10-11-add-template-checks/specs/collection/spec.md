# Spec Delta

## ADDED Requirements

### Requirement: Template checks on save
Saving a card type SHALL be refused, with a message naming the problem, when a template has a syntax error (an unclosed, unmatched or crossed section, a `{{` without `}}`, or a tag without a field name), refers to a field that is neither the note type's nor special, or when the front uses no field. A front identical to another card type's SHALL be refused in standard note types, and a front without a `cloze:` filter in cloze note types. A refused save SHALL change nothing.

#### Scenario: Unknown field
- **WHEN** the user saves a front template `{{Frnt}}` in a note type with fields Front and Back
- **THEN** the save is refused with a message that there is no field `Frnt`

#### Scenario: Unclosed section
- **WHEN** the user saves `{{#Hint}}{{Hint}}` without `{{/Hint}}`
- **THEN** the save is refused with a message that `{{#Hint}}` is not closed

#### Scenario: Unclosed tag
- **WHEN** the user saves a back template `{{FrontSide}}<hr id=answer>{{Back`
- **THEN** the save is refused with a message that `{{Back` has no closing `}}`

#### Scenario: Tag without a field name
- **WHEN** the user saves a front template `{{Front}}{{}}`
- **THEN** the save is refused with a message that `{{}}` has no field name

#### Scenario: Crossed sections
- **WHEN** the user saves `{{#Hint}}{{#Extra}}{{/Hint}}{{/Extra}}`
- **THEN** the save is refused with one message that `{{/Hint}}` closes `{{#Hint}}` before `{{#Extra}}` is closed

#### Scenario: Missing cloze filter
- **WHEN** the user saves the front template `{{Text}}` in a Cloze note type
- **THEN** the save is refused with a message that the front must use `{{cloze:Text}}` or another cloze field

#### Scenario: Duplicate front
- **WHEN** a standard note type has two card types and the user makes the second front identical to the first
- **THEN** the save is refused with a message that both card types would show the same question

### Requirement: Live template feedback
The template editor SHALL show the result of the save checks while the user edits, before saving.

#### Scenario: Error shown while typing
- **WHEN** the user types `{{Frnt}}` into the front template
- **THEN** the editor shows that there is no field `Frnt` without the user saving

### Requirement: Reserved field names
Adding a field or renaming a field to `Tags`, `Type`, `Deck`, `Card` or `FrontSide`, compared case-insensitively, SHALL be refused with an explanation.

#### Scenario: Reserved name
- **WHEN** the user adds a field named `tags`
- **THEN** the field is not added and the message says the name is reserved

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
- **WHEN** the user mistakenly changes a card type's front template from `{{Front}}` to `{{Back}}` in a note type with 1,000 notes whose Back is empty
- **THEN** no card is deleted, and the 1,000 cards are empty cards
