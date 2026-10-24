# Spec Delta

## ADDED Requirements

### Requirement: Change note type
The user SHALL be able to change notes that share one note type to another note type, mapping each target field to one source field or to nothing, and, between standard note types, each target card type to one source card type or to nothing. Cards whose card type is mapped SHALL keep their scheduling and review history. Unmapped source cards SHALL be deleted, and cards the target generates that have no mapped source SHALL be created as New.

#### Scenario: Basic to reversed
- **WHEN** 10 Basic notes are changed to `Basic (and reversed card)` with Front→Front, Back→Back and Card 1→Card 1
- **THEN** the 10 forward cards keep their scheduling and 10 new reverse cards are created as New

#### Scenario: Unmapped field
- **WHEN** a note with fields Front `a`, Back `b`, Extra `c` is changed to Basic with Extra mapped to nothing
- **THEN** the note has Front `a` and Back `b`, and `c` is gone

#### Scenario: Cloze numbers kept
- **WHEN** a Cloze note is changed to another cloze note type with Text→Text
- **THEN** its cards for each cloze number keep their scheduling

### Requirement: Change preview and confirmation
Before notes change type, the app SHALL show how many notes change, how many cards will be deleted and created, and which source fields will lose their content, and SHALL apply the change only after confirmation, as one operation.

#### Scenario: Confirmation counts
- **WHEN** 5 `Basic (and reversed card)` notes are changed to Basic with Card 2 mapped to nothing
- **THEN** the confirmation says 5 notes change and 5 cards will be deleted
