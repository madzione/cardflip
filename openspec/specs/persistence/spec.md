# persistence Specification

## Purpose

Keeps the user's collection and media stored durably on the device, on every platform, so that no data is lost between launches, after crashes, or across app updates.

## Requirements

### Requirement: Local-only storage
All collection data and media SHALL be stored on the device, in the platform's application data location (native) or the browser's persistent site storage (web). The app SHALL NOT send collection data over the network, except to devices the user has paired, over encrypted connections, on networks the user has marked as trusted. Collection data SHALL never be sent to a server or over the internet.

#### Scenario: Offline use
- **WHEN** the device has no network connection
- **THEN** every feature of the app works normally, apart from syncing with paired devices

#### Scenario: No pairing, no traffic
- **WHEN** the user has never paired a device
- **THEN** the app sends no collection data over any network

### Requirement: Data survives restarts
Every completed change SHALL still be present after the app is closed and reopened, including additions, edits, deletions, answers, imports and option changes.

#### Scenario: Restart
- **WHEN** the user adds a note, closes the app and reopens it
- **THEN** the note is present

### Requirement: Crash safety
A crash or power loss during a save SHALL NOT corrupt the collection. On next start, the collection SHALL open in a consistent state containing every change that finished saving.

#### Scenario: Killed during save
- **WHEN** the app process is killed while writing the collection
- **THEN** the next launch opens the collection without errors, and at most the change being written is lost

### Requirement: Unreadable storage is reported
If stored data cannot be read at startup, the app SHALL show an error screen. That screen SHALL offer to restore from a backup or to start a new collection after an explicit confirmation, and SHALL NOT silently discard data.

#### Scenario: Corrupt data file
- **WHEN** the collection storage is unreadable at startup
- **THEN** the app shows the error screen and keeps the unreadable data untouched until the user chooses an action

### Requirement: Web storage persistence
On the web, the app SHALL ask the browser for persistent storage, so that the browser does not evict the data under storage pressure. If the browser refuses, the app SHALL tell the user once that browser data may be cleared, and SHALL recommend regular backups.

#### Scenario: Persistence refused
- **WHEN** the browser does not grant persistent storage
- **THEN** a one-time notice recommends exporting backups regularly

### Requirement: Scale
Opening a collection of 50,000 cards with 500,000 review log entries SHALL take less than 3 seconds on a typical desktop, and saving an answer SHALL not block the UI.

#### Scenario: Large collection startup
- **WHEN** the app starts with a collection of 50,000 cards
- **THEN** the deck list is usable within 3 seconds on a typical desktop

### Requirement: Check media
The user SHALL be able to check media and see the files no note, template or styling refers to, and the file names notes refer to that are missing, with the space used. Files whose names start with `_` SHALL NOT be reported as unused.

#### Scenario: Unused and missing
- **WHEN** the media contains `old.jpg`, which nothing refers to, and a note refers to `gone.mp3`, which is not in the media
- **THEN** the report lists `old.jpg` as unused and `gone.mp3` as missing

#### Scenario: Static template file
- **WHEN** the media contains `_logo.png`, which no note refers to
- **THEN** `_logo.png` is not listed as unused

### Requirement: Media trash
Deleting unused media from the report SHALL move the files to a media trash. The user SHALL be able to restore the trash or empty it. Storage compaction SHALL move unreferenced media to the trash instead of deleting them, and SHALL delete files that have been in the trash for more than 30 days.

#### Scenario: Restore deleted files
- **WHEN** the user deletes unused files and then chooses Restore deleted
- **THEN** the files are back in the media

### Requirement: Tag notes with missing media
From the report, the user SHALL be able to add the tag `missing-media` to every note that refers to a missing file.

#### Scenario: Find notes to fix
- **WHEN** 3 notes refer to missing files and the user chooses Tag missing
- **THEN** those 3 notes have the tag `missing-media`

### Requirement: Check and repair the collection
The user SHALL be able to check the collection for inconsistencies and repair them in one operation: cards without a note or note type, notes with a wrong number of fields, cards for missing templates or cloze numbers, missing cards, cards in missing decks, invalid scheduling values, and invalid tags. The check SHALL report each kind of problem found with its count, or that none was found. If the repair cannot be saved, the collection SHALL stay unchanged.

#### Scenario: Orphaned card
- **WHEN** the collection contains a card whose note no longer exists and the user checks the database
- **THEN** the card is deleted and the report says 1 card without a note was removed

#### Scenario: Healthy collection
- **WHEN** the user checks a consistent collection
- **THEN** the report says no problems were found and nothing changes
