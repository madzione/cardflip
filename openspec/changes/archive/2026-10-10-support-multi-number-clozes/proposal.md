# Proposal

## Why

Anki lets one cloze deletion belong to several cards: `{{c1,2,3::Paris}}` hides "Paris" on cards 1, 2 and 3. Cardflip's cloze parser accepts only a single number after `c`, so such deletions are not recognised at all: imported decks that use them show the raw `{{c1,2::…}}` markup and generate too few cards. This blocks imported cloze decks from rendering correctly.

## What Changes

- The cloze parser accepts `{{cN,M,…::answer}}` and `{{cN,M,…::answer::hint}}`: a comma-separated list of positive numbers, with optional spaces around commas.
- Such a deletion is treated as a deletion for every listed number: hidden on the question side of each of those cards and highlighted on their answer side, and shown as plain text on other cards.
- Card generation counts every listed number, so `{{c1,2::x}}` alone creates two cards.
- Nested deletions and hints work with multi-number deletions as with single-number ones.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirement for multi-number cloze deletions.
- `collection`: new requirement that multi-number deletions generate one card per listed number.

## Impact

- Go: `golang/render/cloze.go` (`clozeOpen` returns a set of numbers, `clozePiece.ord` becomes a list, `writeCloze`, `ClozeNumbers`), no changes outside `render`.
- Tests: `golang/render/cloze_test.go`, a card-generation case in `golang/core/notes_test.go`, and an `.apkg` fixture note in `golang/importer/apkg/apkg_test.go`.
- No storage or API change.
