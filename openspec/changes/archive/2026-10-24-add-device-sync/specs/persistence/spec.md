# Spec Delta

## MODIFIED Requirements

### Requirement: Local-only storage
All collection data and media SHALL be stored on the device, in the platform's application data location (native) or the browser's persistent site storage (web). The app SHALL NOT send collection data over the network, except to devices the user has paired, over encrypted connections, on networks the user has marked as trusted. Collection data SHALL never be sent to a server or over the internet.

#### Scenario: Offline use
- **WHEN** the device has no network connection
- **THEN** every feature of the app works normally, apart from syncing with paired devices

#### Scenario: No pairing, no traffic
- **WHEN** the user has never paired a device
- **THEN** the app sends no collection data over any network
