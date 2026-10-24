# filtered-decks Specification

## Purpose

Filtered decks gather cards matching a search into a temporary deck for studying outside the normal schedule, remember each card's home deck, and offer custom study sessions built on them.

## Requirements

### Requirement: Build filtered decks
The user SHALL be able to create a filtered deck with a search, a card limit and an order, and optionally a second search with its own limit and order. Building SHALL move up to the limit of matching cards into it, excluding suspended and buried cards and cards already in a filtered deck. Each moved card SHALL remember its home deck. Rebuilding SHALL return the current cards home first and build again.

#### Scenario: Build from a tag
- **WHEN** the user creates a filtered deck with search `tag:exam`, limit 50 and order random, and 80 unsuspended cards have the tag
- **THEN** 50 of them are in the filtered deck and remember their home decks

#### Scenario: Suspended excluded
- **WHEN** a matching card is suspended
- **THEN** it stays in its home deck

### Requirement: Return cards home
Emptying a filtered deck, deleting it, or rebuilding it SHALL return its cards to their home decks with their scheduling as it is at that time.

#### Scenario: Empty
- **WHEN** the user empties a filtered deck holding 50 cards
- **THEN** the 50 cards are back in their home decks and the filtered deck is empty

### Requirement: Studying filtered decks
Filtered decks SHALL show all their cards for study without daily limits, using each card's home deck options for scheduling. When "Reschedule cards based on my answers" is on, answers SHALL schedule cards as normal reviews, with cards reviewed before their due date scheduled from the actual time since their last review. A card that finishes learning or is answered as a review SHALL return to its home deck.

#### Scenario: Review ahead
- **WHEN** a card due in 5 days with a 10-day interval is studied in a filtered deck and rated Good
- **THEN** its new interval is shorter than if it had been rated Good on its due date, and it is back in its home deck

### Requirement: Preview mode
When "Reschedule cards based on my answers" is off, answering SHALL NOT change a card's scheduling. Again, Hard and Good SHALL show the card again after the deck's configured delays (60, 600 and 0 seconds by default, 0 meaning not again), and Easy SHALL return it to its home deck.

#### Scenario: Preview new cards
- **WHEN** the user previews a New card and rates it Good with a delay of 0
- **THEN** the card is still New, back in its home deck, and not shown again in the session

### Requirement: Filtered deck limits
A filtered deck SHALL NOT have subdecks, SHALL NOT be the target for adding notes, and SHALL be marked as filtered on the deck list with Rebuild and Empty actions.

#### Scenario: Add screen
- **WHEN** the user opens the add screen
- **THEN** filtered decks are not offered as the deck

### Requirement: Custom study sessions
Custom study SHALL offer: review forgotten cards (rated Again in the last N days), review ahead (due in the next N days), preview new cards (added in the last N days, in preview mode), and study by card state or tag (new, due or all cards, optionally limited to tags, with a card count). Each SHALL create or replace a filtered deck named "Custom Study Session" for the chosen deck and start it.

#### Scenario: Forgotten cards
- **WHEN** the user chooses "Review forgotten cards" for `Spanish` with 2 days
- **THEN** a "Custom Study Session" deck holds the cards of `Spanish` rated Again in the last 2 days

#### Scenario: Replace previous session
- **WHEN** a "Custom Study Session" exists and the user starts another custom study
- **THEN** the previous session's cards return home and the deck is rebuilt for the new choice
