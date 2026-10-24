# Proposal

## Why

Cardflip's free-text search matches substrings in all fields. Anki users rely on more precise matching: searching a single field (`front:dog`), wildcards (`d_g`, `d*g`), whole words (`w:dog`), accent-insensitive search (`nc:uber`), searching cloze text without its markup, regular expressions, and finding notes with an empty field. Without field-limited search in particular, many cleanup tasks are impossible.

## What Changes

- Wildcards in free text: `_` matches one character, `*` any run of characters; `\_` and `\*` match them literally.
- `w:` matches whole words: `w:dog` finds "dog" but not "doggy"; `w:dog*` and `w:*dog` allow prefixes and suffixes.
- Field search: `front:dog` matches when the field Front is exactly "dog" (case-insensitive, HTML stripped); wildcards make partial matches (`front:*dog*`); field names may use wildcards (`fr*:text`); quoting covers spaces (`"animal front:a dog"`).
- Empty fields: `front:` matches an empty Front, `front:_*` a non-empty one, `front:*` any note that has a Front field.
- `nc:` ignores accents and other combining marks (`nc:uber` finds "über"); a setting **Ignore accents in search** applies this to all text and field searches.
- `sc:` searches text with cloze markup removed (`sc:capital of France` finds `{{c1::capital}} of France`).
- Regular expressions: `re:` on all fields, `field:re:` on one field and `tag:re:` on tags, matched against the raw field HTML, case-insensitive unless the expression starts with `(?-i)`.
- Text search decodes HTML entities, so `&` finds a field stored as `&amp;`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for wildcards, word search, field search, accent and cloze handling, regular expressions and entity decoding.
- `app-shell`: new requirement for the "Ignore accents in search" setting.

## Impact

- Go: `golang/search` gains term kinds (wildcard text, word, field with name pattern, `nc:`, `sc:`, `re:`, field `re:`, tag `re:`); `search.Card` needs per-field text (stripped and raw) and field names, filled in `core/search.go` with caching; regexes compiled once per query (RE2).
- Flutter: `settings_screen.dart`.
- Tests: `golang/search/search_test.go`, `golang/core/search_test.go`, `test/settings_test.dart`; a performance check of field and regex searches on 50,000 cards.
- Overlap: the shared tokenizer with `add-boolean-search` (see that change).
