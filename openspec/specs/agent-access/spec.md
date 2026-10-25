# agent-access Specification

## Purpose

Lets AI agents create and manage cards in the open collection through a Model Context Protocol server that runs inside the desktop app while it is open, only after the user turns it on.

## Requirements

### Requirement: Opt-in MCP server
On Linux, macOS and Windows, Settings SHALL offer a switch "Allow agents (MCP)", off by default. While it is on and the app is running, the app SHALL serve MCP over Streamable HTTP at `http://127.0.0.1:<port>/mcp`. Turning it off SHALL stop the server at once and end open agent sessions.

#### Scenario: Off by default
- **WHEN** the app runs for the first time on Linux
- **THEN** the switch is off and nothing listens on the MCP port

#### Scenario: Turn on
- **WHEN** the user turns the switch on
- **THEN** the server accepts MCP requests at the shown URL without restarting the app

#### Scenario: Turn off
- **WHEN** an agent is connected and the user turns the switch off
- **THEN** the agent's next request fails to connect

### Requirement: Server lives with the app
The MCP server SHALL run only inside the running app. It SHALL stop when the app quits and start again when the app is next opened with the switch on. The app SHALL NOT offer a way to run the server without its window.

#### Scenario: App quit
- **WHEN** the switch is on and the user quits the app
- **THEN** nothing listens on the MCP port anymore

#### Scenario: Next start
- **WHEN** the switch was on when the app quit and the user opens the app again
- **THEN** the server runs again with the same URL and token

### Requirement: Device-level setting
The switch, port and token SHALL belong to the device: they SHALL be the same for every profile, SHALL NOT be stored in the collection, and SHALL NOT be included in backups.

#### Scenario: Backup restored elsewhere
- **WHEN** a backup is made on a computer with the server on and restored on another computer
- **THEN** the server is still off on the other computer

#### Scenario: Profile switch
- **WHEN** the server is on and the user opens another profile
- **THEN** the server keeps running, and agents from then on work on the newly opened profile's collection

### Requirement: Not offered on other platforms
On the web, Android and iOS, the app SHALL NOT show the MCP settings and SHALL NOT listen for connections.

#### Scenario: Web build
- **WHEN** the user opens Settings in the web build
- **THEN** there is no "Agents (MCP)" section

### Requirement: Connection details in Settings
While the switch is on, Settings SHALL show the server URL, the token with actions to copy and regenerate it, a client configuration ready to paste that contains both, and the port, which the user SHALL be able to change (1024 to 65535). Regenerating the token SHALL make the old one invalid at once.

#### Scenario: Copy configuration
- **WHEN** the user copies the client configuration
- **THEN** the clipboard holds a configuration with the server URL and an `Authorization: Bearer <token>` header

#### Scenario: Regenerate token
- **WHEN** an agent uses the old token after the user regenerated it
- **THEN** the request is rejected as unauthorized

#### Scenario: Change port
- **WHEN** the user changes the port from 47811 to 50000
- **THEN** the server moves to port 50000 and the shown URL and configuration change with it

### Requirement: Server errors shown
When the server cannot start, Settings SHALL show why, such as a port already in use, and the switch SHALL stay on so the server starts once the problem is fixed by changing the port or on the next start. A failing server SHALL NOT affect the rest of the app.

#### Scenario: Port in use
- **WHEN** another program listens on the configured port and the user turns the switch on
- **THEN** Settings says the port is in use, and the user can pick another port

### Requirement: Token required
Every MCP request SHALL carry the token as `Authorization: Bearer <token>`. Requests without it or with a wrong token SHALL be rejected with HTTP 401 and SHALL NOT read or change anything. The token SHALL be random, at least 128 bits, and created when the server is first turned on.

#### Scenario: Missing token
- **WHEN** a client sends an MCP request without an Authorization header
- **THEN** the server answers 401 and the collection is unchanged

### Requirement: Local connections only
The server SHALL listen only on the loopback interface. It SHALL reject requests with an `Origin` header that is not a loopback origin, so web pages in a browser cannot reach it.

#### Scenario: Other computer
- **WHEN** another computer on the network connects to the device's address on the MCP port
- **THEN** the connection is refused

#### Scenario: Web page request
- **WHEN** a web page at `https://example.com` sends a request to the server with the right token
- **THEN** the server answers 403

### Requirement: Open collection required
Tools SHALL act on the collection of the profile that is open in the app. When no collection is open (the profile list is shown at startup, or the collection could not be read), every tool SHALL fail with a message saying so.

#### Scenario: Profile list at startup
- **WHEN** the app shows the profile list at startup and an agent calls a tool
- **THEN** the tool fails with a message that no collection is open

### Requirement: Same rules as the app
Every tool that changes the collection SHALL apply the same validation and messages as the same action in the app, and SHALL change nothing when it fails. A failed tool call SHALL return the reason as an MCP tool error, in English.

#### Scenario: Duplicate deck name
- **WHEN** an agent creates deck `Spanish` and a deck named `spanish` already exists at the top level
- **THEN** the tool fails with the message the app shows for a taken deck name, and no deck is created

### Requirement: Changes show in the app
Changes made by agents SHALL appear in open screens without user action, as changes made in the app do.

#### Scenario: Browser open
- **WHEN** the browser shows deck `Spanish` and an agent adds a note to it
- **THEN** the new card appears in the list

### Requirement: Agent changes can be undone
Each tool call that changes the collection SHALL be one undo step, named as the same action in the app is, and the app's Undo SHALL revert it.

#### Scenario: Undo a batch
- **WHEN** an agent adds 40 notes in one call and the user chooses Undo
- **THEN** all 40 notes are removed, and the Undo item had read "Undo Add 40 notes"

### Requirement: Recent agent activity
Settings SHALL list the agent's most recent changes of this app session, newest first: at most 50 entries, each with its time and the undo step's name. Failed calls and calls that only read SHALL NOT be listed.

#### Scenario: Activity listed
- **WHEN** an agent adds 3 notes and then deletes 1 note
- **THEN** the list shows "Delete 1 note" above "Add 3 notes", each with its time

### Requirement: Read tools
The server SHALL offer tools to list decks with their due counts, show a deck's overview, list note types with their fields and templates, list tags, search cards and notes with the app's search syntax, and get notes and cards by ID. Search results SHALL be paged, with at most 500 per call.

#### Scenario: Search
- **WHEN** an agent searches for `deck:Spanish is:due` with limit 20
- **THEN** it gets at most 20 matching cards, their total count, and an offset for the next page

#### Scenario: Invalid search
- **WHEN** an agent searches with an unbalanced parenthesis
- **THEN** the tool fails with the message the browser shows for that search

### Requirement: Card content for agents
Getting a card SHALL return its note's fields by name, its tags, deck, card type, scheduling (state, due date, interval, reviews, lapses, suspended, flag), and its question and answer rendered as plain text, without HTML and without the question repeated in the answer.

#### Scenario: Basic card
- **WHEN** an agent gets the card of a Basic note with Front `hola` and Back `<b>hello</b>`
- **THEN** the question text is `hola` and the answer text is `hello`

#### Scenario: Cloze card
- **WHEN** an agent gets card 2 of the cloze note `{{c1::Paris}} is the capital of {{c2::France}}`
- **THEN** the question text is `Paris is the capital of [...]` and the answer text is `Paris is the capital of France`

### Requirement: Collection statistics tool
The server SHALL offer a tool returning, for the collection or one deck, today's studied cards and time, the counts of new, learning, review and suspended cards, and the number of cards due on each of the next 30 days. It SHALL also return the collection's current streak, whether today is a study day, the daily goal with today's progress, or that no goal is set, and the collection's level with its XP and the XP needed for the next level.

#### Scenario: Deck forecast
- **WHEN** an agent asks for the statistics of deck `Spanish`
- **THEN** it gets those counts for `Spanish` and its subdecks only

#### Scenario: Goal progress
- **WHEN** the goal is 30 cards, 20 cards were answered today and the streak is 12 days, and an agent asks for statistics
- **THEN** it gets a current streak of 12 days, today studied, and goal progress of 20 of 30 cards

#### Scenario: Level
- **WHEN** the user has 3,760 XP and an agent asks for the statistics of deck `Spanish`
- **THEN** it gets the collection's level 12, 3,760 XP and 3,900 XP for level 13

### Requirement: Add notes
The server SHALL offer a tool to add 1 to 500 notes in one call, each with a note type, a deck, fields given by name and tags. Unknown field names SHALL fail the call. Missing fields SHALL be empty. With duplicate checking on (the default), a note whose first field matches an existing note of the same note type SHALL be skipped and reported. The call SHALL return each added note's ID and cards.

#### Scenario: Batch add
- **WHEN** an agent adds 3 Basic notes to deck `Spanish` with fields Front and Back
- **THEN** 3 notes with 3 cards are added, and their IDs are returned

#### Scenario: Unknown field
- **WHEN** an agent adds a Basic note with a field `Answer`
- **THEN** the call fails naming the unknown field and the note type's fields, and no note of the call is added

#### Scenario: Duplicate skipped
- **WHEN** an agent adds a Basic note whose Front is `hola` and a Basic note with Front `hola` exists
- **THEN** that note is skipped and reported as a duplicate, and the other notes of the call are added

### Requirement: Added notes keep the editor's choices
Notes added by agents SHALL NOT change the note type and deck the Add screen starts with.

#### Scenario: Add screen unchanged
- **WHEN** the user last added a note of type Basic to `Default` and an agent then adds a Cloze note to `Spanish`
- **THEN** the Add screen still starts with Basic and `Default`

### Requirement: Edit and delete notes
The server SHALL offer tools to update a note's fields (by name) and tags, to delete notes, to add and remove tags on notes, and to rename a tag. Deleting SHALL report how many notes and cards were deleted.

#### Scenario: Update a field
- **WHEN** an agent sets field Back of a note to `hello`
- **THEN** the note's other fields and its scheduling are unchanged

### Requirement: Media for fields
The server SHALL offer a tool that stores a media file (image, audio or video, base64, at most 20 MB) and returns the file name to use in fields, such as `<img src="name">` or `[sound:name]`, renamed as the editor would when the name is taken by a different file.

#### Scenario: Image for a note
- **WHEN** an agent stores `map.png` and adds a note whose Back is `<img src="map.png">`
- **THEN** the card shows the image in the app

### Requirement: Card action tools
The server SHALL offer tools to suspend and unsuspend cards, set or clear their flag, move them to a deck, reset them to new, and set their due date with the same input as the app's Set due date.

#### Scenario: Suspend
- **WHEN** an agent suspends 5 cards
- **THEN** the cards are suspended and no longer shown when studying

### Requirement: Deck tools
The server SHALL offer tools to create a deck, including nested decks with `::`, and to rename a deck. Agents SHALL NOT be able to delete decks or change deck options.

#### Scenario: Nested deck
- **WHEN** an agent creates `Spanish::Verbs`
- **THEN** `Spanish` and its subdeck `Verbs` exist

### Requirement: Note type tools
The server SHALL offer tools to create a note type from a stock type or as a copy of another, rename it, add, rename, remove and reorder its fields, add, change, remove and reorder its card templates, set its CSS, and delete it. Changes that remove fields, templates or a note type SHALL report how many notes or cards they affect.

#### Scenario: Add a field
- **WHEN** an agent adds field `Example` to note type `Vocab`
- **THEN** every note of `Vocab` has an empty `Example` field

#### Scenario: Template error
- **WHEN** an agent sets a front template that references an unknown field
- **THEN** the tool fails with the template check's message and the note type is unchanged
