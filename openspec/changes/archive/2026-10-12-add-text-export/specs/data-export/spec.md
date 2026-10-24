# Spec Delta

## Purpose

Exporting decks, selected notes and the whole collection to files that other spaced-repetition apps, in particular Anki, can import.

## ADDED Requirements

### Requirement: Notes as plain text
The user SHALL be able to export notes of a deck and its subdecks, of a selection, or of the whole collection to a tab-separated file with one note per line and fields as stored, optionally adding tags, deck, note type and GUID columns. The file SHALL start with headers describing the separator, HTML and the added columns. Fields containing tabs, quotes or line breaks SHALL be quoted.

#### Scenario: Export with tags
- **WHEN** the user exports deck `Spanish` as notes with tags
- **THEN** each line has the note's fields followed by its tags, and the header names the tags column

#### Scenario: Line break in a field
- **WHEN** a field contains a line break
- **THEN** that field is quoted, and importing the file gives the same field content

### Requirement: Cards as plain text
The user SHALL be able to export cards to a tab-separated file with one card per line holding its rendered question and answer as plain text.

#### Scenario: Printable list
- **WHEN** the user exports `Spanish` as cards
- **THEN** each line holds a card's question and answer without HTML

### Requirement: Text export round trip
Importing a notes text file exported by Cardflip, with the duplicate option "update", SHALL update the original notes instead of creating new ones, as long as their first field is unchanged.

#### Scenario: Edit outside and re-import
- **WHEN** the user exports notes, edits the Back field of one note in a spreadsheet and imports the file with "update"
- **THEN** the original note shows the edit and no note is duplicated
