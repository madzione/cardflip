# Spec Delta

## ADDED Requirements

### Requirement: Unicode normalisation
Field contents, tags, deck names and descriptions, note type names, field names, card templates, note type CSS and media file names SHALL be stored in Unicode NFC form, whether they come from editing, renaming, adding media or importing. Search queries SHALL be normalised the same way. A collection created by an earlier version SHALL be normalised once when it is first opened; stored media files whose names are not in NFC SHALL be renamed to their NFC form, so that references in fields, templates and CSS keep finding them.

#### Scenario: Decomposed input
- **WHEN** the user adds a note whose Front is `e` followed by a combining acute accent
- **THEN** the stored Front is the single character `é`

#### Scenario: Search matches either form
- **WHEN** a note contains `café` and the user searches `café` typed in decomposed form
- **THEN** the note is found

#### Scenario: Decomposed media name
- **WHEN** a collection from an earlier version holds a media file named `café.jpg` in decomposed form, and a note shows it with `<img src="café.jpg">` written the same way
- **THEN** after the collection is first opened, the file is stored as `café.jpg` in NFC, the note's reference is in NFC, and the image still shows
