# Spec Delta

## ADDED Requirements

### Requirement: LaTeX image format on Anki import
Importing an `.apkg` SHALL keep each note type's choice between PNG and SVG LaTeX images, and SHALL import SVG media files.

#### Scenario: SVG deck
- **WHEN** a package whose note type uses SVG LaTeX images is imported
- **THEN** its `[latex]` blocks show the imported SVG images
