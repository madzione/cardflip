# Tasks

## 1. Selector engine

- [x] 1.1 Add an element tree builder for sanitized card HTML in `golang/render` and a serializer that writes merged `style` attributes; existing `render` tests must still pass. Verify `go test ./render/...` passes.
- [x] 1.2 Compile selectors (compounds, combinators, ids, attributes, universal, the listed pseudo-classes, selector lists) and match right-to-left with a rule index by rightmost key. Add `css_test.go` cases for each selector kind and for unsupported selectors being ignored. Verify they pass.
- [x] 1.3 Implement the cascade (importance, specificity, order, inline styles). Add `css_test.go` cases for the specificity scenario, `!important`, and an inline style beating a class rule. Add a benchmark with 2,000 rules and 500 elements and verify it stays under 10 ms on a typical desktop.

## 2. Images and backgrounds

- [x] 2.1 Honour inlined `width`, `height`, `max-width`, `max-height` and attributes in the `card_view.dart` image builder, with horizontal scrolling for overflowing images. Add `test/card_view_test.dart` tests for a fixed width and `max-width: none`. Verify they pass.
- [x] 2.2 Show `background-image` media behind elements through the widget factory, ignoring remote URLs. Add a test with the card background scenario. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Compare three popular shared decks side by side with Anki on desktop.
