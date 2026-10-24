# Proposal

## Why

Entering many cards in a row is the main use of the add screen. Anki keeps chosen fields across adds (sticky fields, for example a shared "Source" or "Chapter" field), adds with Ctrl+Enter without reaching for the mouse, and can remember the last deck per note type. Cardflip clears every field, has no add shortcut, and remembers only one deck for all note types.

## What Changes

- Each field in the add screen has a pin. Pinned fields keep their content after **Add**; the pins are remembered per note type.
- Ctrl+Enter (Cmd+Enter on macOS) adds the card from any field; in the edit screen it saves.
- A setting "Deck for new cards": *Current deck* (today's behaviour: one remembered deck) or *Change deck depending on note type* (the add screen switches to the deck last used with the chosen note type).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: the "Add cards" requirement changes so pinned fields are kept after an add; new requirements for the add shortcut and the deck-per-note-type setting.

## Impact

- Go: `core.NoteType` gains sticky flags per field (stored with the note type, so they are in backups); `core.Settings` gains the deck mode and a `LastDeckByNoteType map[ID]ID`; `NotesService.Defaults(typeID)` returns the deck for a note type in that mode.
- Flutter: `note_editor_screen.dart` (pins, shortcut, deck switching), `settings_screen.dart`.
- Tests: `golang/svc_test.go`, `test/editor_test.dart`, `test/settings_test.dart`.
- Overlap: `add-rich-text-editor` and `add-field-options` also add per-field data to note types; all use the same per-field options list, which the change landing second extends.
