# Design

## Context

- The collection lives in the Go core (`golang/core`). It is held in memory and persisted by `storage.Store` as a snapshot plus a journal of ops. Every mutation goes through `Collection.commitAs`: it builds an `op` (whole new objects plus `Del*` ID lists), journals it durably, applies it, records an undo step and publishes a `core.Change`. Apply is deterministic, so replaying the journal rebuilds the state. Model objects are never modified in place.
- IDs (`core/ids.go`) are epoch milliseconds, made strictly increasing per collection and kept below 2^53 for JavaScript. `added:` searches and "newer objects have larger IDs" rely on this. Two devices can therefore create different objects with the same ID.
- Notes have a `GUID` and `Modified`. Cards and note types have `Modified`, which imports set to the source's time. Decks, presets and settings have no modification time. The review log lives in monthly segments keyed by entry ID. Deleted notes go to the deletion log.
- Deck, preset and note type names are unique, ignoring case. Every new collection creates its own `Default` deck and built-in note types with fresh IDs, while the Default preset always has ID 1.
- `core.Settings` mixes study settings (rollover hour, learn-ahead limit, flag names, review display options and others) with device-ish state (theme, language, last-used pickers, current and recent decks, the web storage notice).
- Media is stored by name (`Store.PutMedia`). `AddMedia` never overwrites a name: it reuses identical content or adds a numbered suffix. Media writes are not ops.
- Device-level state already exists outside collections: `profiles.json` and `mcp.json` in the data root (`storage/profiles_files.go`, `svc_mcp.go`). Each profile has its own data dir, `profiles/<id>`.
- The network code of the MCP server sits in `!js` files with a js stub, because the fluffy generator type-checks the web build. `golang/go.mod` already has `klauspost/compress`.
- Browsers cannot do mDNS or listen on sockets, so the web build is out of scope (see proposal).

## Goals / Non-Goals

**Goals:**
- Convergence: any set of linked devices that sync in any order and topology ends with identical collections, except for device-level settings.
- No data loss on concurrent work. Only true conflicts, where the same object field group changes on two devices, lose an edit, and the later one wins.
- Merge logic in plain Go, testable without a network, and shared by both directions.
- Unchanged behaviour and negligible overhead for people who never pair a device.

**Non-Goals:**
- Internet sync, relays, NAT traversal, or a hosted server.
- Field-level merging of note text, or keeping both versions of a conflicting note.
- Background sync while the app is closed or suspended (no OS background tasks).
- Syncing with Anki or AnkiWeb.
- Sync between profiles on the same device.

## Decisions

### 1. State-based merge with per-object stamps (LWW registers), not op replay

Each syncable object is a last-writer-wins register identified by `(kind, ID)`, with a **stamp** from a hybrid logical clock (HLC: wall-clock milliseconds, a counter, and the device's slot as tie-breaker). Cards are two registers: *scheduling* (`CardState`, `NewPos`, `OriginalPos`, `BuriedUntil`, `PreviewDue`) and *placement and marks* (`DeckID`, `HomeDeckID`, `FilteredPos`, `Flag`, `Suspended`). The synced part of `Settings` is one register, and so is each media name. Review log entries are immutable and are combined by union. Deletions are **tombstones** `(kind, ID, stamp)`. An incoming object replaces the local one only if its stamp is greater. A tombstone deletes the object only if its stamp is greater than the object's stamp. A note tombstone also deletes the note's cards. Cards deleted with their note get no tombstone of their own (card tombstones are for cards deleted alone, such as empty cards), and a delta that carries a note also carries its cards, so a note edited after its deletion elsewhere comes back there with its cards. Equal stamps, possible only between collections without device slots, are decided by the greater JSON, so both sides agree.

LWW with a union for the log is commutative, associative and idempotent, so convergence does not depend on sync order or on how many devices take part. Interrupted sessions can simply be repeated.

*Alternative:* replay the op journal on the other device (operation-based CRDT). Rejected: ops are compacted away, contain derived data (positions, burying), and would need causal delivery across three or more devices.
*Alternative:* the existing `Modified` fields as LWW time. Rejected: imports set them to a source's time, decks and presets lack them, and raw wall clocks give no causal order.

### 2. Sync metadata lives beside the objects, not in them

`Collection` gains `meta map[metaKey]objMeta{stamp, seq}` (and two stamps for cards), `tombstones`, a local change counter `seq`, an `epoch` (a random collection lineage ID) and the slot table (decision 4). They are saved in the snapshot as compact arrays and journaled as part of each `op` (`op.Meta`, `op.Tombs`), so replay stays deterministic. The model structs that Dart sees (`Card`, `Note`, …) do not change. That avoids stamps above 2^53 reaching JavaScript, and it keeps undo before-images free of sync metadata. Review log entries are the exception, as they outnumber everything else by far: their `seq` is stored in their review log segment record, with the highest `seq` per month in the snapshot, and they count as stamp 0. Only an entry added again after its deletion (redo of an answer) gets a register, so that it beats its tombstone.

`commitAs` stamps every object and tombstone the op touches with a fresh HLC stamp and the next `seq`, **in every mode**. Unlike `stampModified`, this includes undo and redo, so an undo after a sync is synced as a new change (collection delta, "Undo after sync"). The HLC stays at or above every stamp seen, including incoming ones.

### 3. Deltas by local sequence number

Every local or incoming change gets the next local `seq`. A link stores the highest peer `seq` it has fully received. A session asks the peer for "everything with seq > N", ordered by kind (presets, decks, note types, notes, cards, review entries, settings, tombstones, media manifest) and then by seq, in batches of about 500 objects or 4 MB. The receiver commits each batch as one `sync` op and saves a resume cursor `(kind, seq)` in the link state, so an interrupted session continues where it stopped (spec: Interrupted sync is safe). Changes that arrived from a peer are recorded with that peer's link ID in a range list, so they are not echoed back to the same peer.

*Alternative:* compare stamps directly ("send what is newer than your newest stamp from me"). Rejected: with three devices, a change relayed late through a third device can have an older stamp than the watermark, and it would be lost.

### 4. Collision-free IDs once a profile is linked: device slots

A synced slot table maps device ID to a slot in 0–31. A device joining a link group takes the lowest slot not in the merged table. Once a collection has a slot, its `idGen` hands out the smallest ID that is above the last ID and congruent to the slot modulo 32. IDs stay epoch milliseconds (at most 31 ms off for single additions), increasing, below 2^53 and unique across the group. Unlinked collections keep today's allocation. If two devices joining at the same time through different peers get the same slot, the merged table shows it, and the device with the larger device ID moves to a free slot.

Objects created before linking can still collide when two independent collections are merged on the first sync. The first-sync Merge (decision 7) detects collisions by identity: the same note ID with a different GUID, the same card ID with a different note or ordinal, or the same review entry ID with a different card. The device whose objects collide gives its own objects new IDs, including their references, in one local op before the exchange. Decks, presets and note types with the same ID are treated as the same object, which is the expected case for collections copied by backup.

*Alternative:* a creator ID on every object. Rejected: it costs about 20 bytes per card and review entry (about 10 MB for the scale target), and renumbering on collision across three devices needs coordination.

### 5. Deterministic clean-up after each merge

After an incoming batch is applied, the receiver runs the same deterministic clean-up that the sender would run:
- **Deck name clashes**: the deck with the larger ID is merged into the smaller one. Its cards, home-deck references and subdecks move, and it gets a tombstone that records a *redirect* to the kept deck, so cards that arrive later for the removed deck land in the kept one. When one of the two is a filtered deck and the other a regular one, they cannot be combined (a filtered deck holds no subdecks and no cards of its own), so the filtered one is renamed `Name (2)` instead.
- **Note type clashes**: if the fields and templates are equal, the note type is merged into the smaller ID (also with a redirect). Otherwise the larger ID is renamed `Name (2)`, or the next free number. **Presets** are always renamed.
- **Dangling references** that remain, such as a card whose deck is gone without a redirect, are repaired with the existing Check Database rules.

Both devices compute identical results, so the clean-up ops converge even though their stamps differ. Clean-up ops are part of the `sync` op and are not undo steps.

### 6. Settings split

A `syncedSettings` list in `core` names the synced fields: rollover hour, learn-ahead limit, auto-continue, flag names, the review display and timer settings, ignore accents, paste options, the timebox and the deck-by-note-type switch. Applying the incoming settings register copies only those fields. Theme, language, the last-used pickers, current and recent decks, and the web notice stay local. A test fails when a new `Settings` field is not classified as synced or local.

### 7. Epochs and the first-sync choice

`epoch` changes when a collection is replaced: by a restore, a `.colpkg` import, a new collection, or receiving a full replacement. A link stores both sides' epochs from its last sync. If an epoch differs, or the link is new, or it has not synced for longer than the tombstone retention (decision 10), the session stops before exchanging data, and the device that started it asks the user Merge / Replace other / Replace this (spec: First sync of a link). Until the user answers, the indicator shows "needs your choice", and automatic syncs skip that link.

- **Replace**: the receiving side makes an automatic backup (existing `backup.auto` hook). It receives the snapshot, review log and media into a new storage generation and flips the pointer, as restore does today. It then takes the sender's epoch.
- **Merge**: both sides make an automatic backup, run the collision renumbering (decision 4), and then exchange full deltas from seq 0 in both directions. Both keep their own epochs, which the link records. (A shared new epoch would change the epoch of a device that already syncs with a third one, and that link would ask the first-sync question again.)

### 8. Transport: mutual TLS 1.3 with pinned Ed25519 device keys

Each device has an Ed25519 key in `sync.json` (mode `0600`) and a self-signed certificate. The device ID is the base32 of the first 10 bytes of SHA-256(public key). Connections use TCP with TLS 1.3. Both sides present certificates, and `VerifyPeerCertificate` accepts only paired keys, plus unknown keys during the pairing window, which may then only send a pairing request. Messages are length-prefixed JSON, compressed with zstd (`klauspost/compress`), with a 16 MB limit per frame. Media blobs are streamed in chunks.

A session begins with `hello`: protocol version, snapshot version, device ID and name, wall time, and the link ID and epochs. The receiver refuses if the protocol version is newer than its own, naming the device to update (spec: Compatible versions), or if the clocks differ by more than 10 minutes after half the round-trip time is subtracted (spec: Clocks must agree). Either device can start a session. When both start one at the same time, the session started by the smaller device ID wins.

The listener uses port **47812** (one above the MCP default) and falls back to a random free port. The actual port is announced via mDNS and shown on the pairing screen.

*Alternative:* Noise protocol or a custom handshake. Rejected: `crypto/tls` is in the standard library, audited, and gives `ExportKeyingMaterial` for the pairing code.

### 9. Pairing: numeric comparison bound to the TLS session

During pairing, both sides accept the unknown certificate. Both then compute the code: six digits from `ExportKeyingMaterial("cardflip pairing", nil, 8)`. A man in the middle would run two TLS sessions with different exporter values, so the codes would differ. The receiving side's user first accepts the request. Each side then waits for its own user's confirmation and the other side's `confirmed` message, with a 2-minute timeout. Only then do both store `{deviceID, name, publicKey}`.

Unpairing removes the entry and adds the key to a `revoked` list. A revoked device can complete the handshake only to receive an `unpaired` message, after which it removes the pairing on its side (spec: Unpair while away). Pairing requests are refused outside the pairing window, and more than one pending request at a time is refused.

*Alternative:* QR codes. Deferred: they need a camera and a scanner dependency on every platform, and desktops often have no camera. The numeric code works everywhere.

### 10. Tombstone retention

Tombstones are kept until every link of the profile has synced past their seq, and at most 180 days. A link that has not synced for more than 180 days must answer the first-sync question again (decision 7), so a missing tombstone cannot bring deleted objects back silently.

### 11. Discovery and network identity in Dart, everything else in Go

- **Go `SyncService`** (bound in `fluffy.Start`): device identity, trusted networks, paired devices, links, the listener and sessions, scheduling, and a `Status` stream (listening port, advertise/pairing flags, reachable devices, per-link state, last error). The network code (identities, TLS, framing, the listener, pairing) is the package `golang/syncnet`, in `net.go` (`!js`) with `net_js.go` reporting "not supported", so that the headless test peer `golang/cmd/syncpeer` can use it too; the app's root package cannot be imported. The merge engine is in a new package `golang/sync` (pure Go, no fluffy). It works on an interface implemented by `core.Collection` (`Delta(since)`, `ApplyBatch`, `Snapshot`/`Replace`, `Epoch`).
- **Go → Dart** through a `//fluffy:dart` interface `SyncUI`: `ConfirmPairing(ctx, deviceName, code) (bool, error)`, `PairingRequest(ctx, deviceName) (bool, error)`, `PairingFinished(deviceName, paired)`, `ChooseFirstSync(ctx, linkInfo) (choice, error)`. Sync now returns its summary to the caller, so no separate `Synced` callback is needed.
- **Dart**: `bonsoir` advertises and browses `_cardflip._tcp` through the platform's own service (NsdManager on Android, Bonjour on Apple platforms, which avoids the iOS multicast entitlement, the Windows DNS-SD API, Avahi on Linux). The TXT record holds `id`, `v` and `pair=1` while pairing. Dart passes found services to `SyncService.Seen(id, host, port)`. `connectivity_plus` reports network changes. `network_info_plus` gives the Wi-Fi name, IP and mask, and Go computes the subnet for wired desktops from `net.Interfaces`. Dart calls `SyncService.SetNetwork(kind, name, subnet)`. Go decides whether the network is trusted and starts or stops the listener. Dart starts or stops bonsoir from the status stream. Go never connects or listens on an untrusted network, even if Dart misreports.
- The network identity is `wifi:<SSID>` or `subnet:<CIDR>`. Cellular networks are never trusted.
- `AppLifecycleListener` tells Go `Foreground()` / `Background()`. Go schedules automatic syncs itself: on foreground, 30 s after the last local change event, when a linked device is first seen, and on background with a 5-second budget.
- The default device name comes from `os.Hostname` on desktops and from `device_info_plus` (model name) on Android and iOS.

### 12. Serving a profile that is not open

When a peer syncs a link whose local profile is not the open one, `app` opens that profile's collection for the session through a reference-counted registry of open collections. Switching to that profile meanwhile reuses the same `*Collection`. Automatic syncs started *by* this device cover only the open profile.

### 13. Media

The collection keeps a media manifest of name → `{sha1, stamp, seq, trashed}`, journaled through a new `media` op that `AddMedia`, imports, the media trash and Check media emit. The manifest is built by hashing the stored files when a profile gets its first link, with a yield in the loop. After that it is kept up to date. Missing blobs are fetched after the object batches. Same name with different content: the entry with the smaller stamp keeps the name. The device that added the other file renames it (`stem-<sha8>.ext`) and rewrites the notes it changed since its last sync that refer to it, as an ordinary local change. Compaction treats files that a pending session has announced as referenced.

### 14. Undo interplay

The `sync` op is untracked, so it is not an undo step. After applying it, `hist` is cleared if any key the op changed or deleted appears in the before-images of an undo or redo step. Otherwise the history is kept (collection delta).

### 15. Storage files

- `<data>/sync.json` (device-level, mode `0600`): key, name, trusted networks, paired devices, revoked keys.
- `<data>/profiles/<id>/sync-links.json`: the links of that profile (peer device ID, remote profile ID, both epochs, received seq, resume cursor, echo ranges, last synced, needs-choice). It sits outside the collection backend, so it is not in backups.

Both files are read and written through new functions in `storage/profiles_files.go`, files backend only.

## Risks / Trade-offs

- [Clock skew favours a device whose clock is fast] → HLC plus the 10-minute refusal. A device that is a few minutes fast can still win close conflicts. That is accepted.
- [The Wi-Fi name needs location permission on Android, iOS, macOS 14+ and Windows 11 24H2] → It is asked for only when trusting a network. If refused, the network is identified by its subnet (spec: Wi-Fi name permission). Subnet identity is weak, because many home routers use `192.168.1.0/24`. That is acceptable because trust only controls whether the device announces itself: all data still needs pinned keys.
- [Same-named Wi-Fi networks, such as a café using a router's default name] → Same as above: at worst the device announces itself there, and no data is exchanged with unpaired devices. The announcement contains the device name. The Sync section explains this when a network is trusted.
- [mDNS blocked on guest and corporate Wi-Fi, or client isolation] → Pairing by address. The last known address of each paired device is tried as well.
- [A desktop firewall (ufw and firewalld are on by default on several Linux distributions) lets mDNS through but drops connections to the sync port] → Found in the manual test: the phone found the desktop, but pairing from the phone and every sync the phone started timed out, while the desktop's syncs with the phone worked. A connection that times out is reported as "did not answer", with the port to allow on the other device (spec: Unanswered connections), and the README shows the ufw and firewalld commands. A successful session clears the link's error whichever device started it, so a link that works in one direction does not keep showing the other direction's failure. Changes made on the phone still reach the desktop only when the desktop starts a session, so the port has to be opened.
- [Automatic syncs dial a paired device's last known address when it is not seen, which times out while it is switched off] → Such a failure is not an error for automatic syncs (spec: Automatic sync), nor is a busy refusal when both devices start a session at once. Sync now still reports both.
- [Snapshot growth from metadata (about 30 bytes per object)] → Compact array encoding. Measure against the 3-second startup target with the 50k-card fixture, and fail the task if the target is exceeded.
- [The first merge of two large collections is slow on phones] → Batches, progress in the indicator, resume cursor. The UI stays responsive because batches commit under short write locks.
- [iOS and Android suspend the app] → Sync only in the foreground, plus a best-effort sync when going to the background. Documented.
- [LWW loses one of two concurrent edits to the same note] → Accepted, as in Anki. The automatic backups before first syncs and the periodic backups are the fallback.
- [A slot table race (decision 4)] → Detected and repaired. The remaining window needs the same millisecond on two devices joining at once.
- [New dependencies (`bonsoir`, `network_info_plus`, `connectivity_plus`, `device_info_plus`, `permission_handler`)] → All are widely used Flutter plugins. Pin current versions at implementation time. `bonsoir`'s Linux support needs Avahi at runtime, which goes in the README prerequisites. `permission_handler` asks for the location permission when a network is trusted (Android and iOS); under Swift Package Manager its iOS permissions follow from the `NS*UsageDescription` keys in `Info.plist`. `network_info_plus` and `connectivity_plus` are held one patch release back, because their latest releases need `dbus` 0.8 while `bonsoir_linux` needs 0.7.

## Migration Plan

- Snapshot version 4: on load, give every object a stamp derived from its `Modified` time or, failing that, its ID time, with counter 0 and no slot. Seqs start at 0, the epoch is random, and there are no tombstones. The existing compaction on version change writes the new snapshot. Backups from version 4 are refused by older apps with the existing "update Cardflip" message. Version 4 restores version 3 backups.
- Nothing changes until the user opens the Sync section. No network code runs before a network is trusted.
- Rollback: an older app cannot open a version-4 collection. Users restore a backup made before updating, which the automatic backup thinning keeps for 8 days by default. Note this in the release notes.

## Measurements

Startup with the 50,000-card perf fixture (50,000 Basic notes imported from a TSV file, 60 answered, as in `integration_test/perf/stats_flow.dart`), measured by `core/startup_perf_test.go` (`TestStartup50k`) on the files backend, Linux desktop, 2026-10-23:

| | Open (median of 5) | Snapshot |
|---|---|---|
| Before (snapshot version 3) | 136–142 ms | 15.1 MB |
| After (version 4, every object stamped by the import) | 200 ms | 17.5 MB |
| First open of the version 3 collection (migration and compaction) | 373 ms, then 191 ms | 17.5 MB |

All are well under the 3-second target. The review log carries its sequence numbers in the segment records, so its size adds nothing to the snapshot or to startup.

## Open Questions

- Should QR codes be offered later as a faster alternative to the numeric code on phones? This does not change the protocol: the QR would carry the address and key fingerprint.
- Exact batch sizes and debounce intervals are tuning values, to be set from measurements on a mid-range Android phone.
