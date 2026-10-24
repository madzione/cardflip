# Spec Delta

## Purpose

Brings cards into the collection from outside sources: delimited text files (CSV/TSV) and Anki deck packages (`.apkg`), on every supported platform including the web.

## ADDED Requirements

### Requirement: Choose an import file
The user SHALL be able to pick a `.csv`, `.tsv`, `.txt` or `.apkg` file with the platform's file picker. On the web, this SHALL work without uploading the file anywhere: all processing happens on the device.

#### Scenario: Web import is local
- **WHEN** the user imports an `.apkg` in the web build
- **THEN** the import completes with no network request carrying the file

### Requirement: CSV/TSV import with preview
For delimited text, the system SHALL detect the delimiter (comma, tab or semicolon) and the encoding (UTF-8, with or without BOM), and show a preview of the first rows. The user SHALL then choose the note type, the target deck, the mapping from columns to fields or tags, and whether the first row is a header.

#### Scenario: Tab-separated file
- **WHEN** the user picks a file whose lines are `front<TAB>back`
- **THEN** the preview shows two columns, mapped by default to Front and Back of note type Basic

#### Scenario: Quoted fields
- **WHEN** a CSV line is `"a, b","line1\nline2"` with an embedded newline inside quotes
- **THEN** it imports as one note with Front `a, b` and Back containing both lines

### Requirement: CSV duplicate handling
When importing delimited text, a row whose first mapped field equals an existing note's first field (for the same note type) SHALL be handled according to the user's choice: skip it (default), update the existing note, or add it as a duplicate.

#### Scenario: Skip duplicates
- **WHEN** the file contains a row whose Front matches an existing Basic note and "skip" is chosen
- **THEN** no new note is created for that row, and the import summary counts it as skipped

### Requirement: Anki package import
The system SHALL import `.apkg` files exported by Anki 2.1.x and later, in both the legacy format (`collection.anki2` / `collection.anki21`) and the current format (`collection.anki21b`), including their decks, note types (fields, templates, CSS), notes, tags, cards and image media.

#### Scenario: Modern Anki export
- **WHEN** the user imports an `.apkg` exported by current Anki with default settings
- **THEN** its decks, notes and cards appear in Cardflip, and cards render using the original templates

#### Scenario: Legacy export
- **WHEN** the user imports an `.apkg` exported with "support older Anki versions"
- **THEN** the same content is imported as from the modern format

### Requirement: Scheduling on Anki import
Before an `.apkg` import, the user SHALL choose between keeping the scheduling and importing all cards as New. If they keep scheduling, new cards stay New, suspended cards stay suspended, and review cards keep their due date, with FSRS memory states taken from Anki when present and estimated from interval and ease otherwise.

#### Scenario: Keep scheduling
- **WHEN** a review card due in 15 days in Anki is imported with "keep scheduling"
- **THEN** it is a Review card due in 15 days in Cardflip

#### Scenario: Reset scheduling
- **WHEN** the same package is imported with "import as new"
- **THEN** every imported card is New

### Requirement: Re-importing a package
Importing an `.apkg` whose notes already exist in the collection (matched by Anki's note GUID) SHALL update those notes' content instead of duplicating them, and SHALL leave the existing cards' scheduling unchanged.

#### Scenario: Import twice
- **WHEN** the same `.apkg` is imported twice
- **THEN** the collection holds each note once

### Requirement: Import progress and summary
Imports SHALL show progress and SHALL NOT freeze the UI on any platform. When finished, the app SHALL show a summary of notes added, updated and skipped, cards added, media files imported, and any warnings, such as unsupported media.

#### Scenario: Summary shown
- **WHEN** an import of 2,000 notes with 3 audio files finishes
- **THEN** the summary shows 2,000 notes added and a warning that 3 audio files were skipped

### Requirement: Failed imports change nothing
If an import fails, for example because of a corrupt file, an unsupported format or insufficient storage, the collection SHALL be left exactly as it was before the import, and the user SHALL see an error that says what went wrong.

#### Scenario: Corrupt package
- **WHEN** the user imports a truncated `.apkg`
- **THEN** an error explains that the file could not be read, and no decks, notes or media are added
