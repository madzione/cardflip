# Spec Delta

## MODIFIED Requirements

### Requirement: Choose an import file
The user SHALL be able to pick a `.csv`, `.tsv`, `.txt`, `.apkg` or `.colpkg` file with the platform's file picker. On the web, this SHALL work without uploading the file anywhere: all processing happens on the device.

#### Scenario: Web import is local
- **WHEN** the user imports an `.apkg` in the web build
- **THEN** the import completes with no network request carrying the file

## ADDED Requirements

### Requirement: Collection package import
Importing a `.colpkg` SHALL replace the current collection with the package's decks, note types, notes, tags, cards with scheduling, and media, after a confirmation that warns the current data will be overwritten. Cardflip's app settings SHALL be kept. Cards in Anki filtered decks SHALL be placed in their home decks. If the import fails, the current collection SHALL remain unchanged.

#### Scenario: Move from Anki
- **WHEN** the user imports a `.colpkg` exported by current Anki and confirms
- **THEN** the deck list shows the Anki collection's decks, including empty ones, with their due counts

#### Scenario: Cancel
- **WHEN** the user picks a `.colpkg` and cancels the confirmation
- **THEN** nothing changes

#### Scenario: Corrupt package
- **WHEN** the `.colpkg` is truncated
- **THEN** an error explains the file could not be read, and the current collection is unchanged
