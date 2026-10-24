# Spec Delta

## Purpose

Decides when each card is next shown, using the FSRS spaced-repetition algorithm, and which cards are due today within the deck's daily limits.

## ADDED Requirements

### Requirement: Card states
Every card SHALL be in exactly one scheduling state: New, Learning, Review or Relearning. A card SHALL also record its due time, its FSRS memory state (stability and difficulty, absent while New), its review count, its lapse count and its last review time.

#### Scenario: New card has no memory state
- **WHEN** a card is created
- **THEN** it is in state New, has no stability or difficulty, and has review and lapse counts of 0

### Requirement: FSRS scheduling
When a card is rated Again, Hard, Good or Easy, the system SHALL compute its new state, memory state and due time with FSRS (default parameters) so that the predicted recall probability at the due time equals the deck's desired retention.

#### Scenario: Higher retention means shorter intervals
- **WHEN** two identical review cards are rated Good, one in a deck with desired retention 0.95 and one in a deck with 0.85
- **THEN** the card in the 0.95 deck gets a shorter interval

#### Scenario: Ratings are ordered
- **WHEN** the intervals for Again, Hard, Good and Easy are computed for the same card
- **THEN** each interval is greater than or equal to the previous one, and Easy's is strictly greater than Again's

### Requirement: Lapses
Rating a Review card Again SHALL increment its lapse count and move it to Relearning with a short re-study step.

#### Scenario: Forgotten review card
- **WHEN** a Review card with 2 lapses is rated Again
- **THEN** it has 3 lapses, is in Relearning, and is due again within the same day

### Requirement: Learning steps
New and Relearning cards SHALL use short intraday steps (default 1 minute and 10 minutes for learning, 10 minutes for relearning) before they graduate to Review. Rating Easy SHALL graduate immediately.

#### Scenario: New card rated Good twice
- **WHEN** a New card is rated Good, and then rated Good again once its 10-minute step is due
- **THEN** after the second rating it is in Review with a due date at least one day away

### Requirement: Interval preview
For any card, the system SHALL provide the next due interval for each of the four ratings without changing the card.

#### Scenario: Preview is side-effect free
- **WHEN** the previews for a card are requested and then the card is rated Good
- **THEN** the card's resulting interval equals the Good preview

### Requirement: Review log
Every rating SHALL append an entry to the review history: card, time, rating, state before, interval before and after, and time taken to answer. The entry SHALL record enough of the previous state that the rating can be undone exactly.

#### Scenario: Log entry written
- **WHEN** the user rates a card Hard after 7 seconds
- **THEN** the review history contains an entry for that card with rating Hard and a duration of 7 seconds

### Requirement: Day boundary
The study day SHALL roll over at a configurable hour in the device's local time (default 04:00). Daily limits, burying and "due today" SHALL be measured against this boundary.

#### Scenario: Late-night review counts for the previous day
- **WHEN** the rollover hour is 04:00 and the user reviews at 02:30
- **THEN** that review counts toward the previous calendar day's limits

### Requirement: Due cards and daily limits
A deck's due cards SHALL be: Learning and Relearning cards due by now; Review cards due by the end of the current study day, up to the reviews-per-day limit minus reviews already done today; and New cards, up to the new-per-day limit minus new cards already introduced today. Counts for a parent deck SHALL include its children, and each child's own limits SHALL also apply.

#### Scenario: New limit reached
- **WHEN** a deck has a new-per-day limit of 20, 50 new cards, and 20 new cards have been introduced today
- **THEN** the deck shows 0 new cards due until the next rollover

#### Scenario: Learning ignores limits
- **WHEN** the review limit for today is used up and a Relearning card becomes due
- **THEN** the Relearning card is still shown

### Requirement: Order of new cards
New cards SHALL be introduced in the order they were added to the collection (import order for imported cards). Cards of the same note SHALL NOT be shown on the same day once one of them has been studied (sibling burying).

#### Scenario: Sibling buried
- **WHEN** the forward card of a reversed note is studied today
- **THEN** the reverse card is not shown again until the next study day
