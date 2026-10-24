# Spec Delta

## ADDED Requirements

### Requirement: Card terminology
On the deck list, Browse, add and edit screens and the study screen, user-facing text SHALL call what the user adds, edits, suspends and deletes a "card", including buttons, menu items, titles, tooltips, confirmations and editor error messages. "Note type" SHALL stay as the name of the template that decides which cards are made.

#### Scenario: Add and edit labels
- **WHEN** the user opens the add screen, and later chooses to edit a card while studying
- **THEN** the add screen's title and button say "Add card", and the study menu item and the editor's title say "Edit card"

#### Scenario: Editor error
- **WHEN** the user tries to add a Basic card with an empty Front field
- **THEN** the error says the entry would produce no cards and does not call it a note

#### Scenario: Note type kept
- **WHEN** the user opens the add screen
- **THEN** the template picker is labelled "Note type"

## MODIFIED Requirements

### Requirement: Deck list home screen
The home screen SHALL show the deck tree with new, learning and review counts for each deck. Parent decks SHALL be collapsible. Each deck SHALL have actions to study, add a card, open its options, rename it and delete it.

#### Scenario: Due counts visible
- **WHEN** a deck has 5 new, 2 learning and 30 review cards due
- **THEN** its row shows 5, 2 and 30 in three distinguishable colours

### Requirement: First run
On first launch, the app SHALL open on the deck list with the `Default` deck and a short empty-state hint offering to add a card or import a file.

#### Scenario: Empty state
- **WHEN** the app is launched for the first time
- **THEN** the deck list shows `Default` and buttons to add a card or import
