# Spec Delta

## ADDED Requirements

### Requirement: Pre-generated LaTeX images
`[latex]…[/latex]`, `[$]…[/$]` and `[$$]…[/$$]` SHALL be replaced by the media image Anki would have generated for that source, found by the same file name Anki uses (PNG, or SVG when the note type is set to SVG). When that image is not in the collection's media, the LaTeX source SHALL be shown as preformatted text. SVG images SHALL display on every platform.

#### Scenario: Image shipped with the deck
- **WHEN** a card from an imported deck contains `[$]x^2[/$]` and the deck's media includes the image Anki generated for it
- **THEN** the card shows that image

#### Scenario: Image missing
- **WHEN** a card contains `[latex]\LaTeX[/latex]` and no matching image exists
- **THEN** the card shows `\LaTeX` as preformatted text and no broken image
