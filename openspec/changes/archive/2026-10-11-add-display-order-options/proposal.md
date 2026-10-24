# Proposal

## Why

The order in which cards are shown matters for learning: new cards by deck or at random, reviews before new cards, the most at-risk reviews first when there is a backlog. Anki's Display Order section offers these choices. Cardflip always introduces new cards in the order added and interleaves reviews and new cards, which users cannot change.

## What Changes

- Deck options in a **Display order** section:
  - **New card gather order**: deck, deck then random notes, ascending position (default, today's behaviour), descending position, random notes, random cards.
  - **New card sort order**: order gathered (default), card type then order gathered, card type then random, random note then card type, random.
  - **New/review order**: mix with reviews (default, today's behaviour), show before reviews, show after reviews.
  - **Review sort order**: due date then random (default), due date then deck, deck then due date, ascending intervals, descending intervals, ascending difficulty, descending difficulty, ascending retrievability, relative overdueness.
- "Random" orders are stable within a study day, so the queue does not reshuffle after every answer.
- The options of the deck being studied apply to the whole session, including its subdecks' cards (as in Anki).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for the display order options.
- `scheduling`: "Order of new cards" refers to the gather and sort options.
- `study-session`: "Card queue order" refers to the new/review and review sort options.

## Impact

- Go: `core.DeckOptions` gains four enums (absent = defaults); `core/queue.go` replaces `newOrder` and `dueOrder` with option-driven gathering (which cards fill the daily limit) and sorting (the order within the session); `PickNext` honours new/review order; random orders are seeded by card or note ID and the study day number.
- Flutter: `deck_options_dialog.dart`.
- Tests: `golang/core/queue_test.go`, `golang/svc_test.go`.
- Overlap: `add-burying-options` also modifies "Order of new cards" (its sibling part), and `configure-learning-steps` adds interday learning ordering in the same queue code; the change landing later merges the requirement text.
