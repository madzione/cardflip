# Spec Delta

## Purpose

The review experience: studying the due cards of a deck one at a time, revealing answers, rating recall, undoing mistakes, and finishing the session.

## ADDED Requirements

### Requirement: Start a session
The user SHALL be able to start studying any deck that has due cards. The session SHALL include the due cards of the deck and its children, according to the scheduling rules. Selecting a deck with nothing due SHALL show a "congratulations, nothing due" screen that says when the next card is due.

#### Scenario: Nothing due
- **WHEN** the user opens a deck with no due cards and the next card is due tomorrow
- **THEN** the app shows that studying is complete for now and that the next card is due tomorrow

### Requirement: Card queue order
Within a session, cards due from learning SHALL take priority once due. Otherwise review cards and new cards SHALL be interleaved. The session SHALL show remaining counts for new, learning and review cards, and update them after every answer.

#### Scenario: Counts decrease
- **WHEN** the session shows 5 review cards remaining and the user answers a review card Good
- **THEN** the review count shows 4

### Requirement: Reveal answer
Each card SHALL first show its question only. Revealing SHALL show the answer side, and only then SHALL the four rating buttons appear.

#### Scenario: Reveal
- **WHEN** the user taps "Show answer"
- **THEN** the answer is shown together with the Again, Hard, Good and Easy buttons

### Requirement: Rate with interval previews
Each rating button SHALL show the interval the card would receive (for example `<10m`, `3d`, `1.2mo`). Choosing a rating SHALL schedule the card, record how long the user took to answer, and advance to the next card.

#### Scenario: Rate Good
- **WHEN** the user presses Good, which shows `3d`
- **THEN** the card is due in 3 days and the next card's question is shown

### Requirement: Undo
The user SHALL be able to undo the most recent answers in the current session, one at a time. Undo SHALL restore the card's previous scheduling state, remove the review log entry, and show that card's question again.

#### Scenario: Undo last answer
- **WHEN** the user rates a card Again by mistake and chooses Undo
- **THEN** the card is restored to its exact prior state and is shown again unrevealed

### Requirement: Keyboard control
On platforms with a keyboard, Space or Enter SHALL reveal the answer and then choose Good. Keys 1–4 SHALL choose Again, Hard, Good and Easy. Ctrl+Z (Cmd+Z on macOS) SHALL undo. Escape SHALL leave the session.

#### Scenario: Keyboard review
- **WHEN** the user presses Space, then 3
- **THEN** the answer is revealed and the card is rated Good

### Requirement: Card actions during study
While studying, the user SHALL be able to edit the current note, suspend the current card, or bury the current card. Each of these SHALL take effect immediately and advance the queue where applicable.

#### Scenario: Edit while studying
- **WHEN** the user edits the current card's note to fix a typo
- **THEN** the corrected content is shown when they return to the card

### Requirement: Session completion
When no due cards remain, the session SHALL end with a summary showing the number of cards studied and the time spent. If learning cards will become due soon, the summary SHALL say so.

#### Scenario: Finish
- **WHEN** the last due card is answered and no learning card is due within the session
- **THEN** the summary screen shows the number of cards studied and the time spent

### Requirement: Session survives interruption
Every answer SHALL be saved as soon as it is given, so closing or killing the app mid-session loses at most the card currently on screen.

#### Scenario: App killed mid-session
- **WHEN** the user answers 10 cards and the app process is killed
- **THEN** after restart those 10 answers are preserved and the due counts reflect them
