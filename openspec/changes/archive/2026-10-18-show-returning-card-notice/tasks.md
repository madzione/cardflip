# Tasks

## 1. Go

- [x] 1.1 Add `DueEarly` to `StudyCard` and set it in `studyCard` for learning cards shown before their due time (design decision 1). Add `svc_test.go` cases: a new card answered Good with learn ahead on comes back with `DueEarly` 10 minutes ahead; an already due card has none. Run `fluffy generate`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 2. Study screen

- [x] 2.1 Add the shown-again flag and the notice to `study_screen.dart` (design decisions 2 and 3), and `dueEarly` to the study card in `test/study_test.dart`. Add `test/study_test.dart` tests for each scenario of "Card shown again": shown early, due again now, gone after Show answer, no notice for another card (a sibling), and none after undo. Verify `fvm flutter test` passes.
- [x] 2.2 Show the notice on any card shown early, not only the card just answered (design decision 3, `_ReturningCardNotice`): "Shown early: …" for another card, "This card again: …" for the card just answered. Add `test/study_test.dart` tests for the scenarios of "Card shown early or again" not covered yet: shown early (another card, also the first card of a session), and no notice for another card that is due. Verify `fvm flutter test` passes.

## 3. Integration check

- [x] 3.1 Run `tool/check.sh` and verify all checks pass. Study the Image Occlusion fixture deck with burying on and check that the notice explains the cards coming back early, whether the same card or the other note's.
