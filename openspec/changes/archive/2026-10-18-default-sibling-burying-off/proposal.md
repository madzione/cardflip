# Proposal

## Why

Cardflip buries new and review siblings by default, Anki does not. With note types that make many cards per note, such as Image Occlusion (one card per shape) or long cloze notes, the default shows one card of each note per day and holds back the rest, which looks like a bug to users coming from Anki. Users who want sibling burying can still turn it on per preset.

## What Changes

- The options **Bury new siblings** and **Bury review siblings** are off by default, as in Anki. New collections and newly added presets get them off. **BREAKING** for the default behaviour: siblings are shown on the same day unless a preset buries them.
- Existing collections: a one-time migration turns both options off in the **Default** preset Cardflip created. Presets the user added, and presets imported from Anki, keep their values.
- Options stored before this change keep reading a missing burying value as on, so no other preset changes on its own.
- Bury interday learning siblings stays off by default (unchanged).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: "Burying options" changes its defaults and describes the migration of the Default preset.
- `scheduling`: "Order of new cards" gets a scenario that names the burying option instead of assuming it is on.

## Impact

- Go: `core.DefaultDeckOptions` (`golang/core/model.go`), the decoding of stored options (`DeckOptions.UnmarshalJSON`), the preset migration of old collections and backups (`golang/core/presets.go`, `golang/core/backup.go`), and the snapshot version (`snapshotVersion`).
- Flutter: the initial values in `lib/features/decks/deck_options_screen.dart`; `test/support/fake_go.dart` defaults.
- Tests that rely on burying by default: `golang/core/queue_test.go`, `golang/core/decks_test.go`, `golang/svc_test.go`, `test/deck_list_test.dart`, and the Linux/web integration flow (`integration_test/flow.dart`, which expects "1 buried card").
- Anki packages: unaffected. Imported presets carry Anki's values, and exports without presets already use Anki's defaults (`ankiDefaultOptions`).
