# Proposal

## Why

Cardflip has no statistics. Users cannot see how much they studied today, how many reviews are coming in the next weeks, whether their retention matches the target, or how their cards' memory is distributed. Anki's statistics screen answers these, and with an FSRS-only scheduler the retention and memory graphs are the main way to check that scheduling works. Cardflip already keeps the full review history, so the data exists.

## What Changes

- A **Statistics** screen, opened from the deck list (for the whole collection) and from each deck's menu (for that deck).
- **Scope**: the deck with its subdecks, the whole collection, or any search; period: 1 month, 3 months, 1 year, all history (default 1 month for time-based graphs).
- **Today**: cards studied, time spent, Again count and correct percentage, and the split into learning, review, relearning and filtered reviews.
- **Future due**: forecast of due cards per day for the period, overdue cards, cumulative line, and the average daily load.
- **Calendar**: heatmap of reviews per day over the year.
- **Reviews** and **Review time**: per day, week or month, split into learning, young (interval < 21 days), mature, relearning and filtered.
- **Card counts**: new, learning, relearning, young, mature, suspended and buried.
- **Review intervals**: histogram.
- **Stability**, **Difficulty** and **Retrievability**: histograms, average retrievability, and estimated total knowledge (sum of retrievability).
- **Hourly breakdown**: reviews and success rate by hour of day.
- **Answer buttons**: ratings per button for learning, young and mature cards, with correct percentage.
- **True retention**: pass and fail counts and retention per period (today, yesterday, last week, month, year, all time) for young and mature cards, compared with the desired retention.
- **Save as PDF** of the whole screen.
- Collection-wide statistics include reviews of deleted notes, since the review log keeps them.
- Statistics are computed in the background; the screen shows sections as they are ready.

## Capabilities

### New Capabilities

- `statistics`: the statistics screen, its scope and period, and each graph and table.

### Modified Capabilities

- `app-shell`: new requirement for opening statistics.

## Impact

- Go: a new `golang/stats` package computing every section from cards and the review log for a set of card IDs (from deck, collection or search) and a period, with the study day boundary; `StatsService` with one call per section so the UI can load them independently; deleted cards' entries counted in collection scope.
- Flutter: new `lib/features/stats/` (screen, section widgets, charts with `fl_chart`, PDF with `pdf` and `printing`).
- Performance: all sections for 50,000 cards and 500,000 review entries within 2 seconds on a typical desktop.
- Tests: `golang/stats/*_test.go` with a fixture collection and known answers, `golang/svc_test.go`, widget tests per section.
