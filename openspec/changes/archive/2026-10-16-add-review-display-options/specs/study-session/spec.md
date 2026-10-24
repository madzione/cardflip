# Spec Delta

## MODIFIED Requirements

### Requirement: Keyboard control
On platforms with a keyboard, Space or Enter SHALL reveal the answer and then, while the setting "Space/Enter also answers Good" is on, choose Good. Keys 1–4 SHALL choose Again, Hard, Good and Easy. Ctrl+Z (Cmd+Z on macOS) SHALL undo. Escape SHALL leave the session.

#### Scenario: Keyboard review
- **WHEN** the user presses Space, then 3
- **THEN** the answer is revealed and the card is rated Good

#### Scenario: Space does not rate
- **WHEN** "Space/Enter also answers Good" is off and the user presses Space twice on a new card
- **THEN** the answer is revealed and the card is not rated

### Requirement: Rate with interval previews
Each rating button SHALL show the interval the card would receive (for example `<10m`, `3d`, `1.2mo`), unless the setting "Show next review time above answer buttons" is off. Choosing a rating SHALL schedule the card, record how long the user took to answer, and advance to the next card.

#### Scenario: Rate Good
- **WHEN** the user presses Good, which shows `3d`
- **THEN** the card is due in 3 days and the next card's question is shown

#### Scenario: Intervals hidden
- **WHEN** the setting is off and the answer is shown
- **THEN** the buttons show only Again, Hard, Good and Easy

## ADDED Requirements

### Requirement: Remaining count visibility
When the setting "Show remaining card count" is off, the study screen SHALL NOT show the remaining new, learning and review counts.

#### Scenario: Counts hidden
- **WHEN** the setting is off and a session is open
- **THEN** no remaining counts are visible

### Requirement: Hidden bars during review
When the setting "Hide bars during review" is on, the study screen SHALL hide the app's navigation and app bar while keeping the card, the counts (if shown) and the buttons. The user SHALL still be able to leave the session with Escape or a control at the top edge.

#### Scenario: Distraction-free
- **WHEN** the setting is on and the user starts a session on a phone
- **THEN** the bottom navigation and app bar are hidden until the session is left

### Requirement: Question fade
Each new question SHALL fade in over the configured question fade time; 0 SHALL show it immediately.

#### Scenario: No fade
- **WHEN** the question fade is 0 and the next card is shown
- **THEN** its question appears without animation

### Requirement: Scroll to answer
When the answer is revealed and contains an element with id `answer`, and the card is taller than the visible area, the card SHALL scroll so that element is at the top.

#### Scenario: Long front side
- **WHEN** the answer template is `{{FrontSide}}<hr id=answer>{{Back}}` and the front side fills the screen
- **THEN** on reveal the view scrolls to the divider and shows the Back field
