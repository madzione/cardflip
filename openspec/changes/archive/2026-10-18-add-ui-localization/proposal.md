# Proposal

## Why

Cardflip's interface is English only, while Anki is translated into dozens of languages. Many learners, particularly beginners and younger students, are more comfortable in their own language. Translating also requires every user-facing string to go through one mechanism, which the code does not have yet: strings are inline in Dart widgets, and Go returns English error messages.

## What Changes

- All user-facing text in the Flutter UI moves to translation files, including dialogs, errors, empty states and the relative time formats (`<10m`, `3d`, `1.2mo`).
- Go returns user errors as a code with parameters, alongside the English text, so the app can show them translated.
- Dates and numbers use the selected locale's formats.
- A **Language** setting: "System default" (default) or a specific language. The first shipped translation is German, next to English; adding a language means adding one translation file.
- Missing translations fall back to English per string.
- Card content is never translated.
- Built-in names follow the interface language when they are created: the stock note types (name, fields, card types), the first deck and the default preset of a new collection. Note types, decks and presets that already exist keep their names.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `app-shell`: new requirements for the language setting and translated text.

## Impact

- Dependencies: `flutter_localizations` and `intl` (Flutter SDK), `gen_l10n` with ARB files in `lib/l10n/`.
- Go: user errors (`userErr` in `golang/core`) gain a stable code and parameters; the services pass them to Dart in the error payload; `core.Settings` gains `Language`.
- Go: `core.BuiltinNames`, set by the app (`app.SetBuiltinNames`), names what new collections and stock note types are created with; Image Occlusion note types are recognised by their template instead of English field names.
- Flutter: every screen under `lib/features/` and `lib/widgets/time_format.dart`, `lib/app/app.dart` (locales, delegates), `lib/app/errors.dart` (mapping codes to messages), `settings_screen.dart`.
- Tests: a test that every ARB file has the same keys as English; widget tests run in English by default; `test/settings_test.dart` switches to German.
- Assumption: German is chosen as the first translation; any other language can be added later the same way.
