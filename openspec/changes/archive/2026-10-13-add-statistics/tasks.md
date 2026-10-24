# Tasks

## 1. Computation

- [x] 1.1 Add `golang/stats` with scope resolution (deck subtree, collection including deleted cards' entries, search), period handling and study-day bucketing, and a fixture collection with a hand-computed review log. Verify `go test ./stats/...` passes.
- [x] 1.2 Implement the today summary, future due, and card counts. Add tests with known answers for each spec scenario. Verify they pass.
- [x] 1.3 Implement calendar, reviews, review time, hourly breakdown and answer buttons. Add tests. Verify they pass.
- [x] 1.4 Implement interval, stability, difficulty and retrievability histograms, average retrievability, estimated knowledge, and true retention. Add tests. Verify they pass. Add a benchmark for all sections on 50,000 cards and 500,000 entries and verify it stays under 2 seconds.

## 2. Service

- [x] 2.1 Add `StatsService` with one call per section taking scope and period. Run `fluffy generate`, update `test/support/fake_go.dart`. Verify `go test ./...` passes and `fluffy generate --check` is clean.

## 3. UI

- [x] 3.1 Add `lib/features/stats/stats_screen.dart` with scope and period controls and sections loading independently, and entries on the deck list and deck menu. Add widget tests for opening with each scope. Verify they pass.
- [x] 3.2 Add the section widgets with `fl_chart` (today, future due, calendar, reviews, review time, counts, intervals, memory histograms, hourly, buttons, true retention). Add a widget test per section with fake data. Verify they pass.
- [x] 3.3 Add Save as PDF with `pdf` and `printing`. Add a test that the PDF builder includes every section. Verify it passes and `fluffy build web` succeeds.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Open statistics on the 50,000-card fixture on desktop and web and check every section appears.
