# Spec Delta

## Purpose

Keeps the user's collection and media stored durably on the device, on every platform, so that no data is lost between launches, after crashes, or across app updates.

## ADDED Requirements

### Requirement: Local-only storage
All collection data and media SHALL be stored on the device, in the platform's application data location (native) or the browser's persistent site storage (web). The app SHALL NOT send collection data over the network.

#### Scenario: Offline use
- **WHEN** the device has no network connection
- **THEN** every feature of the app works normally

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
