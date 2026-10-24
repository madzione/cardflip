# Proposal

## Why

Choosing a deck in Cardflip starts studying immediately. Users don't see what is waiting (and how much is buried), they can't read the deck's description that shared decks ship with, and there is no place for deck-level actions before studying. In particular, cards buried by mistake can only come back at the next rollover. Anki solves this with a deck overview screen.

## What Changes

- Choosing a deck opens an overview: the deck name, its new, learning and review counts for today, the number of buried cards, and its description.
- The overview has **Study now**, **Options**, and, when cards are buried, **Unbury**, which makes all buried cards in the deck and its subdecks available again immediately.
- When nothing is due, the overview says so and when the next card is due, which replaces the separate "nothing due" screen as the first thing shown.
- Each deck can have a description (plain text or simple HTML), edited from the overview; links open in the system browser.
- `.apkg` import keeps deck descriptions.
- Starting a study session from elsewhere (the learning-due reminder, the deck list's study action) still goes straight to studying.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-shell`: new requirement for the deck overview.
- `study-session`: "Start a session" changes so that choosing a deck shows the overview first.
- `collection`: new requirements for deck descriptions and unburying.
- `data-import`: new requirement that `.apkg` import keeps deck descriptions.

## Impact

- Go: `core.Deck` gains `Description string` (`omitempty`); `Collection.SetDeckDescription`, `Collection.Unbury(deckID)` (clears manual buries in the subtree and the sibling burying for today); `DecksService.Overview(deckID)` returns counts, buried count, description and next due; `importer/apkg` reads `desc` from deck JSON.
- Flutter: new `lib/features/decks/deck_overview_screen.dart`; `deck_list_screen.dart` opens it on tap; the study screen's "nothing due" stays for sessions started directly.
- Tests: `golang/core/decks_test.go`, `golang/core/queue_test.go`, `golang/svc_test.go`, `golang/importer/apkg/apkg_test.go`, `test/deck_list_test.dart`, new `test/deck_overview_test.dart`.
- Overlap: sibling burying is computed from today's answers (not stored), so "Unbury" needs a per-day "unburied siblings" marker for the deck tree; see design.
