# Proposal

## Why

Cardflip can search by deck, tag, note type and a few states, but not by the properties users need to manage their reviews: cards with long intervals, many lapses, low retrievability or high difficulty, buried cards, a particular card type, notes without tags, or specific IDs. Anki's `prop:`, `card:`, `is:buried` and related searches cover these and are widely used in guides and saved searches.

## What Changes

- `prop:` comparisons with `=`, `!=`, `<`, `<=`, `>`, `>=` on: `ivl` (interval in days), `due` (days from today; negative = overdue), `reps`, `lapses`, `pos` (new-card position), `s` (stability in days), `d` (difficulty on Anki's 0–1 scale), `r` (current retrievability 0–1). `prop:ease` reports that Cardflip uses FSRS difficulty (`prop:d`) instead.
- `is:buried` (any burying), `is:buried-sibling`, `is:buried-manually`.
- `is:learn` also matches relearning cards and `is:review` also matches relearning cards, as in Anki, so `is:learn is:review` finds lapsed cards being relearned and `is:review -is:learn` finds plain review cards.
- `card:<name>` by card type name (wildcards allowed) and `card:<n>` by card type number (cloze number for cloze notes).
- `tag:none` for notes without tags.
- `deck:current` for the deck most recently studied, and `deck:name::*` for only the subdecks of a deck (so `deck:french -deck:french::*` is the deck without its subdecks).
- `nid:` and `cid:` with one ID or a comma-separated list.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for property, burial, state, card type, tag, deck and ID searches.

## Impact

- Go: `golang/search` term kinds and comparison parsing; `search.Card` gains interval, due day offset, reps, lapses, position, stability, difficulty, retrievability, buried kinds, card ordinal and card type name, note and card IDs; `core.Settings` gains `CurrentDeckID` (set when a study session starts).
- Tests: `golang/search/search_test.go`, `golang/core/search_test.go`.
- Not included: Anki's custom-data searches (`has-cd:`, `prop:cdn:`, `prop:cds:`), because Cardflip has no custom scheduling data.
