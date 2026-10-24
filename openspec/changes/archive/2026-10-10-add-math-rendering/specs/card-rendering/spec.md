# Spec Delta

## ADDED Requirements

### Requirement: MathJax typesetting
Text between `\(` and `\)` SHALL be typeset as inline math, and text between `\[` and `\]` as display math on its own centred line. Typesetting SHALL work offline on every supported platform. HTML tags and entities inside the delimiters SHALL be ignored as markup.

#### Scenario: Inline formula
- **WHEN** a field contains `Area is \(\pi r^2\)`
- **THEN** the card shows "Area is" followed by a typeset π r² in the line

#### Scenario: Display formula
- **WHEN** a field contains `\[\frac{a}{b}\]`
- **THEN** the card shows a typeset fraction centred on its own line

### Requirement: Chemistry formulas
Inside math, `\ce{…}` SHALL be typeset as a chemical formula, supporting element symbols with subscript counts, charges, stoichiometric numbers, states such as `(aq)`, `+` between species and reaction arrows.

#### Scenario: Reaction
- **WHEN** a field contains `\(\ce{2H2 + O2 -> 2H2O}\)`
- **THEN** the card shows the reaction with subscript 2s, upright element symbols and a right arrow

### Requirement: Math errors and clozes
A formula that cannot be typeset SHALL be shown as its TeX source with an error style, and the rest of the card SHALL render. A cloze deletion inside math SHALL show as `[...]` within the formula on the question side and highlighted on the answer side.

#### Scenario: Unknown command
- **WHEN** a field contains `\(\unknowncommand{x}\)` and `text`
- **THEN** the card shows `\unknowncommand{x}` with an error style and `text` normally

#### Scenario: Cloze in a formula
- **WHEN** card 1 of `\(E = {{c1::mc^2}}\)` is rendered
- **THEN** the question shows `E = [...]` typeset and the answer shows `E = mc²` with `mc²` highlighted
