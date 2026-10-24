# Spec Delta

## ADDED Requirements

### Requirement: Card root classes
The rendered card SHALL carry the classes `card` and `cardN`, where N is the card type number starting at 1 (the cloze number for cloze cards), `nightMode` and `night_mode` while the app shows the dark theme, and the class or classes of the current platform: `win`, `mac`, `linux`, `android`, `iphone`, `ipad`, and `mobile` on phones and tablets. CSS rules selecting these classes, on their own or followed by a descendant selector, SHALL apply.

#### Scenario: Reverse card styling
- **WHEN** the CSS contains `.card2 { color: red }` and the reverse card of a `Basic (and reversed card)` note is shown
- **THEN** the reverse card's text is red and the forward card's is not

#### Scenario: Night mode rule
- **WHEN** the CSS contains `.nightMode .cloze { color: lightblue }` and the app is in dark mode
- **THEN** cloze highlights are light blue, and in light mode they use the other rules

#### Scenario: Phone rule
- **WHEN** the CSS contains `.mobile .card { font-size: 30px }` and the card is shown on an Android phone
- **THEN** the card text is 30 px

### Requirement: Environment media queries
`@media` rules SHALL apply when their conditions match: `prefers-color-scheme` SHALL follow the app's current theme, and `min-width`, `max-width` and `orientation` SHALL be evaluated against the card's display area. Unsupported conditions SHALL make the rule not apply. When the theme or display width changes, the shown card SHALL update to the matching rules.

#### Scenario: Dark scheme query
- **WHEN** the CSS contains `@media (prefers-color-scheme: dark) { .card { background: black } }` and the user switches the app to dark mode while a card is shown
- **THEN** the card background turns black without leaving the card
