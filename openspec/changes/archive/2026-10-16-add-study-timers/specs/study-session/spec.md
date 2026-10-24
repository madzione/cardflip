# Spec Delta

## ADDED Requirements

### Requirement: On-screen answer timer
When the deck's "show answer timer" option is on, the study screen SHALL show the time spent on the current card, changing colour once the maximum answer seconds are reached. With "stop timer on answer" on, the timer SHALL stop when the answer is shown. The timer SHALL NOT affect what is recorded.

#### Scenario: Timer stops on answer
- **WHEN** the timer is shown, stop timer on answer is on, and the user reveals the answer after 8 seconds
- **THEN** the timer shows 8 seconds and stays there until the card is rated

### Requirement: Timebox
When the timebox limit is set, each time that many minutes of a session pass, the study screen SHALL show how many cards were studied in that time, with Continue and Finish. Continue SHALL start the next timebox; Finish SHALL end the session with the usual summary.

#### Scenario: Ten-minute timebox
- **WHEN** the limit is 10 minutes and 10 minutes of studying pass with 42 cards answered
- **THEN** a message says 42 cards were studied in 10 minutes, with Continue and Finish
