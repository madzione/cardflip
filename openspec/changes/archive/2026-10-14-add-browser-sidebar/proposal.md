# Proposal

## Why

Finding cards in Cardflip means typing search syntax. Anki's browser has a sidebar tree of decks, note types, tags and card states: one click searches, and modifier clicks combine searches. The same sidebar is where Anki users rename, delete and reorganise tags and decks. Cardflip cannot rename, delete or reparent tags anywhere, and decks only from the deck list.

## What Changes

- A sidebar in the browser (always visible on wide screens, a drawer on narrow ones) with sections: Card states (New, Learning, Review, Due, Suspended), Decks (tree), Note types, and Tags (tree by `::`).
- Clicking an item replaces the search with that item's search (`deck:"…"`, `tag:…`, `note:"…"`, `is:…`). Ctrl-click (Cmd on macOS) adds it with AND; Alt-click (Option) adds it negated. Long-press on touch offers the same choices.
- Tags in the sidebar can be renamed (including all child tags), deleted (from all notes, after a confirmation with the note count), and moved under another tag by drag-and-drop; several can be selected.
- **Clear unused tags** removes tags that no note uses from the tag list.
- Decks in the sidebar can be renamed, deleted and moved under another deck by drag-and-drop, with the same rules and confirmations as on the deck list.
- A filter field at the top of the sidebar narrows the tree.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for the sidebar and clicking to search.
- `collection`: new requirements for renaming, deleting and reparenting tags, and clearing unused tags.

## Impact

- Go: `Collection.RenameTag(old, new)` (with children, merging into existing tags), `ReparentTags(tags, parent)` and `DeleteTags(tags)` (several selected tags in one operation, so one undo step), `TagUsage(tags)` (notes counted once), `ClearUnusedTags()` (the tag list is derived from notes today, so this removes remembered tags such as those kept for the tag picker; if there are none, it reports zero); `NotesService` gains `RenameTag`, `MoveTags`, `DeleteTags`, `TagUsage` and `ClearUnusedTags`.
- Flutter: new `lib/features/browse/browse_sidebar.dart`, `browse_screen.dart` (layout and search composition), drag-and-drop with `Draggable`/`DragTarget`.
- Tests: `golang/core/notes_test.go`, `golang/svc_test.go`, `test/browse_test.dart`.
- Overlap: `add-boolean-search` adds OR; once it exists, Shift-click adds an item with OR. `add-saved-searches` and `add-card-flags` can add their own sections. `add-recent-event-search` can add a Today section.
