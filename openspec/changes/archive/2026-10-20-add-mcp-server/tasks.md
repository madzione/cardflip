# Tasks

## 1. Core: batch add

- [x] 1.1 Add `Collection.AddNotes` with `KeepDefaults` and `SkipDuplicates` options as one `addNotes` op (design decision 6), sharing per-note building with `AddNote`; add the "Add N notes" step name with its German translation in `lib/l10n/go_messages.dart`/arb. Add `golang/core/notes_test.go` cases: one undo step reverts the batch, defaults unchanged, duplicates skipped (also within the batch), an invalid note fails the whole call with nothing committed. Verify `go test ./core/...` passes.

## 2. Server and settings in Go

- [x] 2.1 Add `storage` read/write of `mcp.json` in the data root (files backend, mode `0600`) with tests in `storage/profiles_files_test.go` style. Verify `go test ./storage/...` passes.
- [x] 2.2 Add `github.com/modelcontextprotocol/go-sdk` (newest release) to `golang/go.mod`. Add `svc_mcp.go` (`McpService`: `Supported`, `Status`, `SetEnabled`, `SetPort`, `RegenerateToken`, `Activity`, `StatusChanges`), `mcp_server.go` (`!js`: loopback listener, origin check, bearer auth, Streamable HTTP handler, start/stop/restart, port-in-use message) and `mcp_server_js.go` (stub). Bind it in `golang/main.go` and start the server from `app.open` when enabled. Add `svc_test.go` cases: off by default, enable/disable, token created once and kept, regenerate invalidates the old token (401), missing token 401, foreign `Origin` 403, port in use reported with the switch kept on, port change moves the server, server keeps running across a profile switch. Verify `go test ./...` and `GOOS=js GOARCH=wasm go vet ./...` pass.

## 3. Tools

- [x] 3.1 Add the shared adapter pieces in `mcp_tools.go`: fields by name ↔ positional, plain-text rendering of question/answer, English tool errors via `usermsg.Plain`, "no collection open" handling, activity recording from `UndoInfo`. Add unit tests for the field mapping (unknown name lists valid ones) and the text rendering (Basic and cloze scenarios from the spec). Verify `go test ./...` passes.
- [x] 3.2 Add the read tools (`list_decks`, `get_deck_overview`, `list_note_types`, `list_tags`, `search_cards`, `search_notes`, `get_notes`, `get_cards`, `get_stats`) with paging (default 50, max 500). Test each through an in-process MCP client session (`mcp.NewInMemoryTransports`): paging and `nextOffset`, invalid search message, deck-scoped stats. Verify `go test ./...` passes.
- [x] 3.3 Add the note tools (`add_notes` via `AddNotes`, `update_note`, `delete_notes`, `add_tags`, `remove_tags`, `rename_tag`, `store_media` with the 20 MB limit). Test batch add, unknown field, duplicate skipped, Add screen defaults unchanged, delete counts, media rename on a name clash, and that a changing call adds one activity entry while a read call adds none. Verify `go test ./...` passes.
- [x] 3.4 Add the card and deck tools (`suspend_cards`, `unsuspend_cards`, `set_flag`, `move_cards`, `reset_cards`, `set_due_date`, `create_deck`, `rename_deck`). Test suspend, nested deck creation and the taken-name message. Verify `go test ./...` passes.
- [x] 3.5 Add the note type tools (`create_note_type`, `rename_note_type`, `add_field`, `rename_field`, `remove_field`, `move_field`, `add_template`, `update_template`, `remove_template`, `move_template`, `set_note_type_css`, `delete_note_type`) with affected counts on removal and `readOnlyHint`/`destructiveHint` annotations on every tool. Test add field, the refused template with its message, and delete counts. Verify `go test ./...` passes.
- [x] 3.6 Add an HTTP-level test against a real loopback listener: initialize, `tools/list` lists every tool from design decision 5, one `add_notes` call succeeds and its change arrives on `app.changes`. Verify `go test ./...` passes.

## 4. App

- [x] 4.1 Run `fluffy generate`, add `mcp.*` handlers to `test/support/fake_go.dart`. Verify `fluffy generate --check` is clean and `fvm flutter analyze` passes.
- [x] 4.2 Add the "Agents (MCP)" section to `lib/features/settings/settings_screen.dart` (switch with first-time explanation dialog, status/URL, port field 1024–65535, token copy and regenerate with confirmation, copy client configuration, activity list with translated step names), shown only on Linux, macOS and Windows; strings in `app_en.arb` and `app_de.arb`. Add `test/settings_test.dart` tests: hidden on web/Android, enable shows URL, port error shown, copied configuration contains URL and bearer header, activity listed newest first. Verify `fvm flutter test` passes, including `no_literal_text_test.dart` and `l10n_test.dart`.
- [x] 4.3 Add `com.apple.security.network.server` to `macos/Runner/Release.entitlements`. Verify the file matches `DebugProfile.entitlements` for that key.
- [x] 4.4 Add an "Agents (MCP)" section to `README.md`: what it does, desktop only, how to turn it on, and the Claude Code command `claude mcp add --transport http cardflip http://127.0.0.1:47811/mcp --header "Authorization: Bearer <token>"`. Verify the command works against the running app.

## 5. Integration check

- [x] 5.1 Run `go test ./...` and `GOOS=js GOARCH=wasm go vet ./...` in `golang/`, `fvm flutter analyze` and `LANG=en_US.UTF-8 fvm flutter test`, and verify all pass. Add an integration test (Linux) that enables the server in Settings, adds a note over HTTP with the shown token and sees the card in the browser, then undoes it from the app. Run it on Linux, and run the existing web integration test to confirm the web build is unaffected. Compare the native library size before and after and note it in the change.

## Workflow follow-up

- Add `add-mcp-server` to `ROADMAP.md` under "12. App-wide" and tick it once archived.
- Archive with `/opsx:archive add-mcp-server` after review.
