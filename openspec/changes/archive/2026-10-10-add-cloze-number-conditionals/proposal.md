# Proposal

## Why

Cloze templates in shared decks often show extra content only on certain cloze cards, using `{{#c1}}…{{/c1}}` (render only on card 1) and `{{^c1}}…{{/c1}}`. Cardflip treats `c1` as an unknown field name, so these sections render an "unknown field" marker or the wrong content on every card.

## What Changes

- In cloze note types, a section named `cN` (N a positive number) is a cloze-number condition: `{{#cN}}…{{/cN}}` renders only when the card being rendered is cloze card N, and `{{^cN}}…{{/cN}}` renders only when it is not.
- A real field named `c1` takes precedence over the cloze-number meaning, as in Anki.
- In standard note types the sections keep their current meaning (a field lookup).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirement for cloze-number conditional sections.

## Impact

- Go: `golang/render/template.go` (`context.render` for `sectionNode`, using the existing `clozeOrd` in the render context).
- Tests: `golang/render/template_test.go`.
- No storage or API change.
