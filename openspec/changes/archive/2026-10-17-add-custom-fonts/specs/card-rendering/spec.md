# Spec Delta

## ADDED Requirements

### Requirement: Fonts from media
An `@font-face` rule in the note type's CSS whose `src` names a `.ttf` or `.otf` file in the collection's media SHALL make that font available under the rule's `font-family` name, and declarations using that name SHALL display text in that font. Fonts in other formats, remote URLs and missing files SHALL be ignored, and text SHALL use the next listed family or the default font.

#### Scenario: Deck font applied
- **WHEN** the CSS contains `@font-face { font-family: jp; src: url("_jp.ttf") }` and `.card { font-family: jp }`, and `_jp.ttf` is in the media
- **THEN** the card text is displayed in that font

#### Scenario: Missing font file
- **WHEN** the CSS names `_missing.ttf` in `@font-face` and the file is not in the media
- **THEN** the card is shown in the default font without an error

#### Scenario: Unsupported format
- **WHEN** the `src` is `url("_font.woff2")`
- **THEN** the rule is ignored and the default font is used
