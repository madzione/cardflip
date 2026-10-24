# Spec Delta

## MODIFIED Requirements

### Requirement: Type-in-the-answer comparison
On the answer side, `{{type:Field}}` SHALL show the typed text compared with the field's text, with HTML stripped and surrounding whitespace ignored. Matching characters, wrong typed characters and missing expected characters SHALL be marked distinguishably. When nothing was typed, it SHALL show the expected text. Only the first `type:` filter of a card SHALL be active. When the answer template has no `type:` filter of its own, the comparison SHALL appear where the question's `type:` filter sits inside `{{FrontSide}}`, and an `<hr id=answer>` line on the answer SHALL move to just before it. When the answer template has its own, `{{FrontSide}}` SHALL show no input or comparison.

#### Scenario: Exact match
- **WHEN** the expected text is `perro` and the user typed `perro`
- **THEN** the answer shows `perro` marked as correct

#### Scenario: Typo
- **WHEN** the expected text is `perro` and the user typed `pero`
- **THEN** the answer shows the typed text and the expected text, with the missing `r` marked

#### Scenario: Preview without input
- **WHEN** a card with `{{type:Back}}` is shown in the browser details
- **THEN** no input is shown and the answer shows the expected text

#### Scenario: Type filter only on the question
- **WHEN** the question template is `{{Front}} {{type:Back}}`, the answer template is `{{FrontSide}}<hr id=answer>{{Back}}`, Front is `dog`, Back is `perro` and the user typed `pero`
- **THEN** the answer shows `dog`, then the answer line, then the comparison of `pero` with `perro`, then `perro`

#### Scenario: Question-only type filter keeps its modifiers
- **WHEN** the question template has `{{type:nc:Back}}`, the answer template is `{{FrontSide}}<hr id=answer>{{Back}}`, Back is `über` and the user typed `uber`
- **THEN** the comparison shows the answer as correct

#### Scenario: Question-only type filter in a preview
- **WHEN** a card whose question template is `{{Front}} {{type:Back}}` and whose answer template is `{{FrontSide}}<hr id=answer>{{Back}}` is shown in the browser details
- **THEN** the answer shows the expected text below the answer line, and the browser's answer column includes it

#### Scenario: Type filter on both sides
- **WHEN** the question template is `{{Front}} {{type:Back}}`, the answer template is `{{FrontSide}}<hr id=answer>{{type:Back}}` and the user typed `perro`
- **THEN** the answer shows the comparison once, below the answer line
