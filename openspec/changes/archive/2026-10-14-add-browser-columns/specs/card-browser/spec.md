# Spec Delta

## MODIFIED Requirements

### Requirement: Card list
The browser SHALL list cards in columns chosen by the user from: question, answer, sort field, card type, note type, deck, due, interval, difficulty, stability, retrievability, reviews, lapses, tags, created, card modified and note modified, with text shown without HTML. The user SHALL be able to show, hide and reorder columns, and the layout SHALL be remembered. The list SHALL be sortable by any column, ascending or descending, and SHALL stay responsive with 50,000 cards. By default it SHALL show question, deck, due and note type.

#### Scenario: Large collection
- **WHEN** the collection has 50,000 cards and the user opens the browser
- **THEN** the first rows appear within 1 second on a typical desktop, and scrolling remains smooth

#### Scenario: Sort by lapses
- **WHEN** the user shows the Lapses column and sorts it descending
- **THEN** the card with the most lapses is listed first

#### Scenario: Layout remembered
- **WHEN** the user hides the Note type column, adds Interval, and restarts the app
- **THEN** the browser shows Question, Deck, Due and Interval
