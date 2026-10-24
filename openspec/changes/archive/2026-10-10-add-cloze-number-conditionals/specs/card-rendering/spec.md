# Spec Delta

## ADDED Requirements

### Requirement: Cloze-number conditionals
When rendering a card of a cloze note type, a section named `cN`, where N is a positive number and the note type has no field named `cN`, SHALL be a cloze-number condition: `{{#cN}}…{{/cN}}` SHALL render its content only on cloze card N, and `{{^cN}}…{{/cN}}` SHALL render its content on every other cloze card. This SHALL apply on both the question and the answer side. When the note type has a field named `cN`, the section SHALL test that field as usual.

#### Scenario: Content only on card 1
- **WHEN** the question template is `{{cloze:Text}}{{#c1}} (capital){{/c1}}` and cards 1 and 2 are rendered
- **THEN** card 1's question ends with ` (capital)` and card 2's question does not

#### Scenario: Inverted condition
- **WHEN** the template contains `{{^c2}}Not card 2{{/c2}}`
- **THEN** cards 1 and 3 show `Not card 2` and card 2 does not

#### Scenario: Field named c1 wins
- **WHEN** a cloze note type has a field `c1` that is empty and the template contains `{{#c1}}X{{/c1}}`
- **THEN** no card shows `X`
