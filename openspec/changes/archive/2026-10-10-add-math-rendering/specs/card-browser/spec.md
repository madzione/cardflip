# Spec Delta

## ADDED Requirements

### Requirement: Insert math in the editor
The note editor SHALL offer a math menu that wraps the selected text, or inserts an empty pair at the cursor, with inline math `\(…\)`, display math `\[…\]`, chemistry `\(\ce{…}\)`, `[latex]…[/latex]`, `[$]…[/$]` or `[$$]…[/$$]`. On platforms with a keyboard the same actions SHALL be available as Ctrl+M M, Ctrl+M E, Ctrl+M C, Ctrl+T T, Ctrl+T E and Ctrl+T M (Cmd on macOS).

#### Scenario: Wrap selection
- **WHEN** the user selects `x^2` in a field and chooses inline math
- **THEN** the field contains `\(x^2\)`

#### Scenario: Empty insert
- **WHEN** nothing is selected and the user presses Ctrl+M then E
- **THEN** `\[\]` is inserted and the cursor is between the delimiters
