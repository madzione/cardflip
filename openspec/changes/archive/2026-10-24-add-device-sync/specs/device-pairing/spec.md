# Spec Delta

## Purpose

Lets the user's own devices recognise each other securely on networks the user trusts, so that they can sync without an account or a server and without announcing themselves on other networks.

## ADDED Requirements

### Requirement: Sync section
On Linux, Windows, macOS, Android and iOS, Settings SHALL have a Sync section with this device's name, the trusted networks, the paired devices, "Pair a device" and the open profile's links. On the web, Cardflip SHALL NOT show the section and SHALL NOT use the network for sync.

#### Scenario: Web build
- **WHEN** the user opens Settings in the web build
- **THEN** there is no Sync section

#### Scenario: Android
- **WHEN** the user opens Settings on Android
- **THEN** the Sync section shows this device's name and an empty list of paired devices

### Requirement: Device identity
Each installation SHALL have a device identity, created on first use of the Sync section, and a device name that defaults to the computer's or phone's name and can be changed (1 to 40 characters). Identity, name, trusted networks and paired devices SHALL belong to the device: they are the same for every profile, are not stored in any collection and are not included in backups.

#### Scenario: Rename device
- **WHEN** the user changes the device name to `Kitchen tablet`
- **THEN** paired devices show `Kitchen tablet` after their next connection to it

#### Scenario: Backup restored on another device
- **WHEN** a backup made on a paired laptop is restored on a new phone
- **THEN** the phone has no paired devices and its own device identity

### Requirement: Trusted networks only
Cardflip SHALL announce itself, search for devices, accept connections and connect to paired devices only while the current network is trusted. On any other network, or with no network, it SHALL send and accept no sync traffic.

#### Scenario: Café network
- **WHEN** the phone joins a café's Wi-Fi that is not trusted and the app is open
- **THEN** Cardflip does not announce itself, does not search for devices and does not accept connections

#### Scenario: Back home
- **WHEN** the phone returns to the trusted home Wi-Fi with the app open
- **THEN** Cardflip announces itself and syncs with paired devices it finds

### Requirement: Trust a network
The Sync section SHALL show the current network and whether it is trusted, and SHALL let the user trust it or stop trusting it. A Wi-Fi network SHALL be identified by its name, and any other network by its subnet. Networks SHALL NOT be trusted by default.

#### Scenario: Trust home Wi-Fi
- **WHEN** the device is on the Wi-Fi network `Home` and the user taps "Trust this network"
- **THEN** `Home` appears in the list of trusted networks, and sync becomes available while on it

#### Scenario: Wired network
- **WHEN** a desktop is connected by cable to the subnet `192.168.178.0/24` and the user trusts it
- **THEN** the list shows the trusted network as `Wired network 192.168.178.0/24`

#### Scenario: Remove trust
- **WHEN** the user removes `Home` from the trusted networks while on it
- **THEN** Cardflip stops announcing itself and closes open sync connections at once

### Requirement: Wi-Fi name permission
When the platform requires a permission to read the Wi-Fi name, Cardflip SHALL ask for it only when the user is about to trust a network, and SHALL explain why. If the user refuses, the network SHALL be identified by its subnet instead.

#### Scenario: Location refused on Android
- **WHEN** the user taps "Trust this network" on Android and refuses the location permission
- **THEN** the network is trusted by its subnet, and the list shows it as `Network 192.168.1.0/24`

### Requirement: Pair a device
"Pair a device" SHALL make this device visible for pairing for 5 minutes or until the screen is closed, and SHALL list the other devices on the network that are visible for pairing. Choosing a device SHALL send it a pairing request, which its user accepts or declines.

#### Scenario: Request shown
- **WHEN** the laptop and the phone both have "Pair a device" open and the user chooses the laptop on the phone
- **THEN** the laptop shows a pairing request from the phone's device name

#### Scenario: Not visible outside pairing
- **WHEN** the laptop is not paired with the phone and has "Pair a device" closed
- **THEN** the phone's "Pair a device" list does not show the laptop

### Requirement: Confirm the pairing code
When a pairing request is accepted, both devices SHALL show the same 6-digit code derived from the secure connection, and the pairing SHALL complete only after the user confirms on both devices that the codes match. If either side declines, cancels or does not confirm within 2 minutes, nothing SHALL be stored.

#### Scenario: Codes match
- **WHEN** both devices show `482 913` and the user confirms on both
- **THEN** each device lists the other under paired devices

#### Scenario: Code declined
- **WHEN** the user says the codes do not match on one device
- **THEN** both devices show that pairing failed and neither stores the other

### Requirement: Pair by address
"Pair a device" SHALL show this device's address and port, and SHALL let the user enter another device's address and port. The user can then pair with a device the network does not let them find, with the same code confirmation.

#### Scenario: Multicast blocked
- **WHEN** the router blocks device discovery and the user enters `192.168.1.20:47812` shown on the laptop
- **THEN** the pairing request reaches the laptop and continues with the code confirmation

### Requirement: Unanswered connections
When another device on the network does not answer a connection for pairing or sync within the time limit, Cardflip SHALL say that it did not answer, that a firewall on that device may block sync, and which TCP port to allow there. For pairing, it SHALL also suggest starting the pairing on that device.

#### Scenario: Desktop firewall
- **WHEN** the laptop's firewall drops incoming connections and the user chooses the laptop under Devices nearby on the phone
- **THEN** the phone says the laptop did not answer and that its firewall may block sync, names port 47812, and suggests starting the pairing on the laptop

### Requirement: Only paired devices sync
Cardflip SHALL exchange collection data only over encrypted connections with paired devices, verified by the keys stored at pairing. A connection from a device that is not paired, or that presents a different key, SHALL be refused without exchanging collection data.

#### Scenario: Impostor with the same name
- **WHEN** an unpaired device on the trusted network announces itself with the name of a paired device and tries to sync
- **THEN** the connection is refused and nothing is exchanged

### Requirement: Unpair a device
The user SHALL be able to unpair a device from the paired devices list after a confirmation. Unpairing SHALL remove the device and its profile links. If the other device is reachable, it SHALL remove this device as well. Otherwise it SHALL do so the next time it tries to connect.

#### Scenario: Unpair while away
- **WHEN** the user unpairs the phone on the laptop while the phone is not on the network
- **THEN** the laptop no longer lists the phone, and the next connection attempt by the phone is refused, after which the phone shows the laptop as unpaired

### Requirement: Paired device status
The paired devices list SHALL show for each device its name, whether it is reachable now, and when this profile last synced with it.

#### Scenario: Device away
- **WHEN** the phone is paired and switched off
- **THEN** the laptop lists the phone as not reachable, with the time of the last sync
