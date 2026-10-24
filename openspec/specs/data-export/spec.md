# data-export Specification

## Purpose

Exporting decks, selected notes and the whole collection to files that other spaced-repetition apps, in particular Anki, can import.

## Requirements

### Requirement: Deck package export
The user SHALL be able to export a deck and its subdecks to an `.apkg` file, with the options include scheduling information (off by default), include deck presets (off by default), include media (on by default) and support older Anki versions (off by default). The file SHALL import into current Anki and, with the last option, into Anki 2.1.x.

#### Scenario: Share a deck
- **WHEN** the user exports deck `Spanish` with default options and a friend imports the file into Anki
- **THEN** the friend has the deck `Spanish` with all its notes, note types, images and audio, and every card is new

#### Scenario: With scheduling
- **WHEN** the user exports with scheduling and imports the file into Anki
- **THEN** review cards keep their due dates, intervals and review history

### Requirement: Collection package export
The user SHALL be able to export the whole collection to a `.colpkg` file, with scheduling and history, with or without media, in the current or the older format.

#### Scenario: Back to Anki
- **WHEN** the user exports the collection and opens it in Anki
- **THEN** Anki shows all decks, notes and cards with their scheduling

### Requirement: Export round trip
Notes exported to a package SHALL keep their GUIDs, so importing the package into Cardflip again SHALL update the existing notes instead of duplicating them.

#### Scenario: Re-import own export
- **WHEN** the user exports `Spanish` and imports the file into the same collection
- **THEN** no note is duplicated

### Requirement: Export of card type options
A package export SHALL keep each card type's browser appearance templates. It SHALL keep each card type's deck override when that deck is part of the export, naming it as the package's copy of that deck; otherwise the card type SHALL have no override in the package. This SHALL hold for both package formats.

#### Scenario: Round trip with an override
- **WHEN** card type `Production` overrides to deck `Spanish::Production` and has the browser question template `{{Back}}`, the user exports deck `Spanish` and imports the file into another collection
- **THEN** the imported `Production` card type overrides to the imported `Spanish::Production` deck and has the browser question template `{{Back}}`

#### Scenario: Override deck not exported
- **WHEN** card type `Production` overrides to deck `Production` and the user exports deck `Spanish`, which does not contain it
- **THEN** the package's `Production` card type has no deck override, and its browser appearance templates are kept

#### Scenario: Older format
- **WHEN** the user exports with "support older Anki versions"
- **THEN** the package keeps the card type options in the same way

### Requirement: Saving exports
On native platforms the user SHALL choose where to save the exported file; on the web the file SHALL be downloaded. Export SHALL show progress and SHALL NOT freeze the interface.

#### Scenario: Web export
- **WHEN** the user exports a deck in the web build
- **THEN** the browser downloads `Spanish.apkg`

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
