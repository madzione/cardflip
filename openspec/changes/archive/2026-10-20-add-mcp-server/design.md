# Design

## Context

- The whole collection lives in the Go core, loaded into the Flutter process as a library (dart:ffi) on native platforms and as WebAssembly on the web. Dart reaches it only through fluffy-bound services (`golang/main.go`). Go never runs `main` on native platforms; setup happens in `fluffy.Start`.
- Service calls already arrive concurrently (fluffy dispatches each request on its own goroutine) and `core.Collection` guards its state with `c.mu`, so a second caller inside the process (the MCP server) is safe with no new locking.
- Every committed op emits a `core.Change`. `app.changes` forwards them to Dart's `ChangeHub` (`lib/data/changes.dart`), which refreshes open screens. Every op is also an undo step named by `core.stepName`, as a tagged text (`internal/usermsg`) that Dart translates.
- `app.collection()` returns the open profile's collection, or an error when none is open or it is corrupt. Profiles switch inside the same process; `profiles.json` in the data root (`storage.WriteProfiles`) is device-level state outside any collection.
- `core.AddNote` adds one note per op and stores the note type and deck as the Add screen's defaults (`LastNoteTypeID`, `LastDeckID`, `LastDeckByNoteType`).
- The fluffy generator type-checks the web build (`GOOS=js`), so every bound service must compile there. Platform-specific internals go in `!js` files with a js counterpart.
- `macos/Runner/DebugProfile.entitlements` already allows `network.server`; `Release.entitlements` does not.

## Goals / Non-Goals

**Goals:**
- The server and its tools run in Go, next to the collection, without a round trip through Dart.
- Tools reuse the existing service methods, so validation, messages and undo match the UI exactly.
- Nothing about the server reaches the web build's wasm size or behavior.

**Non-Goals:**
- MCP resources, prompts, sampling or elicitation: tools only.
- Studying (answering cards), deck options and presets, filtered decks, deleting decks, import/export, backups, profile switching, app settings.
- Remote access, TLS, OAuth, several tokens or per-tool permissions.
- A stdio transport or a bridge binary.

## Decisions

### 1. The server runs in Go, owned by an `McpService`

A new `McpService` is bound in `fluffy.Start` with the app's `*app`. It exposes to Dart: `Supported() bool`, `Status() McpStatus` (enabled, port, token, URL, running, error message), `SetEnabled(bool)`, `SetPort(int)`, `RegenerateToken()`, `Activity() []McpActivity`, and `StatusChanges(ctx) <-chan McpStatus` so Settings updates live when the server fails or an agent changes something.

The service type and its methods live in `svc_mcp.go` (builds everywhere). The listener, the MCP handlers and the tools live in `mcp_server.go` / `mcp_tools*.go` with `//go:build !js`; `mcp_server_js.go` provides a `startServer` that returns "not supported" and `Supported()` is false there. `Supported()` is also false when `runtime.GOOS` is not `linux`, `darwin` or `windows`, so Android and iOS builds never listen even if the Dart side were wrong.

*Alternative:* run the server in Dart (`dart:io` `HttpServer`) and call the Go services from it. Rejected: every tool call would cross the FFI bridge twice, and Dart has no MCP SDK as mature as Go's.

### 2. The server starts with the collection, not with the window

`app.open` (after loading the profile registry, so the data root is known) reads `mcp.json` and starts the server when it is enabled. Starting it in Go rather than from Dart's bootstrap means a hot restart, after which `open` returns early, leaves the running server alone instead of starting a second one. The server does not depend on a collection being open: tools call `app.collection()` per request, which follows profile switches and reports "no collection open" from the profile list or a corrupt collection (spec: Open collection required).

The process ending ends the listener; there is no separate shutdown path to forget. `SetEnabled(false)` calls `http.Server.Shutdown` with a short timeout and then `Close`, so open SSE streams end.

### 3. Device settings in `mcp.json`

`{ "enabled": bool, "port": int, "token": string }` next to `profiles.json` in the data root, read and written through two new functions in `storage` (files backend only; the web never calls them). Not in `core.Settings`, so it stays out of collections and backups, and is the same for every profile (spec: Device-level setting).

Default port **47811** (unassigned by IANA, fixed so client configurations survive restarts). The token is 32 bytes from `crypto/rand`, base64url, created on the first `SetEnabled(true)` and kept when the server is turned off and on again. The file is written with mode `0600`.

*Alternative:* a random free port per start. Rejected: users would have to update their agent configuration every launch.

### 4. Streamable HTTP with the official Go SDK

Use `github.com/modelcontextprotocol/go-sdk` (`mcp.NewServer`, `mcp.AddTool` with typed input/output structs, `mcp.NewStreamableHTTPHandler`), pinned to the newest release at implementation time. Typed tool handlers give JSON schemas for inputs and structured outputs from Go structs, with doc tags as descriptions.

The handler is wrapped in middleware, in this order:
1. **Origin check**: a present `Origin` that is not `http(s)://127.0.0.1[:port]`, `http(s)://localhost[:port]` or `http(s)://[::1][:port]` → 403 (DNS rebinding protection, as the MCP transport spec requires).
2. **Bearer token** compared with `subtle.ConstantTimeCompare` → 401 with `WWW-Authenticate: Bearer`.

The listener is `net.Listen("tcp", "127.0.0.1:<port>")`. A bind error is kept as the status's error message ("Port 47811 is already in use.") and the switch stays enabled (spec: Server errors shown). `SetPort` restarts the listener on the new port.

### 5. Tools wrap the existing services

Each tool is a thin adapter over a method of `NotesService`, `BrowseService`, `DecksService` or `StatsService` constructed with the same `*app`. This gives identical validation and undo steps for free. Adapters do three things the UI does in Dart instead:

- **Fields by name.** Tools take and return `map[string]string` keyed by field name; the adapter maps to and from positional fields using the note type, and fails on unknown names listing the valid ones.
- **Plain text.** Card question and answer are rendered with the current render environment and reduced to text: the answer drops everything up to `<hr id=answer>` when present, then HTML is converted to text (block elements to newlines, entities decoded, `<img>` as `[image: name]`, sounds as `[sound: name]`), reusing `golang/render`'s HTML tree.
- **English errors.** Tool errors are `usermsg.Plain(err.Error())`, returned as MCP tool errors (`isError: true`), not protocol errors, so agents can read and correct them. Undo step names in the activity list stay tagged so Settings shows them translated.

Tools are annotated `readOnlyHint` or `destructiveHint` so clients can ask the user before destructive calls. Tool set (names are the MCP tool names):

| Group | Tools |
|---|---|
| Read | `list_decks`, `get_deck_overview`, `list_note_types`, `list_tags`, `search_cards`, `search_notes`, `get_notes`, `get_cards`, `get_stats` |
| Notes | `add_notes`, `update_note`, `delete_notes`, `add_tags`, `remove_tags`, `rename_tag`, `store_media` |
| Cards | `suspend_cards`, `unsuspend_cards`, `set_flag`, `move_cards`, `reset_cards`, `set_due_date` |
| Decks | `create_deck`, `rename_deck` |
| Note types | `create_note_type`, `rename_note_type`, `add_field`, `rename_field`, `remove_field`, `move_field`, `add_template`, `update_template`, `remove_template`, `move_template`, `set_note_type_css`, `delete_note_type` |

Search tools take `query`, `limit` (default 50, max 500) and `offset`, call `Search`/`SearchNotes` for the sorted ID list and `Rows`/`NoteRows` for the page, and return `total` and `nextOffset`. Destructive note type tools (`remove_field`, `remove_template`, `delete_note_type`) report affected counts from the existing `TemplateCardCount` / `TypeDeletionInfo` calls in their result.

`get_stats` combines `StatsService.Today`, `Counts` and `FutureDue` (30 days) for a deck scope or the whole collection.

### 6. Batch add as one op that keeps the editor's defaults

`core` gains `AddNotes(in []NoteInput, opts AddNotesOptions) ([]AddResult, error)`: one op of kind `addNotes` (step name "Add N notes", a new tagged text with a German translation) holding all notes and cards, so one Undo reverts the batch (spec: Agent changes can be undone). `opts.KeepDefaults` skips updating the remembered note type and deck (spec: Added notes keep the editor's choices); `opts.SkipDuplicates` checks each note's first field against existing notes of its type and against earlier notes of the same batch, using the same normalization as `NotesService.Duplicates`. Validation of every note happens before anything is committed, so an invalid note fails the whole call.

`core.AddNote` is kept for the editor; `AddNote` and `AddNotes` share the per-note building code.

*Alternative:* call `AddNote` in a loop. Rejected: 500 undo steps, of which only the last 30 survive, and the Add screen's defaults would jump to whatever the agent used last.

### 7. Activity list from the change stream

`McpService` keeps a ring buffer of 50 `McpActivity{Time, Step}`. After a changing tool succeeds, its adapter reads the step name from the top of the undo stack (`Collection.UndoInfo()`), which is the step it just committed. The buffer lives in memory for the app session (spec: Recent agent activity) and is pushed to Dart through `StatusChanges`.

*Alternative:* return the step name from every service method. Rejected: it changes dozens of signatures the Dart side uses, for a list that is informational.

### 8. Settings UI

A new "Agents (MCP)" section in `settings_screen.dart`, built only when `!kIsWeb` and the target platform is Linux, macOS or Windows (and `golang.mcp.supported()`). It holds the switch, then, when on: status line (running at URL / error), port field with validation 1024–65535, token row (obscured, copy, regenerate with confirmation), "Copy client configuration" producing

```json
{ "mcpServers": { "cardflip": { "type": "http", "url": "http://127.0.0.1:47811/mcp",
  "headers": { "Authorization": "Bearer <token>" } } } }
```

and the activity list. Turning the switch on for the first time shows a dialog explaining that any program on this computer with the token can read and change the collection. All strings go into `app_en.arb` and `app_de.arb`.

## Risks / Trade-offs

- [Any local process that reads `mcp.json` gets the token] → File mode `0600`; same trust boundary as the collection files next to it, which such a process could edit directly anyway.
- [An agent fills the undo stack (30 steps) and pushes the user's own steps out] → Batch tools keep most agent work to few steps; the activity list shows what happened. Accepted.
- [An agent edits a note the user has open in the editor] → Same as today with two windows of the browser: the editor's save writes its fields over the agent's. Accepted; the Change stream already refreshes the browser list.
- [A UI change committed between an agent's commit and the activity read gives the activity entry the UI step's name] → The window is microseconds wide and the list is informational; undo itself is unaffected. Accepted.
- [Large batches block the UI] → Batches are capped at 500 notes; adding 500 notes is a single locked commit comparable to a small CSV import, which the app already does in one op.
- [The SDK's protocol versions move quickly] → Pin the SDK, cover the transport with an HTTP-level test (initialize, `tools/list`, one tool call) so an upgrade that breaks it fails `go test`.
- [macOS release builds cannot listen without the entitlement] → Add `com.apple.security.network.server` to `Release.entitlements`; the macOS smoke test enables the server.
- [New dependency enlarges the native library] → Only `!js` files import the SDK; the web build is unchanged. Check the native library size before and after.

## Migration Plan

No data migration: without `mcp.json` the server is off. Removing the feature later only leaves an unused `mcp.json`.

## Open Questions

- Exact default port: 47811 unless it turns out to clash with a common local tool during implementation; changing it touches only the default constant and the spec scenario.

## Implementation Notes

- **Native library size** (Linux amd64, `fluffy lib linux`): 12,209,128 → 23,782,960 bytes unstripped, 8,777,896 → 17,168,680 bytes stripped (+8.4 MB). Before this change the library linked neither `net/http` nor `crypto/tls`; the growth is Go's HTTP stack (with TLS and HTTP/2) plus the SDK and its dependencies (`jsonschema-go`, `oauth2`, `encoding/json/v2`). The web build links none of it (`GOOS=js go list -deps` has no SDK package).
- **Server start**: the hook runs after the first successful load of the profile registry, which both `app.open` and `AppService.LoadProfiles` do, so the server also runs while the profile list is shown at startup (spec: Open collection required). A hot restart loads nothing new and starts nothing.
- **Activity**: `core.Collection.LastStep` returns a push counter with the top step's name, so a tool call that changed nothing (e.g. adding a tag a note has) is not listed under the previous step's name.
- **Request size**: the Streamable HTTP handler's body limit is raised from the SDK's 4 MiB to 32 MiB so that `store_media` can take 20 MB files as base64.
- **Tool inputs**: decks and note types are named (full deck name, case-insensitive); notes and cards by ID. `set_flag` only sets the flag on cards that lack it, rather than the browser's toggle-off when all have it.
