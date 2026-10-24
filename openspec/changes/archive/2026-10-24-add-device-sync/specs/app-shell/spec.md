# Spec Delta

## MODIFIED Requirements

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
