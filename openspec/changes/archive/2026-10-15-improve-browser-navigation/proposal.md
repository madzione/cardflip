# Proposal

## Why

Selecting and moving around in Cardflip's browser is basic: select all and clear only, a fixed split between list and details, and no keyboard navigation. Anki adds invert selection, "select notes" (extend the selection to all cards of the selected notes), resizable panes and keyboard shortcuts for moving through the list, which matter when working through large result sets.

## What Changes

- **Invert selection** (Ctrl+Alt+S / Cmd+Option+S) and **Select notes** (selects every card of the notes of the current selection) in the selection menu.
- On wide screens, the divider between the list and the details can be dragged; the position is remembered.
- Keyboard navigation in the list: Up/Down move the current row (details follow), Shift+Up/Down extend the selection, Home/End jump to the first and last row, Ctrl+F (Cmd+F) focuses the search, Ctrl+Shift+F focuses the list, Enter opens the editor of the current card, Delete deletes after confirmation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for selection commands, the resizable split and keyboard navigation.

## Impact

- Go: `BrowseService.CardsOfNotes(cardIDs)` for Select notes; `core.Settings` gains `BrowserSplit`.
- Flutter: `browse_screen.dart` (selection menu, draggable divider, `Shortcuts`/`Actions` for navigation).
- Tests: `golang/svc_test.go`, `test/browse_test.dart`.
