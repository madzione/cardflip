# Spec Delta

## MODIFIED Requirements

### Requirement: Add cards
The user SHALL be able to add cards from the browser or the deck list, choosing a note type and deck that are remembered for next time. After a successful add, the editor SHALL clear the fields that are not pinned, keeping pinned fields, deck, note type and tags, for rapid entry. Pins SHALL be remembered per note type.

#### Scenario: Rapid entry
- **WHEN** the user adds a card and the add succeeds
- **THEN** the fields are empty, the deck, note type and tags are unchanged, and focus is in the first field

#### Scenario: Pinned field kept
- **WHEN** the Source field is pinned, the user adds a card with Source `Book 1`, and the add succeeds
- **THEN** Source still contains `Book 1` and the other fields are empty

## ADDED Requirements

### Requirement: Add shortcut
On platforms with a keyboard, Ctrl+Enter (Cmd+Enter on macOS) in any field SHALL add the card in the add screen and save the note in the edit screen.

#### Scenario: Add from the keyboard
- **WHEN** the cursor is in the Back field and the user presses Ctrl+Enter
- **THEN** the card is added as if Add was pressed

### Requirement: Deck per note type
Settings SHALL offer "Deck for new cards" with the choices "Current deck" (default) and "Change deck depending on note type". With the second choice, choosing a note type in the add screen SHALL select the deck last used to add a card of that note type.

#### Scenario: Deck follows note type
- **WHEN** the setting is "Change deck depending on note type", the user last added Cloze cards to `Biology` and Basic cards to `Spanish`, and switches the note type to Cloze
- **THEN** the deck switches to `Biology`
