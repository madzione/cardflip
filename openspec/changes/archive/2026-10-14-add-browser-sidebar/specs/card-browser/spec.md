# Spec Delta

## ADDED Requirements

### Requirement: Browser sidebar
The browser SHALL show a sidebar with Card states, Decks, Note types and Tags, with decks and tags as trees. On narrow screens it SHALL be available as a drawer. A filter field SHALL narrow the sidebar to matching items.

#### Scenario: Tag tree
- **WHEN** notes have the tags `verb::irregular` and `verb::regular`
- **THEN** the sidebar shows `verb` with the children `irregular` and `regular`

### Requirement: Search from the sidebar
Clicking a sidebar item SHALL replace the search with that item's search. Ctrl-click (Cmd-click on macOS) SHALL add it to the current search with AND, and Alt-click (Option-click) SHALL add it negated. On touch devices, a long press SHALL offer these choices.

#### Scenario: Click a deck
- **WHEN** the user clicks the deck `Spanish` in the sidebar
- **THEN** the search becomes `deck:Spanish` and the matching cards are listed

#### Scenario: Combine
- **WHEN** the search is `deck:Spanish` and the user Alt-clicks the tag `verb`
- **THEN** the search becomes `deck:Spanish -tag:verb`

### Requirement: Manage decks from the sidebar
The user SHALL be able to rename and delete decks from the sidebar and to move a deck under another deck by dragging it, with the same rules and confirmations as on the deck list.

#### Scenario: Drag to reparent
- **WHEN** the user drags deck `Verbs` onto `Spanish`
- **THEN** the deck is renamed to `Spanish::Verbs` and keeps its cards
