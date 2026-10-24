# collection Specification

## Purpose

Defines the user's flash card collection: decks, note types, notes, the cards generated from them, tags, and the operations that create, change, suspend and delete them.

## Requirements

### Requirement: Hierarchical decks
The system SHALL let the user create, rename, move and delete decks. A deck name containing `::` denotes nesting (for example `Languages::Spanish`), and parent decks that do not exist SHALL be created automatically. Deck names SHALL be unique among siblings, compared case-insensitively.

#### Scenario: Create nested deck
- **WHEN** the user creates a deck named `Languages::Spanish` and no deck `Languages` exists
- **THEN** the collection contains a deck `Languages` with a child deck `Spanish`

#### Scenario: Duplicate sibling name rejected
- **WHEN** the user renames a deck to a name that a sibling already uses, ignoring case
- **THEN** the rename is rejected with an error explaining that the name is taken, and nothing changes

### Requirement: Default deck
The collection SHALL always contain at least one deck. A new collection SHALL start with a deck named `Default`. The last remaining deck cannot be deleted.

#### Scenario: Fresh collection
- **WHEN** the app opens a collection for the first time
- **THEN** the deck list shows exactly one deck named `Default`

#### Scenario: Deleting the last deck
- **WHEN** the user tries to delete the only deck
- **THEN** the deletion is rejected and the deck remains

### Requirement: Deck deletion
Deleting a deck SHALL delete its child decks and all cards in them. Notes left with no cards SHALL be deleted. The user SHALL confirm a deletion that removes one or more cards, and the confirmation SHALL state how many cards will be deleted.

#### Scenario: Delete deck with cards
- **WHEN** the user deletes a deck that, together with its children, contains 12 cards
- **THEN** the app asks for confirmation mentioning 12 cards, and after confirmation those decks and cards are gone

### Requirement: Deck options
Deck options SHALL be kept in named presets. Each deck SHALL use exactly one preset, and decks using the same preset SHALL share its options. A preset SHALL include new cards per day (default 20), maximum reviews per day (default 200), desired retention (default 0.90, allowed range 0.70–0.99), and every other deck option. A child deck SHALL use its own preset independently of its parent. A new collection SHALL have the preset `Default`, used by every deck.

#### Scenario: Out-of-range retention
- **WHEN** the user sets desired retention to 0.5
- **THEN** the value is rejected and the previous value is kept

#### Scenario: Shared preset
- **WHEN** decks `Spanish` and `French` use preset `Languages` and the user sets its desired retention to 0.85 from `Spanish`
- **THEN** both decks use 0.85

#### Scenario: Existing options kept
- **WHEN** a collection whose deck `Biology` had a new card limit of 5 is opened by a version with presets
- **THEN** `Biology` still has a new card limit of 5 and its other options are unchanged

### Requirement: Audio deck options
Each deck SHALL have the options "Don't play audio automatically" and "Skip question when replaying answer", both off by default.

#### Scenario: Options persist
- **WHEN** the user turns on "Don't play audio automatically" for a deck and restarts the app
- **THEN** the option is still on for that deck and is included in backups

### Requirement: Step deck options
Each deck SHALL have the options learning steps and relearning steps, written as space-separated durations with units `s`, `m`, `h` or `d` (defaults `1m 10m` and `10m`), interday learning/review order (mix with reviews, before reviews, after reviews; default mix), and bury interday learning siblings (off by default). Steps that are not positive, longer than 365 days, or more than 20 SHALL be rejected with an explanation.

#### Scenario: Valid steps
- **WHEN** the user sets learning steps to `30s 5m 1h 2d`
- **THEN** the steps are saved and shown the same way when the options are reopened

#### Scenario: Invalid step
- **WHEN** the user enters `5x`
- **THEN** the value is rejected with an explanation and the previous steps are kept

### Requirement: Burying options
Each deck SHALL have the options bury new siblings and bury review siblings, both off by default, as in Anki. Collections created before these defaults SHALL have both options turned off once in their Default preset; presets the user added and presets imported from Anki SHALL keep their values.

#### Scenario: Defaults
- **WHEN** a deck is created
- **THEN** both burying options are off

#### Scenario: New preset
- **WHEN** the user adds a preset that starts with the default options
- **THEN** both burying options of the preset are off

#### Scenario: Existing collection
- **WHEN** a collection whose Default preset buries new and review siblings, and whose preset `Exam` (added by the user) buries them too, is opened after the update
- **THEN** the Default preset buries neither, and `Exam` still buries both

#### Scenario: Older backup
- **WHEN** the user restores a backup made before these defaults
- **THEN** its Default preset buries neither new nor review siblings, and its other presets keep their burying options

### Requirement: Built-in note types
A new collection SHALL contain the note types `Basic` (fields Front, Back; one card), `Basic (and reversed card)` (fields Front, Back; two cards), `Cloze` (fields Text, Back Extra; one card per cloze number) and `Image Occlusion` (fields Occlusion, Image, Header, Back Extra, Comments; one card per shape number). A collection created before Image Occlusion was built in SHALL get the note type when the user first adds an Image Occlusion note; when its name is taken by a different note type, a numbered name SHALL be used.

#### Scenario: Built-ins available
- **WHEN** the user opens the note type picker in a fresh collection
- **THEN** Basic, Basic (and reversed card), Cloze and Image Occlusion are listed

#### Scenario: Older collection
- **WHEN** the user adds the first Image Occlusion note in a collection without that note type, which has a Basic note type named `Image Occlusion`
- **THEN** a note type `Image Occlusion 2` is created for it, and later Image Occlusion notes use it too

### Requirement: Note types
A note type SHALL have a name, an ordered list of uniquely named fields, one or more card templates (each a name, a question format and an answer format), shared styling (CSS), and a kind, either standard or cloze. Cloze note types SHALL have exactly one template. The user SHALL be able to rename a note type and its fields, and to add or remove fields and templates.

#### Scenario: Remove a field
- **WHEN** the user removes field `Back` from a note type with existing notes
- **THEN** after confirmation the field and its content are removed from all notes of that type

### Requirement: Notes generate cards
A note SHALL hold one value per field of its note type, plus tags. For standard note types, the system SHALL create a card for each template whose question renders non-empty and that has no card yet. For cloze note types, it SHALL create a card for each cloze number in the note's fields that has no card yet. Cards whose question becomes empty, or whose cloze number disappears, after an edit SHALL be kept as empty cards with their scheduling until the user removes them with Empty Cards.

#### Scenario: Reversed card
- **WHEN** the user adds a `Basic (and reversed card)` note with Front `hola` and Back `hello`
- **THEN** two cards are created, one asking `hola` and one asking `hello`

#### Scenario: Cloze cards
- **WHEN** the user adds a Cloze note with Text `{{c1::Paris}} is the capital of {{c2::France}}`
- **THEN** two cards are created, for cloze 1 and cloze 2

#### Scenario: Editing removes a cloze
- **WHEN** the user edits that note and removes `{{c2::France}}`
- **THEN** the card for cloze 2 is kept as an empty card, and the card for cloze 1 keeps its scheduling state

#### Scenario: Template typo
- **WHEN** the user mistakenly changes a card type's front template from `{{Front}}` to `{{Back}}` in a note type with 1,000 notes whose Back is empty
- **THEN** no card is deleted, and the 1,000 cards are empty cards

### Requirement: Multi-number clozes generate cards
For cloze note types, every number listed in a multi-number deletion SHALL count as a cloze number of the note when cards are generated.

#### Scenario: Two cards from one deletion
- **WHEN** the user adds a Cloze note with Text `{{c1,2::Paris}} is the capital of France`
- **THEN** two cards are created, for cloze 1 and cloze 2

#### Scenario: Removing a number
- **WHEN** that note is edited to `{{c1::Paris}} is the capital of France`
- **THEN** the card for cloze 2 is kept as an empty card, and the card for cloze 1 keeps its scheduling state

### Requirement: Empty note rejected
The system SHALL reject adding a note that would generate no cards, and SHALL explain why.

#### Scenario: Empty front
- **WHEN** the user tries to add a Basic note with an empty Front field
- **THEN** the note is not added and the editor reports that the note would produce no cards

### Requirement: Tags
Notes SHALL carry zero or more tags. Tags are compared case-insensitively, contain no whitespace, and may use `::` for hierarchy. The system SHALL be able to list all tags in use.

#### Scenario: Add tag with spaces
- **WHEN** the user enters the tag text `verb irregular`
- **THEN** the note gets two tags, `verb` and `irregular`

### Requirement: Suspend and bury
The user SHALL be able to suspend and unsuspend cards. Suspended cards are never shown for study. The user SHALL be able to bury a card until the next day.

#### Scenario: Suspended card skipped
- **WHEN** a due card is suspended
- **THEN** it does not appear in any study session or due count until it is unsuspended

#### Scenario: Buried card returns
- **WHEN** a card is buried during a study session
- **THEN** it is excluded from study for the rest of the current day and is eligible again after the next day rollover

### Requirement: Move cards between decks
The user SHALL be able to move one or more cards to another deck without changing their scheduling state.

#### Scenario: Move preserves schedule
- **WHEN** a review card due in 10 days is moved to another deck
- **THEN** it is still due in 10 days, now in the target deck

### Requirement: Change notifications
Whenever the collection changes, including changes made by import, restore or study, every open screen that shows affected data SHALL refresh without the user taking action.

#### Scenario: Counts update after import
- **WHEN** an import finishes while the deck list is visible
- **THEN** the deck list shows the new decks and updated due counts

### Requirement: Empty Cards tool
The user SHALL be able to list all empty cards, grouped by note with their card type or cloze number and a total count, and delete them after confirmation in one operation. For a note whose cards are all empty, one card SHALL be kept so that no note is deleted, and the list SHALL say so.

#### Scenario: Remove empty cards
- **WHEN** 3 notes each have one empty card among other cards and the user confirms Empty Cards
- **THEN** those 3 cards are deleted and every note still has its other cards

#### Scenario: All cards empty
- **WHEN** a note's only card is empty
- **THEN** the list marks the note, and deleting empty cards keeps that card

### Requirement: Undo changes
Every change to the collection made through the app SHALL be undoable, as a named step, in reverse order, for at least the last 30 steps of the current app session. Undoing SHALL restore exactly the previous state of everything the step changed. The undo history SHALL be cleared when the app closes, when a backup is restored, and when a new collection is started. Changes received by device sync SHALL NOT be undo steps; when they change or delete anything that a step in the undo or redo history changed, the history SHALL be cleared. Otherwise it SHALL be kept. Undoing a step that was already synced SHALL be synced as a new change.

#### Scenario: Undo a delete
- **WHEN** the user deletes 3 notes in the browser and chooses Undo
- **THEN** the 3 notes and their cards are back with their scheduling, review history and images

#### Scenario: Undo an import
- **WHEN** the user imports a package that adds 200 notes and chooses Undo
- **THEN** the 200 notes, their cards and the decks the import created are gone

#### Scenario: History cleared by restore
- **WHEN** the user restores a backup
- **THEN** there is nothing to undo

#### Scenario: Unrelated synced change
- **WHEN** the user edits note A on the laptop, and a sync then brings an edit of note B from the phone
- **THEN** Undo on the laptop still reverts the edit of note A, and note B keeps the phone's edit

#### Scenario: Synced change to the same note
- **WHEN** the user edits note A on the laptop, and a sync then brings a later edit of note A from the phone
- **THEN** there is nothing to undo on the laptop

#### Scenario: Undo after sync
- **WHEN** the user deletes a note on the laptop, the deletion syncs to the phone, and the user then chooses Undo on the laptop
- **THEN** the note is back on the laptop, and after the next sync it is back on the phone too

### Requirement: Redo
After an undo, and until another change is made, the user SHALL be able to redo the undone step.

#### Scenario: Redo a move
- **WHEN** the user moves cards to `Archive`, undoes it, and chooses Redo
- **THEN** the cards are in `Archive` again

### Requirement: Deletion log
Whenever notes are deleted, for any reason, each note's deletion time, ID, note type name, fields and tags SHALL be added to a deletion log stored with the collection and included in backups.

#### Scenario: Deleted with a deck
- **WHEN** the user deletes a deck containing 5 notes
- **THEN** the deletion log has 5 new entries with those notes' fields and tags

### Requirement: Export and clear the deletion log
The user SHALL be able to export the deletion log as a tab-separated text file that the text import can read, and to clear the log after a confirmation.

#### Scenario: Recover a note
- **WHEN** the user exports the deletion log and imports the file with the deleted note's note type
- **THEN** the deleted note's fields and tags are back in a new note

### Requirement: Card flags
Each card SHALL have either no flag or exactly one of seven flags: 1 red, 2 orange, 3 green, 4 blue, 5 pink, 6 turquoise, 7 purple. Flags SHALL belong to the card, not to its note. Setting or clearing a flag SHALL NOT change the card's scheduling. Flags SHALL be included in backups.

#### Scenario: Flag one card of a note
- **WHEN** the user flags the forward card of a `Basic (and reversed card)` note red
- **THEN** the forward card has flag 1 and the reverse card has no flag

#### Scenario: Flag kept across restore
- **WHEN** a collection with flagged cards is exported to a backup and restored
- **THEN** every card has the same flag as before

### Requirement: Flag names
The user SHALL be able to rename each of the seven flags in Settings. A flag without a custom name SHALL be shown by its colour name. Clearing a custom name SHALL restore the colour name.

#### Scenario: Rename a flag
- **WHEN** the user renames flag 1 to "Fix later"
- **THEN** the study menu, the browser's flag action and the settings show "Fix later" for the red flag

### Requirement: Create note types
The user SHALL be able to create a note type from a stock type, giving it a name. The stock types SHALL be Basic, Basic (and reversed card), Basic (optional reversed card), Basic (type in the answer), Cloze and Image Occlusion. Note type names SHALL be unique, compared case-insensitively.

#### Scenario: Add from stock
- **WHEN** the user adds a note type from stock type Basic named `Vocabulary`
- **THEN** a note type `Vocabulary` with fields Front and Back and one card type exists, and it has no notes

#### Scenario: Name taken
- **WHEN** the user adds a note type named `basic` while `Basic` exists
- **THEN** the creation is rejected with an error that the name is taken

#### Scenario: Image Occlusion from stock
- **WHEN** the user adds a note type from stock type Image Occlusion named `Anatomy`
- **THEN** `Anatomy` has the fields Occlusion, Image, Header, Back Extra and Comments, and its notes are edited with the mask editor

### Requirement: Clone note types
The user SHALL be able to clone a note type, which SHALL create a new note type with a new name and the same kind, fields, card types and styling, without notes.

#### Scenario: Clone imported type
- **WHEN** the user clones an imported note type with 4 fields and 2 card types as `Experiment`
- **THEN** `Experiment` has the same 4 fields, 2 card types and CSS, and no notes

### Requirement: Delete note types
The user SHALL be able to delete a note type, which SHALL delete all its notes and their cards. The app SHALL ask for confirmation stating how many notes and cards will be deleted. The last remaining note type SHALL NOT be deletable.

#### Scenario: Delete with notes
- **WHEN** the user deletes a note type that has 10 notes with 20 cards and confirms
- **THEN** the note type, its 10 notes and 20 cards are gone

#### Scenario: Last note type
- **WHEN** only one note type exists and the user tries to delete it
- **THEN** the deletion is rejected and the note type remains

### Requirement: Optional reversed card
The stock type Basic (optional reversed card) SHALL have the fields Front, Back and Add Reverse, and SHALL generate the reverse card only when Add Reverse is not empty.

#### Scenario: Reverse on demand
- **WHEN** the user adds a note of that type with Front `hola`, Back `hello` and Add Reverse `y`
- **THEN** two cards are created, and a note with an empty Add Reverse gets only the forward card

### Requirement: Template preview
While editing a card type's front or back template or the note type's CSS, the app SHALL show the rendered front and back using the unsaved text, updating within a second of the last change. The preview SHALL use a note of that type chosen by the user, or placeholder field text `(FieldName)` when the note type has no notes. Previewing SHALL NOT change the collection.

#### Scenario: Live update
- **WHEN** the user changes the front template from `{{Front}}` to `<b>{{Front}}</b>` without saving
- **THEN** the preview shows the front field in bold, and cards in the collection are unchanged

#### Scenario: No notes yet
- **WHEN** the user edits a template of a note type with no notes
- **THEN** the preview shows `(Front)` where `{{Front}}` is used

#### Scenario: Cloze number
- **WHEN** the user previews a Cloze note with deletions c1 and c2 and chooses card 2
- **THEN** the preview shows the question of cloze card 2

### Requirement: Template checks on save
Saving a card type SHALL be refused, with a message naming the problem, when a template has a syntax error (an unclosed, unmatched or crossed section, a `{{` without `}}`, or a tag without a field name), refers to a field that is neither the note type's nor special, or when the front uses no field. A front identical to another card type's SHALL be refused in standard note types, and a front without a `cloze:` filter in cloze note types. A refused save SHALL change nothing.

#### Scenario: Unknown field
- **WHEN** the user saves a front template `{{Frnt}}` in a note type with fields Front and Back
- **THEN** the save is refused with a message that there is no field `Frnt`

#### Scenario: Unclosed section
- **WHEN** the user saves `{{#Hint}}{{Hint}}` without `{{/Hint}}`
- **THEN** the save is refused with a message that `{{#Hint}}` is not closed

#### Scenario: Unclosed tag
- **WHEN** the user saves a back template `{{FrontSide}}<hr id=answer>{{Back`
- **THEN** the save is refused with a message that `{{Back` has no closing `}}`

#### Scenario: Tag without a field name
- **WHEN** the user saves a front template `{{Front}}{{}}`
- **THEN** the save is refused with a message that `{{}}` has no field name

#### Scenario: Crossed sections
- **WHEN** the user saves `{{#Hint}}{{#Extra}}{{/Hint}}{{/Extra}}`
- **THEN** the save is refused with one message that `{{/Hint}}` closes `{{#Hint}}` before `{{#Extra}}` is closed

#### Scenario: Missing cloze filter
- **WHEN** the user saves the front template `{{Text}}` in a Cloze note type
- **THEN** the save is refused with a message that the front must use `{{cloze:Text}}` or another cloze field

#### Scenario: Duplicate front
- **WHEN** a standard note type has two card types and the user makes the second front identical to the first
- **THEN** the save is refused with a message that both card types would show the same question

### Requirement: Live template feedback
The template editor SHALL show the result of the save checks while the user edits, before saving.

#### Scenario: Error shown while typing
- **WHEN** the user types `{{Frnt}}` into the front template
- **THEN** the editor shows that there is no field `Frnt` without the user saving

### Requirement: Reserved field names
Adding a field or renaming a field to `Tags`, `Type`, `Deck`, `Card` or `FrontSide`, compared case-insensitively, SHALL be refused with an explanation.

#### Scenario: Reserved name
- **WHEN** the user adds a field named `tags`
- **THEN** the field is not added and the message says the name is reserved

### Requirement: Leech options
Each deck SHALL have a leech threshold in lapses (default 8, allowed 1 to 99) and a leech action: tag only (default), or suspend card and tag.

#### Scenario: Options stored
- **WHEN** the user sets the threshold to 5 and the action to suspend
- **THEN** the values persist for the deck and are included in backups

### Requirement: Today-only limits
A deck's new and review limits SHALL be overridable for the current study day only. The override SHALL be ignored from the next rollover on, and SHALL be included in backups.

#### Scenario: Override expires
- **WHEN** the new card limit of `Spanish` is 20 with a today-only override of 30, and the next study day starts
- **THEN** the effective new card limit is 20

### Requirement: Limit behaviour options
Each deck SHALL have the options new cards ignore review limit (on by default) and limits start from top (off by default).

#### Scenario: Defaults
- **WHEN** a deck is created
- **THEN** new cards ignore the review limit and limits do not start from the top

### Requirement: Interval limit options
Each deck SHALL have the options maximum interval (default 36,500 days) and minimum interval after relearning (default 1 day), each allowed from 1 to 36,500 days, with the minimum not above the maximum.

#### Scenario: Minimum above maximum
- **WHEN** the user sets the maximum interval to 30 and the minimum interval after relearning to 60
- **THEN** the change is rejected with an explanation

### Requirement: New card order options
Each deck SHALL have the options new card gather order (deck, deck then random notes, ascending position, descending position, random notes, random cards; default ascending position) and new card sort order (order gathered, card type then order gathered, card type then random, random note then card type, random; default order gathered).

#### Scenario: New card defaults
- **WHEN** a new deck is created
- **THEN** its new card gather order is ascending position and its sort order is order gathered

### Requirement: Review order options
Each deck SHALL have the options new/review order (mix, before reviews, after reviews; default mix) and review sort order (due date then random, due date then deck, deck then due date, ascending intervals, descending intervals, ascending difficulty, descending difficulty, ascending retrievability, relative overdueness; default due date then random).

#### Scenario: Review defaults
- **WHEN** a new deck is created
- **THEN** its new/review order is mix and its review sort order is due date then random

### Requirement: Reposition new cards
The user SHALL be able to give selected New cards consecutive positions from a start position with a step, optionally in random order, and optionally shifting the positions of other new cards at or after the start so no positions are shared. Cards that are not New SHALL be left unchanged.

#### Scenario: Move a chapter to the front
- **WHEN** the user repositions 20 new cards with start 0, step 1 and shifting on
- **THEN** those 20 cards are introduced before all other new cards, in their previous relative order

#### Scenario: Mixed selection
- **WHEN** the selection contains 5 new and 3 review cards
- **THEN** only the 5 new cards are repositioned and the app says 3 cards were not new

### Requirement: Insertion order
Each deck SHALL have an insertion order option, sequential (default) or random. With random, cards added to the deck SHALL get random positions among its new cards. Changing the option SHALL reposition the deck's existing new cards: shuffled for random, in order added for sequential.

#### Scenario: Random insertion
- **WHEN** a deck's insertion order is random and 100 cards are imported into it
- **THEN** the new cards are not introduced in the import order

#### Scenario: Back to sequential
- **WHEN** the option is changed back to sequential
- **THEN** the deck's new cards are introduced in the order they were added

### Requirement: Manage presets
The user SHALL be able to add, clone, rename and delete presets, choose a deck's preset, and assign a deck's preset to all its subdecks. Deleting a preset SHALL move the decks using it to `Default`; `Default` SHALL NOT be deletable. Preset names SHALL be unique, compared case-insensitively. The options screen SHALL show how many decks use the preset.

#### Scenario: Save to all subdecks
- **WHEN** the user chooses "Save to all subdecks" on `Languages`, which uses preset `Languages`
- **THEN** every descendant of `Languages` uses preset `Languages`

#### Scenario: Delete a preset
- **WHEN** the user deletes preset `Exam`, used by 3 decks
- **THEN** those 3 decks use `Default`

### Requirement: This-deck limits
A deck SHALL optionally have its own new and review limits that override its preset's limits for that deck only.

#### Scenario: One deck lower
- **WHEN** `Spanish` and `French` use a preset with 20 new cards per day and `French` has a this-deck new limit of 5
- **THEN** `French` shows at most 5 new cards and `Spanish` up to 20

### Requirement: Find and replace in notes
The system SHALL replace text in a set of notes, in all fields, in one named field, or in tags. Matching SHALL be literal unless regular expressions are requested, in which case the replacement MAY refer to capture groups. Matching SHALL ignore case unless the user turns that off. All changes SHALL be applied as one operation, and cards SHALL be regenerated as for an edit.

#### Scenario: Fix a spelling
- **WHEN** the user replaces `recieve` with `receive` in all fields of 200 notes, 15 of which contain it
- **THEN** those 15 notes contain `receive` and the report says 15 notes changed

#### Scenario: Regex with group
- **WHEN** the user replaces `(\d+) kg` with `$1 kilograms` using regular expressions in the field Back
- **THEN** `5 kg` in Back becomes `5 kilograms`, and other fields are unchanged

#### Scenario: Rename a tag pattern
- **WHEN** the user replaces `ch0` with `chapter` in tags
- **THEN** the tag `ch01` becomes `chapter1`

#### Scenario: Invalid expression
- **WHEN** the user enters `(` with regular expressions on
- **THEN** the replacement is refused with an error and nothing changes

### Requirement: Rename tags
The user SHALL be able to rename a tag on all notes that have it, including its child tags. Renaming into an existing tag SHALL merge them. Moving a tag under another tag SHALL be a rename to `parent::tag`.

#### Scenario: Rename with children
- **WHEN** the user renames `verb` to `verbs`
- **THEN** notes tagged `verb::irregular` are tagged `verbs::irregular`

#### Scenario: Merge
- **WHEN** the user renames `noun` to `nouns` and some notes already have `nouns`
- **THEN** those notes have `nouns` once and no note has `noun`

### Requirement: Delete tags
The user SHALL be able to delete a tag and its child tags from all notes, after a confirmation stating how many notes are affected. Notes SHALL NOT be deleted.

#### Scenario: Delete a tag
- **WHEN** the user deletes the tag `old` used by 12 notes and confirms
- **THEN** no note has `old`, and the 12 notes still exist

### Requirement: Clear unused tags
The user SHALL be able to clear tags that no note uses from every tag list.

#### Scenario: Tag left behind
- **WHEN** the tag `draft` remains in the tag list but no note uses it, and the user clears unused tags
- **THEN** `draft` no longer appears in the sidebar or the tag suggestions

### Requirement: Change note type
The user SHALL be able to change notes that share one note type to another note type, mapping each target field to one source field or to nothing, and, between standard note types, each target card type to one source card type or to nothing. Cards whose card type is mapped SHALL keep their scheduling and review history. Unmapped source cards SHALL be deleted, and cards the target generates that have no mapped source SHALL be created as New.

#### Scenario: Basic to reversed
- **WHEN** 10 Basic notes are changed to `Basic (and reversed card)` with Front→Front, Back→Back and Card 1→Card 1
- **THEN** the 10 forward cards keep their scheduling and 10 new reverse cards are created as New

#### Scenario: Unmapped field
- **WHEN** a note with fields Front `a`, Back `b`, Extra `c` is changed to Basic with Extra mapped to nothing
- **THEN** the note has Front `a` and Back `b`, and `c` is gone

#### Scenario: Cloze numbers kept
- **WHEN** a Cloze note is changed to another cloze note type with Text→Text
- **THEN** its cards for each cloze number keep their scheduling

### Requirement: Change preview and confirmation
Before notes change type, the app SHALL show how many notes change, how many cards will be deleted and created, and which source fields will lose their content, and SHALL apply the change only after confirmation, as one operation.

#### Scenario: Confirmation counts
- **WHEN** 5 `Basic (and reversed card)` notes are changed to Basic with Card 2 mapped to nothing
- **THEN** the confirmation says 5 notes change and 5 cards will be deleted

### Requirement: Marked notes
Toggling the mark on notes SHALL add the tag `marked` to all of them when any lacks it, and otherwise remove it from all of them. The tag SHALL behave as any other tag.

#### Scenario: Mark a note
- **WHEN** the user toggles the mark on a note without the `marked` tag
- **THEN** the note has the tag `marked`

### Requirement: Deck descriptions
Each deck SHALL have an optional description that the user can edit, which SHALL be shown on the deck's overview as sanitized HTML, without running scripts.

#### Scenario: Edit description
- **WHEN** the user sets the description of `Spanish` to `Vocabulary from <b>Book 1</b>`
- **THEN** the overview shows the text with "Book 1" in bold

### Requirement: Unbury
The user SHALL be able to unbury a deck, which SHALL make every buried card in the deck and its subdecks available for study immediately, including cards held back today by sibling burying.

#### Scenario: Unbury manually buried cards
- **WHEN** the user buried 3 due cards of `Spanish` and then chooses Unbury on its overview
- **THEN** the 3 cards are due again and the overview's buried count is 0

#### Scenario: Unbury siblings
- **WHEN** the reverse card of a studied note is held back today by sibling burying and the user chooses Unbury
- **THEN** the reverse card is available in today's session

### Requirement: Reset options
Resetting cards to New SHALL offer to restore each card's original position among new cards, where it is known (on by default), and to reset its review and lapse counts to zero (off by default). Without restoring, reset cards SHALL be placed after all other new cards. Every reset SHALL add a manual entry to the card's review history.

#### Scenario: Restore position
- **WHEN** a card that was the 3rd new card when first studied is reset with "restore original position"
- **THEN** it is introduced before the new cards that were added after it

#### Scenario: Reset counts
- **WHEN** a card with 12 reviews and 4 lapses is reset with "reset repetition and lapse counts"
- **THEN** it is New with 0 reviews and 0 lapses, and its earlier review entries are still in its history

### Requirement: Original position
A card SHALL remember its new-card position when it is first answered, and keep it across later reviews, moves and backups.

#### Scenario: Position kept
- **WHEN** a new card at position 42 is answered and later reviewed many times
- **THEN** its remembered original position is still 42

### Requirement: Timer deck options
Each deck SHALL have the options maximum answer seconds (default 60, allowed 1 to 9999), show answer timer (off by default) and stop timer on answer (on by default).

#### Scenario: Out of range
- **WHEN** the user sets maximum answer seconds to 0
- **THEN** the value is rejected and the previous value is kept

### Requirement: Auto advance deck options
Each deck SHALL have the options seconds to show question for and seconds to show answer for (0 to 9999, default 0 meaning disabled), the question action (show answer, or show reminder; default show answer) and the answer action (bury card, answer Again, answer Hard, answer Good, or show reminder; default bury card).

#### Scenario: Options stored
- **WHEN** the user sets 5 seconds for the question and 3 for the answer with the answer action "answer Good"
- **THEN** those values persist for the deck and are included in backups

### Requirement: Deleting filtered decks
Deleting a filtered deck SHALL return its cards to their home decks and SHALL NOT delete any card. Deleting a normal deck SHALL also delete its cards that are currently in filtered decks, and the confirmation SHALL count them.

#### Scenario: Delete filtered deck
- **WHEN** the user deletes a filtered deck holding 20 cards
- **THEN** no card is deleted and the 20 cards are in their home decks

### Requirement: FSRS parameter options
Each deck SHALL have FSRS parameters (empty meaning default parameters), a search that selects the reviews used for optimizing (default: the deck and its subdecks), an optional date before which reviewed cards are ignored, and a historical retention (default 0.90, allowed 0.50 to 0.99).

#### Scenario: Default parameters
- **WHEN** a deck has never been optimized
- **THEN** its parameters are empty and the default parameters are used

### Requirement: Reschedule on change
Each deck SHALL have the option "Reschedule cards on change", off by default. When it is on and the deck's desired retention or parameters change, the due dates of its review cards SHALL be recomputed from their review history with the new values, after a confirmation stating how many cards will be rescheduled.

#### Scenario: Lower retention reschedules
- **WHEN** the option is on and the user lowers desired retention from 0.90 to 0.85 and confirms
- **THEN** the deck's review cards get later due dates

#### Scenario: Option off
- **WHEN** the option is off and the user changes desired retention
- **THEN** no due date changes until cards are answered

### Requirement: Workload simulation
The user SHALL be able to simulate a deck's expected daily reviews and number of memorised cards over a chosen number of days, starting from its current cards, for given new and review limits and desired retention, and to compare two retentions. Simulating SHALL NOT change the collection and SHALL be reproducible for the same inputs and collection.

#### Scenario: Compare retentions
- **WHEN** the user simulates 365 days at retention 0.90 and compares with 0.95
- **THEN** a chart shows both curves, and the 0.95 curve has more reviews per day

### Requirement: Retention suggestion
The user SHALL be able to request a suggested desired retention, which SHALL be the retention between 0.70 and 0.99 with the lowest simulated review time per memorised card. Applying the suggestion SHALL set the deck's desired retention only after the user confirms.

#### Scenario: Help me decide
- **WHEN** the user chooses "Help me decide"
- **THEN** a suggested retention is shown with the cost curve, and the deck's retention is unchanged until the user applies it

### Requirement: Easy Days option
Each deck SHALL have an Easy Days setting with a level for each weekday: Normal (default), Reduced or Minimum.

#### Scenario: Weekend reduced
- **WHEN** the user sets Saturday and Sunday to Reduced
- **THEN** the setting persists for the deck and is included in backups

### Requirement: Image Occlusion note type
The app SHALL offer an Image Occlusion note type with the fields Occlusion, Image, Header, Back Extra and Comments. The Occlusion field SHALL store shapes as cloze deletions in Anki's image occlusion syntax, and one card SHALL be generated per shape number, so shapes sharing a number are asked together.

#### Scenario: Three labels
- **WHEN** the user creates an Image Occlusion note with three shapes numbered 1, 2 and 3
- **THEN** three cards are created

#### Scenario: Grouped shapes
- **WHEN** two of those shapes are grouped
- **THEN** the note has two cards, and the grouped shapes are asked together

### Requirement: Reposition card types
The user SHALL be able to move a card type of a standard note type to another position. Every existing card SHALL keep its scheduling and stay attached to the same card type.

#### Scenario: Swap two card types
- **WHEN** a note type has card types `Recognition` and `Production` and the user moves `Production` to the first position
- **THEN** `Production` is listed first, and every card keeps its card type, deck and scheduling

### Requirement: Card type deck override
A card type SHALL optionally name a deck. When a note is added or edited so that a new card of that card type is generated, that card SHALL be placed in the named deck instead of the note's deck. Existing cards SHALL NOT be moved. When the named deck no longer exists, the note's deck SHALL be used.

#### Scenario: Production cards to their own deck
- **WHEN** the card type `Production` overrides to deck `Spanish::Production` and the user adds a note in deck `Spanish`
- **THEN** the `Recognition` card is in `Spanish` and the `Production` card is in `Spanish::Production`

#### Scenario: Override deck deleted
- **WHEN** the override deck has been deleted and the user adds a note
- **THEN** all its cards are in the deck chosen in the editor

### Requirement: Import of card type options
Importing an `.apkg` SHALL keep each card type's deck override when that deck is imported or exists, and its browser appearance templates.

#### Scenario: Imported override
- **WHEN** a package's card type overrides to a deck that the package also contains
- **THEN** the imported card type overrides to that deck

### Requirement: Reposition fields
The user SHALL be able to move a field of a note type to another position. Every note of that type SHALL keep each value in the same named field, and cards SHALL render as before.

#### Scenario: Move Back first
- **WHEN** a Basic note type's field Back is moved before Front
- **THEN** the editor lists Back first, every note's Back value is unchanged, and cards show the same content

### Requirement: Field options
Each field SHALL have the options editing font, editing size, right-to-left, collapse by default, and exclude from unqualified searches. Each note type SHALL have exactly one sort field, the first field by default. Options SHALL follow the field when it is renamed or moved, and SHALL be included in backups and kept on `.apkg` import.

#### Scenario: Sort field follows move
- **WHEN** the sort field is Front and Front is moved to the second position
- **THEN** Front is still the sort field

#### Scenario: Imported options
- **WHEN** an `.apkg` whose note type has a right-to-left field and sorts by its second field is imported
- **THEN** the imported note type has the same right-to-left field and sort field

### Requirement: Unicode normalisation
Field contents, tags, deck names and descriptions, note type names, field names, card templates, note type CSS and media file names SHALL be stored in Unicode NFC form, whether they come from editing, renaming, adding media or importing. Search queries SHALL be normalised the same way. A collection created by an earlier version SHALL be normalised once when it is first opened; stored media files whose names are not in NFC SHALL be renamed to their NFC form, so that references in fields, templates and CSS keep finding them.

#### Scenario: Decomposed input
- **WHEN** the user adds a note whose Front is `e` followed by a combining acute accent
- **THEN** the stored Front is the single character `é`

#### Scenario: Search matches either form
- **WHEN** a note contains `café` and the user searches `café` typed in decomposed form
- **THEN** the note is found

#### Scenario: Decomposed media name
- **WHEN** a collection from an earlier version holds a media file named `café.jpg` in decomposed form, and a note shows it with `<img src="café.jpg">` written the same way
- **THEN** after the collection is first opened, the file is stored as `café.jpg` in NFC, the note's reference is in NFC, and the image still shows
