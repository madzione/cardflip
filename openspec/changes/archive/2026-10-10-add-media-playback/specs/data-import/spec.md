# Spec Delta

## MODIFIED Requirements

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

### Requirement: Import progress and summary
Imports SHALL show progress and SHALL NOT freeze the UI on any platform. When finished, the app SHALL show a summary of notes added, updated and skipped, cards added, media files imported, and any warnings, such as media files in formats Cardflip cannot play.

#### Scenario: Summary shown
- **WHEN** an import of 2,000 notes with 3 audio files finishes
- **THEN** the summary shows 2,000 notes added and 3 more media files imported, with no warning about audio
