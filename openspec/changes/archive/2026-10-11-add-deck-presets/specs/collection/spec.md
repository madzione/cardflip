# Spec Delta

## MODIFIED Requirements

### Requirement: Deck options
Deck options SHALL be kept in named presets. Each deck SHALL use exactly one preset, and decks using the same preset SHALL share its options. A preset SHALL include new cards per day (default 20), maximum reviews per day (default 200), desired retention (default 0.90, allowed range 0.70–0.99), and every other deck option. A child deck SHALL use its own preset independently of its parent. A new collection SHALL have the preset `Default`, used by every deck.

#### Scenario: Out-of-range retention
- **WHEN** the user sets desired retention to 0.5
- **THEN** the value is rejected and the previous value is kept

#### Scenario: Shared preset
- **WHEN** decks `Spanish` and `French` use preset `Languages` and the user sets its desired retention to 0.85 from `Spanish`
- **THEN** both decks use 0.85

#### Scenario: Existing options kept
- **WHEN** a collection whose deck `Biology` had a new card limit of 5 is opened by a version with presets
- **THEN** `Biology` still has a new card limit of 5 and its other options are unchanged

## ADDED Requirements

### Requirement: Manage presets
The user SHALL be able to add, clone, rename and delete presets, choose a deck's preset, and assign a deck's preset to all its subdecks. Deleting a preset SHALL move the decks using it to `Default`; `Default` SHALL NOT be deletable. Preset names SHALL be unique, compared case-insensitively. The options screen SHALL show how many decks use the preset.

#### Scenario: Save to all subdecks
- **WHEN** the user chooses "Save to all subdecks" on `Languages`, which uses preset `Languages`
- **THEN** every descendant of `Languages` uses preset `Languages`

#### Scenario: Delete a preset
- **WHEN** the user deletes preset `Exam`, used by 3 decks
- **THEN** those 3 decks use `Default`

### Requirement: This-deck limits
A deck SHALL optionally have its own new and review limits that override its preset's limits for that deck only.

#### Scenario: One deck lower
- **WHEN** `Spanish` and `French` use a preset with 20 new cards per day and `French` has a this-deck new limit of 5
- **THEN** `French` shows at most 5 new cards and `Spanish` up to 20
