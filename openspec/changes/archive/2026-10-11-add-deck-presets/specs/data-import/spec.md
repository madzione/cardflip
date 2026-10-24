# Spec Delta

## ADDED Requirements

### Requirement: Presets on Anki import
Importing an `.apkg` that contains deck presets SHALL import them and assign them to the decks the import creates; existing decks SHALL keep their preset. A package preset whose name is taken by a preset with different options SHALL be imported under a new name; one with the same name and options SHALL be reused. Anki's default preset (ID 1) SHALL map to the collection's `Default` preset, as in Anki, whatever its options.

#### Scenario: Shared deck settings
- **WHEN** a package whose deck uses a preset with 10 new cards per day is imported into a collection without that deck
- **THEN** the new deck uses an imported preset with 10 new cards per day

#### Scenario: Anki's default preset
- **WHEN** a package whose decks use Anki's default preset is imported
- **THEN** the new decks use `Default`, and no preset is added
