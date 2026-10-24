# Spec Delta

## ADDED Requirements

### Requirement: Learning cards becoming due while waiting
When a session shows no card ("Session complete" or "Nothing due right now") and a learning card becomes due later today, the study screen SHALL react at that time without user action. With auto-continue off (the default), it SHALL show that learning cards are due now with a Continue button that shows the card. With auto-continue on, it SHALL show the due card directly. A manual "Check again" SHALL keep working before that time.

#### Scenario: Prompt by default
- **WHEN** auto-continue is off, the summary says more learning cards will be due at 14:05, and 14:05 passes
- **THEN** the screen says learning cards are due now and offers Continue; choosing Continue shows the card's question

#### Scenario: Auto-continue
- **WHEN** auto-continue is on, the summary says more learning cards will be due at 14:05, and 14:05 passes
- **THEN** the screen shows the due card's question without the user tapping anything

#### Scenario: Opened before the card is due
- **WHEN** the user opens a deck whose only card due today is a learning card due in 1 minute, and the minute passes
- **THEN** the "Nothing due right now" screen reacts the same way as the summary, according to the setting

#### Scenario: Nothing more today
- **WHEN** the session is complete and no learning card is due later today
- **THEN** the screen does not change on its own

### Requirement: Session counts stay current
While a card is on screen, the session's remaining counts SHALL update when a learning card becomes due, without changing or un-revealing the card on screen.

#### Scenario: Learning count during a review
- **WHEN** the user is looking at the revealed answer of a review card and a learning card becomes due
- **THEN** the learning count increases by one and the same answer stays on screen

## MODIFIED Requirements

### Requirement: Session completion
When no due cards remain, the session SHALL end with a summary showing the number of cards studied and the time spent. If learning cards will become due later today, the summary SHALL say when, and SHALL react when they become due as described in "Learning cards becoming due while waiting".

#### Scenario: Finish
- **WHEN** the last due card is answered and no learning card is due within the session
- **THEN** the summary screen shows the number of cards studied and the time spent

#### Scenario: Finish with learning cards pending
- **WHEN** the last due card is answered and a learning card is due in 1 minute
- **THEN** the summary shows the cards studied, the time spent, and when the learning card will be due
