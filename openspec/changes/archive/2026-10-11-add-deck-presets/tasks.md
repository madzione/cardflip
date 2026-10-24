# Tasks

## 1. Model and migration

- [x] 1.1 Add `core.Preset`, `Deck.PresetID` and `Deck.Limits`, and route every options read through `c.optionsOf(deckID)` (design decision 1). Keep all existing tests green. Verify `go test ./...` passes.
- [x] 1.2 Implement the v1 → v2 migration for snapshots, journals and restored backups (design decision 2), with the property test and the "Existing options kept" scenario. Bump the snapshot version and the backup format; add a `backup_test.go` case restoring a v1 backup. Verify `go test ./...` passes.
- [x] 1.3 Add preset operations (add, clone, rename, delete, assign, save to subdecks) and this-deck limits to `core`, with `decks_test.go` and `queue_test.go` cases for the spec scenarios. Verify `go test ./core/...` passes.

## 2. Search and import

- [x] 2.1 Add the `preset:` term to `golang/search` and fill it in `core/search.go`. Add `search_test.go` cases. Verify `go test ./search/... ./core/...` passes.
- [x] 2.2 Read presets from modern (`deck_config`) and legacy (`dconf`) packages and assign them per design decision 3. Add `apkg_test.go` cases for new, reused and renamed presets. Verify `go test ./importer/...` passes.

## 3. Service and UI

- [x] 3.1 Add preset calls and a deck options DTO with preset name, usage count and this-deck limits to `DecksService`. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean.
- [x] 3.2 Turn `deck_options_dialog.dart` into an options screen with the preset bar (picker, add, clone, rename, delete, save to subdecks, "Used by N decks") and this-deck limits. Add widget tests for switching presets, deleting, and this-deck limits. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Open a copy of a real v1 collection and confirm every deck's options are unchanged.
