# Spec Delta

## ADDED Requirements

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
