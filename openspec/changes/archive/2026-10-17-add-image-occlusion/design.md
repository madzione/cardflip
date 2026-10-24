# Design

## Context

- Anki stores occlusions as cloze deletions in the Occlusion field: `{{c3::image-occlusion:rect:left=.1:top=.2:width=.3:height=.1:oi=1}}`, coordinates relative to the image (0–1). Polygons use `points=x,y x,y …`, text uses `text=…:scale=…:fs=…`. `oi=1` means "occlude inactive" (Hide all, guess one). Its template relies on `anki.imageOcclusion.setup()` in a `<script>`.
- Cardflip strips scripts and renders cards through `flutter_widget_from_html_core`, with custom elements mapped to widgets.
- Cloze card generation already creates one card per number in the note's fields.

## Goals / Non-Goals

**Goals:** interchange with Anki (same note type fields and Occlusion syntax); native drawing on all platforms; an editor usable with mouse and touch.

**Non-Goals:** freehand drawing, automatic text detection (OCR), editing masks of many notes at once.

## Decisions

### 1. Render from the Occlusion field, not from the template

When the cloze field being rendered contains `image-occlusion:` deletions, the renderer emits `<cf-occlusion image="name" shapes='[…]' active="N" side="q|a">` at the place of `{{cloze:Occlusion}}` and removes the template's container markup and script. Detecting by content, not by note type name, also renders notes whose note type was renamed or cloned.

### 2. Draw in Flutter with `CustomPaint`

`occlusion_view.dart` lays out the image at the card width and paints shapes in relative coordinates: inactive shapes in the mask colour when `oi=1`, the active shape in the highlight colour on the question, and outlined on the answer. Colours come from the note type CSS variables Anki uses (`--inactive-shape-color`, `--active-shape-color`, `--inactive-shape-border`…) when present, otherwise defaults.

### 3. Editor writes Anki's syntax

The mask editor keeps shapes in relative coordinates and serialises them into the Occlusion field as cloze deletions, numbering groups 1..N in creation order. Editing an existing note parses the field back and keeps each shape's number, so its card stays with it; a removed shape's card is kept as an empty card (as for any cloze note), and new shapes get numbers that no card of the note has, so they never take over an empty card's scheduling. The Image field holds `<img src="file">`, with the image stored as media through the normal media path.

### 4. Note type availability

The note type is added to new collections. For existing collections, it is created on first use (with a fixed name; if the name is taken by a different type, a numbered suffix is used), so the change needs no migration.

## Risks / Trade-offs

- [Anki changes the occlusion syntax] → Unknown shape kinds or properties are ignored and the rest still renders; fixtures from current Anki pin the format.
- [Editor usability on phones] → Touch handles are at least 24 px; pinch-to-zoom is supported in the editor canvas.
- [Large images] → Images are decoded at display size; the editor works on a downscaled preview and stores the original.
