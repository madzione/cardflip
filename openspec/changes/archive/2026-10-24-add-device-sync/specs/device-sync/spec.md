# Spec Delta

## Purpose

Keeps a profile's collection, review history and media the same on the user's paired devices, so that cards can be added and studied on any of them without losing work on another.

## ADDED Requirements

### Requirement: Profile links
Sync SHALL be per profile. The user SHALL be able to link the open profile with one profile on a paired device, chosen from that device's profiles or created there as a new profile with the same name, and SHALL be able to remove the link. A local profile SHALL have at most one link per paired device.

#### Scenario: Link to a new profile
- **WHEN** the user links profile `Anna` on the laptop to "New profile" on the paired phone
- **THEN** the phone gets a profile `Anna` linked to the laptop's `Anna`

#### Scenario: Other profiles stay separate
- **WHEN** only profile `Anna` is linked and the user adds a note in profile `Ben` on the laptop
- **THEN** no sync carries that note to the phone

### Requirement: First sync of a link
The first sync of a link SHALL ask the user to merge both collections, to replace the other device's profile with this one, or to replace this profile with the other device's. Before any of these choices changes a collection, an automatic backup of that collection SHALL be made.

#### Scenario: Merge two collections
- **WHEN** the laptop has deck `Spanish` and the phone has deck `French`, and the user chooses Merge
- **THEN** both devices show `Spanish` and `French`

#### Scenario: Replace the phone's profile
- **WHEN** the user chooses to replace the phone's profile with the laptop's
- **THEN** an automatic backup of the phone's profile exists, and the phone's profile then has exactly the laptop's decks, cards and review history

### Requirement: Replacing the collection asks again
After a backup is restored, a collection package is imported or a new collection is started in a linked profile, the next sync of each of its links SHALL ask the first-sync question again instead of merging silently.

#### Scenario: Restore then sync
- **WHEN** the user restores last week's backup on the laptop and the phone then syncs
- **THEN** the user is asked whether to merge or which side to replace, and nothing is changed before the answer

### Requirement: What is synced
Sync SHALL carry decks, deck descriptions and limits, presets with their options, note types, notes, cards with their scheduling state, flags, suspension and burying, filtered decks, the review history, deletions of all of these, media files, and the collection settings that affect studying and card content. It SHALL NOT carry device settings.

#### Scenario: Synced study settings
- **WHEN** the user changes the rollover hour to 6 on the laptop and syncs
- **THEN** the phone uses 6 as the rollover hour

#### Scenario: Device settings stay
- **WHEN** the laptop uses dark mode and German, the phone uses light mode and English, and they sync
- **THEN** each keeps its own theme, language, UI size, last-used deck and note type, and MCP settings

### Requirement: Changes on both sides are combined
When both devices changed a collection since their last sync, sync SHALL keep the changes of both. For one object changed on both, such as the same note, the later change SHALL win. A card's scheduling SHALL be merged separately from its flag, suspension and deck.

#### Scenario: Different notes edited
- **WHEN** note A is edited on the laptop and note B on the phone, and they sync
- **THEN** both devices have both edits

#### Scenario: Same note edited
- **WHEN** the same note is edited on the laptop at 10:00 and on the phone at 10:05, and they sync
- **THEN** both devices have the phone's version

#### Scenario: Review and flag
- **WHEN** a card is answered Good on the phone and flagged red on the laptop before they sync
- **THEN** after syncing, the card has the phone's scheduling and the red flag on both devices

### Requirement: Review history is combined
Sync SHALL combine the review histories, so that each answer given on any linked device appears exactly once on every device. Statistics, card info and the FSRS optimizer SHALL use the combined history.

#### Scenario: Studied on both
- **WHEN** the user answers 30 cards on the laptop and 20 other cards on the phone, and they sync
- **THEN** today's statistics on both devices count 50 answers

### Requirement: Deletions are synced
A note, card, deck, preset or note type deleted on one device SHALL be deleted on its linked devices by the next sync, unless it was changed on the other device after the deletion. Notes deleted by sync SHALL be written to the deletion log of the device that deletes them.

#### Scenario: Delete propagates
- **WHEN** the user deletes 3 notes on the laptop and the phone syncs
- **THEN** the 3 notes and their cards are gone on the phone, and the phone's deletion log lists them

#### Scenario: Edited after deletion
- **WHEN** a note is deleted on the laptop at 10:00 and edited on the phone at 10:10 before they sync
- **THEN** the note exists on both devices with the phone's edit

### Requirement: Name clashes after a merge
If a merge leaves two decks with the same name, their cards and subdecks SHALL be combined into the older deck. If it leaves two presets or two note types with the same name, the newer one SHALL be renamed with ` (2)` (or the next free number), except that two note types with the same fields and card types SHALL be combined into the older one.

#### Scenario: Two Default decks
- **WHEN** two separately created collections are merged, each with a `Default` deck
- **THEN** there is one `Default` deck holding the cards of both

#### Scenario: Different note types with one name
- **WHEN** both collections have a note type `Vocab` with different fields
- **THEN** the newer one is named `Vocab (2)` on both devices

### Requirement: Media is synced
Media files used by synced notes, note types and templates SHALL be copied to the linked device. A file that Check media moves to the media trash on one device SHALL be moved to the trash on the linked device too. Files with the same name and different content SHALL both be kept, and the notes that refer to them SHALL keep showing their own file.

#### Scenario: Image added on phone
- **WHEN** the user adds a note with a photo on the phone and syncs with the laptop
- **THEN** the laptop shows the note with the photo

### Requirement: Automatic sync
While the app is open, the open profile SHALL sync with each linked device that is reachable on a trusted network: when the app starts or returns to the foreground, within 1 minute after a local change, when a linked device becomes reachable, and when the app is closed or sent to the background, if time allows. An automatic sync with a linked device that is not seen on the network and does not answer at its last known address SHALL NOT be reported as an error, since that device is away.

#### Scenario: Study on the phone
- **WHEN** the user answers cards on the phone at home while the laptop is open
- **THEN** within 2 minutes the laptop shows the reduced due counts

#### Scenario: Laptop switched off
- **WHEN** the laptop is switched off and the user answers cards on the phone
- **THEN** the phone's automatic syncs find the laptop away, and the phone's indicator shows no error

### Requirement: Sync now
The user SHALL be able to sync the open profile with all reachable linked devices at once from the sync indicator on the deck list. Afterwards, Cardflip SHALL report how many changes were received and sent, or why the sync failed.

#### Scenario: Manual sync
- **WHEN** the user taps the sync indicator while the linked laptop is reachable
- **THEN** the profile syncs, and a message says how many changes were received and sent

#### Scenario: Nothing reachable
- **WHEN** the user taps the sync indicator while no linked device is reachable
- **THEN** a message says that no linked device is reachable on this network

### Requirement: Sync indicator
When the open profile has at least one link, the deck list SHALL show a sync indicator that shows syncing, synced (with the time of the last sync), not on a trusted network, or an error. Automatic syncs that fail SHALL be shown in the indicator only, not as dialogs. A successful sync of a link SHALL clear that link's error, whichever device started it.

#### Scenario: Error kept visible
- **WHEN** an automatic sync fails because the connection drops
- **THEN** the indicator shows an error, and tapping it shows the reason and offers to sync again

#### Scenario: Synced by the other device
- **WHEN** the phone's last sync failed because the laptop did not answer, and the laptop then syncs with the phone
- **THEN** the phone's indicator shows the time of that sync instead of the error

### Requirement: Sync keeps working while studying
Sync SHALL NOT block the interface. Changes received during a study session SHALL update the session's counts and queue as changes made through the app do, and a sync SHALL NOT discard an answer the user is giving.

#### Scenario: Card deleted elsewhere during study
- **WHEN** the phone is studying deck `Spanish` and the laptop deletes the note of the next card, then syncs
- **THEN** the phone's session skips that card and continues with the next one

### Requirement: Interrupted sync is safe
A sync that is interrupted by a lost connection, a closed app or a crash SHALL leave each collection consistent, with every change that was received completely, and the next sync SHALL continue where the interrupted one stopped.

#### Scenario: Wi-Fi lost mid-sync
- **WHEN** the connection drops while the first sync of a large collection is copying cards
- **THEN** the phone opens normally afterwards, and the next sync completes the copy without duplicating anything

### Requirement: Clocks must agree
Cardflip SHALL refuse to sync when the clocks of the two devices differ by more than 10 minutes. The message SHALL name the other device and the difference, and SHALL ask the user to set the correct time.

#### Scenario: Wrong clock
- **WHEN** the phone's clock is 2 hours ahead of the laptop's and they try to sync
- **THEN** sync fails with a message that the clocks of this device and the laptop differ by 2 hours

### Requirement: Compatible versions
Cardflip SHALL refuse to sync with a device whose sync format is newer than its own. The message SHALL say which device to update. A device SHALL sync with older versions as long as their sync format is supported.

#### Scenario: Old app on the phone
- **WHEN** the laptop has a newer sync format than the phone's app supports
- **THEN** the phone says to update Cardflip on the phone, and nothing is exchanged
