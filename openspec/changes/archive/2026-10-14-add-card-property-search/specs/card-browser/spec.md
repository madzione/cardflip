# Spec Delta

## ADDED Requirements

### Requirement: Property search
`prop:<property><op><number>` SHALL compare a card property with `=`, `!=`, `<`, `<=`, `>` or `>=`, for the properties `ivl` (days), `due` (days relative to today), `reps`, `lapses`, `pos`, `s` (stability in days), `d` (difficulty from 0 to 1) and `r` (retrievability from 0 to 1). Cards without the property, such as a New card for `ivl`, SHALL NOT match. Unknown properties SHALL show an error; `prop:ease` SHALL explain that `prop:d` is used instead.

#### Scenario: Long intervals
- **WHEN** the user searches `prop:ivl>=30`
- **THEN** cards with an interval of 30 days or more are listed

#### Scenario: Overdue
- **WHEN** the user searches `prop:due<0`
- **THEN** review cards that were due before today are listed

#### Scenario: At risk
- **WHEN** the user searches `prop:r<0.8`
- **THEN** cards whose current retrievability is below 0.8 are listed

### Requirement: Burial and state search
`is:buried` SHALL match cards buried manually or held back today by sibling burying; `is:buried-manually` and `is:buried-sibling` SHALL match each kind. `is:learn` SHALL match learning and relearning cards, and `is:review` SHALL match review and relearning cards.

#### Scenario: Lapsed cards
- **WHEN** the user searches `is:learn is:review`
- **THEN** only relearning cards are listed

### Requirement: Card type and tag search
`card:<name>` SHALL match cards of card types with that name, with wildcards; `card:<n>` SHALL match card type number n, or cloze number n for cloze notes. `tag:none` SHALL match notes without tags.

#### Scenario: Reverse cards
- **WHEN** the user searches `card:2`
- **THEN** the second card of each reversed note and the cloze 2 cards are listed

### Requirement: Deck and ID search
`deck:current` SHALL match the deck most recently studied and its subdecks. `deck:<name>::*` SHALL match only that deck's subdecks. `nid:` and `cid:` SHALL match notes or cards by one ID or a comma-separated list.

#### Scenario: Without subdecks
- **WHEN** the user searches `deck:French -deck:French::*`
- **THEN** cards directly in `French` are listed, and cards in its subdecks are not

#### Scenario: By card IDs
- **WHEN** the user searches `cid:101,102`
- **THEN** exactly those two cards are listed
