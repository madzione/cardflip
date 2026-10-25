# Spec Delta

## ADDED Requirements

### Requirement: New achievements on the summary
When achievement tiers were earned during the session, the session-complete screen SHALL show one line naming them, after the streak and goal lines. The session SHALL NOT be interrupted when a tier is earned.

#### Scenario: Earned during the session
- **WHEN** the user's 1,000th answer is in this session
- **THEN** the summary shows "New: 1,000 answers" and no message appeared while studying

#### Scenario: Several earned
- **WHEN** the session earned Answers 1,000 and Study days 100
- **THEN** the summary shows one line naming both
