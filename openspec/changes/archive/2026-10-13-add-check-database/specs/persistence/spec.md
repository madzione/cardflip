# Spec Delta

## ADDED Requirements

### Requirement: Check and repair the collection
The user SHALL be able to check the collection for inconsistencies and repair them in one operation: cards without a note or note type, notes with a wrong number of fields, cards for missing templates or cloze numbers, missing cards, cards in missing decks, invalid scheduling values, and invalid tags. The check SHALL report each kind of problem found with its count, or that none was found. If the repair cannot be saved, the collection SHALL stay unchanged.

#### Scenario: Orphaned card
- **WHEN** the collection contains a card whose note no longer exists and the user checks the database
- **THEN** the card is deleted and the report says 1 card without a note was removed

#### Scenario: Healthy collection
- **WHEN** the user checks a consistent collection
- **THEN** the report says no problems were found and nothing changes
