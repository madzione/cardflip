# study-session Specification

## Purpose

The review experience: studying the due cards of a deck one at a time, revealing answers, rating recall, undoing mistakes, and finishing the session.

## Requirements

### Requirement: Start a session
The user SHALL be able to start studying any deck that has due cards, from the deck's overview or directly from the deck list's study action. The session SHALL include the due cards of the deck and its children, according to the scheduling rules. Starting a session directly for a deck with nothing due SHALL show a "congratulations, nothing due" screen that says when the next card is due.

#### Scenario: Nothing due
- **WHEN** the user starts studying a deck with no due cards from the deck list's study action and the next card is due tomorrow
- **THEN** the app shows that studying is complete for now and that the next card is due tomorrow

#### Scenario: Study from the overview
- **WHEN** the user chooses Study now on a deck's overview
- **THEN** the session starts with that deck's due cards

### Requirement: Card queue order
Within a session, cards due from learning SHALL take priority once due. Otherwise review cards and new cards SHALL be interleaved, shown before, or shown after each other according to the deck's new/review order, and review cards SHALL follow its review sort order. The session SHALL show remaining counts for new, learning and review cards, and update them after every answer.

#### Scenario: Counts decrease
- **WHEN** the session shows 5 review cards remaining and the user answers a review card Good
- **THEN** the review count shows 4

#### Scenario: New cards after reviews
- **WHEN** the new/review order is "after reviews" and a session has 10 review and 5 new cards
- **THEN** all 10 review cards are shown before the first new card

#### Scenario: Most at risk first
- **WHEN** the review sort order is "ascending retrievability"
- **THEN** due review cards are shown starting with the one least likely to be recalled

### Requirement: Interday learning order
Interday learning cards SHALL be shown mixed with review cards, before them, or after them, according to the deck's interday learning/review order.

#### Scenario: Before reviews
- **WHEN** the order is "before reviews" and a session has 2 interday learning and 10 review cards due
- **THEN** the 2 interday learning cards are shown before any review card

### Requirement: Reveal answer
Each card SHALL first show its question only. Revealing SHALL show the answer side, and only then SHALL the four rating buttons appear.

#### Scenario: Reveal
- **WHEN** the user taps "Show answer"
- **THEN** the answer is shown together with the Again, Hard, Good and Easy buttons

### Requirement: Rate with interval previews
Each rating button SHALL show the interval the card would receive (for example `<10m`, `3d`, `1.2mo`), unless the setting "Show next review time above answer buttons" is off. Choosing a rating SHALL schedule the card, record how long the user took to answer, and advance to the next card.

#### Scenario: Rate Good
- **WHEN** the user presses Good, which shows `3d`
- **THEN** the card is due in 3 days and the next card's question is shown

#### Scenario: Intervals hidden
- **WHEN** the setting is off and the answer is shown
- **THEN** the buttons show only Again, Hard, Good and Easy

### Requirement: Undo
The user SHALL be able to undo the most recent answers in the current session, one at a time. Undo SHALL restore the card's previous scheduling state, remove the review log entry, and show that card's question again.

#### Scenario: Undo last answer
- **WHEN** the user rates a card Again by mistake and chooses Undo
- **THEN** the card is restored to its exact prior state and is shown again unrevealed

### Requirement: Keyboard control
On platforms with a keyboard, Space or Enter SHALL reveal the answer and then, while the setting "Space/Enter also answers Good" is on, choose Good. Keys 1–4 SHALL choose Again, Hard, Good and Easy. Ctrl+Z (Cmd+Z on macOS) SHALL undo. Escape SHALL leave the session.

#### Scenario: Keyboard review
- **WHEN** the user presses Space, then 3
- **THEN** the answer is revealed and the card is rated Good

#### Scenario: Space does not rate
- **WHEN** "Space/Enter also answers Good" is off and the user presses Space twice on a new card
- **THEN** the answer is revealed and the card is not rated

### Requirement: Card actions during study
While studying, the user SHALL be able to edit the current note, suspend the current card, or bury the current card. Each of these SHALL take effect immediately and advance the queue where applicable.

#### Scenario: Edit while studying
- **WHEN** the user edits the current card's note to fix a typo
- **THEN** the corrected content is shown when they return to the card

### Requirement: Session completion
When no due cards remain and no learning card is due within the learn ahead limit, the session SHALL end with a summary showing the number of cards studied and the time spent. If learning cards will become due later today, the summary SHALL say when, and SHALL react when they become due as described in "Learning cards becoming due while waiting".

#### Scenario: Finish
- **WHEN** the last due card is answered and no learning card is due within the session
- **THEN** the summary screen shows the number of cards studied and the time spent

#### Scenario: Finish with learning cards pending
- **WHEN** the learn ahead limit is 20 minutes, the last due card is answered and a learning card is due in 1 hour
- **THEN** the summary shows the cards studied, the time spent, and when the learning card will be due

### Requirement: Session survives interruption
Every answer SHALL be saved as soon as it is given, so closing or killing the app mid-session loses at most the card currently on screen.

#### Scenario: App killed mid-session
- **WHEN** the user answers 10 cards and the app process is killed
- **THEN** after restart those 10 answers are preserved and the due counts reflect them

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

### Requirement: Session counts stay current
While a card is on screen, the session's remaining counts SHALL update when a learning card becomes due, without changing or un-revealing the card on screen.

#### Scenario: Learning count during a review
- **WHEN** the user is looking at the revealed answer of a review card and a learning card becomes due
- **THEN** the learning count increases by one and the same answer stays on screen

### Requirement: Empty cards during study
When the card being studied is an empty card, the study screen SHALL show a notice that the card is empty and can be removed with Empty Cards, together with the normal reveal and rating controls.

#### Scenario: Empty cloze card
- **WHEN** the card for a removed cloze number comes up for review
- **THEN** the screen says the card is empty and how to remove it, and the user can still rate it

### Requirement: Typing the answer
When the question contains a `type:` filter, the study screen SHALL show a text input in its place and focus it. Enter in the input SHALL reveal the answer, as SHALL "Show answer". While the input has focus, typed characters SHALL go into the input and SHALL NOT trigger study shortcuts. After the answer is revealed, the rating shortcuts SHALL work again.

#### Scenario: Type and reveal
- **WHEN** the user types `pero` into the input and presses Enter
- **THEN** the answer is shown with the comparison of `pero` against the expected text, together with the rating buttons

#### Scenario: Digits are typed
- **WHEN** the input has focus and the user presses 3
- **THEN** `3` is typed into the input and the card is not rated

### Requirement: Flag the current card
While studying, the user SHALL be able to set or clear a flag on the current card from the study menu, and with Ctrl+1 to Ctrl+7 (Cmd on macOS). Choosing the flag the card already has SHALL clear it. The current card's flag SHALL be visible on the study screen as a marker in the flag's colour.

#### Scenario: Toggle with keyboard
- **WHEN** the user presses Ctrl+1 on a card without a flag, and then Ctrl+1 again
- **THEN** the card first shows a red flag marker, and then no marker

#### Scenario: Change flag
- **WHEN** the current card has flag 1 and the user presses Ctrl+3
- **THEN** the card has flag 3

### Requirement: Leech message and undo
When an answer makes a card a leech, the study screen SHALL show a message saying so, and that the card was suspended when it was. Undoing that answer SHALL also undo the suspension and remove the `leech` tag if that answer added it.

#### Scenario: Message
- **WHEN** an answer makes the current card a leech with the suspend action
- **THEN** a message says the card is a leech and was suspended

#### Scenario: Undo leech
- **WHEN** the user undoes that answer
- **THEN** the card is unsuspended, its note no longer has the tag `leech`, and the card is shown again

### Requirement: Card info during study
While studying, `I` and a menu item SHALL open card info for the current card, and Ctrl+Alt+I (Cmd+Option+I on macOS) SHALL open it for the previously answered card. Opening card info SHALL NOT change the session, and closing it SHALL return to the same card in the same revealed state.

#### Scenario: Previous card
- **WHEN** the user rates a card Again and presses Ctrl+Alt+I
- **THEN** card info for that card opens, and its history includes the Again rating

#### Scenario: Return unchanged
- **WHEN** the user opens card info on a revealed answer and closes it
- **THEN** the same answer is still shown with the rating buttons

### Requirement: Mark during study
While studying, `*` and a menu item SHALL toggle the mark on the current card's note, and a marked card SHALL show a star.

#### Scenario: Toggle with key
- **WHEN** the user presses `*` on an unmarked card, and again
- **THEN** a star appears and the note has `marked`, and then the star and tag are gone

### Requirement: Set due date during study
While studying, a menu item and Ctrl+Shift+D (Cmd+Shift+D on macOS) SHALL set the current card's due date. When the card is no longer due today, the session SHALL advance to the next card.

#### Scenario: Push a card to next week
- **WHEN** the user sets the current card's due date to `7`
- **THEN** the next card is shown and the review count decreases by one

### Requirement: Learn ahead
When no learning, review or new card is due in a session, the session SHALL show the learning card that becomes due soonest, if it is due within the learn ahead limit. With a limit of 0, learning cards SHALL only be shown once due.

#### Scenario: Finish without waiting
- **WHEN** the limit is 20 minutes, the last review card is answered and a learning card is due in 5 minutes
- **THEN** the learning card is shown immediately

#### Scenario: Limit off
- **WHEN** the limit is 0 and the only remaining card is a learning card due in 5 minutes
- **THEN** the session shows when that card will be due instead of the card

### Requirement: Note actions during study
While studying, the user SHALL be able to bury all cards of the current note, suspend all cards of the current note, and delete the current note with all its cards after a confirmation that states how many cards will be deleted. Each SHALL take effect immediately and advance to the next card.

#### Scenario: Suspend note
- **WHEN** the user chooses Suspend note on the forward card of a `Basic (and reversed card)` note
- **THEN** both cards are suspended and the next card is shown

#### Scenario: Delete note
- **WHEN** the user chooses Delete note on a card whose note has 2 cards and confirms
- **THEN** the note and both cards are deleted and the next card is shown

### Requirement: Create copy during study
While studying, the user SHALL be able to open the add screen prefilled with the current note's note type, deck, fields and tags. Closing the add screen SHALL return to the same card.

#### Scenario: Copy and adjust
- **WHEN** the user chooses Create copy, changes the Front field and adds the card
- **THEN** a new note is added and the session shows the same card as before

### Requirement: Deck options during study
While studying, `O` and a menu item SHALL open the options of the deck being studied. Changed options SHALL apply to the rest of the session.

#### Scenario: Lower the new card limit
- **WHEN** the user opens the options during study and lowers new cards per day below the number already introduced today
- **THEN** after closing the options, no further new cards are shown in the session

### Requirement: Card action keys
On platforms with a keyboard, `E` SHALL edit the current card, `-` SHALL bury it, `@` SHALL suspend it, `=` SHALL bury its note, `!` SHALL suspend its note, and Ctrl+Delete (Cmd+Delete on macOS) SHALL delete its note after confirmation.

#### Scenario: Bury with key
- **WHEN** the user presses `-` on a card
- **THEN** the card is buried and the next card is shown

### Requirement: Reset during study
While studying, the user SHALL be able to reset the current card to New with the reset options, which SHALL take effect immediately and advance to the next card.

#### Scenario: Reset from study
- **WHEN** the user chooses Reset card in the study menu and confirms with the default options
- **THEN** the card is New in its original position and the next card is shown

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

### Requirement: Auto advance
The user SHALL be able to turn auto advance on and off for the current session from the study menu and with Shift+A; it SHALL be off when a session starts. While on, after the deck's question time the question action SHALL run, and after the answer time the answer action SHALL run; a time of 0 SHALL disable that side. User input SHALL restart the countdown for the new state. The study screen SHALL show that auto advance is on.

#### Scenario: Hands-free review
- **WHEN** auto advance is on with 5 seconds for the question, 3 for the answer and the answer action "answer Good"
- **THEN** each card's answer appears after 5 seconds, and 3 seconds later the card is rated Good and the next card appears

#### Scenario: Off by default
- **WHEN** a session starts for a deck with auto advance times set
- **THEN** nothing happens until the user turns auto advance on

### Requirement: Speaking during study
When a question or answer with utterances is shown, the study screen SHALL speak them automatically in order of appearance. Showing the answer or leaving the card SHALL stop speech from the previous side. The user SHALL be able to replay a single utterance with its button, and all utterances of the current side with R or F5. When the device has no voice for an utterance's language, its button SHALL be disabled with an explanation, and nothing else SHALL be affected.

#### Scenario: Autoplay on question
- **WHEN** a card whose question has `{{tts es_ES:Front}}` is shown
- **THEN** the Front text is spoken without user action

#### Scenario: Replay with keyboard
- **WHEN** the answer is shown and the user presses R
- **THEN** the answer side's utterances are spoken again

#### Scenario: No voice
- **WHEN** an utterance is in a language with no voice on the device
- **THEN** its replay button is disabled and explains that no voice is installed for that language

### Requirement: Card shown early or again
A card shown before it is due because nothing else is due (learn ahead) SHALL have a notice on its question saying so and when it is due. The card just answered, shown again next, SHALL have a notice saying so, and that it is shown early or due again now; not after undo. The notice SHALL NOT cover the card or the buttons, and SHALL disappear when the answer or another card is shown.

#### Scenario: Shown early
- **WHEN** the learn ahead limit is 20 minutes, nothing else is due, and the next card is a learning card due in 11 minutes, another card than the one just answered
- **THEN** its question is shown with a notice that it is shown early, due in 11 minutes, because nothing else is due

#### Scenario: This card again, shown early
- **WHEN** the learn ahead limit is 20 minutes, the only card left is a new card, and the user answers it Good, which makes it due in 11 minutes
- **THEN** the card's question is shown again with a notice that it is this card again, due in 11 minutes and shown early because nothing else is due

#### Scenario: Due again now
- **WHEN** the user answers a learning card Again and it is the next card because it is already due
- **THEN** its question is shown with a notice that it is due again now

#### Scenario: Notice goes away
- **WHEN** the notice is shown and the user shows the answer
- **THEN** the notice is gone

#### Scenario: Another card that is due
- **WHEN** the card after an answer is a different card that is due, also a sibling of the answered note
- **THEN** no notice is shown

#### Scenario: Undo
- **WHEN** the user answers a card and undoes the answer
- **THEN** the card's question is shown without the notice
