# Spec Delta

## ADDED Requirements

### Requirement: Backup settings
Settings SHALL include the backup interval in minutes (default 30, 0 turns off periodic backups), the number of daily, weekly and monthly backups to keep (default 8 each), the list of automatic backups, and the storage space they use.

#### Scenario: Periodic backups off
- **WHEN** the interval is 0 and the user imports a collection package
- **THEN** a backup is still made before the import, and no periodic backups are made
