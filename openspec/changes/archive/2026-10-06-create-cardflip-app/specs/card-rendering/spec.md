# Spec Delta

## Purpose

Turns a note and one of its note type's templates into the question and answer the user sees, using a subset of Anki's template language, and displays the result as styled HTML with images.

## ADDED Requirements

### Requirement: Field substitution
A template SHALL replace `{{FieldName}}` with that field's content. Field names SHALL match exactly, except that surrounding whitespace inside the braces SHALL be ignored. A reference to an unknown field SHALL render an inline error marker naming the field, instead of failing the card.

#### Scenario: Basic substitution
- **WHEN** the question template is `{{Front}}` and the note's Front is `hola`
- **THEN** the question renders as `hola`

#### Scenario: Unknown field
- **WHEN** a template references `{{Missing}}`
- **THEN** the card renders, with a visible marker saying field `Missing` was not found

### Requirement: FrontSide
In an answer template, `{{FrontSide}}` SHALL be replaced by the rendered question, excluding any audio.

#### Scenario: Answer includes front
- **WHEN** the answer template is `{{FrontSide}}<hr id=answer>{{Back}}`
- **THEN** the answer shows the question, a divider, and then the Back field

### Requirement: Conditional sections
`{{#Field}}…{{/Field}}` SHALL render its content only when the field is non-empty after stripping HTML and whitespace. `{{^Field}}…{{/Field}}` SHALL render its content only when the field is empty.

#### Scenario: Optional hint omitted
- **WHEN** the template contains `{{#Hint}}Hint: {{Hint}}{{/Hint}}` and the Hint field is empty
- **THEN** no "Hint:" text appears

### Requirement: Field filters
The renderer SHALL support the filters `text:` (strip HTML), `hint:` (collapsed content revealed by tapping), `type:` (rendered as the plain field content, without a type-in box), and `cloze:`. Unknown filters SHALL be ignored, and the field content SHALL render unfiltered.

#### Scenario: Text filter
- **WHEN** the template has `{{text:Front}}` and Front is `<b>bold</b>`
- **THEN** the output is `bold` without markup

### Requirement: Cloze deletions
For cloze card number N, `{{cloze:Field}}` SHALL show cloze N as `[...]` (or `[hint]` when written `{{cN::answer::hint}}`) on the question side, and highlight its answer on the answer side. Other cloze numbers SHALL show their answer text unhighlighted on both sides.

#### Scenario: Cloze question and answer
- **WHEN** card 1 of `{{c1::Paris}} is in {{c2::France}}` is rendered
- **THEN** the question shows `[...] is in France` and the answer shows `Paris is in France` with `Paris` highlighted

#### Scenario: Cloze hint
- **WHEN** card 1 of `{{c1::Paris::city}} is in France` is rendered
- **THEN** the question shows `[city] is in France`

### Requirement: HTML display with styling
Rendered cards SHALL be displayed as HTML, supporting at least: text formatting, lists, tables, line breaks, links, colours, and the note type's CSS for font size, colour, alignment and background. Scripts SHALL NOT be executed.

#### Scenario: Script ignored
- **WHEN** a field contains `<script>alert(1)</script>Text`
- **THEN** only `Text` is displayed and no script runs

### Requirement: Images
`<img src="name">` referencing a file in the collection's media SHALL display that image, scaled to fit the card width. A missing media file SHALL show a placeholder instead of failing.

#### Scenario: Imported image shown
- **WHEN** a card from an imported `.apkg` references `cat.jpg`, which was imported with the deck
- **THEN** the image is displayed on the card

### Requirement: Unsupported media degrades gracefully
Audio and video references (`[sound:…]` tags, `<audio>`, `<video>`) SHALL be omitted from display, without errors.

#### Scenario: Sound tag
- **WHEN** a field contains `[sound:word.mp3] palabra`
- **THEN** the card shows `palabra` and nothing else from the sound tag

### Requirement: Theme-aware card display
Cards SHALL be readable in both light and dark app themes. In dark mode, the default card background and text colour SHALL follow the theme unless the note type's CSS sets them explicitly.

#### Scenario: Dark mode default
- **WHEN** the app is in dark mode and the note type CSS sets no colours
- **THEN** the card shows light text on a dark background
