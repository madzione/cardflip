# Spec Delta

## ADDED Requirements

### Requirement: Image Occlusion note type
The app SHALL offer an Image Occlusion note type with the fields Occlusion, Image, Header, Back Extra and Comments. The Occlusion field SHALL store shapes as cloze deletions in Anki's image occlusion syntax, and one card SHALL be generated per shape number, so shapes sharing a number are asked together.

#### Scenario: Three labels
- **WHEN** the user creates an Image Occlusion note with three shapes numbered 1, 2 and 3
- **THEN** three cards are created

#### Scenario: Grouped shapes
- **WHEN** two of those shapes are grouped
- **THEN** the note has two cards, and the grouped shapes are asked together

## MODIFIED Requirements

### Requirement: Built-in note types
A new collection SHALL contain the note types `Basic` (fields Front, Back; one card), `Basic (and reversed card)` (fields Front, Back; two cards), `Cloze` (fields Text, Back Extra; one card per cloze number) and `Image Occlusion` (fields Occlusion, Image, Header, Back Extra, Comments; one card per shape number). A collection created before Image Occlusion was built in SHALL get the note type when the user first adds an Image Occlusion note; when its name is taken by a different note type, a numbered name SHALL be used.

#### Scenario: Built-ins available
- **WHEN** the user opens the note type picker in a fresh collection
- **THEN** Basic, Basic (and reversed card), Cloze and Image Occlusion are listed

#### Scenario: Older collection
- **WHEN** the user adds the first Image Occlusion note in a collection without that note type, which has a Basic note type named `Image Occlusion`
- **THEN** a note type `Image Occlusion 2` is created for it, and later Image Occlusion notes use it too

### Requirement: Create note types
The user SHALL be able to create a note type from a stock type, giving it a name. The stock types SHALL be Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer), Cloze and Image Occlusion. Note type names SHALL be unique, compared case-insensitively.

#### Scenario: Add from stock
- **WHEN** the user adds a note type from stock type Basic named `Vocabulary`
- **THEN** a note type `Vocabulary` with fields Front and Back and one card type exists, and it has no notes

#### Scenario: Name taken
- **WHEN** the user adds a note type named `basic` while `Basic` exists
- **THEN** the creation is rejected with an error that the name is taken

#### Scenario: Image Occlusion from stock
- **WHEN** the user adds a note type from stock type Image Occlusion named `Anatomy`
- **THEN** `Anatomy` has the fields Occlusion, Image, Header, Back Extra and Comments, and its notes are edited with the mask editor
