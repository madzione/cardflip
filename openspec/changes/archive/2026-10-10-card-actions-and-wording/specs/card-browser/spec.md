# Spec Delta

## MODIFIED Requirements

### Requirement: Card details and editing
Selecting a card SHALL show its rendered question and answer, its scheduling information (state, due, interval, stability, difficulty, reviews, lapses), and an editor for its fields and tags. The editor SHALL come directly after the question and answer, before the scheduling information. Saves SHALL regenerate cards according to the collection rules. When the card's note has other cards, the editor SHALL say that changes also apply to them.

#### Scenario: Edit from browser
- **WHEN** the user changes a field in the browser editor and saves
- **THEN** the list row and the preview show the new content

#### Scenario: Shared note is disclosed
- **WHEN** the user opens a card whose `Basic (and reversed card)` note has two cards
- **THEN** the editor says that changes also apply to 1 other card

### Requirement: Bulk actions
The user SHALL be able to select multiple cards and suspend or unsuspend them, move them to a deck, add or remove tags, reset them to New, or delete them. The Browse toolbar SHALL have a visible control that starts selecting, and long-press on a row SHALL also start selecting. Destructive actions SHALL require confirmation.

#### Scenario: Bulk move
- **WHEN** the user selects 30 cards and moves them to deck `Archive`
- **THEN** all 30 cards are in `Archive` with unchanged scheduling

#### Scenario: Reset to new
- **WHEN** the user resets a Review card to New
- **THEN** the card is New, its memory state is cleared, and its review history is kept

#### Scenario: Select without long-press
- **WHEN** the user clicks the Select button in the Browse toolbar and then clicks two rows
- **THEN** both cards are selected and the bulk actions are available

### Requirement: Add cards
The user SHALL be able to add cards from the browser or the deck list, choosing a note type and deck that are remembered for next time. After a successful add, the editor SHALL clear the fields (keeping deck, note type and tags) for rapid entry.

#### Scenario: Rapid entry
- **WHEN** the user adds a card and the add succeeds
- **THEN** the fields are empty, the deck, note type and tags are unchanged, and focus is in the first field

## RENAMED Requirements

- FROM: `### Requirement: Add notes`
- TO: `### Requirement: Add cards`

## ADDED Requirements

### Requirement: Single-card actions
The card details SHALL offer Edit, Delete, and Suspend or Unsuspend (whichever applies to the card's state), on wide and narrow layouts. Edit SHALL bring the editor into view and focus its first field. Suspend and Unsuspend SHALL take effect immediately, and the details SHALL show the new state.

#### Scenario: Unsuspend from details
- **WHEN** the user opens a suspended card in the browser and chooses Unsuspend
- **THEN** the card is no longer suspended, its row no longer says "suspended", and the action changes to Suspend

#### Scenario: Edit action
- **WHEN** the user chooses Edit in the card details
- **THEN** the editor is visible and its first field has focus

### Requirement: Deleting cards
Deleting a card SHALL delete its note with all of that note's cards. Before deleting, the browser SHALL ask for confirmation that states how many cards will be deleted, and SHALL say separately how many of them the user did not pick. After deleting from the details, the details SHALL close.

#### Scenario: Delete a single card
- **WHEN** the user chooses Delete in the details of a `Basic` card and confirms
- **THEN** the card is deleted, it disappears from the list, and the details close

#### Scenario: Delete reaches a sibling card
- **WHEN** the user deletes one card of a `Basic (and reversed card)` note
- **THEN** the confirmation says 2 cards will be deleted, including 1 other card from the same note

#### Scenario: Cancel delete
- **WHEN** the user chooses Delete and then cancels the confirmation
- **THEN** no card is deleted
