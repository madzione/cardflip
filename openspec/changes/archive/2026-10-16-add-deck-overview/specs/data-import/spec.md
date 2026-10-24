# Spec Delta

## ADDED Requirements

### Requirement: Deck descriptions on Anki import
Importing an `.apkg` SHALL keep the descriptions of the decks it creates. Existing decks SHALL keep their current description.

#### Scenario: Shared deck description
- **WHEN** a shared deck with a description is imported into a collection without that deck
- **THEN** the new deck's overview shows the description
