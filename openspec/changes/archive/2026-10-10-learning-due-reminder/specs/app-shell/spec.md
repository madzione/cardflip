# Spec Delta

## ADDED Requirements

### Requirement: Due counts stay current over time
The deck list's due counts SHALL update without user action when they change because time passed: when a learning or relearning card becomes due, and when a new study day starts at the rollover hour. They SHALL also update when the app returns to the foreground. The update SHALL happen within a few seconds of the due time while the app is in the foreground.

#### Scenario: Learning card becomes due on the deck list
- **WHEN** the user rates a card Again with a 1-minute step, returns to the deck list, and waits one minute
- **THEN** that deck's learning count increases by one without the user tapping anything

#### Scenario: Day rollover
- **WHEN** the deck list is open across the rollover hour and review cards are due in the new study day
- **THEN** the review counts update to the new day's due cards

#### Scenario: Returning to the app
- **WHEN** the app was in the background while a learning card became due and the user brings it back to the foreground
- **THEN** the deck list shows the updated learning count

### Requirement: Learning due reminder
When learning cards become due while the app is in the foreground and the user is not in a study session, the app SHALL show an in-app reminder that names the deck and offers a Study action that opens that deck's study session. When cards become due in several top-level decks at once, the reminder SHALL say so and its action SHALL show the deck list. The app SHALL NOT show system notifications.

#### Scenario: Reminder on another tab
- **WHEN** the user is on the Browse tab and a learning card in deck "Spanish" becomes due
- **THEN** a reminder says learning cards are due in "Spanish" with a Study action

#### Scenario: Study from the reminder
- **WHEN** the user chooses Study on the reminder for "Spanish"
- **THEN** the app shows the Decks tab with a study session for "Spanish" open, starting with the due learning card

#### Scenario: No reminder while studying
- **WHEN** a learning card becomes due while a study session is open
- **THEN** no reminder is shown outside the study screen; the study screen handles it

#### Scenario: Several decks at once
- **WHEN** learning cards become due in the top-level decks "Spanish" and "Biology" at the same moment
- **THEN** one reminder says learning cards are due in 2 decks, and its action shows the deck list

## MODIFIED Requirements

### Requirement: App settings
Settings SHALL include theme mode, the day rollover hour, whether to continue automatically when learning cards become due during a study session (off by default), export backup, restore backup, import, and an About section showing the app version.

#### Scenario: Change rollover hour
- **WHEN** the user changes the rollover hour from 4 to 6
- **THEN** due counts and daily limits use 06:00 as the day boundary from then on

#### Scenario: Auto-continue setting persists
- **WHEN** the user turns on "Continue automatically when learning cards are due" and restarts the app
- **THEN** the setting is still on, and it is included in backups
