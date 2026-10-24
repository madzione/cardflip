# Tasks

## 1. Infrastructure

- [x] 1.1 Enable `gen_l10n` (`l10n.yaml`, `lib/l10n/app_en.arb`), add localisation delegates and supported locales to `lib/app/app.dart`, and add `Language` to `core.Settings` with a setter. Verify `fvm flutter gen-l10n` runs and `fvm flutter test` passes.
- [x] 1.2 Give `userErr` a code and parameters in `golang/core` and the services, and send them in the error payload; map them in `lib/app/errors.dart` with English fallback. Add Go tests that errors carry codes and a Dart test for the fallback. Verify `go test ./...` and `fvm flutter test` pass.

## 2. Extraction

- [x] 2.1 Move all strings in `lib/features/decks`, `lib/features/study` and `lib/widgets` (including `time_format.dart`) to ARB. Verify the existing tests pass in English.
- [x] 2.2 Move all strings in `lib/features/browse`, `lib/features/editor`, `lib/features/import`, `lib/features/settings` and `lib/app` to ARB. Add the literal-text check test from the design. Verify it and the suite pass.
  - Conventions used: widgets read `context.l10n`; code without a context reads `appL10n`; Go texts go through `goText()`; count phrases use ICU plurals; int placeholders named max/min/total/number/… have no `decimalPattern` format (no grouping). After changing a `go*` ARB message run `fvm dart run tool/gen_go_messages.dart`.

## 3. German and setting

- [x] 3.1 Add `app_de.arb` with all keys and a test that every ARB has the same keys as English. Verify it passes.
- [x] 3.2 Add the Language setting to `settings_screen.dart`. Add `test/settings_test.dart` tests for switching to German and for the system default. Verify they pass.

## 4. Built-in names

- [x] 4.1 Add `core.BuiltinNames` (deck, preset, stock note types with fields and card types) with English defaults and validation, and `render.RenameFields`. Use it for a new collection, Add note type, Image Occlusion on first use, a replaced collection without note types and Check database's default deck. Recognise Image Occlusion by its template and report its field indexes in the note type API; accept the built-in Basic name in the Basic fallbacks. Add `app.SetBuiltinNames`. Add Go tests. Verify `go test ./...` passes.
- [x] 4.2 Add the built-in names to the ARB files, send them from Dart before opening and on language changes, show stock note types translated in Add note type, use the reported Image Occlusion field indexes in the editor, and name the default preset in `presetUsedBy` from the data. Add Dart tests for the German add dialog and editor. Verify `fvm flutter test` passes.

## 5. Integration check

- [x] 5.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Check every screen in German on a phone layout for overflow.
  - Progress (2026-10-18): `go test ./...`, `fvm flutter analyze` and `fvm flutter test` (645 tests, after the built-in names) pass. A temporary widget-test walk-through in German at 400×800 and 360×740 (deck list with its menu, rename, new deck and filtered deck dialogs; deck overview and options; study with the answer and card menu; browser with details; add; settings; statistics; import) found no overflow; the same run at 220 px did report overflows, so detection works. Left: a visual pass with real fonts on a phone or phone-sized window, including the dialogs not walked through.
