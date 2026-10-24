# Proposal

## Why

Anki users mark cards for follow-up with seven coloured flags (fix later, ask a teacher, leech candidates, …), and many shared decks and personal workflows rely on them. Cardflip has no flags at all, and `.apkg` import silently drops them, so a user switching from Anki loses that information on the first import.

## What Changes

- Every card can carry one of seven flags (1 red, 2 orange, 3 green, 4 blue, 5 pink, 6 turquoise, 7 purple) or none. Flags belong to the card, not the note.
- The user can rename each flag; the default names are the colour names.
- Study screen: Ctrl+1…7 (Cmd on macOS) toggles the flag on the current card, the **More** menu has a Flag submenu, and the current card's flag is shown as a coloured marker.
- Browser: a bulk **Flag** action (set or clear), and a coloured flag marker on each row. Rows of suspended cards get a yellow tint, so the "suspended" state is visible at a glance, not only in text.
- Search: `flag:0` (no flag) to `flag:7`.
- Templates: the special field `{{CardFlag}}` renders `flag1` … `flag7`, or nothing when there is no flag.
- `.apkg` import keeps flags when scheduling is kept. Backups include flags.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `collection`: new requirements for card flags and flag names.
- `study-session`: new requirement for flagging the current card.
- `card-browser`: new requirements for the flag bulk action, `flag:` search and row colours.
- `card-rendering`: new requirement for the `{{CardFlag}}` special field.
- `data-import`: new requirement that `.apkg` import keeps flags.

## Impact

- Go: `core.Card` gains `Flag int` (`json:"flag,omitempty"`, so old snapshots and backups read as "no flag" without migration); `core.Settings` gains `FlagNames [7]string`; `Collection.SetFlag(ids, flag)`; `search` gains a `flag` term; `render` gains the `CardFlag` special field; `importer/apkg` reads the low 3 bits of `cards.flags`; `BrowseService`, `StudyService` and `AppService` gain flag calls; `Row` and `StudyCard` DTOs carry the flag.
- Flutter: `lib/features/study/study_screen.dart` (shortcuts, menu, marker), `lib/features/browse/browse_screen.dart` (bulk action, row marker and tint), `lib/features/browse/card_details.dart`, `lib/features/settings/settings_screen.dart` (flag names), a shared flag colour table in `lib/app/theme.dart`.
- Tests: `golang/core/cards_test.go`, `golang/search/search_test.go`, `golang/render/template_test.go`, `golang/importer/apkg/apkg_test.go`, `golang/svc_test.go`, `test/study_test.dart`, `test/browse_test.dart`, `test/support/fake_go.dart`.
