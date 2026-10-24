# Design

## Context

- No WebView: card HTML is laid out by `flutter_widget_from_html_core`, so MathJax itself cannot run.
- Rendering happens in Go (`render.RenderCard`), then the HTML is sanitized and CSS inlined; Dart only displays.

## Goals / Non-Goals

**Goals:**
- Typeset the TeX that typical decks use, identically on all platforms, offline.
- Never fail a card because of a formula.

**Non-Goals:**
- Full MathJax compatibility (custom macros defined in templates, `\require`, MathML input).
- Full mhchem (the subset is listed in decision 3).
- Generating images for `[latex]` (that needs a TeX installation); see the `add-latex-image-lookup` change for pre-generated images.

## Decisions

### 1. Typeset in Dart with `flutter_math_fork`

It implements KaTeX's layout in pure Dart, supports the commands most decks use (fractions, roots, matrices, `\text`, `\color`, `\boxed`, environments like `aligned`), works on web, and has no native setup.

Alternatives considered:
- *MathJax in a WebView or headless JS.* Rejected: no WebView by design, and JS is not available on native.
- *Render to SVG in Go.* Rejected: no maintained pure-Go TeX layout engine.

### 2. Go marks math, Dart lays it out

After field substitution and cloze rendering, Go finds `\(…\)` and `\[…\]` outside HTML tags and replaces each with `<cf-math display="inline|block">` containing the HTML-escaped TeX, with HTML tags inside the math removed and entities decoded (fields edited in a rich editor often contain `<br>` or `&nbsp;` inside formulas). Dart maps the element to `Math.tex`, and on a parse error shows the source in a monospace span with an error colour.

Doing the detection in Go keeps it next to cloze rendering: a cloze inside math becomes `\text{[...]}` on the question and `\color{<highlight>}{…}` on the answer, as Anki renders it.

### 3. mhchem subset translated in Go

`\ce{…}` is translated to plain TeX before it reaches Dart: element symbols upright, digits after symbols as subscripts, charges `^{2+}`, leading stoichiometric numbers, `(s)`, `(aq)` and similar states, arrows `->`, `<-`, `<->`, `<=>` with optional `[above][below]` text, `+` between species, and `\pu{}` units as upright text. Anything else inside `\ce` is passed through as `\text{…}`.

### 4. Editor menu

The Fx menu and shortcuts insert delimiters around the current selection in the focused field's `TextEditingController`, placing the cursor inside when the selection is empty. Chords (Ctrl+M then M) use a short-lived "pending chord" state in the editor's `Shortcuts`.

## Risks / Trade-offs

- [`flutter_math_fork` lacks a command used by a deck] → The error fallback shows the source for that formula only; the test corpus in `math_test.go` documents what is covered.
- [Detecting `\(` in fields that contain literal backslashes] → Only balanced pairs are converted; unbalanced delimiters stay text.
- [Large formulas overflow narrow screens] → Display math scrolls horizontally inside its own box.
