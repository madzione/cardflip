# Spec Delta

## ADDED Requirements

### Requirement: Auto advance deck options
Each deck SHALL have the options seconds to show question for and seconds to show answer for (0 to 9999, default 0 meaning disabled), the question action (show answer, or show reminder; default show answer) and the answer action (bury card, answer Again, answer Hard, answer Good, or show reminder; default bury card).

#### Scenario: Options stored
- **WHEN** the user sets 5 seconds for the question and 3 for the answer with the answer action "answer Good"
- **THEN** those values persist for the deck and are included in backups
