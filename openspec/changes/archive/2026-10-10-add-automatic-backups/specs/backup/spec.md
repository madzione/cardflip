# Spec Delta

## ADDED Requirements

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
