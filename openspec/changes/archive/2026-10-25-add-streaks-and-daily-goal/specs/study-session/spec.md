# Spec Delta

## MODIFIED Requirements

### Requirement: Start a session
The user SHALL be able to start studying any deck that has due cards, from the deck's overview or directly from the deck list's study action. The session SHALL include the due cards of the deck and its children, according to the scheduling rules. Starting a session directly for a deck with nothing due SHALL show a "congratulations, nothing due" screen that says when the next card is due, and the streak and goal lines described in the `progress` capability.

#### Scenario: Nothing due
- **WHEN** the user starts studying a deck with no due cards from the deck list's study action and the next card is due tomorrow
- **THEN** the app shows that studying is complete for now and that the next card is due tomorrow

#### Scenario: Nothing due with a goal
- **WHEN** the user starts studying a deck with no due cards, the goal is 30 cards and 18 cards were answered today
- **THEN** the nothing-due screen also shows the current streak and "12 more cards to reach today's goal"

#### Scenario: Study from the overview
- **WHEN** the user chooses Study now on a deck's overview
- **THEN** the session starts with that deck's due cards

### Requirement: Session completion
When no due cards remain and no learning card is due within the learn ahead limit, the session SHALL end with a summary showing the number of cards studied and the time spent, followed by the streak and goal lines described in the `progress` capability. If learning cards will become due later today, the summary SHALL say when, and SHALL react when they become due as described in "Learning cards becoming due while waiting".

#### Scenario: Finish
- **WHEN** the last due card is answered and no learning card is due within the session
- **THEN** the summary screen shows the number of cards studied and the time spent

#### Scenario: Finish with learning cards pending
- **WHEN** the learn ahead limit is 20 minutes, the last due card is answered and a learning card is due in 1 hour
- **THEN** the summary shows the cards studied, the time spent, and when the learning card will be due

#### Scenario: Finish with streak and goal
- **WHEN** the user finishes the first session of the day, the streak becomes 13 days and the goal of 30 cards is reached
- **THEN** the summary also shows "Streak extended to 13 days" and "Daily goal reached"
