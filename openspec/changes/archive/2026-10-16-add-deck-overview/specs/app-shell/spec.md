# Spec Delta

## ADDED Requirements

### Requirement: Deck overview
Choosing a deck on the deck list SHALL open an overview showing the deck's name, its new, learning and review counts for today, the number of buried cards, and its description, with actions Study now and Options, and Unbury when cards are buried. When nothing is due, the overview SHALL say so and when the next card is due.

#### Scenario: Overview before studying
- **WHEN** the user taps deck `Spanish` with 5 new and 30 review cards due
- **THEN** the overview shows 5 and 30 with a Study now button, and no session has started

#### Scenario: Nothing due
- **WHEN** the user taps a deck with nothing due and the next card due tomorrow
- **THEN** the overview says nothing is due and that the next card is due tomorrow
