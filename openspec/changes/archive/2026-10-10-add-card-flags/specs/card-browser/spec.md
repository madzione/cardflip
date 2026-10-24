# Spec Delta

## ADDED Requirements

### Requirement: Flag bulk action
The user SHALL be able to set one of the seven flags on the selected cards, or clear their flags, from the browser's bulk actions and from the card details. When every selected card already has the chosen flag, choosing it SHALL clear it.

#### Scenario: Flag selection
- **WHEN** the user selects 5 cards and chooses flag 4
- **THEN** all 5 cards have flag 4

#### Scenario: Clear flags
- **WHEN** the user selects flagged cards and chooses "No flag"
- **THEN** none of them has a flag

### Requirement: Flag search
The search SHALL support `flag:N` for N from 0 to 7, where `flag:0` matches cards without a flag. Other values SHALL be reported as an invalid term.

#### Scenario: Search red flags
- **WHEN** the user searches `flag:1`
- **THEN** only cards with the red flag are listed

#### Scenario: Unflagged cards
- **WHEN** the user searches `deck:Spanish flag:0`
- **THEN** only cards in Spanish without a flag are listed

### Requirement: Row colours
Each browser row SHALL show the card's flag as a marker in the flag's colour, and SHALL be tinted with that colour. A suspended card without a flag SHALL be tinted yellow. Tints SHALL keep the row text readable in light and dark themes.

#### Scenario: Suspended row
- **WHEN** a suspended card without a flag is listed
- **THEN** its row has a yellow tint and still says "suspended"

#### Scenario: Flag wins over suspended
- **WHEN** a suspended card has flag 2
- **THEN** its row is tinted orange
