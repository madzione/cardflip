# Spec Delta

## ADDED Requirements

### Requirement: Weekly recap
On the first time the deck list is shown in a new week, it SHALL show a card at the top with the previous week's study days, answers, study time and true retention, each compared to the week before. Weeks SHALL start on the first day of the week of the interface language and follow the study day boundary.

#### Scenario: New week
- **WHEN** it is Monday, last week the user studied on 6 days with 420 answers in 95 minutes and 91 % retention, and the week before on 4 days with 300 answers
- **THEN** the deck list shows a recap with 6 days, 420 answers, 95 minutes and 91 %, each with its change from the week before

#### Scenario: Open statistics
- **WHEN** the user taps the recap
- **THEN** the collection statistics open

#### Scenario: Nothing studied last week
- **WHEN** the user did not study in the previous week
- **THEN** no recap is shown

### Requirement: Dismiss the recap
The user SHALL be able to dismiss the recap. A dismissed recap SHALL stay hidden on this device until the next week. When *Show streaks and goals* is off, no recap SHALL be shown.

#### Scenario: Dismissed
- **WHEN** the user dismisses this week's recap and restarts the app
- **THEN** the deck list shows no recap until next week

### Requirement: Celebrations
When the daily goal was reached or an achievement tier was earned during a session, the session-complete screen SHALL play a short animation of at most two seconds that does not block any action. There SHALL be no animation during studying, when reduce motion is in effect, or when *Show streaks and goals* is off.

#### Scenario: Goal reached
- **WHEN** the session reached the daily goal
- **THEN** the summary plays a short celebration once while its buttons stay usable

#### Scenario: Reduce motion
- **WHEN** reduce motion is in effect and the session reached the daily goal
- **THEN** the summary shows "Daily goal reached" without an animation

#### Scenario: Goal already reached earlier
- **WHEN** the goal was reached in the morning and the user finishes another session in the evening
- **THEN** no celebration is played
