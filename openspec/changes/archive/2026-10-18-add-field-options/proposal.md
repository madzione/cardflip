# Proposal

## Why

Anki's Fields dialog lets users reorder fields and set per-field options that matter in daily use: an editing font and size (essential for CJK and Arabic scripts), right-to-left editing, collapsing long fields, which field the browser sorts by, and excluding a field (such as a long "Notes" field) from plain text searches. Cardflip can add, rename and remove fields, but not reorder them, and has none of these options.

## What Changes

- Fields can be moved to another position. Notes keep their content in the right field, and templates are unaffected (they refer to fields by name).
- Per-field options:
  - **Editing font and size**: used for that field in the editor only, not on cards.
  - **Right-to-left**: the editor lays out and aligns that field right to left.
  - **Collapse by default**: the field starts collapsed in the editor and expands on tap.
  - **Sort by this field**: one field per note type is the sort field (the first field by default); the browser's question-text sort is joined by a "Sort field" sort order that uses it.
  - **Exclude from unqualified searches**: free text in the browser search does not look in that field (field-qualified searches still can).
- Cards render `dir="rtl"` from templates and fields right to left.
- `.apkg` import keeps the field options Anki stores (font, size, RTL, sort field, collapse, exclude from search) and the field order.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for reordering fields and for field options.
- `card-browser`: new requirements for the sort field sort order, unqualified search exclusion and the editor's per-field presentation.
- `card-rendering`: new requirement for right-to-left text.

## Impact

- Go: `core.NoteType` gains per-field options (`FieldOptions` slice parallel to `Fields`, `omitempty`) and `SortField int`; `Collection.MoveField(typeID, from, to)` reorders fields and note values in one op; `core/search.go` builds search text without excluded fields and adds a `sortField` order; `importer/apkg` maps Anki's `font`, `size`, `rtl`, `sortf`, `collapsed`, `excludeFromSearch`.
- Flutter: `note_type_manager.dart` (field reorder and options sheet), `note_editor_screen.dart` (font, size, direction, collapse), `browse_screen.dart` (sort menu).
- Tests: `golang/core/notes_test.go`, `golang/core/search_test.go`, `golang/importer/apkg/apkg_test.go`, `test/editor_test.dart`, `test/browse_test.dart`, `test/card_view_test.dart`.
- Overlap: `add-rich-text-editor` and `improve-rapid-entry` add their own per-field options to the same list; the change landing second extends the struct instead of adding another.
