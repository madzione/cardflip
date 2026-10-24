# Spec Delta

## ADDED Requirements

### Requirement: Find duplicates
The browser SHALL find sets of notes whose chosen field has the same content, comparing with HTML stripped, case ignored and surrounding whitespace ignored, among notes of note types that have a field of that name, optionally limited by a search. Empty fields SHALL NOT be reported. Each set SHALL show its content and note count, and selecting it SHALL list those notes.

#### Scenario: Two duplicates
- **WHEN** two Basic notes have Front `perro` and `Perro ` and the user finds duplicates in Front
- **THEN** one set `perro` with 2 notes is listed

#### Scenario: Limited by search
- **WHEN** the same duplicates exist in decks `A` and `B` and the user limits the search to `deck:A`
- **THEN** only duplicates among notes in `A` are reported

### Requirement: Tag duplicates
From the duplicate results, the user SHALL be able to add the tag `duplicate` to every note in all listed sets.

#### Scenario: Tag for review
- **WHEN** the results contain 3 sets with 7 notes and the user chooses Tag duplicates
- **THEN** those 7 notes have the tag `duplicate`
