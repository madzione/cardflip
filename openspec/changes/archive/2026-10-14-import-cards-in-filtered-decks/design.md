# Design

## Context

`apkg.Read` (`golang/importer/apkg/reader.go`) already reads every deck with a `Filtered` flag (legacy JSON `dyn`, modern protobuf kind) and every card's `DeckID` and `ODeckID` (home deck). `apkg.Convert` (`convert.go`) picks a card's deck as follows:

- if `ODeckID != 0`, use the home deck (and `ODue`);
- look the deck up with `deckName`, which returns `""` for a filtered deck;
- an empty name becomes `core.DefaultDeckName`.

So a card whose `DeckID` is a filtered deck and whose `ODeckID` is 0 lands in Default. `ConvertCollection` additionally lists all non-filtered decks so that empty ones are created too.

In a collection Anki is working with, a card in a filtered deck always has a home deck. Home-less cards in filtered decks come from old shared decks (as in the reported `Japanese_Hiragana.apkg`, a 2014 `collection.anki2`), where the export stripped the scheduling but left the cards in the filtered deck.

## Goals / Non-Goals

**Goals:**
- Home-less cards in filtered decks end up in a deck named like the filtered deck.

**Non-Goals:**
- Importing filtered decks as filtered decks (their search terms, rescheduling). That is `add-filtered-decks`.
- A warning in the import summary. The cards end up where the package's deck names say, so there is nothing for the user to act on.

## Decisions

1. **Decide per card, not per deck.** In `Convert`, a card with `ODeckID == 0` whose `DeckID` is a filtered deck takes that deck's name. The deck lookup gains a flag for "use the name even if filtered", used only for this case. A filtered deck in the same package that only holds cards with home decks therefore still creates no deck.
   - Alternative: treat every filtered deck in the package as a regular deck. That would create empty decks for ordinary filtered decks whose cards all went home, which the `.colpkg` import deliberately avoids today.
2. **No preset mapping for these decks.** `convertPresets` keeps skipping filtered decks, so the new deck gets Cardflip's Default preset. A filtered deck's `conf` is not a real preset reference.
3. **`ConvertCollection` needs no extra deck entries.** The decks of home-less cards are created through the cards' deck names, like any other non-empty deck; the explicit deck list only exists for empty decks.

## Risks / Trade-offs

- [A regular deck in the target collection already has the filtered deck's name] → the cards join that deck, the same as for any imported deck name; this is the existing merge rule for `.apkg` imports.
