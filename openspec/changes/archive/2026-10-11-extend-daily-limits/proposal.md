# Proposal

## Why

The most common Anki custom-study request is "let me do 10 more new cards today" without changing the deck's settings. Cardflip's limits are fixed per deck, so users raise the limit and forget to lower it again. Anki also lets limits apply for today only, decides whether the review limit also caps new cards, and can apply a parent deck's limits when studying a subdeck.

## What Changes

- **Today-only limits**: a deck's new and review limits can be overridden for the current study day; the override expires at the next rollover.
- **Custom study** dialog from the deck's menu with **Increase today's new card limit** and **Increase today's review card limit** by a number. It sets a today-only limit equal to the current effective limit plus that number. The dialog shows how many more cards are available.
- Deck option **New cards ignore review limit** (on by default, today's behaviour): when off, new cards are only shown while the review limit has room, as in Anki's default.
- Deck option **Limits start from top** (off by default): when on and the user studies a subdeck, the limits of its parent decks also apply.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for today-only limits and the two limit options.
- `scheduling`: "Due cards and daily limits" covers today-only limits and the two options.
- `app-shell`: new requirement for the Custom study dialog on the deck list.

## Impact

- Go: `core.Deck` gains `TodayLimits{Day int, New, Review *int}` (ignored when `Day` is not today); `core.DeckOptions` gains `NewIgnoresReviewLimit` (stored inverted, absent = on) and `LimitsStartFromTop`; `core/queue.go` (`limited`) uses effective limits and the two options; `DecksService.CustomStudyInfo` and `IncreaseToday(deckID, newBy, reviewBy)`.
- Flutter: new `lib/features/decks/custom_study_dialog.dart`, a deck menu entry in `deck_list_screen.dart`, the two options in `deck_options_dialog.dart`.
- Tests: `golang/core/queue_test.go`, `golang/svc_test.go`, `test/deck_list_test.dart`.
- Overlap: `add-filtered-decks` adds further custom study choices to the same dialog; `add-deck-overview` can add a Custom study button that opens it; `add-deck-presets` adds "This deck" limits next to today-only ones.
