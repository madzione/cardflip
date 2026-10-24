# Proposal

## Why

STEM decks write formulas as MathJax: `\(…\)` inline and `\[…\]` as display math, often with mhchem's `\ce{…}` for chemistry. Cardflip shows the raw TeX source, which makes these decks unreadable. There is also no quick way to insert math in the editor.

## What Changes

- Card display typesets `\(…\)` as inline math and `\[…\]` as centred display math, on every platform, without a WebView and without network access.
- A common subset of mhchem is supported: `\ce{…}` with formulas, charges, stoichiometric numbers, states and reaction arrows.
- Math that cannot be typeset is shown as its TeX source with an error style, and the rest of the card renders.
- Cloze deletions inside math work: the question shows `[...]` inside the formula and the answer highlights the revealed part.
- The editor gets an **Fx** menu that wraps the selection in `\(…\)`, `\[…\]`, `\ce{…}` (inside inline math), `[latex]…[/latex]`, `[$]…[/$]` or `[$$]…[/$$]`, with Anki's shortcuts (Ctrl+M M, Ctrl+M E, Ctrl+M C, Ctrl+T T, Ctrl+T E, Ctrl+T M; Cmd on macOS).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-rendering`: new requirements for MathJax typesetting, chemistry and errors.
- `card-browser`: new requirement for the editor's math insertion menu.

## Impact

- Dependency: `flutter_math_fork` (pure Dart TeX layout, all six platforms).
- Go: `golang/render` wraps math spans in a `<cf-math>` element after field substitution and cloze rendering, translating `\ce{}` to plain TeX; `sanitize.go` allows `cf-math`.
- Flutter: `lib/widgets/card_view.dart` maps `cf-math` to `Math.tex`; `lib/features/editor/math_insert.dart` holds the Fx menu, shortcuts and wrapping, used by `note_editor_screen.dart` and the browser's editor in `lib/features/browse/card_details.dart`.
- Tests: new `golang/render/math_test.go`, `golang/render/mhchem_test.go`, `test/card_view_test.dart`, `test/editor_test.dart`.
