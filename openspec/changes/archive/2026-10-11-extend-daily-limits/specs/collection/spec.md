# Spec Delta

## ADDED Requirements

### Requirement: Today-only limits
A deck's new and review limits SHALL be overridable for the current study day only. The override SHALL be ignored from the next rollover on, and SHALL be included in backups.

#### Scenario: Override expires
- **WHEN** the new card limit of `Spanish` is 20 with a today-only override of 30, and the next study day starts
- **THEN** the effective new card limit is 20

### Requirement: Limit behaviour options
Each deck SHALL have the options new cards ignore review limit (on by default) and limits start from top (off by default).

#### Scenario: Defaults
- **WHEN** a deck is created
- **THEN** new cards ignore the review limit and limits do not start from the top
