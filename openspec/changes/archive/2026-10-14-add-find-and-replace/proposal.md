# Proposal

## Why

Fixing a systematic mistake (a misspelt word across a deck, an outdated term, a formatting pattern from an import) currently means editing notes one by one. Anki's Find and Replace changes text across many notes at once, in all fields, one field or tags, with optional regular expressions. It is one of the most used maintenance tools for imported decks.

## What Changes

- A **Find and Replace** action in the browser, applied to the selected notes or, by choice, to all notes.
- Inputs: find text, replacement, where (all fields, one named field, or tags), "Treat input as regular expression", "Ignore case" (on by default), and "Selected notes only".
- Plain text matches literally. With regular expressions, the replacement can use groups (`$1`, `${name}`), using RE2 syntax.
- Field replacements apply to the stored HTML of the field, as in Anki.
- Before applying, the app shows how many notes will change; afterwards it reports how many changed. Cards are regenerated as for any note edit (a field that becomes empty can remove cards, with the usual rules).
- The whole replacement is one operation, so it is applied completely or not at all.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirement for replacing text in notes.
- `card-browser`: new requirement for the Find and Replace dialog.

## Impact

- Go: `Collection.FindReplace(noteIDs, find, replace, target, regex, ignoreCase)` and a preview count, in `golang/core/notes.go`, applied in one op with card regeneration; `BrowseService` calls.
- Flutter: a dialog in `lib/features/browse/find_replace.dart` and a toolbar action in `browse_screen.dart`.
- Tests: `golang/core/notes_test.go`, `golang/svc_test.go`, `test/browse_test.dart`.
