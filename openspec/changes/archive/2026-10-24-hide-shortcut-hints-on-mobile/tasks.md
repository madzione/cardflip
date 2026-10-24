# Tasks

## 1. Hint switch and helpers

- [x] 1.1 Add `showShortcutHints` (false on Android and iOS), `withShortcut(label, shortcut)` and the `labelWithShortcut` ARB string `"{label} ({shortcut})"` (en and de) to `lib/app/shortcuts.dart` and the ARB files, plus `keyEsc`; regenerate localizations with `fvm flutter gen-l10n` and verify `fvm flutter analyze` is clean
- [x] 1.2 Add label getters in `lib/app/shortcuts.dart` for the shortcuts that are typed into strings today (E, R, Shift+V, V, *, I, -, =, @, !, Shift+A, O, /, Esc) and for the editor's Ctrl/Cmd+B, I, U, =, Shift+=, R and Shift+X, moving `commandKey` out of `rich_field.dart`; verify `fvm flutter analyze` is clean
- [x] 1.3 Make `MenuEntryLabel` drop the shortcut when `showShortcutHints` is false instead of on `narrowScreen`, with an optional leading widget for the flag icon; verify the existing editor menu tests in `test/editor_test.dart` pass once they run as Linux (`TargetPlatformVariant.only(TargetPlatform.linux)`) and add an Android case that finds "Inline math" without "Ctrl+M, M"

## 2. Study screen

- [x] 2.1 Replace the combined strings in the study more menu (`lib/features/study/study_screen.dart`) with plain labels through `withShortcut`, add the missing plain keys to `app_en.arb` and `app_de.arb` (German taken from the current strings without the parenthesis) and delete the combined keys; verify with `grep` that no `*Shortcut` key used only there is left in the ARB files
- [x] 2.2 Use `MenuEntryLabel` for the flag menu entries and `withShortcut` for the Undo, Stop recording and Leave tooltips; verify `fvm flutter analyze` is clean
- [x] 2.3 Update `test/study_test.dart` finders to the plain labels ("Edit card", "Suspend note", …), keep the hint checks (previous card info, flags) on Linux, and add tests for the spec scenarios "Study menu on a phone", "Study menu on desktop", "Flag menu on a phone", "Hardware keyboard on a tablet" and "German key names"; verify `fvm flutter test test/study_test.dart` passes

## 3. App bar, browser and deck list

- [x] 3.1 Use `MenuEntryLabel` for the undo/redo items in `lib/app/scaffold.dart`; add the "Undo menu on iOS" scenario to `test/app_shell_test.dart` and a Linux case that still shows the shortcut; verify `fvm flutter test test/app_shell_test.dart` passes
- [x] 3.2 In `lib/features/browse/browse_screen.dart`, use `MenuEntryLabel` for invert selection (hint by platform, not by `narrow`) and `withShortcut` for set due date, deleting `setDueDateWithShortcut`; add the "Mobile browser" scenario (iOS) and a Linux case to `test/browse_test.dart`; verify `fvm flutter test test/browse_test.dart` passes
- [x] 3.3 Use `withShortcut` for the Study deck tooltip in `lib/features/decks/deck_list_screen.dart`, deleting `studyDeckShortcut`; update the tooltip finder in `test/deck_list_test.dart`; verify `fvm flutter test test/deck_list_test.dart` passes

## 4. Editor

- [x] 4.1 Replace the formatting and HTML/visual toggle tooltips in `lib/features/editor/rich_field.dart` with plain labels through `withShortcut`, deleting the combined keys from both ARB files; update tooltip finders in `test/editor_test.dart` and `test/rich_field_test.dart` and add the "Tooltip on a tablet" scenario; verify both test files pass
- [x] 4.2 Leave out the Add/Save button's `Tooltip` in `lib/features/editor/note_editor_screen.dart` when hints are hidden, and add the "Narrow desktop window" scenario (Linux, width below 360, math menu shows "Ctrl+M, M"); verify `fvm flutter test test/editor_test.dart test/editor_edit_test.dart` passes

## 5. Settings

- [x] 5.1 Hide the "Create backup now" subtitle in `lib/features/settings/settings_screen.dart` when hints are hidden; run the existing subtitle test on Windows or Linux and add the "Backup entry on a phone" scenario; verify `fvm flutter test test/settings_test.dart` passes

## 6. Integration checks

- [x] 6.1 Verify `grep -rn "narrowScreen" lib` no longer guards any shortcut hint, and that no ARB value still embeds a keyboard key in parentheses (`grep -nE '\((Ctrl|Shift|Esc|\{key\}|\{shortcut\}|.)\)"' lib/l10n/*.arb` returns only `labelWithShortcut`)
- [x] 6.2 Run `fvm flutter analyze` and the full `fvm flutter test` suite (including `test/l10n_test.dart` and `test/no_literal_text_test.dart`) and verify both pass
- [x] 6.3 Run the app on Android (or a Chrome mobile emulation of the web build) and on Linux, open the study more menu, the editor's math menu and Settings, and verify hints are absent on Android and present on Linux
