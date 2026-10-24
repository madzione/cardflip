# Spec Delta

## ADDED Requirements

### Requirement: OR and grouping
The search SHALL support `or` between terms (case-insensitive), an optional explicit `and`, parentheses for grouping, and `-` before a group to negate it. AND, explicit or implicit, SHALL bind more tightly than OR.

#### Scenario: Either deck
- **WHEN** the user searches `(deck:Spanish or deck:French) is:due`
- **THEN** due cards from either deck are listed

#### Scenario: Negated group
- **WHEN** the user searches `-(tag:verb or tag:noun)`
- **THEN** cards whose notes have neither tag are listed

#### Scenario: Precedence
- **WHEN** the user searches `deck:A tag:x or tag:y`
- **THEN** cards in deck A with tag x, and cards anywhere with tag y, are listed

### Requirement: Escaping in searches
In search text, `\"`, `\(`, `\)`, `\-`, `\:` and `\\` SHALL stand for the literal characters. Quoted text SHALL be one term, with spaces and parentheses taken literally.

#### Scenario: Literal parenthesis
- **WHEN** a note contains `f(x)` and the user searches `"f(x)"`
- **THEN** the note's cards are listed

#### Scenario: Leading hyphen
- **WHEN** a note contains `-ing` and the user searches `\-ing`
- **THEN** the note's cards are listed, and the term is not treated as negation

### Requirement: Search syntax errors
A search with unbalanced parentheses, an `or` missing a term on either side, or an empty group SHALL show an error naming the problem, instead of a result list.

#### Scenario: Unclosed group
- **WHEN** the user searches `(deck:Spanish or deck:French`
- **THEN** the browser shows an error that a parenthesis is not closed
