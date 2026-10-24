# Spec Delta

## Purpose

The application frame around all features: which platforms Cardflip runs on, how the user navigates between screens, the layout across screen sizes, theming, and app-wide settings.

## ADDED Requirements

### Requirement: Supported platforms
Cardflip SHALL run with identical features on Linux, Windows, macOS, Android, iOS and the web (current Chrome, Firefox, Safari and Edge).

#### Scenario: Same behavior on web
- **WHEN** the same backup is restored in the web build and in the Linux build
- **THEN** both show the same decks, due counts and card contents

### Requirement: Main navigation
The app SHALL provide top-level destinations for Decks (home), Browse, Add and Settings. On wide screens (at least 840 logical pixels), navigation SHALL be a side rail. On narrow screens, it SHALL be a bottom navigation bar.

#### Scenario: Phone layout
- **WHEN** the app runs on a phone in portrait
- **THEN** navigation is a bottom bar with Decks, Browse, Add and Settings

### Requirement: Deck list home screen
The home screen SHALL show the deck tree with new, learning and review counts for each deck. Parent decks SHALL be collapsible. Each deck SHALL have actions to study, add a note, open its options, rename it and delete it.

#### Scenario: Due counts visible
- **WHEN** a deck has 5 new, 2 learning and 30 review cards due
- **THEN** its row shows 5, 2 and 30 in three distinguishable colours

### Requirement: Theming
The app SHALL support light, dark and system theme modes, defaulting to system. The choice SHALL persist.

#### Scenario: Theme persists
- **WHEN** the user selects dark mode and restarts the app
- **THEN** the app starts in dark mode

### Requirement: App settings
Settings SHALL include theme mode, the day rollover hour, export backup, restore backup, import, and an About section showing the app version.

#### Scenario: Change rollover hour
- **WHEN** the user changes the rollover hour from 4 to 6
- **THEN** due counts and daily limits use 06:00 as the day boundary from then on

### Requirement: First run
On first launch, the app SHALL open on the deck list with the `Default` deck and a short empty-state hint offering to add a note or import a file.

#### Scenario: Empty state
- **WHEN** the app is launched for the first time
- **THEN** the deck list shows `Default` and buttons to add a note or import

### Requirement: Errors are shown, not swallowed
Any failed operation SHALL show the user a readable message. Unexpected errors SHALL NOT leave the UI in a loading state indefinitely.

#### Scenario: Backend error
- **WHEN** saving a note fails
- **THEN** a message explains the failure, and the editor keeps the user's input
