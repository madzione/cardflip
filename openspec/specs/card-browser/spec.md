# card-browser Specification

## Purpose

Lets the user find, inspect and edit any card or note in the collection, independent of studying.

## Requirements

### Requirement: Card list
The browser SHALL list cards in columns chosen by the user from: question, answer, sort field, card type, note type, deck, due, interval, difficulty, stability, retrievability, reviews, lapses, tags, created, card modified and note modified, with text shown without HTML. The user SHALL be able to show, hide and reorder columns, and the layout SHALL be remembered. The list SHALL be sortable by any column, ascending or descending, and SHALL stay responsive with 50,000 cards. By default it SHALL show question, deck, due and note type.

#### Scenario: Large collection
- **WHEN** the collection has 50,000 cards and the user opens the browser
- **THEN** the first rows appear within 1 second on a typical desktop, and scrolling remains smooth

#### Scenario: Sort by lapses
- **WHEN** the user shows the Lapses column and sorts it descending
- **THEN** the card with the most lapses is listed first

#### Scenario: Layout remembered
- **WHEN** the user hides the Note type column, adds Interval, and restarts the app
- **THEN** the browser shows Question, Deck, Due and Interval

### Requirement: Search
The browser SHALL filter by free text, matched case-insensitively against all fields with HTML stripped. It SHALL also support the qualifiers `deck:<name>` (including children), `tag:<tag>`, `is:new`, `is:due`, `is:review`, `is:learn`, `is:suspended` and `note:<note type>`. Several terms SHALL be combined with AND, and a term prefixed with `-` SHALL be negated.

#### Scenario: Combined search
- **WHEN** the user searches `deck:Spanish tag:verb -is:suspended`
- **THEN** only unsuspended cards in Spanish or its subdecks whose note has tag `verb` are listed

#### Scenario: Invalid qualifier
- **WHEN** the user searches `is:bogus`
- **THEN** the browser shows an error for that term instead of an empty list

### Requirement: Card details and editing
Selecting a card SHALL show its rendered question and answer, its scheduling information (state, due, interval, stability, difficulty, reviews, lapses), and an editor for its fields and tags. The editor SHALL come directly after the question and answer, before the scheduling information. Saves SHALL regenerate cards according to the collection rules. When the card's note has other cards, the editor SHALL say that changes also apply to them. For an Image Occlusion note, the editor SHALL be the mask editor in place of the Occlusion and Image fields, followed by the note's other fields.

#### Scenario: Edit from browser
- **WHEN** the user changes a field in the browser editor and saves
- **THEN** the list row and the preview show the new content

#### Scenario: Shared note is disclosed
- **WHEN** the user opens a card whose `Basic (and reversed card)` note has two cards
- **THEN** the editor says that changes also apply to 1 other card

#### Scenario: Image Occlusion card in the browser
- **WHEN** the user selects a card of an Image Occlusion note in the browser
- **THEN** the editor shows the note's image with its shapes and the shape tools, and Header, Back Extra and Comments as fields; the raw Occlusion and Image fields are not shown

#### Scenario: Edit action on an Image Occlusion card
- **WHEN** the user chooses Edit on a card of an Image Occlusion note
- **THEN** the editor scrolls into view and the Header field gets the focus

### Requirement: Bulk actions
The user SHALL be able to select multiple cards and suspend or unsuspend them, move them to a deck, add or remove tags, reset them to New, or delete them. The Browse toolbar SHALL have a visible control that starts selecting, and long-press on a row SHALL also start selecting. Destructive actions SHALL require confirmation.

#### Scenario: Bulk move
- **WHEN** the user selects 30 cards and moves them to deck `Archive`
- **THEN** all 30 cards are in `Archive` with unchanged scheduling

#### Scenario: Reset to new
- **WHEN** the user resets a Review card to New
- **THEN** the card is New, its memory state is cleared, and its review history is kept

#### Scenario: Select without long-press
- **WHEN** the user clicks the Select button in the Browse toolbar and then clicks two rows
- **THEN** both cards are selected and the bulk actions are available

### Requirement: Add cards
The user SHALL be able to add cards from the browser or the deck list, choosing a note type and deck that are remembered for next time. After a successful add, the editor SHALL clear the fields that are not pinned, keeping pinned fields, deck, note type and tags, for rapid entry. Pins SHALL be remembered per note type.

#### Scenario: Rapid entry
- **WHEN** the user adds a card and the add succeeds
- **THEN** the fields are empty, the deck, note type and tags are unchanged, and focus is in the first field

#### Scenario: Pinned field kept
- **WHEN** the Source field is pinned, the user adds a card with Source `Book 1`, and the add succeeds
- **THEN** Source still contains `Book 1` and the other fields are empty

### Requirement: Single-card actions
The card details SHALL offer Edit, Delete, and Suspend or Unsuspend (whichever applies to the card's state), on wide and narrow layouts. Edit SHALL bring the editor into view and focus its first field. Suspend and Unsuspend SHALL take effect immediately, and the details SHALL show the new state.

#### Scenario: Unsuspend from details
- **WHEN** the user opens a suspended card in the browser and chooses Unsuspend
- **THEN** the card is no longer suspended, its row no longer says "suspended", and the action changes to Suspend

#### Scenario: Edit action
- **WHEN** the user chooses Edit in the card details
- **THEN** the editor is visible and its first field has focus

### Requirement: Deleting cards
Deleting a card SHALL delete its note with all of that note's cards. Before deleting, the browser SHALL ask for confirmation that states how many cards will be deleted, and SHALL say separately how many of them the user did not pick. After deleting from the details, the details SHALL close.

#### Scenario: Delete a single card
- **WHEN** the user chooses Delete in the details of a `Basic` card and confirms
- **THEN** the card is deleted, it disappears from the list, and the details close

#### Scenario: Delete reaches a sibling card
- **WHEN** the user deletes one card of a `Basic (and reversed card)` note
- **THEN** the confirmation says 2 cards will be deleted, including 1 other card from the same note

#### Scenario: Cancel delete
- **WHEN** the user chooses Delete and then cancels the confirmation
- **THEN** no card is deleted

### Requirement: Insert math in the editor
The note editor SHALL offer a math menu that wraps the selected text, or inserts an empty pair at the cursor, with inline math `\(…\)`, display math `\[…\]`, chemistry `\(\ce{…}\)`, `[latex]…[/latex]`, `[$]…[/$]` or `[$$]…[/$$]`. On platforms with a keyboard the same actions SHALL be available as Ctrl+M M, Ctrl+M E, Ctrl+M C, Ctrl+T T, Ctrl+T E and Ctrl+T M (Cmd on macOS).

#### Scenario: Wrap selection
- **WHEN** the user selects `x^2` in a field and chooses inline math
- **THEN** the field contains `\(x^2\)`

#### Scenario: Empty insert
- **WHEN** nothing is selected and the user presses Ctrl+M then E
- **THEN** `\[\]` is inserted and the cursor is between the delimiters

### Requirement: Flag bulk action
The user SHALL be able to set one of the seven flags on the selected cards, or clear their flags, from the browser's bulk actions and from the card details. When every selected card already has the chosen flag, choosing it SHALL clear it.

#### Scenario: Flag selection
- **WHEN** the user selects 5 cards and chooses flag 4
- **THEN** all 5 cards have flag 4

#### Scenario: Clear flags
- **WHEN** the user selects flagged cards and chooses "No flag"
- **THEN** none of them has a flag

### Requirement: Flag search
The search SHALL support `flag:N` for N from 0 to 7, where `flag:0` matches cards without a flag. Other values SHALL be reported as an invalid term.

#### Scenario: Search red flags
- **WHEN** the user searches `flag:1`
- **THEN** only cards with the red flag are listed

#### Scenario: Unflagged cards
- **WHEN** the user searches `deck:Spanish flag:0`
- **THEN** only cards in Spanish without a flag are listed

### Requirement: Row colours
Each browser row SHALL show the card's flag as a marker in the flag's colour, and SHALL be tinted with that colour. A suspended card without a flag SHALL be tinted yellow. Tints SHALL keep the row text readable in light and dark themes.

#### Scenario: Suspended row
- **WHEN** a suspended card without a flag is listed
- **THEN** its row has a yellow tint and still says "suspended"

#### Scenario: Flag wins over suspended
- **WHEN** a suspended card has flag 2
- **THEN** its row is tinted orange

### Requirement: Visual field editor
Note fields in the add and edit screens SHALL be edited visually with a toolbar for bold, italic, underline, subscript, superscript, text colour, highlight colour, clear formatting, bulleted and numbered lists, alignment and indentation. On platforms with a keyboard, Ctrl+B, Ctrl+I and Ctrl+U (Cmd on macOS) SHALL toggle bold, italic and underline. The saved field SHALL be HTML that renders the same formatting on cards.

#### Scenario: Make a word bold
- **WHEN** the user selects `perro` in the Front field, presses Ctrl+B and saves
- **THEN** the Front field contains `<b>perro</b>` and the card shows the word in bold

### Requirement: HTML source toggle
Each field SHALL have a toggle between the visual editor and its HTML source, also available as Ctrl+Shift+X (Cmd+Shift+X on macOS). A per-field note type option SHALL open that field in HTML mode by default.

#### Scenario: Edit source
- **WHEN** the user toggles the Back field to HTML, types `<i>x</i>` and toggles back
- **THEN** the visual editor shows an italic `x`

### Requirement: No silent loss of formatting
A field containing HTML that the visual editor cannot represent without loss SHALL open in HTML mode with a note saying so. Switching it to the visual editor SHALL require confirmation.

#### Scenario: Imported table
- **WHEN** the user edits an imported note whose Back field contains a `<table>`
- **THEN** the Back field opens in HTML mode with a note, and saving without changes keeps the table exactly

### Requirement: Paste with formatting
Pasting into a field SHALL keep bold, italic, underline, colours, lists and line breaks, and drop other formatting. Ctrl+Shift+V (Cmd+Shift+V on macOS) SHALL paste plain text.

#### Scenario: Plain paste
- **WHEN** the user copies bold text from a web page and pastes it with Ctrl+Shift+V
- **THEN** the field contains the text without bold

### Requirement: Resize images in the editor
Images in a field SHALL be shown in the visual editor, and the user SHALL be able to resize one by dragging a handle, which sets its width while keeping its aspect ratio.

#### Scenario: Shrink an image
- **WHEN** the user drags an image's handle to half its width and saves
- **THEN** the image is shown at that width on the card

### Requirement: Attach media in the editor
The editor SHALL let the user pick image, audio and video files and insert them at the cursor of the focused field: images as `<img>` and audio and video as `[sound:…]`. Dropping files onto a field on desktop and web SHALL do the same. The files SHALL be stored in the collection's media.

#### Scenario: Attach an image
- **WHEN** the user picks `cat.jpg` with the Attach button while the Back field has focus, and adds the card
- **THEN** the Back field contains an `<img>` referencing the stored file and the card shows the image

#### Scenario: Attach audio
- **WHEN** the user attaches `hola.mp3` to the Front field
- **THEN** the Front field contains `[sound:hola.mp3]`

### Requirement: Paste images
Pasting an image from the clipboard into a field SHALL store it as a media file and insert it as an `<img>`, in JPEG format unless the setting "Paste images as PNG" is on.

#### Scenario: Paste a screenshot
- **WHEN** the user copies a screenshot and pastes it into a field
- **THEN** the field shows the image and the media contains a new JPEG file

### Requirement: Record audio in the editor
The editor SHALL let the user record from the microphone and insert the recording into the focused field as `[sound:…]`. If microphone access is refused, the editor SHALL explain this and keep the user's input.

#### Scenario: Record pronunciation
- **WHEN** the user records three seconds of audio and confirms
- **THEN** the focused field contains a `[sound:…]` tag for the new recording

### Requirement: Media file naming
A media file added in the editor SHALL keep its extension. Adding a file with the same content as an existing media file SHALL reuse the existing file. Adding a different file whose name is already taken SHALL store it under a new name.

#### Scenario: Same image twice
- **WHEN** the user attaches the same `cat.jpg` to two notes
- **THEN** the media contains one copy, referenced by both notes

#### Scenario: Name collision
- **WHEN** the media already has a different `cat.jpg` and the user attaches another file named `cat.jpg`
- **THEN** the new file is stored under a different name and the existing one is unchanged

### Requirement: Editing buttons in the field toolbar
The buttons that change a field's content, Cloze deletion (in cloze note types), Math, Attach and Record audio, SHALL be together, in that order, at the end of the field toolbar, set apart from the formatting buttons, in the add and edit screens and the browser's edit pane, and nowhere else in those editors. They SHALL act on the focused field, or the last focused one, and SHALL stay enabled while it shows its HTML source.

#### Scenario: Basic note type
- **WHEN** the user opens the Add card screen with the Basic note type
- **THEN** the field toolbar ends with Math, Attach and Record audio, and the app bar has the Note types button and none of these

#### Scenario: Cloze note type
- **WHEN** the user switches the Add card screen to the Cloze note type
- **THEN** the field toolbar ends with Cloze deletion, Math, Attach and Record audio

#### Scenario: Same group when editing in the browser
- **WHEN** the user selects a card of a cloze note in the browser
- **THEN** the edit pane's field toolbar ends with Cloze deletion, Math, Attach and Record audio, and the row with Save has only Save

#### Scenario: Cloze in the browser
- **WHEN** the Text field in the browser's edit pane contains `{{c1::Paris}} is in France`, the user selects `France` and chooses Cloze deletion
- **THEN** the field contains `{{c1::Paris}} is in {{c2::France}}`

#### Scenario: HTML source
- **WHEN** the Back field shows its HTML source, the formatting buttons are disabled, and the user chooses inline math and then attaches `cat.jpg`
- **THEN** Math and Attach are enabled and the Back field's source contains `\(\)` and `<img src="cat.jpg">`

### Requirement: Add shortcut
On platforms with a keyboard, Ctrl+Enter (Cmd+Enter on macOS) in any field SHALL add the card in the add screen and save the note in the edit screen.

#### Scenario: Add from the keyboard
- **WHEN** the cursor is in the Back field and the user presses Ctrl+Enter
- **THEN** the card is added as if Add was pressed

### Requirement: Deck per note type
Settings SHALL offer "Deck for new cards" with the choices "Current deck" (default) and "Change deck depending on note type". With the second choice, choosing a note type in the add screen SHALL select the deck last used to add a card of that note type.

#### Scenario: Deck follows note type
- **WHEN** the setting is "Change deck depending on note type", the user last added Cloze cards to `Biology` and Basic cards to `Spanish`, and switches the note type to Cloze
- **THEN** the deck switches to `Biology`

### Requirement: Duplicate warning
While the user enters a note, the editor SHALL warn when its first field, with HTML stripped and case and surrounding whitespace ignored, equals the first field of another note of the same note type. The warning SHALL NOT prevent adding, SHALL ignore the note being edited and empty first fields, and SHALL offer to show the matching notes in the browser.

#### Scenario: Duplicate word
- **WHEN** a Basic note with Front `perro` exists and the user types `Perro ` into the Front field of a new Basic note
- **THEN** the field is marked and a message says this is a duplicate, with an action to show it

#### Scenario: Different note type
- **WHEN** only a Cloze note contains `perro` as its first field and the user types `perro` into a Basic note
- **THEN** no warning is shown

#### Scenario: Add anyway
- **WHEN** the warning is shown and the user presses Add
- **THEN** the note is added

### Requirement: Cloze insertion shortcuts
In cloze note types, Ctrl+Shift+C (Cmd+Shift+C on macOS) SHALL wrap the selection in a deletion with the next unused cloze number of the note, and Ctrl+Alt+Shift+C (Cmd+Option+Shift+C on macOS) SHALL use the highest number already in the note, or 1 when there is none. Numbers SHALL be taken from all fields of the note. The same two actions SHALL be available from the toolbar.

#### Scenario: Next number
- **WHEN** the Text field contains `{{c1::Paris}} is in France`, the user selects `France` and presses Ctrl+Shift+C
- **THEN** the field contains `{{c1::Paris}} is in {{c2::France}}`

#### Scenario: Same number
- **WHEN** the user instead presses Ctrl+Alt+Shift+C
- **THEN** the field contains `{{c1::Paris}} is in {{c1::France}}`

#### Scenario: Number from another field
- **WHEN** Back Extra contains `{{c3::x}}`, Text has no clozes, and the user presses Ctrl+Shift+C in Text
- **THEN** the new deletion in Text is `c4`

### Requirement: Reposition action and positions
The browser SHALL offer "Reposition" for selected cards with the start, step, randomize and shift options, and SHALL show each new card's position in its due column and details.

#### Scenario: Position shown
- **WHEN** a new card has position 42
- **THEN** its row shows `New #42`

### Requirement: Preset search
The search SHALL support `preset:<name>`, matching cards in decks that use the preset with that name, case-insensitively; quotes SHALL allow names with spaces.

#### Scenario: Search by preset
- **WHEN** decks `Spanish` and `French` use preset `Languages` and the user searches `preset:Languages`
- **THEN** the cards of both decks are listed

### Requirement: Export selected notes
The browser SHALL offer to export the notes of the selected cards as an `.apkg`, with the same options as deck export.

#### Scenario: Export a selection
- **WHEN** the user selects 10 cards from 8 notes and exports them
- **THEN** the package contains those 8 notes with all their cards

### Requirement: Export selection as text
The browser SHALL offer to export the selected cards' notes, or the selected cards, as plain text.

#### Scenario: Export selected notes
- **WHEN** the user selects 5 cards of 5 notes and exports them as notes
- **THEN** the file has 5 note lines

### Requirement: Card info from the browser
The card details SHALL offer an action that opens card info for the selected card.

#### Scenario: Open from details
- **WHEN** the user chooses Card info in a card's details
- **THEN** card info for that card opens with its review history

### Requirement: OR and grouping
The search SHALL support `or` between terms (case-insensitive), an optional explicit `and`, parentheses for grouping, and `-` before a group to negate it. AND, explicit or implicit, SHALL bind more tightly than OR.

#### Scenario: Either deck
- **WHEN** the user searches `(deck:Spanish or deck:French) is:due`
- **THEN** due cards from either deck are listed

#### Scenario: Negated group
- **WHEN** the user searches `-(tag:verb or tag:noun)`
- **THEN** cards whose notes have neither tag are listed

#### Scenario: Precedence
- **WHEN** the user searches `deck:A tag:x or tag:y`
- **THEN** cards in deck A with tag x, and cards anywhere with tag y, are listed

### Requirement: Escaping in searches
In search text, `\"`, `\(`, `\)`, `\-`, `\:` and `\\` SHALL stand for the literal characters. Quoted text SHALL be one term, with spaces and parentheses taken literally.

#### Scenario: Literal parenthesis
- **WHEN** a note contains `f(x)` and the user searches `"f(x)"`
- **THEN** the note's cards are listed

#### Scenario: Leading hyphen
- **WHEN** a note contains `-ing` and the user searches `\-ing`
- **THEN** the note's cards are listed, and the term is not treated as negation

### Requirement: Search syntax errors
A search with unbalanced parentheses, an `or` missing a term on either side, or an empty group SHALL show an error naming the problem, instead of a result list.

#### Scenario: Unclosed group
- **WHEN** the user searches `(deck:Spanish or deck:French`
- **THEN** the browser shows an error that a parenthesis is not closed

### Requirement: Wildcards and words
In free-text terms, `_` SHALL match any single character and `*` any run of characters, while `\_` and `\*` SHALL match the literal characters. `w:` SHALL match whole words, where a `*` at either end allows a prefix or suffix.

#### Scenario: Single-character wildcard
- **WHEN** notes contain `dog` and `dig` and the user searches `d_g`
- **THEN** both notes' cards are listed

#### Scenario: Whole word
- **WHEN** notes contain `dog` and `doggy` and the user searches `w:dog`
- **THEN** only the note with `dog` is listed

### Requirement: Field search
`<field>:<text>` SHALL match notes whose field of that name, with HTML stripped and case ignored, equals the text, where wildcards allow partial matches and the field name may contain `*`. `<field>:` SHALL match an empty field, `<field>:_*` a non-empty field, and `<field>:*` any note that has the field.

#### Scenario: Exact field
- **WHEN** a note's Front is `dog` and another's is `a dog`, and the user searches `front:dog`
- **THEN** only the first note is listed

#### Scenario: Empty field
- **WHEN** the user searches `back:`
- **THEN** notes whose Back field is empty are listed

### Requirement: Accent and cloze handling
`nc:` SHALL match text while ignoring accents and other combining marks, and `sc:` SHALL match text with cloze markup removed. With the setting "Ignore accents in search" on, all text and field terms SHALL ignore accents.

#### Scenario: Ignore accents
- **WHEN** a note contains `über` and the user searches `nc:uber`
- **THEN** the note is listed

#### Scenario: Across a cloze
- **WHEN** a note's Text is `{{c1::capital}} of France` and the user searches `"sc:capital of France"`
- **THEN** the note is listed

### Requirement: Regular expression search
`re:<expr>` SHALL match notes whose raw field content matches the expression, `<field>:re:<expr>` only that field, and `tag:re:<expr>` any tag. Matching SHALL ignore case unless the expression starts with `(?-i)`. An invalid expression SHALL show an error.

#### Scenario: Field regex
- **WHEN** the user searches `front:re:^\d+$`
- **THEN** notes whose Front consists only of digits are listed

### Requirement: Entity decoding
Text and field terms SHALL match field text with HTML entities decoded.

#### Scenario: Ampersand
- **WHEN** a field is stored as `salt &amp; pepper` and the user searches `"salt & pepper"`
- **THEN** the note is listed

### Requirement: Event searches
The search SHALL support `added:N`, `edited:N`, `rated:N`, `rated:N:R`, `introduced:N` and `resched:N`, matching cards added, whose note was edited, answered (with rating R from 1 to 4), first answered, or manually rescheduled within the last N study days, where 1 means today. `rated:` SHALL accept N up to 365. `prop:rated` and `prop:resched` SHALL compare the study day of the latest answer or manual reschedule with today.

#### Scenario: Added today
- **WHEN** the user added 12 cards today and searches `added:1`
- **THEN** those 12 cards are listed

#### Scenario: Failed this week
- **WHEN** the user searches `rated:7:1`
- **THEN** cards rated Again in the last 7 study days are listed

#### Scenario: Before rollover
- **WHEN** the rollover hour is 04:00, a card was added yesterday at 02:00, and the user searches `added:1` today at 10:00
- **THEN** that card is not listed, because it belongs to the previous study day

#### Scenario: Out of range
- **WHEN** the user searches `rated:400`
- **THEN** an error says the number of days must be at most 365

### Requirement: Recently added history
The add screen SHALL offer a History action that opens the browser with the search `added:1`.

#### Scenario: Review what was added
- **WHEN** the user adds 5 cards and chooses History
- **THEN** the browser opens with `added:1` and lists those cards

### Requirement: Property search
`prop:<property><op><number>` SHALL compare a card property with `=`, `!=`, `<`, `<=`, `>` or `>=`, for the properties `ivl` (days), `due` (days relative to today), `reps`, `lapses`, `pos`, `s` (stability in days), `d` (difficulty from 0 to 1) and `r` (retrievability from 0 to 1). Cards without the property, such as a New card for `ivl`, SHALL NOT match. Unknown properties SHALL show an error; `prop:ease` SHALL explain that `prop:d` is used instead.

#### Scenario: Long intervals
- **WHEN** the user searches `prop:ivl>=30`
- **THEN** cards with an interval of 30 days or more are listed

#### Scenario: Overdue
- **WHEN** the user searches `prop:due<0`
- **THEN** review cards that were due before today are listed

#### Scenario: At risk
- **WHEN** the user searches `prop:r<0.8`
- **THEN** cards whose current retrievability is below 0.8 are listed

### Requirement: Burial and state search
`is:buried` SHALL match cards buried manually or held back today by sibling burying; `is:buried-manually` and `is:buried-sibling` SHALL match each kind. `is:learn` SHALL match learning and relearning cards, and `is:review` SHALL match review and relearning cards.

#### Scenario: Lapsed cards
- **WHEN** the user searches `is:learn is:review`
- **THEN** only relearning cards are listed

### Requirement: Card type and tag search
`card:<name>` SHALL match cards of card types with that name, with wildcards; `card:<n>` SHALL match card type number n, or cloze number n for cloze notes. `tag:none` SHALL match notes without tags.

#### Scenario: Reverse cards
- **WHEN** the user searches `card:2`
- **THEN** the second card of each reversed note and the cloze 2 cards are listed

### Requirement: Deck and ID search
`deck:current` SHALL match the deck most recently studied and its subdecks. `deck:<name>::*` SHALL match only that deck's subdecks. `nid:` and `cid:` SHALL match notes or cards by one ID or a comma-separated list.

#### Scenario: Without subdecks
- **WHEN** the user searches `deck:French -deck:French::*`
- **THEN** cards directly in `French` are listed, and cards in its subdecks are not

#### Scenario: By card IDs
- **WHEN** the user searches `cid:101,102`
- **THEN** exactly those two cards are listed

### Requirement: Find and Replace dialog
The browser SHALL offer Find and Replace with the inputs find, replace with, in (all fields, a field of the selected notes' note types, or tags), selected notes only (on when cards are selected), treat input as regular expression, and ignore case. Before applying, it SHALL show how many notes will change.

#### Scenario: Preview count
- **WHEN** the user enters a find text that occurs in 7 of the selected notes
- **THEN** the dialog says 7 notes will change before the user confirms

### Requirement: Notes mode
The browser SHALL offer a Notes mode, remembered across restarts, that lists one row per note showing its first field, note type, card count, deck (or "multiple" when its cards are in several decks), tags and the earliest due date of its cards. A note SHALL be listed when any of its cards matches the search. Notes mode SHALL be sortable by created date, first field and earliest due date.

#### Scenario: Reversed note once
- **WHEN** the collection has one `Basic (and reversed card)` note and the user switches to Notes mode
- **THEN** the browser lists one row showing 2 cards

#### Scenario: Match through one card
- **WHEN** only one of a note's cards is suspended and the user searches `is:suspended` in Notes mode
- **THEN** the note is listed

### Requirement: Bulk actions in Notes mode
In Notes mode, bulk actions SHALL apply to all cards of the selected notes.

#### Scenario: Suspend notes
- **WHEN** the user selects 3 reversed notes in Notes mode and suspends them
- **THEN** all 6 cards are suspended

### Requirement: Browser sidebar
The browser SHALL show a sidebar with Card states, Decks, Note types and Tags, with decks and tags as trees. On narrow screens it SHALL be available as a drawer. A filter field SHALL narrow the sidebar to matching items.

#### Scenario: Tag tree
- **WHEN** notes have the tags `verb::irregular` and `verb::regular`
- **THEN** the sidebar shows `verb` with the children `irregular` and `regular`

### Requirement: Search from the sidebar
Clicking a sidebar item SHALL replace the search with that item's search. Ctrl-click (Cmd-click on macOS) SHALL add it to the current search with AND, and Alt-click (Option-click) SHALL add it negated. On touch devices, a long press SHALL offer these choices.

#### Scenario: Click a deck
- **WHEN** the user clicks the deck `Spanish` in the sidebar
- **THEN** the search becomes `deck:Spanish` and the matching cards are listed

#### Scenario: Combine
- **WHEN** the search is `deck:Spanish` and the user Alt-clicks the tag `verb`
- **THEN** the search becomes `deck:Spanish -tag:verb`

### Requirement: Manage decks from the sidebar
The user SHALL be able to rename and delete decks from the sidebar and to move a deck under another deck by dragging it, with the same rules and confirmations as on the deck list.

#### Scenario: Drag to reparent
- **WHEN** the user drags deck `Verbs` onto `Spanish`
- **THEN** the deck is renamed to `Spanish::Verbs` and keeps its cards

### Requirement: Saved searches
The user SHALL be able to save the current search under a unique name, run a saved search from a list next to the search field, and rename or delete saved searches. A search with an invalid term SHALL NOT be saved. Saved searches SHALL be included in backups.

#### Scenario: Save and run
- **WHEN** the user saves `deck:Spanish tag:leech` as `Spanish leeches` and later chooses it
- **THEN** the search field contains `deck:Spanish tag:leech` and the matching cards are listed

#### Scenario: Invalid search
- **WHEN** the user tries to save `is:bogus`
- **THEN** saving is refused with the search error

### Requirement: Default browser search
Settings SHALL include a default browser search, empty by default. When the browser opens, it SHALL start with that search.

#### Scenario: Start filtered
- **WHEN** the default search is `is:due` and the user opens Browse
- **THEN** the search field contains `is:due` and only due cards are listed

### Requirement: Find duplicates
The browser SHALL find sets of notes whose chosen field has the same content, comparing with HTML stripped, case ignored and surrounding whitespace ignored, among notes of note types that have a field of that name, optionally limited by a search. Empty fields SHALL NOT be reported. Each set SHALL show its content and note count, and selecting it SHALL list those notes.

#### Scenario: Two duplicates
- **WHEN** two Basic notes have Front `perro` and `Perro ` and the user finds duplicates in Front
- **THEN** one set `perro` with 2 notes is listed

#### Scenario: Limited by search
- **WHEN** the same duplicates exist in decks `A` and `B` and the user limits the search to `deck:A`
- **THEN** only duplicates among notes in `A` are reported

### Requirement: Tag duplicates
From the duplicate results, the user SHALL be able to add the tag `duplicate` to every note in all listed sets.

#### Scenario: Tag for review
- **WHEN** the results contain 3 sets with 7 notes and the user chooses Tag duplicates
- **THEN** those 7 notes have the tag `duplicate`

### Requirement: Change note type action
The browser SHALL offer "Change note type" for a selection whose notes all share one note type, with a dialog to choose the target note type and map fields and card types, defaulting to equal names and then positions. For a selection with several note types, the action SHALL explain that the notes must share one note type.

#### Scenario: Mixed selection
- **WHEN** the selection contains Basic and Cloze notes and the user chooses "Change note type"
- **THEN** the app explains that all selected cards must use the same note type, and nothing changes

### Requirement: Mark action and marked rows
The browser SHALL offer "Toggle mark" for the selected cards' notes, and SHALL tint rows of marked notes purple.

#### Scenario: Bulk mark
- **WHEN** the user selects 3 cards of unmarked notes and chooses "Toggle mark"
- **THEN** the 3 notes have the tag `marked` and their rows are tinted purple

### Requirement: Selection commands
The browser SHALL offer Invert selection, which selects exactly the listed cards that were not selected, and Select notes, which adds every listed card of the notes of the selected cards to the selection.

#### Scenario: Invert
- **WHEN** 3 of 10 listed cards are selected and the user inverts the selection
- **THEN** the other 7 cards are selected

#### Scenario: Select notes
- **WHEN** the forward card of a reversed note is selected and both cards are listed, and the user chooses Select notes
- **THEN** both cards are selected

### Requirement: Resizable split
On wide screens, the user SHALL be able to drag the divider between the card list and the details, and the position SHALL be remembered.

#### Scenario: Remembered width
- **WHEN** the user drags the divider to make the list wider and reopens the browser
- **THEN** the list has the same width

### Requirement: Keyboard navigation
On platforms with a keyboard, Up and Down SHALL move the current row and show its details, Shift with Up or Down SHALL extend the selection, Home and End SHALL go to the first and last row, Ctrl+F (Cmd+F on macOS) SHALL focus the search field, Enter SHALL open the editor of the current card, and Delete SHALL delete the selected cards' notes after confirmation.

#### Scenario: Browse with arrows
- **WHEN** the list has focus on the first row and the user presses Down twice
- **THEN** the third row is current and its details are shown

### Requirement: Set due date action
The browser SHALL offer "Set due date" for the selected cards, with an input that accepts the set-due-date syntax, explains it, and remembers the last value.

#### Scenario: Bulk postpone
- **WHEN** the user selects 20 cards and sets the due date to `3`
- **THEN** all 20 cards are due in 3 days

### Requirement: Reset dialog
Resetting cards in the browser SHALL show the reset options with their defaults before applying the reset to the selected cards.

#### Scenario: Options shown
- **WHEN** the user selects 5 cards and chooses Reset
- **THEN** a dialog shows "Restore original position where possible" on and "Reset repetition and lapse counts" off, and nothing changes until the user confirms

### Requirement: Filtered deck from search
The browser SHALL offer "Create filtered deck" prefilled with the current search.

#### Scenario: From search
- **WHEN** the search is `tag:exam is:due` and the user chooses Create filtered deck
- **THEN** the filtered deck options open with the search `tag:exam is:due`

### Requirement: Filtered deck search
`deck:filtered` SHALL match cards that are currently in a filtered deck.

#### Scenario: Cards away from home
- **WHEN** 50 cards are in filtered decks and the user searches `deck:filtered`
- **THEN** those 50 cards are listed

### Requirement: Mask editor
When adding or editing an Image Occlusion note, in the note editor or in the browser's card details, the editor SHALL let the user choose an image, draw rectangles, ellipses, polygons and text labels on it, select, move, resize and delete shapes, group and ungroup them, choose between "Hide all, guess one" and "Hide one, guess one", and fill the other fields. It SHALL work with mouse and touch. Shapes of an existing note SHALL keep their numbers when the note is edited. A note without an image or without shapes SHALL NOT be saved.

#### Scenario: Create from an image
- **WHEN** the user picks a diagram, draws two rectangles and adds the note
- **THEN** a note with two cards is created, and its image is stored in the collection's media

#### Scenario: Edit masks
- **WHEN** the user opens an existing Image Occlusion note in the editor, deletes one shape and saves
- **THEN** the card for that shape is kept as an empty card until the user removes it with Empty Cards, and the other cards keep their scheduling

#### Scenario: Numbers not reused
- **WHEN** the user later adds a shape to that note
- **THEN** it gets a number no card of the note has, so it does not take over the empty card's scheduling

#### Scenario: Edit masks in the browser
- **WHEN** the user selects a card of an Image Occlusion note with shapes 1 and 2 in the browser, draws a rectangle, and saves
- **THEN** the note gets a third card for the new shape, shapes 1 and 2 keep their numbers, and the list and preview show the change

#### Scenario: Change image in the browser
- **WHEN** the user chooses Change image in the browser's editor and picks another picture
- **THEN** the mask editor shows the new picture with the shapes where they were, and saving stores it in the media and in the Image field

#### Scenario: No shapes left
- **WHEN** the user deletes every shape of an Image Occlusion note in the browser and saves
- **THEN** the note is not saved and the editor asks the user to draw shapes

### Requirement: Browser appearance templates
A card type SHALL optionally have a browser question template and a browser answer template. When set, the browser list SHALL show the question text from the browser template, rendered and with HTML stripped. The card details and study SHALL keep using the normal templates.

#### Scenario: Simplified question
- **WHEN** a card type's front is `<div class=hint>{{Hint}}</div>{{Word}}` and its browser question template is `{{Word}}`
- **THEN** the browser row shows only the word, and studying shows the hint too

### Requirement: Sort by sort field
The browser SHALL offer a "Sort field" sort order, which sorts cards by their note's sort field with HTML stripped, case-insensitively.

#### Scenario: Sort by second field
- **WHEN** a note type's sort field is Meaning and the user sorts by "Sort field"
- **THEN** its cards are listed in order of their Meaning field

### Requirement: Fields excluded from unqualified search
Free-text search terms SHALL NOT match text in fields marked "exclude from unqualified searches".

#### Scenario: Notes field excluded
- **WHEN** the Notes field is excluded and contains `irregular`, and the user searches `irregular`
- **THEN** that card is not listed unless another field contains `irregular`

### Requirement: Per-field editor presentation
The editor SHALL show each field in its editing font and size, lay out right-to-left fields right to left, and start fields marked "collapse by default" collapsed, expanding them when tapped.

#### Scenario: Right-to-left field
- **WHEN** the Arabic field is marked right-to-left and the user types in it
- **THEN** the text is aligned right and the cursor moves right to left
