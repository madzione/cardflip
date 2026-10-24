# scheduling Specification

## Purpose

Decides when each card is next shown, using the FSRS spaced-repetition algorithm, and which cards are due today within the deck's daily limits.

## Requirements

### Requirement: Card states
Every card SHALL be in exactly one scheduling state: New, Learning, Review or Relearning. A card SHALL also record its due time, its FSRS memory state (stability and difficulty, absent while New), its review count, its lapse count and its last review time.

#### Scenario: New card has no memory state
- **WHEN** a card is created
- **THEN** it is in state New, has no stability or difficulty, and has review and lapse counts of 0

### Requirement: FSRS scheduling
When a card is rated Again, Hard, Good or Easy, the system SHALL compute its new state, memory state and due time with FSRS, using the parameters of the card's deck or the default parameters when none are set, so that the predicted recall probability at the due time equals the deck's desired retention.

#### Scenario: Higher retention means shorter intervals
- **WHEN** two identical review cards are rated Good, one in a deck with desired retention 0.95 and one in a deck with 0.85
- **THEN** the card in the 0.95 deck gets a shorter interval

#### Scenario: Ratings are ordered
- **WHEN** the intervals for Again, Hard, Good and Easy are computed for the same card
- **THEN** each interval is greater than or equal to the previous one, and Easy's is strictly greater than Again's

### Requirement: Lapses
Rating a Review card Again SHALL increment its lapse count and move it to Relearning on the first relearning step. When the deck's relearning steps are empty, the delays SHALL come from FSRS's short-term scheduling as for empty learning steps.

#### Scenario: Forgotten review card
- **WHEN** a Review card with 2 lapses is rated Again in a deck with the default relearning step
- **THEN** it has 3 lapses, is in Relearning, and is due again within the same day

### Requirement: Learning steps
New cards SHALL go through the deck's learning steps and Relearning cards through its relearning steps before they graduate to Review (defaults: 1 minute and 10 minutes for learning, 10 minutes for relearning). Again SHALL return to the first step, Hard SHALL repeat the current step (on the first step, with the average of the first two steps), Good SHALL move to the next step, and Easy SHALL graduate immediately. When the deck's step list is empty, the delays SHALL come from FSRS's short-term scheduling until it schedules a delay of at least one day. Intraday step delays SHALL get a random extra delay of up to 25 % of the step and at most 5 minutes, and the interval preview SHALL include it.

#### Scenario: New card rated Good twice
- **WHEN** a New card in a deck with the default steps is rated Good, and then rated Good again once its 10-minute step is due
- **THEN** after the second rating it is in Review with a due date at least one day away

#### Scenario: Custom steps
- **WHEN** a deck's learning steps are `1m 10m 1h` and a New card is rated Good twice
- **THEN** the card is due again about one hour later and is still learning

#### Scenario: Step fuzz bounded
- **WHEN** a learning card is rated Good onto a 10-minute step
- **THEN** it is due between 10 and 12.5 minutes later, and the Good button showed that delay

### Requirement: Interday learning
A learning or relearning step that ends on a later study day SHALL make the card due on that study day instead of at an exact time. Such a card SHALL be shown from the start of that day, SHALL count as learning in due counts, and SHALL count against the review limit.

#### Scenario: One-day step
- **WHEN** a deck's learning steps are `10m 1d` and a card is rated Good on its 10-minute step at 15:00
- **THEN** it is due from the start of the next study day, not at 15:00 the next day

### Requirement: Interval preview
For any card, the system SHALL provide the next due interval for each of the four ratings without changing the card.

#### Scenario: Preview is side-effect free
- **WHEN** the previews for a card are requested and then the card is rated Good
- **THEN** the card's resulting interval equals the Good preview

### Requirement: Review log
Every rating SHALL append an entry to the review history: card, time, rating, state before, interval before and after, and time taken to answer. The entry SHALL record enough of the previous state that the rating can be undone exactly.

#### Scenario: Log entry written
- **WHEN** the user rates a card Hard after 7 seconds
- **THEN** the review history contains an entry for that card with rating Hard and a duration of 7 seconds

### Requirement: Day boundary
The study day SHALL roll over at a configurable hour in the device's local time (default 04:00). Daily limits, burying and "due today" SHALL be measured against this boundary.

#### Scenario: Late-night review counts for the previous day
- **WHEN** the rollover hour is 04:00 and the user reviews at 02:30
- **THEN** that review counts toward the previous calendar day's limits

### Requirement: Due cards and daily limits
A deck's due cards SHALL be: Learning and Relearning cards due by now; Review cards due by the end of the current study day, up to the effective review limit minus reviews already done today; and New cards, up to the effective new limit minus new cards already introduced today. The effective limit SHALL be the deck's today-only limit when set for the current day, otherwise its daily limit. When the deck's "new cards ignore review limit" option is off, new cards SHALL also be limited to the review limit's remaining room after due reviews. Counts for a parent deck SHALL include its children, and each child's own limits SHALL also apply. When studying a subdeck whose "limits start from top" option is on, the limits of its parent decks SHALL also apply.

#### Scenario: New limit reached
- **WHEN** a deck has a new-per-day limit of 20, 50 new cards, and 20 new cards have been introduced today
- **THEN** the deck shows 0 new cards due until the next rollover

#### Scenario: Learning ignores limits
- **WHEN** the review limit for today is used up and a Relearning card becomes due
- **THEN** the Relearning card is still shown

#### Scenario: Today-only increase
- **WHEN** 20 of 20 new cards were introduced today and the user increases today's new card limit by 10
- **THEN** the deck shows 10 new cards due

#### Scenario: Review limit caps new cards
- **WHEN** "new cards ignore review limit" is off, the review limit is 100, 95 reviews are due and the new limit is 20
- **THEN** the deck shows 95 review and 5 new cards

#### Scenario: Limits from the top
- **WHEN** `Languages` has a new limit of 10, `Languages::Spanish` has 20 and "limits start from top" is on for `Spanish`, and the user studies `Spanish`
- **THEN** at most 10 new cards are shown

### Requirement: Order of new cards
New cards SHALL be selected for the day according to the studied deck's new card gather order, and shown according to its new card sort order; with the defaults, they are introduced in the order they were added to the collection (import order for imported cards). Random orders SHALL stay the same within a study day. Once a card of a note has been studied today, the note's other New cards SHALL NOT be shown that day when their deck buries new siblings, and its other Review cards SHALL NOT be shown that day when their deck buries review siblings (sibling burying).

#### Scenario: Sibling buried
- **WHEN** a deck buries new siblings and the forward card of a new reversed note is studied today
- **THEN** the reverse card is not shown again until the next study day

#### Scenario: Siblings by default
- **WHEN** a deck uses the default options and the first card of a new cloze note with three deletions is studied
- **THEN** the note's other two cards can be shown on the same day

#### Scenario: Review siblings allowed
- **WHEN** a deck does not bury review siblings and both cards of a reversed note are due for review today
- **THEN** both cards are shown today

#### Scenario: Random notes
- **WHEN** the gather order is "random notes" and the new limit is 10 of 500 new cards
- **THEN** today's 10 new cards come from randomly chosen notes, and the selection does not change after answering one

### Requirement: Leech detection
When rating a Review card Again brings its lapse count to the deck's leech threshold, or to the threshold plus a multiple of half the threshold rounded up, the card's note SHALL get the tag `leech`, and with the suspend action the card SHALL be suspended, in the same operation as the answer.

#### Scenario: Threshold reached
- **WHEN** the threshold is 8, the action is suspend, and a card with 7 lapses is rated Again
- **THEN** the card has 8 lapses, is suspended, and its note has the tag `leech`

#### Scenario: Repeat warning
- **WHEN** the threshold is 8, the action is tag only, and the card reaches 12 lapses
- **THEN** it is detected as a leech again

### Requirement: Interval limits
No review interval, including the Easy interval and fuzz, SHALL exceed the deck's maximum interval. A card graduating from relearning SHALL get at least the deck's minimum interval after relearning. The interval previews SHALL reflect both limits.

#### Scenario: Capped interval
- **WHEN** the deck's maximum interval is 180 days and a mature card's Easy interval would be 400 days
- **THEN** Easy shows and gives 180 days

#### Scenario: Relearned card
- **WHEN** the minimum interval after relearning is 3 days and a relearning card graduates with a computed interval of 1 day
- **THEN** it is due in 3 days

### Requirement: Set due date
The user SHALL be able to set cards to be due on a chosen study day: `N` days from today (`0` = today), or a random day per card in `N-M`. A trailing `!` SHALL also set the card's interval to the days until that due date; without it, a review card's interval SHALL stay unchanged. Invalid input SHALL be rejected with an explanation, and no card SHALL change.

#### Scenario: Postpone
- **WHEN** a review card with a 10-day interval, due today, is set to `7`
- **THEN** it is due in 7 days and its interval is still 10 days

#### Scenario: Range
- **WHEN** 100 new cards are set to `1-10`
- **THEN** they are Review cards, each due on a day from tomorrow to 10 days from now, spread across that range

#### Scenario: Interval too
- **WHEN** a review card is set to `30!`
- **THEN** it is due in 30 days with an interval of 30 days

#### Scenario: Invalid input
- **WHEN** the user enters `10-5`
- **THEN** the input is rejected and no card changes

### Requirement: Card state after setting a due date
Cards whose due date is set SHALL become Review cards, whatever their previous state. Memory states SHALL be kept, and a New card SHALL get the memory state of a first Good rating. Each change SHALL add a manual entry to the card's review history.

#### Scenario: New card scheduled
- **WHEN** a New card is set to `2`
- **THEN** it is a Review card due in 2 days, with a stability and difficulty, and its history has one manual entry

### Requirement: Capped answer time
The answer time recorded in the review history SHALL be capped at the maximum answer seconds of the card's deck.

#### Scenario: Long pause
- **WHEN** the deck's maximum is 60 seconds and the user answers a card after 5 minutes
- **THEN** the review entry records 60 seconds

### Requirement: Optimize parameters
The user SHALL be able to optimize a deck's FSRS parameters, or all decks' parameters, from the review history selected by its search. Optimizing SHALL require at least 400 reviews, SHALL keep the current parameters unless the new ones fit the history better, SHALL run without blocking the app, SHALL show progress, and SHALL be cancellable. An optional health check SHALL report unreliable results instead of applying them.

#### Scenario: Optimize
- **WHEN** a deck has 5,000 reviews and the user chooses Optimize
- **THEN** progress is shown, and afterwards the deck has parameters that fit its history at least as well as the previous ones

#### Scenario: Not enough reviews
- **WHEN** a deck has 150 reviews and the user chooses Optimize
- **THEN** the app explains that at least 400 reviews are needed and the parameters are unchanged

### Requirement: Due day choice with Easy Days
When a review interval allows a range of due days through fuzz, the due day SHALL be chosen within that range so that Reduced weekdays receive fewer reviews than Normal days and Minimum weekdays receive reviews only when the range contains no other day, taking into account the reviews already due on each day. The interval previews SHALL show the chosen day.

#### Scenario: Avoid a minimum day
- **WHEN** Sunday is Minimum and a card's interval range spans Friday to Monday
- **THEN** the card is not scheduled on Sunday

#### Scenario: Only a minimum day
- **WHEN** Sunday is Minimum and a card's interval range is only Sunday
- **THEN** the card is scheduled on Sunday

#### Scenario: No Easy Days set
- **WHEN** every weekday is Normal
- **THEN** due days are chosen as before
