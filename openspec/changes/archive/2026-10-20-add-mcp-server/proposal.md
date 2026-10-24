# Proposal

## Why

People increasingly use AI agents (Claude Code, Claude Desktop, other MCP clients) to turn notes, articles and course material into flash cards. Today an agent can only reach Cardflip by writing a CSV or `.apkg` file that the user then imports by hand, one round trip per batch, with no way to look at existing cards, fix a typo or avoid duplicates. An MCP server in the running app lets an agent work directly on the open collection while the user watches the result appear in the app.

## What Changes

- **MCP server in the app**: on Linux, macOS and Windows, Cardflip can run a [Model Context Protocol](https://modelcontextprotocol.io) server (Streamable HTTP) on `127.0.0.1`, while the app is open. It stops when the app quits. There is no headless mode and no separate server binary.
- **Opt-in**: off by default. A new **Agents (MCP)** section in Settings turns it on, shows the server URL, a bearer token (copy, regenerate) and a ready-to-paste client configuration, the port (changeable), errors such as a port in use, and the agent's recent changes. The setting belongs to the device, not to the collection: it is not in backups and not per profile. The section is not shown on the web, Android or iOS.
- **Security**: every request needs the bearer token; the server listens only on the loopback interface and rejects requests from web pages (`Origin` check).
- **Tools for agents**, acting on the open profile's collection through the same rules as the UI:
  - read: decks with due counts, deck overviews, note types, tags, search for cards and notes with the app's search syntax, notes and cards (fields, rendered question and answer as text, scheduling), collection statistics;
  - notes and cards: add notes (in batches, with a duplicate check), update and delete notes, add and remove tags, store media files for use in fields;
  - card actions: suspend and unsuspend, flag, move to a deck, reset to new, set due date;
  - decks: create and rename;
  - note types: create (from a stock type or a copy), rename, edit fields, card templates and CSS, delete.
- Agent changes are ordinary collection changes: open screens refresh, each tool call is one undo step named like the UI's (a batch of added notes is one step), and the app's Undo reverts them.

## Capabilities

### New Capabilities

- `agent-access`: the opt-in MCP server, its settings, security and lifetime, and the tools it offers agents.

### Modified Capabilities

- `app-shell`: "Supported platforms" gains an exception: the MCP server is only available on the desktop platforms.

## Impact

- Go: new `McpService` bound with fluffy (status, enable/disable, port, token, recent activity); the server itself in `!js` files using the official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk`, new dependency), with a stub for the web build. Tools call the existing services (`NotesService`, `BrowseService`, `DecksService`, `StatsService`). Device settings in `mcp.json` next to `profiles.json` in the data root. `core` gains a batch note add that is one undo step and leaves the editor's remembered note type and deck alone.
- Flutter: Settings section and its strings (English, German), `test/support/fake_go.dart` handlers.
- Platform: `macos/Runner/Release.entitlements` gains `com.apple.security.network.server` (DebugProfile already has it).
- Tests: Go tests for every tool, auth, origin check, lifecycle and the batch add; `test/settings_test.dart`; an integration test on Linux that enables the server and adds a note over HTTP.
