# Spec Delta

## Purpose

Lets the user save the entire collection to a portable `.cardflip` file and restore it later, on the same or another device or platform.

## ADDED Requirements

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
