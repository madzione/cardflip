# Design

## Context

- `OcclusionEditor` and `OcclusionController` (`lib/features/editor/occlusion_editor.dart`) are self-contained: the widget takes a controller and image bytes and already has all the shape tools, zoom, mouse and touch handling.
- The code around it lives only in `_NoteEditorScreenState` (`lib/features/editor/note_editor_screen.dart`): building the controller with `keepNumbers` and the note's `cardOrds + 1` as `cardNumbers` (`_resetFields`), `_loadOcclusionImage` (the `<img src>` of the Image field via `loadMediaBytes`), `_pickOcclusionImage` (picker → `golang.media.add` → new `<img>` HTML), `_occlusionSection` (title row, Change image / Choose image, editor or placeholder), the save check (no image → `chooseImageToDrawOn`, no shapes → `drawShapes`), and writing `occlusion.field` / image HTML back into the field list. `_shownAsField` hides the Occlusion and Image fields.
- `NoteFieldsEditor` (`lib/features/browse/card_details.dart`) builds one `RichField` per field from `golang.CardDetails.note` / `.noteType`, which already carry `imageOcclusion`, `occlusionField`, `imageField` and `cardOrds`. It is keyed by note id, so it survives the pane's reloads after a save. `CardDetailsPane.edit()` focuses `firstFieldFocusNode`, which is wired to field 0, the Occlusion field.

## Goals / Non-Goals

**Goals:**
- One implementation of the occlusion section, used by both editors, so they cannot drift.
- The note editor screen's behavior and widget keys stay as they are, so `test/editor_test.dart` passes unchanged.

**Non-Goals:**
- Opening masks in a separate full-screen editor from the browser.
- Editing masks of several selected notes at once.
- Changes to the mask editor's tools themselves.

## Decisions

**Inline mask editor, not a button to the note editor screen.** The pane is where the user already edits the note's fields; a button that pushes `NoteEditorScreen(noteId:)` would split one note across two editors with separate saves. The canvas sizes itself to the available width (`LayoutBuilder`, `maxHeight` 480), so it fits the resizable split and the narrow `CardDetailsScreen`. Alternative considered: an "Edit masks" button only — less code, but the browser editor would still show raw Occlusion/Image fields, which is the problem being fixed.

**Extract an `OcclusionFields` mixin plus `OcclusionSection` widget** into a new `lib/features/editor/occlusion_fields.dart`, following the `FieldTools` pattern (a mixin on `State` used by both editors):
- The mixin owns `OcclusionController? occlusion`, `String imageHtml`, `Uint8List? occlusionImage`; offers `initOcclusion(type, rawFields, {editing, cardNumbers})`, `disposeOcclusion()`, `loadOcclusionImage()`, `pickOcclusionImage()`, `bool shownAsField(int i)`, `String? occlusionProblem()` (the save check) and `applyOcclusion(List<String> fields)`; it reports errors through an abstract `showOcclusionError(String)`.
- `occlusionSection(BuildContext)` returns the existing section UI with the existing keys (`occlusion-section`, `occlusion-pick-image`).
- `NoteEditorScreen` switches to it with no visible change; `NoteFieldsEditor` uses it with `editing: true`.
Alternative considered: a standalone `StatefulWidget` owning image state — cleaner encapsulation, but the parent's save must read the controller and image HTML and validate, so the state would have to be lifted back out anyway; a mixin matches how `FieldTools` already shares editor behavior.

**Field 0 focus.** `NoteFieldsEditor` attaches `firstFieldFocusNode` to the first field that `shownAsField` accepts (Header for the stock type) and starts `_lastField` there, so Edit and the field toolbar target a visible field.

**Keys in the browser.** The browser editor uses the same occlusion keys as the note editor; only one editor is on screen at a time, and tests can reuse the `occlusion-*` helpers from `editor_test.dart`.

**Errors in the pane** use the pane's existing inline `_error` text; no new strings are needed.

## Risks / Trade-offs

- [The pane reloads on every collection change, rebuilding `NoteFieldsEditor` with new `note` data] → it is keyed by note id and reads its initial values once (`late final`), so unsaved shapes are not overwritten; this matches how text fields behave there today.
- [`cardNumbers` are captured when the editor is created; after a save adds a card, a later new shape must still avoid it] → `OcclusionController._nextGroup` only grows, so numbers handed out in the session are never reused.
- [Mouse-wheel scrolling over the canvas in the details pane's `SingleChildScrollView`] → only Ctrl+wheel is claimed by the canvas; plain wheel still scrolls the pane, as in the note editor.
- [Refactoring the note editor screen could change its behavior] → keep keys and messages identical and run `test/editor_test.dart` and `test/occlusion_editor_test.dart` before and after.
