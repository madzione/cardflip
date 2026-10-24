# Spec Delta

## ADDED Requirements

### Requirement: Browser appearance templates
A card type SHALL optionally have a browser question template and a browser answer template. When set, the browser list SHALL show the question text from the browser template, rendered and with HTML stripped. The card details and study SHALL keep using the normal templates.

#### Scenario: Simplified question
- **WHEN** a card type's front is `<div class=hint>{{Hint}}</div>{{Word}}` and its browser question template is `{{Word}}`
- **THEN** the browser row shows only the word, and studying shows the hint too
