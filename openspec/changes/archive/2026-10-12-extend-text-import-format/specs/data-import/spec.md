# Spec Delta

## ADDED Requirements

### Requirement: Comment lines
In delimited text files, lines that start with `#` after the header lines SHALL be skipped, and the import summary SHALL count them as comments.

#### Scenario: Comment in the middle
- **WHEN** a file has the rows `a,b`, `# chapter 2`, `c,d`
- **THEN** two notes are imported and the summary reports 1 comment line

### Requirement: HTML in text imports
The import screen SHALL offer "Allow HTML in fields", on by default or as set by a `#html:` header. When off, field text SHALL be imported literally, so that `<`, `>` and `&` display as typed and line breaks are kept.

#### Scenario: Literal text
- **WHEN** "Allow HTML in fields" is off and a field is `a < b & c`
- **THEN** the card shows `a < b & c`

### Requirement: Tags header
A `#tags:` header SHALL add its space-separated tags to every note imported from the file.

#### Scenario: Header tags
- **WHEN** a file starts with `#tags:chapter1 vocab`
- **THEN** every imported note has the tags `chapter1` and `vocab`

### Requirement: Separator choice
The `#separator:` header SHALL accept Comma, Semicolon, Tab, Space, Pipe and Colon, by name or as the character. The import screen SHALL show the separator in use and let the user choose another one, updating the preview.

#### Scenario: Override detection
- **WHEN** a file is detected as comma-separated but uses `|` and the user chooses Pipe
- **THEN** the preview shows the columns split at `|`
