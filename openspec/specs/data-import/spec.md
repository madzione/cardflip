# data-import Specification

## Purpose

Brings cards into the collection from outside sources: delimited text files (CSV/TSV) and Anki deck packages (`.apkg`), on every supported platform including the web.

## Requirements

### Requirement: Choose an import file
The user SHALL be able to pick a `.csv`, `.tsv`, `.txt`, `.apkg` or `.colpkg` file with the platform's file picker. On the web, this SHALL work without uploading the file anywhere: all processing happens on the device.

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
When importing delimited text, a row whose GUID matches an existing note of the same note type SHALL update that note. Otherwise, a row whose first mapped field equals an existing note's first field, for the same note type, or for the same note type and deck when the user chooses that scope, SHALL be handled according to the user's choice: skip it (default), update the existing note, or add it as a duplicate.

#### Scenario: Skip duplicates
- **WHEN** the file contains a row whose Front matches an existing Basic note and "skip" is chosen
- **THEN** no new note is created for that row, and the import summary counts it as skipped

#### Scenario: GUID update
- **WHEN** a row's GUID matches an existing note and its Back field differs
- **THEN** the existing note's Back is updated, whatever the duplicate choice

#### Scenario: Deck scope
- **WHEN** the scope is "note type and deck" and a row's Front matches a Basic note in another deck
- **THEN** the row is added as a new note

### Requirement: Anki package import
The system SHALL import `.apkg` files exported by Anki 2.1.x and later, in both the legacy format (`collection.anki2` / `collection.anki21`) and the current format (`collection.anki21b`), including their decks, note types (fields, templates, CSS), notes, tags, cards and media, including images, audio and video.

#### Scenario: Modern Anki export
- **WHEN** the user imports an `.apkg` exported by current Anki with default settings
- **THEN** its decks, notes and cards appear in Cardflip, and cards render using the original templates

#### Scenario: Legacy export
- **WHEN** the user imports an `.apkg` exported with "support older Anki versions"
- **THEN** the same content is imported as from the modern format

#### Scenario: Audio imported
- **WHEN** the user imports a package containing `hola.mp3` referenced by a note
- **THEN** the file is in the collection's media and the card plays it

### Requirement: Scheduling on Anki import
Before an `.apkg` import, the user SHALL choose between keeping the scheduling and importing all cards as New. If they keep scheduling, new cards stay New, suspended cards stay suspended, and review cards keep their due date, with FSRS memory states taken from Anki when present and estimated from interval and ease otherwise.

#### Scenario: Keep scheduling
- **WHEN** a review card due in 15 days in Anki is imported with "keep scheduling"
- **THEN** it is a Review card due in 15 days in Cardflip

#### Scenario: Reset scheduling
- **WHEN** the same package is imported with "import as new"
- **THEN** every imported card is New

### Requirement: Re-importing a package
Importing an `.apkg` whose notes already exist in the collection (matched by Anki's note GUID) SHALL NOT duplicate them. It SHALL update their content when the "Update notes" option is "always", or when it is "if newer" (the default) and the package's note was modified after the collection's; with "never", existing notes SHALL stay unchanged. Existing cards' scheduling SHALL stay unchanged.

#### Scenario: Import twice
- **WHEN** the same `.apkg` is imported twice
- **THEN** the collection holds each note once

#### Scenario: Local edit kept
- **WHEN** the user edited a note after importing a deck, and re-imports the same package with "if newer"
- **THEN** the user's edit is kept

#### Scenario: Updated deck
- **WHEN** the deck's author fixed a note and the user imports the new package with "if newer"
- **THEN** the note shows the author's fix

### Requirement: Import progress and summary
Imports SHALL show progress and SHALL NOT freeze the UI on any platform. When finished, the app SHALL show a summary of notes added, updated and skipped, cards added, media files imported, and any warnings, such as media files in formats Cardflip cannot play.

#### Scenario: Summary shown
- **WHEN** an import of 2,000 notes with 3 audio files finishes
- **THEN** the summary shows 2,000 notes added and 3 more media files imported, with no warning about audio

### Requirement: Failed imports change nothing
If an import fails, for example because of a corrupt file, an unsupported format or insufficient storage, the collection SHALL be left exactly as it was before the import, and the user SHALL see an error that says what went wrong.

#### Scenario: Corrupt package
- **WHEN** the user imports a truncated `.apkg`
- **THEN** an error explains that the file could not be read, and no decks, notes or media are added

### Requirement: Flags on Anki import
When an `.apkg` is imported with "keep scheduling", newly added cards SHALL keep their Anki flag. When imported as New, cards SHALL have no flag. Cards that already exist SHALL keep their current flag.

#### Scenario: Flag kept
- **WHEN** a package containing a card with the green flag is imported with "keep scheduling"
- **THEN** the imported card has flag 3

#### Scenario: Import as new
- **WHEN** the same package is imported with "import as new"
- **THEN** the imported card has no flag

### Requirement: Review history on Anki import
When an `.apkg` is imported with "keep scheduling", the review history of each card the import adds SHALL be imported with time, rating, kind, intervals and answer time. Anki's manual and rescheduling entries SHALL become manual entries. Importing the same package again SHALL NOT duplicate entries. The summary SHALL report how many review entries were imported.

#### Scenario: History kept
- **WHEN** a package with a card reviewed 12 times in Anki is imported with "keep scheduling"
- **THEN** that card's history in Cardflip has the 12 reviews with their dates and ratings

#### Scenario: Import as new
- **WHEN** the same package is imported with "import as new"
- **THEN** no review history is imported

#### Scenario: Re-import
- **WHEN** the package is imported a second time with "keep scheduling"
- **THEN** the card still has 12 review entries

### Requirement: Memory state from history
When a card imported with its scheduling has no FSRS memory state in Anki but has review history, its memory state SHALL be computed from that history with FSRS.

#### Scenario: Card scheduled with SM-2
- **WHEN** a card reviewed only under Anki's SM-2 scheduler is imported with its 8 reviews
- **THEN** its stability and difficulty result from replaying those 8 reviews, and its due date is kept

### Requirement: Leech tag on import as new
When an `.apkg` is imported with "import as new", the tag `leech` SHALL be removed from the imported notes.

#### Scenario: Shared deck with leeches
- **WHEN** a package with notes tagged `leech` is imported as New
- **THEN** none of the imported notes has the tag `leech`

### Requirement: Presets on Anki import
Importing an `.apkg` that contains deck presets SHALL import them and assign them to the decks the import creates; existing decks SHALL keep their preset. A package preset whose name is taken by a preset with different options SHALL be imported under a new name; one with the same name and options SHALL be reused. Anki's default preset (ID 1) SHALL map to the collection's `Default` preset, as in Anki, whatever its options.

#### Scenario: Shared deck settings
- **WHEN** a package whose deck uses a preset with 10 new cards per day is imported into a collection without that deck
- **THEN** the new deck uses an imported preset with 10 new cards per day

#### Scenario: Anki's default preset
- **WHEN** a package whose decks use Anki's default preset is imported
- **THEN** the new decks use `Default`, and no preset is added

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

### Requirement: Note type updates on import
When a package's note type already exists, the "Update note types" option (if newer, always, never) SHALL decide whether its templates and styling are updated. When the existing and package versions differ in fields or templates and "Merge note types" is on, fields and templates SHALL be matched by name and missing ones added, without removing any; when it is off, the package's version SHALL be imported as a separate note type for the package's notes.

#### Scenario: Merge adds a field
- **WHEN** the package's note type has an extra field `Audio` and merging is on
- **THEN** the existing note type gains the field `Audio`, and the updated notes have its content

#### Scenario: No merge
- **WHEN** the same package is imported with merging off
- **THEN** a new note type is created for the package's notes, and the existing note type is unchanged

### Requirement: Comment lines
In delimited text files, lines that start with `#` after the header lines SHALL be skipped, and the import summary SHALL count them as comments.

#### Scenario: Comment in the middle
- **WHEN** a file has the rows `a,b`, `# chapter 2`, `c,d`
- **THEN** two notes are imported and the summary reports 1 comment line

### Requirement: HTML in text imports
The import screen SHALL offer "Allow HTML in fields", on by default or as set by a `#html:` header. When off, field text SHALL be imported literally, so that `<`, `>` and `&` display as typed and line breaks are kept.

#### Scenario: Literal text
- **WHEN** "Allow HTML in fields" is off and a field is `a < b & c`
- **THEN** the card shows `a < b & c`

### Requirement: Tags header
A `#tags:` header SHALL add its space-separated tags to every note imported from the file.

#### Scenario: Header tags
- **WHEN** a file starts with `#tags:chapter1 vocab`
- **THEN** every imported note has the tags `chapter1` and `vocab`

### Requirement: Separator choice
The `#separator:` header SHALL accept Comma, Semicolon, Tab, Space, Pipe and Colon, by name or as the character. The import screen SHALL show the separator in use and let the user choose another one, updating the preview.

#### Scenario: Override detection
- **WHEN** a file is detected as comma-separated but uses `|` and the user chooses Pipe
- **THEN** the preview shows the columns split at `|`

### Requirement: Per-row note type, deck, tags and GUID
Columns SHALL be assignable to note type, deck, tags or GUID, from `#notetype column:`, `#deck column:`, `#tags column:` and `#guid column:` headers or in the import screen. Each row SHALL be imported into its note type, by name, and its deck, created if missing, with field columns mapped by position within that note type. `#notetype:` and `#deck:` headers SHALL set defaults for empty column values. A row naming an unknown note type SHALL be skipped and reported.

#### Scenario: Anki notes export
- **WHEN** a file exported by Anki with deck, note type and GUID columns is imported
- **THEN** each note is created in its own deck with its own note type and GUID

#### Scenario: Unknown note type
- **WHEN** a row names a note type that does not exist
- **THEN** the row is skipped and the summary names the note type

### Requirement: Cards in filtered decks without a home deck
When an Anki package (`.apkg` or `.colpkg`) has cards in a filtered deck but without a home deck, the import SHALL put those cards in a regular deck with the filtered deck's full name, creating it and its parent decks as needed, with the Default preset. Cards in a filtered deck that have a home deck SHALL still go to their home deck. Such cards SHALL never be placed in Default for lack of a deck.

#### Scenario: Shared deck with filtered subdecks
- **WHEN** the user imports a package with the regular deck `Kana (Hiragana)` holding 5 cards and the filtered decks `Kana (Hiragana)::Basic` and `Kana (Hiragana)::Advanced` holding 46 and 53 cards without a home deck
- **THEN** the deck list shows `Kana (Hiragana)` with the subdecks `Basic` and `Advanced`, holding 5, 46 and 53 cards, and no card of the package is in Default

#### Scenario: Card with a home deck
- **WHEN** a package has a card in a filtered deck whose home deck is `Fixture::Spanish`
- **THEN** the card is imported into `Fixture::Spanish`, and no deck is created for the filtered deck

#### Scenario: Imported as new
- **WHEN** the same package as in "Shared deck with filtered subdecks" is imported without scheduling
- **THEN** the cards are in the same decks, as new cards

### Requirement: Marked tag on import as new
When an `.apkg` is imported with "import as new", the tag `marked` SHALL be removed from the imported notes.

#### Scenario: Shared deck with marks
- **WHEN** a package with notes tagged `marked` is imported as New
- **THEN** none of the imported notes has the tag `marked`

### Requirement: Deck descriptions on Anki import
Importing an `.apkg` SHALL keep the descriptions of the decks it creates. Existing decks SHALL keep their current description.

#### Scenario: Shared deck description
- **WHEN** a shared deck with a description is imported into a collection without that deck
- **THEN** the new deck's overview shows the description

### Requirement: LaTeX image format on Anki import
Importing an `.apkg` SHALL keep each note type's choice between PNG and SVG LaTeX images, and SHALL import SVG media files.

#### Scenario: SVG deck
- **WHEN** a package whose note type uses SVG LaTeX images is imported
- **THEN** its `[latex]` blocks show the imported SVG images

### Requirement: Image Occlusion import
Image Occlusion notes in an `.apkg` SHALL be imported with their note type and image, and their cards SHALL render with masks.

#### Scenario: Imported occlusion deck
- **WHEN** the user imports an anatomy deck made with Anki's Image Occlusion
- **THEN** its cards show the diagram with masks, one card per shape group
