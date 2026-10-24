# Spec Delta

## ADDED Requirements

### Requirement: Deck descriptions
Each deck SHALL have an optional description that the user can edit, which SHALL be shown on the deck's overview as sanitized HTML, without running scripts.

#### Scenario: Edit description
- **WHEN** the user sets the description of `Spanish` to `Vocabulary from <b>Book 1</b>`
- **THEN** the overview shows the text with "Book 1" in bold

### Requirement: Unbury
The user SHALL be able to unbury a deck, which SHALL make every buried card in the deck and its subdecks available for study immediately, including cards held back today by sibling burying.

#### Scenario: Unbury manually buried cards
- **WHEN** the user buried 3 due cards of `Spanish` and then chooses Unbury on its overview
- **THEN** the 3 cards are due again and the overview's buried count is 0

#### Scenario: Unbury siblings
- **WHEN** the reverse card of a studied note is held back today by sibling burying and the user chooses Unbury
- **THEN** the reverse card is available in today's session
