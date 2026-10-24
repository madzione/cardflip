# Spec Delta

## ADDED Requirements

### Requirement: Notes mode
The browser SHALL offer a Notes mode, remembered across restarts, that lists one row per note showing its first field, note type, card count, deck (or "multiple" when its cards are in several decks), tags and the earliest due date of its cards. A note SHALL be listed when any of its cards matches the search. Notes mode SHALL be sortable by created date, first field and earliest due date.

#### Scenario: Reversed note once
- **WHEN** the collection has one `Basic (and reversed card)` note and the user switches to Notes mode
- **THEN** the browser lists one row showing 2 cards

#### Scenario: Match through one card
- **WHEN** only one of a note's cards is suspended and the user searches `is:suspended` in Notes mode
- **THEN** the note is listed

### Requirement: Bulk actions in Notes mode
In Notes mode, bulk actions SHALL apply to all cards of the selected notes.

#### Scenario: Suspend notes
- **WHEN** the user selects 3 reversed notes in Notes mode and suspends them
- **THEN** all 6 cards are suspended
