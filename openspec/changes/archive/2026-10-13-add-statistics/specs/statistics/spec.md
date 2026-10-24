# Spec Delta

## Purpose

The statistics screen: what the user studied, what is coming, and how well cards are remembered, for a deck, the whole collection or a search, over a chosen period.

## ADDED Requirements

### Requirement: Scope and period
Statistics SHALL be shown for a deck with its subdecks, the whole collection, or the cards matching a search, and time-based sections SHALL cover a chosen period of 1 month, 3 months, 1 year or all history. Days SHALL follow the study day boundary. Collection-wide statistics SHALL include reviews of deleted notes.

#### Scenario: Deck scope
- **WHEN** the user opens statistics from the menu of `Spanish`
- **THEN** every section covers only cards in `Spanish` and its subdecks

### Requirement: Today summary
Statistics SHALL show, for the current study day, the number of cards studied, the time spent, the number of Again answers with the correct percentage, and how many reviews were learning, review, relearning and filtered.

#### Scenario: After a session
- **WHEN** the user answered 40 cards today, 4 of them Again, in 10 minutes
- **THEN** the summary shows 40 cards, 10 minutes, 4 Again and 90 % correct

### Requirement: Future due
Statistics SHALL show the number of cards due on each upcoming day of the period, the number currently overdue, a cumulative total, and the average daily load.

#### Scenario: Forecast
- **WHEN** 30 review cards are due tomorrow and 12 the day after
- **THEN** the forecast shows 30 and 12 for those days

### Requirement: Review history graphs
Statistics SHALL show a calendar heatmap of reviews per day, and the number of reviews and the review time per day, week or month, each split into learning, young, mature, relearning and filtered, where young means an interval under 21 days. They SHALL also show reviews and success rate by hour of day, and the answer button counts with correct percentage for learning, young and mature cards.

#### Scenario: Calendar
- **WHEN** the user reviewed on 20 of the last 30 days
- **THEN** the calendar marks those 20 days, with darker colours for more reviews

### Requirement: Card state graphs
Statistics SHALL show card counts by state (new, learning, relearning, young, mature, suspended, buried), a histogram of review intervals, and histograms of stability, difficulty and retrievability, together with the average retrievability and the estimated total knowledge (the sum of all cards' retrievability).

#### Scenario: Retrievability
- **WHEN** a deck has 100 review cards with an average retrievability of 0.92
- **THEN** the retrievability section shows 92 % on average and an estimated knowledge of about 92 cards

### Requirement: True retention
Statistics SHALL show, for today, yesterday, the last week, month and year, and all time, the number of passed and failed reviews of young and mature cards and the resulting retention, next to the deck's desired retention.

#### Scenario: Compare with target
- **WHEN** last month's mature reviews were 180 passed and 20 failed and the desired retention is 0.90
- **THEN** the table shows 90.0 % for last month's mature cards and the target 90 %

### Requirement: Save statistics
The user SHALL be able to save the statistics screen as a PDF.

#### Scenario: PDF
- **WHEN** the user chooses Save as PDF
- **THEN** a PDF with all sections is saved or downloaded
