# app-shell Specification

## Purpose

The application frame around all features: which platforms Cardflip runs on, how the user navigates between screens, the layout across screen sizes, theming, and app-wide settings.

## Requirements

### Requirement: Supported platforms
Cardflip SHALL run with identical features on Linux, Windows, macOS, Android, iOS and the web (current Chrome, Firefox, Safari and Edge), with two exceptions: the MCP server for agents SHALL be available only on Linux, Windows and macOS, and sync with paired devices SHALL be available on every platform except the web.

#### Scenario: Same behavior on web
- **WHEN** the same backup is restored in the web build and in the Linux build
- **THEN** both show the same decks, due counts and card contents

#### Scenario: MCP server only on desktop
- **WHEN** the user opens Settings on Android
- **THEN** there is no "Agents (MCP)" section, while on Linux there is

#### Scenario: No device sync on the web
- **WHEN** the user opens Settings in the web build
- **THEN** there is no Sync section, while on Android and Linux there is

### Requirement: Main navigation
The app SHALL provide top-level destinations for Decks (home), Browse, Add and Settings. On wide screens (at least 840 logical pixels), navigation SHALL be a side rail. On narrow screens, it SHALL be a bottom navigation bar.

#### Scenario: Phone layout
- **WHEN** the app runs on a phone in portrait
- **THEN** navigation is a bottom bar with Decks, Browse, Add and Settings

### Requirement: Deck list home screen
The home screen SHALL show the deck tree with new, learning and review counts for each deck. Parent decks SHALL be collapsible. Each deck SHALL have actions to study, add a card, open its options, rename it and delete it.

#### Scenario: Due counts visible
- **WHEN** a deck has 5 new, 2 learning and 30 review cards due
- **THEN** its row shows 5, 2 and 30 in three distinguishable colours

### Requirement: Theming
The app SHALL support light, dark and system theme modes, defaulting to system. The choice SHALL persist.

#### Scenario: Theme persists
- **WHEN** the user selects dark mode and restarts the app
- **THEN** the app starts in dark mode

### Requirement: App settings
Settings SHALL include theme mode, the day rollover hour, whether to continue automatically when learning cards become due during a study session (off by default), export backup, restore backup, import, and an About section showing the app version.

#### Scenario: Change rollover hour
- **WHEN** the user changes the rollover hour from 4 to 6
- **THEN** due counts and daily limits use 06:00 as the day boundary from then on

#### Scenario: Auto-continue setting persists
- **WHEN** the user turns on "Continue automatically when learning cards are due" and restarts the app
- **THEN** the setting is still on, and it is included in backups

### Requirement: First run
On first launch, the app SHALL open on the deck list with the `Default` deck and a short empty-state hint offering to add a card or import a file.

#### Scenario: Empty state
- **WHEN** the app is launched for the first time
- **THEN** the deck list shows `Default` and buttons to add a card or import

### Requirement: Errors are shown, not swallowed
Any failed operation SHALL show the user a readable message. Unexpected errors SHALL NOT leave the UI in a loading state indefinitely.

#### Scenario: Backend error
- **WHEN** saving a note fails
- **THEN** a message explains the failure, and the editor keeps the user's input

### Requirement: Card terminology
On the deck list, Browse, add and edit screens and the study screen, user-facing text SHALL call what the user adds, edits, suspends and deletes a "card", including buttons, menu items, titles, tooltips, confirmations and editor error messages. "Note type" SHALL stay as the name of the template that decides which cards are made.

#### Scenario: Add and edit labels
- **WHEN** the user opens the add screen, and later chooses to edit a card while studying
- **THEN** the add screen's title and button say "Add card", and the study menu item and the editor's title say "Edit card"

#### Scenario: Editor error
- **WHEN** the user tries to add a Basic card with an empty Front field
- **THEN** the error says the entry would produce no cards and does not call it a note

#### Scenario: Note type kept
- **WHEN** the user opens the add screen
- **THEN** the template picker is labelled "Note type"

### Requirement: Due counts stay current over time
The deck list's due counts SHALL update without user action when they change because time passed: when a learning or relearning card becomes due, and when a new study day starts at the rollover hour. They SHALL also update when the app returns to the foreground. The update SHALL happen within a few seconds of the due time while the app is in the foreground.

#### Scenario: Learning card becomes due on the deck list
- **WHEN** the user rates a card Again with a 1-minute step, returns to the deck list, and waits one minute
- **THEN** that deck's learning count increases by one without the user tapping anything

#### Scenario: Day rollover
- **WHEN** the deck list is open across the rollover hour and review cards are due in the new study day
- **THEN** the review counts update to the new day's due cards

#### Scenario: Returning to the app
- **WHEN** the app was in the background while a learning card became due and the user brings it back to the foreground
- **THEN** the deck list shows the updated learning count

### Requirement: Learning due reminder
When learning cards become due while the app is in the foreground and the user is not in a study session, the app SHALL show an in-app reminder that names the deck and offers a Study action that opens that deck's study session. When cards become due in several top-level decks at once, the reminder SHALL say so and its action SHALL show the deck list. The app SHALL NOT show system notifications.

#### Scenario: Reminder on another tab
- **WHEN** the user is on the Browse tab and a learning card in deck "Spanish" becomes due
- **THEN** a reminder says learning cards are due in "Spanish" with a Study action

#### Scenario: Study from the reminder
- **WHEN** the user chooses Study on the reminder for "Spanish"
- **THEN** the app shows the Decks tab with a study session for "Spanish" open, starting with the due learning card

#### Scenario: No reminder while studying
- **WHEN** a learning card becomes due while a study session is open
- **THEN** no reminder is shown outside the study screen; the study screen handles it

#### Scenario: Several decks at once
- **WHEN** learning cards become due in the top-level decks "Spanish" and "Biology" at the same moment
- **THEN** one reminder says learning cards are due in 2 decks, and its action shows the deck list

### Requirement: Backup settings
Settings SHALL include the backup interval in minutes (default 30, 0 turns off periodic backups), the number of daily, weekly and monthly backups to keep (default 8 each), the list of automatic backups, and the storage space they use.

#### Scenario: Periodic backups off
- **WHEN** the interval is 0 and the user imports a collection package
- **THEN** a backup is still made before the import, and no periodic backups are made

### Requirement: Empty Cards entry
Settings SHALL offer Empty Cards, which shows the empty cards list and its delete action.

#### Scenario: Nothing to remove
- **WHEN** the user opens Empty Cards and no card is empty
- **THEN** the dialog says there are no empty cards

### Requirement: Undo controls
The app SHALL offer Undo and Redo that name the step they act on, through menu items and, outside text fields, Ctrl+Z, Ctrl+Shift+Z and Ctrl+Y (Cmd on macOS). After deleting, moving in bulk or importing, the app SHALL show a message with an Undo action.

#### Scenario: Undo from the snackbar
- **WHEN** the user moves 30 cards and taps Undo in the message that appears
- **THEN** the cards are back in their previous decks

#### Scenario: Named step
- **WHEN** the last change was deleting 3 notes
- **THEN** the Undo menu item reads "Undo Delete 3 notes"

### Requirement: Paste formatting setting
Settings SHALL include "Strip formatting when pasting", off by default. When on, a normal paste SHALL insert plain text and Ctrl+Shift+V (Cmd+Shift+V on macOS) SHALL keep formatting.

#### Scenario: Setting inverts paste
- **WHEN** the setting is on and the user pastes bold text with Ctrl+V
- **THEN** the field contains the text without bold

### Requirement: Paste images as PNG setting
Settings SHALL include "Paste images as PNG", off by default, which makes pasted images be stored as PNG instead of JPEG.

#### Scenario: PNG paste
- **WHEN** the setting is on and the user pastes an image
- **THEN** the stored media file is a PNG

### Requirement: Custom study limits
Each deck's menu SHALL offer Custom study, with choices to increase today's new card limit and today's review card limit by a number, showing how many more cards are available for each.

#### Scenario: Ten more new cards
- **WHEN** the user opens Custom study for `Spanish`, chooses to increase today's new card limit by 10 and confirms
- **THEN** the deck list shows up to 10 more new cards for `Spanish` today, and its options are unchanged

### Requirement: Export entries
Each deck's menu on the deck list SHALL offer Export, and Settings SHALL offer Export collection.

#### Scenario: Export from the deck menu
- **WHEN** the user chooses Export in the menu of `Spanish`
- **THEN** the export options for `Spanish` are shown

### Requirement: Check media entry
Settings SHALL offer Check media, which shows the media report with its actions.

#### Scenario: Open the report
- **WHEN** the user chooses Check media
- **THEN** the report with unused and missing files and the trash size is shown

### Requirement: Check database entry
Settings SHALL offer Check database, showing progress while it runs and its report when it finishes.

#### Scenario: Run the check
- **WHEN** the user chooses Check database
- **THEN** progress is shown, and then the report

### Requirement: Open statistics
The deck list SHALL offer Statistics for the whole collection, and each deck's menu SHALL offer Statistics for that deck.

#### Scenario: Collection statistics
- **WHEN** the user chooses Statistics on the deck list
- **THEN** the statistics screen opens for the whole collection

### Requirement: Ignore accents setting
Settings SHALL include "Ignore accents in search", off by default. The value SHALL persist and be included in backups.

#### Scenario: Setting on
- **WHEN** the setting is on and the user searches `cafe`
- **THEN** notes containing `café` are listed

### Requirement: Deck overview
Choosing a deck on the deck list SHALL open an overview showing the deck's name, its new, learning and review counts for today, the number of buried cards, its description, and its progress: a bar of new, learning, young and mature cards of the deck and its subdecks, leaving out suspended cards, with the share of mature cards. It SHALL have actions Study now and Options, and Unbury when cards are buried. When nothing is due, the overview SHALL say so and when the next card is due.

#### Scenario: Overview before studying
- **WHEN** the user taps deck `Spanish` with 5 new and 30 review cards due
- **THEN** the overview shows 5 and 30 with a Study now button, and no session has started

#### Scenario: Nothing due
- **WHEN** the user taps a deck with nothing due and the next card due tomorrow
- **THEN** the overview says nothing is due and that the next card is due tomorrow

#### Scenario: Deck progress
- **WHEN** `Spanish` has 100 cards, 10 of them suspended, and of the other 90 there are 18 new, 4 learning, 12 young and 56 mature
- **THEN** the overview shows a bar with those four parts and "62 % mature"

#### Scenario: Buried cards count by state
- **WHEN** a mature card of `Spanish` is buried
- **THEN** it still counts as mature in the progress bar

### Requirement: Learn ahead limit setting
Settings SHALL include the learn ahead limit in minutes, from 0 to 1440, defaulting to 20. The value SHALL persist and be included in backups.

#### Scenario: Turn off learning ahead
- **WHEN** the user sets the learn ahead limit to 0 and restarts the app
- **THEN** the limit is still 0

### Requirement: Timebox setting
Settings SHALL include the timebox time limit in minutes, where 0 (the default) turns it off. The value SHALL persist and be included in backups.

#### Scenario: Enable timebox
- **WHEN** the user sets the timebox to 15 minutes and restarts the app
- **THEN** the setting is still 15 minutes

### Requirement: Review settings
Settings SHALL include a Review section with: show remaining card count (on), show next review time above answer buttons (on), Space/Enter also answers Good (on), hide bars during review (off), and question fade in milliseconds (100, from 0 to 1000). The values SHALL persist and be included in backups.

#### Scenario: Setting persists
- **WHEN** the user turns off "Show remaining card count" and restarts the app
- **THEN** the setting is still off

### Requirement: Study deck picker
Pressing `/` outside a text field, or choosing "Study deck" on the deck list, SHALL open a picker that filters decks by their full names, matching all typed words in any order, case-insensitively, and lists recently studied decks first when nothing is typed. Choosing a deck SHALL start studying it.

#### Scenario: Find a subdeck
- **WHEN** the user presses `/`, types `span verb` and presses Enter
- **THEN** studying starts for `Languages::Spanish::Verbs`

#### Scenario: Recent decks
- **WHEN** the user last studied `Biology` and opens the picker without typing
- **THEN** `Biology` is listed first

#### Scenario: Typing in a field
- **WHEN** the cursor is in a note field and the user types `/`
- **THEN** a `/` is inserted and no picker opens

### Requirement: Interface language
Settings SHALL include the interface language: "System default" (default), English or German. With "System default", the app SHALL use the first of the device's preferred languages it has a translation for, otherwise English. Changing the language SHALL apply immediately.

#### Scenario: German interface
- **WHEN** the user chooses German
- **THEN** navigation, buttons, dialogs and settings are shown in German without restarting

#### Scenario: Unsupported system language
- **WHEN** the device's only preferred language is Finnish and the setting is "System default"
- **THEN** the interface is in English

### Requirement: Translated text
All interface text, error messages, relative intervals, dates and numbers SHALL follow the selected language. A text without a translation SHALL be shown in English. Card content SHALL NOT be translated.

#### Scenario: Error in German
- **WHEN** the interface is German and the user renames a deck to a name a sibling already uses
- **THEN** the error explains in German that the name is taken

#### Scenario: Interval label
- **WHEN** the interface is German and a rating button's interval is 3 days
- **THEN** the button shows the interval in German formatting

### Requirement: Built-in names
Stock note types (their names, field names and card type names), and the first deck and the default preset of a new collection, SHALL be created with names in the interface language. Note types, decks and presets that already exist SHALL keep their names when the language changes. Notes of an Image Occlusion note type SHALL be edited with the mask editor whatever its field names.

#### Scenario: New collection in German
- **WHEN** a new collection is created while the interface is German
- **THEN** its deck is named "Standard", its preset "Standard", and it has the note type "Einfach" with the fields "Vorderseite" and "Rückseite", whose card type "Karte 1" shows them

#### Scenario: Adding a stock note type in German
- **WHEN** the interface is German and the user adds the stock note type Cloze
- **THEN** the list of stock note types shows "Lückentext", and the new note type has the fields "Text" and "Rückseite Extra"

#### Scenario: Existing note types are kept
- **WHEN** a collection has the note type "Basic" and the user switches the interface to German
- **THEN** the note type is still named "Basic" with the fields "Front" and "Back"

#### Scenario: German Image Occlusion
- **WHEN** the interface is German and the user adds an Image Occlusion note
- **THEN** the note type "Bildverdeckung" is created with the fields "Verdeckung" and "Bild", and the note is edited with the mask editor

### Requirement: UI size
Settings SHALL include a UI size from 75 % to 200 % (default 100 %) that scales text, icons and spacing throughout the app, combined with the operating system's text scale. Changes SHALL apply immediately and persist. At every UI size, every visible element SHALL respond to taps, clicks, hovers, drags and scrolling wherever it appears in the window, including near its right and bottom edges.

#### Scenario: Larger interface
- **WHEN** the user sets the UI size to 150 %
- **THEN** text and icons in every screen are 1.5 times their previous size without restarting

#### Scenario: Smaller interface stays usable
- **WHEN** the UI size is 75 % and the user taps a control at the right or bottom edge of the window
- **THEN** the control responds as it does at 100 %

#### Scenario: Switching back from a smaller size
- **WHEN** the UI size is 75 % and the user opens Settings and picks 100 % in the UI size control
- **THEN** the UI size changes to 100 % at once and stays 100 % after a restart

### Requirement: Minimalist mode
Settings SHALL include a minimalist mode, off by default, which removes shadows, elevation tints and decorative fills from the interface in light and dark themes.

#### Scenario: Flat interface
- **WHEN** the user turns on minimalist mode
- **THEN** cards, dialogs and bars are shown without shadows

### Requirement: Reduce motion
Settings SHALL include reduce motion with the choices follow system (default), on and off. When reduce motion is in effect, page transitions, expand and collapse animations and fades SHALL be replaced by immediate changes.

#### Scenario: System preference
- **WHEN** the OS requests reduced motion and the setting is "follow system"
- **THEN** opening a study session shows it without a transition animation

### Requirement: Keyboard shortcut hints
On Linux, Windows, macOS and the web on a desktop operating system, menu entries, tooltips and settings entries that have a keyboard shortcut SHALL show it next to their label, at every window width. On Android and iOS, including the web app in a mobile browser, they SHALL show only the label, without the shortcut. The shortcuts SHALL keep working on every platform.

#### Scenario: Study menu on a phone
- **WHEN** the user opens the study screen's more menu on Android
- **THEN** it lists "Edit card", "Card info" and "Set due date" without "(E)", "(I)" or "(Ctrl+Shift+D)"

#### Scenario: Study menu on desktop
- **WHEN** the user opens the study screen's more menu on Linux
- **THEN** it lists "Edit card (E)", "Card info (I)" and "Set due date (Ctrl+Shift+D)"

#### Scenario: Narrow desktop window
- **WHEN** the Linux app window is narrower than a phone and the user opens the editor's math menu
- **THEN** each entry still shows its shortcut, such as "Ctrl+M, M"

#### Scenario: Undo menu on iOS
- **WHEN** the last change was deleting 3 notes and the user opens the app bar menu on iOS
- **THEN** the Undo item reads "Undo Delete 3 notes" with no "⌘Z" or "Ctrl+Z" beside it

#### Scenario: Flag menu on a phone
- **WHEN** the user opens the flag menu while studying on Android
- **THEN** each flag shows its colour and name without "Ctrl+1" to "Ctrl+7"

#### Scenario: Tooltip on a tablet
- **WHEN** the user long-presses the Bold button in the field editor on an Android tablet
- **THEN** the tooltip reads "Bold", not "Bold (Ctrl+B)"

#### Scenario: Backup entry on a phone
- **WHEN** the user opens Settings on Android
- **THEN** "Create backup now" has no "Ctrl+Alt+B" below it, while on Windows it does

#### Scenario: Mobile browser
- **WHEN** the web app is opened in Safari on an iPhone and the user opens the browser's bulk-action menu
- **THEN** it shows "Set due date" without a shortcut

#### Scenario: Hardware keyboard on a tablet
- **WHEN** a keyboard is connected to an Android tablet and the user presses `E` while studying
- **THEN** the card opens in the editor, although the menu does not show "(E)"

#### Scenario: German key names
- **WHEN** the interface is German and the user opens the study screen's more menu on Windows
- **THEN** the record entry reads "Eigene Stimme aufnehmen (Umschalt+V)"

### Requirement: License and source in About
Settings' About section and the licenses page SHALL show Cardflip's copyright notice and state that Cardflip is free software under the GNU GPL version 3 or later and comes with no warranty, and the About section SHALL offer "Source code", which opens the project's public repository in the browser.

#### Scenario: License named
- **WHEN** the user opens Settings
- **THEN** the About section shows the version, the copyright notice, that Cardflip is licensed under the GNU GPL version 3 or later and that it comes with no warranty

#### Scenario: Source code opens the repository
- **WHEN** the user taps "Source code" in the About section
- **THEN** the project's repository opens in the system browser, or in a new tab in the web build

### Requirement: Open-source licenses page
Settings' About section SHALL offer "Open-source licenses", which opens a page that lists every component the running app contains, by name, and shows each one's license text when it is selected. The page SHALL work offline, on every platform including the web build, and in the interface language.

#### Scenario: Open the page
- **WHEN** the user taps "Open-source licenses" in the About section
- **THEN** a page opens with Cardflip's name, version and license at the top and a list of components below

#### Scenario: Read a license
- **WHEN** the user selects `flutter_quill` on the licenses page
- **THEN** its MIT license text is shown

#### Scenario: Offline
- **WHEN** the device has no network connection and the user opens the licenses page
- **THEN** all components and their license texts are shown

#### Scenario: German interface
- **WHEN** the interface language is German and the user opens Settings
- **THEN** the entry reads "Open-Source-Lizenzen" and the page's own labels are German

### Requirement: Licenses page contents
The licenses page SHALL list Cardflip itself with the GNU GPL version 3 text; the code taken from other projects (fsrs-rs, the Anki Image Occlusion template, fluffy) with their licenses; every Dart and Flutter package, including the Flutter engine; every Go module linked into the app, and the Go standard library; and, on the platforms and in the packages where the app bundles them, libmpv and FFmpeg.

#### Scenario: Cardflip and ported code
- **WHEN** the user opens the licenses page
- **THEN** it lists Cardflip once, as its package `cardflip`, with the GPL version 3, "fsrs-rs" with the BSD 3-Clause license, "Anki" with the GNU AGPL version 3 and "fluffy" with the MIT license

#### Scenario: Go modules
- **WHEN** the user opens the licenses page
- **THEN** it lists `github.com/klauspost/compress`, `github.com/open-spaced-repetition/go-fsrs/v3` and the Go standard library, each with its license text

#### Scenario: Web build
- **WHEN** the user opens the licenses page in the web build
- **THEN** it lists the Go modules compiled into the web build's Go code, the same as the native apps

#### Scenario: Media libraries on Windows
- **WHEN** the user opens the licenses page in the Windows app
- **THEN** it lists libmpv and FFmpeg with the licenses of the bundled builds

#### Scenario: Media libraries on Linux
- **WHEN** the user opens the licenses page in the Linux app, which uses the system's libmpv
- **THEN** libmpv and FFmpeg are not listed

#### Scenario: Media libraries in the AppImage
- **WHEN** the user opens the licenses page in the Linux app started from the AppImage, which bundles libmpv
- **THEN** it lists libmpv and FFmpeg with the GNU GPL version 3 text, the versions bundled and where their source is published

### Requirement: Licenses list kept current
The list of Go module licenses SHALL match the modules the app links on every platform. The verification gates SHALL fail when a module is added, removed or updated without the list being regenerated, or when a linked module has no license file.

#### Scenario: Dependency added
- **WHEN** a developer adds a Go module that the app uses and runs the verification gates without regenerating the list
- **THEN** the gates fail and name the license list as out of date

#### Scenario: Module without license
- **WHEN** a linked Go module has no license file
- **THEN** generating the list fails and names the module

### Requirement: Deck progress in the deck list
Settings SHALL include *Show deck progress in the deck list*, off by default and belonging to the device. When it is on, each deck row SHALL show a thin bar of the deck's new, learning, young and mature cards, including its subdecks and leaving out suspended cards.

#### Scenario: Bars shown
- **WHEN** the user turns on *Show deck progress in the deck list*
- **THEN** each deck row shows a thin progress bar under its name

#### Scenario: Off by default
- **WHEN** the user has not changed the setting
- **THEN** deck rows show no progress bar
