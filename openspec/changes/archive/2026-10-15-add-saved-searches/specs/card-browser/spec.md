# Spec Delta

## ADDED Requirements

### Requirement: Saved searches
The user SHALL be able to save the current search under a unique name, run a saved search from a list next to the search field, and rename or delete saved searches. A search with an invalid term SHALL NOT be saved. Saved searches SHALL be included in backups.

#### Scenario: Save and run
- **WHEN** the user saves `deck:Spanish tag:leech` as `Spanish leeches` and later chooses it
- **THEN** the search field contains `deck:Spanish tag:leech` and the matching cards are listed

#### Scenario: Invalid search
- **WHEN** the user tries to save `is:bogus`
- **THEN** saving is refused with the search error

### Requirement: Default browser search
Settings SHALL include a default browser search, empty by default. When the browser opens, it SHALL start with that search.

#### Scenario: Start filtered
- **WHEN** the default search is `is:due` and the user opens Browse
- **THEN** the search field contains `is:due` and only due cards are listed
