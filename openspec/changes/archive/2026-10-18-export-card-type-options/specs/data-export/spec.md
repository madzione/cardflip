# Spec Delta

## ADDED Requirements

### Requirement: Export of card type options
A package export SHALL keep each card type's browser appearance templates. It SHALL keep each card type's deck override when that deck is part of the export, naming it as the package's copy of that deck; otherwise the card type SHALL have no override in the package. This SHALL hold for both package formats.

#### Scenario: Round trip with an override
- **WHEN** card type `Production` overrides to deck `Spanish::Production` and has the browser question template `{{Back}}`, the user exports deck `Spanish` and imports the file into another collection
- **THEN** the imported `Production` card type overrides to the imported `Spanish::Production` deck and has the browser question template `{{Back}}`

#### Scenario: Override deck not exported
- **WHEN** card type `Production` overrides to deck `Production` and the user exports deck `Spanish`, which does not contain it
- **THEN** the package's `Production` card type has no deck override, and its browser appearance templates are kept

#### Scenario: Older format
- **WHEN** the user exports with "support older Anki versions"
- **THEN** the package keeps the card type options in the same way
