# Design

## Context

- Cards belong to exactly one deck (`Card.DeckID`); queue building, counts and scheduling parameters all key off that deck.
- Search already returns card IDs for a query with sorting.

## Goals / Non-Goals

**Goals:** Anki-compatible semantics; cards always recoverable to their home deck, also after crashes and restores.

**Non-Goals:** nested filtered decks; filtered decks in `.apkg` export (they are exported as their home decks).

## Decisions

### 1. Home deck on the card

Building sets `HomeDeckID = DeckID` and `DeckID = filteredDeck`. Every place that needs options (`Collection.params`, leech and step logic) uses `HomeDeckID` when set. Emptying restores `DeckID` and clears `HomeDeckID`, in one op per action.

### 2. Preview mode state

When rescheduling is off, the card's scheduling state is not touched; answers record a preview review entry (so history shows them, but stats and undo-of-answers treat them separately) and set a transient `PreviewDue` on the card. Easy or emptying clears it and returns the card.

### 3. Queue

Filtered decks bypass daily limits: all their due cards are available. New cards in a filtered deck count as new; review-ahead cards are shown as reviews. Ordering uses the deck's chosen order at build time (the build stores positions); learning cards follow the normal learning priority.

### 4. Custom study

Each custom study choice produces a filter config with a generated search (`deck:"X" rated:N:1`, `deck:"X" prop:due<=N prop:due>0`, `deck:"X" added:N is:new` with preview mode, state/tag searches), replaces any existing "Custom Study Session" deck (emptying it first), and builds it. Where the generated search needs syntax that the current search lacks (for example `rated:`), the generator selects the cards directly in Go instead of through search text.

## Risks / Trade-offs

- [Cards stranded in a filtered deck after deleting a home deck] → Deleting a home deck also deletes its cards wherever they are, as Anki does; the deck deletion confirmation counts them.
- [Many places use `DeckID`] → An audit test: for every exported query (counts, browse deck column, export), cards in a filtered deck report the filtered deck for display and the home deck for options.
