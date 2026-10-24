# Spec Delta

## ADDED Requirements

### Requirement: Reset options
Resetting cards to New SHALL offer to restore each card's original position among new cards, where it is known (on by default), and to reset its review and lapse counts to zero (off by default). Without restoring, reset cards SHALL be placed after all other new cards. Every reset SHALL add a manual entry to the card's review history.

#### Scenario: Restore position
- **WHEN** a card that was the 3rd new card when first studied is reset with "restore original position"
- **THEN** it is introduced before the new cards that were added after it

#### Scenario: Reset counts
- **WHEN** a card with 12 reviews and 4 lapses is reset with "reset repetition and lapse counts"
- **THEN** it is New with 0 reviews and 0 lapses, and its earlier review entries are still in its history

### Requirement: Original position
A card SHALL remember its new-card position when it is first answered, and keep it across later reviews, moves and backups.

#### Scenario: Position kept
- **WHEN** a new card at position 42 is answered and later reviewed many times
- **THEN** its remembered original position is still 42
