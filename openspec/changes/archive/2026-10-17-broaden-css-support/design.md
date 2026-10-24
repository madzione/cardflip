# Design

## Context

`css.go` parses rules with `tdewolff/parse/css` and inlines matching declarations by streaming over HTML tokens, which is why only element and class selectors work: a streaming pass does not know ancestors or siblings.

## Goals / Non-Goals

**Goals:** browser-like matching for the listed selectors; deterministic cascade; bounded cost for typical cards (a few hundred elements, a few hundred rules).

**Non-Goals:** pseudo-elements, dynamic pseudo-classes, `:has()`, CSS variables, inheritance computed in Go (the HTML widget already inherits text properties from parents).

## Decisions

### 1. Build a tree, then match right-to-left

Parse the sanitized HTML into a lightweight element tree (tag, id, classes, attributes, parent, previous sibling, index among siblings). Each selector is compiled into compounds with combinators and matched right-to-left from each element, like browsers do. The tree is serialized back with merged `style` attributes.

Alternative: switch to `golang.org/x/net/html` and a third-party selector library (`cascadia`). Viable, but adds dependencies to a WASM build that is size-sensitive; the subset here is small enough to own. Revisit if the subset grows.

### 2. Cascade

Collect matching declarations per element with (importance, specificity, order); sort; last wins per property; inline `style` attributes from the field content sit above normal author rules and below `!important`, as in browsers.

### 3. Images

The card view's image builder reads inlined `width`, `height`, `max-width`, `max-height` (px and %), and the `width`/`height` attributes. Default behaviour stays "scale to fit the card width". `max-width: none` with a larger image lets the image overflow into a horizontal scroll area.

## Risks / Trade-offs

- [Pathological CSS (thousands of rules) slows rendering] → Index rules by their rightmost compound's id, class and tag, as browsers do; benchmark with a 2,000-rule stylesheet in tests.
- [Behaviour change for decks that relied on rules being ignored] → This brings rendering closer to Anki, which is the reference.
