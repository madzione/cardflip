# Proposal

## Why

With many decks, keyboard users switch decks in Anki by pressing `/`, typing part of a name and pressing Enter. In Cardflip they must scroll the deck tree with the mouse, expanding parents to find a subdeck.

## What Changes

- Pressing `/` anywhere outside a text field (deck list, study screen, browser) opens a **Study deck** picker: a search field over full deck names (`Languages::Spanish`), matching all typed words in any order, case-insensitively.
- Enter or a tap starts studying the chosen deck (through the deck's overview when that screen exists); the arrow keys move the selection.
- The picker lists the most recently studied decks first when the search is empty.
- A **Study deck** entry in the deck list's toolbar opens the same picker on touch devices.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-shell`: new requirement for the deck picker.

## Impact

- Go: `core.Settings` gains `RecentDecks []ID` (last 10 studied, updated when a session starts).
- Flutter: new `lib/features/decks/deck_picker.dart`, a global `/` shortcut in `lib/app/scaffold.dart`, a toolbar button in `deck_list_screen.dart`.
- Tests: `golang/svc_test.go`, `test/app_shell_test.dart`, `test/deck_list_test.dart`.
