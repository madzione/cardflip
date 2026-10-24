# Proposal

## Why

Filtered decks let users study a slice of the collection outside the normal schedule: cards failed this week, cards with a tag before a test, new cards previewed without affecting scheduling, or reviews ahead of a trip. Anki's Custom Study builds on them for its common presets. Cardflip has neither; users can only study whole decks on their schedule.

## What Changes

- **Filtered decks**: a deck built from a search, with a card limit (default 100) and an order (oldest seen first, random, increasing intervals, decreasing intervals, most lapses, order added, order due, latest added first, relative overdueness/retrievability). An optional second filter with its own search, limit and order.
- **Build / Rebuild / Empty**: building moves matching cards into the filtered deck; they remember their **home deck**. Emptying, deleting the filtered deck, or rebuilding returns cards to their home deck. Suspended, buried and cards already in another filtered deck are never pulled in.
- **Reschedule cards based on my answers** (on by default): answers count as normal reviews and use the home deck's options. When off, the deck is a **preview**: Again, Hard and Good re-show the card after configurable delays (defaults 60 s, 600 s, 0 s meaning not again), Easy returns it home, and scheduling is unchanged.
- Cards reviewed before they are due (review ahead) are scheduled from the actual time since their last review, so early reviews give appropriately shorter interval growth.
- Filtered decks show all their cards without daily limits; they appear in the deck list marked as filtered, cannot have subdecks, and cannot receive new cards from the editor.
- **Custom Study** choices that create a filtered deck named "Custom Study Session" (replacing the previous one): review forgotten cards (rated Again in the last N days), review ahead (due in the next N days), preview new cards (added in the last N days), and study by card state or tag (new/due/all cards, optional tags, a card count).
- The browser's **Create filtered deck** uses the current search.
- Search: `deck:filtered` matches cards in filtered decks.

## Capabilities

### New Capabilities

- `filtered-decks`: building, rebuilding, emptying and studying filtered decks, home decks, preview mode, and the custom study sessions that create them.

### Modified Capabilities

- `card-browser`: new requirements for creating a filtered deck from the search and for `deck:filtered`.
- `collection`: new requirement that deleting a filtered deck returns its cards.

## Impact

- Go: `core.Deck` gains `Filter *FilterConfig` (searches, limits, orders, reschedule, preview delays); `core.Card` gains `HomeDeckID` (and the original due for preview mode); `Collection.BuildFiltered`, `EmptyFiltered`; queue rules for filtered decks (no daily limits, preview delays, home deck options for steps); `DecksService` calls including custom study; `search` gains `deck:filtered`; `.apkg` import keeps putting cards from filtered decks into their home decks.
- Flutter: a filtered deck options screen, filtered deck styling and actions (Rebuild, Empty) in `deck_list_screen.dart`, custom study choices in the custom study dialog (shared with the daily-limit choices if that change exists, otherwise a new dialog from the deck menu), and a toolbar action in `browse_screen.dart`.
- Tests: `golang/core/filtered_test.go` (new), `golang/core/queue_test.go`, `golang/svc_test.go`, `golang/search/search_test.go`, widget tests for the screens.
- Overlap: `extend-daily-limits` creates the Custom study dialog with limit increases; whichever change lands second adds its choices to the existing dialog.
