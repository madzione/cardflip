# Proposal

## Why

Families and classrooms often share one device, and some users keep separate collections for unrelated subjects. Anki supports several profiles, each with its own collection and settings. Cardflip has one collection per device or browser, so sharing a tablet means sharing (and mixing) everyone's cards.

## What Changes

- **Profiles**: each profile has its own collection, media, settings, backups and review history, fully separated from the others.
- A **Switch profile** entry in Settings opens the profile list: open, add, rename and delete profiles. Deleting asks for confirmation naming the profile and its card count, and cannot delete the open profile.
- On start, the app opens the last used profile. A setting "Show profile list at startup" (off by default) shows the list instead.
- The current profile's name is shown in Settings; with more than one profile, it is also shown on the deck list.
- Existing installations become a profile named "Default" without moving data.
- `.cardflip` backups contain one profile's collection; restoring replaces only the open profile.

## Capabilities

### New Capabilities

- `profiles`: several separate collections on one device and switching between them.

### Modified Capabilities

None.

## Impact

- Go: `golang/app.go` opens a collection from a profile-specific data location (native: a subdirectory per profile, the existing data directory serving as "Default"; web: an IndexedDB database per profile, the existing one serving as "Default"); a small profile registry stored outside any collection; `AppService` gains list, add, rename, delete and switch; switching closes the collection, resets services' state (study session, staged imports) and opens the other.
- Flutter: profile list screen, Settings entry, startup flow in `lib/app/bootstrap.dart`, profile name on the deck list.
- Tests: `golang/svc_test.go` (separation, switch, delete), `golang/storage` tests for per-profile locations, `test/app_shell_test.dart`, `test/settings_test.dart`, integration test switching profiles on Linux and web.
