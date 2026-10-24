# Tasks

## 1. Contrast helper

- [x] 1.1 Add `double contrastRatio(Color a, Color b)` (WCAG 2: `(L1 + 0.05) / (L2 + 0.05)`, lighter over darker, via `computeLuminance()`) to `lib/widgets/css_color.dart`. Extend the "CSS colours" test in `test/card_view_test.dart` with ratios for black/white (21), black/`#ff0000` (≈5.25) and a colour against itself (1). Verify with `fvm flutter test test/card_view_test.dart`.

## 2. Comparison colour rewrite

- [x] 2.1 In `lib/widgets/card_view.dart`, add a private static function that takes the HTML, `bg` and `fg`. It returns the HTML unchanged when it contains no `typeans`. Otherwise it parses the HTML with `package:html`, walks `code#typeans` and its descendants, and works out each element's effective text colour (reuse `_textColor`) and effective background (own parseable non-transparent `background-color`/`background`, else the nearest ancestor's, else `bg`, with alpha composited). When the contrast is below 3:1, it sets the element's inline `color` to `fg`, or to black/white when `fg` is also below 3:1. It then serialises the result. Call it in `build` before `_surface` so `HtmlWidget` gets the rewritten HTML. Verify with `fvm flutter analyze`.
- [x] 2.2 Add widget tests in `test/card_view_test.dart`, one per spec scenario. The HTML should be what Go produces, with the classes' styles inlined: dark theme with `typeGood` `color:black;background-color:light green` gets a light text colour; `typeBad` `color:black;background-color:red` stays black; light theme with `typeGood` `color:black` stays black; `color:#ddd;background-color:#eee` with card colour `white` becomes black; and a `color:black` span outside `code#typeans` in dark mode stays black. Read colours from the `RichText` spans as the existing dark-mode test does. Verify with `fvm flutter test test/card_view_test.dart`.
- [x] 2.3 Add a test that an answer with marks, the `typearrow` span, an `<hr id=answer>` and a `cf-sound` placeholder still renders the same widgets (sound button, both text runs) after the rewrite, and that HTML without `typeans` is passed to `HtmlWidget` unchanged. Verify with `fvm flutter test test/card_view_test.dart`.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass.
- [x] 3.2 In the app, switch to dark mode and study a "Hiragana Note" card. Type a partly wrong answer, reveal it, and check that the correct characters are readable in the card's orange, while the red and grey marks keep black text on their highlights. Switch to light mode and check that the comparison looks as before.

## Workflow follow-up

- Archive the change with `/opsx:archive fix-type-answer-text-contrast` and tick it in `ROADMAP.md` (phase 0).
