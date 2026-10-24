# Spec Delta

## Purpose

Several separate collections on one device or browser, each with its own cards, media, settings and backups, and switching between them.

## ADDED Requirements

### Requirement: Separate profiles
Each profile SHALL have its own collection, media, settings, automatic backups and review history. Nothing done in one profile SHALL be visible in another.

#### Scenario: Separate decks
- **WHEN** profile `Anna` has deck `Spanish` and the user switches to profile `Ben`
- **THEN** the deck list does not show `Spanish`

### Requirement: Manage profiles
The user SHALL be able to add, rename, open and delete profiles from a profile list reached from Settings. Profile names SHALL be unique, compared case-insensitively. Deleting a profile SHALL require a confirmation naming it and its card count, and the open profile SHALL NOT be deletable.

#### Scenario: Add and open
- **WHEN** the user adds profile `Ben` and opens it
- **THEN** the app shows a fresh collection with the `Default` deck

#### Scenario: Delete
- **WHEN** the user deletes profile `Ben`, which has 300 cards, and confirms
- **THEN** `Ben` and all its data are removed, and other profiles are unchanged

### Requirement: Startup profile
On start, the app SHALL open the last used profile, unless "Show profile list at startup" is on, in which case it SHALL show the profile list first. An installation from before profiles existed SHALL open its existing collection as profile `Default`.

#### Scenario: Upgrade
- **WHEN** an existing user updates to a version with profiles
- **THEN** the app opens their collection as profile `Default` with all data intact

### Requirement: Backups per profile
Exporting a backup SHALL contain only the open profile's collection, and restoring a backup SHALL replace only the open profile's collection.

#### Scenario: Restore into one profile
- **WHEN** the user restores a backup while profile `Ben` is open
- **THEN** only `Ben`'s collection is replaced
