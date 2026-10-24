# Design

## Context

- Strings are inline `Text('…')` across `lib/features/**`. Errors from Go arrive as messages built with `userErr(format, args…)` and are shown through `lib/app/errors.dart`.
- The relative interval labels come from `lib/widgets/time_format.dart`; some summaries are built in Go (for example import warnings).

## Goals / Non-Goals

**Goals:** one translation mechanism for all UI text; Go stays locale-agnostic.

**Non-Goals:** right-to-left UI languages in this change (layout is ready through Flutter, but no RTL translation ships), translating card content, or renaming note types, decks and presets that already exist.

## Decisions

### 1. `gen_l10n` with ARB files

Flutter's built-in generator gives typed accessors (`context.l10n.addCard`), plural and select forms, and needs no extra runtime package.

### 2. Go errors carry codes

`userErr` becomes `userErr(code, format, args…)`. The error payload sent to Dart includes `code` and `args` (as strings) along with the English message. Dart looks up `code` in the ARB messages and falls back to the English message when the code is unknown, so Go-side additions never break the UI. Import warnings and summaries use the same scheme.

Alternative: a message catalog in Go selected by the app's locale. Rejected: two translation systems, and WASM size grows with every language.

### 3. Locale resolution

`Language` empty = system: the first of the platform's preferred locales that has a translation, else English.

### 4. Built-in names come from Dart

Stock note types (name, field names, card type names), a new collection's first deck and its default preset are created in the interface language, as in Anki. Go has no translations (decision 2), so Dart sends the names in `core.BuiltinNames` through `app.SetBuiltinNames`: before the collection is opened (from the device languages, since a new collection has no language setting yet) and whenever the interface language changes. Go uses them in every place that creates one of these: a new collection, Add note type, Image Occlusion on first use, a replaced collection without note types, and Check database recreating the default deck. Names that are empty, duplicated or not allowed fall back to English for that note type. The API keeps the English stock names as identifiers; Dart shows them translated.

Stock templates are written with English field names and rewritten to the translated ones in one pass (`render.RenameFields`), so references stay correct whatever the names.

Existing note types, decks and presets keep their names: they are user data that searches (`note:`, `deck:`, `preset:`), saved searches and imported text files refer to.

Image Occlusion is recognised from the card template instead of the field names Occlusion and Image: a cloze note type whose template passes a field through `cloze:` and shows a field inside `image-occlusion-container`, as Anki's template does. The note type API reports the two fields' indexes, which the editor uses. The old name rule stays as a fallback.

Code that falls back to the Basic note type by name accepts the English name and the current built-in one.

## Risks / Trade-offs

- [Strings missed during extraction] → A lint-style test greps `lib/` for `Text('` and `Text("` with literal text outside allowlisted files.
- [German text is longer and overflows] → Run the widget tests and a manual pass on a phone layout in German.
- [Collections in several languages] → A German collection has "Einfach" next to an English "Basic" from an import. Both work; nothing matches note types by stock name except the Basic fallback, which accepts both names.
