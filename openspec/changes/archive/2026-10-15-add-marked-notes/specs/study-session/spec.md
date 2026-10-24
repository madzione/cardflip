# Spec Delta

## ADDED Requirements

### Requirement: Mark during study
While studying, `*` and a menu item SHALL toggle the mark on the current card's note, and a marked card SHALL show a star.

#### Scenario: Toggle with key
- **WHEN** the user presses `*` on an unmarked card, and again
- **THEN** a star appears and the note has `marked`, and then the star and tag are gone
