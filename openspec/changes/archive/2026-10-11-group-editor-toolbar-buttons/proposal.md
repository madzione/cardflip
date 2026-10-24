# Proposal

## Why

The buttons that change a field's content sit where there happened to be room. In the add/edit screen, Cloze, Math, Attach and Record audio are app bar actions next to the screen-wide Note types button, away from the fields they insert into. In the browser's edit pane, Math, Attach and Record audio are in the row with Save, and there is no Cloze button at all. The two editors place them differently, and in neither are they in the field toolbar, which already acts on the field being edited. On a phone they also crowd the app bar to up to five icons.

## What Changes

- Cloze deletion (cloze note types only), Math, Attach and Record audio move into the field toolbar (the bold, italic, colours, lists, … row above the fields). They form one group in that order at its end, set apart from the formatting buttons.
- Both editors get the same group: the add/edit screen and the browser's edit pane. The browser pane gains the Cloze button for cloze notes, which it does not have today.
- They are removed from the app bar of the add/edit screen and from the row with Save in the browser's edit pane.
- Unlike the formatting buttons, they stay enabled while a field shows its HTML source, because they insert text or markup, which works there too.
- They act on the focused field, or the last focused one, as the rest of the toolbar does.
- What stays where it is: screen-wide actions in the app bar (Note types, and such buttons from other changes, e.g. the History button of `add-recent-event-search`), Save, each field's HTML toggle, and the card actions in the browser pane's header (Edit, Flag, Suspend, Delete).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirement for where the editing buttons are in the editor.

## Impact

- Flutter: `RichToolbar` in `lib/features/editor/rich_field.dart` (room for trailing buttons that stay enabled in HTML mode). The shared mixin in `lib/features/editor/field_media.dart` grows into one for all editing buttons and takes over the cloze insertion from `note_editor_screen.dart`. `note_editor_screen.dart` and `lib/features/browse/card_details.dart` pass the group to the toolbar instead of the app bar or the Save row. `MathMenuButton` in `math_insert.dart` is reused as is.
- Tests: `test/editor_test.dart` and `test/browse_test.dart` (placement, cloze in the browser, HTML mode). The existing math, cloze and media tests find the buttons by key or tooltip and keep working.
- No Go change.
- Overlap: requires `add-editor-media`, whose requirements for attaching and recording are not yet in the main `card-browser` spec, so land this after that change is archived. `improve-cloze-editing` adds a same-number menu and shortcuts to the Cloze button and edits `_insertCloze` in `note_editor_screen.dart`, which this change moves into the shared mixin; whichever lands second applies its part to the other's code, and the menu then lives on the toolbar button in both editors.
