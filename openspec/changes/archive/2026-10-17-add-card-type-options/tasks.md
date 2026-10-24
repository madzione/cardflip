# Tasks

## 1. Collection

- [x] 1.1 Add `DeckOverride ID`, `BrowserQFmt` and `BrowserAFmt` to `core.CardTemplate` and the service DTO (all `omitempty`). Add a snapshot round-trip test. Verify `go test ./core/...` passes.
- [x] 1.2 Add `Collection.MoveTemplate(typeID, from, to)` that reorders templates and remaps `Card.Ord` for all notes of the type in one op. Add tests for the swap scenario (scheduling and deck unchanged) and that cloze types are refused. Verify `go test ./core/...` passes.
- [x] 1.3 Use the override in card generation (`syncCards`) for newly created cards only, falling back to the note's deck when the deck is missing. Add tests for both override scenarios and for editing a note that creates a new card. Verify `go test ./core/...` passes.
- [x] 1.4 Use `BrowserQFmt` for the question text cache in `core/search.go` when set. Add a `search_test.go` case. Verify it passes.
- [x] 1.5 Map Anki's template `did` and `bqfmt`/`bafmt` in `importer/apkg` (override only when the deck exists after import). Add `apkg_test.go` coverage. Verify `go test ./importer/...` passes.

## 2. Service and UI

- [x] 2.1 Add `NotesService.MoveTemplate` and extend `UpdateTemplate` for the new fields; run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 2.2 Add a reorder handle and an "Options" sheet (deck override picker, browser templates) per card type in `note_type_manager.dart`. Add `test/editor_test.dart` tests for moving a card type and setting an override. Verify they pass.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
