# Spec Delta

## ADDED Requirements

### Requirement: Due day choice with Easy Days
When a review interval allows a range of due days through fuzz, the due day SHALL be chosen within that range so that Reduced weekdays receive fewer reviews than Normal days and Minimum weekdays receive reviews only when the range contains no other day, taking into account the reviews already due on each day. The interval previews SHALL show the chosen day.

#### Scenario: Avoid a minimum day
- **WHEN** Sunday is Minimum and a card's interval range spans Friday to Monday
- **THEN** the card is not scheduled on Sunday

#### Scenario: Only a minimum day
- **WHEN** Sunday is Minimum and a card's interval range is only Sunday
- **THEN** the card is scheduled on Sunday

#### Scenario: No Easy Days set
- **WHEN** every weekday is Normal
- **THEN** due days are chosen as before
