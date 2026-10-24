# Spec Delta

## ADDED Requirements

### Requirement: Review history on Anki import
When an `.apkg` is imported with "keep scheduling", the review history of each card the import adds SHALL be imported with time, rating, kind, intervals and answer time. Anki's manual and rescheduling entries SHALL become manual entries. Importing the same package again SHALL NOT duplicate entries. The summary SHALL report how many review entries were imported.

#### Scenario: History kept
- **WHEN** a package with a card reviewed 12 times in Anki is imported with "keep scheduling"
- **THEN** that card's history in Cardflip has the 12 reviews with their dates and ratings

#### Scenario: Import as new
- **WHEN** the same package is imported with "import as new"
- **THEN** no review history is imported

#### Scenario: Re-import
- **WHEN** the package is imported a second time with "keep scheduling"
- **THEN** the card still has 12 review entries

### Requirement: Memory state from history
When a card imported with its scheduling has no FSRS memory state in Anki but has review history, its memory state SHALL be computed from that history with FSRS.

#### Scenario: Card scheduled with SM-2
- **WHEN** a card reviewed only under Anki's SM-2 scheduler is imported with its 8 reviews
- **THEN** its stability and difficulty result from replaying those 8 reviews, and its due date is kept
