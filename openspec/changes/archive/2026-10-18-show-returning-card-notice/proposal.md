# Proposal

## Why

When nothing else is due, a learning card within the learn ahead limit (20 minutes by default) is shown early, without a word, while the counts show nothing due. When it is the card just answered, the study screen flips straight back to its question; when it is another card of the same image (Image Occlusion notes with buried siblings), it looks almost the same. Either way users think their click did not register, and that cards come back before their time, as reported while studying Image Occlusion notes.

## What Changes

- When a card is shown before it is due because nothing else is due, the study screen shows a notice on its question that says so and when the card is due, for example "Shown early: due in 11 minutes, because nothing else is due."
- When the card is the one just answered, the notice says so: "This card again: due in 11 minutes, shown early because nothing else is due." When that card is already due, it says it is due again now.
- The notice does not cover the card or the buttons, and goes away when the answer is shown or another card comes.
- Go tells the app when a card is shown before it is due, and its due time.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `study-session`: new requirement for the notice on a card shown early, or shown again right after its answer.

## Impact

- Go: `StudyCard` in `golang/svc_study.go` gains the due time of a learning card shown before it is due (`studyCard`); `golang/svc_test.go`. Run `fluffy generate`.
- Flutter: `lib/features/study/study_screen.dart` (the notice, next to the empty-card notice), `test/study_test.dart`, `test/support/fake_go.dart`.
- Not affected: undo, which brings back the previous card on purpose, does not say "this card again".
