# Design

## Context

- A card's persistent state is `core.Card` (snapshot JSON plus journaled ops). Card-level writes go through `Collection.commit` with an op listing the changed cards; `Change.Cards` then refreshes open screens.
- The browser's rows come from `Collection.Rows` → `BrowseService.Rows`; the study screen gets `StudyCard` from `StudyService.state`.
- Search terms are parsed in `golang/search/search.go` and matched against a `search.Card` the collection fills per card.
- `.apkg` cards are read in `importer/apkg/reader.go`; Anki stores the flag in the low 3 bits of `cards.flags` (legacy and modern schema alike).

## Goals / Non-Goals

**Goals:**
- One integer per card, no new storage structures, readable by old code paths as "no flag".
- Flag colours defined once and used by study, browser rows and settings.

**Non-Goals:**
- Flags on notes, or more than seven flags.
- A sidebar entry for flags (the browser sidebar is a separate change).

## Decisions

### 1. `Card.Flag int` with 0 = none, 1–7 = flag

Stored with `omitempty`, so existing snapshots, journals and `.cardflip` backups load unchanged and no backup format bump is needed. Setting a flag is a card op like suspend; it does not touch scheduling or the review log.

Alternative: flags as tags on the note (`flag::red`). Rejected: Anki flags are per card, and a note with a reversed card would flag both.

### 2. Toggle semantics

`SetFlag(ids, n)`: if every given card already has flag `n`, clear it; otherwise set `n` on all of them. This gives Anki's toggle behaviour for the single study card and a predictable result for a mixed browser selection. The browser also offers an explicit "No flag" item, which always clears.

### 3. Flag names in `Settings`

`FlagNames [7]string`; an empty entry means the default colour name. Renaming is in Settings (a "Flags" section). Names appear in the study menu, the browser's bulk action and the row tooltip.

### 4. Row colours

Flag colour wins over suspended. The tint is a low-alpha overlay of the flag colour (or yellow for suspended) on the row background, so it stays readable in light and dark themes. The row keeps its existing "suspended" text for accessibility.

### 5. `{{CardFlag}}`

Added to the special fields map in `render.Note.context`, which needs the card's flag passed in alongside the card ordinal. The value is `flagN`, as in Anki, so shared-deck CSS like `.flag1` keeps working.

## Risks / Trade-offs

- [Ctrl+1…7 collides with OS or browser shortcuts on the web (tab switching in Chrome)] → Register the shortcuts on the study screen's `Shortcuts` widget, which receives them first in Flutter web; if a browser still steals them, the **More › Flag** menu remains.
- [Re-importing a package with "keep scheduling" overwrites flags changed in Cardflip] → Re-import leaves existing cards' scheduling unchanged per the data-import spec; flags follow the same rule and are only set on newly added cards.
