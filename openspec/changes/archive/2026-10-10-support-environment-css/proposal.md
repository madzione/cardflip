# Proposal

## Why

Anki puts classes on the card that describe where it is shown: `.card1`, `.card2`… for the card type, `.nightMode` in dark mode, and `.win`, `.mac`, `.linux`, `.android`, `.iphone`, `.ipad`, `.mobile` for the platform. Shared decks use them to style reverse cards differently, to fix colours in dark mode and to enlarge text on phones, often together with `@media` queries. Cardflip ignores all of these, so dark-mode rules are lost (text can become unreadable) and per-card-type styling disappears.

## What Changes

- The card root has the classes `card`, `cardN` (N = card type number, starting at 1; for cloze cards the cloze number), `nightMode` and `night_mode` in dark mode, and the platform classes for the device it is shown on.
- CSS rules that select these classes apply: on the card itself (`.card1`, `.card.nightMode`, `.mobile`) and as an ancestor prefix (`.nightMode .cloze`, `.card2 b`, `.mobile .front`).
- `@media (prefers-color-scheme: dark|light)` follows the app's theme, and `@media` width conditions are evaluated against the card's display width.
- Changing the theme or rotating the device re-renders the shown card with the matching rules.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirements for card root classes and environment media queries.

## Impact

- Go: `golang/render/css.go` (root classes as compound and ancestor selectors, `@media` evaluation), `golang/render/render.go` (a render environment: night mode, platform, width bucket), the services that render cards accept the environment, which Dart sets once via `AppService.SetRenderEnvironment`.
- Flutter: `lib/app/app.dart` or `lib/data/scope.dart` reports theme brightness, platform and card width changes; `lib/widgets/card_view.dart` re-requests HTML when they change.
- Tests: `golang/render/css_test.go` (new), `golang/svc_test.go`, `test/card_view_test.dart`.
- Overlap: `broaden-css-support` generalises selectors; both touch `css.go` matching and should be rebased on each other.
