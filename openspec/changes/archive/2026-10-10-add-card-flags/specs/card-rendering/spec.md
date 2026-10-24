# Spec Delta

## ADDED Requirements

### Requirement: CardFlag special field
`{{CardFlag}}` SHALL render `flag1` to `flag7` for a card with that flag, and nothing for a card without a flag.

#### Scenario: Flag as CSS class
- **WHEN** the question template is `<div class="{{CardFlag}}">{{Front}}</div>` and the card has flag 3
- **THEN** the question renders as `<div class="flag3">` around the Front field
