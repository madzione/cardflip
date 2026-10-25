# Spec Delta

## MODIFIED Requirements

### Requirement: New achievements on the summary
When achievement tiers were earned during the session, the session-complete screen SHALL show one line naming them, after the streak and goal lines. A Level tier SHALL be left out of that line when the level line already says the level was reached. The session SHALL NOT be interrupted when a tier is earned.

#### Scenario: Earned during the session
- **WHEN** the user's 1,000th answer is in this session
- **THEN** the summary shows "New: 1,000 answers" and no message appeared while studying

#### Scenario: Several earned
- **WHEN** the session earned Answers 1,000 and Study days 100
- **THEN** the summary shows one line naming both

#### Scenario: Level tier with the level line
- **WHEN** the session raised the level to 10 and earned the Level 10 tier and Answers 1,000
- **THEN** the summary says "Reached level 10" and "New: 1,000 answers"
