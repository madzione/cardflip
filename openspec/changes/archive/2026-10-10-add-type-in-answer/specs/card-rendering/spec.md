# Spec Delta

## MODIFIED Requirements

### Requirement: Field filters
The renderer SHALL support the filters `text:` (strip HTML), `hint:` (collapsed content revealed by tapping), `type:` (an answer input on the question side and a comparison on the answer side, as described in "Type-in-the-answer comparison"), and `cloze:`. Unknown filters SHALL be ignored, and the field content SHALL render unfiltered.

#### Scenario: Text filter
- **WHEN** the template has `{{text:Front}}` and Front is `<b>bold</b>`
- **THEN** the output is `bold` without markup

## ADDED Requirements

### Requirement: Type-in-the-answer comparison
On the answer side, `{{type:Field}}` SHALL show the typed text compared with the field's text, with HTML stripped and surrounding whitespace ignored. Matching characters, wrong typed characters and missing expected characters SHALL be marked distinguishably. When nothing was typed, it SHALL show the expected text. Only the first `type:` filter of a card SHALL be active.

#### Scenario: Exact match
- **WHEN** the expected text is `perro` and the user typed `perro`
- **THEN** the answer shows `perro` marked as correct

#### Scenario: Typo
- **WHEN** the expected text is `perro` and the user typed `pero`
- **THEN** the answer shows the typed text and the expected text, with the missing `r` marked

#### Scenario: Preview without input
- **WHEN** a card with `{{type:Back}}` is shown in the browser details
- **THEN** no input is shown and the answer shows the expected text

### Requirement: Type-in variants
`{{type:cloze:Field}}` SHALL compare against the answer of the active cloze number, joining several deletions of that number with `, `. `{{type:nc:Field}}` SHALL ignore accents and other combining marks on both sides when comparing.

#### Scenario: Cloze answer
- **WHEN** card 1 of `{{c1::Paris}} is in {{c2::France}}` uses `{{type:cloze:Text}}` and the user typed `Paris`
- **THEN** the comparison shows `Paris` as correct

#### Scenario: Ignore accents
- **WHEN** the template uses `{{type:nc:Back}}`, Back is `über` and the user typed `uber`
- **THEN** the comparison shows the answer as correct
