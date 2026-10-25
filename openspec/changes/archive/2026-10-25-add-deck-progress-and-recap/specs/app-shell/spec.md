# Spec Delta

## ADDED Requirements

### Requirement: Deck progress in the deck list
Settings SHALL include *Show deck progress in the deck list*, off by default and belonging to the device. When it is on, each deck row SHALL show a thin bar of the deck's new, learning, young and mature cards, including its subdecks and leaving out suspended cards.

#### Scenario: Bars shown
- **WHEN** the user turns on *Show deck progress in the deck list*
- **THEN** each deck row shows a thin progress bar under its name

#### Scenario: Off by default
- **WHEN** the user has not changed the setting
- **THEN** deck rows show no progress bar

## MODIFIED Requirements

### Requirement: Deck overview
Choosing a deck on the deck list SHALL open an overview showing the deck's name, its new, learning and review counts for today, the number of buried cards, its description, and its progress: a bar of new, learning, young and mature cards of the deck and its subdecks, leaving out suspended cards, with the share of mature cards. It SHALL have actions Study now and Options, and Unbury when cards are buried. When nothing is due, the overview SHALL say so and when the next card is due.

#### Scenario: Overview before studying
- **WHEN** the user taps deck `Spanish` with 5 new and 30 review cards due
- **THEN** the overview shows 5 and 30 with a Study now button, and no session has started

#### Scenario: Nothing due
- **WHEN** the user taps a deck with nothing due and the next card due tomorrow
- **THEN** the overview says nothing is due and that the next card is due tomorrow

#### Scenario: Deck progress
- **WHEN** `Spanish` has 100 cards, 10 of them suspended, and of the other 90 there are 18 new, 4 learning, 12 young and 56 mature
- **THEN** the overview shows a bar with those four parts and "62 % mature"

#### Scenario: Buried cards count by state
- **WHEN** a mature card of `Spanish` is buried
- **THEN** it still counts as mature in the progress bar
