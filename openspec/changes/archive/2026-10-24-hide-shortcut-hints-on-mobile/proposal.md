# Proposal

## Why

On Android and iOS the app's menus and tooltips show keyboard shortcuts such as "Edit card (E)", "Set due date (Ctrl+Shift+D)" or a dimmed "Ctrl+Z" next to Undo. Phone and tablet users almost never have a hardware keyboard, so these hints are noise: they make menu entries longer, crowd narrow menus and point at keys the user cannot press. The only existing guard (`narrowScreen`, width < 360) hides a few of them on very small phones, so most phones still show them.

## What Changes

- On Android and iOS (native apps and the web app opened in a mobile browser), the app no longer shows keyboard shortcut hints in:
  - popup and context menus: the study screen's more menu and flag menu, the browser's selection and bulk-action menus, the app bar's Undo/Redo menu, the editor's cloze and math menus;
  - tooltips: Study deck, study Undo, Stop recording, Leave (Esc), the field editor's formatting buttons and HTML/visual toggle, and the Add/Save button;
  - Settings: the "Create backup now" entry's shortcut subtitle.
- Each of these keeps its plain label ("Edit card", "Bold", "Undo Delete 3 notes"); only the shortcut part goes away.
- The shortcuts themselves keep working on every platform, for example with a hardware keyboard on a tablet.
- Linux, Windows, macOS and the desktop web keep showing the hints as today, at every window width.
- Interface strings that embed a shortcut in the label ("Card info (I)") are split into a plain label and the shortcut, combined only where hints are shown.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `app-shell`: adds a requirement for where keyboard shortcut hints appear, by platform.

## Impact

- `lib/app/shortcuts.dart`: a single place that decides whether hints are shown and how a label and its shortcut are combined.
- Menus and tooltips in `lib/app/scaffold.dart`, `lib/features/study/study_screen.dart`, `lib/features/browse/browse_screen.dart`, `lib/features/editor/math_insert.dart`, `lib/features/editor/field_tools.dart`, `lib/features/editor/rich_field.dart`, `lib/features/editor/note_editor_screen.dart`, `lib/features/decks/deck_list_screen.dart`, `lib/features/settings/settings_screen.dart`.
- `lib/l10n/app_en.arb` and `lib/l10n/app_de.arb`: the combined "label (shortcut)" strings become plain labels plus one shared format string; generated localizations are regenerated.
- Widget tests that find menu entries or tooltips by their shortcut text. `flutter test` runs as Android by default, so they need the plain label or an explicit desktop platform.
- No change to the Go backend, data, sync or backups.
