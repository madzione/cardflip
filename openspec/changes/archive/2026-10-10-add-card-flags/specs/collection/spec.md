# Spec Delta

## ADDED Requirements

### Requirement: Card flags
Each card SHALL have either no flag or exactly one of seven flags: 1 red, 2 orange, 3 green, 4 blue, 5 pink, 6 turquoise, 7 purple. Flags SHALL belong to the card, not to its note. Setting or clearing a flag SHALL NOT change the card's scheduling. Flags SHALL be included in backups.

#### Scenario: Flag one card of a note
- **WHEN** the user flags the forward card of a `Basic (and reversed card)` note red
- **THEN** the forward card has flag 1 and the reverse card has no flag

#### Scenario: Flag kept across restore
- **WHEN** a collection with flagged cards is exported to a backup and restored
- **THEN** every card has the same flag as before

### Requirement: Flag names
The user SHALL be able to rename each of the seven flags in Settings. A flag without a custom name SHALL be shown by its colour name. Clearing a custom name SHALL restore the colour name.

#### Scenario: Rename a flag
- **WHEN** the user renames flag 1 to "Fix later"
- **THEN** the study menu, the browser's flag action and the settings show "Fix later" for the red flag
