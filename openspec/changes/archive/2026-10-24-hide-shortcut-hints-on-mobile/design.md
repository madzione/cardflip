# Design

## Context

Shortcut hints reach the screen in three shapes today:

1. **A separate label string from `lib/app/shortcuts.dart`** (`undoShortcutLabel`, `flagShortcutLabel(n)`, `clozeShortcutLabel`, `setDueDateShortcutLabel`, `backupShortcutLabel`, …), placed beside the menu text:
   - dimmed in a `Row`: `MenuEntryLabel` in `lib/features/editor/math_insert.dart` (math and cloze menus), the undo/redo items in `lib/app/scaffold.dart`, the flag menu in `lib/features/study/study_screen.dart`, invert selection in `lib/features/browse/browse_screen.dart`;
   - alone: the Add/Save button's tooltip in `lib/features/editor/note_editor_screen.dart`, the "Create backup now" subtitle in `lib/features/settings/settings_screen.dart`.
2. **A localized string with a placeholder**: `undoShortcut`, `previousCardInfoShortcut`, `setDueDateShortcut`, `deleteNoteShortcut`, `setDueDateWithShortcut`, and the editor's `boldShortcut`, `italicShortcut`, `underlineShortcut`, `subscriptShortcut`, `superscriptShortcut`, `removeFormattingShortcut`, `visualEditorShortcut`, `htmlEditorShortcut` (with `{key}` for Ctrl or Cmd).
3. **A localized string with the key typed into it**: `studyDeckShortcut` "Study deck (/)", `editCardShortcut` "Edit card (E)", `replayAudioShortcut`, `recordVoiceShortcut` and `stopRecordingShortcut` (Shift+V, "Umschalt+V" in German), `playVoiceShortcut`, `markNoteShortcut`, `unmarkNoteShortcut`, `cardInfoShortcut`, `buryCardShortcut`, `buryNoteShortcut`, `suspendCardShortcut`, `suspendNoteShortcut`, `autoAdvanceShortcut`, `stopAutoAdvanceShortcut`, `deckOptionsShortcut`, `leaveEsc`.

The only guard is `narrowScreen(context)` (width < 360) in `MenuEntryLabel` and the undo/redo items, and `narrow` for invert selection in the browser. It hides some hints on very small phones and none on most.

Most plain labels already exist in the ARB files (`undo`, `cardInfo`, `editCard`, `setDueDate`, `buryCard`, `autoAdvance`, `htmlEditor`, `studyDeck`); the others need new keys.

`flutter test` runs with `defaultTargetPlatform == TargetPlatform.android` unless a test overrides it, and many widget tests find menu entries by their combined text ("Edit card (E)").

## Goals / Non-Goals

**Goals:**
- One switch decides whether hints are shown, so a new menu entry cannot forget it.
- One shape for "label with shortcut" so every hint looks the same within its kind (inline in parentheses, or dimmed on the right).
- Translators translate the label and the key names, not the combination of each.

**Non-Goals:**
- Detecting a connected hardware keyboard and showing hints when there is one.
- Changing which shortcuts exist or how they are handled.
- Restyling the dimmed right-hand hints or moving them into parentheses, or the other way round.

## Decisions

### Decide by platform: `showShortcutHints` in `shortcuts.dart`

Add `bool get showShortcutHints` to `lib/app/shortcuts.dart`, false when `defaultTargetPlatform` is Android or iOS. On the web, Flutter derives `defaultTargetPlatform` from the browser's user agent, so a phone browser counts as mobile and a desktop browser does not. That matches the request without a web-specific branch.

Alternatives: checking for a hardware keyboard (`HardwareKeyboard`/`KeyboardListener` heuristics) is unreliable before the first key press and on the web; keeping the width rule would still show hints on most phones and hide them in narrow desktop windows, where they are useful. The `narrowScreen` checks around shortcut hints are removed in favour of the platform switch. The browser's `narrow` flag keeps controlling which actions go into its menus; only the hint inside "Invert selection" switches to the platform rule.

### Two helpers for the two shapes

- `String withShortcut(String label, String shortcut)` returns `label` when hints are hidden, and otherwise the localized `labelWithShortcut` string, `"{label} ({shortcut})"`. It covers every inline case, in menus and tooltips.
- `MenuEntryLabel(label, shortcut)` stays the widget for dimmed right-hand hints, and drops the shortcut when `showShortcutHints` is false. The undo/redo items, the flag menu and the browser's invert selection item reuse it instead of their own `Row`s. The flag item keeps its icon in front.

Where the shortcut is the whole text (Add/Save tooltip, backup subtitle), the call site omits the `Tooltip` or the subtitle when hints are hidden.

Alternative: keep the combined ARB strings and pick a plain twin on mobile. That doubles every key and leaves translators to keep two versions consistent.

### Split combined strings into label and shortcut

- Shape 3 strings become plain labels (reusing existing keys where they exist, adding e.g. `replayAudio`, `recordVoice`, `stopRecording`, `playVoice`, `markNote`, `unmarkNote`, `buryNote`, `suspendCard`, `suspendNote`, `deleteNote`, `stopAutoAdvance`, `deckOptions`, `leave`). Their keys go into `shortcuts.dart` as label getters next to the existing ones, built from `appL10n.keyShift` and friends, so German keeps "Umschalt+V". Single characters (`E`, `I`, `/`, `*`, `-`, `=`, `@`, `!`, `O`, `R`, `V`) are plain constants; Esc needs a new `keyEsc`.
- Shape 2 strings become plain labels (`bold`, `italic`, `underline`, `subscript`, `superscript`, `removeFormatting`, `visualEditor`, `previousCardInfo`), with their shortcuts as label getters in `shortcuts.dart` (moving `commandKey` from `rich_field.dart` there, or reusing `_ctrl`/`_mac`).
- The combined keys are deleted from `app_en.arb` and `app_de.arb`, and the German plain labels are taken from the current German strings without their parenthesis.

### Tests run on both platform families

Existing tests that find entries by their combined text switch to the plain label (they run as Android). Tests that are about the hint itself (cloze and math menu labels, backup subtitle, the study menu's previous card info) run with `TargetPlatformVariant.only(TargetPlatform.linux)` or a desktop `debugDefaultTargetPlatformOverride`, matching how `settings_test.dart` and `browse_test.dart` already do it. New tests cover the spec scenarios: Android and iOS show plain labels, Linux shows hints, and a narrow Linux window still shows them.

## Risks / Trade-offs

- [An iPad or Android tablet user with a keyboard loses the hints] → The shortcuts still work; the spec makes this explicit. Hardware keyboard detection can be added later behind the same `showShortcutHints` switch.
- [`withShortcut` reads `defaultTargetPlatform` and `appL10n` globally, not from the widget tree] → Same as the existing label getters in `shortcuts.dart`; menus are built on open, so a language change is picked up the next time.
- [Many test finders change at once] → Done per screen together with that screen's code, so a failing test points at one file.
- [`labelWithShortcut` adds a format that some languages may want different] → It is a translatable string, so a translation can reorder or change the brackets.

## Migration Plan

No data or settings are affected. Deleting the combined ARB keys requires regenerating `lib/l10n/app_localizations*.dart` (`flutter gen-l10n`, or the build does it via `l10n.yaml`). Rollback is reverting the change.
