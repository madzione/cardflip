# Proposal

## Why

When a note or template edit makes a card's question empty, Cardflip deletes that card immediately on save, with its scheduling and position in the review history context. A mistaken template edit (for example a typo in a field name in a note type with thousands of notes) can therefore delete thousands of cards at once, without a confirmation. Anki deliberately keeps such cards and offers an **Empty Cards** tool that lists them and deletes them only after the user confirms.

## What Changes

- **BREAKING (behaviour)**: editing a note or a note type no longer deletes cards whose question becomes empty, or cards whose cloze number disappears from the note. Those cards stay, with their scheduling, as **empty cards**. Cards are still created when a template or cloze number starts producing content, as today.
- Deleting a card type in the note type editor still deletes its cards after the existing confirmation.
- **Empty Cards** in Settings lists notes with empty cards (note, card type or cloze number) with the total count, and deletes them after confirmation. Notes whose every card is empty keep one card, so no note is deleted silently; the list marks them.
- An empty card shown during study displays a notice that the card is empty and can be removed with Empty Cards, with its normal rating buttons.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: "Notes generate cards" keeps cards whose question becomes empty; new requirement for the Empty Cards tool.
- `study-session`: new requirement for showing empty cards.
- `app-shell`: new requirement for the Empty Cards entry.

## Impact

- Go: `syncCards` in `golang/core/cards.go` and `notes.go` stops deleting cards whose ordinal no longer renders; `Collection.EmptyCards()` lists them; `Collection.DeleteEmptyCards(ids)` deletes them in one op, keeping at least one card per note; `StudyCard` gains `Empty bool`.
- Flutter: `settings_screen.dart` (Empty Cards dialog), `study_screen.dart` (empty card notice).
- Tests: `golang/core/notes_test.go` (the cloze-removal test changes), `golang/core/cards_test.go`, `golang/svc_test.go`, `test/study_test.dart`, `test/settings_test.dart`.
- Behaviour change for users: removing a cloze deletion no longer removes its card until Empty Cards is run.
