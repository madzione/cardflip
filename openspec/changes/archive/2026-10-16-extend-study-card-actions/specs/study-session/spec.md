# Spec Delta

## ADDED Requirements

### Requirement: Note actions during study
While studying, the user SHALL be able to bury all cards of the current note, suspend all cards of the current note, and delete the current note with all its cards after a confirmation that states how many cards will be deleted. Each SHALL take effect immediately and advance to the next card.

#### Scenario: Suspend note
- **WHEN** the user chooses Suspend note on the forward card of a `Basic (and reversed card)` note
- **THEN** both cards are suspended and the next card is shown

#### Scenario: Delete note
- **WHEN** the user chooses Delete note on a card whose note has 2 cards and confirms
- **THEN** the note and both cards are deleted and the next card is shown

### Requirement: Create copy during study
While studying, the user SHALL be able to open the add screen prefilled with the current note's note type, deck, fields and tags. Closing the add screen SHALL return to the same card.

#### Scenario: Copy and adjust
- **WHEN** the user chooses Create copy, changes the Front field and adds the card
- **THEN** a new note is added and the session shows the same card as before

### Requirement: Deck options during study
While studying, `O` and a menu item SHALL open the options of the deck being studied. Changed options SHALL apply to the rest of the session.

#### Scenario: Lower the new card limit
- **WHEN** the user opens the options during study and lowers new cards per day below the number already introduced today
- **THEN** after closing the options, no further new cards are shown in the session

### Requirement: Card action keys
On platforms with a keyboard, `E` SHALL edit the current card, `-` SHALL bury it, `@` SHALL suspend it, `=` SHALL bury its note, `!` SHALL suspend its note, and Ctrl+Delete (Cmd+Delete on macOS) SHALL delete its note after confirmation.

#### Scenario: Bury with key
- **WHEN** the user presses `-` on a card
- **THEN** the card is buried and the next card is shown
