# Spec Delta

## MODIFIED Requirements

### Requirement: Supported platforms
Cardflip SHALL run with identical features on Linux, Windows, macOS, Android, iOS and the web (current Chrome, Firefox, Safari and Edge), except the MCP server for agents, which SHALL be available only on Linux, Windows and macOS.

#### Scenario: Same behavior on web
- **WHEN** the same backup is restored in the web build and in the Linux build
- **THEN** both show the same decks, due counts and card contents

#### Scenario: MCP server only on desktop
- **WHEN** the user opens Settings on Android
- **THEN** there is no "Agents (MCP)" section, while on Linux there is
