# Spec Delta

## ADDED Requirements

### Requirement: Study deck picker
Pressing `/` outside a text field, or choosing "Study deck" on the deck list, SHALL open a picker that filters decks by their full names, matching all typed words in any order, case-insensitively, and lists recently studied decks first when nothing is typed. Choosing a deck SHALL start studying it.

#### Scenario: Find a subdeck
- **WHEN** the user presses `/`, types `span verb` and presses Enter
- **THEN** studying starts for `Languages::Spanish::Verbs`

#### Scenario: Recent decks
- **WHEN** the user last studied `Biology` and opens the picker without typing
- **THEN** `Biology` is listed first

#### Scenario: Typing in a field
- **WHEN** the cursor is in a note field and the user types `/`
- **THEN** a `/` is inserted and no picker opens
