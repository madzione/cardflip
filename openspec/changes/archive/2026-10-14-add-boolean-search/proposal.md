# Proposal

## Why

Cardflip's search combines terms only with AND. Everyday searches such as "Spanish or French cards due today" (`(deck:Spanish or deck:French) is:due`) or "not verbs and not nouns" (`-(tag:verb or tag:noun)`) cannot be written. OR and grouping are basic in Anki's search syntax, and many guides and shared workflows use them.

## What Changes

- `or` (case-insensitive) between terms matches either; `and` may be written explicitly (it is still the default between terms).
- Parentheses group terms: `dog (cat or mouse)`, and `-( … )` negates a group.
- Precedence as in Anki: AND binds tighter than OR, so `a b or c` means `(a and b) or c`.
- Escaping: `\"`, `\(`, `\)`, `\-`, `\:` and `\\` stand for the literal characters; a quoted term `"a (b)"` or `deck:"My Deck"` is taken literally apart from escapes.
- Errors: unbalanced parentheses, an `or` without a term on one side, and an empty group are reported with the position, instead of returning an empty list.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `card-browser`: new requirements for OR, grouping and escaping in searches.

## Impact

- Go: `golang/search/search.go` gets a tokenizer and a recursive-descent parser producing an expression tree (`and`, `or`, `not`, term); `Query.Match` evaluates the tree; the existing term kinds are unchanged.
- Tests: `golang/search/search_test.go` (precedence, nesting, escapes, errors), `golang/core/search_test.go`.
- Overlap: the other search changes (`add-text-and-field-search`, `add-card-property-search`, `add-recent-event-search`, flags, presets) add term kinds; they plug into either the current term list or this tree without conflict, but the tokenizer must be shared, so whichever lands second rebases its term parsing onto the other's tokenizer.
