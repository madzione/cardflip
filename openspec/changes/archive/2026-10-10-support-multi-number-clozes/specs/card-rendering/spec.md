# Spec Delta

## ADDED Requirements

### Requirement: Multi-number cloze deletions
A cloze deletion SHALL accept a comma-separated list of cloze numbers, written `{{c1,2::answer}}` or `{{c1,2::answer::hint}}`, with optional spaces after commas. For each listed number N, the deletion SHALL render on card N as a single-number deletion `{{cN::…}}` would. On cards whose number is not listed, it SHALL render as its plain answer text.

#### Scenario: Hidden on every listed card
- **WHEN** cards 1, 2 and 3 of `{{c1,2::Paris}} is in {{c3::France}}` are rendered
- **THEN** the questions of cards 1 and 2 show `[...] is in France`, and the question of card 3 shows `Paris is in [...]`

#### Scenario: Hint on a multi-number deletion
- **WHEN** card 2 of `{{c1,2::Paris::city}}` is rendered
- **THEN** the question shows `[city]`

#### Scenario: Malformed list is literal text
- **WHEN** a field contains `{{c1,,2::Paris}}`
- **THEN** the text is shown as written and no cloze is detected
