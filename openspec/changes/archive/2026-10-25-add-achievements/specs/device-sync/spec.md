# Spec Delta

## MODIFIED Requirements

### Requirement: What is synced
Sync SHALL carry decks, deck descriptions and limits, presets with their options, note types, notes, cards with their scheduling state, flags, suspension and burying, filtered decks, the review history, deletions of all of these, media files, and the collection settings that affect studying and card content, including the daily goal and rest days per week. It SHALL NOT carry device settings, including whether streaks and goals are shown and which achievements the device has already shown.

#### Scenario: Synced study settings
- **WHEN** the user changes the rollover hour to 6 on the laptop and syncs
- **THEN** the phone uses 6 as the rollover hour

#### Scenario: Device settings stay
- **WHEN** the laptop uses dark mode and German, the phone uses light mode and English, and they sync
- **THEN** each keeps its own theme, language, UI size, last-used deck and note type, and MCP settings

#### Scenario: Daily goal synced
- **WHEN** the user sets a daily goal of 30 cards on the laptop and syncs
- **THEN** the phone shows progress toward 30 cards

#### Scenario: Visibility stays
- **WHEN** the user turns off *Show streaks and goals* on the phone and syncs
- **THEN** the laptop still shows streaks and goals

#### Scenario: Rest days synced
- **WHEN** the user sets 1 rest day per week on the phone and syncs
- **THEN** the laptop computes streaks with 1 rest day per week

#### Scenario: Achievements shown per device
- **WHEN** the phone has shown the notice for Answers 1,000 and the devices sync
- **THEN** the laptop still shows that notice once
