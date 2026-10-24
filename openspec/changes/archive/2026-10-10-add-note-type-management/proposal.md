# Proposal

## Why

In Cardflip, the only note types are the three built-ins and those that arrive with imports. A user cannot create a note type for vocabulary with Word/Meaning/Example fields, cannot clone an imported type to experiment safely, and cannot delete a type they no longer need. Anki's Manage Note Types screen does all three, and this blocks most custom card designs.

## What Changes

- The note type list gets **Add**, **Clone** and **Delete**.
- **Add** creates a note type from a stock type, under a name the user chooses: Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer), and Cloze.
- **Clone** copies an existing note type (fields, card types, CSS, kind) under a new name, without its notes.
- **Delete** removes a note type with all its notes and their cards after a confirmation that states how many notes and cards will be deleted. The last remaining note type cannot be deleted.
- "Basic (optional reversed card)" has fields Front, Back and Add Reverse; the reverse card is created only when Add Reverse is not empty. "Basic (type in the answer)" has a front template `{{Front}}\n\n{{type:Back}}`.
- Note type names must be unique (case-insensitive); the clone dialog proposes a free name such as "Basic copy" or "Basic copy 2".

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for creating, cloning and deleting note types and for the stock note types.

## Impact

- Go: `golang/core/notetypes.go` gains `AddNoteType(stock, name)`, `CloneNoteType(id, name)`, `DeleteNoteType(id)` and `NoteTypeDeletionInfo(id)`, plus stock definitions next to `builtinNoteTypes()`; `NotesService` gains matching calls and a `StockTypes()` list.
- Flutter: `lib/features/editor/note_type_manager.dart` (`NoteTypeListScreen` actions and dialogs). The editor's remembered note type falls back to the first note type when the remembered one is deleted.
- Tests: `golang/core/notes_test.go` (or a new `notetypes_test.go`), `golang/svc_test.go`, `test/editor_test.dart`, `test/support/fake_go.dart`.
- Behaviour: built-in names stay unique, so adding a second "Basic" requires a different name. Fresh collections keep the three current built-ins.
