# Spec Delta

## ADDED Requirements

### Requirement: Workload simulation
The user SHALL be able to simulate a deck's expected daily reviews and number of memorised cards over a chosen number of days, starting from its current cards, for given new and review limits and desired retention, and to compare two retentions. Simulating SHALL NOT change the collection and SHALL be reproducible for the same inputs and collection.

#### Scenario: Compare retentions
- **WHEN** the user simulates 365 days at retention 0.90 and compares with 0.95
- **THEN** a chart shows both curves, and the 0.95 curve has more reviews per day

### Requirement: Retention suggestion
The user SHALL be able to request a suggested desired retention, which SHALL be the retention between 0.70 and 0.99 with the lowest simulated review time per memorised card. Applying the suggestion SHALL set the deck's desired retention only after the user confirms.

#### Scenario: Help me decide
- **WHEN** the user chooses "Help me decide"
- **THEN** a suggested retention is shown with the cost curve, and the deck's retention is unchanged until the user applies it
