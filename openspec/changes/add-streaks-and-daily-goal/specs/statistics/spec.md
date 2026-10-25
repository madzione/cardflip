# Spec Delta

## ADDED Requirements

### Requirement: Streaks
Statistics SHALL show a Streaks section with the current streak, the longest streak with its dates, the number of study days in the period out of the days in the period, and, when a daily goal is set, the number of days in the period that met it. The section SHALL be computed from the reviews of the scope, using the study day definition of the `progress` capability.

#### Scenario: Collection streaks
- **WHEN** the period is 1 month, the user studied on 24 of the last 30 days, the last 9 in a row, and the longest run ever was 41 days
- **THEN** the section shows a current streak of 9 days, a longest streak of 41 days with its start and end dates, and 24 of 30 study days

#### Scenario: Deck scope
- **WHEN** the user studied every day this week but `Spanish` only on the last 3 days
- **THEN** the statistics of `Spanish` show a current streak of 3 days

#### Scenario: Goal days
- **WHEN** the goal is 30 cards and the user answered 30 or more cards on 18 days of the last month
- **THEN** the section shows 18 days the goal was met

#### Scenario: No goal
- **WHEN** no daily goal is set
- **THEN** the section does not show days the goal was met

## MODIFIED Requirements

### Requirement: Save statistics
The user SHALL be able to save the statistics screen as a PDF.

#### Scenario: PDF
- **WHEN** the user chooses Save as PDF
- **THEN** a PDF with all sections, including Streaks, is saved or downloaded
