# Spec Delta

## ADDED Requirements

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
