# Spec Delta

## ADDED Requirements

### Requirement: Learn ahead
When no learning, review or new card is due in a session, the session SHALL show the learning card that becomes due soonest, if it is due within the learn ahead limit. With a limit of 0, learning cards SHALL only be shown once due.

#### Scenario: Finish without waiting
- **WHEN** the limit is 20 minutes, the last review card is answered and a learning card is due in 5 minutes
- **THEN** the learning card is shown immediately

#### Scenario: Limit off
- **WHEN** the limit is 0 and the only remaining card is a learning card due in 5 minutes
- **THEN** the session shows when that card will be due instead of the card

## MODIFIED Requirements

### Requirement: Session completion
When no due cards remain and no learning card is due within the learn ahead limit, the session SHALL end with a summary showing the number of cards studied and the time spent. If learning cards will become due later today, the summary SHALL say when, and SHALL react when they become due as described in "Learning cards becoming due while waiting".

#### Scenario: Finish
- **WHEN** the last due card is answered and no learning card is due within the session
- **THEN** the summary screen shows the number of cards studied and the time spent

#### Scenario: Finish with learning cards pending
- **WHEN** the learn ahead limit is 20 minutes, the last due card is answered and a learning card is due in 1 hour
- **THEN** the summary shows the cards studied, the time spent, and when the learning card will be due

### Requirement: Learning cards becoming due while waiting
When a session shows no card ("Session complete" or "Nothing due right now") and a learning card becomes due later today, the study screen SHALL react when that card comes within the learn ahead limit, without user action. With auto-continue off (the default), it SHALL show that learning cards are due now with a Continue button that shows the card. With auto-continue on, it SHALL show the due card directly. A manual "Check again" SHALL keep working before that time.

#### Scenario: Prompt by default
- **WHEN** the learn ahead limit is 0, auto-continue is off, the summary says more learning cards will be due at 14:05, and 14:05 passes
- **THEN** the screen says learning cards are due now and offers Continue; choosing Continue shows the card's question

#### Scenario: Auto-continue
- **WHEN** the learn ahead limit is 0, auto-continue is on, the summary says more learning cards will be due at 14:05, and 14:05 passes
- **THEN** the screen shows the due card's question without the user tapping anything

#### Scenario: Opened before the card is due
- **WHEN** the learn ahead limit is 0, the user opens a deck whose only card due today is a learning card due in 1 minute, and the minute passes
- **THEN** the "Nothing due right now" screen reacts the same way as the summary, according to the setting

#### Scenario: Nothing more today
- **WHEN** the session is complete and no learning card is due later today
- **THEN** the screen does not change on its own

#### Scenario: Reacting at the limit
- **WHEN** the learn ahead limit is 20 minutes, auto-continue is off, and the summary says the next learning card is due at 14:30
- **THEN** at 14:10 the screen says learning cards are due now and offers Continue
