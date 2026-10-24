# Tasks

## 1. Collection and search

- [x] 1.1 Add `Flag int` (`json:"flag,omitempty"`) to `core.Card` and `Collection.SetFlag(ids []ID, flag int)` with the toggle rule from design decision 2, rejecting values outside 0–7. Add `core/cards_test.go` cases: set, toggle off, mixed selection sets all, invalid value rejected, scheduling unchanged. Verify `go test ./core/...` passes.
- [x] 1.2 Add `FlagNames [7]string` to `core.Settings` with an `AppService.SetFlagName(flag int, name string)` and the names in the `Settings` DTO. Add a `svc_test.go` case that renames flag 1 and reads it back after reopening. Verify `go test ./...` passes.
- [x] 1.3 Add the `flag:N` term to `golang/search/search.go` (and `Flag` to `search.Card`, filled in `core/search.go`). Add `search_test.go` cases for `flag:0`, `flag:7`, `-flag:1` and the invalid `flag:8`. Verify `go test ./search/... ./core/...` passes.
- [x] 1.4 Add a backup round-trip test asserting flags survive export and restore. Verify `go test ./...` passes.

## 2. Rendering and import

- [x] 2.1 Pass the card's flag into the render context and add the `CardFlag` special field. Add `render/template_test.go` cases for a flagged and an unflagged card. Verify `go test ./render/...` passes.
- [x] 2.2 Read the low 3 bits of `cards.flags` in `importer/apkg/reader.go` and carry them through `ImportNote`/`core.Import` for new cards when scheduling is kept. Extend `apkg_test.go` with a fixture card flagged green, imported with and without scheduling. Verify `go test ./importer/... ./core/...` passes.

## 3. Services and generated API

- [x] 3.1 Add the flag to `StudyCard` and `Row`, add `StudyService.SetFlag(cardID, flag)` (returns the new state) and `BrowseService.SetFlag(cardIDs, flag)`. Add `svc_test.go` cases. Run `fluffy generate` and verify `fluffy generate --check` is clean.
- [x] 3.2 Add the new calls and fields to `test/support/fake_go.dart` and verify the existing `fvm flutter test` suite still passes.

## 4. Flutter

- [x] 4.1 Add a flag colour table (light and dark variants) in `lib/app/theme.dart` and a Flags section in `settings_screen.dart` to rename flags. Add a `test/settings_test.dart` test that renaming calls `app.SetFlagName`. Verify it passes.
- [x] 4.2 In `study_screen.dart`, add Ctrl/Cmd+1…7 shortcuts, a Flag submenu in the More menu using the flag names, and a coloured flag marker on the card. Add `test/study_test.dart` tests for toggling with the keyboard and changing the flag from the menu. Verify they pass.
- [x] 4.3 In `browse_screen.dart`, add the Flag bulk action (seven flags plus "No flag"), the row flag marker and the row tint (flag colour, else yellow for suspended); add the same action to `card_details.dart`. Add `test/browse_test.dart` tests for bulk flagging, the suspended tint and flag-over-suspended. Verify they pass.

## 5. Integration check

- [x] 5.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 5.2 Manually on desktop: import an `.apkg` with flagged cards keeping scheduling, confirm the flags show in Browse, search `flag:1`, flag a card with Ctrl+2 while studying, and confirm the row colour changes.
