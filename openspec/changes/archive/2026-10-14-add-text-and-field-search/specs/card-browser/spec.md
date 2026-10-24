# Spec Delta

## ADDED Requirements

### Requirement: Wildcards and words
In free-text terms, `_` SHALL match any single character and `*` any run of characters, while `\_` and `\*` SHALL match the literal characters. `w:` SHALL match whole words, where a `*` at either end allows a prefix or suffix.

#### Scenario: Single-character wildcard
- **WHEN** notes contain `dog` and `dig` and the user searches `d_g`
- **THEN** both notes' cards are listed

#### Scenario: Whole word
- **WHEN** notes contain `dog` and `doggy` and the user searches `w:dog`
- **THEN** only the note with `dog` is listed

### Requirement: Field search
`<field>:<text>` SHALL match notes whose field of that name, with HTML stripped and case ignored, equals the text, where wildcards allow partial matches and the field name may contain `*`. `<field>:` SHALL match an empty field, `<field>:_*` a non-empty field, and `<field>:*` any note that has the field.

#### Scenario: Exact field
- **WHEN** a note's Front is `dog` and another's is `a dog`, and the user searches `front:dog`
- **THEN** only the first note is listed

#### Scenario: Empty field
- **WHEN** the user searches `back:`
- **THEN** notes whose Back field is empty are listed

### Requirement: Accent and cloze handling
`nc:` SHALL match text while ignoring accents and other combining marks, and `sc:` SHALL match text with cloze markup removed. With the setting "Ignore accents in search" on, all text and field terms SHALL ignore accents.

#### Scenario: Ignore accents
- **WHEN** a note contains `über` and the user searches `nc:uber`
- **THEN** the note is listed

#### Scenario: Across a cloze
- **WHEN** a note's Text is `{{c1::capital}} of France` and the user searches `"sc:capital of France"`
- **THEN** the note is listed

### Requirement: Regular expression search
`re:<expr>` SHALL match notes whose raw field content matches the expression, `<field>:re:<expr>` only that field, and `tag:re:<expr>` any tag. Matching SHALL ignore case unless the expression starts with `(?-i)`. An invalid expression SHALL show an error.

#### Scenario: Field regex
- **WHEN** the user searches `front:re:^\d+$`
- **THEN** notes whose Front consists only of digits are listed

### Requirement: Entity decoding
Text and field terms SHALL match field text with HTML entities decoded.

#### Scenario: Ampersand
- **WHEN** a field is stored as `salt &amp; pepper` and the user searches `"salt & pepper"`
- **THEN** the note is listed
