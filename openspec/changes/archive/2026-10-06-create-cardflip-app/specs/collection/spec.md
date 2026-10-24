# Spec Delta

## Purpose

Defines the user's flash card collection: decks, note types, notes, the cards generated from them, tags, and the operations that create, change, suspend and delete them.

## ADDED Requirements

### Requirement: Hierarchical decks
The system SHALL let the user create, rename, move and delete decks. A deck name containing `::` denotes nesting (for example `Languages::Spanish`), and parent decks that do not exist SHALL be created automatically. Deck names SHALL be unique among siblings, compared case-insensitively.

#### Scenario: Create nested deck
- **WHEN** the user creates a deck named `Languages::Spanish` and no deck `Languages` exists
- **THEN** the collection contains a deck `Languages` with a child deck `Spanish`

#### Scenario: Duplicate sibling name rejected
- **WHEN** the user renames a deck to a name that a sibling already uses, ignoring case
- **THEN** the rename is rejected with an error explaining that the name is taken, and nothing changes

### Requirement: Default deck
The collection SHALL always contain at least one deck. A new collection SHALL start with a deck named `Default`. The last remaining deck cannot be deleted.

#### Scenario: Fresh collection
- **WHEN** the app opens a collection for the first time
- **THEN** the deck list shows exactly one deck named `Default`

#### Scenario: Deleting the last deck
- **WHEN** the user tries to delete the only deck
- **THEN** the deletion is rejected and the deck remains

### Requirement: Deck deletion
Deleting a deck SHALL delete its child decks and all cards in them. Notes left with no cards SHALL be deleted. The user SHALL confirm a deletion that removes one or more cards, and the confirmation SHALL state how many cards will be deleted.

#### Scenario: Delete deck with cards
- **WHEN** the user deletes a deck that, together with its children, contains 12 cards
- **THEN** the app asks for confirmation mentioning 12 cards, and after confirmation those decks and cards are gone

### Requirement: Deck options
Each deck SHALL have options for new cards per day (default 20), maximum reviews per day (default 200), and desired retention (default 0.90, allowed range 0.70–0.99). A child deck SHALL use its own options independently of its parent.

#### Scenario: Out-of-range retention
- **WHEN** the user sets desired retention to 0.5
- **THEN** the value is rejected and the previous value is kept

### Requirement: Built-in note types
A new collection SHALL contain the note types `Basic` (fields Front, Back; one card), `Basic (and reversed card)` (fields Front, Back; two cards) and `Cloze` (fields Text, Back Extra; one card per cloze number).

#### Scenario: Built-ins available
- **WHEN** the user opens the note type picker in a fresh collection
- **THEN** Basic, Basic (and reversed card) and Cloze are listed

### Requirement: Note types
A note type SHALL have a name, an ordered list of uniquely named fields, one or more card templates (each a name, a question format and an answer format), shared styling (CSS), and a kind, either standard or cloze. Cloze note types SHALL have exactly one template. The user SHALL be able to rename a note type and its fields, and to add or remove fields and templates.

#### Scenario: Remove a field
- **WHEN** the user removes field `Back` from a note type with existing notes
- **THEN** after confirmation the field and its content are removed from all notes of that type

### Requirement: Notes generate cards
A note SHALL hold one value per field of its note type, plus tags. For standard note types, the system SHALL keep exactly one card per template whose question renders non-empty. For cloze note types, it SHALL keep one card per distinct cloze number in the note's fields.

#### Scenario: Reversed card
- **WHEN** the user adds a `Basic (and reversed card)` note with Front `hola` and Back `hello`
- **THEN** two cards are created, one asking `hola` and one asking `hello`

#### Scenario: Cloze cards
- **WHEN** the user adds a Cloze note with Text `{{c1::Paris}} is the capital of {{c2::France}}`
- **THEN** two cards are created, for cloze 1 and cloze 2

#### Scenario: Editing removes a cloze
- **WHEN** the user edits that note and removes `{{c2::France}}`
- **THEN** the card for cloze 2 is deleted, and the card for cloze 1 keeps its scheduling state

### Requirement: Empty note rejected
The system SHALL reject adding a note that would generate no cards, and SHALL explain why.

#### Scenario: Empty front
- **WHEN** the user tries to add a Basic note with an empty Front field
- **THEN** the note is not added and the editor reports that the note would produce no cards

### Requirement: Tags
Notes SHALL carry zero or more tags. Tags are compared case-insensitively, contain no whitespace, and may use `::` for hierarchy. The system SHALL be able to list all tags in use.

#### Scenario: Add tag with spaces
- **WHEN** the user enters the tag text `verb irregular`
- **THEN** the note gets two tags, `verb` and `irregular`

### Requirement: Suspend and bury
The user SHALL be able to suspend and unsuspend cards. Suspended cards are never shown for study. The user SHALL be able to bury a card until the next day.

#### Scenario: Suspended card skipped
- **WHEN** a due card is suspended
- **THEN** it does not appear in any study session or due count until it is unsuspended

#### Scenario: Buried card returns
- **WHEN** a card is buried during a study session
- **THEN** it is excluded from study for the rest of the current day and is eligible again after the next day rollover

### Requirement: Move cards between decks
The user SHALL be able to move one or more cards to another deck without changing their scheduling state.

#### Scenario: Move preserves schedule
- **WHEN** a review card due in 10 days is moved to another deck
- **THEN** it is still due in 10 days, now in the target deck

### Requirement: Change notifications
Whenever the collection changes, including changes made by import, restore or study, every open screen that shows affected data SHALL refresh without the user taking action.

#### Scenario: Counts update after import
- **WHEN** an import finishes while the deck list is visible
- **THEN** the deck list shows the new decks and updated due counts
