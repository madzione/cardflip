# Spec Delta

## ADDED Requirements

### Requirement: Visual field editor
Note fields in the add and edit screens SHALL be edited visually with a toolbar for bold, italic, underline, subscript, superscript, text colour, highlight colour, clear formatting, bulleted and numbered lists, alignment and indentation. On platforms with a keyboard, Ctrl+B, Ctrl+I and Ctrl+U (Cmd on macOS) SHALL toggle bold, italic and underline. The saved field SHALL be HTML that renders the same formatting on cards.

#### Scenario: Make a word bold
- **WHEN** the user selects `perro` in the Front field, presses Ctrl+B and saves
- **THEN** the Front field contains `<b>perro</b>` and the card shows the word in bold

### Requirement: HTML source toggle
Each field SHALL have a toggle between the visual editor and its HTML source, also available as Ctrl+Shift+X (Cmd+Shift+X on macOS). A per-field note type option SHALL open that field in HTML mode by default.

#### Scenario: Edit source
- **WHEN** the user toggles the Back field to HTML, types `<i>x</i>` and toggles back
- **THEN** the visual editor shows an italic `x`

### Requirement: No silent loss of formatting
A field containing HTML that the visual editor cannot represent without loss SHALL open in HTML mode with a note saying so. Switching it to the visual editor SHALL require confirmation.

#### Scenario: Imported table
- **WHEN** the user edits an imported note whose Back field contains a `<table>`
- **THEN** the Back field opens in HTML mode with a note, and saving without changes keeps the table exactly

### Requirement: Paste with formatting
Pasting into a field SHALL keep bold, italic, underline, colours, lists and line breaks, and drop other formatting. Ctrl+Shift+V (Cmd+Shift+V on macOS) SHALL paste plain text.

#### Scenario: Plain paste
- **WHEN** the user copies bold text from a web page and pastes it with Ctrl+Shift+V
- **THEN** the field contains the text without bold

### Requirement: Resize images in the editor
Images in a field SHALL be shown in the visual editor, and the user SHALL be able to resize one by dragging a handle, which sets its width while keeping its aspect ratio.

#### Scenario: Shrink an image
- **WHEN** the user drags an image's handle to half its width and saves
- **THEN** the image is shown at that width on the card
