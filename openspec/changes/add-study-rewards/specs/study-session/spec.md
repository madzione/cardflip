# Spec Delta

## MODIFIED Requirements

### Requirement: New achievements on the summary
When achievement tiers were earned during the session, the session-complete screen SHALL name them in the reward card described in the `progress` capability, including tiers a reward banner already showed during the session. A Level tier SHALL be left out of that text when the card already says the level was reached. Earning a tier SHALL NOT pause the session or open a dialog.

#### Scenario: Earned during the session
- **WHEN** the user's 1,000th answer is in this session
- **THEN** the summary's reward card shows "New: 1,000 answers", and studying continued without a dialog

#### Scenario: Several earned
- **WHEN** the session earned Answers 1,000 and Study days 100
- **THEN** the reward card names both

#### Scenario: Level tier with the level line
- **WHEN** the session raised the level to 10 and earned the Level 10 tier and Answers 1,000
- **THEN** the reward card says "Reached level 10" and "New: 1,000 answers"

#### Scenario: Already shown by a banner
- **WHEN** a banner showed "New: 1,000 answers" earlier in the session
- **THEN** the summary's reward card still names it
