# Design

## Context

- Anki (rslib `latex.rs`) extracts each LaTeX block, normalises it (`<br>` to newline, other HTML stripped, entities decoded), wraps `[$]` as `$…$` and `[$$]` as `\begin{displaymath}…\end{displaymath}`, and names the image `latex-` + hex SHA-1 of that text + `.png` or `.svg`.
- Media lookup and missing-media placeholders already exist for `<img>`.

## Goals / Non-Goals

**Goals:** produce exactly Anki's file names so existing pre-generated images are found.

**Non-Goals:** running TeX; the note type's LaTeX header and footer (they affect generation, not the file name).

## Decisions

### 1. Reproduce Anki's naming, verified by fixtures

Port the normalisation and naming from rslib and pin it with test fixtures created in Anki (a deck with `[latex]`, `[$]` and `[$$]` blocks, including HTML and entities inside), asserting our names equal the media file names Anki generated.

### 2. Replace with `<img>`, fall back to source

Render `[latex]` blocks as `<img class=latex src="latex-….png">` when the media exists (checked through the collection's media index) and as `<pre class=latex>` with the escaped source otherwise. Checking existence at render time avoids a broken-image placeholder for decks that never shipped images.

### 3. SVG through `flutter_svg`

Images ending in `.svg` use `SvgPicture.memory` in the card view's image builder. This also benefits any other SVG media.

## Risks / Trade-offs

- [Anki changes its normalisation in a future version] → The fixtures catch a mismatch when fixtures from a newer Anki are added; older decks keep the old names anyway.
- [Malicious SVG] → `flutter_svg` does not execute scripts; external references are not resolved.
