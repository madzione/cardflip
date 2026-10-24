# Design

## Context

- All field editing goes through `NoteEditorScreen` (`lib/features/editor/note_editor_screen.dart`) with one `TextField` per field; the browser editor in `card_details.dart` and the study screen's edit action reuse it.
- Fields are HTML strings stored by Go; rendering sanitizes them, so the editor does not need to sanitize.
- No WebView is used anywhere in the app.

## Goals / Non-Goals

**Goals:** visual editing on all six platforms without a WebView; never lose data from imported HTML; keep stored HTML close to what Anki writes.

**Non-Goals:** tables, arbitrary CSS editing in the visual mode, collaborative or multi-field selection.

## Decisions

### 1. `flutter_quill` with an explicit HTML subset

`flutter_quill` is pure Dart, mature, supports all platforms, custom embeds (images) and keyboard shortcuts. Its document model is not HTML, so we write converters for a fixed subset: `b/strong`, `i/em`, `u`, `sub`, `sup`, `s`, `span style="color|background-color"`, `font color`, `br`, `div`/`p`, `ul/ol/li` (nested), `text-align`, indentation, `img` (src, width), `a href`.

Alternatives considered:
- *`super_editor`*: more flexible model but no maintained HTML serializer; more work for the same subset.
- *WebView with contenteditable (Anki's approach)*: rejected by the app's no-WebView design and web/desktop inconsistencies.
- *Own editor on `EditableText`*: too much work for selection, IME and undo.

### 2. Lossless guard

Before opening a field visually, convert HTML → document → HTML and compare with a normalised form of the original (whitespace and attribute order normalised). If they differ, open in HTML mode with the note "This field contains formatting the visual editor can't keep. Editing as HTML." The user can still switch, after a confirmation that formatting may be lost.

### 3. Output format

Bold as `<b>`, italic `<i>`, underline `<u>`, colour `<span style="color: …">`, line breaks `<br>`, blocks `<div>`, as Anki's editor writes them, so fields look the same when exported back to Anki.

### 4. Paste

On paste, read `text/html` from the clipboard where the platform provides it (desktop and web via the `super_clipboard` package; mobile falls back to plain text), filter to the subset and insert. Plain paste inserts text only.

### 5. Per-field option storage

`NoteType.FieldOptions []FieldOptions` parallel to `Fields`, holding `HTMLByDefault bool` now; other field options can be added later. Missing entries mean defaults, so old snapshots load unchanged. Field rename/add/remove/reorder keep the slices aligned.

## Risks / Trade-offs

- [Converter bugs alter user content] → Round-trip tests over all fields of several real shared decks; the lossless guard catches the rest by opening HTML mode.
- [Package size and startup] → `flutter_quill` is moderate in size; measure web bundle size before and after.
- [IME and accessibility in a custom editor] → `flutter_quill` builds on `EditableText` and supports IME; test with Japanese input on Android and macOS.
