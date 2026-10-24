# Spec Delta

## MODIFIED Requirements

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

### Requirement: Lapses
Rating a Review card Again SHALL increment its lapse count and move it to Relearning on the first relearning step. When the deck's relearning steps are empty, the delays SHALL come from FSRS's short-term scheduling as for empty learning steps.

#### Scenario: Forgotten review card
- **WHEN** a Review card with 2 lapses is rated Again in a deck with the default relearning step
- **THEN** it has 3 lapses, is in Relearning, and is due again within the same day

## ADDED Requirements

### Requirement: Interday learning
A learning or relearning step that ends on a later study day SHALL make the card due on that study day instead of at an exact time. Such a card SHALL be shown from the start of that day, SHALL count as learning in due counts, and SHALL count against the review limit.

#### Scenario: One-day step
- **WHEN** a deck's learning steps are `10m 1d` and a card is rated Good on its 10-minute step at 15:00
- **THEN** it is due from the start of the next study day, not at 15:00 the next day
