# Proposal

## Why

Writing cloze notes quickly depends on keyboard shortcuts. Anki has Ctrl+Shift+C for the next cloze number and Ctrl+Alt+Shift+C to reuse the current number (to hide several parts on the same card). Cardflip has only a toolbar button that inserts the next number, with no shortcut and no same-number variant.

## What Changes

- Ctrl+Shift+C (Cmd+Shift+C on macOS) wraps the selection in the next cloze number, like the existing button.
- Ctrl+Alt+Shift+C (Cmd+Option+Shift+C on macOS) wraps the selection using the highest number already in the note, so both parts are hidden on the same card. With no cloze yet, it uses 1.
- The toolbar button gets a menu for the same-number variant, for touch devices.
- With an empty selection, an empty deletion is inserted with the cursor inside.
- The numbers consider all fields of the note, not only the focused field.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirement for cloze insertion shortcuts.

## Impact

- Flutter: `_insertCloze` in `lib/features/editor/note_editor_screen.dart` (number computed across all fields, same-number variant), shortcuts and the button menu.
- Tests: `test/editor_test.dart`.
- No Go change: the highest number is computed in Dart with the same pattern the Go parser uses (`{{c<digits>`, including multi-number lists such as `c1,2`).
