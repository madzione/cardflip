# Spec Delta

## MODIFIED Requirements

### Requirement: FSRS scheduling
When a card is rated Again, Hard, Good or Easy, the system SHALL compute its new state, memory state and due time with FSRS, using the parameters of the card's deck or the default parameters when none are set, so that the predicted recall probability at the due time equals the deck's desired retention.

#### Scenario: Higher retention means shorter intervals
- **WHEN** two identical review cards are rated Good, one in a deck with desired retention 0.95 and one in a deck with 0.85
- **THEN** the card in the 0.95 deck gets a shorter interval

#### Scenario: Ratings are ordered
- **WHEN** the intervals for Again, Hard, Good and Easy are computed for the same card
- **THEN** each interval is greater than or equal to the previous one, and Easy's is strictly greater than Again's

## ADDED Requirements

### Requirement: Optimize parameters
The user SHALL be able to optimize a deck's FSRS parameters, or all decks' parameters, from the review history selected by its search. Optimizing SHALL require at least 400 reviews, SHALL keep the current parameters unless the new ones fit the history better, SHALL run without blocking the app, SHALL show progress, and SHALL be cancellable. An optional health check SHALL report unreliable results instead of applying them.

#### Scenario: Optimize
- **WHEN** a deck has 5,000 reviews and the user chooses Optimize
- **THEN** progress is shown, and afterwards the deck has parameters that fit its history at least as well as the previous ones

#### Scenario: Not enough reviews
- **WHEN** a deck has 150 reviews and the user chooses Optimize
- **THEN** the app explains that at least 400 reviews are needed and the parameters are unchanged
