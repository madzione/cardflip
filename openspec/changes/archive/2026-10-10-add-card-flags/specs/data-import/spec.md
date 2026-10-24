# Spec Delta

## ADDED Requirements

### Requirement: Flags on Anki import
When an `.apkg` is imported with "keep scheduling", newly added cards SHALL keep their Anki flag. When imported as New, cards SHALL have no flag. Cards that already exist SHALL keep their current flag.

#### Scenario: Flag kept
- **WHEN** a package containing a card with the green flag is imported with "keep scheduling"
- **THEN** the imported card has flag 3

#### Scenario: Import as new
- **WHEN** the same package is imported with "import as new"
- **THEN** the imported card has no flag
