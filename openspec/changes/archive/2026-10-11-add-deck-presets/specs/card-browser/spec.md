# Spec Delta

## ADDED Requirements

### Requirement: Preset search
The search SHALL support `preset:<name>`, matching cards in decks that use the preset with that name, case-insensitively; quotes SHALL allow names with spaces.

#### Scenario: Search by preset
- **WHEN** decks `Spanish` and `French` use preset `Languages` and the user searches `preset:Languages`
- **THEN** the cards of both decks are listed
