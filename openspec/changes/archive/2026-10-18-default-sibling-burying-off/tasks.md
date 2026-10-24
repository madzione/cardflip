# Tasks

## 1. Defaults and migration

- [x] 1.1 Set `BuryNew` and `BuryReview` to false in `core.DefaultDeckOptions`, keep "missing means on" in `DeckOptions.UnmarshalJSON`, and compare version 1 decks with the old defaults in `migratePresets` (design decisions 1 and 2). Add `core` tests: a new collection's Default preset and an added preset bury nothing; a version 1 deck without burying keys keeps burying on. Update the tests that relied on burying by default (`queue_test.go`, `decks_test.go`) to set it on where they test burying. Verify `go test ./core/...` passes.
- [x] 1.2 Raise `snapshotVersion` to 3 and turn burying off in the Default preset when opening an older snapshot and when restoring an older backup, other presets unchanged (design decision 3). Add tests for both, using a version 2 snapshot with a user preset that buries siblings. Verify `go test ./...` passes.

## 2. App

- [x] 2.1 Match the initial burying values in `deck_options_screen.dart` and the defaults in `test/support/fake_go.dart`; update `test/deck_list_test.dart` and the integration flow (`integration_test/flow.dart`, which expects "1 buried card") to the new default. Verify `fvm flutter test` passes.

## 3. Integration check

- [x] 3.1 Run `tool/check.sh` and verify all checks pass. Run the Linux integration flow (`fvm flutter test integration_test/app_flow_test.dart -d linux`) and verify it passes. Open an existing collection and check that its Default preset no longer buries siblings while a user preset still does.
