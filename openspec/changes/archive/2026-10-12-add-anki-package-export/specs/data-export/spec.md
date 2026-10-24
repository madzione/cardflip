# Spec Delta

## Purpose

Exporting decks, selected notes and the whole collection to files that other spaced-repetition apps, in particular Anki, can import.

## ADDED Requirements

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

### Requirement: Saving exports
On native platforms the user SHALL choose where to save the exported file; on the web the file SHALL be downloaded. Export SHALL show progress and SHALL NOT freeze the interface.

#### Scenario: Web export
- **WHEN** the user exports a deck in the web build
- **THEN** the browser downloads `Spanish.apkg`
