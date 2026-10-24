# Spec Delta

## Purpose

Lets the user find, inspect and edit any card or note in the collection, independent of studying.

## ADDED Requirements

### Requirement: Card list
The browser SHALL list cards with at least: question text (HTML stripped), deck, due date or state, and note type. It SHALL be sortable by created date, due date and question text, and SHALL stay responsive with 50,000 cards.

#### Scenario: Large collection
- **WHEN** the collection has 50,000 cards and the user opens the browser
- **THEN** the first rows appear within 1 second on a typical desktop, and scrolling remains smooth

### Requirement: Search
The browser SHALL filter by free text, matched case-insensitively against all fields with HTML stripped. It SHALL also support the qualifiers `deck:<name>` (including children), `tag:<tag>`, `is:new`, `is:due`, `is:review`, `is:learn`, `is:suspended` and `note:<note type>`. Several terms SHALL be combined with AND, and a term prefixed with `-` SHALL be negated.

#### Scenario: Combined search
- **WHEN** the user searches `deck:Spanish tag:verb -is:suspended`
- **THEN** only unsuspended cards in Spanish or its subdecks whose note has tag `verb` are listed

#### Scenario: Invalid qualifier
- **WHEN** the user searches `is:bogus`
- **THEN** the browser shows an error for that term instead of an empty list

### Requirement: Card details and editing
Selecting a card SHALL show its rendered question and answer, its scheduling information (state, due, interval, stability, difficulty, reviews, lapses), and an editor for its note's fields and tags. Saves SHALL regenerate cards according to the collection rules.

#### Scenario: Edit from browser
- **WHEN** the user changes a field in the browser editor and saves
- **THEN** the list row and the preview show the new content

### Requirement: Bulk actions
The user SHALL be able to select multiple cards and suspend or unsuspend them, move them to a deck, add or remove tags, reset them to New, or delete their notes. Destructive actions SHALL require confirmation.

#### Scenario: Bulk move
- **WHEN** the user selects 30 cards and moves them to deck `Archive`
- **THEN** all 30 cards are in `Archive` with unchanged scheduling

#### Scenario: Reset to new
- **WHEN** the user resets a Review card to New
- **THEN** the card is New, its memory state is cleared, and its review history is kept

### Requirement: Add notes
The user SHALL be able to add notes from the browser or the deck list, choosing a note type and deck that are remembered for next time. After a successful add, the editor SHALL clear the fields (keeping deck, note type and tags) for rapid entry.

#### Scenario: Rapid entry
- **WHEN** the user adds a note and the add succeeds
- **THEN** the fields are empty, the deck, note type and tags are unchanged, and focus is in the first field
