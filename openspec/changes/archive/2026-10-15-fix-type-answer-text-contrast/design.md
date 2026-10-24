# Design

## Context

- Go (`golang/render`) inlines the note type's CSS into `style` attributes. `.card` colours are lifted into `Card.Background`/`Card.Color`, and black/white are dropped there so the theme takes over (`cardColors` in `css.go`). Rules for other classes, such as `.typeGood { color: black }`, are inlined on the matching elements unchanged. An invalid value such as `background-color: light green` is inlined as well, and the HTML widget ignores it when building.
- `CardView` (`lib/widgets/card_view.dart`) picks the surface `bg` and text `fg` from those card colours or the theme. It then hands the HTML to `HtmlWidget` with `customStylesBuilder`.
- `flutter_widget_from_html_core` 0.17.4 adds the styles from `customStylesBuilder` before the element's own `style` attribute (`core_build_tree.dart`, `_parseEverything`), so inline declarations win. The builder therefore can't override an inlined `color`. The cloze highlight only works because it applies when the element has no `color` of its own.
- The comparison markup comes from `CompareTyped`: `<code id="typeans">` holding `<span class="typeGood|typeBad|typeMissed">` runs and `<span id="typearrow">`. With nothing typed or very long text, it holds plain text.

## Goals / Non-Goals

**Goals:**
- Fix the readability of the comparison in `CardView`, where the real surface and text colours are known, for every place that shows an answer (study, browser details, previews).

**Non-Goals:**
- A general contrast fix for all card content. Explicit note type colours elsewhere keep winning, as "Theme-aware card display" says. This could be extended later by widening the element filter.
- Default highlight colours for `typeGood`/`typeBad`/`typeMissed` when the note type sets none. Stock "Basic (type in the answer)" CSS has no mark styles today. That is a separate gap.
- Parsing more colour formats (`hsl()`, the full CSS named-colour list). An unparseable colour or background counts as absent, which is what the widget does when drawing.

## Decisions

**Rewrite the comparison's inline colours in Dart before `HtmlWidget`.** `CardView` checks `html.contains('typeans')`. If it is there, `CardView` parses the HTML with `package:html` (already a dependency), walks the `code#typeans` subtree, and corrects each element's `color` declaration in its `style` attribute. It also sets a `color` on `code#typeans` itself when the colour it inherits is unreadable. Then it serialises the HTML and passes it on. HTML without a comparison is passed through untouched, so other cards pay nothing.
- *Alternative: `customStylesBuilder`.* Rejected because inline styles override it (see Context).
- *Alternative: fix it in Go.* Go knows `Env.Night` but not the theme's actual surface and text colours. It would have to guess "dark", and previews with explicit card colours would get it wrong.
- *Alternative: drop all mark colours in dark mode.* Rejected because it throws away readable designs like black on red.

**Effective colours.** An element's text colour is its own inline `color` if parseable, else the nearest ancestor's (the existing `_textColor` walk), else `fg`. Its background is its own parseable, non-transparent `background-color`/`background`, else the nearest ancestor's, else `bg`. Semi-transparent backgrounds are composited over the next one down, using `Color.alphaBlend`.

**Readability threshold.** Use the WCAG 2 contrast ratio, `(L1 + 0.05) / (L2 + 0.05)` from `computeLuminance()`, and require at least 3:1. That is the WCAG minimum for large text, and card text is 20 px or more. It lets saturated highlights through (black on `red` is about 5.3, black on `grey` about 5.3) and catches black on a dark surface (about 1.2). The fallback order is the card's text colour `fg` if it reaches 3:1, else black or white, whichever has the higher ratio.

**Where the helper lives.** Add `double contrastRatio(Color a, Color b)` in `css_color.dart` next to `parseCssColor`. The DOM rewrite is a private static function in `card_view.dart`, next to `_textColor`, which it reuses.

## Risks / Trade-offs

- [Re-serialising the HTML could change other markup] → Only `style` attributes inside `code#typeans` change. Tests compare the rendered output of a full answer (marks, arrow, `<hr id=answer>`, media placeholders) before and after.
- [A deliberately low-contrast design, such as a faint "hint" colour in the comparison, is overridden] → Accepted. The comparison is functional feedback, and readability matters more there.
- [Parsing on every build] → Only cards with a comparison are parsed. They are small, and the parse runs once per side shown. If profiling shows a cost, memoise on `(html, bg, fg)`.
- [`broaden-css-support` changes how inlining works] → The rewrite reads the final inline styles, whatever produced them, so it doesn't depend on selector support.
