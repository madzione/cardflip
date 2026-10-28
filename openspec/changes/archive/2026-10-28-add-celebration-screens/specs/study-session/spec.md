# Spec Delta

## MODIFIED Requirements

### Requirement: Session completion
When no due cards remain and no learning card is due within the learn ahead limit, the session SHALL end with a summary showing the number of cards studied and the time spent, followed by the streak and goal lines described in the `progress` capability. When the session earned something to celebrate, the celebration sequence described in the `progress` capability SHALL be shown before the summary. If learning cards will become due later today, the summary SHALL say when, and SHALL react when they become due as described in "Learning cards becoming due while waiting".

#### Scenario: Finish
- **WHEN** the last due card is answered and no learning card is due within the session
- **THEN** the summary screen shows the number of cards studied and the time spent

#### Scenario: Finish with learning cards pending
- **WHEN** the learn ahead limit is 20 minutes, the last due card is answered and a learning card is due in 1 hour
- **THEN** the summary shows the cards studied, the time spent, and when the learning card will be due

#### Scenario: Finish with streak and goal
- **WHEN** the user finishes the first session of the day, the streak becomes 13 days and the goal of 30 cards is reached
- **THEN** a streak page and a goal page are shown, and then the summary also shows "Streak extended to 13 days" and the reward card says "Daily goal reached"

### Requirement: Keyboard control
On platforms with a keyboard, Space or Enter SHALL reveal the answer and then, while the setting "Space/Enter also answers Good" is on, choose Good. Keys 1–4 SHALL choose Again, Hard, Good and Easy. Ctrl+Z (Cmd+Z on macOS) SHALL undo. Escape SHALL leave the session, except on a celebration page, where it behaves as described in the `progress` capability.

#### Scenario: Keyboard review
- **WHEN** the user presses Space, then 3
- **THEN** the answer is revealed and the card is rated Good

#### Scenario: Space does not rate
- **WHEN** "Space/Enter also answers Good" is off and the user presses Space twice on a new card
- **THEN** the answer is revealed and the card is not rated

#### Scenario: Escape on the level page
- **WHEN** the level page is shown during a session and the user presses Escape
- **THEN** the session goes on with the next card

