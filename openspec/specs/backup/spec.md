# backup Specification

## Purpose

Lets the user save the entire collection to a portable `.cardflip` file and restore it later, on the same or another device or platform.

## Requirements

### Requirement: Export full backup
The user SHALL be able to export the whole collection to a single `.cardflip` file containing all decks, deck options, note types, notes, cards with scheduling state, review history, media and settings. On native platforms, the user chooses where to save the file. On the web, the file is downloaded.

#### Scenario: Export on web
- **WHEN** the user chooses "Export backup" in the web build
- **THEN** the browser downloads a file named `cardflip-<YYYY-MM-DD>.cardflip`

### Requirement: Restore replaces the collection
Restoring a `.cardflip` file SHALL replace the current collection entirely, after a confirmation that warns the current data will be overwritten. After the restore, the collection SHALL be identical to the one that was exported.

#### Scenario: Round trip
- **WHEN** a collection is exported and the backup is restored on another platform
- **THEN** deck list, due counts, card contents, scheduling states and review history match the original

### Requirement: Failed restore keeps current data
If a restore fails, for example because the file is corrupt, is not a Cardflip backup, or comes from a newer, unsupported format version, the current collection SHALL remain unchanged, and the error SHALL say why the restore failed.

#### Scenario: Wrong file
- **WHEN** the user tries to restore a `.zip` that is not a Cardflip backup
- **THEN** an error says the file is not a Cardflip backup, and the collection is unchanged

### Requirement: Versioned format
Every backup SHALL record its format version and the app version that created it. The app SHALL restore any backup whose format version is less than or equal to its own.

#### Scenario: Newer backup
- **WHEN** a backup with a higher format version than the app supports is restored
- **THEN** the restore is refused with a message to update Cardflip

### Requirement: Automatic backups
While the app is open, the collection SHALL be backed up automatically, without media, every backup interval (default 30 minutes) when it has changed since the last backup. A backup SHALL also be made before restoring a backup, importing a collection package, and checking the database. Backups SHALL be stored on the device.

#### Scenario: Periodic backup
- **WHEN** the user studies for 40 minutes with the default interval
- **THEN** at least one automatic backup exists from that time

#### Scenario: Nothing changed
- **WHEN** the app stays open for an hour without any change
- **THEN** no new automatic backup is made

### Requirement: Backup thinning
Automatic backups SHALL be thinned so that all of today's backups are kept, plus the latest backup of each of the last N days, of each of the last N weeks and of each of the last N months, with N configurable per period (default 8). Other automatic backups SHALL be deleted.

#### Scenario: Old backups removed
- **WHEN** backups exist for every day of the last 60 days with default settings
- **THEN** 8 daily, up to 8 weekly and up to 8 monthly backups remain, without duplicates

### Requirement: Restore from automatic backup
The user SHALL be able to list automatic backups with their date and time and restore one, with the same confirmation and all-or-nothing behaviour as restoring a backup file. Media SHALL be left unchanged.

#### Scenario: Undo a bad day
- **WHEN** the user restores this morning's automatic backup and confirms
- **THEN** the collection is as it was this morning

### Requirement: Create backup now
The user SHALL be able to create an automatic backup immediately from Settings and with Ctrl+Alt+B (Cmd+Option+B on macOS), without choosing a location.

#### Scenario: Quick backup
- **WHEN** the user presses Ctrl+Alt+B
- **THEN** a new backup appears at the top of the automatic backup list
