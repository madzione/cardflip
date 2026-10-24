# Proposal

## Why

Many people study on more than one device: they add cards on a laptop and review them on a phone. Today the only way to move a collection is to export a `.cardflip` backup and restore it on the other device. That replaces everything there, so studying on two devices in the same week loses one side's reviews. Cardflip has no account and no server, and it should stay that way. Devices on the same home network can sync directly with each other instead.

## What Changes

- **Device identity**: every native installation gets a device key pair and a device name the user can edit. Both belong to the device, like the MCP settings: they are not stored in collections and not included in backups.
- **Trusted networks**: the app never announces itself, searches for devices or accepts connections on a network the user has not marked as trusted. On an unknown network, the Sync section offers to trust it (identified by Wi-Fi name, or by subnet for wired networks). Trusted networks can be reviewed and removed. Public and corporate networks therefore see no traffic from Cardflip.
- **Pairing**: in Settings → Sync → "Pair a device", two devices on a trusted network find each other (mDNS/DNS-SD, `_cardflip._tcp`). If multicast is blocked, the user can type the address shown on the other device instead. Both screens show the same 6-digit code, and the user confirms it on both. After that, the devices recognise each other by their pinned keys. Unpairing works from either side.
- **Profile links**: sync is per profile. A local profile is linked to one profile on a paired device, which can be an existing one or a new one created for the link. The first sync of a link asks whether to merge the two collections or replace one with the other. An automatic backup is made first.
- **Synchronisation**: decks, presets, note types, notes, cards with their scheduling state, the review history, deletions, media files and the synced subset of collection settings (rollover hour, learn-ahead limit, flag names and other study settings; not theme, language, UI size, the last-used pickers or MCP). Changes are merged object by object, and the newest change wins. A card's scheduling and its flag, suspension and deck are merged separately, so a review on one device and a flag on the other are both kept. Review histories are combined. Deletions propagate. Duplicate deck, preset and note type names that a merge creates are resolved automatically.
- **When it runs**: automatically when the app starts or resumes, shortly after local changes, when a paired device appears on the network and before the app closes, plus a manual "Sync now". Sync runs only on a trusted network, only with paired devices, and only while the app is open. A status indicator on the deck list shows syncing, last synced and errors.
- **Safety**: connections use mutual TLS 1.3 with the pinned device keys. Devices refuse to sync when their clocks differ by more than 10 minutes, or when the other device runs an incompatible version. Restoring a backup, importing a collection package or starting a new collection makes the next sync of each link ask again whether to merge or replace.
- **Platforms**: Linux, Windows, macOS, Android and iOS. **The web build cannot take part**, because browsers can neither discover devices nor accept connections, so it does not show the Sync section.
- **BREAKING (spec)**: persistence's "Local-only storage" changes from "never sends collection data over the network" to "sends it only to paired devices on trusted local networks, encrypted, and never to the internet".

## Capabilities

### New Capabilities

- `device-pairing`: device identity and name, trusted networks, discovery, pairing and unpairing, manual addresses, and the Sync section's device and network management.
- `device-sync`: profile links and the first-sync choice, what is synced and how changes merge, when sync runs, the status indicator, and clock and version checks.

### Modified Capabilities

- `persistence`: "Local-only storage" allows encrypted transfer to paired devices on trusted local networks, and nothing else.
- `app-shell`: "Supported platforms" adds device sync as a feature that is not available on the web.
- `collection`: "Undo changes" says that synced changes are not undo steps, and that the history is cleared when incoming changes touch something it covers.

## Impact

- **Go**: a new `sync` package (merge engine and protocol, plain Go with `go test`), a `SyncService` bound with fluffy, and the network code (TLS listener, client, pairing) in `!js` files with a js stub. `core` gains per-object sync stamps (a hybrid logical clock), tombstones for deleted objects, change sequence numbers, a "sync" op kind that applies incoming changes and resolves name clashes, and snapshot version 4 with a migration. `storage` gains `sync.json` (device key, name, trusted networks, paired devices) in the data root and per-profile link state next to each profile's collection, neither of them in backups.
- **Flutter**: the Sync section in Settings (device name, trusted networks, paired devices, pairing flow, per-profile links), the status indicator on the deck list, the first-sync choice dialog, and the strings (English, German). New dependencies: `bonsoir` (DNS-SD advertise and browse through the platform's own service: NsdManager, Bonjour, Windows DNS-SD, Avahi), `network_info_plus` and `connectivity_plus` (network identity and change events), `device_info_plus` (the default device name on phones) and `permission_handler` (the location permission that reading the Wi-Fi name needs on Android and iOS, asked for only when trusting a network).
- **Platform**: Android needs the `INTERNET`, `ACCESS_NETWORK_STATE`, `ACCESS_WIFI_STATE`, `CHANGE_WIFI_MULTICAST_STATE` and location permissions (location is needed to read the Wi-Fi name). iOS needs `NSLocalNetworkUsageDescription`, `NSBonjourServices`, a location usage string and the Access WiFi Information entitlement. macOS needs `com.apple.security.network.client` in both entitlements files.
- **Docs**: the README's "no sync" statement, `ANKI_FEATURE_GAP.md` (syncing) and `ROADMAP.md` (a new phase).
- **Tests**: Go merge and convergence tests (two and three devices, conflicts, deletions, name clashes, collisions), protocol tests over loopback TLS, widget tests for the Sync section and indicator, and a Linux integration test syncing two app data directories.
